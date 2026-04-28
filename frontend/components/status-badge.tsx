import { Badge } from "@/components/ui/badge";

type Variant = "default" | "secondary" | "destructive" | "success" | "warning" | "info" | "outline";

const wcVariant: Record<string, Variant> = {
  pending: "warning",
  processing: "info",
  "on-hold": "warning",
  completed: "success",
  cancelled: "destructive",
  refunded: "destructive",
  failed: "destructive",
  shipped: "info",
  "in-transit": "info",
  "out-for-delivery": "info",
};

export function StatusBadge({ status, label }: { status: string; label: string }) {
  return <Badge variant={wcVariant[status] ?? "secondary"}>{label}</Badge>;
}

const shipmentVariant: Record<string, Variant> = {
  unknown: "secondary",
  shipped: "info",
  "in-transit": "info",
  "out-for-delivery": "info",
  delivered: "success",
  "delivery-failed": "warning",
  returned: "destructive",
};

export function ShipmentBadge({ status, label }: { status: string; label: string }) {
  return <Badge variant={shipmentVariant[status] ?? "secondary"}>{label}</Badge>;
}
