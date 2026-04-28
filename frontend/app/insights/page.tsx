import { Sparkles, TrendingUp, AlertTriangle, Zap } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const insights = [
  {
    type: "anomalia",
    icon: AlertTriangle,
    tone: "warning" as const,
    title: "3 envios atrasados há mais de 7 dias",
    description:
      "Pedidos #29694, #29710 e #29716 estão acima do prazo médio de entrega da transportadora.",
    action: "Revisar",
  },
  {
    type: "oportunidade",
    icon: TrendingUp,
    tone: "success" as const,
    title: "Tempo médio de preparação caiu 18%",
    description:
      "Comparado ao mês passado, a equipe está despachando pedidos mais rápido. Ótimo trabalho.",
    action: "Ver detalhes",
  },
  {
    type: "automacao",
    icon: Zap,
    tone: "info" as const,
    title: "Automatize notificações de entrega",
    description:
      "67% dos clientes que receberam mensagem de 'Saiu para entrega' avaliaram positivamente.",
    action: "Configurar",
  },
];

export default function InsightsPage() {
  return (
    <div className="mx-auto max-w-[1100px] space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 shadow-glow">
          <Sparkles className="size-5 text-white" />
        </div>
        <div>
          <h1 className="font-display text-3xl font-semibold tracking-tight">
            AI Insights
          </h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Análises e recomendações baseadas nos seus dados de logística.
          </p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {insights.map((insight, i) => (
          <Card
            key={i}
            className="group relative overflow-hidden transition-all hover:shadow-elevated"
          >
            <div className="mesh-gradient absolute inset-0 opacity-0 transition-opacity group-hover:opacity-100" />
            <CardHeader className="relative">
              <div className="flex items-start justify-between">
                <div
                  className={`flex size-10 items-center justify-center rounded-lg bg-${insight.tone}/10`}
                >
                  <insight.icon
                    className={`size-5 text-${insight.tone}`}
                    strokeWidth={2}
                  />
                </div>
                <Badge variant="outline" className="text-[10px] uppercase">
                  {insight.type}
                </Badge>
              </div>
              <CardTitle className="mt-3 leading-snug">
                {insight.title}
              </CardTitle>
            </CardHeader>
            <CardContent className="relative">
              <CardDescription className="leading-relaxed">
                {insight.description}
              </CardDescription>
              <button className="mt-4 text-sm font-medium text-primary hover:underline">
                {insight.action} →
              </button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
