export function formatBRL(value: string | number | null | undefined): string {
  if (value === null || value === undefined) return "—";
  const n = typeof value === "string" ? Number(value) : value;
  if (!Number.isFinite(n)) return "—";
  return n.toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

// Anchor every date format to São Paulo so SSR (Docker UTC) and the
// browser (typically America/Sao_Paulo) agree exactly — a timestamp
// near midnight UTC was rendering "30 abr" on the server and "29 abr"
// on the client, tripping React #418.
const BR_TZ = "America/Sao_Paulo";

export function formatDate(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    timeZone: BR_TZ,
  });
}

export function formatDateTime(value?: string | null): string {
  if (!value) return "—";
  const iso = /^\d{4}-/.test(value)
    ? value.replace(" ", "T")
    : value.replace(/^(\d{2})\/(\d{2})\/(\d{4})\s/, "$3-$2-$1T");
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString("pt-BR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    timeZone: BR_TZ,
  });
}

/**
 * Relative time in Portuguese ("agora", "há 3 minutos", "há 2 dias").
 * Falls back to formatDate for anything older than 7 days.
 */
export function formatRelative(value?: string | null, now: Date = new Date()): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "—";
  const diff = now.getTime() - d.getTime();
  const sec = Math.round(diff / 1000);

  if (sec < 0) return formatDate(value);
  if (sec < 30) return "agora";
  if (sec < 60) return `há ${sec}s`;
  const min = Math.round(sec / 60);
  if (min < 60) return `há ${min} min`;
  const hr = Math.round(min / 60);
  if (hr < 24) return `há ${hr}h`;
  const day = Math.round(hr / 24);
  if (day === 1) return "ontem";
  if (day < 7) return `há ${day} dias`;
  return formatDate(value);
}

/**
 * Removes the "João Silva João Silva" pattern WC themes sometimes emit when
 * they store the full name in both first and last name fields.
 */
export function dedupeName(name: string): string {
  const trimmed = name.trim().replace(/\s+/g, " ");
  if (!trimmed) return trimmed;
  // If the second half equals the first half, drop it.
  const half = Math.floor(trimmed.length / 2);
  if (trimmed.length % 2 === 1 && trimmed[half] === " ") {
    const left = trimmed.slice(0, half);
    const right = trimmed.slice(half + 1);
    if (left.toLowerCase() === right.toLowerCase()) return left;
  }
  // Generic: last token block equals the prefix? "Maria Ivani Maria Ivani" or
  // "Aaaa Bbbb Cccc Aaaa Bbbb Cccc"
  const tokens = trimmed.split(" ");
  for (let len = Math.floor(tokens.length / 2); len >= 1; len--) {
    const first = tokens.slice(0, len).join(" ").toLowerCase();
    const last = tokens.slice(tokens.length - len).join(" ").toLowerCase();
    if (first === last) {
      return tokens.slice(0, tokens.length - len).join(" ");
    }
  }
  return trimmed;
}
