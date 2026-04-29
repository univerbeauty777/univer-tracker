import { fmtHours } from "@/lib/format-hours";
import { cn } from "@/lib/utils";
import type { StageBreakdown } from "@/lib/types";

export function DelayBreakdown({ stages }: { stages: StageBreakdown[] }) {
  const max = Math.max(
    ...stages.map((s) => Math.max(s.actual_hours ?? 0, s.target_hours)),
    1,
  );

  return (
    <div className="space-y-4">
      {stages.map((s) => {
        const targetWidth = (s.target_hours / max) * 100;

        let actualWidth = 0;
        let actualBarColor = "bg-zinc-700";
        let label = "";
        let labelColor = "text-zinc-500";
        let badge: { text: string; cls: string } | null = null;

        if (s.is_pending) {
          const elapsed = s.target_hours - (s.hours_to_target ?? 0);
          actualWidth = Math.max(0, Math.min(100, (Math.max(elapsed, 0) / max) * 100));
          if (s.is_on_time) {
            actualBarColor = "bg-zinc-700/60";
            label = `Pendente — prazo em ${fmtHours(s.hours_to_target ?? 0)}`;
            badge = {
              text: "aguardando",
              cls: "bg-zinc-800 text-zinc-400",
            };
          } else {
            actualBarColor = "bg-rose-500/70";
            label = `Atrasado +${fmtHours(s.delay_hours)} e ainda pendente`;
            labelColor = "text-rose-400";
            badge = { text: "violado", cls: "bg-rose-500/15 text-rose-400" };
          }
        } else {
          actualWidth = ((s.actual_hours ?? 0) / max) * 100;
          if (s.is_on_time) {
            actualBarColor = "bg-emerald-500";
            label = `Concluído em ${fmtHours(s.actual_hours)} (prazo ${fmtHours(s.target_hours)})`;
            labelColor = "text-emerald-400";
            badge = { text: "no prazo", cls: "bg-emerald-500/10 text-emerald-400" };
          } else {
            actualBarColor = "bg-rose-500";
            label = `${fmtHours(s.actual_hours)} (+${fmtHours(s.delay_hours)} além do prazo)`;
            labelColor = "text-rose-400";
            badge = { text: "atrasado", cls: "bg-rose-500/15 text-rose-400" };
          }
        }

        return (
          <div key={s.field}>
            <div className="mb-1.5 flex items-baseline justify-between">
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "text-sm font-medium",
                    s.is_pending && s.is_on_time ? "text-zinc-400" : "text-zinc-200",
                  )}
                >
                  {s.label}
                </span>
                {badge ? (
                  <span
                    className={cn(
                      "rounded-full px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider",
                      badge.cls,
                    )}
                  >
                    {badge.text}
                  </span>
                ) : null}
              </div>
              <span className={cn("text-xs", labelColor)}>{label}</span>
            </div>
            <div className="relative h-7 overflow-hidden rounded-md bg-zinc-900">
              <div
                className={cn("absolute inset-y-0 left-0 opacity-85", actualBarColor)}
                style={{ width: `${actualWidth}%` }}
              />
              <div
                className="absolute inset-y-0 border-l-2 border-dashed border-zinc-500/70"
                style={{ left: `${targetWidth}%` }}
              />
              <span className="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] text-zinc-500">
                prazo {fmtHours(s.target_hours)}
              </span>
            </div>
          </div>
        );
      })}
      <p className="mt-2 text-[11px] text-zinc-500">
        Linha tracejada vertical: prazo SLA cumulativo. Barra colorida: tempo real consumido (ou em
        andamento) na etapa.
      </p>
    </div>
  );
}

export function CascadeBreakdown({ stages, total }: { stages: StageBreakdown[]; total: number }) {
  const cascading = stages.filter((s) => s.cascade_contribution > 0);
  if (cascading.length === 0) return null;
  const max = Math.max(...cascading.map((s) => s.cascade_contribution));
  return (
    <div className="rounded-xl border border-amber-500/20 bg-amber-500/5 p-5">
      <div className="mb-3 flex items-center gap-2">
        <span className="text-sm font-semibold text-amber-300">
          Decomposição do atraso (efeito cascata)
        </span>
        <span className="ml-auto text-xs text-zinc-500">
          total: <span className="font-mono text-rose-400">+{fmtHours(total)}</span>
        </span>
      </div>
      <p className="mb-4 text-xs text-zinc-400">
        Quanto cada etapa adicionou ao tempo total além do que a anterior já tinha atrasado.
      </p>
      <div className="space-y-2">
        {cascading.map((s) => {
          const pct = (s.cascade_contribution / max) * 100;
          return (
            <div key={s.field} className="flex items-center gap-3">
              <span className="w-44 shrink-0 text-sm text-zinc-300">{s.label}</span>
              <div className="relative h-5 flex-1 overflow-hidden rounded bg-zinc-900">
                <div
                  className="absolute inset-y-0 left-0 bg-gradient-to-r from-amber-500 to-rose-500 opacity-80"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <span className="w-16 text-right font-mono text-sm text-rose-400">
                +{fmtHours(s.cascade_contribution)}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
