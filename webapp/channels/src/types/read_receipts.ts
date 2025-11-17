// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

export type ReadCursor = {
    channel_id: string;
    user_id: string;
    last_post_seq: number;
    updated_at: number;
};

export type ReadReceiptsState = {
    // 每个频道的读游标: channel_id -> user_id -> cursor
    cursors: {
        [channelId: string]: {
            [userId: string]: ReadCursor;
        };
    };
};
