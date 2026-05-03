"use client";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";
import { HistoryPoint } from "@/hooks/useHistory";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";

interface Props {
  data: HistoryPoint[];
  loading: boolean;
  error: string | null;
  days: number;
  onDaysChange: (d: number) => void;
  dayOptions: number[];
}

function fmtGB(bytes: number) {
  const gb = bytes / 1e9;
  return `${gb.toFixed(2)} GB`;
}

export function TrafficChart({ data, loading, error, days, onDaysChange, dayOptions }: Props) {
  const router = useRouter();

  const chartData = data.map(d => ({
    date: d.date.slice(5, 10),
    入流量: parseFloat((d.bytes_in / 1e9).toFixed(3)),
    出流量: parseFloat((d.bytes_out / 1e9).toFixed(3)),
  }));

  return (
    <div className="min-h-screen bg-slate-900 text-white p-6 space-y-5">
      <div className="flex items-center gap-3">
        <button onClick={() => router.push("/servers")} className="text-slate-400 hover:text-white text-sm">← 返回</button>
        <h1 className="text-xl font-bold">流量历史</h1>
      </div>

      <div className="flex items-center gap-2">
        {dayOptions.map(d => (
          <Button key={d} size="sm" variant={days === d ? "default" : "ghost"} onClick={() => onDaysChange(d)}>
            {d} 天
          </Button>
        ))}
      </div>

      <div className="bg-slate-800 border border-slate-700 rounded-xl p-6 space-y-6">
        {loading && <div className="flex items-center justify-center h-48 text-slate-400">加载中…</div>}
        {error && <div className="flex items-center justify-center h-48 text-red-400">无法连接到服务器，请稍后重试</div>}
        {!loading && !error && data.length === 0 && (
          <div className="flex items-center justify-center h-48 text-slate-500">暂无历史数据</div>
        )}
        {!loading && !error && data.length > 0 && (
          <>
            <p className="text-sm text-slate-400">每日流量 (GB)</p>
            <ResponsiveContainer width="100%" height={260}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="date" tick={{ fill: "#94a3b8", fontSize: 11 }} />
                <YAxis tick={{ fill: "#94a3b8", fontSize: 11 }} unit=" GB" />
                <Tooltip
                  contentStyle={{ background: "#1e293b", border: "1px solid #475569", borderRadius: 8 }}
                  labelStyle={{ color: "#e2e8f0" }}
                  itemStyle={{ color: "#94a3b8" }}
                  formatter={(val) => typeof val === "number" ? fmtGB(val * 1e9) : String(val)}
                />
                <Legend wrapperStyle={{ color: "#94a3b8" }} />
                <Line type="monotone" dataKey="入流量" stroke="#06b6d4" dot={false} strokeWidth={2} />
                <Line type="monotone" dataKey="出流量" stroke="#f97316" dot={false} strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </>
        )}
      </div>
    </div>
  );
}
