-- 0008: notification_triggers persists the per-event automation rules
-- the worker consults when a shipment crosses a milestone (postado,
-- in_transit, delivered, breached). One row per (store, event_key);
-- the worker fires once per shipment+event and records it in the
-- notifications table for de-dup.

CREATE TABLE IF NOT EXISTS notification_triggers (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT NOT NULL,
    event_key   TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    template    TEXT NOT NULL DEFAULT '',
    session     TEXT NOT NULL DEFAULT '',
    cooldown_minutes INT NOT NULL DEFAULT 60,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id, event_key)
);

-- Seed the four canonical events so the UI has rows to render even
-- before the user saves anything (idempotent: ON CONFLICT DO NOTHING).
INSERT INTO notification_triggers (store_id, event_key, enabled, template)
VALUES
    (1, 'postado',    FALSE,
     'Olá {first_name}, seu pedido #{order_id} foi postado! 📦' || E'\n\n' ||
     'Código: {tracking}' || E'\n' ||
     'Acompanhe: {track_url}'),
    (1, 'in_transit', FALSE,
     'Oi {first_name}, seu pedido #{order_id} está a caminho 🚛' || E'\n\n' ||
     'Último evento: {last_event}' || E'\n' ||
     'Previsão: {eta}' || E'\n' ||
     'Rastreie: {track_url}'),
    (1, 'delivered',  FALSE,
     '{first_name}, seu pedido #{order_id} foi entregue ✅' || E'\n\n' ||
     'Esperamos que você ame seus produtos. Conta com a gente!'),
    (1, 'breached',   FALSE,
     'Oi {first_name}, queria te avisar que seu pedido #{order_id} está com atraso na transportadora.' || E'\n\n' ||
     'Último evento: {last_event}' || E'\n\n' ||
     'Já estamos em contato com a transportadora pra agilizar.')
ON CONFLICT (store_id, event_key) DO NOTHING;
