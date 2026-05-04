"use client";
import { Server } from "@/hooks/useServers";
import { LiveStats } from "@/hooks/useLiveStats";
import { formatRate, formatBytes, formatUptime, daysUntilExpire } from "@/lib/format";

interface Props {
  server: Server;
  live?: LiveStats;
  selected?: boolean;
  onSelect?: (id: number) => void;
  onClick?: (id: number) => void;
}

function MiniBar({ value, color }: { value: number; color: string }) {
  const pct = Math.min(100, Math.max(0, value));
  return (
    <div className="w-full bg-slate-700 rounded-full h-1.5 overflow-hidden">
      <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
    </div>
  );
}

function expireBorderClass(expireAt: string | null | undefined): string {
  if (!expireAt) return "border-slate-700";
  const days = daysUntilExpire(expireAt);
  if (days < 0) return "border-red-500";
  if (days <= 7) return "border-orange-500";
  return "border-slate-700";
}

function ExpireBadge({ expireAt }: { expireAt: string | null | undefined }) {
  if (!expireAt) return null;
  const days = daysUntilExpire(expireAt);
  if (days < 0) return <span className="text-xs bg-red-900/60 text-red-300 px-1.5 py-0.5 rounded">已过期</span>;
  if (days <= 7) return <span className="text-xs bg-orange-900/60 text-orange-300 px-1.5 py-0.5 rounded">即将到期</span>;
  return null;
}

function expireDateClass(expireAt: string | null | undefined): string {
  if (!expireAt) return "text-slate-500";
  const days = daysUntilExpire(expireAt);
  if (days < 0) return "text-red-400";
  if (days <= 7) return "text-orange-400";
  if (days <= 30) return "text-yellow-400";
  return "text-slate-400";
}

export function ServerCard({ server, live, selected, onSelect, onClick }: Props) {
  const isOnline = live?.status === "online" || server.status === "online";
  const borderCls = expireBorderClass(live?.expire_at ?? server.expire_at);

  return (
    <div
      className={`relative bg-slate-800/70 border ${borderCls} rounded-xl p-4 space-y-3 cursor-pointer hover:bg-slate-800 transition-colors`}
      onClick={() => onClick?.(server.id)}
    >
      {/* Checkbox */}
      {onSelect && (
        <div className="absolute top-3 left-3" onClick={(e) => { e.stopPropagation(); onSelect(server.id); }}>
          <input type="checkbox" checked={!!selected} onChange={() => onSelect(server.id)}
            className="accent-blue-500 w-3.5 h-3.5" />
        </div>
      )}

      {/* Header */}
      <div className="flex items-start justify-between pl-5">
        <div className="flex items-center gap-2 min-w-0">
          <span className={`w-2 h-2 rounded-full shrink-0 ${isOnline ? "bg-green-400" : "bg-slate-500"}`} />
          <span className="text-white text-sm font-medium truncate">{server.name}</span>
        </div>
        <div className="flex items-center gap-1 shrink-0 ml-2">
          {server.carrier_type && (
            <span className="text-xs bg-slate-700 text-slate-300 px-1.5 py-0.5 rounded">{server.carrier_type}</span>
          )}
          <ExpireBadge expireAt={live?.expire_at ?? server.expire_at} />
        </div>
      </div>

      {/* Uptime */}
      <div className="text-xs text-slate-500 pl-5">
        运行时间：{live ? formatUptime(live.uptime_sec) : "—"}
      </div>

      {/* CPU / Mem / Disk */}
      <div className="space-y-1.5">
        {[
          { label: "CPU", value: live?.cpu_percent ?? 0, color: "bg-blue-500" },
          { label: "内存", value: live?.mem_percent ?? 0, color: "bg-green-500" },
          { label: "硬盘", value: live?.disk_percent ?? 0, color: "bg-purple-500" },
        ].map(({ label, value, color }) => (
          <div key={label} className="flex items-center gap-2">
            <span className="text-xs text-slate-400 w-8 shrink-0">{label}</span>
            <MiniBar value={value} color={color} />
            <span className="text-xs text-slate-300 w-10 text-right shrink-0">{value.toFixed(1)}%</span>
          </div>
        ))}
      </div>

      {/* Network rates */}
      <div className="flex justify-between text-xs text-slate-400">
        <span>↑ {live ? formatRate(live.net_out_rate) : "—"}</span>
        <span>↓ {live ? formatRate(live.net_in_rate) : "—"}</span>
      </div>

      {/* Traffic totals */}
      {live && live.total_quota_gb > 0 && (
        <div className="space-y-1">
          <div className="flex justify-between text-xs text-slate-400">
            <span>已用 {formatBytes(live.used_bytes)}</span>
            <span>共 {live.total_quota_gb} GB</span>
          </div>
          <MiniBar
            value={(live.used_bytes / (live.total_quota_gb * 1e9)) * 100}
            color={live.used_bytes / (live.total_quota_gb * 1e9) > 0.9 ? "bg-red-500" : "bg-cyan-500"}
          />
        </div>
      )}

      {/* Footer */}
      <div className="flex justify-between text-xs pt-1 border-t border-slate-700/50">
        <span className="text-slate-500">¥{server.monthly_rent}/月</span>
        {(live?.expire_at ?? server.expire_at) ? (
          <span className={expireDateClass(live?.expire_at ?? server.expire_at)}>
            {(live?.expire_at ?? server.expire_at)!.slice(0, 10)} 到期
          </span>
        ) : (
          <span className="text-slate-600">无到期日</span>
        )}
      </div>
    </div>
  );
}
