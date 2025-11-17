// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"context"
	"encoding/json"

	"github.com/redis/rueidis"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// AdvanceChannelReadCursor updates the user's read cursor in a channel
// This is the core method that tracks what messages a user has read
func (a *App) AdvanceChannelReadCursor(rctx request.CTX, userId, channelId string, newSeq int64) (*model.ChannelReadCursor, *model.AppError) {
	// 1. Permission check - user must have read access to the channel
	if !a.SessionHasPermissionToChannel(rctx, *rctx.Session(), channelId, model.PermissionReadChannelContent) {
		return nil, model.NewAppError("AdvanceChannelReadCursor", "api.channel.read_cursor.permission.app_error", nil, "", 403)
	}

	// 2. Get the old cursor (if exists)
	oldCursor, err := a.Srv().Store().ChannelReadCursor().Get(channelId, userId)
	prevSeq := int64(0)
	if err == nil {
		prevSeq = oldCursor.LastPostSeq
	}

	// 3. Only update if the new sequence is greater (prevent cursor rollback)
	if newSeq <= prevSeq {
		rctx.Logger().Debug("Read cursor not advanced - new seq not greater than previous",
			mlog.String("user_id", userId),
			mlog.String("channel_id", channelId),
			mlog.Int("prev_seq", int(prevSeq)),
			mlog.Int("new_seq", int(newSeq)),
		)
		return oldCursor, nil
	}

	// 4. Create and save the new cursor
	cursor := &model.ChannelReadCursor{
		ChannelId:   channelId,
		UserId:      userId,
		LastPostSeq: newSeq,
		UpdatedAt:   model.GetMillis(),
	}

	if err := a.Srv().Store().ChannelReadCursor().Upsert(cursor); err != nil {
		return nil, model.NewAppError("AdvanceChannelReadCursor", "app.channel.read_cursor.save.app_error", nil, err.Error(), 500)
	}

	// 5. Publish event for ReadIndexService to consume
	event := &model.ReadCursorEvent{
		Type:        "channel_read_advanced",
		EventId:     model.NewId(),
		ChannelId:   channelId,
		UserId:      userId,
		PrevLastSeq: prevSeq,
		NewLastSeq:  newSeq,
		Timestamp:   cursor.UpdatedAt,
	}

	if err := a.publishReadCursorEvent(rctx, event); err != nil {
		rctx.Logger().Error("Failed to publish read cursor event", mlog.Err(err))
		// Don't fail the request if event publishing fails
	}

	// 6. Send WebSocket event to notify other users in the channel
	a.publishReadCursorWebSocketEvent(rctx, channelId, userId, newSeq)

	rctx.Logger().Debug("Read cursor advanced",
		mlog.String("user_id", userId),
		mlog.String("channel_id", channelId),
		mlog.Int("prev_seq", int(prevSeq)),
		mlog.Int("new_seq", int(newSeq)),
	)

	return cursor, nil
}

// AdvanceChannelReadCursorByPost is a convenience method that derives the sequence from a post
func (a *App) AdvanceChannelReadCursorByPost(rctx request.CTX, userId, postId string) (*model.ChannelReadCursor, *model.AppError) {
	post, err := a.GetSinglePost(rctx, postId, false)
	if err != nil {
		return nil, err
	}

	// Use CreateAt as the sequence number (it's monotonically increasing)
	return a.AdvanceChannelReadCursor(rctx, userId, post.ChannelId, post.CreateAt)
}

// GetChannelReadCursor retrieves the read cursor for a user in a channel
func (a *App) GetChannelReadCursor(rctx request.CTX, userId, channelId string) (*model.ChannelReadCursor, *model.AppError) {
	cursor, err := a.Srv().Store().ChannelReadCursor().Get(channelId, userId)
	if err != nil {
		return nil, model.NewAppError("GetChannelReadCursor", "app.channel.read_cursor.get.app_error", nil, err.Error(), 404)
	}
	return cursor, nil
}

// GetChannelReadCursorsForUser returns all read cursors for a user across all channels
func (a *App) GetChannelReadCursorsForUser(rctx request.CTX, userId string) ([]*model.ChannelReadCursor, *model.AppError) {
	cursors, err := a.Srv().Store().ChannelReadCursor().GetForUser(userId)
	if err != nil {
		return nil, model.NewAppError("GetChannelReadCursorsForUser", "app.channel.read_cursor.get_for_user.app_error", nil, err.Error(), 500)
	}
	return cursors, nil
}

// GetChannelReadCursors returns all read cursors for a channel
func (a *App) GetChannelReadCursors(rctx request.CTX, channelId string) ([]*model.ChannelReadCursor, *model.AppError) {
	cursors, err := a.Srv().Store().ChannelReadCursor().GetForChannel(channelId)
	if err != nil {
		return nil, model.NewAppError("GetChannelReadCursors", "app.channel.read_cursor.get_for_channel.app_error", nil, err.Error(), 500)
	}
	return cursors, nil
}

