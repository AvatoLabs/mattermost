-- Rollback migration for channel_read_cursors table

DROP INDEX IF EXISTS idx_channel_read_cursors_updated;
DROP INDEX IF EXISTS idx_channel_read_cursors_channel;
DROP TABLE IF EXISTS channel_read_cursors;
