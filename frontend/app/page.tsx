import Link from "next/link";
import {
  Activity,
  AlertTriangle,
  Clock,
  Hourglass,
  Package,
  PackageCheck,
  Timer,
  TrendingUp,
  Truck,
} from "lucide-react";
import { ExportButton } from "@/components/export-button";
import { Kpi, absDelta, pctDelta } from "@/components/kpi";
import { LastSyncBanner } from "@/components/last-sync-banner";
import { SlaBadge } from "@/components/sla-badge";
import { fetchOrders, fetchOverview } from "@/lib/api";
import { dedupeName, formatDate, formatRelative } from "@/lib/format";
import { fmtHours } from "@/lib/format-hours";
import type { CarrierStats, OrderListItem, Overview, SLAState } from "@/lib/types";

export const dynamic = "force-dynamic";

export default async function PainelPage() {
  let overview: Overview | null = null;
  let carriers: CarrierStats[] = [];
  let alerts: OrderListItem[] = [];
  let err: string | null = null;

  try {
    const [ov, alert] = await Promise.all([
      fetchOverview(),
      fetchOrders({ health: "at_risk,breached", per_page: 8, sort: "last_event", dir: "asc" }),
    ]);
    overview = ov.overview;
    carriers = ov.carriers ?? [];
    alerts = alert.orders ?? [];
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  const prev = overview?.previous_period;
  const fmtPhase = (h?: number) =>
    h && Number.isFinite(h) && h > 0 ? fmtHours(h) : "—";

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-display text-2xl font-semibold text-zinc-100">Painel logístico</h1>
          <p className="mt-1 text-sm text-zinc-500">
            Visão consolidada com SLA por etapa, gargalos da operação e alertas em tempo real.
          </p>
        </div>
        <ExportButton />
      </div>

      <LastSyncBanner />

      <section className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
        <Kpi
          label="Total de envios"
          value={overview?.total_30d ?? 0}
          hint={`${overview?.delivered_30d ?? 0} entregues`}
          icon={Package}
          delta={prev ? pctDelta(overview?.total_30d ?? 0, prev.total_30d) : null}
        />
        <Kpi
          label="Em andamento"
          value={overview?.in_progress ?? 0}
          hint="no pipeline ativo"
          icon={Activity}
        />
        <Kpi
          label="Taxa OTD"
          value={`${Math.round((overview?.on_time_rate ?? 0) * 100)}%`}
          hint="entregas no prazo"
          icon={TrendingUp}
          tone="success"
          delta={prev ? absDelta(overview?.on_time_rate ?? 0, prev.on_time_rate) : null}
        />
        <Kpi
          label="Em risco"
          value={overview?.at_risk ?? 0}
          hint=">80% do prazo"
          icon={AlertTriangle}
          tone="warning"
          positiveIsGood={false}
        />
        <Kpi
          label="Atrasados"
          value={overview?.breached ?? 0}
          hint="SLA violado"
          icon={AlertTriangle}
          tone="destructive"
          positiveIsGood={false}
        />
        <Kpi
          label="Lead time médio"
          value={
            overview && overview.avg_delivery_days > 0
              ? `${overview.avg_delivery_days.toFixed(1)}d`
              : "—"
          }
          hint="pedido → entrega"
          icon={Timer}
          delta={
            prev && overview && overview.avg_delivery_days > 0 && prev.avg_delivery_days > 0
              ? pctDelta(overview.avg_delivery_days, prev.avg_delivery_days)
              : null
          }
          positiveIsGood={false}
        />
      </section>

      {/* KPIs por fase */}
      <section className="grid grid-cols-1 gap-3 md:grid-cols-4">
        <Kpi
          label="Preparação"
          value={fmtPhase(overview?.avg_preparing_hours)}
          hint="pedido → pronto p/ coleta"
          icon={Hourglass}
          tone="info"
        />
        <Kpi
          label="Trânsito"
          value={fmtPhase(overview?.avg_in_transit_hours)}
          hint="postagem → entrega"
          icon={Truck}
          tone="info"
        />
        <Kpi
          label="Last mile"
          value={fmtPhase(overview?.avg_last_mile_hours)}
          hint="saiu p/ entrega → entregue"
          icon={PackageCheck}
          tone="info"
        />
        <Kpi
          label="Sem evento >4d"
          value={overview?.idle_alarms ?? 0}
          hint="alarmes ociosos"
          icon={Clock}
          tone="warning"
          positiveIsGood={false}
        />
      </section>

      {/* Performance por transportadora */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <h2 className="text-base font-semibold text-zinc-100">Performance por transportadora</h2>
            <p className="text-xs text-zinc-500">Volume e quebras de SLA nos últimos 30 dias.</p>
          </div>
          <Truck className="size-4 text-zinc-500" />
        </div>
        {carriers.length === 0 ? (
          <p className="py-6 text-center text-sm text-zinc-500">
            Sem dados suficientes ainda. Aguardando primeiras entregas.
          </p>
        ) : (
          <div className="overflow-hidden rounded-lg border border-zinc-800">
            <table className="w-full text-sm">
              <thead className="bg-zinc-900/80">
                <tr>
                  <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                    Transportadora
                  </th>
                  <th className="px-4 py-2.5 text-right text-xs font-medium uppercase tracking-wider text-zinc-500">
                    Volume
                  </th>
                  <th className="px-4 py-2.5 text-right text-xs font-medium uppercase tracking-wider text-zinc-500">
                    SLA quebrado
                  </th>
                  <th className="px-4 py-2.5 text-right text-xs font-medium uppercase tracking-wider text-zinc-500">
                    Lead médio
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800">
                {carriers.map((c) => (
                  <tr key={c.carrier} className="hover:bg-zinc-900/40">
                    <td className="px-4 py-2.5 text-zinc-200">{c.carrier}</td>
                    <td className="px-4 py-2.5 text-right text-zinc-300">{c.total}</td>
                    <td className="px-4 py-2.5 text-right">
                      {c.breached > 0 ? (
                        <span className="text-rose-400">
                          {c.breached} ({Math.round((c.breached / c.total) * 100)}%)
                        </span>
                      ) : (
                        <span className="text-zinc-500">—</span>
                      )}
                    </td>
                    <td className="px-4 py-2.5 text-right text-zinc-400">
                      {c.avg_delivery_days > 0 ? `${c.avg_delivery_days.toFixed(1)}d` : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Alertas */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-base font-semibold text-zinc-100">Envios em risco e atrasados</h2>
          <span className="rounded-full bg-rose-500/15 px-2.5 py-1 text-xs font-medium text-rose-400">
            {alerts.length} alertas
          </span>
        </div>
        {err ? (
          <p className="text-sm text-rose-400">Não foi possível carregar alertas: {err}</p>
        ) : alerts.length === 0 ? (
          <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-4 text-center text-sm text-emerald-400">
            Nenhum envio em risco no momento.
          </div>
        ) : (
          <div className="space-y-2">
            {alerts.map((o) => {
              const breached = o.tracking.sla_state === "BREACHED";
              return (
                <Link
                  key={o.id}
                  href={`/pedidos/${o.id}`}
                  className={`flex items-center justify-between rounded-lg border p-3 transition-colors ${
                    breached
                      ? "border-rose-500/30 bg-rose-500/5 hover:bg-rose-500/10"
                      : "border-amber-500/30 bg-amber-500/5 hover:bg-amber-500/10"
                  }`}
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <code className="text-xs text-zinc-300">
                        {o.tracking.number || `#${o.id}`}
                      </code>
                      <span className="text-xs text-zinc-500">·</span>
                      <span className="text-xs text-zinc-400">{o.tracking.carrier || "—"}</span>
                    </div>
                    <p className="mt-1 truncate text-sm text-zinc-200">
                      {dedupeName(o.customer_name) || "—"}
                    </p>
                    <p className="mt-0.5 text-[11px] text-zinc-500">
                      Último evento: {formatRelative(o.tracking.last_event_at)}
                    </p>
                  </div>
                  <div className="ml-4 shrink-0 text-right">
                    <SlaBadge state={o.tracking.sla_state as SLAState | undefined} />
                    <p className="mt-1 text-[11px] text-zinc-500">
                      ETA: {formatDate(o.tracking.estimated_delivery)}
                    </p>
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
