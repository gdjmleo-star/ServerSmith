"use client";

import { Separator } from "@/components/ui/separator";
import StatCard from "@/components/StatCard";
import CarrierCard from "@/components/CarrierCard";
import { CostSummary } from "@/components/CostSummary";
import { useDashboard } from "@/hooks/useDashboard";
import { useAlerts } from "@/hooks/useAlerts";

function formatBytes(gb: number): string {
  if (gb >= 1000) return `${(gb / 1000).toFixed(1)} TB`;
  return `${gb.toFixed(1)} GB`;
}

export default function DashboardPage() {
  const { dashboard, carriers, loading } = useDashboard();
  const { alerts } = useAlerts();

  if (loading) {
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <p className="text-gray-500 animate-pulse">正在加载驾驶舱数据...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white p-4 md:p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">ServerSmith</h1>
            <p className="text-sm text-gray-500 mt-1">运维驾驶舱 · 实时总览</p>
          </div>
          <a
            href="/servers"
            className="text-sm text-blue-400 hover:text-blue-300 border border-blue-800 hover:border-blue-600 rounded-lg px-3 py-1.5 transition"
          >
            服务器管理 →
          </a>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3 mb-8">
        <StatCard title="总服务器" value={dashboard?.total_servers ?? 0} />
        <StatCard title="在线" value={dashboard?.online_count ?? 0} color="text-green-400" />
        <StatCard title="离线" value={dashboard?.offline_count ?? 0} color="text-red-400" />
        <StatCard title="月总支出" value={`¥${(dashboard?.total_monthly_rent ?? 0).toFixed(0)}`} />
        <StatCard
          title="总剩余流量"
          value={
            dashboard
              ? formatBytes(Math.max(0, dashboard.total_quota_gb - dashboard.total_used_gb))
              : "0 GB"
          }
        />
        <StatCard
          title="未读告警"
          value={dashboard?.alerts_unread ?? 0}
          color={dashboard && dashboard.alerts_unread > 0 ? "text-red-400" : ""}
        />
      </div>

      {/* Cost Summary */}
      <div className="mb-8">
        <h2 className="text-lg font-medium mb-4 text-gray-300">成本分析</h2>
        <CostSummary />
      </div>

      {/* Carrier Cards */}
      <h2 className="text-lg font-medium mb-4 text-gray-300">线路流量汇总</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        {carriers.length === 0 && (
          <p className="text-gray-600 col-span-full text-sm">
            暂无线路数据，请先添加流量节点
          </p>
        )}
        {carriers.map((c) => (
          <CarrierCard key={c.carrier_type} data={c} />
        ))}
      </div>

      {/* Alerts */}
      <div className="mt-8">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-medium text-gray-300">最近告警</h2>
          <a href="/alerts" className="text-sm text-blue-400 hover:text-blue-300">查看全部 →</a>
        </div>
        {alerts.length === 0 ? (
          <p className="text-gray-600 text-sm">暂无告警</p>
        ) : (
          <div className="space-y-2">
            {alerts.slice(0, 5).map((a) => (
              <div key={a.id} className={`flex items-start gap-3 px-4 py-2 rounded-lg border ${
                a.acknowledged ? "border-slate-800 bg-slate-900/30" : "border-red-900/50 bg-red-950/20"
              }`}>
                <span className="text-xs text-slate-500 whitespace-nowrap mt-0.5">{a.alert_type}</span>
                <span className="text-sm text-slate-300 flex-1">{a.message}</span>
                {!a.acknowledged && <span className="text-red-400 text-xs">未读</span>}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <Separator className="my-8 bg-gray-800" />
      <p className="text-xs text-gray-700 text-center">
        ServerSmith v1.0.0 · 数据每 30 秒自动更新
      </p>
    </div>
  );
}
