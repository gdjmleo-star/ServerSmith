"use client";
import { useCallback, useEffect, useState } from "react";
import { API_BASE } from "@/lib/config";

export interface Server {
  id: number;
  name: string;
  host: string;
  ssh_port: number;
  ssh_user: string;
  server_type: string;
  carrier_type: string | null;
  monthly_rent: number;
  status: string;
  last_report_at: string | null;
  agent_version: string | null;
  created_at: string;
  plan_type: string;
  total_quota: number | null;
  used_bytes: number;
  used_gb: number;
  remaining_gb: number;
  usage_percent: number;
  next_cycle_at: string | null;
  expire_at: string | null;
  cycle_day: number | null;
  cycle_time: string | null;
  expire_notify_days: number | null;
  avg_cpu: number | null;
  avg_mem: number | null;
}

export interface ServerFormData {
  name: string;
  host: string;
  ssh_port: number;
  ssh_user: string;
  server_type: string;
  carrier_type: string;
  monthly_rent: number;
  plan_type: string;
  total_quota: number | null;
  initial_used_gb: number | null;
  cycle_day: number | null;
  cycle_time: string;
  expire_at: string;
  expire_notify_days: number | null;
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "Unknown error");
    throw new Error(text || `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export function useServers() {
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchServers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiFetch<Server[]>("/api/servers");
      setServers(data ?? []);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchServers();
  }, [fetchServers]);

  const createServer = useCallback(async (body: ServerFormData) => {
    await apiFetch("/api/servers", {
      method: "POST",
      body: JSON.stringify(body),
    });
  }, []);

  const updateServer = useCallback(
    async (id: number, body: ServerFormData) => {
      await apiFetch(`/api/servers/${id}`, {
        method: "PUT",
        body: JSON.stringify(body),
      });
    },
    []
  );

  const deleteServer = useCallback(async (id: number) => {
    await apiFetch(`/api/servers/${id}`, { method: "DELETE" });
  }, []);

  const getServer = useCallback(async (id: number): Promise<Server> => {
    return apiFetch<Server>(`/api/servers/${id}`);
  }, []);

  return {
    servers,
    loading,
    error,
    refresh: fetchServers,
    createServer,
    updateServer,
    deleteServer,
    getServer,
  };
}
