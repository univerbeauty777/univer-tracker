export function fmtHours(h: number | null | undefined): string {
  if (h == null || !Number.isFinite(h) || h <= 0) return "—";
  if (h < 1) return `${(h * 60).toFixed(0)}min`;
  if (h < 48) return `${h.toFixed(1)}h`;
  return `${(h / 24).toFixed(1)}d`;
}
