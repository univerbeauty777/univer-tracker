import { AlertTriangle } from "lucide-react";
import { fetchFunnel, fetchTransitions } from "@/lib/api";
import { fmtHours } from "@/lib/format-hours";
import type { FunnelStage, Transition } from "@/lib/types";

export const dynamic = "force-dynamic";

export default async function GargalosPage() {
  let funnel: FunnelStage[] = [];
  let transitions: Transition[] = [];
  let err: string | null = null;

  try {
    const [f, t] = await Promise.all([fetchFunnel(30), fetchTransitions(30)]);
    funnel = f.stages;
    transitions = t.transitions;
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  // Pior gargalo (maior breach rate com count > 0)
  const worst = transitions
    .filter((t) => t.count > 0)
    .sort((a, b) => b.breach_rate - a.breach_rate)[0];

  const carriers = Array.from(
    new Set(
      transitions
        .flatMap((t) => Object.keys(t.by_carrier))
        .filter((c) => c && c !== "desconhecida"),
    ),
  ).sort();

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="font-display text-2xl font-semibold text-zinc-100">Análise de gargalos</h1>
        <p className="mt-1 text-sm text-zinc-500">
          Onde a operação trava entre o pedido e a entrega. Cada transição é analisada com média,
          percentis e taxa de violação.
        </p>
      </div>

      {err ? (
        <div className="rounded-xl border border-rose-500/30 bg-rose-500/5 p-5 text-sm text-rose-300">
          {err}
        </div>
      ) : null}

      {worst ? (
        <section className="rounded-xl border border-rose-500/30 bg-gradient-to-br from-rose-500/10 to-rose-500/5 p-5">
          <div className="flex items-start gap-3">
            <div className="grid size-10 shrink-0 place-items-center rounded-lg bg-rose-500/15">
              <AlertTriangle className="size-5 text-rose-400" />
            </div>
            <div>
              <p className="text-xs font-medium uppercase tracking-wider text-rose-400">
                Maior gargalo da operação
              </p>
              <p className="mt-1 text-lg font-semibold text-zinc-100">{worst.label}</p>
              <p className="mt-1 text-sm text-zinc-400">
                <strong className="text-rose-300">{worst.breach_rate.toFixed(0)}%</strong> dos
                envios passam dessa etapa <em>fora do prazo</em>. Tempo médio:{" "}
                <strong className="text-zinc-200">{fmtHours(worst.avg_hours)}</strong> · p90:{" "}
                <strong className="text-zinc-200">{fmtHours(worst.p90_hours)}</strong>
              </p>
            </div>
          </div>
        </section>
      ) : null}

      {/* Funil */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-base font-semibold text-zinc-100">Funil de envios por etapa</h2>
          <span className="text-xs text-zinc-500">quantos alcançaram cada etapa</span>
        </div>
        <Funnel stages={funnel} />
      </section>

      {/* Transitions table */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-base font-semibold text-zinc-100">
            Tempo por transição entre etapas
          </h2>
          <span className="text-xs text-zinc-500">média · p50 · p90 · violações</span>
        </div>
        <TransitionsTable transitions={transitions} />
      </section>

      {/* Heatmap */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-base font-semibold text-zinc-100">
            Mapa de calor: violação por etapa × transportadora
          </h2>
          <span className="text-xs text-zinc-500">% de envios atrasados em cada transição</span>
        </div>
        <Heatmap transitions={transitions} carriers={carriers} />
      </section>
    </div>
  );
}

function Funnel({ stages }: { stages: FunnelStage[] }) {
  const max = Math.max(...stages.map((s) => s.count), 1);
  return (
    <div className="space-y-1.5">
      {stages.map((stage, idx) => {
        const pct = (stage.count / max) * 100;
        const drop = idx > 0 ? stages[idx - 1].count - stage.count : 0;
        const dropPct =
          idx > 0 && stages[idx - 1].count > 0 ? (drop / stages[idx - 1].count) * 100 : 0;
        return (
          <div key={stage.field} className="flex items-center gap-3">
            <span className="w-44 shrink-0 text-sm text-zinc-300">{stage.label}</span>
            <div className="relative h-9 flex-1 overflow-hidden rounded-md bg-zinc-900">
              <div
                className="absolute inset-y-0 left-0 flex items-center bg-gradient-to-r from-violet-500/80 to-fuchsia-500/80 px-3"
                style={{ width: `${pct}%` }}
              >
                <span className="text-xs font-semibold text-white">{stage.count}</span>
              </div>
            </div>
            <div className="w-32 shrink-0 text-right text-xs">
              {idx > 0 && drop > 0 ? (
                <>
                  <span className="text-rose-400">−{drop}</span>{" "}
                  <span className="text-zinc-600">({dropPct.toFixed(1)}%)</span>
                </>
              ) : (
                <span className="text-zinc-600">—</span>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

function TransitionsTable({ transitions }: { transitions: Transition[] }) {
  return (
    <div className="overflow-hidden rounded-lg border border-zinc-800">
      <table className="w-full text-sm">
        <thead className="bg-zinc-900/80">
          <tr>
            <Th>Transição</Th>
            <Th align="right">Envios</Th>
            <Th align="right">Médio</Th>
            <Th align="right">p50</Th>
            <Th align="right">p90</Th>
            <Th align="right">% violação</Th>
          </tr>
        </thead>
        <tbody className="divide-y divide-zinc-800">
          {transitions.map((t) => {
            const breachColor =
              t.breach_rate >= 30
                ? "text-rose-400"
                : t.breach_rate >= 15
                ? "text-amber-400"
                : "text-emerald-400";
            return (
              <tr key={t.field} className="hover:bg-zinc-900/40">
                <td className="px-4 py-3 font-medium text-zinc-200">{t.label}</td>
                <td className="px-4 py-3 text-right text-zinc-400">{t.count}</td>
                <td className="px-4 py-3 text-right font-medium text-zinc-200">
                  {fmtHours(t.avg_hours)}
                </td>
                <td className="px-4 py-3 text-right text-zinc-400">{fmtHours(t.p50_hours)}</td>
                <td className="px-4 py-3 text-right text-zinc-400">{fmtHours(t.p90_hours)}</td>
                <td className={`px-4 py-3 text-right font-semibold ${breachColor}`}>
                  {t.breach_rate.toFixed(1)}%
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function Heatmap({
  transitions,
  carriers,
}: {
  transitions: Transition[];
  carriers: string[];
}) {
  const cellColor = (rate: number) => {
    if (rate >= 50) return "bg-rose-500/80 text-white";
    if (rate >= 30) return "bg-rose-500/50 text-rose-100";
    if (rate >= 15) return "bg-amber-500/40 text-amber-100";
    if (rate > 0) return "bg-emerald-500/25 text-emerald-100";
    return "bg-zinc-800 text-zinc-500";
  };

  if (carriers.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-zinc-500">
        Sem dados de carriers suficientes para o heatmap.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-separate border-spacing-1 text-sm">
        <thead>
          <tr>
            <th className="px-3 py-2 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
              Transição
            </th>
            {carriers.map((c) => (
              <th key={c} className="px-3 py-2 text-center text-xs font-medium text-zinc-500">
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {transitions.map((t) => (
            <tr key={t.field}>
              <td className="whitespace-nowrap px-3 py-1.5 text-sm text-zinc-300">{t.label}</td>
              {carriers.map((c) => {
                const cell = t.by_carrier[c];
                if (!cell || cell.count === 0) {
                  return (
                    <td
                      key={c}
                      className="rounded-md bg-zinc-900/40 text-center text-xs text-zinc-700"
                    >
                      —
                    </td>
                  );
                }
                return (
                  <td
                    key={c}
                    className={`rounded-md py-2 text-center text-xs font-semibold ${cellColor(cell.breach_rate)}`}
                  >
                    <div>{cell.breach_rate.toFixed(0)}%</div>
                    <div className="mt-0.5 text-[9px] opacity-75">{fmtHours(cell.avg_hours)}</div>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Th({
  children,
  align = "left",
}: {
  children: React.ReactNode;
  align?: "left" | "right";
}) {
  return (
    <th
      className={`px-4 py-3 text-xs font-medium uppercase tracking-wider text-zinc-500 ${
        align === "right" ? "text-right" : "text-left"
      }`}
    >
      {children}
    </th>
  );
}
