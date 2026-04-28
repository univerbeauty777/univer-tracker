-- 0002 — sync state + shipment health/idle tracking
-- Created: 2026-04-28
--
-- Idempotent on purpose: if a previous run failed mid-way, replaying
-- this migration should converge to the desired state instead of
-- exploding on "column already exists".

BEGIN;

-- =============================================================================
-- sync_state: tracks last successful sync per entity so the worker is
-- restart-safe and can do incremental fetches.
-- =============================================================================
CREATE TABLE IF NOT EXISTS sync_state (
    entity          TEXT PRIMARY KEY,
    last_synced_at  TIMESTAMPTZ,
    extra           JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- shipments: add health and idle tracking so the dashboard can surface
-- at-risk and breached SLAs without recomputing on every request.
-- =============================================================================
ALTER TABLE shipments
    ADD COLUMN IF NOT EXISTS health        TEXT        NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS idle_since    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS risk_score    SMALLINT    NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_shipments_health     ON shipments (health);
CREATE INDEX IF NOT EXISTS idx_shipments_idle_since ON shipments (idle_since NULLS LAST);

-- =============================================================================
-- orders: cache the shipping_method so the table view doesn't need to
-- reach back to WooCommerce for every row. customer_phone already exists
-- in migration 0001, so it is intentionally not added here.
-- =============================================================================
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS shipping_method TEXT NOT NULL DEFAULT '';

-- =============================================================================
-- Default store (Lizzon). Multi-tenant comes later; for now everything we
-- sync is rooted here so the worker has a known FK target.
-- =============================================================================
INSERT INTO stores (slug, name, wc_url)
VALUES ('lizzon', 'Lizzon', 'https://lizzon.com.br')
ON CONFLICT (slug) DO NOTHING;

COMMIT;
