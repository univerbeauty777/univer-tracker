"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Loader2, RefreshCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateOrderStatus } from "@/lib/api";

const OPTIONS: { value: string; label: string }[] = [
  { value: "pending", label: "Aguardando pagamento" },
  { value: "processing", label: "Processando" },
  { value: "separacao", label: "Em separação" },
  { value: "on-hold", label: "Em espera" },
  { value: "enviado", label: "Enviado" },
  { value: "em-rota", label: "Saiu para entrega" },
  { value: "entregue", label: "Entregue" },
  { value: "completed", label: "Concluído" },
  { value: "cancelled", label: "Cancelado" },
  { value: "refunded", label: "Estornado" },
  { value: "failed", label: "Falhou" },
  { value: "retornado", label: "Retornado" },
];

export function ChangeStatusAction({
  orderId,
  currentStatus,
}: {
  orderId: number;
  currentStatus: string;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [status, setStatus] = useState(currentStatus);
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  function submit() {
    setError(null);
    startTransition(async () => {
      try {
        await updateOrderStatus(orderId, status, note.trim() || undefined);
        setOpen(false);
        setNote("");
        router.refresh();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Erro ao atualizar");
      }
    });
  }

  return (
    <>
      <Button size="sm" variant="outline" onClick={() => setOpen(true)}>
        <RefreshCcw className="size-3.5" />
        Alterar status
      </Button>

      {open ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
          onClick={(e) => {
            if (e.target === e.currentTarget) setOpen(false);
          }}
        >
          <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-2xl">
            <div className="mb-4">
              <h2 className="font-display text-lg font-semibold">Alterar status do pedido</h2>
              <p className="text-xs text-muted-foreground">
                A mudança é aplicada direto no WooCommerce.
              </p>
            </div>

            <label className="block text-xs font-medium text-muted-foreground">
              Novo status
            </label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>

            <label className="mt-4 block text-xs font-medium text-muted-foreground">
              Nota interna (opcional)
            </label>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              rows={3}
              placeholder="Ex.: trocado SEDEX por PAC após contato com cliente"
              className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />

            {error ? (
              <div className="mt-3 rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {error}
              </div>
            ) : null}

            <div className="mt-5 flex justify-end gap-2">
              <Button variant="ghost" size="sm" onClick={() => setOpen(false)} disabled={pending}>
                Cancelar
              </Button>
              <Button size="sm" onClick={submit} disabled={pending || status === currentStatus}>
                {pending ? <Loader2 className="size-3.5 animate-spin" /> : null}
                Confirmar
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
