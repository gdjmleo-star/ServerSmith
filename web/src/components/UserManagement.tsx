"use client";
import { useState, useEffect, useCallback } from "react";
import { useAuth, authFetch } from "@/hooks/useAuth";

interface UserInfo {
  id: number;
  username: string;
  created_at: string;
}

interface Props {
  onToast?: (msg: string, type: "ok" | "err") => void;
}

export function UserManagement({ onToast }: Props) {
  const { username: currentUser } = useAuth();
  const [users, setUsers] = useState<UserInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [newUsername, setNewUsername] = useState("");
  const [newUserPwd, setNewUserPwd] = useState("");
  const [createLoading, setCreateLoading] = useState(false);

  const notify = (msg: string, type: "ok" | "err") => onToast?.(msg, type);

  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const data = await authFetch<UserInfo[]>("/api/users");
      setUsers(Array.isArray(data) ? data : []);
    } catch {
      notify("加载用户列表失败", "err");
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => { loadUsers(); }, [loadUsers]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newUsername.trim() || !newUserPwd) { notify("用户名和密码不能为空", "err"); return; }
    setCreateLoading(true);
    try {
      await authFetch("/api/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: newUsername.trim(), password: newUserPwd }),
      });
      notify(`用户 ${newUsername} 创建成功`, "ok");
      setNewUsername(""); setNewUserPwd("");
      loadUsers();
    } catch (err: unknown) {
      const raw = err instanceof Error ? err.message : "创建失败";
      try { notify(JSON.parse(raw).error ?? raw, "err"); } catch { notify(raw, "err"); }
    } finally {
      setCreateLoading(false);
    }
  };

  const handleDelete = async (user: UserInfo) => {
    if (!confirm(`确认删除用户「${user.username}」？此操作不可撤销。`)) return;
    try {
      await authFetch(`/api/users/${user.id}`, { method: "DELETE" });
      notify(`用户 ${user.username} 已删除`, "ok");
      loadUsers();
    } catch (err: unknown) {
      const raw = err instanceof Error ? err.message : "删除失败";
      notify(raw, "err");
    }
  };

  return (
    <section className="bg-slate-800/60 border border-slate-700 rounded-xl p-6 space-y-4">
      <h2 className="text-white font-semibold text-base">管理员账号管理</h2>

      <div className="space-y-2">
        {loading ? (
          <p className="text-slate-500 text-sm">加载中...</p>
        ) : (
          users.map((u) => (
            <div key={u.id}
              className="flex items-center justify-between bg-slate-900 border border-slate-700 rounded-lg px-4 py-3">
              <div>
                <span className="text-white text-sm font-mono">{u.username}</span>
                {u.username === currentUser && (
                  <span className="ml-2 text-xs text-blue-400 bg-blue-900/40 px-2 py-0.5 rounded-full">
                    当前账号
                  </span>
                )}
                <p className="text-slate-500 text-xs mt-0.5">创建于 {u.created_at.slice(0, 10)}</p>
              </div>
              {u.username !== currentUser && (
                <button onClick={() => handleDelete(u)}
                  className="text-xs text-red-400 hover:text-red-300 hover:bg-red-900/30 px-3 py-1.5 rounded-lg transition-colors">
                  删除
                </button>
              )}
            </div>
          ))
        )}
      </div>

      <div className="border-t border-slate-700 pt-4">
        <h3 className="text-slate-300 text-sm font-medium mb-3">添加管理员账号</h3>
        <form onSubmit={handleCreate} className="space-y-3">
          <div className="flex gap-3">
            <input type="text" value={newUsername} onChange={(e) => setNewUsername(e.target.value)}
              className="flex-1 bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="用户名" />
            <input type="password" value={newUserPwd} onChange={(e) => setNewUserPwd(e.target.value)}
              className="flex-1 bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-blue-500"
              placeholder="密码（至少 4 位）" />
            <button type="submit" disabled={createLoading}
              className="px-4 py-2 bg-green-700 hover:bg-green-600 disabled:opacity-50 text-white text-sm rounded-lg font-medium transition-colors whitespace-nowrap">
              {createLoading ? "添加中..." : "添加"}
            </button>
          </div>
        </form>
      </div>
    </section>
  );
}
