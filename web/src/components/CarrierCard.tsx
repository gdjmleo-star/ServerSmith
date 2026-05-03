import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface CarrierSummary {
  carrier_type: string;
  server_count: number;
  total_quota_gb: number;
  used_gb: number;
  remaining_gb: number;
  usage_percent: number;
  total_rent: number;
  cost_per_gb: number;
  abnormal_count: number;
}

function formatBytes(gb: number): string {
  if (gb >= 1000) return `${(gb / 1000).toFixed(1)} TB`;
  return `${gb.toFixed(1)} GB`;
}

function usageColor(pct: number): string {
  if (pct >= 95) return "text-red-400";
  if (pct >= 80) return "text-yellow-400";
  return "text-green-400";
}

function progressBarColor(pct: number): string {
  if (pct >= 95) return "bg-red-500";
  if (pct >= 80) return "bg-yellow-500";
  return "bg-green-500";
}

interface Props {
  data: CarrierSummary;
}

export default function CarrierCard({ data }: Props) {
  const pct = Math.min(data.usage_percent, 100);
  return (
    <Card className="bg-zinc-950 border-zinc-800 hover:border-zinc-700 transition-colors">
      <CardHeader className="p-4 pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-gray-200">
            {data.carrier_type}
          </CardTitle>
          <Badge variant="outline" className="text-xs border-zinc-700 text-gray-400">
            {data.server_count} 节点
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="p-4 pt-2 space-y-3">
        <div className="space-y-1">
          <div className="flex justify-between text-xs text-gray-500">
            <span>{formatBytes(data.used_gb)} 已用</span>
            <span className={usageColor(pct)}>{pct.toFixed(0)}%</span>
          </div>
          <div className="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full transition-all ${progressBarColor(pct)}`}
              style={{ width: `${pct}%` }}
            />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2 text-xs">
          <div>
            <span className="text-gray-600">剩余</span>
            <p className="text-gray-300">{formatBytes(Math.max(0, data.remaining_gb))}</p>
          </div>
          <div>
            <span className="text-gray-600">总额度</span>
            <p className="text-gray-300">{formatBytes(data.total_quota_gb)}</p>
          </div>
          <div>
            <span className="text-gray-600">月租</span>
            <p className="text-gray-300">¥{data.total_rent.toFixed(0)}</p>
          </div>
          <div>
            <span className="text-gray-600">每 GB 成本</span>
            <p className="text-gray-300">¥{data.cost_per_gb.toFixed(3)}</p>
          </div>
        </div>
        {data.abnormal_count > 0 && (
          <p className="text-xs text-red-400">⚠ {data.abnormal_count} 节点异常</p>
        )}
      </CardContent>
    </Card>
  );
}
