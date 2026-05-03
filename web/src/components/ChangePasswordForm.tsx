"use client";
import { useState } from "react";
import { useAuth, authFetch } from "@/hooks/useAuth";

interface Props {
  onSuccess?: () => void;
}

export function ChangePasswordForm({ onSuccess }: Props) {
  const { username: currentUser } = useAuth();
  const [curPwd, setCurPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [confirmPwd, setConfirmPwd] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const inputCls =
    "w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (newPwd !== confirmPwd) { setError("两次输入的新密码不一致"); return; }
    if (newPwd.length < 4) { setError("新密码至少 4 位"); return; }
    setLoading(true);
    try {
      await authFetch("/api/users/change-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ current_password: curPwd, new_password: newPwd }),
      });
      setSuccess(true);
      setCurPwd(""); setNewPwd(""); setConfirmPwd("");
      onSuccess?.();
      setTimeout(() => setSuccess(false), 3000);
    } catch (err: unknown) {
      const raw = err instanceof Error ? err.message : "修改失败";
      try { setError(JSON.parse(raw).error ?? raw); } catch { setError(raw); }
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="bg-slate-800/60 border border-slate-700 rounded-xl p-6 space-y-4">
      <h2 className="text-white font-semibold text-base">修改密码</h2>
      <p className="text-slate-400 text-sm">
        当前登录账号：<span className="text-blue-400 font-mono">{currentUser}</span>
      </p>
      {success && (
        <p className="text-green-400 text-sm">✅ 密码修改成功，下次登录生效</p>
      )}
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label className="text-slate-400 text-xs block mb-1">当前密码</label>
          <input type="password" value={curPwd} onChange={(e) => setCurPwd(e.target.value)}
            className={inputCls} placeholder="输入当前密码" required />
        </div>
        <div>
          <label className="text-slate-400 text-xs block mb-1">新密码</label>
          <input type="password" value={newPwd} onChange={(e) => setNewPwd(e.target.value)}
            className={inputCls} placeholder="至少 4 位" required />
        </div>
        <div>
          <label className="text-slate-400 text-xs block mb-1">确认新密码</label>
          <input type="password" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)}
            className={inputCls} placeholder="再次输入新密码" required />
        </div>
        <button type="submit" disabled={loading}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm rounded-lg font-medium transition-colors">
          {loading ? "保存中..." : "保存新密码"}
        </button>
      </form>
    </section>
  );
}
