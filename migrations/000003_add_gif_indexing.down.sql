-- Down Migration
DROP INDEX IF EXISTS idx_messages_gif_provider;
DROP INDEX IF EXISTS idx_messages_metadata_gin;
COMMENT ON COLUMN messages.metadata IS NULL;
