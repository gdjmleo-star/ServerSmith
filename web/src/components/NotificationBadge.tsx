"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ALERT_POLL_MS } from "@/lib/config";
import { authFetch } from "@/hooks/useAuth";

export function NotificationBadge() {
  const router = useRouter();
  const [count, setCount] = useState(0);

  useEffect(() => {
    async function poll() {
      try {
        const data = await authFetch<unknown[]>('/api/alerts?acknowledged=false&limit=100');
        setCount(Array.isArray(data) ? data.length : 0);
      } catch {
        // silently ignore — badge degraded gracefully
      }
    }
    poll();
    const t = setInterval(poll, ALERT_POLL_MS);
    return () => clearInterval(t);
  }, []);

  return (
    <button
      onClick={() => router.push("/alerts")}
      className="relative flex items-center gap-1.5 text-slate-400 hover:text-white text-sm transition"
    >
      <span>🔔</span>
      <span>告警</span>
      {count > 0 && (
        <span className="absolute -top-1.5 -right-2 bg-red-500 text-white text-[10px] font-bold rounded-full min-w-[16px] h-[16px] flex items-center justify-center px-0.5">
          {count > 99 ? "99+" : count}
        </span>
      )}
    </button>
  );
}
