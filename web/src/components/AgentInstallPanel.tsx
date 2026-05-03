"use client";
import { useState } from "react";
import { useServers } from "@/hooks/useServers";

const TABS = ["Linux", "Windows", "macOS"] as const;
type Tab = typeof TABS[number];

function getManagerUrl(): string {
  if (typeof window !== "undefined") {
    return `${window.location.protocol}//${window.location.host}`;
  }
  return "";
}

function getInstallCmd(tab: Tab, serverId: number, managerUrl: string): string {
  switch (tab) {
    case "Linux":
      return `curl -sSL ${managerUrl}/agent/install.sh | MANAGER_URL=${managerUrl} SERVER_ID=${serverId} bash`;
    case "macOS":
      return `curl -sSL ${managerUrl}/agent/install.sh | MANAGER_URL=${managerUrl} SERVER_ID=${serverId} bash`;
    case "Windows":
      return `$env:MANAGER_URL="${managerUrl}"; $env:SERVER_ID="${serverId}"; irm ${managerUrl}/agent/install.ps1 | iex`;
  }
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };
  return (
    <button
      onClick={copy}
      className="shrink-0 px-3 py-1.5 rounded-md text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors"
    >
      {copied ? "✅ 已复制" : "📋 复制"}
    </button>
  );
}

export function AgentInstallPanel() {
  const { servers } = useServers();
  const [activeTab, setActiveTab] = useState<Tab>("Linux");
  const [selectedId, setSelectedId] = useState<number>(servers[0]?.id ?? 0);
  const managerUrl = getManagerUrl();

  const currentServerId = selectedId || servers[0]?.id || 0;
  const cmd = currentServerId ? getInstallCmd(activeTab, currentServerId, managerUrl) : "";

  return (
    <div className="space-y-6 max-w-3xl">
      {/* 说明 */}
      <div className="bg-slate-800/60 border border-slate-700 rounded-xl p-5 space-y-2">
        <h2 className="text-white font-semibold text-base">如何安装探针？</h2>
        <ol className="text-slate-300 text-sm space-y-1 list-decimal list-inside">
          <li>在下方选择目标服务器</li>
          <li>选择被管理服务器的操作系统</li>
          <li>复制安装命令</li>
          <li>SSH 登录到被管理服务器，粘贴命令并执行</li>
          <li>约 30 秒后，服务器列表中该服务器状态变为「在线」</li>
        </ol>
      </div>

      {/* 选择服务器 */}
      <div className="space-y-2">
        <label className="text-slate-400 text-sm">选择服务器</label>
        <select
          value={currentServerId}
          onChange={(e) => setSelectedId(Number(e.target.value))}
          className="w-full bg-slate-800 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
        >
          {servers.length === 0 && <option value={0}>暂无服务器，请先在「服务器管理」中添加</option>}
          {servers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} ({s.host}) — ID: {s.id}
            </option>
          ))}
        </select>
      </div>

      {/* 系统 Tab */}
      <div className="space-y-3">
        <div className="flex gap-2">
          {TABS.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                activeTab === tab
                  ? "bg-blue-600 text-white"
                  : "bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700"
              }`}
            >
              {tab === "Linux" && "🐧 "}
              {tab === "Windows" && "🪟 "}
              {tab === "macOS" && "🍎 "}
              {tab}
            </button>
          ))}
        </div>

        {/* 安装命令 */}
        {currentServerId ? (
          <div className="space-y-2">
            <p className="text-slate-400 text-xs">
              {activeTab === "Windows"
                ? "以管理员身份打开 PowerShell，粘贴以下命令："
                : "在终端中粘贴以下命令（需要 root 权限）："}
            </p>
            <div className="flex items-start gap-3 bg-slate-900 border border-slate-700 rounded-xl p-4">
              <code className="flex-1 text-green-300 text-xs font-mono break-all leading-relaxed">
                {cmd}
              </code>
              <CopyButton text={cmd} />
            </div>
            {activeTab === "Windows" && (
              <p className="text-slate-500 text-xs">
                💡 提示：如遇 ExecutionPolicy 报错，先运行 <code className="text-slate-300">Set-ExecutionPolicy RemoteSigned -Scope CurrentUser</code>
              </p>
            )}
          </div>
        ) : (
          <div className="text-slate-500 text-sm py-4 text-center">请先选择服务器</div>
        )}
      </div>
    </div>
  );
}
