import type { OrderDetail, OrdersResponse, OverviewResponse } from "./types";

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

function url(path: string, params?: Record<string, string | number | undefined>): string {
  const u = new URL(path, baseURL());
  for (const [k, v] of Object.entries(params ?? {})) {
    if (v !== undefined && v !== "") u.searchParams.set(k, String(v));
  }
  return u.toString();
}

export async function fetchOrders(params: {
  status?: string;
  health?: string;
  q?: string;
  page?: number;
  per_page?: number;
} = {}): Promise<OrdersResponse> {
  const res = await fetch(url("/api/v1/orders", params), { cache: "no-store" });
  if (!res.ok) throw new Error(`orders fetch failed: ${res.status}`);
  return res.json();
}

export async function fetchOverview(): Promise<OverviewResponse> {
  const res = await fetch(url("/api/v1/analytics/overview"), { cache: "no-store" });
  if (!res.ok) throw new Error(`overview fetch failed: ${res.status}`);
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
