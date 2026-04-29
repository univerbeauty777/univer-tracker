BEGIN;
DROP INDEX IF EXISTS idx_orders_tags;
ALTER TABLE orders DROP COLUMN IF EXISTS declared_value, DROP COLUMN IF EXISTS tags;
DROP INDEX IF EXISTS idx_shipments_posted_at;
DROP INDEX IF EXISTS idx_shipments_sla_state;
ALTER TABLE shipments
    DROP COLUMN IF EXISTS sla_breached_stage,
    DROP COLUMN IF EXISTS sla_state,
    DROP COLUMN IF EXISTS out_for_delivery_at,
    DROP COLUMN IF EXISTS at_destination_city_at,
    DROP COLUMN IF EXISTS in_transit_at,
    DROP COLUMN IF EXISTS posted_at,
    DROP COLUMN IF EXISTS ready_for_pickup_at,
    DROP COLUMN IF EXISTS preparing_at,
    DROP COLUMN IF EXISTS label_issued_at;
COMMIT;
