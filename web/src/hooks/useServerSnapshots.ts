import { useState, useEffect, useCallback } from "react";
import { authFetch } from "@/hooks/useAuth";

export interface SnapshotPoint {
  date: string;
  bytes_in: number;
  bytes_out: number;
  avg_cpu: number;
  avg_mem: number;
  avg_disk: number;
}

export function useServerSnapshots(serverId: number, hours = 24) {
  const [snapshots, setSnapshots] = useState<SnapshotPoint[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    setLoading(true);
    try {
      // history API uses ?days= param; pass fractional day for hours
      const days = Math.max(1, Math.ceil(hours / 24));
      const data = await authFetch<SnapshotPoint[]>(
        `/api/servers/${serverId}/history?days=${days}`
      );
      setSnapshots(Array.isArray(data) ? data : []);
    } catch {
      setSnapshots([]);
    } finally {
      setLoading(false);
    }
  }, [serverId, hours]);

  useEffect(() => { fetch(); }, [fetch]);

  return { snapshots, loading };
}
