"use client";
import { useEffect, useState } from "react";
import { useAlertConfigs } from "@/hooks/useAlerts";
import { Button } from "@/components/ui/button";
import { Toast } from "@/components/Toast";

interface Props {
  serverId: number;
  onClose: () => void;
}

const DEFAULTS = {
  offline: { enabled: true, threshold: "5" },
  traffic_usage: { enabled: true, threshold: "90" },
};

export function AlertConfigForm({ serverId, onClose }: Props) {
  const { configs, saveConfig } = useAlertConfigs(serverId);
  const [offline, setOffline] = useState(DEFAULTS.offline);
  const [traffic, setTraffic] = useState(DEFAULTS.traffic_usage);
  const [saving, setSaving] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);

  useEffect(() => {
    for (const cfg of configs) {
      if (cfg.alert_type === "offline") {
        setOffline({ enabled: cfg.enabled, threshold: cfg.threshold });
      } else if (cfg.alert_type === "traffic_usage") {
        setTraffic({ enabled: cfg.enabled, threshold: cfg.threshold });
      }
    }
  }, [configs]);

  async function handleSave() {
    setSaving(true);
    try {
      await saveConfig({ server_id: serverId, alert_type: "offline", ...offline });
      await saveConfig({ server_id: serverId, alert_type: "traffic_usage", ...traffic });
      setToast({ msg: "告警配置已保存", type: "ok" });
      setTimeout(onClose, 800);
    } catch (e) {
      setToast({ msg: `保存失败：${(e as Error).message}`, type: "err" });
    } finally {
      setSaving(false);
    }
  }

  const inputCls = "bg-slate-700 border border-slate-600 rounded-lg px-2 py-1 text-white text-sm w-20 focus:outline-none focus:ring-2 focus:ring-blue-500";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      {toast && <Toast message={toast.msg} type={toast.type} onClose={() => setToast(null)} />}
      <div className="bg-slate-800 border border-slate-600 rounded-xl p-6 w-96 space-y-5">
        <h2 className="text-white font-semibold">告警配置</h2>

        {/* Offline alert */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-slate-300 text-sm font-medium">离线告警</span>
            <button
              onClick={() => setOffline(o => ({ ...o, enabled: !o.enabled }))}
              className={`w-10 h-5 rounded-full transition-colors ${offline.enabled ? "bg-blue-500" : "bg-slate-600"}`}
            >
              <div className={`w-4 h-4 bg-white rounded-full mx-auto transform transition-transform ${offline.enabled ? "translate-x-2.5" : "-translate-x-2.5"}`} />
            </button>
          </div>
          {offline.enabled && (
            <div className="flex items-center gap-2 text-sm text-slate-400">
              <span>连续离线超过</span>
              <input type="number" className={inputCls} value={offline.threshold}
                onChange={e => setOffline(o => ({ ...o, threshold: e.target.value }))} />
              <span>分钟触发告警</span>
            </div>
          )}
        </div>

        {/* Traffic usage alert */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-slate-300 text-sm font-medium">流量超额告警</span>
            <button
              onClick={() => setTraffic(t => ({ ...t, enabled: !t.enabled }))}
              className={`w-10 h-5 rounded-full transition-colors ${traffic.enabled ? "bg-blue-500" : "bg-slate-600"}`}
            >
              <div className={`w-4 h-4 bg-white rounded-full mx-auto transform transition-transform ${traffic.enabled ? "translate-x-2.5" : "-translate-x-2.5"}`} />
            </button>
          </div>
          {traffic.enabled && (
            <div className="flex items-center gap-2 text-sm text-slate-400">
              <span>流量使用超过</span>
              <input type="number" className={inputCls} value={traffic.threshold}
                onChange={e => setTraffic(t => ({ ...t, threshold: e.target.value }))} />
              <span>% 触发告警</span>
            </div>
          )}
        </div>

        <div className="flex gap-3 justify-end pt-2">
          <Button variant="ghost" onClick={onClose}>取消</Button>
          <Button disabled={saving} onClick={handleSave}>{saving ? "保存中…" : "保存配置"}</Button>
        </div>
      </div>
    </div>
  );
}
