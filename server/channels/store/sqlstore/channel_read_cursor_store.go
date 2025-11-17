// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/v8/channels/store"
	"github.com/pkg/errors"
)

type SqlChannelReadCursorStore struct {
	*SqlStore
}

func newSqlChannelReadCursorStore(sqlStore *SqlStore) store.ChannelReadCursorStore {
	return &SqlChannelReadCursorStore{SqlStore: sqlStore}
}

// Upsert inserts or updates a read cursor for a user in a channel
func (s *SqlChannelReadCursorStore) Upsert(cursor *model.ChannelReadCursor) error {
	cursor.PreSave()

	if err := cursor.IsValid(); err != nil {
		return err
	}

	query := s.getQueryBuilder().
		Insert("channel_read_cursors").
		Columns("channel_id", "user_id", "last_post_seq", "updated_at").
		Values(cursor.ChannelId, cursor.UserId, cursor.LastPostSeq, cursor.UpdatedAt).
		Suffix("ON CONFLICT (channel_id, user_id) DO UPDATE SET last_post_seq = GREATEST(channel_read_cursors.last_post_seq, EXCLUDED.last_post_seq), updated_at = EXCLUDED.updated_at")

	queryString, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build upsert query")
	}

	if _, err := s.GetMaster().Exec(queryString, args...); err != nil {
		return errors.Wrap(err, "failed to upsert channel read cursor")
	}

	return nil
}

// Get retrieves the read cursor for a specific user in a channel
func (s *SqlChannelReadCursorStore) Get(channelId, userId string) (*model.ChannelReadCursor, error) {
	var cursor model.ChannelReadCursor

	query := s.getQueryBuilder().
		Select("channel_id", "user_id", "last_post_seq", "updated_at").
		From("channel_read_cursors").
		Where("channel_id = ? AND user_id = ?", channelId, userId)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build get query")
	}

	if err := s.GetReplica().Get(&cursor, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("ChannelReadCursor", fmt.Sprintf("channel=%s,user=%s", channelId, userId))
		}
		return nil, errors.Wrap(err, "failed to get channel read cursor")
	}

	return &cursor, nil
}

// GetForChannel retrieves all read cursors for a channel
func (s *SqlChannelReadCursorStore) GetForChannel(channelId string) ([]*model.ChannelReadCursor, error) {
	var cursors []*model.ChannelReadCursor

	query := s.getQueryBuilder().
		Select("channel_id", "user_id", "last_post_seq", "updated_at").
		From("channel_read_cursors").
		Where("channel_id = ?", channelId).
		OrderBy("last_post_seq DESC")

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build get for channel query")
	}

	if err := s.GetReplica().Select(&cursors, queryString, args...); err != nil {
		return nil, errors.Wrap(err, "failed to get channel read cursors for channel")
	}

	return cursors, nil
}

// GetForUser retrieves all read cursors for a user across all channels
func (s *SqlChannelReadCursorStore) GetForUser(userId string) ([]*model.ChannelReadCursor, error) {
	var cursors []*model.ChannelReadCursor

	query := s.getQueryBuilder().
		Select("channel_id", "user_id", "last_post_seq", "updated_at").
		From("channel_read_cursors").
		Where("user_id = ?", userId).
		OrderBy("updated_at DESC")

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build get for user query")
	}

	if err := s.GetReplica().Select(&cursors, queryString, args...); err != nil {
		return nil, errors.Wrap(err, "failed to get channel read cursors for user")
	}

	return cursors, nil
}

// Delete removes a read cursor
func (s *SqlChannelReadCursorStore) Delete(channelId, userId string) error {
	query := s.getQueryBuilder().
		Delete("channel_read_cursors").
		Where("channel_id = ? AND user_id = ?", channelId, userId)

	queryString, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete query")
	}

	if _, err := s.GetMaster().Exec(queryString, args...); err != nil {
		return errors.Wrap(err, "failed to delete channel read cursor")
	}

	return nil
}

// DeleteForChannel removes all read cursors for a channel
func (s *SqlChannelReadCursorStore) DeleteForChannel(channelId string) error {
	query := s.getQueryBuilder().
		Delete("channel_read_cursors").
		Where("channel_id = ?", channelId)

	queryString, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete for channel query")
	}

	if _, err := s.GetMaster().Exec(queryString, args...); err != nil {
		return errors.Wrap(err, "failed to delete channel read cursors for channel")
	}

	return nil
}

// DeleteOldCursors removes cursors older than the specified timestamp
func (s *SqlChannelReadCursorStore) DeleteOldCursors(olderThan int64) error {
	query := s.getQueryBuilder().
		Delete("channel_read_cursors").
		Where("updated_at < ?", olderThan)

	queryString, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete old cursors query")
	}

	result, err := s.GetMaster().Exec(queryString, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete old channel read cursors")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		s.Logger().Info("Deleted old channel read cursors", mlog.Int("count", int(rowsAffected)), mlog.Int("older_than", int(olderThan)))
	}

	return nil
}
