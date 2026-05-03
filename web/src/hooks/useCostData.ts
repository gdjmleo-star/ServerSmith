"use client";
import { useDashboard, CostPerGBItem } from "@/hooks/useDashboard";

export type { CostPerGBItem };

export function useCostData() {
  const { dashboard, loading } = useDashboard();
  return {
    totalRent: dashboard?.total_monthly_rent ?? 0,
    costPerGB: dashboard?.cost_per_gb ?? [],
    loading,
  };
}
