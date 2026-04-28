-- 0002 — sync state + shipment health/idle tracking
-- Created: 2026-04-28

BEGIN;

-- =============================================================================
-- sync_state: tracks last successful sync per entity so the worker is
-- restart-safe and can do incremental fetches.
-- =============================================================================
CREATE TABLE sync_state (
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
    ADD COLUMN health        TEXT        NOT NULL DEFAULT 'unknown',
    ADD COLUMN idle_since    TIMESTAMPTZ,
    ADD COLUMN risk_score    SMALLINT    NOT NULL DEFAULT 0;

CREATE INDEX idx_shipments_health     ON shipments (health);
CREATE INDEX idx_shipments_idle_since ON shipments (idle_since NULLS LAST);

-- =============================================================================
-- orders: cache the WC line totals so the table view doesn't need to join
-- a second time, and the wc_status mirror so analytics can group fast.
-- =============================================================================
ALTER TABLE orders
    ADD COLUMN customer_phone   TEXT NOT NULL DEFAULT '',
    ADD COLUMN shipping_method  TEXT NOT NULL DEFAULT '';

-- =============================================================================
-- Default store (Lizzon). Multi-tenant comes later; for now everything we
-- sync is rooted here so the worker has a known FK target.
-- =============================================================================
INSERT INTO stores (slug, name, wc_url)
VALUES ('lizzon', 'Lizzon', 'https://lizzon.com.br')
ON CONFLICT (slug) DO NOTHING;

COMMIT;
