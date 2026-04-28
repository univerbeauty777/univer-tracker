import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchOrderHistory } from "@/lib/api";
import { formatDateTime } from "@/lib/format";

export async function OrderHistory({ orderId }: { orderId: number }) {
  let history;
  try {
    history = await fetchOrderHistory(orderId);
  } catch {
    return null;
  }
  const changes = history.changes ?? [];
  const notes = history.notifications ?? [];

  if (changes.length === 0 && notes.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Histórico</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {changes.length > 0 ? (
          <div>
            <div className="mb-2 text-[11px] uppercase tracking-wide text-muted-foreground">
              Mudanças de status
            </div>
            <ul className="space-y-2">
              {changes.map((c) => (
                <li key={c.id} className="flex items-start gap-3 text-xs">
                  <span className="mt-1 inline-block size-1.5 shrink-0 rounded-full bg-info" />
                  <div className="min-w-0 flex-1">
                    <div className="text-sm">
                      <span className="text-muted-foreground">{c.from_status || "—"}</span>
                      {" → "}
                      <span className="font-medium">{c.to_status}</span>
                    </div>
                    <div className="text-[11px] text-muted-foreground">
                      {formatDateTime(c.created_at)} · {c.actor} · {c.source}
                    </div>
                    {c.note ? (
                      <div className="mt-1 rounded bg-muted/40 px-2 py-1 text-xs text-foreground">
                        {c.note}
                      </div>
                    ) : null}
                  </div>
                </li>
              ))}
            </ul>
          </div>
        ) : null}

        {notes.length > 0 ? (
          <div>
            <div className="mb-2 text-[11px] uppercase tracking-wide text-muted-foreground">
              Notificações enviadas
            </div>
            <ul className="space-y-2">
              {notes.map((n) => (
                <li key={n.id} className="flex items-start gap-3 text-xs">
                  <span
                    className={`mt-1 inline-block size-1.5 shrink-0 rounded-full ${
                      n.status === "sent" ? "bg-success" : "bg-destructive"
                    }`}
                  />
                  <div className="min-w-0 flex-1">
                    <div className="text-sm">
                      {n.channel === "waha" ? "WhatsApp" : n.channel}
                      {n.template ? ` · ${n.template}` : ""}
                    </div>
                    <div className="text-[11px] text-muted-foreground">
                      {formatDateTime(n.sent_at)} · {n.status}
                    </div>
                    {n.error ? (
                      <div className="mt-1 rounded bg-destructive/10 px-2 py-1 text-xs text-destructive">
                        {n.error}
                      </div>
                    ) : null}
                  </div>
                </li>
              ))}
            </ul>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
