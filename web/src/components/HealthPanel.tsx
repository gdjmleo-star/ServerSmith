"use client";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
} from "recharts";
import { useServerSnapshots } from "@/hooks/useServerSnapshots";
import { formatRate } from "@/lib/format";

interface Props {
  serverId: number;
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  const pct = Math.min(100, Math.max(0, value));
  return (
    <div className="bg-slate-900 border border-slate-700 rounded-xl p-4 space-y-2">
      <div className="flex justify-between items-baseline">
        <span className="text-slate-400 text-sm">{label}</span>
        <span className="text-white text-2xl font-bold">{value.toFixed(1)}%</span>
      </div>
      <div className="w-full bg-slate-700 rounded-full h-2 overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

function formatTime(dateStr: string): string {
  try {
    const d = new Date(dateStr);
    return `${d.getHours().toString().padStart(2, "0")}:${d.getMinutes().toString().padStart(2, "0")}`;
  } catch {
    return dateStr;
  }
}

// Compute net rate from consecutive snapshots (bytes/s)
function computeRates(snapshots: ReturnType<typeof useServerSnapshots>["snapshots"]) {
  return snapshots.map((s, i) => {
    if (i === 0) return { ...s, in_rate: 0, out_rate: 0, time: formatTime(s.date) };
    const prev = snapshots[i - 1];
    const diffSec = (new Date(s.date).getTime() - new Date(prev.date).getTime()) / 1000;
    const inRate = diffSec > 0 ? Math.max(0, s.bytes_in - prev.bytes_in) / diffSec : 0;
    const outRate = diffSec > 0 ? Math.max(0, s.bytes_out - prev.bytes_out) / diffSec : 0;
    return { ...s, in_rate: inRate / 1024, out_rate: outRate / 1024, time: formatTime(s.date) };
  });
}

export function HealthPanel({ serverId }: Props) {
  const { snapshots, loading } = useServerSnapshots(serverId, 24);

  if (loading) return <div className="text-slate-500 text-sm py-6 text-center">加载健康数据…</div>;
  if (snapshots.length === 0) return (
    <div className="text-slate-500 text-sm py-6 text-center">暂无快照数据，探针上报后自动显示</div>
  );

  const latest = snapshots[snapshots.length - 1];
  const chartData = computeRates(snapshots);

  return (
    <div className="space-y-4">
      {/* Current stats */}
      <div className="grid grid-cols-3 gap-3">
        <StatCard label="CPU" value={latest.avg_cpu} color="bg-blue-500" />
        <StatCard label="内存" value={latest.avg_mem} color="bg-green-500" />
        <StatCard label="硬盘" value={latest.avg_disk} color="bg-purple-500" />
      </div>

      {/* 24h trend chart */}
      <div className="bg-slate-900 border border-slate-700 rounded-xl p-4">
        <h3 className="text-slate-300 text-sm font-medium mb-3">24 小时趋势</h3>
        <ResponsiveContainer width="100%" height={220}>
          <LineChart data={chartData} margin={{ top: 4, right: 8, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="time" tick={{ fill: "#94a3b8", fontSize: 10 }} interval="preserveStartEnd" />
            <YAxis tick={{ fill: "#94a3b8", fontSize: 10 }} />
            <Tooltip
              contentStyle={{ background: "#1e293b", border: "1px solid #334155", borderRadius: 8 }}
              labelStyle={{ color: "#cbd5e1" }}
              formatter={(value, name) => {
                const v = typeof value === "number" ? value : 0;
                if (name === "入站(KB/s)" || name === "出站(KB/s)") return [`${v.toFixed(1)} KB/s`, String(name)];
                return [`${v.toFixed(1)}%`, String(name)];
              }}
            />
            <Legend wrapperStyle={{ fontSize: 11, color: "#94a3b8" }} />
            <Line type="monotone" dataKey="avg_cpu" name="CPU%" stroke="#3b82f6" dot={false} strokeWidth={1.5} />
            <Line type="monotone" dataKey="avg_mem" name="内存%" stroke="#22c55e" dot={false} strokeWidth={1.5} />
            <Line type="monotone" dataKey="in_rate" name="入站(KB/s)" stroke="#06b6d4" dot={false} strokeWidth={1.5} />
            <Line type="monotone" dataKey="out_rate" name="出站(KB/s)" stroke="#f97316" dot={false} strokeWidth={1.5} />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
