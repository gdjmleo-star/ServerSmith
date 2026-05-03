"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ServerFormData } from "@/hooks/useServers";
import { Button } from "@/components/ui/button";
import { Toast } from "@/components/Toast";

interface Props {
  initial?: Partial<ServerFormData>;
  onSubmit: (data: ServerFormData) => Promise<void>;
  submitLabel: string;
  mode?: "create" | "edit";
}

const CARRIER_OPTIONS = ["4837", "CN2_GIA", "CN2_GT", "CMI", "CUII", "软银", "9929", "其他"];

const DEFAULTS: ServerFormData = {
  name: "", host: "", ssh_port: 22, ssh_user: "root",
  server_type: "node", carrier_type: "", monthly_rent: 0,
  plan_type: "traffic", total_quota: null, initial_used_gb: null,
  cycle_day: 1, cycle_time: "02:00", expire_at: "", expire_notify_days: 7,
};

const inputCls = "w-full bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500";

function Field({ label, required, children, hint }: {
  label: string; required?: boolean; children: React.ReactNode; hint?: string;
}) {
  return (
    <div className="space-y-1">
      <label className="text-sm text-slate-300">
        {label}{required && <span className="text-red-400 ml-0.5">*</span>}
      </label>
      {children}
      {hint && <p className="text-xs text-slate-500">{hint}</p>}
    </div>
  );
}

