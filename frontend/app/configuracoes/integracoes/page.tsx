import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchIntegrations } from "@/lib/api";
import { IntegrationsBoard } from "@/components/integrations-board";
import type { IntegrationsResponse } from "@/lib/types";

export const dynamic = "force-dynamic";

export default async function IntegrationsPage() {
  let initial: IntegrationsResponse | null = null;
  let err: string | null = null;
  try {
    initial = await fetchIntegrations();
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="mx-auto max-w-[1000px] space-y-6">
      <div>
        <h1 className="font-display text-3xl font-semibold tracking-tight">Integrações</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Conecte os gateways de logística sem redeployar. As credenciais são salvas com
          segurança no banco e o backend recarrega sozinho a cada poucos segundos.
        </p>
      </div>

      {err ? (
        <Card>
          <CardContent className="space-y-1 p-6 text-sm">
            <div className="font-medium text-destructive">
              Não foi possível carregar as integrações.
            </div>
            <div className="text-muted-foreground">{err}</div>
          </CardContent>
        </Card>
      ) : initial ? (
        <IntegrationsBoard initial={initial} />
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Carregando…</CardTitle>
            <CardDescription>Buscando configuração atual.</CardDescription>
          </CardHeader>
        </Card>
      )}
    </div>
  );
}
