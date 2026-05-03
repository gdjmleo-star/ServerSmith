"use client";
import { useCallback, useEffect, useState } from "react";
import { API_BASE } from "@/lib/config";

export interface RebootTask {
  id: number;
  action: string;
  triggered_at: string;
  executed_at: string | null;
  result: string;
  error_msg: string | null;
}

export function useRebootTasks(serverId: number | null, limit = 5) {
  const [tasks, setTasks] = useState<RebootTask[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchTasks = useCallback(async () => {
    if (!serverId) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/servers/${serverId}/reboot-tasks?limit=${limit}`);
      if (res.ok) setTasks(await res.json());
    } catch {
      // silent degraded
    } finally {
      setLoading(false);
    }
  }, [serverId, limit]);

  useEffect(() => { fetchTasks(); }, [fetchTasks]);

  return { tasks, loading, refresh: fetchTasks };
}
