-- 0004 — recompute SLA basis for shipments synced before the bugfix
-- Created: 2026-04-28
--
-- The first sync run was using the WooCommerce order's created_at as the
-- shipment's start date, which made every order older than the SLA window
-- appear as "breached" the moment we synced it for the first time.
--
-- This migration resets the SLA inputs for non-delivered shipments so the
-- next worker tick (or the inline Compute call after sync) recomputes
-- health from a sane baseline (now). Delivered shipments are left alone —
-- their history is real.

BEGIN;

UPDATE shipments
SET
    created_at         = NOW(),
    health             = 'unknown',
    idle_since         = NULL,
    risk_score         = 0,
    estimated_delivery = NULL
WHERE delivered_at IS NULL;

COMMIT;
