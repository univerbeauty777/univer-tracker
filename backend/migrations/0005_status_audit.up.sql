-- 0005 — status change audit log
-- Created: 2026-04-28
--
-- Every WC status transition (whether triggered by the dashboard, a
-- webhook or the sync worker) lands here so the order detail page can
-- show "quem mudou de pra quando, com qual nota". It also future-proofs
-- the SLA dashboard's "tempo médio em cada etapa" without polling WC.

BEGIN;

CREATE TABLE IF NOT EXISTS status_changes (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status  TEXT NOT NULL DEFAULT '',
    to_status    TEXT NOT NULL,
    source       TEXT NOT NULL DEFAULT 'manual', -- manual | sync | webhook
    note         TEXT NOT NULL DEFAULT '',
    actor        TEXT NOT NULL DEFAULT 'system',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_status_changes_order ON status_changes (order_id, created_at DESC);

-- Per-shipment notification audit (so we don't spam customers and the
-- detail page can show 'cliente notificado da postagem em hh:mm').
CREATE TABLE IF NOT EXISTS notifications (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    channel      TEXT NOT NULL, -- waha | email | sms
    template     TEXT NOT NULL DEFAULT '',
    payload      JSONB NOT NULL DEFAULT '{}'::jsonb,
    status       TEXT NOT NULL DEFAULT 'sent', -- sent | failed
    error        TEXT NOT NULL DEFAULT '',
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_order ON notifications (order_id, sent_at DESC);

COMMIT;
