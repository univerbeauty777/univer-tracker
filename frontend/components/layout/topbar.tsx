"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Bell, Search } from "lucide-react";

export function Topbar() {
  const router = useRouter();
  const params = useSearchParams();
  const initial = params.get("q") ?? "";
  const [value, setValue] = useState(initial);

  useEffect(() => {
    setValue(initial);
  }, [initial]);

  function submit() {
    const t = value.trim();
    router.push(t ? `/envios?q=${encodeURIComponent(t)}` : "/envios");
  }

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-zinc-900 bg-zinc-950/80 px-6 backdrop-blur-xl">
      <form
        className="relative w-full max-w-md"
        onSubmit={(e) => {
          e.preventDefault();
          submit();
        }}
      >
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-600" />
        <input
          type="search"
          placeholder="Buscar envio, cliente, transportadora…"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          className="h-9 w-full rounded-lg border border-zinc-800 bg-zinc-900/50 pl-9 pr-3 text-sm placeholder:text-zinc-600 focus:border-violet-500 focus:outline-none"
        />
      </form>

      <div className="ml-auto flex items-center gap-3">
        <button className="rounded-lg p-2 text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100">
          <Bell className="size-4" />
        </button>
        <div className="flex items-center gap-2 rounded-lg bg-zinc-900 px-2 py-1">
          <div className="grid size-7 place-items-center rounded-full gradient-brand text-xs font-semibold text-white">
            U
          </div>
          <span className="pr-2 text-sm">Operação</span>
        </div>
      </div>
    </header>
  );
}
