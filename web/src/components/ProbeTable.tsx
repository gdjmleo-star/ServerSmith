"use client";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Server, useServers } from "@/hooks/useServers";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatUTCToBeijing } from "@/lib/timezone";
import { authFetch } from "@/hooks/useAuth";

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

// 公钥命令弹窗
function PubkeyCmdModal({ serverId, onClose }: { serverId: number; onClose: () => void }) {
  const [cmd, setCmd] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    authFetch<{install_cmd: string; public_key: string}>(`/api/servers/${serverId}/pubkey-cmd`)
      .then((d) => { setCmd(d.install_cmd ?? ""); setLoading(false); })
      .catch(() => setLoading(false));
  }, [serverId]);

  const copy = () => {
    navigator.clipboard.writeText(cmd).then(() => { setCopied(true); setTimeout(() => setCopied(false), 2000); });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div className="bg-slate-800 border border-slate-600 rounded-xl p-6 w-[640px] max-w-[95vw] space-y-4"
        onClick={(e) => e.stopPropagation()}>
        <h3 className="text-white font-semibold text-base">第一步：在被管理服务器上粘贴此命令</h3>
        <p className="text-slate-400 text-sm">SSH 登录到目标服务器，粘贴以下命令，将管理端公钥写入 authorized_keys（只需执行一次）：</p>
        {loading ? (
          <div className="text-slate-400 text-sm">加载中…</div>
        ) : (
          <div className="bg-slate-900 rounded-lg p-3 font-mono text-xs text-green-300 break-all select-all">
            {cmd}
          </div>
        )}
        <div className="flex gap-3 justify-end">
          <Button size="sm" variant="ghost" onClick={onClose}>关闭</Button>
          <Button size="sm" onClick={copy} disabled={!cmd}>
            {copied ? "✅ 已复制" : "📋 复制命令"}
          </Button>
        </div>
      </div>
    </div>
  );
}

// 一键安装探针
function InstallAgentButton({ server }: { server: Server }) {
  const [status, setStatus] = useState<"idle" | "loading" | "ok" | "err">("idle");
  const [msg, setMsg] = useState("");

  const install = async () => {
    setStatus("loading");
    setMsg("");
    try {
      await authFetch<unknown>(
        `/api/servers/${server.id}/install-agent`, { method: "POST" }
      );
      setStatus("ok");
      setMsg("安装成功！探针将在30秒内开始上报。");
    } catch (e: unknown) {
      setStatus("err");
      setMsg(e instanceof Error ? e.message : "安装失败");
    }
  };

  if (status === "loading") return <span className="text-xs text-slate-400 animate-pulse">安装中…</span>;
  if (status === "ok") return <span className="text-xs text-green-400">✅ {msg}</span>;
  if (status === "err") return (
    <span className="text-xs text-red-400 cursor-pointer" title={msg} onClick={() => setStatus("idle")}>
      ❌ 失败(点重试)
    </span>
  );

  return (
    <Button size="sm" variant="outline"
      className="text-xs border-slate-600 text-slate-300 hover:text-white hover:border-blue-500"
      onClick={install}>
      🚀 安装探针
    </Button>
  );
}

export function ProbeTable() {
  const router = useRouter();
  const { servers, loading } = useServers();
  const [pubkeyTarget, setPubkeyTarget] = useState<number | null>(null);

  if (loading) return <div className="text-slate-400 text-sm py-8 text-center">加载中…</div>;

  const installed = servers.filter((s) => s.last_report_at).length;
  const online = servers.filter((s) => probeStatus(s) === "installed").length;

  return (
    <div className="space-y-4">
      {pubkeyTarget !== null && (
        <PubkeyCmdModal serverId={pubkeyTarget} onClose={() => setPubkeyTarget(null)} />
      )}

      <div className="flex gap-4 text-sm text-slate-400">
        <span>共 <strong className="text-white">{servers.length}</strong> 台</span>
        <span>已安装 <strong className="text-white">{installed}</strong> 台</span>
        <span>当前在线 <strong className="text-green-400">{online}</strong> 台</span>
      </div>

      <div className="bg-blue-950/40 border border-blue-800/50 rounded-lg p-3 text-xs text-blue-300 space-y-1">
        <p className="font-semibold">📋 安装探针两步走：</p>
        <p>① 点「获取公钥命令」→ 复制命令 → 粘贴到目标服务器执行（写入 SSH 公钥，只做一次）</p>
        <p>② 点「🚀 安装探针」→ 管理端自动 SSH 推送探针并启动服务</p>
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
                  <div className="flex items-center justify-end gap-2">
                    <Button size="sm" variant="ghost"
                      className="text-xs text-slate-400 hover:text-yellow-300"
                      onClick={() => setPubkeyTarget(s.id)}>
                      🔑 获取公钥命令
                    </Button>
                    <InstallAgentButton server={s} />
                    <Button size="sm" variant="ghost" className="text-xs text-slate-400 hover:text-white"
                      onClick={() => router.push(`/servers/${s.id}`)}>
                      详情
                    </Button>
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
