"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useServers } from "@/hooks/useServers";
import { useLiveStats } from "@/hooks/useLiveStats";
import { ServerCard } from "@/components/ServerCard";

export function ServerOverview() {
  const router = useRouter();
  const { servers, loading } = useServers();
  const { statsMap } = useLiveStats();
  const [search, setSearch] = useState("");

  if (loading) return <p className="text-slate-500 text-sm">加载中…</p>;

  const filtered = servers.filter(
    (s) =>
      !search ||
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.host.includes(search)
  );

  return (
    <div className="space-y-3">
      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="搜索名称 / IP…"
        className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 w-64"
      />
      {filtered.length === 0 ? (
        <p className="text-slate-600 text-sm">
          {search ? "没有匹配的服务器" : "暂无服务器，请先在「服务器管理」中添加"}
        </p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filtered.map((s) => (
            <ServerCard
              key={s.id}
              server={s}
              live={statsMap.get(s.id)}
              onClick={(id) => router.push(`/servers/${id}`)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
