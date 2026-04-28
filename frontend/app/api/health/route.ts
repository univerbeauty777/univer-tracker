import { NextResponse } from "next/server";

// Trivial health endpoint for the docker healthcheck and external probes.
// Crucially: no fetch, no DB, no external dependency — if this 200s, the
// Next process is up and serving. Probing "/" instead would do an SSR
// render with three upstream fetches, which can mask transient backend
// hiccups as a frontend outage.
export const dynamic = "force-dynamic";

export function GET() {
  return NextResponse.json({ status: "ok" }, { status: 200 });
}
