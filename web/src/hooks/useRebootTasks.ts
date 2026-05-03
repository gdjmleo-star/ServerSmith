"use client";
import { useCallback, useEffect, useState } from "react";
import { authFetch } from "./useAuth";

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
      const data = await authFetch<RebootTask[]>(`/api/servers/${serverId}/reboot-tasks?limit=${limit}`);
      setTasks(data);
    } catch {
      // silent degraded
    } finally {
      setLoading(false);
    }
  }, [serverId, limit]);

  useEffect(() => { fetchTasks(); }, [fetchTasks]);

  return { tasks, loading, refresh: fetchTasks };
}
