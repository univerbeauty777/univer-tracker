"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import {
  AlertTriangle,
  CheckCircle2,
  Loader2,
  PackageCheck,
  Send,
  Truck,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { saveTriggers } from "@/lib/api";
import type {
  NotificationTrigger,
  TriggerEventKey,
  WAHASession,
} from "@/lib/types";
import { cn } from "@/lib/utils";

const META: Record<
  TriggerEventKey,
  {
    title: string;
    subtitle: string;
    icon: typeof Send;
    tone: string;
  }
> = {
  postado: {
    title: "Postado",
    subtitle: "O objeto foi postado e está com a transportadora.",
    icon: Send,
    tone: "text-info",
  },
  in_transit: {
    title: "Em trânsito",
    subtitle: "Objeto deixou a unidade de origem.",
    icon: Truck,
    tone: "text-violet-400",
  },
  delivered: {
    title: "Entregue",
    subtitle: "Carrier confirmou a entrega.",
    icon: PackageCheck,
    tone: "text-success",
  },
  breached: {
    title: "Atrasado (SLA quebrado)",
    subtitle: "O prazo cumulativo foi violado.",
    icon: AlertTriangle,
    tone: "text-destructive",
  },
};

const PLACEHOLDERS = [
  { token: "{first_name}", desc: "primeiro nome do cliente" },
  { token: "{customer_name}", desc: "nome completo" },
  { token: "{order_id}", desc: "número do pedido WC" },
  { token: "{tracking}", desc: "código de rastreio" },
  { token: "{track_url}", desc: "link público de rastreio" },
  { token: "{last_event}", desc: "último evento da transportadora" },
  { token: "{eta}", desc: "previsão de entrega (dd/mm/aaaa)" },
  { token: "{carrier}", desc: "nome da transportadora" },
];

const ORDER: TriggerEventKey[] = ["postado", "in_transit", "delivered", "breached"];

export function TriggersBoard({
  initial,
  sessions,
  defaultSession,
}: {
  initial: NotificationTrigger[];
  sessions: WAHASession[];
  defaultSession: string;
}) {
  const router = useRouter();
  const [rows, setRows] = useState<NotificationTrigger[]>(() => mergeOrder(initial));
  const [pending, start] = useTransition();
  const [saved, setSaved] = useState<{ ok: boolean; text: string } | null>(null);

  function patch(key: TriggerEventKey, p: Partial<NotificationTrigger>) {
    setRows((prev) => prev.map((r) => (r.event_key === key ? { ...r, ...p } : r)));
  }

  function save() {
    setSaved(null);
    start(async () => {
      try {
        const res = await saveTriggers(rows);
        setRows(mergeOrder(res.triggers));
        setSaved({ ok: true, text: "Triggers salvos." });
        router.refresh();
        setTimeout(() => setSaved(null), 3000);
      } catch (e) {
        setSaved({
          ok: false,
          text: e instanceof Error ? e.message : "Erro ao salvar",
        });
      }
    });
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Como funciona</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-xs text-muted-foreground">
          <p>
            A rastreiaki monitora cada envio em tempo real. Quando uma encomenda
            cruza um marco (Postado · Em trânsito · Entregue · Atrasado), o
            trigger correspondente dispara uma mensagem via WAHA pro telefone
            cadastrado no pedido.
          </p>
          <p>
            <strong className="text-foreground">Cooldown</strong> evita disparos
            duplicados — se o evento se repete dentro da janela, ignoramos. 0
            minutos = manda no máximo uma vez por pedido.
          </p>
          <p className="pt-1">
            Tokens disponíveis no template:{" "}
            {PLACEHOLDERS.map((p, i) => (
              <span key={p.token}>
                <code className="rounded bg-muted/50 px-1 py-0.5 font-mono text-[10.5px]">
                  {p.token}
                </code>
                {i < PLACEHOLDERS.length - 1 ? " " : ""}
              </span>
            ))}
          </p>
        </CardContent>
      </Card>

      {rows.map((row) => (
        <TriggerCard
          key={row.event_key}
          row={row}
          sessions={sessions}
          defaultSession={defaultSession}
          onChange={(p) => patch(row.event_key, p)}
        />
      ))}

      <div className="sticky bottom-3 flex flex-wrap items-center justify-end gap-3 rounded-lg border border-border/60 bg-card/80 px-4 py-2.5 backdrop-blur-md">
        {saved ? (
          <span
            className={cn(
              "text-xs",
              saved.ok ? "text-success" : "text-destructive",
            )}
          >
            {saved.text}
          </span>
        ) : null}
        <Button onClick={save} disabled={pending} size="sm">
          {pending ? <Loader2 className="size-3.5 animate-spin" /> : <CheckCircle2 className="size-3.5" />}
          Salvar triggers
        </Button>
      </div>
    </div>
  );
}

function TriggerCard({
  row,
  sessions,
  defaultSession,
  onChange,
}: {
  row: NotificationTrigger;
  sessions: WAHASession[];
  defaultSession: string;
  onChange: (p: Partial<NotificationTrigger>) => void;
}) {
  const meta = META[row.event_key];
  const Icon = meta.icon;
  return (
    <Card className={cn(!row.enabled && "opacity-80")}>
      <CardHeader className="flex flex-row items-start justify-between gap-3 space-y-0 pb-3">
        <div className="flex items-start gap-3">
          <div className="grid size-9 shrink-0 place-items-center rounded-lg bg-muted/40">
            <Icon className={cn("size-4", meta.tone)} strokeWidth={1.75} />
          </div>
          <div>
            <CardTitle className="text-base">{meta.title}</CardTitle>
            <p className="mt-0.5 text-xs text-muted-foreground">{meta.subtitle}</p>
          </div>
        </div>
        <label className="inline-flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
          <input
            type="checkbox"
            checked={row.enabled}
            onChange={(e) => onChange({ enabled: e.target.checked })}
            className="size-4 rounded border-border/80 bg-background accent-violet-500"
          />
          Habilitado
        </label>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="block text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Sessão WAHA
            </label>
            <select
              value={row.session ?? ""}
              onChange={(e) => onChange({ session: e.target.value })}
              className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-1.5 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <option value="">
                — usar padrão{defaultSession ? ` (${defaultSession})` : ""}
              </option>
              {sessions.map((s) => (
                <option key={s.name} value={s.name}>
                  {s.name}
                  {s.status ? ` · ${s.status.toLowerCase()}` : ""}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Cooldown (min)
            </label>
            <input
              type="number"
              min={0}
              max={10080}
              value={row.cooldown_minutes}
              onChange={(e) =>
                onChange({ cooldown_minutes: Math.max(0, Number(e.target.value) || 0) })
              }
              className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-1.5 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
            <p className="mt-1 text-[11px] text-muted-foreground">
              Janela mínima entre disparos do mesmo trigger pro mesmo pedido. 0 =
              manda uma única vez.
            </p>
          </div>
        </div>

        <div>
          <label className="block text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Mensagem
          </label>
          <textarea
            value={row.template}
            onChange={(e) => onChange({ template: e.target.value })}
            rows={6}
            placeholder="Ex.: Olá {first_name}, seu pedido #{order_id} foi postado..."
            className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
      </CardContent>
    </Card>
  );
}

function mergeOrder(rows: NotificationTrigger[]): NotificationTrigger[] {
  const byKey = new Map(rows.map((r) => [r.event_key, r]));
  return ORDER.map(
    (key) =>
      byKey.get(key) ?? {
        event_key: key,
        enabled: false,
        template: "",
        session: "",
        cooldown_minutes: 60,
      },
  );
}
