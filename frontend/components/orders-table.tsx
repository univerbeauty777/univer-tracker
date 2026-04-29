"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowDown, ArrowUp, ArrowUpDown, ChevronRight } from "lucide-react";
import { StatusBadge } from "@/components/status-badge";
import { SlaBadge } from "@/components/sla-badge";
import { TagList } from "@/components/tag-chip";
import { dedupeName, formatBRL, formatDate, formatRelative } from "@/lib/format";
import type { OrderListItem, SLAState } from "@/lib/types";
import { cn } from "@/lib/utils";

type SortKey = "created_at" | "total" | "customer_name" | "last_event";
type SortDir = "asc" | "desc";

export function OrdersTable({
  orders,
  total,
  limit,
  offset,
}: {
  orders: OrderListItem[];
  total: number;
  limit: number;
  offset: number;
}) {
  const router = useRouter();
  const params = useSearchParams();
  const sort = (params.get("sort") || "created_at") as SortKey;
  const dir = (params.get("dir") || "desc") as SortDir;

  function setSort(k: SortKey) {
    const sp = new URLSearchParams(params.toString());
    if (sort === k) {
      sp.set("dir", dir === "asc" ? "desc" : "asc");
    } else {
      sp.set("sort", k);
      sp.set("dir", k === "total" || k === "created_at" ? "desc" : "asc");
    }
    sp.delete("offset");
    router.push(`/?${sp.toString()}`);
  }

  function setOffset(next: number) {
    const sp = new URLSearchParams(params.toString());
    if (next <= 0) sp.delete("offset");
    else sp.set("offset", String(next));
    router.push(`/?${sp.toString()}`);
  }

  const start = total === 0 ? 0 : offset + 1;
  const end = Math.min(offset + orders.length, total);

  return (
    <div className="overflow-hidden rounded-xl border border-border/60 bg-card">
      <div className="flex items-center justify-between border-b border-border/60 px-4 py-2.5 text-xs text-muted-foreground">
        <span>
          {total > 0 ? (
            <>
              <span className="font-medium text-foreground">{start}–{end}</span> de {total} pedidos
            </>
          ) : (
            "Nenhum pedido"
          )}
        </span>
        <div className="flex items-center gap-1">
          <button
            disabled={offset === 0}
            onClick={() => setOffset(Math.max(0, offset - limit))}
            className="rounded px-2 py-1 hover:bg-muted disabled:opacity-30"
          >
            ← anterior
          </button>
          <button
            disabled={end >= total}
            onClick={() => setOffset(offset + limit)}
            className="rounded px-2 py-1 hover:bg-muted disabled:opacity-30"
          >
            próximo →
          </button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border/60 text-[11px] uppercase tracking-wide text-muted-foreground">
              <Th label="Pedido" sortKey="created_at" sort={sort} dir={dir} onSort={setSort} />
              <Th label="Cliente" sortKey="customer_name" sort={sort} dir={dir} onSort={setSort} />
              <th className="px-4 py-3 text-left font-medium">SLA</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Tags</th>
              <Th label="Último evento" sortKey="last_event" sort={sort} dir={dir} onSort={setSort} />
              <Th label="Total" sortKey="total" sort={sort} dir={dir} onSort={setSort} align="right" />
              <th className="w-10" />
            </tr>
          </thead>
          <tbody>
            {orders.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-6 py-16 text-center text-sm text-muted-foreground">
                  Nenhum pedido bate com esses filtros.
                </td>
              </tr>
            ) : (
              orders.map((o) => <Row key={o.id} order={o} />)
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function Th({
  label,
  sortKey,
  sort,
  dir,
  onSort,
  align,
}: {
  label: string;
  sortKey: SortKey;
  sort: SortKey;
  dir: SortDir;
  onSort: (k: SortKey) => void;
  align?: "left" | "right";
}) {
  const active = sort === sortKey;
  return (
    <th className={cn("px-4 py-3 font-medium", align === "right" ? "text-right" : "text-left")}>
      <button
        onClick={() => onSort(sortKey)}
        className={cn(
          "inline-flex items-center gap-1 rounded transition-colors hover:text-foreground",
          active && "text-foreground",
        )}
      >
        {label}
        {active ? (
          dir === "asc" ? <ArrowUp className="size-3" /> : <ArrowDown className="size-3" />
        ) : (
          <ArrowUpDown className="size-3 opacity-30" />
        )}
      </button>
    </th>
  );
}

function Row({ order: o }: { order: OrderListItem }) {
  const idleDays = o.tracking.last_event_at
    ? Math.max(0, Math.floor((Date.now() - new Date(o.tracking.last_event_at).getTime()) / 86_400_000))
    : null;
  const customer = dedupeName(o.customer_name);
  return (
    <tr className="border-b border-border/40 transition-colors last:border-0 hover:bg-muted/40">
      <td className="px-4 py-3">
        <Link href={`/pedidos/${o.id}`} className="font-mono text-xs font-medium hover:underline">
          #{o.id}
        </Link>
        <div className="mt-0.5 text-[11px] text-muted-foreground" title={formatDate(o.created_at)}>
          {formatRelative(o.created_at)}
        </div>
      </td>
      <td className="px-4 py-3">
        <div className="font-medium text-foreground">{customer || "—"}</div>
        <div className="mt-0.5 text-[11px] text-muted-foreground">
          {[o.customer_city, o.customer_state].filter(Boolean).join(" · ") || "—"}
        </div>
      </td>
      <td className="px-4 py-3">
        <SlaBadge state={o.tracking.sla_state as SLAState | undefined} />
        {o.tracking.estimated_delivery ? (
          <div className="mt-1 text-[11px] text-muted-foreground">
            ETA · {formatDate(o.tracking.estimated_delivery)}
          </div>
        ) : null}
      </td>
      <td className="px-4 py-3">
        <StatusBadge status={o.status} label={o.status_label} />
        {o.tracking.status !== "unknown" ? (
          <div className="mt-0.5 text-[11px] text-muted-foreground">{o.tracking.status_label}</div>
        ) : null}
      </td>
      <td className="px-4 py-3">
        <TagList tags={o.tags} />
      </td>
      <td className="px-4 py-3">
        {o.tracking.last_event ? (
          <div>
            <div className="line-clamp-1 max-w-[280px] text-xs">{o.tracking.last_event}</div>
            <div className="text-[11px] text-muted-foreground" title={formatDate(o.tracking.last_event_at)}>
              {formatRelative(o.tracking.last_event_at)}
              {idleDays !== null && idleDays >= 4 ? (
                <span className="ml-1 text-warning">· {idleDays}d sem evento</span>
              ) : null}
            </div>
          </div>
        ) : o.tracking.number ? (
          <span className="text-xs text-muted-foreground">aguardando primeira leitura</span>
        ) : (
          <span className="text-xs text-muted-foreground">sem código</span>
        )}
      </td>
      <td className="px-4 py-3 text-right font-medium">{formatBRL(o.total)}</td>
      <td className="px-3 py-3 text-muted-foreground">
        <Link href={`/pedidos/${o.id}`} aria-label={`Abrir pedido ${o.id}`}>
          <ChevronRight className="size-4" />
        </Link>
      </td>
    </tr>
  );
}
