"use client";
import { useCallback, useEffect, useState } from "react";
import { API_BASE, DEFAULT_HISTORY_DAYS } from "@/lib/config";

export interface HistoryPoint {
  date: string;
  bytes_in: number;
  bytes_out: number;
  avg_cpu: number;
  avg_mem: number;
}

export function useHistory(serverId: number, days: number = DEFAULT_HISTORY_DAYS) {
  const [data, setData] = useState<HistoryPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch_ = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(
        `${API_BASE}/api/servers/${serverId}/history?days=${days}`
      );
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      setData(json ?? []);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [serverId, days]);

  useEffect(() => {
    if (serverId > 0) fetch_();
  }, [fetch_, serverId]);

  return { data, loading, error };
}
