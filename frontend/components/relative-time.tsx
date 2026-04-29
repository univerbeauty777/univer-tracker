"use client";

import { useEffect, useState } from "react";
import { formatDateTime, formatRelative } from "@/lib/format";

/**
 * Renders nothing on the server (just an empty <time> element) and
 * fills in the relative timestamp client-side after mount. Anything
 * computed via Intl on the server can drift from Chrome's Intl
 * (full-icu builds vs Chrome ICU snapshot, NBSP differences, etc),
 * which trips React #418. By keeping SSR deterministic-empty we
 * eliminate the mismatch by construction.
 */
export function RelativeTime({
  value,
  fallback,
}: {
  value?: string | null;
  fallback?: string;
}) {
  const [text, setText] = useState<string>(fallback ?? "");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const tick = () => setText(formatRelative(value));
    tick();
    const id = setInterval(tick, 60_000);
    return () => clearInterval(id);
  }, [value]);

  return (
    <time
      dateTime={value ?? undefined}
      title={value ? formatDateTime(value) : undefined}
      suppressHydrationWarning
    >
      {mounted ? text : (fallback ?? "\u00A0")}
    </time>
  );
}
