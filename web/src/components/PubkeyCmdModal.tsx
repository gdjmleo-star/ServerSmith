"use client";
import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { authFetch } from "@/hooks/useAuth";

interface Props {
  serverId: number;
  onClose: () => void;
}

export function PubkeyCmdModal({ serverId, onClose }: Props) {
  const [cmd, setCmd] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    authFetch<{ install_cmd: string; public_key: string }>(
      `/api/servers/${serverId}/pubkey-cmd`
    )
      .then((d) => { setCmd(d.install_cmd ?? ""); setLoading(false); })
      .catch(() => setLoading(false));
  }, [serverId]);

  const copy = () => {
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(cmd).then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }).catch(fallbackCopy);
    } else {
      fallbackCopy();
    }
  };

  const fallbackCopy = () => {
    const el = document.createElement("textarea");
    el.value = cmd;
    el.style.position = "fixed";
    el.style.opacity = "0";
    document.body.appendChild(el);
    el.focus();
    el.select();
    try {
      document.execCommand("copy");
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {/* ignore */} finally {
      document.body.removeChild(el);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      onClick={onClose}
    >
      <div
        className="bg-slate-800 border border-slate-600 rounded-xl p-6 w-[640px] max-w-[95vw] space-y-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-white font-semibold text-base">
          第一步：在被管理服务器上粘贴此命令
        </h3>
        <p className="text-slate-400 text-sm">
          SSH 登录到目标服务器，粘贴以下命令，将管理端公钥写入 authorized_keys（只需执行一次）：
        </p>
        {loading ? (
          <div className="text-slate-400 text-sm">加载中…</div>
        ) : (
          <div className="bg-slate-900 rounded-lg p-3 font-mono text-xs text-green-300 break-all select-all">
            {cmd}
          </div>
        )}
        <div className="flex gap-3 justify-end">
          <Button size="sm" variant="ghost" onClick={onClose}>
            关闭
          </Button>
          <Button size="sm" onClick={copy} disabled={!cmd}>
            {copied ? "✅ 已复制" : "📋 复制命令"}
          </Button>
        </div>
      </div>
    </div>
  );
}
