export const dynamic = "force-dynamic";

const POLICIES: { carrier: string; color: string; stages: number[] }[] = [
  { carrier: "Correios - PAC", color: "#ef4444", stages: [4, 12, 24, 48, 120, 168] },
  { carrier: "Correios - SEDEX", color: "#f97316", stages: [2, 8, 16, 24, 72, 96] },
  { carrier: "Jadlog", color: "#f59e0b", stages: [4, 12, 24, 36, 72, 96] },
  { carrier: "Loggi", color: "#10b981", stages: [4, 12, 24, 36, 60, 72] },
  { carrier: "DHL", color: "#3b82f6", stages: [2, 6, 12, 24, 48, 72] },
  { carrier: "FedEx", color: "#a855f7", stages: [2, 6, 12, 24, 48, 72] },
  { carrier: "J&T Express", color: "#ec4899", stages: [4, 12, 24, 48, 96, 120] },
  { carrier: "Motoboy", color: "#14b8a6", stages: [1, 2, 3, 4, 6, 8] },
];

const HEADERS = [
  "Etiqueta",
  "Preparação",
  "Coleta",
  "Postado",
  "Saiu p/ entrega",
  "Entrega total",
];

export default function SlaPage() {
  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="font-display text-2xl font-semibold text-zinc-100">
          Configuração de SLA
        </h1>
        <p className="mt-1 text-sm text-zinc-500">
          Prazos cumulativos em horas, contados a partir do pedido pago.
        </p>
      </div>

      <div className="overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900/50">
        <table className="w-full text-sm">
          <thead className="bg-zinc-900/80">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                Transportadora
              </th>
              {HEADERS.map((h) => (
                <th
                  key={h}
                  className="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-zinc-500"
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800">
            {POLICIES.map((p) => (
              <tr key={p.carrier} className="hover:bg-zinc-900/40">
                <td className="px-4 py-3 font-medium">
                  <div className="flex items-center gap-2 text-zinc-200">
                    <span className="size-3 rounded-full" style={{ background: p.color }} />
                    {p.carrier}
                  </div>
                </td>
                {p.stages.map((h, i) => (
                  <td
                    key={i}
                    className={`px-4 py-3 text-right ${
                      i === p.stages.length - 1
                        ? "font-semibold text-violet-400"
                        : "text-zinc-300"
                    }`}
                  >
                    {h}h
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <h3 className="mb-2 text-base font-semibold text-zinc-100">Como o sistema avalia</h3>
        <ul className="space-y-2 text-sm text-zinc-400">
          <li className="flex gap-2">
            <span className="text-emerald-400">●</span>
            <span>
              <strong className="text-zinc-200">ON_TRACK</strong> — tempo decorrido &lt; 80% do
              SLA total e nenhuma etapa atrasada.
            </span>
          </li>
          <li className="flex gap-2">
            <span className="text-amber-400">●</span>
            <span>
              <strong className="text-zinc-200">AT_RISK</strong> — entre 80% e 100% do SLA, ainda
              sem violação confirmada.
            </span>
          </li>
          <li className="flex gap-2">
            <span className="text-rose-400">●</span>
            <span>
              <strong className="text-zinc-200">BREACHED</strong> — alguma etapa específica violou
              seu prazo OU o prazo total expirou.
            </span>
          </li>
        </ul>
        <p className="mt-3 text-xs text-zinc-500">
          A Frenet é consultada a cada 10 minutos. Cada evento recebido reavalia o SLA do envio
          imediatamente.
        </p>
      </div>
    </div>
  );
}
