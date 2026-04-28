import Link from "next/link";
import { ChevronRight, Package, Truck, CheckCircle2, AlertTriangle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { StatusBadge, ShipmentBadge } from "@/components/status-badge";
import { fetchOrders } from "@/lib/api";
import { formatBRL, formatDate } from "@/lib/format";
import type { OrderListItem } from "@/lib/types";

export const dynamic = "force-dynamic";

export default async function OrdersPage() {
  let orders: OrderListItem[] = [];
  let err: string | null = null;

  try {
    const res = await fetchOrders({
      status: "processing,on-hold,shipped,in-transit,out-for-delivery,completed",
      per_page: 100,
    });
    orders = res.orders;
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  const stats = computeStats(orders);

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="font-display text-3xl font-semibold tracking-tight">
          Pedidos
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Pedidos do WooCommerce com status de entrega cruzado com a Frenet.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <StatCard label="Total" value={stats.total} icon={Package} tone="text-foreground" />
        <StatCard label="Em trânsito" value={stats.transit} icon={Truck} tone="text-info" />
        <StatCard label="Entregues" value={stats.delivered} icon={CheckCircle2} tone="text-success" />
        <StatCard label="Atenção" value={stats.attention} icon={AlertTriangle} tone="text-warning" />
      </div>

      {err ? (
        <Card>
          <CardContent className="space-y-1 p-6 text-sm">
            <div className="font-medium text-destructive">Não foi possível carregar pedidos.</div>
            <div className="text-muted-foreground">{err}</div>
            <div className="text-xs text-muted-foreground">
              Verifique as credenciais WooCommerce nas variáveis de ambiente do backend.
            </div>
          </CardContent>
        </Card>
      ) : orders.length === 0 ? (
        <Card>
          <CardContent className="p-10 text-center text-sm text-muted-foreground">
            Nenhum pedido encontrado.
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="px-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border/60 text-xs uppercase tracking-wide text-muted-foreground">
                    <th className="px-6 py-3 text-left font-medium">Pedido</th>
                    <th className="px-6 py-3 text-left font-medium">Cliente</th>
                    <th className="px-6 py-3 text-left font-medium">Status WooCommerce</th>
                    <th className="px-6 py-3 text-left font-medium">Rastreio</th>
                    <th className="px-6 py-3 text-left font-medium">Entrega</th>
                    <th className="px-6 py-3 text-right font-medium">Total</th>
                    <th className="w-10" />
                  </tr>
                </thead>
                <tbody>
                  {orders.map((o) => (
                    <tr
                      key={o.id}
                      className="border-b border-border/40 transition-colors last:border-0 hover:bg-muted/30"
                    >
                      <td className="px-6 py-4">
                        <Link
                          href={`/pedidos/${o.id}`}
                          className="font-mono text-xs font-medium hover:underline"
                        >
                          #{o.id}
                        </Link>
                        <div className="text-[11px] text-muted-foreground">
                          {formatDate(o.created_at)}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="font-medium">{o.customer_name || "—"}</div>
                        <div className="text-[11px] text-muted-foreground">
                          {[o.customer_city, o.customer_state].filter(Boolean).join(" · ") || "—"}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <StatusBadge status={o.status} label={o.status_label} />
                      </td>
                      <td className="px-6 py-4">
                        {o.tracking.number ? (
                          <div>
                            <div className="font-mono text-xs">{o.tracking.number}</div>
                            <div className="text-[11px] text-muted-foreground">
                              {o.tracking.carrier || "transportadora desconhecida"}
                            </div>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">sem código</span>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        <ShipmentBadge status={o.tracking.status} label={o.tracking.status_label} />
                      </td>
                      <td className="px-6 py-4 text-right font-medium">
                        {formatBRL(o.total)}
                      </td>
                      <td className="px-3 py-4 text-muted-foreground">
                        <Link href={`/pedidos/${o.id}`} aria-label={`Abrir pedido ${o.id}`}>
                          <ChevronRight className="size-4" />
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function computeStats(orders: OrderListItem[]) {
  const transit = orders.filter((o) =>
    ["shipped", "in-transit", "out-for-delivery"].includes(o.tracking.status) ||
    ["shipped", "in-transit", "out-for-delivery"].includes(o.status),
  ).length;
  const delivered = orders.filter(
    (o) => o.tracking.status === "delivered" || o.status === "completed",
  ).length;
  const attention = orders.filter(
    (o) => ["delivery-failed", "returned"].includes(o.tracking.status) || o.status === "on-hold",
  ).length;
  return { total: orders.length, transit, delivered, attention };
}

function StatCard({
  label,
  value,
  icon: Icon,
  tone,
}: {
  label: string;
  value: number;
  icon: React.ComponentType<{ className?: string; strokeWidth?: number }>;
  tone: string;
}) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex size-9 items-center justify-center rounded-lg bg-secondary">
          <Icon className={`size-[18px] ${tone}`} strokeWidth={2} />
        </div>
        <div className="mt-3">
          <div className={`font-display text-3xl font-semibold ${tone}`}>{value}</div>
          <div className="mt-0.5 text-xs text-muted-foreground">{label}</div>
        </div>
      </CardContent>
    </Card>
  );
}
