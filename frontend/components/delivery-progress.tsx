import { Check } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ShipmentStatus } from "@/lib/types";

type Step = {
  key: ShipmentStatus | "label-created";
  label: string;
};

const STEPS: Step[] = [
  { key: "label-created", label: "Etiqueta" },
  { key: "shipped", label: "Postado" },
  { key: "in-transit", label: "Em trânsito" },
  { key: "out-for-delivery", label: "Saiu" },
  { key: "delivered", label: "Entregue" },
];

const ORDER: Record<string, number> = {
  unknown: -1,
  "label-created": 0,
  shipped: 1,
  "in-transit": 2,
  "out-for-delivery": 3,
  delivered: 4,
  "delivery-failed": 2.5,
  returned: 4,
};

/**
 * Horizontal progress bar with named milestones. The current step is
 * filled in primary; failed/returned states show in destructive.
 */
export function DeliveryProgress({ status }: { status: ShipmentStatus | "label-created" }) {
  const idx = ORDER[status] ?? -1;
  const failed = status === "delivery-failed" || status === "returned";

  return (
    <div className="grid grid-cols-5 gap-2">
      {STEPS.map((s, i) => {
        const reached = idx >= i;
        const current = Math.floor(idx) === i;
        return (
          <div key={s.key} className="flex flex-col items-center gap-2">
            <div className="flex w-full items-center gap-2">
              {i > 0 ? (
                <div
                  className={cn(
                    "h-px flex-1",
                    reached ? (failed ? "bg-destructive" : "bg-success") : "bg-border",
                  )}
                />
              ) : (
                <div className="flex-1" />
              )}
              <div
                className={cn(
                  "flex size-7 shrink-0 items-center justify-center rounded-full border text-[11px] font-medium transition-colors",
                  reached
                    ? failed && current
                      ? "border-destructive bg-destructive text-destructive-foreground"
                      : "border-success bg-success text-success-foreground"
                    : "border-border bg-card text-muted-foreground",
                  current && !failed && "ring-4 ring-success/20",
                )}
              >
                {reached && !current ? <Check className="size-3.5" strokeWidth={3} /> : i + 1}
              </div>
              {i < STEPS.length - 1 ? (
                <div
                  className={cn(
                    "h-px flex-1",
                    idx > i ? (failed ? "bg-destructive" : "bg-success") : "bg-border",
                  )}
                />
              ) : (
                <div className="flex-1" />
              )}
            </div>
            <div
              className={cn(
                "text-center text-[11px]",
                current ? "font-medium text-foreground" : "text-muted-foreground",
              )}
            >
              {s.label}
            </div>
          </div>
        );
      })}
    </div>
  );
}
