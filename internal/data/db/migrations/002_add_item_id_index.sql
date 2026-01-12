-- Add index on item_id column for faster lookups
-- This complements the composite primary key (data_version_id, item_id)
-- and optimizes queries that filter by item_id alone

-- Drop old index if it exists (from previous migration attempts)
DROP INDEX IF EXISTS idx_items_item_id;

-- Create new index with unique name
CREATE INDEX IF NOT EXISTS idx_items_lookup_by_item_id ON items(item_id);
