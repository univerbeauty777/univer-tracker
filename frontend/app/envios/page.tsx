import { Card, CardContent } from "@/components/ui/card";
import { ExportButton } from "@/components/export-button";
import { OrdersFilterBar } from "@/components/orders-filter-bar";
import { OrdersTable } from "@/components/orders-table";
import { fetchFacets, fetchOrders } from "@/lib/api";
import type { Facets, OrderListItem } from "@/lib/types";

export const dynamic = "force-dynamic";

const PER_PAGE = 50;

type SP = Promise<Record<string, string | string[] | undefined>>;

export default async function EnviosPage({ searchParams }: { searchParams: SP }) {
  const sp = await searchParams;
  const get = (k: string) => (typeof sp[k] === "string" ? (sp[k] as string) : undefined);
  const offset = Number(get("offset") ?? 0) || 0;

  const query = {
    status: get("status"),
    health: get("health"),
    carrier: get("carrier"),
    uf: get("uf"),
    q: get("q"),
    since: get("since"),
    until: get("until"),
    sort: (get("sort") as "created_at" | "total" | "customer_name" | "last_event") ?? "created_at",
    dir: (get("dir") as "asc" | "desc") ?? "desc",
    per_page: PER_PAGE,
    offset,
  };

  let orders: OrderListItem[] = [];
  let total = 0;
  let facets: Facets = { carriers: [], ufs: [], statuses: [], health: [] };
  let err: string | null = null;

  try {
    const [list, fc] = await Promise.all([fetchOrders(query), fetchFacets()]);
    orders = list.orders ?? [];
    total = list.total;
    facets = fc;
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-display text-2xl font-semibold text-zinc-100">Envios</h1>
          <p className="mt-1 text-sm text-zinc-500">
            {total} envios sincronizados · ordenados por urgência de SLA
          </p>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <OrdersFilterBar facets={facets} />
        <ExportButton />
      </div>

      {err ? (
        <Card>
          <CardContent className="space-y-1 p-6 text-sm">
            <div className="font-medium text-rose-400">Não foi possível carregar pedidos.</div>
            <div className="text-zinc-500">{err}</div>
          </CardContent>
        </Card>
      ) : (
        <OrdersTable orders={orders} total={total} limit={query.per_page} offset={query.offset} />
      )}
    </div>
  );
}
