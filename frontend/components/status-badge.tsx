import { Badge } from "@/components/ui/badge";

type Variant = "default" | "secondary" | "destructive" | "success" | "warning" | "info" | "outline";

const wcVariant: Record<string, Variant> = {
  pending: "warning",
  processing: "info",
  separacao: "info",
  aguardando: "warning",
  "on-hold": "warning",
  enviado: "info",
  shipped: "info",
  "in-transit": "info",
  "em-transito": "info",
  "out-for-delivery": "info",
  "em-rota": "info",
  entregue: "success",
  completed: "success",
  retornado: "destructive",
  cancelled: "destructive",
  refunded: "destructive",
  failed: "destructive",
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
