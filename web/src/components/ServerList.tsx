"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { Server } from "@/hooks/useServers";
import { useLiveStats } from "@/hooks/useLiveStats";
import { authFetch } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Toast } from "@/components/Toast";
import { ServerCard } from "@/components/ServerCard";

interface Props {
  servers: Server[];
  loading: boolean;
  error: string | null;
  onRefresh: () => void;
  onDelete: (id: number) => Promise<void>;
}

export function ServerList({ servers, loading, error, onRefresh, onDelete }: Props) {
  const router = useRouter();
  const { statsMap } = useLiveStats();
  const [confirmServer, setConfirmServer] = useState<Server | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [batchRebooting, setBatchRebooting] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);
  const [search, setSearch] = useState("");

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

  const filtered = servers.filter((s) =>
    !search || s.name.toLowerCase().includes(search.toLowerCase()) || s.host.includes(search)
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

      {/* Card grid */}
      {filtered.length === 0 ? (
        <div className="text-center text-slate-500 py-16">
          {search ? "没有匹配的服务器" : "暂无服务器，点击「新增服务器」开始"}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filtered.map((s) => (
            <div key={s.id} className="relative group">
              <ServerCard
                server={s}
                live={statsMap.get(s.id)}
                selected={selected.has(s.id)}
                onSelect={toggleOne}
                onClick={(id) => router.push(`/servers/${id}`)}
              />
              {/* Edit / Delete overlay */}
              <div className="absolute bottom-3 right-3 hidden group-hover:flex gap-1">
                <Button size="sm" variant="ghost"
                  className="text-xs text-slate-400 hover:text-white bg-slate-900/80 backdrop-blur-sm"
                  onClick={(e) => { e.stopPropagation(); router.push(`/servers/${s.id}/edit`); }}>
                  编辑
                </Button>
                <Button size="sm" variant="ghost"
                  className="text-xs text-red-400 hover:text-red-300 bg-slate-900/80 backdrop-blur-sm"
                  onClick={(e) => { e.stopPropagation(); setConfirmServer(s); }}>
                  删除
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