// publishReadCursorEvent publishes the read cursor event to Redis Stream for ReadIndexService
func (a *App) publishReadCursorEvent(rctx request.CTX, event *model.ReadCursorEvent) error {
	// Get Redis client from platform
	redisClientInterface := a.Srv().Platform().GetRedisClient()
	if redisClientInterface == nil {
		// Redis not configured, just log
		rctx.Logger().Debug("Redis not configured, skipping event publish",
			mlog.String("event_id", event.EventId),
		)
		return nil
	}

	// Type assert to rueidis.Client
	redisClient, ok := redisClientInterface.(rueidis.Client)
	if !ok {
		rctx.Logger().Error("Redis client type assertion failed",
			mlog.String("event_id", event.EventId),
		)
		return nil
	}

	// Serialize event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		rctx.Logger().Error("Failed to marshal read cursor event",
			mlog.String("event_id", event.EventId),
			mlog.Err(err),
		)
		return err
	}

	// Publish to Redis Stream using rueidis
	streamName := "read_cursor_events"
	ctx := context.Background()
	
	// Use XADD command to add to stream
	cmd := redisClient.B().Xadd().
		Key(streamName).
		Id("*").  // Auto-generate ID
		FieldValue().
		FieldValue("data", string(eventData)).
		Build()
	
	err = redisClient.Do(ctx, cmd).Error()
	if err != nil {
		rctx.Logger().Error("Failed to publish read cursor event to Redis Stream",
			mlog.String("stream", streamName),
			mlog.String("event_id", event.EventId),
			mlog.Err(err),
		)
		return err
	}

	rctx.Logger().Info("Published read cursor event to Redis Stream",
		mlog.String("stream", streamName),
		mlog.String("event_id", event.EventId),
		mlog.String("channel_id", event.ChannelId),
		mlog.String("user_id", event.UserId),
		mlog.Int("new_seq", int(event.NewLastSeq)),
	)

	return nil
}

// invalidateReadReceiptsCacheForChannel invalidates all cached read receipt counts for a channel
func (a *App) invalidateReadReceiptsCacheForChannel(channelId string) {
	// Note: In a production system, you might want to:
	// 1. Keep a list of post IDs per channel
	// 2. Or use a cache key pattern like "post_read_count:channel:{channelId}:*"
	// For now, we rely on the 30-second TTL to eventually update
	// The WebSocket event will trigger immediate UI updates
}

// publishReadCursorWebSocketEvent sends a WebSocket event to notify users about read cursor changes
func (a *App) publishReadCursorWebSocketEvent(rctx request.CTX, channelId, userId string, lastPostSeq int64) {
	message := model.NewWebSocketEvent(model.WebsocketEventReadCursorAdvanced, "", channelId, "", nil, "")
	message.Add("user_id", userId)
	message.Add("last_post_seq", lastPostSeq)
	message.Add("channel_id", channelId)
	
	a.Publish(message)
}

// AutoAdvanceReadCursorOnChannelView automatically advances the read cursor when user views a channel
// This is called from the ViewChannel API
func (a *App) AutoAdvanceReadCursorOnChannelView(rctx request.CTX, userId, channelId string) *model.AppError {
	// Get the latest post in the channel to determine the sequence
	posts, err := a.GetPostsPage(rctx, model.GetPostsOptions{
		ChannelId: channelId,
		Page:      0,
		PerPage:   1,
	})
	
	if err != nil {
		// If we can't get posts, just log and continue (don't fail the view operation)
		rctx.Logger().Debug("Could not get latest post for auto-advance cursor",
			mlog.String("channel_id", channelId),
			mlog.Err(err),
		)
		return nil
	}

	if len(posts.Posts) == 0 {
		// No posts in channel, nothing to advance
		return nil
	}

	// Get the latest post's CreateAt as the sequence
	var latestSeq int64
	for _, post := range posts.Posts {
		if post.CreateAt > latestSeq {
			latestSeq = post.CreateAt
		}
	}

	if latestSeq == 0 {
		return nil
	}

	// Advance the cursor
	_, appErr := a.AdvanceChannelReadCursor(rctx, userId, channelId, latestSeq)
	if appErr != nil {
		// Log but don't fail - this is a best-effort operation
		rctx.Logger().Warn("Failed to auto-advance read cursor on channel view",
			mlog.String("user_id", userId),
			mlog.String("channel_id", channelId),
			mlog.Err(appErr),
		)
	}

	return nil
}

// CleanupOldReadCursors removes read cursors older than the specified days
// This should be called by a scheduled job
func (a *App) CleanupOldReadCursors(rctx request.CTX, olderThanDays int) *model.AppError {
	if olderThanDays <= 0 {
		return model.NewAppError("CleanupOldReadCursors", "app.channel.read_cursor.cleanup.invalid_days.app_error", nil, "", 400)
	}

	olderThan := model.GetMillis() - int64(olderThanDays*24*60*60*1000)
	
	if err := a.Srv().Store().ChannelReadCursor().DeleteOldCursors(olderThan); err != nil {
		return model.NewAppError("CleanupOldReadCursors", "app.channel.read_cursor.cleanup.app_error", nil, err.Error(), 500)
	}

	rctx.Logger().Info("Cleaned up old read cursors",
		mlog.Int("older_than_days", olderThanDays),
	)

	return nil
}
