"use client";

import { useSearchParams } from "next/navigation";
import { Download } from "lucide-react";
import { ordersExportURL } from "@/lib/api";

export function ExportButton() {
  const params = useSearchParams();

  function download() {
    const obj: Record<string, string> = {};
    params.forEach((v, k) => {
      obj[k] = v;
    });
    const url = ordersExportURL(obj);
    window.open(url, "_blank");
  }

  return (
    <button
      onClick={download}
      className="inline-flex items-center gap-1.5 rounded-md border border-border/60 px-2.5 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
    >
      <Download className="size-3.5" />
      Exportar CSV
    </button>
  );
}
