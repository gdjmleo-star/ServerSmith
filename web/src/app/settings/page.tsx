"use client";
import { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/hooks/useAuth";

interface UserInfo {
  id: number;
  username: string;
  created_at: string;
}

function Toast({ msg, type }: { msg: string; type: "ok" | "err" }) {
  return (
    <div
      className={`fixed top-4 right-4 z-50 px-4 py-3 rounded-lg text-sm font-medium shadow-lg ${
        type === "ok" ? "bg-green-600 text-white" : "bg-red-600 text-white"
      }`}
    >
      {msg}
    </div>
  );
}

export default function SettingsPage() {
  const { authFetch, username: currentUser } = useAuth();

  // ── Change password state ──
  const [curPwd, setCurPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [confirmPwd, setConfirmPwd] = useState("");
  const [pwdLoading, setPwdLoading] = useState(false);

  // ── User list state ──
  const [users, setUsers] = useState<UserInfo[]>([]);
  const [usersLoading, setUsersLoading] = useState(true);
  const [newUsername, setNewUsername] = useState("");
  const [newUserPwd, setNewUserPwd] = useState("");
  const [createLoading, setCreateLoading] = useState(false);

  // ── Toast ──
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);
  const showToast = (msg: string, type: "ok" | "err") => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const loadUsers = useCallback(async () => {
    setUsersLoading(true);
    try {
      const res = await authFetch("/api/users");
      if (res.ok) {
        const data = await res.json();
        setUsers(data.data ?? data);
      }
    } finally {
      setUsersLoading(false);
    }
  }, [authFetch]);

  useEffect(() => { loadUsers(); }, [loadUsers]);

  // ── Change password ──
  const handleChangePwd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPwd !== confirmPwd) { showToast("两次输入的新密码不一致", "err"); return; }
    if (newPwd.length < 4) { showToast("新密码至少 4 位", "err"); return; }
    setPwdLoading(true);
    try {
      const res = await authFetch("/api/users/change-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ current_password: curPwd, new_password: newPwd }),
      });
      const data = await res.json();
      if (res.ok) {
        showToast("密码修改成功，下次登录生效", "ok");
        setCurPwd(""); setNewPwd(""); setConfirmPwd("");
      } else {
        showToast(data.error ?? "修改失败", "err");
      }
    } finally {
      setPwdLoading(false);
    }
  };

  // ── Create user ──
  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newUsername.trim() || !newUserPwd) { showToast("用户名和密码不能为空", "err"); return; }
    setCreateLoading(true);
    try {
      const res = await authFetch("/api/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: newUsername.trim(), password: newUserPwd }),
      });
      const data = await res.json();
      if (res.ok) {
        showToast(`用户 ${newUsername} 创建成功`, "ok");
        setNewUsername(""); setNewUserPwd("");
        loadUsers();
      } else {
        showToast(data.error ?? "创建失败", "err");
      }
    } finally {
      setCreateLoading(false);
    }
  };

  // ── Delete user ──
  const handleDeleteUser = async (user: UserInfo) => {
    if (!confirm(`确认删除用户「${user.username}」？此操作不可撤销。`)) return;
    const res = await authFetch(`/api/users/${user.id}`, { method: "DELETE" });
    const data = await res.json();
    if (res.ok) {
      showToast(`用户 ${user.username} 已删除`, "ok");
      loadUsers();
    } else {
      showToast(data.error ?? "删除失败", "err");
    }
  };

  return (
    <div className="space-y-8 max-w-2xl">
      {toast && <Toast msg={toast.msg} type={toast.type} />}

      <h1 className="text-xl font-bold text-white">系统设置</h1>

      {/* ── Change Password ── */}
      <section className="bg-slate-800/60 border border-slate-700 rounded-xl p-6 space-y-4">
        <h2 className="text-white font-semibold text-base">修改密码</h2>
        <p className="text-slate-400 text-sm">当前登录账号：<span className="text-blue-400 font-mono">{currentUser}</span></p>
        <form onSubmit={handleChangePwd} className="space-y-3">
          <div>
            <label className="text-slate-400 text-xs block mb-1">当前密码</label>
            <input
              type="password"
              value={curPwd}
              onChange={(e) => setCurPwd(e.target.value)}
              className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="输入当前密码"
              required
            />
          </div>
          <div>
            <label className="text-slate-400 text-xs block mb-1">新密码</label>
            <input
              type="password"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
              className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="至少 4 位"
              required
            />
          </div>
          <div>
            <label className="text-slate-400 text-xs block mb-1">确认新密码</label>
            <input
              type="password"
              value={confirmPwd}
              onChange={(e) => setConfirmPwd(e.target.value)}
              className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="再次输入新密码"
              required
            />
          </div>
          <button
            type="submit"
            disabled={pwdLoading}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm rounded-lg font-medium transition-colors"
          >
            {pwdLoading ? "保存中..." : "保存新密码"}
          </button>
        </form>
      </section>

      {/* ── User Management ── */}
      <section className="bg-slate-800/60 border border-slate-700 rounded-xl p-6 space-y-4">
        <h2 className="text-white font-semibold text-base">管理员账号管理</h2>

        {/* User list */}
        <div className="space-y-2">
          {usersLoading ? (
            <p className="text-slate-500 text-sm">加载中...</p>
          ) : (
            users.map((u) => (
              <div
                key={u.id}
                className="flex items-center justify-between bg-slate-900 border border-slate-700 rounded-lg px-4 py-3"
              >
                <div>
                  <span className="text-white text-sm font-mono">{u.username}</span>
                  {u.username === currentUser && (
                    <span className="ml-2 text-xs text-blue-400 bg-blue-900/40 px-2 py-0.5 rounded-full">当前账号</span>
                  )}
                  <p className="text-slate-500 text-xs mt-0.5">创建于 {u.created_at.slice(0, 10)}</p>
                </div>
                {u.username !== currentUser && (
                  <button
                    onClick={() => handleDeleteUser(u)}
                    className="text-xs text-red-400 hover:text-red-300 hover:bg-red-900/30 px-3 py-1.5 rounded-lg transition-colors"
                  >
                    删除
                  </button>
                )}
              </div>
            ))
          )}
        </div>

        {/* Add user */}
        <div className="border-t border-slate-700 pt-4">
          <h3 className="text-slate-300 text-sm font-medium mb-3">添加管理员账号</h3>
          <form onSubmit={handleCreateUser} className="space-y-3">
            <div className="flex gap-3">
              <input
                type="text"
                value={newUsername}
                onChange={(e) => setNewUsername(e.target.value)}
                className="flex-1 bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
                placeholder="用户名"
              />
              <input
                type="password"
                value={newUserPwd}
                onChange={(e) => setNewUserPwd(e.target.value)}
                className="flex-1 bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
                placeholder="密码（至少 4 位）"
              />
              <button
                type="submit"
                disabled={createLoading}
                className="px-4 py-2 bg-green-700 hover:bg-green-600 disabled:opacity-50 text-white text-sm rounded-lg font-medium transition-colors whitespace-nowrap"
              >
                {createLoading ? "添加中..." : "添加"}
              </button>
            </div>
          </form>
        </div>
      </section>
    </div>
  );
}
