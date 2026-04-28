-- Univer Tracker — initial schema
-- Created: 2026-04-28

BEGIN;

-- =============================================================================
-- Stores: support multi-tenant from day one (Lizzon, UniverBeauty, UniverSkin).
-- =============================================================================
CREATE TABLE stores (
    id              BIGSERIAL PRIMARY KEY,
    slug            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    wc_url          TEXT NOT NULL,
    wc_consumer_key TEXT,
    wc_secret       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- Orders: mirror of WooCommerce orders we care about (post-payment).
-- =============================================================================
CREATE TABLE orders (
    id                BIGSERIAL PRIMARY KEY,
    store_id          BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    wc_order_id       BIGINT NOT NULL,
    status            TEXT NOT NULL,
    customer_name     TEXT NOT NULL DEFAULT '',
    customer_email    TEXT NOT NULL DEFAULT '',
    customer_phone    TEXT NOT NULL DEFAULT '',
    customer_city     TEXT NOT NULL DEFAULT '',
    customer_uf       TEXT NOT NULL DEFAULT '',
    total_brl         NUMERIC(12, 2) NOT NULL DEFAULT 0,
    paid_at           TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id, wc_order_id)
);

CREATE INDEX idx_orders_status      ON orders (store_id, status);
CREATE INDEX idx_orders_paid_at     ON orders (store_id, paid_at DESC);

-- =============================================================================
-- Shipments: a tracked shipment, one per order (typically).
-- =============================================================================
CREATE TABLE shipments (
    id                  BIGSERIAL PRIMARY KEY,
    order_id            BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tracking_code       TEXT NOT NULL,
    carrier             TEXT NOT NULL DEFAULT '',
    service             TEXT NOT NULL DEFAULT '',
    service_code        TEXT NOT NULL DEFAULT '',
    tracking_url        TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'created',
    last_event          TEXT NOT NULL DEFAULT '',
    last_event_at       TIMESTAMPTZ,
    estimated_delivery  DATE,
    delivered_at        TIMESTAMPTZ,
    last_synced_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_id, tracking_code)
);

CREATE INDEX idx_shipments_status         ON shipments (status);
CREATE INDEX idx_shipments_last_synced_at ON shipments (last_synced_at NULLS FIRST);
CREATE INDEX idx_shipments_tracking_code  ON shipments (tracking_code);

-- =============================================================================
-- Tracking events: every status update from the carrier.
-- =============================================================================
CREATE TABLE tracking_events (
    id              BIGSERIAL PRIMARY KEY,
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    occurred_at     TIMESTAMPTZ NOT NULL,
    type            TEXT NOT NULL DEFAULT 'unknown',
    description     TEXT NOT NULL,
    location        TEXT NOT NULL DEFAULT '',
    raw             JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shipment_id, occurred_at, description)
);

CREATE INDEX idx_tracking_events_shipment ON tracking_events (shipment_id, occurred_at DESC);
CREATE INDEX idx_tracking_events_type     ON tracking_events (type);

-- =============================================================================
-- Webhook deliveries: idempotency + audit trail for incoming events.
-- =============================================================================
CREATE TABLE webhook_deliveries (
    id              BIGSERIAL PRIMARY KEY,
    source          TEXT NOT NULL,
    event_id        TEXT NOT NULL,
    payload         JSONB NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    error           TEXT,
    UNIQUE (source, event_id)
);

CREATE INDEX idx_webhook_deliveries_received ON webhook_deliveries (received_at DESC);

COMMIT;
