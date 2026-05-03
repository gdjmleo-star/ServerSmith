"use client";
import { useParams } from "next/navigation";
import { useServers } from "@/hooks/useServers";
import { ServerDetail } from "@/components/ServerDetail";

export default function ServerDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { servers, loading } = useServers();
  const server = servers.find((s) => s.id === Number(id));

  if (loading) return <div className="min-h-screen bg-slate-900 flex items-center justify-center text-slate-400">加载中…</div>;
  if (!server) return <div className="min-h-screen bg-slate-900 flex items-center justify-center text-slate-400">服务器不存在</div>;

  return <ServerDetail server={server} />;
}
