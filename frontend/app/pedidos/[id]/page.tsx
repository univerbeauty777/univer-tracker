import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft, MapPin, Package, ExternalLink, Activity } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/status-badge";
import { HealthBadge } from "@/components/health-badge";
import { SlaTracker } from "@/components/sla-tracker";
import { DeliveryProgress } from "@/components/delivery-progress";
import { ChangeStatusAction } from "@/components/change-status-action";
import { fetchOrder } from "@/lib/api";
import { dedupeName, formatBRL, formatDate, formatDateTime, formatRelative } from "@/lib/format";
import type { Health, ShipmentStatus } from "@/lib/types";

export const dynamic = "force-dynamic";

export default async function OrderDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const orderId = Number(id);
  if (!Number.isFinite(orderId)) notFound();

  let order;
  try {
    order = await fetchOrder(orderId);
  } catch {
    notFound();
  }

  const events = order.tracking.events ?? [];

  return (
    <div className="mx-auto max-w-[1200px] space-y-6">
      <div>
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-3.5" />
          Pedidos
        </Link>

        <div className="mt-3 flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="font-display text-3xl font-semibold tracking-tight">
              Pedido #{order.id}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {dedupeName(order.customer_name)} · criado em {formatDate(order.created_at)}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <StatusBadge status={order.status} label={order.status_label} />
            <HealthBadge
              health={order.tracking.health as Health}
              label={order.tracking.health_label}
            />
            <ChangeStatusAction orderId={order.id} currentStatus={order.status} />
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="size-4 text-info" />
                Rastreamento
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {order.tracking.number ? (
                <div className="space-y-5">
                  <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border/60 bg-muted/30 p-4">
                    <div>
                      <div className="text-[11px] uppercase tracking-wide text-muted-foreground">
                        Código
                      </div>
                      <div className="font-mono text-sm font-medium">{order.tracking.number}</div>
                      <div className="text-xs text-muted-foreground">
                        {order.tracking.carrier || "transportadora desconhecida"}
                        {order.tracking.service ? ` · ${order.tracking.service}` : ""}
                      </div>
                    </div>
                    {order.tracking.url ? (
                      <a
                        href={order.tracking.url}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                      >
                        Rastrear no site
                        <ExternalLink className="size-3" />
                      </a>
                    ) : null}
                  </div>

                  <DeliveryProgress status={order.tracking.status as ShipmentStatus} />

                  <SlaTracker
                    createdAt={order.created_at}
                    estimatedDelivery={order.tracking.estimated_delivery}
                    deliveredAt={order.tracking.delivered_at}
                    health={order.tracking.health as Health}
                  />

                  <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                    <SlaStat label="Risco" value={`${order.tracking.risk_score}/100`} icon={Activity} />
                    <SlaStat
                      label="Sem evento há"
                      value={
                        order.tracking.idle_since
                          ? `${idleDays(order.tracking.idle_since)} dias`
                          : "—"
                      }
                    />
                    <SlaStat
                      label="ETA"
                      value={
                        order.tracking.estimated_delivery
                          ? formatDate(order.tracking.estimated_delivery)
                          : "—"
                      }
                    />
                    <SlaStat
                      label="Entregue em"
                      value={
                        order.tracking.delivered_at
                          ? formatDate(order.tracking.delivered_at)
                          : "—"
                      }
                    />
                  </div>
                </div>
              ) : (
                <div className="rounded-lg border border-dashed border-border/60 p-6 text-center text-sm text-muted-foreground">
                  Sem código de rastreio. A integração com Frenet linka automaticamente
                  quando a etiqueta é gerada — também é possível adicionar manualmente
                  no WooCommerce.
                </div>
              )}

              {events.length > 0 ? (
                <ol className="relative space-y-4 border-l border-border/60 pl-6">
                  {events.map((e, i) => (
                    <li key={`${e.occurred_at}-${i}`} className="relative">
                      <span className="absolute -left-[27px] top-1 flex size-3 items-center justify-center rounded-full bg-info ring-4 ring-background" />
                      <div className="text-sm font-medium">{e.description}</div>
                      <div className="mt-0.5 flex flex-wrap gap-3 text-xs text-muted-foreground">
                        <span>{formatDateTime(e.occurred_at)}</span>
                        {e.location ? (
                          <span className="inline-flex items-center gap-1">
                            <MapPin className="size-3" />
                            {e.location}
                          </span>
                        ) : null}
                      </div>
                    </li>
                  ))}
                </ol>
              ) : order.tracking.number ? (
                <div className="rounded-lg border border-dashed border-border/60 p-6 text-center text-sm text-muted-foreground">
                  A Frenet ainda não retornou eventos para este código. Tente novamente em
                  alguns minutos — a transportadora costuma demorar até a primeira leitura.
                </div>
              ) : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Itens</CardTitle>
            </CardHeader>
            <CardContent className="px-0">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border/60 text-xs uppercase tracking-wide text-muted-foreground">
                      <th className="px-6 py-3 text-left font-medium">Produto</th>
                      <th className="px-6 py-3 text-right font-medium">Qtd</th>
                      <th className="px-6 py-3 text-right font-medium">Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {order.line_items.map((li) => (
                      <tr key={li.id} className="border-b border-border/40 last:border-0">
                        <td className="px-6 py-3">{li.name}</td>
                        <td className="px-6 py-3 text-right text-muted-foreground">
                          {li.quantity}
                        </td>
                        <td className="px-6 py-3 text-right font-medium">
                          {formatBRL(li.total)}
                        </td>
                      </tr>
                    ))}
                    <tr className="bg-muted/30">
                      <td className="px-6 py-3 text-right text-muted-foreground" colSpan={2}>
                        Total do pedido
                      </td>
                      <td className="px-6 py-3 text-right font-display text-base font-semibold">
                        {formatBRL(order.total)}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Cliente</CardTitle>
            </CardHeader>
            <CardContent className="space-y-1 text-sm">
              <div className="font-medium">{dedupeName(order.customer_name) || "—"}</div>
              <div className="text-muted-foreground">{order.email || "—"}</div>
              <div className="text-muted-foreground">{order.phone || "—"}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Entrega</CardTitle>
            </CardHeader>
            <CardContent className="space-y-1 text-sm">
              <div>{dedupeName(`${order.shipping.first_name} ${order.shipping.last_name}`)}</div>
              <div className="text-muted-foreground">
                {order.shipping.city} · {order.shipping.state} · {order.shipping.postcode}
              </div>
              {order.shipping_method ? (
                <div className="pt-2 text-xs text-muted-foreground">
                  Método: {order.shipping_method}
                </div>
              ) : null}
              {order.paid_at ? (
                <div className="text-xs text-muted-foreground">
                  Pago em {formatDate(order.paid_at)}
                </div>
              ) : null}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function SlaStat({
  label,
  value,
  icon: Icon,
}: {
  label: string;
  value: string;
  icon?: React.ComponentType<{ className?: string }>;
}) {
  return (
    <div className="rounded-lg border border-border/60 bg-card p-3">
      <div className="flex items-center gap-1.5 text-[10px] uppercase tracking-wide text-muted-foreground">
        {Icon ? <Icon className="size-3" /> : null}
        {label}
      </div>
      <div className="mt-1 font-display text-base font-semibold">{value}</div>
    </div>
  );
}

function idleDays(idleSince: string): number {
  const ms = Date.now() - new Date(idleSince).getTime();
  return Math.max(0, Math.floor(ms / (1000 * 60 * 60 * 24)));
}
