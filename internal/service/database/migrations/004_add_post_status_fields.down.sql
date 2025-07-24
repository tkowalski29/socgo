-- Remove status fields from posts table
DROP INDEX IF EXISTS idx_posts_status;
ALTER TABLE posts DROP COLUMN error_message;
ALTER TABLE posts DROP COLUMN status;
ALTER TABLE posts DROP COLUMN published_at; 