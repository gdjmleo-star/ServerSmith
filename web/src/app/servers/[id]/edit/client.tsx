"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { ServerForm } from "@/components/ServerForm";
import { AlertConfigForm } from "@/components/AlertConfigForm";
import { useServers, ServerFormData, Server } from "@/hooks/useServers";
import { Button } from "@/components/ui/button";

function toForm(s: Server): ServerFormData {
  return {
    name: s.name, host: s.host, ssh_port: s.ssh_port, ssh_user: s.ssh_user,
    server_type: s.server_type, carrier_type: s.carrier_type ?? "",
    monthly_rent: s.monthly_rent, plan_type: s.plan_type,
    total_quota: s.total_quota ?? null, initial_used_gb: null,
    cycle_day: s.cycle_day ?? null, cycle_time: s.cycle_time ?? "02:00",
    expire_at: s.expire_at ? s.expire_at.slice(0, 10) : "",
    expire_notify_days: s.expire_notify_days ?? 7,
  };
}

export default function EditServerPage() {
  const id = Number(useParams()?.id);
  const { getServer, updateServer } = useServers();
  const [initial, setInitial] = useState<ServerFormData | undefined>();
  const [alertOpen, setAlertOpen] = useState(false);

  useEffect(() => { getServer(id).then((s) => setInitial(toForm(s))); }, [id]);

  if (!initial) return <div className="min-h-screen bg-slate-900 flex items-center justify-center text-slate-400">加载中…</div>;

  return (
    <>
      {alertOpen && <AlertConfigForm serverId={id} onClose={() => setAlertOpen(false)} />}
      <div className="max-w-2xl mx-auto p-6">
        <div className="flex justify-end mb-2">
          <Button size="sm" variant="outline" className="border-slate-600 text-slate-300 hover:text-white" onClick={() => setAlertOpen(true)}>
            🔔 告警配置
          </Button>
        </div>
        <ServerForm initial={initial} onSubmit={(d) => updateServer(id, d)} submitLabel="保存修改" mode="edit" />
      </div>
    </>
  );
}
