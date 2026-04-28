"use client";

import { Bell, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/theme-toggle";

export function Topbar() {
  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl lg:px-6">
      {/* Search */}
      <div className="relative ml-auto w-full max-w-sm">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="search"
          placeholder="Buscar..."
          className="h-9 w-full rounded-lg border border-input bg-secondary/40 pl-9 pr-12 text-sm placeholder:text-muted-foreground focus-visible:bg-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
        <kbd className="pointer-events-none absolute right-2 top-1/2 hidden h-5 -translate-y-1/2 select-none items-center gap-1 rounded border border-border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground sm:inline-flex">
          <span className="text-xs">⌘</span>K
        </kbd>
      </div>

      <Button variant="ghost" size="icon" className="rounded-full" aria-label="Notificações">
        <Bell className="size-4" />
      </Button>

      <ThemeToggle />

      {/* User */}
      <div className="flex items-center gap-3 pl-2">
        <div className="flex size-8 items-center justify-center rounded-full bg-gradient-to-br from-violet-500 to-purple-600 text-xs font-semibold text-white shadow-sm">
          DK
        </div>
        <div className="hidden text-sm md:block">
          <div className="font-medium leading-none">Diego Kennedy</div>
          <div className="text-xs text-muted-foreground">Admin</div>
        </div>
      </div>
    </header>
  );
}
