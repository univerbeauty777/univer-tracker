"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Loader2, MessageCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { notifyOrder } from "@/lib/api";

const TEMPLATES: { key: string; label: string; build: (ctx: NotifyContext) => string }[] = [
  {
    key: "shipped",
    label: "Postado",
    build: (c) => `Olá ${c.firstName}, seu pedido #${c.orderId} foi postado! 📦\n\nCódigo de rastreio: ${c.tracking}\nAcompanhe: ${c.trackUrl ?? "(link em breve)"}\n\nQualquer dúvida, é só chamar!`,
  },
  {
    key: "in_transit",
    label: "Em trânsito",
    build: (c) => `Oi ${c.firstName}! Seu pedido #${c.orderId} está a caminho 🚛\n\nÚltimo evento: ${c.lastEvent ?? "em transferência"}\nPrevisão de entrega: ${c.eta ?? "em breve"}\n\nRastreie aqui: ${c.trackUrl ?? ""}`,
  },
  {
    key: "delivered",
    label: "Entregue",
    build: (c) => `${c.firstName}, seu pedido #${c.orderId} foi entregue! ✅\n\nEsperamos que você ame seus produtos. Qualquer coisa, conta com a gente.`,
  },
  {
    key: "delayed",
    label: "Atrasado",
    build: (c) => `Oi ${c.firstName}, queria te avisar que seu pedido #${c.orderId} está com atraso na transportadora.\n\nÚltimo evento: ${c.lastEvent ?? "—"}\n\nJá estamos em contato com a transportadora pra agilizar. Te aviso assim que tiver novidade.`,
  },
];

export interface NotifyContext {
  orderId: number;
  firstName: string;
  tracking: string;
  trackUrl?: string;
  lastEvent?: string;
  eta?: string;
}

export function NotifyAction({ context, hasPhone }: { context: NotifyContext; hasPhone: boolean }) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [template, setTemplate] = useState<string>(TEMPLATES[0].key);
  const [message, setMessage] = useState<string>(TEMPLATES[0].build(context));
  const [pending, start] = useTransition();
  const [result, setResult] = useState<{ ok: boolean; text: string } | null>(null);

  function pickTemplate(key: string) {
    setTemplate(key);
    const t = TEMPLATES.find((x) => x.key === key);
    if (t) setMessage(t.build(context));
  }

  function send() {
    start(async () => {
      try {
        const r = await notifyOrder(context.orderId, message, template);
        setResult({ ok: r.ok, text: r.message ?? r.error ?? (r.ok ? "Enviado" : "Falhou") });
        if (r.ok) {
          setTimeout(() => {
            setOpen(false);
            setResult(null);
            router.refresh();
          }, 1500);
        }
      } catch (e) {
        setResult({ ok: false, text: e instanceof Error ? e.message : "erro" });
      }
    });
  }

  if (!hasPhone) {
    return (
      <Button size="sm" variant="outline" disabled title="Sem telefone cadastrado">
        <MessageCircle className="size-3.5" />
        Notificar
      </Button>
    );
  }

  return (
    <>
      <Button size="sm" variant="outline" onClick={() => setOpen(true)}>
        <MessageCircle className="size-3.5" />
        Notificar via WhatsApp
      </Button>

      {open ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
          onClick={(e) => {
            if (e.target === e.currentTarget) setOpen(false);
          }}
        >
          <div className="w-full max-w-lg rounded-xl border border-border bg-card p-6 shadow-2xl">
            <div className="mb-4">
              <h2 className="font-display text-lg font-semibold">Notificar cliente</h2>
              <p className="text-xs text-muted-foreground">
                Envia uma mensagem via WhatsApp através da integração WAHA.
              </p>
            </div>

            <div className="mb-3 flex flex-wrap gap-1.5">
              {TEMPLATES.map((t) => (
                <button
                  key={t.key}
                  onClick={() => pickTemplate(t.key)}
                  className={`rounded-full px-2.5 py-1 text-[11px] font-medium transition-colors ${
                    template === t.key
                      ? "bg-foreground text-background"
                      : "border border-border/60 text-muted-foreground hover:bg-muted"
                  }`}
                >
                  {t.label}
                </button>
              ))}
            </div>

            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={8}
              className="w-full rounded-lg border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />

            {result ? (
              <div
                className={`mt-3 rounded-md px-3 py-2 text-xs ${
                  result.ok ? "bg-success/10 text-success" : "bg-destructive/10 text-destructive"
                }`}
              >
                {result.text}
              </div>
            ) : null}

            <div className="mt-5 flex justify-end gap-2">
              <Button variant="ghost" size="sm" onClick={() => setOpen(false)} disabled={pending}>
                Cancelar
              </Button>
              <Button size="sm" onClick={send} disabled={pending || !message.trim()}>
                {pending ? <Loader2 className="size-3.5 animate-spin" /> : null}
                Enviar
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
