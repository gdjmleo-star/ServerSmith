import { useEffect, useState } from "react";
import { API_BASE } from "@/lib/config";

export interface CostPerGBItem {
  carrier: string;
  total_quota_gb: number;
  monthly_rent: number;
  cost_per_gb: number;
}

export interface DashboardData {
  total_servers: number;
  online_count: number;
  offline_count: number;
  total_monthly_rent: number;
  total_quota_gb: number;
  total_used_gb: number;
  alerts_unread: number;
  cost_per_gb: CostPerGBItem[];
}

export interface CarrierSummary {
  carrier_type: string;
  server_count: number;
  total_quota_gb: number;
  used_gb: number;
  remaining_gb: number;
  usage_percent: number;
  total_rent: number;
  cost_per_gb: number;
  abnormal_count: number;
}

export interface AlertItem {
  id: number;
  server_name: string;
  alert_type: string;
  message: string;
  level: string;
  created_at: string;
  acknowledged: boolean;
}

export interface DashboardState {
  dashboard: DashboardData | null;
  carriers: CarrierSummary[];
  alerts: AlertItem[];
  loading: boolean;
}

export function useDashboard(): DashboardState {
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [carriers, setCarriers] = useState<CarrierSummary[]>([]);
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      fetch(`${API_BASE}/api/dashboard`).then((r) => r.json()),
      fetch(`${API_BASE}/api/dashboard/by-carrier`).then((r) => r.json()),
      fetch(`${API_BASE}/api/alerts?limit=10`).then((r) => r.json()),
    ])
      .then(([dash, car, alts]) => {
        setDashboard(dash);
        setCarriers(car);
        setAlerts(alts);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  return { dashboard, carriers, alerts, loading };
}
