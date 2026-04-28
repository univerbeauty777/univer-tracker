"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useTransition } from "react";
import { Calendar, ChevronDown, Filter, Truck, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Facets } from "@/lib/types";

const HEALTH_TABS: { key: string; label: string }[] = [
  { key: "", label: "Todos" },
  { key: "at_risk", label: "Em risco" },
  { key: "breached", label: "SLA quebrado" },
  { key: "on_track", label: "No prazo" },
];

const RANGE_PRESETS: { key: string; label: string; days: number | null }[] = [
  { key: "7d", label: "7 dias", days: 7 },
  { key: "30d", label: "30 dias", days: 30 },
  { key: "90d", label: "90 dias", days: 90 },
  { key: "all", label: "Tudo", days: null },
];

export function OrdersFilterBar({ facets }: { facets: Facets }) {
  const router = useRouter();
  const params = useSearchParams();
  const [pending, start] = useTransition();

  const health = params.get("health") ?? "";
  const carrier = params.get("carrier") ?? "";
  const uf = params.get("uf") ?? "";
  const status = params.get("status") ?? "";
  const since = params.get("since") ?? "";
  const until = params.get("until") ?? "";

  function update(next: Record<string, string | null>) {
    const sp = new URLSearchParams(params.toString());
    for (const [k, v] of Object.entries(next)) {
      if (v === null || v === "") sp.delete(k);
      else sp.set(k, v);
    }
    sp.delete("offset");
    start(() => router.push(`/?${sp.toString()}`));
  }

  function applyRange(days: number | null) {
    if (days === null) {
      update({ since: null, until: null });
      return;
    }
    const end = new Date();
    const start = new Date();
    start.setDate(start.getDate() - days);
    update({
      since: start.toISOString().slice(0, 10),
      until: end.toISOString().slice(0, 10),
    });
  }

  const activeCount = [health, carrier, uf, status, since, until].filter(Boolean).length;

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        {HEALTH_TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => update({ health: t.key || null })}
            className={cn(
              "rounded-full px-3 py-1 text-xs font-medium transition-colors",
              t.key === health
                ? "bg-foreground text-background"
                : "border border-border/60 text-muted-foreground hover:bg-muted",
            )}
          >
            {t.label}
          </button>
        ))}

        <div className="mx-2 h-5 w-px bg-border/60" />

        <DateRangeButton
          label={rangeLabel(since, until)}
          onPick={applyRange}
          onClear={() => update({ since: null, until: null })}
          active={Boolean(since || until)}
        />

        <FacetSelect
          icon={<Truck className="size-3.5" />}
          label="Transportadora"
          value={carrier}
          values={facets.carriers}
          onChange={(v) => update({ carrier: v })}
        />

        <FacetSelect
          label="UF"
          value={uf}
          values={facets.ufs}
          onChange={(v) => update({ uf: v })}
        />

        <FacetSelect
          label="Status WC"
          value={status}
          values={facets.statuses}
          onChange={(v) => update({ status: v })}
        />

        {activeCount > 0 ? (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => update({ health: null, carrier: null, uf: null, status: null, since: null, until: null })}
            disabled={pending}
            className="ml-auto text-xs text-muted-foreground"
          >
            <X className="size-3" /> Limpar filtros ({activeCount})
          </Button>
        ) : null}
      </div>
    </div>
  );
}

function rangeLabel(since: string, until: string): string {
  if (!since && !until) return "Período";
  const s = since ? new Date(since).toLocaleDateString("pt-BR", { day: "2-digit", month: "short" }) : "?";
  const u = until ? new Date(until).toLocaleDateString("pt-BR", { day: "2-digit", month: "short" }) : "hoje";
  return `${s} – ${u}`;
}

function DateRangeButton({
  label,
  onPick,
  onClear,
  active,
}: {
  label: string;
  onPick: (days: number | null) => void;
  onClear: () => void;
  active: boolean;
}) {
  return (
    <details className="group relative">
      <summary
        className={cn(
          "flex cursor-pointer list-none items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors",
          active
            ? "border-foreground bg-foreground text-background"
            : "border-border/60 text-muted-foreground hover:bg-muted",
        )}
      >
        <Calendar className="size-3.5" />
        {label}
        <ChevronDown className="size-3 opacity-60" />
      </summary>
      <div className="absolute left-0 top-full z-50 mt-1 w-44 rounded-lg border border-border bg-popover p-1 text-xs shadow-elevated">
        {RANGE_PRESETS.map((p) => (
          <button
            key={p.key}
            onClick={(e) => {
              (e.currentTarget.closest("details") as HTMLDetailsElement).open = false;
              onPick(p.days);
            }}
            className="flex w-full items-center rounded px-2 py-1.5 text-left hover:bg-muted"
          >
            {p.label}
          </button>
        ))}
        {active ? (
          <button
            onClick={(e) => {
              (e.currentTarget.closest("details") as HTMLDetailsElement).open = false;
              onClear();
            }}
            className="mt-1 flex w-full items-center border-t border-border/60 px-2 py-1.5 pt-2 text-left text-destructive hover:bg-muted"
          >
            <X className="mr-1 size-3" /> Remover período
          </button>
        ) : null}
      </div>
    </details>
  );
}

function FacetSelect({
  label,
  value,
  values,
  onChange,
  icon,
}: {
  label: string;
  value: string;
  values: { value: string; count: number }[];
  onChange: (v: string | null) => void;
  icon?: React.ReactNode;
}) {
  if (values.length === 0) return null;
  const active = Boolean(value);
  return (
    <details className="group relative">
      <summary
        className={cn(
          "flex cursor-pointer list-none items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors",
          active
            ? "border-foreground bg-foreground text-background"
            : "border-border/60 text-muted-foreground hover:bg-muted",
        )}
      >
        {icon ?? <Filter className="size-3.5" />}
        {value || label}
        <ChevronDown className="size-3 opacity-60" />
      </summary>
      <div className="absolute left-0 top-full z-50 mt-1 max-h-72 w-56 overflow-y-auto rounded-lg border border-border bg-popover p-1 text-xs shadow-elevated">
        {values.map((opt) => (
          <button
            key={opt.value}
            onClick={(e) => {
              (e.currentTarget.closest("details") as HTMLDetailsElement).open = false;
              onChange(opt.value === value ? null : opt.value);
            }}
            className={cn(
              "flex w-full items-center justify-between rounded px-2 py-1.5 text-left",
              opt.value === value ? "bg-foreground text-background" : "hover:bg-muted",
            )}
          >
            <span className="truncate">{opt.value}</span>
            <span className="ml-2 text-[10px] opacity-60">{opt.count}</span>
          </button>
        ))}
      </div>
    </details>
  );
}
