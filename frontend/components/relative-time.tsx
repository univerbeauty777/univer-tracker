"use client";

import { useEffect, useState } from "react";
import { formatDate, formatDateTime, formatRelative } from "@/lib/format";

/**
 * Server renders the absolute date (deterministic, no Date.now()), client
 * upgrades to a relative timestamp on hydration. This avoids the React
 * #418 hydration mismatch caused by clock drift between SSR and the
 * browser tick.
 */
export function RelativeTime({
  value,
  fallback,
}: {
  value?: string | null;
  fallback?: string;
}) {
  const seed = fallback ?? formatDate(value);
  const [text, setText] = useState(seed);

  useEffect(() => {
    setText(formatRelative(value));
    const id = setInterval(() => setText(formatRelative(value)), 60_000);
    return () => clearInterval(id);
  }, [value]);

  return (
    <time
      dateTime={value ?? undefined}
      title={value ? formatDateTime(value) : undefined}
      suppressHydrationWarning
    >
      {text}
    </time>
  );
}
