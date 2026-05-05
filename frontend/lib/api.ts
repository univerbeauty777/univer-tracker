import type {
  BreakdownResponse,
  Facets,
  FrenetIntegration,
  FunnelResponse,
  IntegrationsResponse,
  NotificationTrigger,
  OrderDetail,
  OrderHistoryResponse,
  OrdersResponse,
  OverviewResponse,
  SyncStatusResponse,
  TestResult,
  TransitionsResponse,
  TriggersResponse,
  WAHAIntegration,
  WAHASessionsResponse,
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
    return qs ? `${path}?${qs}` : path;
  }
  const u = new URL(path, base);
  for (const [k, v] of sp.entries()) u.searchParams.set(k, v);
  return u.toString();
}

// Default upper bound on a single API call. The backend itself enforces
// 25s; we add ~5s headroom so the network error surfaces here as a
// proper timeout instead of a generic failed fetch when the backend is
// stuck. Long endpoints (CSV export) bypass this helper.
const DEFAULT_TIMEOUT_MS = 30_000;

interface ApiFetchOptions extends RequestInit {
  timeoutMs?: number;
}

async function apiFetch(input: string, init: ApiFetchOptions = {}): Promise<Response> {
  const { timeoutMs = DEFAULT_TIMEOUT_MS, signal: externalSignal, ...rest } = init;

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(new Error("timeout")), timeoutMs);

  // If the caller passed their own signal (component unmount, etc.) wire
  // it through so an abort there cancels our fetch too.
  if (externalSignal) {
    if (externalSignal.aborted) {
      controller.abort(externalSignal.reason);
    } else {
      externalSignal.addEventListener("abort", () => controller.abort(externalSignal.reason), {
        once: true,
      });
    }
  }

  try {
    return await fetch(input, { cache: "no-store", ...rest, signal: controller.signal });
  } finally {
    clearTimeout(timer);
  }
}

async function getJSON<T>(path: string, params?: object, init?: ApiFetchOptions): Promise<T> {
  const res = await apiFetch(url(path, params as Record<string, unknown> | undefined), init);
  if (!res.ok) throw new Error(`${path} ${res.status}`);
  return res.json() as Promise<T>;
}

export async function fetchOrders(params: OrdersQuery = {}, init?: ApiFetchOptions): Promise<OrdersResponse> {
  return getJSON("/api/v1/orders", params, init);
}

export async function fetchFacets(init?: ApiFetchOptions): Promise<Facets> {
  return getJSON("/api/v1/orders/facets", undefined, init);
}

export async function fetchSyncStatus(init?: ApiFetchOptions): Promise<SyncStatusResponse> {
  return getJSON("/api/v1/sync/status", undefined, init);
}

export async function triggerSync(init?: ApiFetchOptions): Promise<void> {
  const res = await apiFetch(url("/api/v1/sync/run"), { method: "POST", ...init });
  if (!res.ok) throw new Error(`sync trigger failed: ${res.status}`);
}

export async function fetchOrderHistory(id: number, init?: ApiFetchOptions): Promise<OrderHistoryResponse> {
  return getJSON(`/api/v1/orders/${id}/history`, undefined, init);
}

export async function notifyOrder(
  id: number,
  message: string,
  template?: string,
  session?: string,
): Promise<{ ok: boolean; message?: string; error?: string }> {
  const res = await apiFetch(url(`/api/v1/orders/${id}/notify`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message, template, session }),
  });
  return res.json();
}

export function ordersExportURL(params: Record<string, string | undefined>): string {
  return url("/api/v1/orders/export.csv", params);
}

export async function bulkHideOrders(ids: number[]): Promise<{ hidden: number; requested: number }> {
  const res = await apiFetch(url("/api/v1/orders/bulk-hide"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ids }),
  });
  if (!res.ok) {
    const txt = await res.text().catch(() => "");
    throw new Error(`bulk hide failed: ${res.status} ${txt}`);
  }
  return res.json();
}

export async function fetchWAHASessions(init?: ApiFetchOptions): Promise<WAHASessionsResponse> {
  return getJSON("/api/v1/settings/integrations/waha/sessions", undefined, init);
}

export async function fetchTriggers(init?: ApiFetchOptions): Promise<TriggersResponse> {
  return getJSON("/api/v1/settings/triggers", undefined, init);
}

export async function saveTriggers(
  triggers: NotificationTrigger[],
): Promise<TriggersResponse> {
  const res = await apiFetch(url("/api/v1/settings/triggers"), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ triggers }),
  });
  if (!res.ok) {
    const txt = await res.text().catch(() => "");
    throw new Error(`triggers save failed: ${res.status} ${txt}`);
  }
  return res.json();
}

export async function fetchFunnel(days = 30, init?: ApiFetchOptions): Promise<FunnelResponse> {
  return getJSON("/api/v1/analytics/funnel", { days }, init);
}

export async function fetchTransitions(days = 30, init?: ApiFetchOptions): Promise<TransitionsResponse> {
  return getJSON("/api/v1/analytics/transitions", { days }, init);
}

export async function fetchBreakdown(orderId: number, init?: ApiFetchOptions): Promise<BreakdownResponse> {
  return getJSON(`/api/v1/orders/${orderId}/breakdown`, undefined, init);
}

export async function fetchOverview(init?: ApiFetchOptions): Promise<OverviewResponse> {
  return getJSON("/api/v1/analytics/overview", undefined, init);
}

export async function fetchIntegrations(init?: ApiFetchOptions): Promise<IntegrationsResponse> {
  return getJSON("/api/v1/settings/integrations", undefined, init);
}

export async function updateWooCommerceIntegration(
  body: Partial<WooCommerceIntegration>,
): Promise<IntegrationsResponse> {
  const res = await apiFetch(url("/api/v1/settings/integrations/woocommerce"), {
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
  const res = await apiFetch(url("/api/v1/settings/integrations/frenet"), {
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
  const res = await apiFetch(url("/api/v1/settings/integrations/waha"), {
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
  const res = await apiFetch(url(`/api/v1/settings/integrations/${provider}/test`), {
    method: "POST",
    timeoutMs: 15_000,
  });
  if (!res.ok) throw new Error(`test failed: ${res.status}`);
  return res.json();
}

export async function fetchOrder(id: number, init?: ApiFetchOptions): Promise<OrderDetail> {
  return getJSON(`/api/v1/orders/${id}`, undefined, init);
}

export async function updateOrderStatus(
  id: number,
  status: string,
  note?: string,
): Promise<OrderDetail> {
  const res = await apiFetch(url(`/api/v1/orders/${id}/status`), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status, note }),
  });
  if (!res.ok) throw new Error(`status update failed: ${res.status}`);
  return res.json();
}
