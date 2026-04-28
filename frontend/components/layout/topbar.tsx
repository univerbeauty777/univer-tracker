"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Search } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";

export function Topbar() {
  const router = useRouter();
  const params = useSearchParams();
  const initial = params.get("q") ?? "";
  const [value, setValue] = useState(initial);

  // Keep input in sync if URL changes externally (filter chips, etc.).
  useEffect(() => {
    setValue(initial);
  }, [initial]);

  function submit() {
    const sp = new URLSearchParams(params.toString());
    const t = value.trim();
    if (t) sp.set("q", t);
    else sp.delete("q");
    sp.delete("offset");
    router.push(`/?${sp.toString()}`);
  }

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl lg:px-6">
      <form
        className="relative ml-auto w-full max-w-md"
        onSubmit={(e) => {
          e.preventDefault();
          submit();
        }}
      >
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="search"
          placeholder="Buscar pedido, cliente ou código de rastreio…"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          className="h-9 w-full rounded-lg border border-input bg-secondary/40 pl-9 pr-3 text-sm placeholder:text-muted-foreground focus-visible:bg-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </form>

      <ThemeToggle />
    </header>
  );
}
