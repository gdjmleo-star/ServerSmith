// Server component — required for generateStaticParams with output: 'export'
import ServerDetailClient from "./client";

export async function generateStaticParams() {
  try {
    const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:18080";
    const res = await fetch(`${apiBase}/api/servers`, {
      signal: AbortSignal.timeout(5000),
    });
    if (!res.ok) return [{ id: "1" }];
    const servers: { id: number }[] = await res.json();
    return servers.map((s) => ({ id: String(s.id) }));
  } catch {
    // API unreachable — return fallback, won't block build
    return [{ id: "1" }];
  }
}

export default function ServerDetailPage() {
  return <ServerDetailClient />;
}
