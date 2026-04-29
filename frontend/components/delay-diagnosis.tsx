import { AlertTriangle, Clock } from "lucide-react";
import { fmtHours } from "@/lib/format-hours";
import { formatDateTime } from "@/lib/format";
import type { BreakdownDiagnosis, SLAState } from "@/lib/types";

export function DelayDiagnosis({
  diagnosis,
  slaState,
  expectedAt,
  status,
}: {
  diagnosis: BreakdownDiagnosis;
  slaState?: SLAState;
  expectedAt?: string | null;
  status: string;
}) {
  const hasDelay = Boolean(diagnosis.first_delay_field || diagnosis.worst_delay_field);

  if (!hasDelay) {
    if (slaState === "BREACHED") {
      return (
        <div className="rounded-xl border border-rose-500/30 bg-rose-500/5 p-5">
          <div className="flex items-start gap-3">
            <div className="grid size-9 shrink-0 place-items-center rounded-lg bg-rose-500/15">
              <AlertTriangle className="size-5 text-rose-400" />
            </div>
            <div>
              <p className="text-sm font-semibold text-rose-300">
                SLA violado — prazo total expirado
              </p>
              <p className="mt-1 text-sm text-zinc-400">
                Passou de <strong className="text-zinc-200">{formatDateTime(expectedAt)}</strong>{" "}
                sem ser entregue. Status atual: <strong className="text-zinc-200">{status}</strong>.
              </p>
            </div>
          </div>
        </div>
      );
    }
    if (slaState === "AT_RISK") {
      return (
        <div className="rounded-xl border border-amber-500/30 bg-amber-500/5 p-5">
          <div className="flex items-start gap-3">
            <div className="grid size-9 shrink-0 place-items-center rounded-lg bg-amber-500/15">
              <Clock className="size-5 text-amber-400" />
            </div>
            <div>
              <p className="text-sm font-semibold text-amber-300">Em risco — prazo se aproximando</p>
              <p className="mt-1 text-sm text-zinc-400">
                Já consumiu mais de 80% do prazo SLA. Acompanhe de perto.
              </p>
            </div>
          </div>
        </div>
      );
    }
    return null;
  }

  return (
    <div className="rounded-xl border border-rose-500/30 bg-rose-500/5 p-5">
      <div className="flex items-start gap-3">
        <div className="grid size-9 shrink-0 place-items-center rounded-lg bg-rose-500/15">
          <AlertTriangle className="size-5 text-rose-400" />
        </div>
        <div className="flex-1">
          <p className="text-sm font-semibold text-rose-300">Diagnóstico do atraso</p>
          <p className="mt-1 text-sm text-zinc-300">
            Atraso começou em <strong className="text-zinc-100">{diagnosis.first_delay_label}</strong>
            {" "}
            (<strong className="text-rose-300">+{fmtHours(diagnosis.first_delay_hours)}</strong>),
            comprometendo as etapas seguintes.
          </p>
          <div className="mt-3 grid grid-cols-3 gap-3 text-xs">
            <Cell label="Primeiro atraso" value={diagnosis.first_delay_label ?? "—"} />
            <Cell
              label="Pior etapa"
              value={`${diagnosis.worst_delay_label ?? "—"} +${fmtHours(diagnosis.worst_delay_hours)}`}
            />
            <Cell
              label="Cascata acumulada"
              value={`+${fmtHours(diagnosis.total_cascade_delay)}`}
              tone="rose"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function Cell({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone?: "rose";
}) {
  return (
    <div>
      <p className="uppercase tracking-wider text-zinc-500">{label}</p>
      <p
        className={`mt-1 font-medium ${
          tone === "rose" ? "text-rose-300" : "text-zinc-200"
        }`}
      >
        {value}
      </p>
    </div>
  );
}
