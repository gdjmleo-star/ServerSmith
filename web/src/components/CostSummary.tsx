"use client";
import { useCostData } from "@/hooks/useCostData";

export function CostSummary() {
  const { totalRent, costPerGB, loading } = useCostData();

  if (loading) return null;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between px-4 py-3 bg-slate-800 border border-slate-700 rounded-xl">
        <span className="text-slate-400 text-sm">月度总支出</span>
        <span className="text-white font-semibold text-lg">¥{totalRent.toFixed(2)}</span>
      </div>

      {costPerGB.length > 0 && (
        <div className="bg-slate-800 border border-slate-700 rounded-xl overflow-hidden">
          <div className="px-4 py-2 border-b border-slate-700">
            <span className="text-slate-400 text-sm">各线路成本</span>
          </div>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700">
                <th className="px-4 py-2 text-left text-slate-400 font-normal">线路</th>
                <th className="px-4 py-2 text-right text-slate-400 font-normal">月租</th>
                <th className="px-4 py-2 text-right text-slate-400 font-normal">总额度</th>
                <th className="px-4 py-2 text-right text-slate-400 font-normal">单价</th>
              </tr>
            </thead>
            <tbody>
              {costPerGB.map((item) => (
                <tr key={item.carrier} className="border-b border-slate-800 hover:bg-slate-700/30">
                  <td className="px-4 py-2 text-white font-medium">{item.carrier}</td>
                  <td className="px-4 py-2 text-right text-slate-300">¥{item.monthly_rent.toFixed(0)}/月</td>
                  <td className="px-4 py-2 text-right text-slate-300">{item.total_quota_gb.toFixed(0)} GB</td>
                  <td className="px-4 py-2 text-right">
                    <span className={item.cost_per_gb > 1 ? "text-red-400" : "text-green-400"}>
                      ¥{item.cost_per_gb.toFixed(3)}/GB
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
