import { cn } from "@/lib/utils";

const TAGS: Record<string, { label: string; icon: string; cls: string }> = {
  vip:           { label: "VIP",             icon: "★",  cls: "border-yellow-500/30 bg-yellow-500/15 text-yellow-300" },
  urgente:       { label: "Urgente",         icon: "⚡", cls: "border-red-500/30 bg-red-500/15 text-red-300" },
  fragil:        { label: "Frágil",          icon: "◇",  cls: "border-orange-500/30 bg-orange-500/15 text-orange-300" },
  alto_valor:    { label: "Alto valor",      icon: "$",  cls: "border-emerald-500/30 bg-emerald-500/15 text-emerald-300" },
  reentrega:     { label: "Reentrega",       icon: "↻",  cls: "border-purple-500/30 bg-purple-500/15 text-purple-300" },
  brinde:        { label: "Brinde",          icon: "◈",  cls: "border-pink-500/30 bg-pink-500/15 text-pink-300" },
  primeira:      { label: "Primeira compra", icon: "✦",  cls: "border-blue-500/30 bg-blue-500/15 text-blue-300" },
  volumoso:      { label: "Volumoso",        icon: "▣",  cls: "border-amber-500/30 bg-amber-500/15 text-amber-300" },
  termolabil:    { label: "Termolábil",      icon: "❄",  cls: "border-cyan-500/30 bg-cyan-500/15 text-cyan-300" },
  frete_gratis:  { label: "Frete grátis",    icon: "✓",  cls: "border-emerald-500/30 bg-emerald-500/10 text-emerald-300" },
};

export function TagChip({ tag }: { tag: string }) {
  const def = TAGS[tag];
  if (!def) {
    return (
      <span className="inline-flex items-center rounded-full border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-[10px] font-medium text-zinc-400">
        {tag}
      </span>
    );
  }
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[10px] font-medium",
        def.cls,
      )}
      title={def.label}
    >
      <span className="leading-none">{def.icon}</span>
      {def.label}
    </span>
  );
}

export function TagList({ tags, max = 3 }: { tags?: string[]; max?: number }) {
  if (!tags || tags.length === 0) {
    return <span className="text-xs text-zinc-700">—</span>;
  }
  const visible = tags.slice(0, max);
  const extra = tags.length - visible.length;
  return (
    <div className="flex flex-wrap items-center gap-1">
      {visible.map((t) => (
        <TagChip key={t} tag={t} />
      ))}
      {extra > 0 ? (
        <span className="text-[10px] text-zinc-500" title={tags.slice(max).join(", ")}>
          +{extra}
        </span>
      ) : null}
    </div>
  );
}
