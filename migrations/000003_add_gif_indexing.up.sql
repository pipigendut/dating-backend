-- Up Migration
CREATE INDEX idx_messages_metadata_gin ON messages USING GIN (metadata);
CREATE INDEX idx_messages_gif_provider ON messages ((metadata->>'gif_provider')) WHERE metadata->>'gif_provider' IS NOT NULL;

-- Add comment to clarify usage
COMMENT ON COLUMN messages.metadata IS 'Stores additional message data like GIF provider, image dimensions, etc.';
