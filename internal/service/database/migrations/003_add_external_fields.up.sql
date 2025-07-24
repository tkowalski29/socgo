-- Add external fields to posts table
ALTER TABLE posts ADD COLUMN external_id TEXT;
ALTER TABLE posts ADD COLUMN external_url TEXT;

-- Create index on external_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_posts_external_id ON posts(external_id); 