export function ServerForm({ initial, onSubmit, submitLabel, mode = "create" }: Props) {
  const router = useRouter();
  const [form, setForm] = useState<ServerFormData>({ ...DEFAULTS, ...initial });
  const [errors, setErrors] = useState<Partial<Record<keyof ServerFormData, string>>>({});
  const [submitting, setSubmitting] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);

  useEffect(() => { if (initial) setForm({ ...DEFAULTS, ...initial }); }, [initial]);

  function set<K extends keyof ServerFormData>(k: K, v: ServerFormData[K]) {
    setForm(f => ({ ...f, [k]: v }));
    if (errors[k]) setErrors(e => ({ ...e, [k]: undefined }));
  }

  const isTraffic = form.plan_type === "traffic";

  function validate(): boolean {
    const e: Partial<Record<keyof ServerFormData, string>> = {};
    if (!form.name.trim()) e.name = "名称不能为空";
    if (!form.host.trim()) e.host = "主机地址不能为空";
    if (form.monthly_rent <= 0) e.monthly_rent = "月租不能为空";
    if (isTraffic) {
      if (!form.carrier_type) e.carrier_type = "线路类型不能为空";
      if (!form.total_quota) e.total_quota = "流量额度不能为空";
      if (mode === "create" && !form.initial_used_gb && form.initial_used_gb !== 0) {
        e.initial_used_gb = "期初已用流量不能为空";
      }
      if (!form.cycle_day) e.cycle_day = "计费日不能为空";
    }
    setErrors(e);
    return Object.keys(e).length === 0;
  }

  async function handleSubmit(ev: React.FormEvent) {
    ev.preventDefault();
    if (!validate()) return;
    setSubmitting(true);
    try {
      await onSubmit(form);
      setToast({ msg: "保存成功", type: "ok" });
      setTimeout(() => router.push("/servers"), 800);
    } catch (e) {
      setToast({ msg: `保存失败：${(e as Error).message}`, type: "err" });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-900 text-white p-6">
      {toast && <Toast message={toast.msg} type={toast.type} onClose={() => setToast(null)} />}
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => router.push("/servers")} className="text-slate-400 hover:text-white text-sm">← 返回</button>
        <h1 className="text-xl font-bold">{mode === "create" ? "新增服务器" : "编辑服务器"}</h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5 max-w-xl bg-slate-800 border border-slate-700 rounded-xl p-6">
        <div className="grid grid-cols-2 gap-4">
          <Field label="服务器名称" required>
            <input className={inputCls} value={form.name} onChange={e => set("name", e.target.value)} placeholder="HK-01" />
            {errors.name && <p className="text-xs text-red-400">{errors.name}</p>}
          </Field>
          <Field label="主机 / IP" required>
            <input className={inputCls} value={form.host} onChange={e => set("host", e.target.value)} placeholder="1.2.3.4" />
            {errors.host && <p className="text-xs text-red-400">{errors.host}</p>}
          </Field>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <Field label="SSH 端口">
            <input type="number" className={inputCls} value={form.ssh_port}
              onChange={e => set("ssh_port", parseInt(e.target.value) || 22)} />
          </Field>
          <Field label="SSH 用户">
            <input className={inputCls} value={form.ssh_user} onChange={e => set("ssh_user", e.target.value)} />
          </Field>
        </div>

        <Field label="服务器类型">
          <select className={inputCls} value={form.plan_type}
            onChange={e => set("plan_type", e.target.value)}>
            <option value="traffic">流量节点 (按量计费)</option>
            <option value="no_limit">项目服务器 (不限流量)</option>
          </select>
        </Field>

        <Field label="月租费用 (¥)" required>
          <input type="number" step="0.01" className={inputCls} value={form.monthly_rent}
            onChange={e => set("monthly_rent", parseFloat(e.target.value) || 0)} />
          {errors.monthly_rent && <p className="text-xs text-red-400">{errors.monthly_rent}</p>}
        </Field>

        {isTraffic && (
          <>
            <Field label="线路类型" required>
              <select className={inputCls} value={form.carrier_type}
                onChange={e => set("carrier_type", e.target.value)}>
                <option value="">请选择</option>
                {CARRIER_OPTIONS.map(c => <option key={c} value={c}>{c}</option>)}
              </select>
              {errors.carrier_type && <p className="text-xs text-red-400">{errors.carrier_type}</p>}
            </Field>

            <Field label="月流量额度 (GB)" required>
              <input type="number" step="0.1" className={inputCls} value={form.total_quota ?? ""}
                onChange={e => set("total_quota", parseFloat(e.target.value) || null)} />
              {errors.total_quota && <p className="text-xs text-red-400">{errors.total_quota}</p>}
            </Field>

            {mode === "create" && (
              <Field label="期初已用流量 (GB)" required hint="续费服务器填入当前已用量，新开填 0">
                <input type="number" step="0.1" className={inputCls} value={form.initial_used_gb ?? ""}
                  onChange={e => set("initial_used_gb", parseFloat(e.target.value) ?? null)} />
                {errors.initial_used_gb && <p className="text-xs text-red-400">{errors.initial_used_gb}</p>}
              </Field>
            )}

            <div className="grid grid-cols-2 gap-4">
              <Field label="计费复位日 (1-31)" required>
                <input type="number" min={1} max={31} className={inputCls} value={form.cycle_day ?? ""}
                  onChange={e => set("cycle_day", parseInt(e.target.value) || null)} />
                {errors.cycle_day && <p className="text-xs text-red-400">{errors.cycle_day}</p>}
              </Field>
              <Field label="复位时间 (北京时间)">
                <input className={inputCls} value={form.cycle_time} placeholder="02:00"
                  onChange={e => set("cycle_time", e.target.value)} />
              </Field>
            </div>
          </>
        )}

        <div className="grid grid-cols-2 gap-4">
          <Field label="到期日" hint="不填则不跟踪">
            <input type="date" className={inputCls} value={form.expire_at}
              onChange={e => set("expire_at", e.target.value)} />
          </Field>
          <Field label="提前提醒天数">
            <input type="number" className={inputCls} value={form.expire_notify_days ?? 7}
              onChange={e => set("expire_notify_days", parseInt(e.target.value) || null)} />
          </Field>
        </div>

        <div className="flex gap-3 pt-2">
          <Button type="submit" disabled={submitting}>{submitting ? "保存中…" : submitLabel}</Button>
          <Button type="button" variant="ghost" onClick={() => router.push("/servers")}>取消</Button>
        </div>
      </form>
    </div>
  );
}
