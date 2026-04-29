import { cn } from "@/lib/utils";

/**
 * Compact ring chart that visualises a 0-100 risk score.
 * Green up to 30, amber up to 70, red above. Uses currentColor so the
 * stroke colour comes from the wrapping `text-*` class (Tailwind doesn't
 * generate stroke-{theme-color} by default).
 */
export function RiskRing({ score, size = 56 }: { score: number; size?: number }) {
  const v = Math.max(0, Math.min(100, score));
  const stroke = 5;
  const r = (size - stroke) / 2;
  const c = 2 * Math.PI * r;
  const offset = c - (v / 100) * c;

  const tone =
    v < 30 ? "text-success" : v < 70 ? "text-warning" : "text-destructive";

  return (
    <div
      className={cn("relative inline-flex items-center justify-center", tone)}
      style={{ width: size, height: size }}
    >
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          strokeWidth={stroke}
          stroke="hsl(var(--secondary))"
          fill="none"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          strokeWidth={stroke}
          stroke="currentColor"
          fill="none"
          strokeLinecap="round"
          strokeDasharray={c}
          strokeDashoffset={offset}
          className="transition-[stroke-dashoffset]"
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center text-[11px] font-semibold text-foreground">
        {v}
      </div>
    </div>
  );
}
