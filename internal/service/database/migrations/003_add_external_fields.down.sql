-- Remove external fields from posts table
DROP INDEX IF EXISTS idx_posts_external_id;
ALTER TABLE posts DROP COLUMN external_url;
ALTER TABLE posts DROP COLUMN external_id; 