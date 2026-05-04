DROP INDEX IF EXISTS idx_orders_hidden_at;
ALTER TABLE orders DROP COLUMN IF EXISTS hidden_at;
