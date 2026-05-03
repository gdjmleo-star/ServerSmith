"use client";
import { ServerList } from "@/components/ServerList";
import { useServers } from "@/hooks/useServers";

export default function ServersPage() {
  const { servers, loading, error, refresh, deleteServer } = useServers();
  return (
    <ServerList
      servers={servers}
      loading={loading}
      error={error}
      onRefresh={refresh}
      onDelete={deleteServer}
    />
  );
}
