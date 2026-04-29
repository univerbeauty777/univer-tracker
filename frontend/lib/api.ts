import type {
  BreakdownResponse,
  Facets,
  FrenetIntegration,
  FunnelResponse,
  IntegrationsResponse,
  OrderDetail,
  OrderHistoryResponse,
  OrdersResponse,
  OverviewResponse,
  SyncStatusResponse,
  TestResult,
  TransitionsResponse,
  WAHAIntegration,
  WooCommerceIntegration,
} from "./types";

export interface OrdersQuery {
  status?: string;
  health?: string;
  carrier?: string;
  uf?: string;
  q?: string;
  since?: string;
  until?: string;
  sort?: "created_at" | "total" | "customer_name" | "last_event";
  dir?: "asc" | "desc";
  per_page?: number;
  offset?: number;
}

// Browser fetch always uses a relative URL — Next rewrites /api/* to the
// internal backend, so the client never needs to know the hostname.
// Server-side fetch (RSC) goes direct to the docker network for speed.
function baseURL(): string {
  if (typeof window === "undefined") {
    return (
      process.env.INTERNAL_API_URL ??
      process.env.NEXT_PUBLIC_API_URL ??
      "http://backend:8080"
    );
  }
  return ""; // relative URL; Next rewrites /api/* → backend
}

function url(path: string, params?: Record<string, unknown>): string {
  const base = baseURL();
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params ?? {})) {
    if (v === undefined || v === null || v === "") continue;
    sp.set(k, String(v));
  }
  const qs = sp.toString();
  if (base === "") {
    // Browser: relative URL
    return qs ? `${path}?${qs}` : path;
  }
  const u = new URL(path, base);
  for (const [k, v] of sp.entries()) u.searchParams.set(k, v);
  return u.toString();
}

export async function fetchOrders(params: OrdersQuery = {}): Promise<OrdersResponse> {
  const res = await fetch(url("/api/v1/orders", { ...params }), { cache: "no-store" });
  if (!res.ok) throw new Error(`orders fetch failed: ${res.status}`);
  return res.json();
}

export async function fetchFacets(): Promise<Facets> {
  const res = await fetch(url("/api/v1/orders/facets"), { cache: "no-store" });
  if (!res.ok) throw new Error(`facets fetch failed: ${res.status}`);
  return res.json();
}

export async function fetchSyncStatus(): Promise<SyncStatusResponse> {
  const res = await fetch(url("/api/v1/sync/status"), { cache: "no-store" });
  if (!res.ok) throw new Error(`sync status failed: ${res.status}`);
  return res.json();
}

export async function triggerSync(): Promise<void> {
  const res = await fetch(url("/api/v1/sync/run"), { method: "POST" });
  if (!res.ok) throw new Error(`sync trigger failed: ${res.status}`);
}

export async function fetchOrderHistory(id: number): Promise<OrderHistoryResponse> {
  const res = await fetch(url(`/api/v1/orders/${id}/history`), { cache: "no-store" });
  if (!res.ok) throw new Error(`history fetch failed: ${res.status}`);
  return res.json();
}

export async function notifyOrder(
  id: number,
  message: string,
  template?: string,
): Promise<{ ok: boolean; message?: string; error?: string }> {
  const res = await fetch(url(`/api/v1/orders/${id}/notify`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message, template }),
  });
  return res.json();
}

export function ordersExportURL(params: Record<string, string | undefined>): string {
  return url("/api/v1/orders/export.csv", params);
}

export async function fetchFunnel(days = 30): Promise<FunnelResponse> {
  const res = await fetch(url("/api/v1/analytics/funnel", { days }), { cache: "no-store" });
  if (!res.ok) throw new Error(`funnel ${res.status}`);
  return res.json();
}

export async function fetchTransitions(days = 30): Promise<TransitionsResponse> {
  const res = await fetch(url("/api/v1/analytics/transitions", { days }), { cache: "no-store" });
  if (!res.ok) throw new Error(`transitions ${res.status}`);
  return res.json();
}

export async function fetchBreakdown(orderId: number): Promise<BreakdownResponse> {
  const res = await fetch(url(`/api/v1/orders/${orderId}/breakdown`), { cache: "no-store" });
  if (!res.ok) throw new Error(`breakdown ${res.status}`);
  return res.json();
}

export async function fetchOverview(): Promise<OverviewResponse> {
  const res = await fetch(url("/api/v1/analytics/overview"), { cache: "no-store" });
  if (!res.ok) throw new Error(`overview fetch failed: ${res.status}`);
  return res.json();
}

export async function fetchIntegrations(): Promise<IntegrationsResponse> {
  const res = await fetch(url("/api/v1/settings/integrations"), { cache: "no-store" });
  if (!res.ok) throw new Error(`integrations fetch failed: ${res.status}`);
  return res.json();
}

export async function updateWooCommerceIntegration(
  body: Partial<WooCommerceIntegration>,
): Promise<IntegrationsResponse> {
  const res = await fetch(url("/api/v1/settings/integrations/woocommerce"), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`save failed: ${res.status}`);
  return res.json();
}

export async function updateFrenetIntegration(
  body: Partial<FrenetIntegration>,
): Promise<IntegrationsResponse> {
  const res = await fetch(url("/api/v1/settings/integrations/frenet"), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`save failed: ${res.status}`);
  return res.json();
}

export async function updateWAHAIntegration(
  body: Partial<WAHAIntegration>,
): Promise<IntegrationsResponse> {
  const res = await fetch(url("/api/v1/settings/integrations/waha"), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`save failed: ${res.status}`);
  return res.json();
}

export async function testIntegration(
  provider: "woocommerce" | "frenet" | "waha",
): Promise<TestResult> {
  const res = await fetch(url(`/api/v1/settings/integrations/${provider}/test`), {
    method: "POST",
  });
  if (!res.ok) throw new Error(`test failed: ${res.status}`);
  return res.json();
}

export async function fetchOrder(id: number): Promise<OrderDetail> {
  const res = await fetch(url(`/api/v1/orders/${id}`), { cache: "no-store" });
  if (!res.ok) throw new Error(`order ${id} fetch failed: ${res.status}`);
  return res.json();
}

export async function updateOrderStatus(
  id: number,
  status: string,
  note?: string,
): Promise<OrderDetail> {
  const res = await fetch(url(`/api/v1/orders/${id}/status`), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status, note }),
  });
  if (!res.ok) throw new Error(`status update failed: ${res.status}`);
  return res.json();
}
