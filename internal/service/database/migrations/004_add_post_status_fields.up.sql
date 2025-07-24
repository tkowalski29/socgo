-- Add status fields to posts table
ALTER TABLE posts ADD COLUMN published_at DATETIME;
ALTER TABLE posts ADD COLUMN status TEXT DEFAULT 'pending';
ALTER TABLE posts ADD COLUMN error_message TEXT;

-- Create index on status for faster lookups
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status); 