-- 0003 — application settings
-- Created: 2026-04-28
--
-- Holds dynamic configuration the dashboard can edit at runtime
-- (integrations: WooCommerce, Frenet, WAHA). Keys follow the
-- "domain.subdomain" convention so we can list/group from SQL.

BEGIN;

CREATE TABLE IF NOT EXISTS app_settings (
    key         TEXT PRIMARY KEY,
    value       JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO app_settings (key, value) VALUES
    ('integration.woocommerce', '{}'::jsonb),
    ('integration.frenet',      '{}'::jsonb),
    ('integration.waha',        '{}'::jsonb)
ON CONFLICT (key) DO NOTHING;

COMMIT;
