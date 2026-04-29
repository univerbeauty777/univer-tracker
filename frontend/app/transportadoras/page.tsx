import { Truck } from "lucide-react";
import { fetchOverview } from "@/lib/api";
import { fmtHours } from "@/lib/format-hours";
import type { CarrierStats } from "@/lib/types";

export const dynamic = "force-dynamic";

const COLOR: Record<string, string> = {
  "Correios - PAC": "#ef4444",
  "Correios - Sedex": "#f97316",
  Correios: "#ef4444",
  "Jadlog (Melhor Envio)": "#f59e0b",
  Jadlog: "#f59e0b",
  Loggi: "#10b981",
  DHL: "#3b82f6",
  FedEx: "#a855f7",
  "Azul Cargo": "#0ea5e9",
  Motoboy: "#14b8a6",
};

export default async function TransportadorasPage() {
  let carriers: CarrierStats[] = [];
  let err: string | null = null;
  try {
    const ov = await fetchOverview();
    carriers = ov.carriers ?? [];
  } catch (e) {
    err = e instanceof Error ? e.message : String(e);
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="font-display text-2xl font-semibold text-zinc-100">Transportadoras</h1>
        <p className="mt-1 text-sm text-zinc-500">
          Comparativo de performance e SLA das transportadoras integradas (últimos 30 dias).
        </p>
      </div>

      {err ? (
        <div className="rounded-xl border border-rose-500/30 bg-rose-500/5 p-5 text-sm text-rose-300">
          {err}
        </div>
      ) : null}

      {carriers.length === 0 && !err ? (
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-10 text-center text-sm text-zinc-500">
          Nenhum dado de carrier ainda. Aguardando primeiras entregas.
        </div>
      ) : null}

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {carriers.map((c) => {
          const otd = c.total > 0 ? ((c.total - c.breached) / c.total) * 100 : 0;
          return (
            <div
              key={c.carrier}
              className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span
                    className="size-7 rounded-lg"
                    style={{ background: COLOR[c.carrier] || "#71717a" }}
                  />
                  <h3 className="text-base font-semibold text-zinc-100">{c.carrier}</h3>
                </div>
                <span className="text-xs text-zinc-500">{c.total} envios</span>
              </div>

              <div className="mt-4 grid grid-cols-2 gap-3">
                <Stat
                  label="OTD"
                  value={`${otd.toFixed(1)}%`}
                  tone={otd >= 95 ? "ok" : otd >= 85 ? "warn" : "bad"}
                />
                <Stat label="Lead time" value={fmtHours(c.avg_delivery_days * 24)} />
                <Stat label="Atrasos" value={c.breached.toString()} tone={c.breached > 0 ? "bad" : "muted"} />
                <Stat
                  label="No prazo"
                  value={(c.total - c.breached).toString()}
                  tone="muted"
                />
              </div>
            </div>
          );
        })}
      </div>

      {carriers.length === 0 && !err ? null : (
        <div className="flex items-center gap-2 text-xs text-zinc-500">
          <Truck className="size-3.5" /> Dados sincronizados a cada 10 minutos com a Frenet.
        </div>
      )}
    </div>
  );
}

function Stat({
  label,
  value,
  tone = "default",
}: {
  label: string;
  value: string;
  tone?: "default" | "ok" | "warn" | "bad" | "muted";
}) {
  const cls =
    tone === "ok"
      ? "text-emerald-400"
      : tone === "warn"
      ? "text-amber-400"
      : tone === "bad"
      ? "text-rose-400"
      : tone === "muted"
      ? "text-zinc-300"
      : "text-zinc-100";
  return (
    <div>
      <p className="text-xs uppercase tracking-wider text-zinc-500">{label}</p>
      <p className={`text-2xl font-semibold ${cls}`}>{value}</p>
    </div>
  );
}
