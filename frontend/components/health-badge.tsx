import { Badge } from "@/components/ui/badge";
import type { Health } from "@/lib/types";

const tone: Record<Health, "success" | "warning" | "destructive" | "secondary"> = {
  unknown: "secondary",
  on_track: "success",
  at_risk: "warning",
  breached: "destructive",
};

export function HealthBadge({ health, label }: { health: Health; label: string }) {
  return <Badge variant={tone[health] ?? "secondary"}>{label}</Badge>;
}
