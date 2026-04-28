"use client";

import { useState } from "react";
import { Search, Package, Truck, MapPin, CheckCircle2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

type EventType = "shipped" | "in-transit" | "out-for-delivery" | "delivered";

const mockEvents: Array<{
  type: EventType;
  date: string;
  location: string;
  description: string;
}> = [
  {
    type: "delivered",
    date: "13/04/2026 13:22",
    location: "Brasília — DF",
    description: "Entregue ao destinatário",
  },
  {
    type: "out-for-delivery",
    date: "13/04/2026 10:17",
    location: "Brasília — DF",
    description: "Saiu para entrega",
  },
  {
    type: "in-transit",
    date: "12/04/2026 08:53",
    location: "Brasília — DF",
    description: "Em trânsito para o centro de distribuição",
  },
  {
    type: "in-transit",
    date: "07/04/2026 12:16",
    location: "Belo Horizonte — MG",
    description: "Em trânsito",
  },
  {
    type: "shipped",
    date: "27/03/2026 13:46",
    location: "Belo Horizonte — MG",
    description: "Objeto postado pelo remetente",
  },
];

const eventIcon: Record<EventType, typeof Package> = {
  shipped: Package,
  "in-transit": Truck,
  "out-for-delivery": MapPin,
  delivered: CheckCircle2,
};

const eventTone: Record<EventType, string> = {
  shipped: "text-info bg-info/10",
  "in-transit": "text-warning bg-warning/10",
  "out-for-delivery": "text-primary bg-primary/10",
  delivered: "text-success bg-success/10",
};

export default function RastreamentoPage() {
  const [code, setCode] = useState("");
  const [showResults, setShowResults] = useState(false);

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="font-display text-3xl font-semibold tracking-tight">
          Rastreamento
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Consulte qualquer código de rastreio em tempo real.
        </p>
      </div>

      {/* Search card */}
      <Card className="overflow-hidden">
        <CardContent className="p-6">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (code.trim()) setShowResults(true);
            }}
            className="flex flex-col gap-3 sm:flex-row"
          >
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="AN 815 335 210 BR"
                className="h-11 w-full rounded-lg border border-input bg-background pl-10 pr-4 font-mono text-sm placeholder:text-muted-foreground/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>
            <Button
              type="submit"
              size="lg"
              className="bg-gradient-to-r from-violet-500 to-purple-600 text-white shadow-glow hover:opacity-90"
            >
              Rastrear
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Results */}
      {showResults && (
        <>
          {/* Status header */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <div>
                <CardDescription>Status atual</CardDescription>
                <CardTitle className="mt-1 flex items-center gap-2 font-display text-2xl">
                  Entregue
                </CardTitle>
              </div>
              <Badge variant="success">Concluído</Badge>
            </CardHeader>
            <CardContent className="space-y-2 border-t border-border/40 pt-4 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Código</span>
                <span className="font-mono font-medium">
                  {code || "AN 815 335 210 BR"}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Transportadora</span>
                <span>Correios — Sedex</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Origem → Destino</span>
                <span>Belo Horizonte — Brasília</span>
              </div>
            </CardContent>
          </Card>

          {/* Timeline */}
          <Card>
            <CardHeader>
              <CardTitle>Histórico de movimentação</CardTitle>
              <CardDescription>
                {mockEvents.length} eventos registrados
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ol className="relative space-y-6 before:absolute before:left-[15px] before:top-2 before:h-[calc(100%-1rem)] before:w-px before:bg-border">
                {mockEvents.map((event, i) => {
                  const Icon = eventIcon[event.type];
                  const tone = eventTone[event.type];
                  return (
                    <li key={i} className="flex gap-4">
                      <div
                        className={`relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full ${tone}`}
                      >
                        <Icon className="size-4" strokeWidth={2} />
                      </div>
                      <div className="flex-1 pt-0.5">
                        <div className="font-medium">{event.description}</div>
                        <div className="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
                          <span>{event.date}</span>
                          <span className="size-1 rounded-full bg-muted-foreground/40" />
                          <span>{event.location}</span>
                        </div>
                      </div>
                    </li>
                  );
                })}
              </ol>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
