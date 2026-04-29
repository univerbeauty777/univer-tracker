"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Activity,
  AlertTriangle,
  LayoutGrid,
  MapPin,
  Package,
  Settings,
  Truck,
  Zap,
} from "lucide-react";
import { cn } from "@/lib/utils";

type NavItem = { label: string; href: string; icon: typeof LayoutGrid };
type Section = { title?: string; items: NavItem[] };

const NAV: Section[] = [
  {
    items: [{ label: "Painel", href: "/", icon: LayoutGrid }],
  },
  {
    title: "Logística",
    items: [
      { label: "Envios", href: "/envios", icon: Package },
      { label: "Análise de gargalos", href: "/gargalos", icon: Zap },
      { label: "Transportadoras", href: "/transportadoras", icon: Truck },
      { label: "SLA por etapa", href: "/sla", icon: Activity },
    ],
  },
  {
    title: "Sistema",
    items: [
      { label: "Configurações", href: "/configuracoes/integracoes", icon: Settings },
    ],
  },
];

export function Sidebar() {
  const pathname = usePathname() ?? "/";

  return (
    <aside className="sticky top-0 z-40 hidden h-screen w-64 shrink-0 flex-col self-start border-r border-sidebar-border bg-sidebar text-sidebar-foreground lg:flex">
      <div className="flex h-16 shrink-0 items-center gap-2.5 border-b border-sidebar-border px-5">
        <div className="grid size-8 place-items-center rounded-lg gradient-brand glow-violet">
          <MapPin className="size-[18px] text-white" strokeWidth={2.2} />
        </div>
        <div className="flex flex-col leading-none">
          <span className="text-[15px] font-semibold tracking-tight">rastreiaki</span>
          <span className="mt-0.5 text-[10px] uppercase tracking-[0.14em] text-zinc-500">
            SLA logístico
          </span>
        </div>
      </div>

      <nav className="flex-1 overflow-y-auto px-3 py-4">
        {NAV.map((section, i) => (
          <div key={i} className={i === 0 ? "" : "mt-5"}>
            {section.title ? (
              <div className="mb-1.5 px-3 text-[10px] font-medium uppercase tracking-wider text-zinc-500">
                {section.title}
              </div>
            ) : null}
            <ul className="space-y-0.5">
              {section.items.map((item) => {
                const active =
                  pathname === item.href ||
                  (item.href !== "/" && pathname.startsWith(item.href));
                return (
                  <li key={item.href}>
                    <Link
                      href={item.href}
                      className={cn(
                        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
                        active
                          ? "bg-zinc-900 text-zinc-100"
                          : "text-zinc-400 hover:bg-zinc-900/60 hover:text-zinc-100",
                      )}
                    >
                      <item.icon className="size-4 shrink-0" strokeWidth={2} />
                      {item.label}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>

      <div className="p-3">
        <div className="rounded-lg border border-violet-500/20 bg-gradient-to-br from-violet-500/10 to-fuchsia-500/5 p-3">
          <div className="flex items-center gap-1.5">
            <span className="inline-block size-1.5 animate-pulse rounded-full bg-violet-400" />
            <p className="text-xs font-medium text-violet-300">Operação ao vivo</p>
          </div>
          <p className="mt-1.5 text-xs leading-relaxed text-zinc-500">
            Sincronização contínua com WooCommerce e Frenet. SLA recalculado a cada evento.
          </p>
        </div>
      </div>
    </aside>
  );
}
