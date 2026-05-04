import { useState, useEffect, useCallback } from "react";
import { authFetch } from "@/hooks/useAuth";

export interface LiveStats {
  server_id: number;
  server_name: string;
  status: string;
  cpu_percent: number;
  mem_percent: number;
  disk_percent: number;
  net_in_rate: number;    // bytes/s
  net_out_rate: number;   // bytes/s
  net_in_total: number;   // bytes
  net_out_total: number;  // bytes
  used_bytes: number;
  used_bytes_in: number;
  used_bytes_out: number;
  total_quota_gb: number;
  uptime_sec: number;
  expire_at: string | null;
  last_report_at: string | null;
}

export function useLiveStats() {
  const [stats, setStats] = useState<LiveStats[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const data = await authFetch<LiveStats[]>("/api/servers/live-stats");
      setStats(Array.isArray(data) ? data : []);
    } catch {
      // silently keep stale data on error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetch();
    const timer = setInterval(fetch, 30_000);
    return () => clearInterval(timer);
  }, [fetch]);

  // Map by server_id for O(1) lookup
  const statsMap = new Map<number, LiveStats>(stats.map((s) => [s.server_id, s]));

  return { stats, statsMap, loading };
}
