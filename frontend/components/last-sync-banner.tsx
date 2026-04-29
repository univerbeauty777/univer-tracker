"use client";

import { useEffect, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Loader2, RefreshCw } from "lucide-react";
import { fetchSyncStatus, triggerSync } from "@/lib/api";
import { formatRelative } from "@/lib/format";
import type { SyncSource } from "@/lib/types";
import { cn } from "@/lib/utils";

const REFRESH_MS = 30_000;

export function LastSyncBanner() {
  const router = useRouter();
  const [sources, setSources] = useState<SyncSource[] | null>(null);
  const [pending, start] = useTransition();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    async function load() {
      try {
        const s = await fetchSyncStatus();
        if (mounted) {
          setSources(s.sources);
          setError(null); // recovery — clear stale error from a previous tick
        }
      } catch (e) {
        if (mounted) setError(e instanceof Error ? e.message : "erro");
      }
    }
    load();
    const t = setInterval(load, REFRESH_MS);
    return () => {
      mounted = false;
      clearInterval(t);
    };
  }, []);

  function syncNow() {
    setError(null);
    start(async () => {
      try {
        // Snapshot timestamps to detect when the worker finishes; a
        // hard-coded 4s pause was racing the WC pull on big catalogues.
        const before = sources?.map((s) => s.last_synced_at).join("|") ?? "";
        await triggerSync();
        let attempt = 0;
        while (attempt < 12) {
          await new Promise((r) => setTimeout(r, 2500));
          try {
            const s = await fetchSyncStatus();
            const after = s.sources.map((x) => x.last_synced_at).join("|");
            if (after !== before) {
              setSources(s.sources);
              router.refresh();
              return;
            }
            // Keep showing fresh state even mid-poll.
            setSources(s.sources);
          } catch {
            /* keep polling */
          }
          attempt++;
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : "erro ao sincronizar");
      }
    });
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border/50 bg-card/40 px-4 py-2.5 text-xs">
      <div className="flex flex-wrap items-center gap-x-5 gap-y-1 text-muted-foreground">
        {sources?.map((s) => (
          <SourcePill key={s.entity} source={s} />
        )) ?? <span className="text-muted-foreground">Carregando status…</span>}
        {error ? <span className="text-destructive">{error}</span> : null}
      </div>

      <button
        onClick={syncNow}
        disabled={pending}
        className={cn(
          "inline-flex items-center gap-1.5 rounded-md border border-border/60 px-2.5 py-1 font-medium transition-colors",
          pending ? "text-muted-foreground" : "hover:bg-muted",
        )}
      >
        {pending ? <Loader2 className="size-3 animate-spin" /> : <RefreshCw className="size-3" />}
        {pending ? "Sincronizando…" : "Sincronizar agora"}
      </button>
    </div>
  );
}

function SourcePill({ source }: { source: SyncSource }) {
  const stale = source.seconds_ago < 0 || source.seconds_ago > 600;
  const label =
    source.entity === "wc_orders" ? "WooCommerce" :
    source.entity === "frenet" ? "Frenet" :
    source.entity;

  return (
    <span className="inline-flex items-center gap-1.5">
      <span
        className={cn(
          "size-1.5 rounded-full",
          source.last_synced_at == null ? "bg-muted-foreground/40" :
          stale ? "bg-warning" : "bg-success",
        )}
        aria-hidden
      />
      <span className="text-foreground/80">{label}</span>
      <span className="text-muted-foreground">
        {source.last_synced_at ? formatRelative(source.last_synced_at) : "nunca"}
      </span>
    </span>
  );
}
