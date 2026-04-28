import Link from "next/link";
import {
  AlertTriangle,
  ChevronRight,
  Clock,
  Package,
  Timer,
  TrendingUp,
  Truck,
} from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/status-badge";
import { HealthBadge } from "@/components/health-badge";
import { fetchOrders, fetchOverview } from "@/lib/api";
import { formatBRL, formatDate, formatDateTime } from "@/lib/format";
import type {
  CarrierStats,
  OrderListItem,
  Overview,
  Health,
} from "@/lib/types";

export const dynamic = "force-dynamic";

const HEALTH_FILTERS: { key: string; label: string; param?: string }[] = [
  { key: "all", label: "Todos" },
  { key: "at_risk", label: "Em risco", param: "at_risk" },
  { key: "breached", label: "SLA quebrado", param: "breached" },
  { key: "on_track", label: "No prazo", param: "on_track" },
];

export default async function DashboardPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const sp = await searchParams;
  const healthRaw = typeof sp.health === "string" ? sp.health : "";
  const search = typeof sp.q === "string" ? sp.q : "";

  let overview: Overview | null = null;
  let carriers: CarrierStats[] = [];
  let orders: OrderListItem[] = [];
  let err: string | null = null;

  try {
    const [ov, list] = await Promise.all([
      fetchOverview(),
      fetchOrders({
        status: "processing,on-hold,shipped,in-transit,out-for-delivery,completed",
        health: healthRaw || undefined,
        q: search || undefined,
        per_page: 100,
      }),
    ]);
    overview = ov.overview;
    carriers = ov.carriers ?? [];
    orders = list.orders;
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-display text-3xl font-semibold tracking-tight">
            Painel logístico
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Pedidos do WooCommerce com SLA, saúde de entrega e timeline da Frenet.
          </p>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
        <Kpi
          label="Pedidos (30d)"
          value={overview?.total_30d ?? 0}
          icon={Package}
          tone="text-foreground"
        />
        <Kpi
          label="No prazo"
          value={`${Math.round((overview?.on_time_rate ?? 0) * 100)}%`}
          hint={`${overview?.on_time_30d ?? 0}/${overview?.delivered_30d ?? 0}`}
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
          icon={Timer}
          tone="text-info"
        />
        <Kpi
          label="Em risco"
          value={overview?.at_risk ?? 0}
          icon={AlertTriangle}
          tone="text-warning"
        />
        <Kpi
          label="SLA quebrado"
          value={overview?.breached ?? 0}
          icon={AlertTriangle}
          tone="text-destructive"
        />
        <Kpi
          label="Sem evento >4d"
          value={overview?.idle_alarms ?? 0}
          icon={Clock}
          tone="text-warning"
        />
      </div>

      {/* Carriers ranking */}
      {carriers.length > 0 ? (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
            <div>
              <CardTitle className="text-base">Transportadoras (30 dias)</CardTitle>
              <CardDescription>Volume e quebras de SLA por transportadora.</CardDescription>
            </div>
            <Truck className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="px-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/60 text-xs uppercase tracking-wide text-muted-foreground">
                  <th className="px-6 py-2 text-left font-medium">Transportadora</th>
                  <th className="px-6 py-2 text-right font-medium">Volume</th>
                  <th className="px-6 py-2 text-right font-medium">SLA quebrado</th>
                  <th className="px-6 py-2 text-right font-medium">Tempo médio</th>
                </tr>
              </thead>
              <tbody>
                {carriers.map((c) => (
                  <tr key={c.carrier} className="border-b border-border/40 last:border-0">
                    <td className="px-6 py-2 capitalize">{c.carrier}</td>
                    <td className="px-6 py-2 text-right font-medium">{c.total}</td>
                    <td className="px-6 py-2 text-right text-destructive">
                      {c.breached > 0
                        ? `${c.breached} (${Math.round((c.breached / c.total) * 100)}%)`
                        : "—"}
                    </td>
                    <td className="px-6 py-2 text-right text-muted-foreground">
                      {c.avg_delivery_days > 0 ? `${c.avg_delivery_days.toFixed(1)}d` : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      ) : null}

      {/* Filtros */}
      <div className="flex flex-wrap items-center gap-2">
        {HEALTH_FILTERS.map((f) => {
          const active = (f.param ?? "") === healthRaw;
          const href = f.param
            ? `/?health=${f.param}${search ? `&q=${encodeURIComponent(search)}` : ""}`
            : `/${search ? `?q=${encodeURIComponent(search)}` : ""}`;
          return (
            <Link
              key={f.key}
              href={href}
              className={
                active
                  ? "inline-flex items-center rounded-full bg-foreground px-3 py-1 text-xs font-medium text-background"
                  : "inline-flex items-center rounded-full border border-border/60 px-3 py-1 text-xs text-muted-foreground hover:bg-muted"
              }
            >
              {f.label}
            </Link>
          );
        })}
      </div>

      {/* Orders */}
      {err ? (
        <Card>
          <CardContent className="space-y-1 p-6 text-sm">
            <div className="font-medium text-destructive">Não foi possível carregar pedidos.</div>
            <div className="text-muted-foreground">{err}</div>
          </CardContent>
        </Card>
      ) : orders.length === 0 ? (
        <Card>
          <CardContent className="p-10 text-center text-sm text-muted-foreground">
            Nenhum pedido bate com o filtro escolhido.
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="px-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/60 text-xs uppercase tracking-wide text-muted-foreground">
                    <th className="px-6 py-3 text-left font-medium">Pedido</th>
                    <th className="px-6 py-3 text-left font-medium">Cliente</th>
                    <th className="px-6 py-3 text-left font-medium">SLA</th>
                    <th className="px-6 py-3 text-left font-medium">Status</th>
                    <th className="px-6 py-3 text-left font-medium">Último evento</th>
                    <th className="px-6 py-3 text-right font-medium">Total</th>
                    <th className="w-10" />
                  </tr>
                </thead>
                <tbody>
                  {orders.map((o) => (
                    <OrderRow key={o.id} order={o} />
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function OrderRow({ order: o }: { order: OrderListItem }) {
  const idle = idleDays(o.tracking.last_event_at, o.created_at);
  return (
    <tr className="border-b border-border/40 transition-colors last:border-0 hover:bg-muted/30">
      <td className="px-6 py-4">
        <Link
          href={`/pedidos/${o.id}`}
          className="font-mono text-xs font-medium hover:underline"
        >
          #{o.id}
        </Link>
        <div className="text-[11px] text-muted-foreground">{formatDate(o.created_at)}</div>
      </td>
      <td className="px-6 py-4">
        <div className="font-medium">{o.customer_name || "—"}</div>
        <div className="text-[11px] text-muted-foreground">
          {[o.customer_city, o.customer_state].filter(Boolean).join(" · ") || "—"}
        </div>
      </td>
      <td className="px-6 py-4">
        <HealthBadge health={o.tracking.health as Health} label={o.tracking.health_label} />
        {o.tracking.estimated_delivery ? (
          <div className="mt-1 text-[11px] text-muted-foreground">
            ETA · {formatDate(o.tracking.estimated_delivery)}
          </div>
        ) : null}
      </td>
      <td className="px-6 py-4">
        <StatusBadge status={o.status} label={o.status_label} />
        <div className="mt-1 text-[11px] text-muted-foreground">{o.tracking.status_label}</div>
      </td>
      <td className="px-6 py-4">
        {o.tracking.last_event ? (
          <div>
            <div className="line-clamp-1 max-w-[280px] text-xs">{o.tracking.last_event}</div>
            <div className="text-[11px] text-muted-foreground">
              {formatDateTime(o.tracking.last_event_at)}
              {idle >= 4 ? <span className="ml-1 text-warning">· {idle}d sem evento</span> : null}
            </div>
          </div>
        ) : o.tracking.number ? (
          <span className="text-xs text-muted-foreground">aguardando primeira leitura</span>
        ) : (
          <span className="text-xs text-muted-foreground">sem código</span>
        )}
      </td>
      <td className="px-6 py-4 text-right font-medium">{formatBRL(o.total)}</td>
      <td className="px-3 py-4 text-muted-foreground">
        <Link href={`/pedidos/${o.id}`} aria-label={`Abrir pedido ${o.id}`}>
          <ChevronRight className="size-4" />
        </Link>
      </td>
    </tr>
  );
}

function idleDays(lastEventAt?: string, createdAt?: string): number {
  const ref = lastEventAt ?? createdAt;
  if (!ref) return 0;
  const ms = Date.now() - new Date(ref).getTime();
  return Math.max(0, Math.floor(ms / (1000 * 60 * 60 * 24)));
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
