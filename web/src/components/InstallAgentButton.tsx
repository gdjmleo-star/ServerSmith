"use client";
import { useState } from "react";
import { Server } from "@/hooks/useServers";
import { Button } from "@/components/ui/button";
import { authFetch } from "@/hooks/useAuth";

interface Props {
  server: Server;
}

export function InstallAgentButton({ server }: Props) {
  const [status, setStatus] = useState<"idle" | "loading" | "ok" | "err">("idle");
  const [msg, setMsg] = useState("");

  const install = async () => {
    setStatus("loading");
    setMsg("");
    try {
      await authFetch<unknown>(`/api/servers/${server.id}/install-agent`, {
        method: "POST",
      });
      setStatus("ok");
      setMsg("安装成功！探针将在30秒内开始上报。");
    } catch (e: unknown) {
      setStatus("err");
      setMsg(e instanceof Error ? e.message : "安装失败");
    }
  };

  if (status === "loading")
    return <span className="text-xs text-slate-400 animate-pulse">安装中…</span>;
  if (status === "ok")
    return <span className="text-xs text-green-400">✅ {msg}</span>;
  if (status === "err")
    return (
      <span
        className="text-xs text-red-400 cursor-pointer"
        title={msg}
        onClick={() => setStatus("idle")}
      >
        ❌ 失败(点重试)
      </span>
    );

  return (
    <Button
      size="sm"
      variant="outline"
      className="text-xs border-slate-600 text-slate-300 hover:text-white hover:border-blue-500"
      onClick={install}
    >
      🚀 安装探针
    </Button>
  );
}
