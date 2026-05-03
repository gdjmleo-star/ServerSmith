"use client";
import { useRouter } from "next/navigation";
import { Server, useServers } from "@/hooks/useServers";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatUTCToBeijing } from "@/lib/timezone";

function probeStatus(server: Server): "installed" | "offline" | "unknown" {
  if (!server.last_report_at) return "unknown";
  const ms = Date.now() - new Date(server.last_report_at).getTime();
  return ms < 5 * 60 * 1000 ? "installed" : "offline";
}

function ProbeStatusBadge({ server }: { server: Server }) {
  const st = probeStatus(server);
  if (st === "installed") return <Badge className="bg-green-600 text-white text-xs">在线</Badge>;
  if (st === "offline") return <Badge className="bg-orange-600 text-white text-xs">离线</Badge>;
  return <Badge className="bg-slate-600 text-white text-xs">未安装</Badge>;
}

export function ProbeTable() {
  const router = useRouter();
  const { servers, loading } = useServers();

  if (loading) return <div className="text-slate-400 text-sm py-8 text-center">加载中…</div>;

  const installed = servers.filter((s) => s.last_report_at).length;
  const online = servers.filter((s) => probeStatus(s) === "installed").length;

  return (
    <div className="space-y-4">
      <div className="flex gap-4 text-sm text-slate-400">
        <span>共 <strong className="text-white">{servers.length}</strong> 台</span>
        <span>已安装 <strong className="text-white">{installed}</strong> 台</span>
        <span>当前在线 <strong className="text-green-400">{online}</strong> 台</span>
      </div>

      <div className="rounded-xl border border-slate-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-700 bg-slate-800/60">
              <th className="px-3 py-2 text-left text-slate-400 font-normal">服务器</th>
              <th className="px-3 py-2 text-left text-slate-400 font-normal">IP</th>
              <th className="px-3 py-2 text-left text-slate-400 font-normal">探针状态</th>
              <th className="px-3 py-2 text-left text-slate-400 font-normal">版本</th>
              <th className="px-3 py-2 text-left text-slate-400 font-normal">最后上报</th>
              <th className="px-3 py-2 text-right text-slate-400 font-normal">操作</th>
            </tr>
          </thead>
          <tbody>
            {servers.length === 0 && (
              <tr><td colSpan={6} className="text-center text-slate-500 py-10">暂无服务器</td></tr>
            )}
            {servers.map((s) => (
              <tr key={s.id} className="border-b border-slate-800 hover:bg-slate-800/40">
                <td className="px-3 py-2">
                  <button className="text-blue-400 hover:text-blue-300 font-medium"
                    onClick={() => router.push(`/servers/${s.id}`)}>
                    {s.name}
                  </button>
                </td>
                <td className="px-3 py-2 text-slate-400 font-mono text-xs">{s.host}</td>
                <td className="px-3 py-2"><ProbeStatusBadge server={s} /></td>
                <td className="px-3 py-2 text-slate-400 text-xs">{s.agent_version ?? "未知"}</td>
                <td className="px-3 py-2 text-slate-400 text-xs">
                  {s.last_report_at ? formatUTCToBeijing(s.last_report_at) : "—"}
                </td>
                <td className="px-3 py-2 text-right">
                  <Button size="sm" variant="ghost" className="text-xs text-slate-400 hover:text-white"
                    onClick={() => router.push(`/servers/${s.id}`)}>
                    详情
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
