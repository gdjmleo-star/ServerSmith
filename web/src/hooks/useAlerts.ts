"use client";
import { useCallback, useEffect, useState } from "react";
import { API_BASE } from "@/lib/config";

export interface Alert {
  id: number;
  server_id: number | null;
  server_name: string;
  alert_type: string;
  message: string;
  level: string;
  created_at: string;
  acknowledged: boolean;
  acknowledged_at: string | null;
}

export interface AlertConfig {
  id: number;
  server_id: number;
  alert_type: string;
  enabled: boolean;
  threshold: string;
  updated_at: string;
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

export function useAlerts(acknowledged?: boolean) {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAlerts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const qs = acknowledged !== undefined ? `?acknowledged=${acknowledged}` : "";
      const data = await apiFetch<Alert[]>(`/api/alerts${qs}`);
      setAlerts(data ?? []);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [acknowledged]);

  useEffect(() => { fetchAlerts(); }, [fetchAlerts]);

  const acknowledge = useCallback(async (id: number) => {
    await apiFetch(`/api/alerts/${id}/acknowledge`, { method: "PUT" });
    await fetchAlerts();
  }, [fetchAlerts]);

  return { alerts, loading, error, refresh: fetchAlerts, acknowledge };
}

export function useAlertConfigs(serverId: number) {
  const [configs, setConfigs] = useState<AlertConfig[]>([]);

  const fetchConfigs = useCallback(async () => {
    if (!serverId) return;
    try {
      const data = await apiFetch<AlertConfig[]>(`/api/alert-configs?server_id=${serverId}`);
      setConfigs(data ?? []);
    } catch {
      // silently ignore
    }
  }, [serverId]);

  useEffect(() => { fetchConfigs(); }, [fetchConfigs]);

  const saveConfig = useCallback(async (cfg: Omit<AlertConfig, "id" | "updated_at">) => {
    await apiFetch("/api/alert-configs", { method: "PUT", body: JSON.stringify(cfg) });
    await fetchConfigs();
  }, [fetchConfigs]);

  return { configs, saveConfig, refresh: fetchConfigs };
}
