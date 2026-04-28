import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/format";
import type { Health } from "@/lib/types";

const fillTone: Record<Health, string> = {
  unknown: "bg-muted-foreground/40",
  on_track: "bg-success",
  at_risk: "bg-warning",
  breached: "bg-destructive",
};

export function SlaTracker({
  createdAt,
  estimatedDelivery,
  deliveredAt,
  health,
}: {
  createdAt: string;
  estimatedDelivery?: string;
  deliveredAt?: string;
  health: Health;
}) {
  if (!estimatedDelivery) return null;

  const start = new Date(createdAt).getTime();
  const eta = new Date(estimatedDelivery).getTime();
  const now = deliveredAt ? new Date(deliveredAt).getTime() : Date.now();

  const total = Math.max(eta - start, 1);
  const elapsed = now - start;
  const percent = Math.max(0, Math.min(100, (elapsed / total) * 100));

  const breach = elapsed > total;
  const tone = deliveredAt ? "bg-success" : fillTone[health];
  const overshoot = breach
    ? Math.min(100, ((elapsed - total) / total) * 100)
    : 0;

  return (
    <div className="space-y-2">
      <div className="relative h-2 w-full overflow-hidden rounded-full bg-secondary">
        <div
          className={cn("h-full transition-all", tone)}
          style={{ width: `${breach ? 100 : percent}%` }}
        />
        {breach && !deliveredAt ? (
          <div
            className="absolute inset-y-0 right-0 bg-destructive/30"
            style={{ width: `${overshoot}%` }}
          />
        ) : null}
      </div>
      <div className="flex justify-between text-[11px] text-muted-foreground">
        <span>Início · {formatDate(createdAt)}</span>
        <span>
          {deliveredAt
            ? `Entregue · ${formatDate(deliveredAt)}`
            : `ETA · ${formatDate(estimatedDelivery)}`}
        </span>
      </div>
    </div>
  );
}
