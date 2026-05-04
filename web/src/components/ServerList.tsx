"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { Server } from "@/hooks/useServers";
import { authFetch } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Toast } from "@/components/Toast";
import { daysUntilExpire } from "@/lib/format";

interface Props {
  servers: Server[];
  loading: boolean;
  error: string | null;
  onRefresh: () => void;
  onDelete: (id: number) => Promise<void>;
}

function StatusBadge({ status }: { status: string }) {
  if (status === "online") return <Badge className="bg-green-600 text-white text-xs">在线</Badge>;
  if (status === "offline") return <Badge className="bg-red-600 text-white text-xs">离线</Badge>;
  return <Badge className="bg-slate-600 text-white text-xs">未知</Badge>;
}

function expireDateClass(expireAt: string | null): string {
  if (!expireAt) return "text-slate-500";
  const days = daysUntilExpire(expireAt);
  if (days < 0) return "text-red-400 font-semibold";
  if (days <= 7) return "text-orange-400 font-semibold";
  if (days <= 30) return "text-yellow-400";
  return "text-slate-300";
}

export function ServerList({ servers, loading, error, onRefresh, onDelete }: Props) {
  const router = useRouter();
  const [confirmServer, setConfirmServer] = useState<Server | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [batchRebooting, setBatchRebooting] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);
  const [search, setSearch] = useState("");

  const filtered = servers.filter(
    (s) => !search || s.name.toLowerCase().includes(search.toLowerCase()) || s.host.includes(search)
  );
  const allSelected = filtered.length > 0 && filtered.every((s) => selected.has(s.id));

  function toggleAll() {
    const ids = filtered.map((s) => s.id);
    setSelected((prev) => {
      const next = new Set(prev);
      allSelected ? ids.forEach((id) => next.delete(id)) : ids.forEach((id) => next.add(id));
      return next;
    });
  }

  function toggleOne(id: number) {
    setSelected((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }

  async function handleDelete() {
    if (!confirmServer) return;
    setDeleting(true);
    try {
      await onDelete(confirmServer.id);
      setToast({ msg: `已删除：${confirmServer.name}`, type: "ok" });
      setConfirmServer(null);
      onRefresh();
    } catch (e) {
      setToast({ msg: `删除失败：${(e as Error).message}`, type: "err" });
    } finally { setDeleting(false); }
  }

  async function handleBatchReboot() {
    setBatchRebooting(true);
    try {
      const results = await authFetch<Array<{ server_id: number; result: string }>>(
        "/api/servers/batch-reboot",
        { method: "POST", body: JSON.stringify({ server_ids: Array.from(selected) }) }
      );
      const ok = results.filter((r) => r.result === "success").length;
      const fail = results.length - ok;
      setToast({ msg: `${ok} 台成功${fail > 0 ? `，${fail} 台失败` : ""}`, type: fail > 0 ? "err" : "ok" });
      setSelected(new Set());
    } catch (e) {
      setToast({ msg: `批量重启失败：${(e as Error).message}`, type: "err" });
    } finally { setBatchRebooting(false); }
  }

  if (loading) return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center text-slate-400">加载中…</div>
  );

  return (
    <div className="min-h-screen bg-slate-900 text-white p-6 space-y-4">
      {toast && <Toast message={toast.msg} type={toast.type} onClose={() => setToast(null)} />}
      {error && <Toast message="无法连接到服务器，请稍后重试" type="err" onClose={() => {}} />}

      {/* Delete confirm modal */}
      {confirmServer && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
          <div className="bg-slate-800 border border-slate-600 rounded-xl p-6 w-96 space-y-4">
            <p className="text-white font-medium">确认删除 {confirmServer.name}？</p>
            <p className="text-slate-400 text-sm">此操作不可恢复。</p>
            <div className="flex gap-3 justify-end">
              <Button variant="ghost" onClick={() => setConfirmServer(null)}>取消</Button>
              <Button variant="destructive" disabled={deleting} onClick={handleDelete}>
                {deleting ? "删除中…" : "确认删除"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div className="flex items-center gap-3">
          <button onClick={() => router.push("/")} className="text-slate-400 hover:text-white text-sm">← 首页</button>
          <h1 className="text-xl font-bold">服务器管理</h1>
          <span className="text-slate-400 text-sm">共 {servers.length} 台</span>
        </div>
        <div className="flex gap-2 flex-wrap items-center">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索名称/IP…"
            className="bg-slate-800 border border-slate-600 rounded-lg px-3 py-1.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 w-44"
          />
          {selected.size > 0 && (
            <>
              <span className="text-slate-400 text-sm">已选 {selected.size} 台</span>
              <Button size="sm" variant="outline" className="border-slate-600 text-slate-300 hover:text-white"
                disabled={batchRebooting} onClick={handleBatchReboot}>
                {batchRebooting ? "重启中…" : "批量重启"}
              </Button>
              <Button size="sm" variant="ghost" className="text-slate-400"
                onClick={() => setSelected(new Set())}>取消选择</Button>
            </>
          )}
          <Button size="sm" onClick={() => router.push("/servers/new")}>+ 新增服务器</Button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto rounded-xl border border-slate-700">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-700 text-slate-400">
              <th className="px-3 py-2 w-8">
                <input type="checkbox" checked={allSelected} onChange={toggleAll} className="accent-blue-500" />
              </th>
              <th className="px-3 py-2 text-left">名称</th>
              <th className="px-3 py-2 text-left">IP</th>
              <th className="px-3 py-2 text-left">线路</th>
              <th className="px-3 py-2 text-left">状态</th>
              <th className="px-3 py-2 text-left">月租</th>
              <th className="px-3 py-2 text-left">到期日</th>
              <th className="px-3 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 && (
              <tr><td colSpan={8} className="text-center text-slate-500 py-10">
                {search ? "没有匹配的服务器" : "暂无服务器"}
              </td></tr>
            )}
            {filtered.map((s) => (
              <tr key={s.id} className={`border-b border-slate-800 hover:bg-slate-800/50 ${selected.has(s.id) ? "bg-blue-950/20" : ""}`}>
                <td className="px-3 py-2">
                  <input type="checkbox" checked={selected.has(s.id)} onChange={() => toggleOne(s.id)} className="accent-blue-500" />
                </td>
                <td className="px-3 py-2">
                  <button className="text-blue-400 hover:text-blue-300 font-medium" onClick={() => router.push(`/servers/${s.id}`)}>
                    {s.name}
                  </button>
                </td>
                <td className="px-3 py-2 font-mono text-slate-300 text-xs">{s.host}</td>
                <td className="px-3 py-2 text-slate-400">{s.carrier_type ?? "—"}</td>
                <td className="px-3 py-2"><StatusBadge status={s.status} /></td>
                <td className="px-3 py-2 text-slate-300">{s.monthly_rent > 0 ? `¥${s.monthly_rent.toFixed(0)}` : "—"}</td>
                <td className={`px-3 py-2 text-xs ${expireDateClass(s.expire_at)}`}>
                  {s.expire_at ? s.expire_at.slice(0, 10) : "—"}
                </td>
                <td className="px-3 py-2 text-right">
                  <div className="flex gap-1 justify-end">
                    <Button size="sm" variant="ghost" className="text-xs text-slate-400 hover:text-white"
                      onClick={() => router.push(`/servers/${s.id}/edit`)}>编辑</Button>
                    <Button size="sm" variant="ghost" className="text-xs text-red-400 hover:text-red-300"
                      onClick={() => setConfirmServer(s)}>删除</Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
