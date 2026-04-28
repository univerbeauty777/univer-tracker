BEGIN;

ALTER TABLE orders
    DROP COLUMN IF EXISTS shipping_method;

DROP INDEX IF EXISTS idx_shipments_idle_since;
DROP INDEX IF EXISTS idx_shipments_health;

ALTER TABLE shipments
    DROP COLUMN IF EXISTS risk_score,
    DROP COLUMN IF EXISTS idle_since,
    DROP COLUMN IF EXISTS health;

DROP TABLE IF EXISTS sync_state;

COMMIT;
