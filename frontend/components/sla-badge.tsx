import { cn } from "@/lib/utils";
import type { SLAState } from "@/lib/types";

const META: Record<SLAState, { label: string; cls: string }> = {
  ON_TRACK:        { label: "No prazo",           cls: "bg-emerald-500/10 text-emerald-400" },
  AT_RISK:         { label: "Em risco",           cls: "bg-amber-500/15 text-amber-400" },
  BREACHED:        { label: "Atrasado",           cls: "bg-rose-500/15 text-rose-400" },
  COMPLETED:       { label: "Entregue no prazo",  cls: "bg-emerald-500/10 text-emerald-400" },
  COMPLETED_LATE:  { label: "Entregue tardio",    cls: "bg-rose-500/10 text-rose-400" },
};

export function SlaBadge({ state }: { state?: SLAState }) {
  const m = state ? META[state] : null;
  if (!m) {
    return (
      <span className="inline-flex items-center rounded-full bg-zinc-900 px-2.5 py-0.5 text-xs font-medium text-zinc-500">
        —
      </span>
    );
  }
  return (
    <span className={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium", m.cls)}>
      {m.label}
    </span>
  );
}
