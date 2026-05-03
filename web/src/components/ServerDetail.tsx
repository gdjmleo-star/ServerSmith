"use client";
import { useRouter } from "next/navigation";
import { Server } from "@/hooks/useServers";
import { RebootTask, useRebootTasks } from "@/hooks/useRebootTasks";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Toast } from "@/components/Toast";
import { formatUTCToBeijing, timeAgo } from "@/lib/timezone";
import { useState } from "react";
import { authFetch } from "@/hooks/useAuth";

interface Props {
  server: Server;
}

function StatusBadge({ status }: { status: string }) {
  const cls = status === "online" ? "bg-green-600" : status === "offline" ? "bg-red-600" : "bg-slate-600";
  const label = status === "online" ? "在线" : status === "offline" ? "离线" : "未知";
  return <Badge className={`${cls} text-white text-xs`}>{label}</Badge>;
}

function actionLabel(action: string) {
  return action === "cycle_reboot" ? "周期复位" : action === "manual_reboot" ? "手动重启" : action;
}

function resultBadge(result: string) {
  if (result === "success") return <span className="text-green-400 text-xs">成功</span>;
  if (result === "failed") return <span className="text-red-400 text-xs">失败</span>;
  return <span className="text-slate-400 text-xs">等待中</span>;
}

export function ServerDetail({ server }: Props) {
  const router = useRouter();
  const { tasks } = useRebootTasks(server.id);
  const [rebooting, setRebooting] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);

  const usedGB = (server.used_bytes ?? 0) / 1e9;
  const totalGB = server.total_quota ?? 0;
  const pct = totalGB > 0 ? Math.min(100, (usedGB / totalGB) * 100) : 0;
  const pctColor = pct >= 90 ? "bg-red-500" : pct >= 75 ? "bg-orange-500" : "bg-cyan-500";

  async function handleReboot() {
    setRebooting(true);
    try {
      await authFetch<unknown>(`/api/servers/${server.id}/reboot`, { method: "POST" });
      setToast({ msg: "重启指令已发送", type: "ok" });
    } catch (e) {
      setToast({ msg: `重启失败：${(e as Error).message}`, type: "err" });
    } finally {
      setRebooting(false);
    }
  }

  const row = (label: string, value: React.ReactNode) => (
    <div className="flex items-start justify-between py-2 border-b border-slate-800 last:border-0">
      <span className="text-slate-400 text-sm">{label}</span>
      <span className="text-white text-sm text-right max-w-[60%]">{value}</span>
    </div>
  );

  return (
    <div className="min-h-screen bg-slate-900 text-white p-6 space-y-5 max-w-2xl mx-auto">
      {toast && <Toast message={toast.msg} type={toast.type} onClose={() => setToast(null)} />}

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <button onClick={() => router.push("/servers")} className="text-slate-400 hover:text-white text-sm">← 服务器列表</button>
          </div>
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold">{server.name}</h1>
            <StatusBadge status={server.status} />
          </div>
          <p className="text-slate-400 text-sm mt-1">{server.host}</p>
        </div>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" className="border-slate-600 text-slate-300 hover:text-white"
            onClick={() => router.push(`/servers/${server.id}/edit`)}>编辑</Button>
          <Button size="sm" disabled={rebooting} onClick={handleReboot}
            className="bg-blue-600 hover:bg-blue-500">
            {rebooting ? "重启中…" : "手动重启"}
          </Button>
        </div>
      </div>

      {/* Basic Info */}
      <div className="bg-slate-800 border border-slate-700 rounded-xl px-4 py-1">
        {row("类型", server.server_type === "node" ? "流量节点" : "项目服务器")}
        {server.carrier_type && row("线路", server.carrier_type)}
        {row("月租", `¥${(server.monthly_rent ?? 0).toFixed(2)}/月`)}
        {row("SSH", `${server.host}:${server.ssh_port} (${server.ssh_user})`)}
        {row("探针版本", server.agent_version ?? "未知")}
        {row("最后上报", server.last_report_at ? timeAgo(server.last_report_at) : "从未")}
      </div>

      {/* Traffic (node) */}
      {server.plan_type === "traffic" && (
        <div className="bg-slate-800 border border-slate-700 rounded-xl px-4 py-3 space-y-3">
          <h2 className="text-sm font-medium text-slate-400">流量使用</h2>
          <div className="w-full bg-slate-700 rounded-full h-2">
            <div className={`h-2 rounded-full transition-all ${pctColor}`} style={{ width: `${pct}%` }} />
          </div>
          <div className="flex justify-between text-xs text-slate-400">
            <span>已用 {usedGB.toFixed(1)} GB</span>
            <span>{pct.toFixed(1)}%</span>
            <span>总额 {totalGB} GB</span>
          </div>
          {server.next_cycle_at && (
            <p className="text-xs text-slate-500">下次复位：{formatUTCToBeijing(server.next_cycle_at)}</p>
          )}
        </div>
      )}

      {/* Expire (no_limit) */}
      {server.plan_type === "no_limit" && server.expire_at && (
        <div className="bg-slate-800 border border-slate-700 rounded-xl px-4 py-3">
          {row("到期时间", formatUTCToBeijing(server.expire_at))}
        </div>
      )}

      {/* Reboot Tasks */}
      <div className="bg-slate-800 border border-slate-700 rounded-xl overflow-hidden">
        <div className="px-4 py-2 border-b border-slate-700">
          <span className="text-sm font-medium text-slate-400">最近重启记录</span>
        </div>
        {tasks.length === 0 ? (
          <p className="text-slate-500 text-sm px-4 py-3">暂无记录</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-slate-400">
                <th className="px-3 py-2 text-left font-normal">操作</th>
                <th className="px-3 py-2 text-left font-normal">触发时间</th>
                <th className="px-3 py-2 text-left font-normal">结果</th>
              </tr>
            </thead>
            <tbody>
              {tasks.map((t: RebootTask) => (
                <tr key={t.id} className="border-b border-slate-800">
                  <td className="px-3 py-2 text-slate-300">{actionLabel(t.action)}</td>
                  <td className="px-3 py-2 text-slate-400 text-xs">{formatUTCToBeijing(t.triggered_at)}</td>
                  <td className="px-3 py-2">{resultBadge(t.result)}{t.error_msg && <span className="ml-1 text-slate-500 text-xs truncate max-w-[100px] inline-block">{t.error_msg}</span>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
