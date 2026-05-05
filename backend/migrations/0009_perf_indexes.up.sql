-- Critical indexes the analytics, listing and transitions endpoints
-- depend on. Without these the dashboard ran sequential scans on the
-- shipments + orders tables every request, holding pgx pool connections
-- long enough for Coolify's reverse proxy to give up with 504.
--
-- All created CONCURRENTLY (when supported) so a re-deploy doesn't
-- exclusive-lock the table on a live database. CONCURRENTLY can't run
-- inside a transaction; the migrate runner already applies each file
-- as its own statement so we're safe here.

-- Orders listing pagination is keyed on (store_id, created_at DESC) and
-- almost every analytics filter joins by store_id then dates back N days.
CREATE INDEX IF NOT EXISTS idx_orders_store_created
    ON orders (store_id, created_at DESC);

-- Carrier filters on the orders table use the WC carrier slug we sync
-- into shipping_method. Frequent in /analytics/funnel and bulk hide.
CREATE INDEX IF NOT EXISTS idx_orders_store_status_created
    ON orders (store_id, status, created_at DESC);

-- shipments.created_at drives every windowed analytics query. The bare
-- index is enough; Postgres uses it in combination with idx_shipments_*
-- via bitmap scans.
CREATE INDEX IF NOT EXISTS idx_shipments_created_at
    ON shipments (created_at DESC);

-- Carrier analytics group by carrier and bucket by created_at. Composite
-- index lets the planner skip the whole-table scan for the GROUP BY.
CREATE INDEX IF NOT EXISTS idx_shipments_carrier_created
    ON shipments (carrier, created_at DESC);

-- order_id is the FK shipments → orders. Postgres does NOT auto-index
-- foreign keys, and store/orders.go joins on it heavily. Without this
-- ON DELETE CASCADE also seq-scans on every parent delete.
CREATE INDEX IF NOT EXISTS idx_shipments_order_id
    ON shipments (order_id);

-- Stage-timestamp partial indexes. The transitions analytics filters
-- WHERE <stage>_at IS NOT NULL — partial indexes give us O(log n)
-- access to the small subset of rows that have actually reached the
-- stage. Cheaper than a full b-tree on a sparse column.
CREATE INDEX IF NOT EXISTS idx_shipments_label_issued_at
    ON shipments (created_at) WHERE label_issued_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_shipments_preparing_at
    ON shipments (created_at) WHERE preparing_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_shipments_ready_for_pickup_at
    ON shipments (created_at) WHERE ready_for_pickup_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_shipments_out_for_delivery_at
    ON shipments (created_at) WHERE out_for_delivery_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_shipments_delivered_at
    ON shipments (created_at) WHERE delivered_at IS NOT NULL;
