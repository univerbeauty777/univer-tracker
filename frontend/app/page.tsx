import { Card, CardContent } from "@/components/ui/card";
import { Loader2 } from "lucide-react";

export default function OrdersPage() {
  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div>
        <h1 className="font-display text-3xl font-semibold tracking-tight">
          Pedidos
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Pedidos do WooCommerce com status de entrega cruzado com a Frenet.
        </p>
      </div>

      <Card>
        <CardContent className="flex h-64 items-center justify-center text-sm text-muted-foreground">
          <Loader2 className="mr-2 size-4 animate-spin" />
          Conectando ao backend…
        </CardContent>
      </Card>
    </div>
  );
}
