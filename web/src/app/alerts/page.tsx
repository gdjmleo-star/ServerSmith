"use client";
import { AlertsList } from "@/components/AlertsList";
import { useAlerts } from "@/hooks/useAlerts";

export default function AlertsPage() {
  const { alerts, loading, error, refresh, acknowledge } = useAlerts();
  return <AlertsList alerts={alerts} loading={loading} error={error} onRefresh={refresh} onAcknowledge={acknowledge} />;
}
