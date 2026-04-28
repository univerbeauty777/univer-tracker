"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Sparkles,
  Workflow,
  Globe,
  Zap,
  CheckSquare,
  FolderKanban,
  Clock,
  Activity,
  DollarSign,
  Users,
  Scale,
  Megaphone,
  Monitor,
  ShoppingCart,
  Briefcase,
  Building2,
  Truck,
  Package,
  MapPin,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

type NavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
  badge?: string | number;
};

type NavSection = {
  title?: string;
  items: NavItem[];
};

const navigation: NavSection[] = [
  {
    items: [
      { label: "Dashboard", href: "/", icon: LayoutDashboard },
      { label: "AI Insights", href: "/insights", icon: Sparkles },
    ],
  },
  {
    title: "Processos",
    items: [
      { label: "Pipes", href: "/pipes", icon: Workflow, badge: 8 },
      { label: "Portal", href: "/portal", icon: Globe },
      { label: "Automações", href: "/automacoes", icon: Zap },
    ],
  },
  {
    title: "Tarefas & Projetos",
    items: [
      { label: "Tarefas", href: "/tarefas", icon: CheckSquare, badge: 7 },
      { label: "Projetos", href: "/projetos", icon: FolderKanban },
      { label: "Time Tracking", href: "/tracking", icon: Clock },
      { label: "Carga de Trabalho", href: "/carga", icon: Activity },
    ],
  },
  {
    title: "Departamentos",
    items: [
      { label: "Financeiro", href: "/financeiro", icon: DollarSign },
      { label: "RH", href: "/rh", icon: Users },
      { label: "Jurídico", href: "/juridico", icon: Scale },
      { label: "Marketing", href: "/marketing", icon: Megaphone },
      { label: "TI", href: "/ti", icon: Monitor },
      { label: "Compras", href: "/compras", icon: ShoppingCart },
      { label: "Comercial", href: "/comercial", icon: Briefcase },
      { label: "Operações", href: "/operacoes", icon: Building2 },
    ],
  },
  {
    title: "Logística",
    items: [
      { label: "Painel Logístico", href: "/logistica", icon: Truck },
      { label: "Envios", href: "/envios", icon: Package },
      { label: "Rastreamento", href: "/rastreamento", icon: MapPin },
    ],
  },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden lg:flex h-screen w-60 shrink-0 flex-col bg-sidebar text-sidebar-foreground border-r border-sidebar-border">
      {/* Brand */}
      <div className="flex h-16 items-center gap-2 px-5">
        <div className="flex size-7 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500 to-purple-600 shadow-sm">
          <Package className="size-4 text-white" strokeWidth={2.5} />
        </div>
        <span className="font-display text-base font-semibold tracking-tight">
          UniverHub
        </span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 pb-4">
        {navigation.map((section, i) => (
          <div key={i} className="mb-4">
            {section.title && (
              <div className="px-2 pb-2 pt-3 text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/40">
                {section.title}
              </div>
            )}
            <ul className="space-y-0.5">
              {section.items.map((item) => {
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
                      {item.badge && (
                        <span
                          className={cn(
                            "flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 text-[10px] font-semibold",
                            active
                              ? "bg-white/20 text-white"
                              : "bg-white/10 text-sidebar-foreground/80",
                          )}
                        >
                          {item.badge}
                        </span>
                      )}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  );
}
