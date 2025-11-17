-- Create channel_read_cursors table for read receipts feature
-- This table stores each user's reading progress (cursor) in each channel

CREATE TABLE IF NOT EXISTS channel_read_cursors (
    channel_id    VARCHAR(26) NOT NULL,
    user_id       VARCHAR(26) NOT NULL,
    last_post_seq BIGINT NOT NULL DEFAULT 0,
    updated_at    BIGINT NOT NULL,
    PRIMARY KEY (channel_id, user_id)
);

-- Index for querying all cursors in a channel
CREATE INDEX idx_channel_read_cursors_channel ON channel_read_cursors(channel_id);

-- Index for cleanup queries based on time
CREATE INDEX idx_channel_read_cursors_updated ON channel_read_cursors(updated_at);

-- Add comment for documentation
COMMENT ON TABLE channel_read_cursors IS 'Stores user reading progress cursors for read receipts feature';
COMMENT ON COLUMN channel_read_cursors.last_post_seq IS 'The sequence number (CreateAt timestamp) of the last post the user has read in this channel';
