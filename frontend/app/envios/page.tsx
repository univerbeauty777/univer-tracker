"use client";

import { useState } from "react";
import { Package } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type ShipmentStatus = "pendente" | "em-transito" | "entregue" | "devolvido";

type Shipment = {
  code: string;
  customer: string;
  address: string;
  carrier: string;
  status: ShipmentStatus;
  eta: string;
  weight: string;
};

const shipments: Shipment[] = [
  {
    code: "BR123456789BR",
    customer: "Cliente Alpha Ltda",
    address: "Av Paulista 1000, SP",
    carrier: "Correios",
    status: "em-transito",
    eta: "2026-05-01",
    weight: "2.5kg",
  },
  {
    code: "JD987654321BR",
    customer: "Tech Solutions SA",
    address: "R Augusta 500, SP",
    carrier: "Jadlog",
    status: "entregue",
    eta: "2026-04-26",
    weight: "1.2kg",
  },
  {
    code: "LG456789123BR",
    customer: "Comércio Bom Preço",
    address: "R 25 de Março 300, SP",
    carrier: "Loggi",
    status: "pendente",
    eta: "2026-05-05",
    weight: "8.0kg",
  },
  {
    code: "DH111222333BR",
    customer: "Global Import Export",
    address: "Porto Santos, SP",
    carrier: "DHL",
    status: "em-transito",
    eta: "2026-05-10",
    weight: "15kg",
  },
  {
    code: "FX999888777BR",
    customer: "Startup Inovação",
    address: "R Faria Lima 2000, SP",
    carrier: "FedEx",
    status: "devolvido",
    eta: "2026-04-23",
    weight: "0.5kg",
  },
];

const TABS = [
  { key: "todos", label: "Todos" },
  { key: "pendente", label: "Pendentes" },
  { key: "em-transito", label: "Em Trânsito" },
  { key: "entregue", label: "Entregues" },
  { key: "devolvido", label: "Devolvidos" },
] as const;

type TabKey = (typeof TABS)[number]["key"];

const statusConfig: Record<
  ShipmentStatus,
  { label: string; variant: "info" | "success" | "warning" | "destructive" }
> = {
  pendente: { label: "Pendente", variant: "warning" },
  "em-transito": { label: "Em Trânsito", variant: "info" },
  entregue: { label: "Entregue", variant: "success" },
  devolvido: { label: "Devolvido", variant: "destructive" },
};

export default function EnviosPage() {
  const [tab, setTab] = useState<TabKey>("todos");

  const filtered =
    tab === "todos"
      ? shipments
      : shipments.filter((s) => s.status === tab);

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      {/* Header */}
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="font-display text-3xl font-semibold tracking-tight">
            Envios
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Gerencie todos os envios da operação.
          </p>
        </div>
        <Button
          size="sm"
          className="bg-gradient-to-r from-violet-500 to-purple-600 text-white shadow-glow hover:opacity-90"
        >
          <Package className="size-4" />
          Novo Envio
        </Button>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-border/60">
        {TABS.map((t) => {
          const active = tab === t.key;
          return (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={cn(
                "relative px-4 py-2.5 text-sm font-medium transition-colors",
                active
                  ? "text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {t.label}
              {active && (
                <span className="absolute inset-x-3 -bottom-px h-0.5 rounded-full bg-gradient-to-r from-violet-500 to-purple-600" />
              )}
            </button>
          );
        })}
      </div>

      {/* Table */}
      <Card>
        <CardContent className="px-0 py-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/60 text-xs uppercase tracking-wide text-muted-foreground">
                  <th className="px-6 py-3 text-left font-medium">Rastreio</th>
                  <th className="px-6 py-3 text-left font-medium">
                    Destinatário
                  </th>
                  <th className="px-6 py-3 text-left font-medium">Endereço</th>
                  <th className="px-6 py-3 text-left font-medium">
                    Transportadora
                  </th>
                  <th className="px-6 py-3 text-left font-medium">Status</th>
                  <th className="px-6 py-3 text-left font-medium">Previsão</th>
                  <th className="px-6 py-3 text-left font-medium">Peso</th>
                </tr>
              </thead>
              <tbody>
                {filtered.length === 0 ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-6 py-16 text-center text-sm text-muted-foreground"
                    >
                      Nenhum envio encontrado nesta categoria.
                    </td>
                  </tr>
                ) : (
                  filtered.map((row) => {
                    const cfg = statusConfig[row.status];
                    return (
                      <tr
                        key={row.code}
                        className="border-b border-border/40 transition-colors last:border-0 hover:bg-muted/30"
                      >
                        <td className="px-6 py-4 font-mono text-xs font-medium">
                          {row.code}
                        </td>
                        <td className="px-6 py-4">{row.customer}</td>
                        <td className="px-6 py-4 text-muted-foreground">
                          {row.address}
                        </td>
                        <td className="px-6 py-4 text-muted-foreground">
                          {row.carrier}
                        </td>
                        <td className="px-6 py-4">
                          <Badge variant={cfg.variant}>{cfg.label}</Badge>
                        </td>
                        <td className="px-6 py-4 text-muted-foreground">
                          {row.eta}
                        </td>
                        <td className="px-6 py-4 text-muted-foreground">
                          {row.weight}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
