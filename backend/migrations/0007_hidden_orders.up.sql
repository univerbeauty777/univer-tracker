-- 0007: hidden_at lets ops "delete" orders from the dashboard without
-- removing them from the WooCommerce source-of-truth. The list/facet
-- queries skip rows where hidden_at IS NOT NULL; the WC sync also
-- preserves the flag when it touches an order so a hidden record never
-- bounces back after the next pull.

ALTER TABLE orders ADD COLUMN IF NOT EXISTS hidden_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_orders_hidden_at
    ON orders (hidden_at)
    WHERE hidden_at IS NOT NULL;
