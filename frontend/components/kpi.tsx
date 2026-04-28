import { ArrowDown, ArrowUp, Minus } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type Tone = "neutral" | "success" | "info" | "warning" | "destructive";

const toneClass: Record<Tone, string> = {
  neutral: "text-foreground",
  success: "text-success",
  info: "text-info",
  warning: "text-warning",
  destructive: "text-destructive",
};

export interface KpiProps {
  label: string;
  value: string | number;
  hint?: string;
  icon: React.ComponentType<{ className?: string; strokeWidth?: number }>;
  tone?: Tone;
  /** Delta percent vs previous period. Positive = increase. */
  delta?: number | null;
  /** Whether an increase is good (default true). */
  positiveIsGood?: boolean;
}

/**
 * KPI card with optional delta arrow. Delta colour is decided by the
 * positiveIsGood flag — for "% no prazo" up is good, for "SLA quebrado"
 * up is bad.
 */
export function Kpi({
  label,
  value,
  hint,
  icon: Icon,
  tone = "neutral",
  delta,
  positiveIsGood = true,
}: KpiProps) {
  const hasDelta = delta !== undefined && delta !== null && Number.isFinite(delta);
  let deltaTone: "success" | "destructive" | "muted" = "muted";
  if (hasDelta) {
    if (Math.abs(delta!) < 0.005) deltaTone = "muted";
    else if ((delta! > 0) === positiveIsGood) deltaTone = "success";
    else deltaTone = "destructive";
  }

  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div className="flex size-9 items-center justify-center rounded-lg bg-secondary">
            <Icon className={`size-[18px] ${toneClass[tone]}`} strokeWidth={2} />
          </div>
          {hasDelta ? (
            <span
              className={cn(
                "inline-flex items-center gap-0.5 rounded-full px-1.5 py-0.5 text-[10px] font-medium",
                deltaTone === "success" && "bg-success/10 text-success",
                deltaTone === "destructive" && "bg-destructive/10 text-destructive",
                deltaTone === "muted" && "bg-muted text-muted-foreground",
              )}
              title="Comparado aos 30 dias anteriores"
            >
              {Math.abs(delta!) < 0.005 ? (
                <Minus className="size-2.5" />
              ) : delta! > 0 ? (
                <ArrowUp className="size-2.5" />
              ) : (
                <ArrowDown className="size-2.5" />
              )}
              {Math.abs(delta! * 100).toFixed(0)}%
            </span>
          ) : null}
        </div>
        <div className="mt-3">
          <div className={cn("font-display text-2xl font-semibold tracking-tight", toneClass[tone])}>
            {value}
          </div>
          <div className="mt-0.5 text-xs text-muted-foreground">{label}</div>
          {hint ? <div className="mt-0.5 text-[11px] text-muted-foreground">{hint}</div> : null}
        </div>
      </CardContent>
    </Card>
  );
}

/** Returns `(curr - prev) / prev` or null if previous is zero. */
export function pctDelta(curr: number, prev: number): number | null {
  if (!Number.isFinite(prev) || prev === 0) return null;
  return (curr - prev) / prev;
}

/** Absolute delta (curr - prev). Useful for rates already in [0,1]. */
export function absDelta(curr: number, prev: number): number | null {
  if (!Number.isFinite(prev)) return null;
  return curr - prev;
}
