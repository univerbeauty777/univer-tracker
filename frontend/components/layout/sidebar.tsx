"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Package, Settings } from "lucide-react";
import { cn } from "@/lib/utils";

const navigation = [
  { label: "Pedidos", href: "/", icon: Package },
  { label: "Configurações", href: "/configuracoes", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden lg:flex h-screen w-60 shrink-0 flex-col bg-sidebar text-sidebar-foreground border-r border-sidebar-border">
      <div className="flex h-16 items-center gap-2 px-5">
        <div className="flex size-7 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500 to-purple-600 shadow-sm">
          <Package className="size-4 text-white" strokeWidth={2.5} />
        </div>
        <span className="font-display text-base font-semibold tracking-tight">
          Univer Tracker
        </span>
      </div>

      <nav className="flex-1 overflow-y-auto px-3 pb-4">
        <ul className="space-y-0.5">
          {navigation.map((item) => {
            const active =
              pathname === item.href ||
              (item.href !== "/" && pathname?.startsWith(item.href));
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={cn(
                    "group flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all",
                    active
                      ? "bg-sidebar-accent text-sidebar-accent-foreground shadow-sm"
                      : "text-sidebar-foreground/70 hover:bg-white/5 hover:text-sidebar-foreground",
                  )}
                >
                  <item.icon className="size-[18px] shrink-0" strokeWidth={1.75} />
                  <span className="flex-1">{item.label}</span>
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
    </aside>
  );
}
