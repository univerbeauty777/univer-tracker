import {
  Package,
  Clock,
  Truck,
  CheckCircle2,
  RotateCcw,
  TrendingUp,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

const stats = [
  {
    label: "Total envios",
    value: "5",
    icon: Package,
    tone: "text-foreground",
    delta: "+12%",
  },
  {
    label: "Pendentes",
    value: "1",
    icon: Clock,
    tone: "text-warning",
    delta: "-3%",
  },
  {
    label: "Em trânsito",
    value: "2",
    icon: Truck,
    tone: "text-info",
    delta: "+8%",
  },
  {
    label: "Entregues",
    value: "1",
    icon: CheckCircle2,
    tone: "text-success",
    delta: "+24%",
  },
  {
    label: "Devolvidos",
    value: "1",
    icon: RotateCcw,
    tone: "text-destructive",
    delta: "0%",
  },
];

const recent = [
  {
    code: "BR123456789BR",
    customer: "Cliente Alpha Ltda",
    carrier: "Correios",
    status: "Em Trânsito",
    statusVariant: "info" as const,
    eta: "2026-05-01",
    weight: "2.5kg",
  },
  {
    code: "JD987654321BR",
    customer: "Tech Solutions SA",
    carrier: "Jadlog",
    status: "Entregue",
    statusVariant: "success" as const,
    eta: "2026-04-26",
    weight: "1.2kg",
  },
  {
    code: "LG456789123BR",
    customer: "Comércio Bom Preço",
    carrier: "Loggi",
    status: "Pendente",
    statusVariant: "warning" as const,
    eta: "2026-05-05",
    weight: "8.0kg",
  },
  {
    code: "DH111222333BR",
    customer: "Global Import Export",
    carrier: "DHL",
    status: "Em Trânsito",
    statusVariant: "info" as const,
    eta: "2026-05-10",
    weight: "15kg",
  },
  {
    code: "FX999888777BR",
    customer: "Startup Inovação",
    carrier: "FedEx",
    status: "Devolvido",
    statusVariant: "destructive" as const,
    eta: "2026-04-23",
    weight: "0.5kg",
  },
];

export default function DashboardPage() {
  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      {/* Header */}
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="font-display text-3xl font-semibold tracking-tight">
            Painel Logístico
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Visão geral em tempo real das operações de envio.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm">
            <RotateCcw className="size-4" />
            Atualizar
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-violet-500 to-purple-600 text-white shadow-glow hover:opacity-90">
            <Package className="size-4" />
            Novo Envio
          </Button>
        </div>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5">
        {stats.map((stat) => (
          <Card key={stat.label} className="overflow-hidden">
            <CardContent className="p-5">
              <div className="flex items-start justify-between">
                <div className="flex size-9 items-center justify-center rounded-lg bg-secondary">
                  <stat.icon className={`size-[18px] ${stat.tone}`} strokeWidth={2} />
                </div>
                <Badge
                  variant="outline"
                  className="gap-1 border-border/40 text-[10px] text-muted-foreground"
                >
                  <TrendingUp className="size-3" />
                  {stat.delta}
                </Badge>
              </div>
              <div className="mt-3">
                <div className={`font-display text-3xl font-semibold ${stat.tone}`}>
                  {stat.value}
                </div>
                <div className="mt-0.5 text-xs text-muted-foreground">
                  {stat.label}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Charts row */}
      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Mapa de Envios</CardTitle>
            <CardDescription>Volume diário — últimos 7 dias</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex h-72 items-center justify-center text-sm text-muted-foreground">
              Gráfico será renderizado aqui (recharts)
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Por Transportadora</CardTitle>
            <CardDescription>Distribuição do mês</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex h-72 items-center justify-center text-sm text-muted-foreground">
              Donut chart aqui
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent shipments table */}
      <Card>
        <CardHeader>
          <CardTitle>Envios Recentes</CardTitle>
          <CardDescription>Últimos 5 envios processados</CardDescription>
        </CardHeader>
        <CardContent className="px-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/60 text-xs uppercase tracking-wide text-muted-foreground">
                  <th className="px-6 py-3 text-left font-medium">Rastreio</th>
                  <th className="px-6 py-3 text-left font-medium">Destinatário</th>
                  <th className="px-6 py-3 text-left font-medium">Transportadora</th>
                  <th className="px-6 py-3 text-left font-medium">Status</th>
                  <th className="px-6 py-3 text-left font-medium">Previsão</th>
                  <th className="px-6 py-3 text-left font-medium">Peso</th>
                </tr>
              </thead>
              <tbody>
                {recent.map((row) => (
                  <tr
                    key={row.code}
                    className="border-b border-border/40 transition-colors last:border-0 hover:bg-muted/30"
                  >
                    <td className="px-6 py-4 font-mono text-xs font-medium">
                      {row.code}
                    </td>
                    <td className="px-6 py-4">{row.customer}</td>
                    <td className="px-6 py-4 text-muted-foreground">
                      {row.carrier}
                    </td>
                    <td className="px-6 py-4">
                      <Badge variant={row.statusVariant}>{row.status}</Badge>
                    </td>
                    <td className="px-6 py-4 text-muted-foreground">{row.eta}</td>
                    <td className="px-6 py-4 text-muted-foreground">{row.weight}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
