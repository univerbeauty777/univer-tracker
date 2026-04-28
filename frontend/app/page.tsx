import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  Package,
  Timer,
  TrendingUp,
  Truck,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { OrdersFilterBar } from "@/components/orders-filter-bar";
import { OrdersTable } from "@/components/orders-table";
import { fetchFacets, fetchOrders, fetchOverview } from "@/lib/api";
import type { CarrierStats, Facets, OrderListItem, Overview } from "@/lib/types";

export const dynamic = "force-dynamic";

const PER_PAGE = 50;

type SP = Promise<Record<string, string | string[] | undefined>>;

export default async function DashboardPage({ searchParams }: { searchParams: SP }) {
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

  let overview: Overview | null = null;
  let carriers: CarrierStats[] = [];
  let orders: OrderListItem[] = [];
  let total = 0;
  let facets: Facets = { carriers: [], ufs: [], statuses: [], health: [] };
  let err: string | null = null;

  try {
    const [ov, list, fc] = await Promise.all([
      fetchOverview(),
      fetchOrders(query),
      fetchFacets(),
    ]);
    overview = ov.overview;
    carriers = ov.carriers ?? [];
    orders = list.orders ?? [];
    total = list.total;
    facets = fc;
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="font-display text-3xl font-semibold tracking-tight">Painel logístico</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Pedidos do WooCommerce com SLA, saúde de entrega e timeline da Frenet.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
        <Kpi label="Pedidos (30d)" value={overview?.total_30d ?? 0} icon={Package} tone="text-foreground" />
        <Kpi
          label="No prazo"
          value={`${Math.round((overview?.on_time_rate ?? 0) * 100)}%`}
          hint={`${overview?.on_time_30d ?? 0}/${overview?.delivered_30d ?? 0} entregues`}
          icon={TrendingUp}
          tone="text-success"
        />
        <Kpi
          label="Tempo médio"
          value={
            overview && overview.avg_delivery_days > 0
              ? `${overview.avg_delivery_days.toFixed(1)}d`
              : "—"
          }
          hint={overview && overview.avg_delivery_days > 0 ? "do pago à entrega" : "sem entregas concluídas"}
          icon={Timer}
          tone="text-info"
        />
        <Kpi label="Em risco" value={overview?.at_risk ?? 0} icon={AlertTriangle} tone="text-warning" />
        <Kpi label="SLA quebrado" value={overview?.breached ?? 0} icon={AlertTriangle} tone="text-destructive" />
        <Kpi label="Sem evento >4d" value={overview?.idle_alarms ?? 0} icon={Clock} tone="text-warning" />
      </div>

      {carriers.length > 0 ? (
        <Card>
          <CardContent className="px-0 pb-0 pt-4">
            <div className="flex items-center justify-between px-5 pb-3">
              <div>
                <div className="font-display text-base font-semibold">Transportadoras</div>
                <div className="text-xs text-muted-foreground">Volume e quebras de SLA nos últimos 30 dias.</div>
              </div>
              <Truck className="size-4 text-muted-foreground" />
            </div>
            <table className="w-full text-sm">
              <thead>
                <tr className="border-y border-border/60 text-[11px] uppercase tracking-wide text-muted-foreground">
                  <th className="px-5 py-2 text-left font-medium">Transportadora</th>
                  <th className="px-5 py-2 text-right font-medium">Volume</th>
                  <th className="px-5 py-2 text-right font-medium">SLA quebrado</th>
                  <th className="px-5 py-2 text-right font-medium">Tempo médio</th>
                </tr>
              </thead>
              <tbody>
                {carriers.map((c) => (
                  <tr key={c.carrier} className="border-b border-border/40 last:border-0">
                    <td className="px-5 py-2.5">{c.carrier}</td>
                    <td className="px-5 py-2.5 text-right font-medium">{c.total}</td>
                    <td className="px-5 py-2.5 text-right">
                      {c.breached > 0 ? (
                        <span className="text-destructive">
                          {c.breached} ({Math.round((c.breached / c.total) * 100)}%)
                        </span>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </td>
                    <td className="px-5 py-2.5 text-right text-muted-foreground">
                      {c.avg_delivery_days > 0 ? `${c.avg_delivery_days.toFixed(1)}d` : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      ) : null}

      <OrdersFilterBar facets={facets} />

      {err ? (
        <Card>
          <CardContent className="space-y-1 p-6 text-sm">
            <div className="font-medium text-destructive">Não foi possível carregar pedidos.</div>
            <div className="text-muted-foreground">{err}</div>
          </CardContent>
        </Card>
      ) : (
        <OrdersTable orders={orders} total={total} limit={query.per_page} offset={query.offset} />
      )}
    </div>
  );
}

function Kpi({
  label,
  value,
  hint,
  icon: Icon,
  tone,
}: {
  label: string;
  value: string | number;
  hint?: string;
  icon: React.ComponentType<{ className?: string; strokeWidth?: number }>;
  tone: string;
}) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex size-9 items-center justify-center rounded-lg bg-secondary">
          <Icon className={`size-[18px] ${tone}`} strokeWidth={2} />
        </div>
        <div className="mt-3">
          <div className={`font-display text-2xl font-semibold ${tone}`}>{value}</div>
          <div className="mt-0.5 text-xs text-muted-foreground">{label}</div>
          {hint ? <div className="mt-0.5 text-[11px] text-muted-foreground">{hint}</div> : null}
        </div>
      </CardContent>
    </Card>
  );
}
