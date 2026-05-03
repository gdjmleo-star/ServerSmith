"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { Alert } from "@/hooks/useAlerts";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Toast } from "@/components/Toast";
import { formatUTCToBeijing } from "@/lib/timezone";

interface Props {
  alerts: Alert[];
  loading: boolean;
  error: string | null;
  onRefresh: () => void;
  onAcknowledge: (id: number) => Promise<void>;
}

function typeBadge(t: string) {
  if (t === "offline") return <Badge className="bg-slate-600 text-white text-xs">离线</Badge>;
  if (t === "traffic_usage") return <Badge className="bg-orange-700 text-white text-xs">流量超额</Badge>;
  return <Badge className="bg-slate-700 text-white text-xs">{t}</Badge>;
}

export function AlertsList({ alerts, loading, error, onRefresh, onAcknowledge }: Props) {
  const router = useRouter();
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);
  const [acking, setAcking] = useState<number | null>(null);

  const unread = alerts.filter(a => !a.acknowledged);

  async function handleAck(id: number) {
    setAcking(id);
    try {
      await onAcknowledge(id);
      setToast({ msg: "已标记为已读", type: "ok" });
    } catch (e) {
      setToast({ msg: `操作失败：${(e as Error).message}`, type: "err" });
    } finally {
      setAcking(null);
    }
  }

  async function handleAckAll() {
    for (const a of unread) {
      await handleAck(a.id);
    }
  }

  if (loading) return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center text-slate-400">加载中…</div>
  );

  return (
    <div className="min-h-screen bg-slate-900 text-white p-6 space-y-4">
      {toast && <Toast message={toast.msg} type={toast.type} onClose={() => setToast(null)} />}
      {error && <Toast message="无法连接到服务器，请稍后重试" type="err" onClose={() => {}} />}

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button onClick={() => router.push("/")} className="text-slate-400 hover:text-white text-sm">← 首页</button>
          <h1 className="text-xl font-bold">告警列表</h1>
          {unread.length > 0 && (
            <span className="bg-red-500 text-white text-xs px-2 py-0.5 rounded-full">{unread.length} 未读</span>
          )}
        </div>
        <div className="flex gap-2">
          <Button size="sm" variant="ghost" onClick={onRefresh}>刷新</Button>
          {unread.length > 0 && (
            <Button size="sm" variant="outline" className="border-slate-600 text-slate-300 hover:text-white" onClick={handleAckAll}>
              全部标记已读
            </Button>
          )}
        </div>
      </div>

      <div className="rounded-xl border border-slate-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-700 text-slate-400">
              <th className="px-3 py-2 text-left">服务器</th>
              <th className="px-3 py-2 text-left">类型</th>
              <th className="px-3 py-2 text-left">消息</th>
              <th className="px-3 py-2 text-left">时间</th>
              <th className="px-3 py-2 text-left">状态</th>
              <th className="px-3 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            {alerts.length === 0 && (
              <tr><td colSpan={6} className="text-center text-slate-500 py-10">暂无告警</td></tr>
            )}
            {alerts.map((a) => (
              <tr key={a.id} className={`border-b border-slate-800 ${!a.acknowledged ? "bg-slate-800/40" : ""}`}>
                <td className="px-3 py-2">
                  {a.server_id ? (
                    <button className="text-blue-400 hover:text-blue-300"
                      onClick={() => router.push(`/servers/${a.server_id}/edit`)}>
                      {a.server_name || "—"}
                    </button>
                  ) : <span className="text-slate-500">—</span>}
                </td>
                <td className="px-3 py-2">{typeBadge(a.alert_type)}</td>
                <td className="px-3 py-2 text-slate-300 max-w-xs truncate">{a.message}</td>
                <td className="px-3 py-2 text-slate-400 text-xs whitespace-nowrap">{formatUTCToBeijing(a.created_at)}</td>
                <td className="px-3 py-2">
                  {a.acknowledged
                    ? <span className="text-slate-500 text-xs">已读</span>
                    : <span className="text-red-400 text-xs font-medium">未读</span>}
                </td>
                <td className="px-3 py-2 text-right">
                  {!a.acknowledged && (
                    <Button size="sm" variant="ghost" className="text-xs text-slate-400 hover:text-white"
                      disabled={acking === a.id} onClick={() => handleAck(a.id)}>
                      {acking === a.id ? "…" : "标记已读"}
                    </Button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
