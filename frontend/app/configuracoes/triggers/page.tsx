import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { TriggersBoard } from "@/components/triggers-board";
import { fetchTriggers, fetchWAHASessions, fetchIntegrations } from "@/lib/api";
import type {
  IntegrationsResponse,
  NotificationTrigger,
  WAHASession,
} from "@/lib/types";

export const dynamic = "force-dynamic";

export default async function TriggersPage() {
  let triggers: NotificationTrigger[] = [];
  let sessions: WAHASession[] = [];
  let waha: IntegrationsResponse["waha"] | null = null;
  let err: string | null = null;

  try {
    const [t, s, i] = await Promise.all([
      fetchTriggers(),
      fetchWAHASessions().catch(() => ({ sessions: [] })),
      fetchIntegrations().catch(() => null),
    ]);
    triggers = t.triggers ?? [];
    sessions = s.sessions ?? [];
    waha = i?.waha ?? null;
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="mx-auto max-w-[1000px] space-y-6">
      <div>
        <h1 className="font-display text-3xl font-semibold tracking-tight">
          Triggers de envio
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Disparos automáticos via WhatsApp quando uma encomenda muda de
          fase. Cada trigger usa a integração WAHA configurada e respeita o
          cooldown pra não duplicar mensagem.
        </p>
      </div>

      {!waha?.configured ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">WhatsApp não configurado</CardTitle>
            <CardDescription>
              Conecte o gateway WAHA antes de habilitar triggers — abra
              Configurações → Integrações.
            </CardDescription>
          </CardHeader>
        </Card>
      ) : null}

      {err ? (
        <Card>
          <CardContent className="space-y-1 p-6 text-sm">
            <div className="font-medium text-destructive">
              Não foi possível carregar os triggers.
            </div>
            <div className="text-muted-foreground">{err}</div>
          </CardContent>
        </Card>
      ) : (
        <TriggersBoard
          initial={triggers}
          sessions={sessions}
          defaultSession={waha?.default_session ?? ""}
        />
      )}
    </div>
  );
}
