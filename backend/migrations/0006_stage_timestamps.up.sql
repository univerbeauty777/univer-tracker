-- 0006 — per-stage timestamps + SLA breach + tags
-- Adds the columns rastreiaki uses to compute SLA per etapa, gargalos
-- e tags operacionais.

BEGIN;

ALTER TABLE shipments
    ADD COLUMN IF NOT EXISTS label_issued_at        TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS preparing_at           TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ready_for_pickup_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS posted_at              TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS in_transit_at          TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS at_destination_city_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS out_for_delivery_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sla_state              TEXT NOT NULL DEFAULT 'ON_TRACK',
    ADD COLUMN IF NOT EXISTS sla_breached_stage     TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_shipments_sla_state ON shipments (sla_state);
CREATE INDEX IF NOT EXISTS idx_shipments_posted_at ON shipments (posted_at NULLS LAST);

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS tags          JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS declared_value NUMERIC(12, 2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_orders_tags ON orders USING gin (tags);

COMMIT;
