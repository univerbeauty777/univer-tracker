import type {
  Facets,
  FrenetIntegration,
  IntegrationsResponse,
  OrderDetail,
  OrdersResponse,
  OverviewResponse,
  TestResult,
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

// Browser fetch goes through the public API hostname (CORS + Traefik).
// Server-side fetch (RSC, route handlers) prefers the Docker-internal
// hostname when set — faster, no public DNS dependency, no TLS round
// trip — and falls back to the public URL otherwise.
function baseURL(): string {
  if (typeof window === "undefined") {
    return (
      process.env.INTERNAL_API_URL ??
      process.env.NEXT_PUBLIC_API_URL ??
      "http://localhost:8080"
    );
  }
  return process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
}

function url(path: string, params?: Record<string, unknown>): string {
  const u = new URL(path, baseURL());
  for (const [k, v] of Object.entries(params ?? {})) {
    if (v === undefined || v === null || v === "") continue;
    u.searchParams.set(k, String(v));
  }
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
