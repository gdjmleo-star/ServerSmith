"use client";
import { useCallback, useEffect, useRef } from "react";
import { AuthProvider, useAuth } from "@/hooks/useAuth";
import { NotificationBadge } from "@/components/NotificationBadge";
import { TOPEMBY_URL } from "@/lib/config";

// Public paths that don't require login
const PUBLIC_PATHS = ["/login", "/public"];

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isLoggedIn, initialized } = useAuth();
  const redirected = useRef(false);

  useEffect(() => {
    if (!initialized) return;
    if (isLoggedIn) return;
    const path = window.location.pathname;
    if (PUBLIC_PATHS.some((p) => path.startsWith(p))) return;
    if (redirected.current) return;
    redirected.current = true;
    window.location.href = "/login/";
  }, [isLoggedIn, initialized]);

  return <>{children}</>;
}

function NavBar() {
  const { isLoggedIn, initialized, logout } = useAuth();
  const showAlerts = initialized && isLoggedIn;

  return (
    <nav className="flex items-center justify-between px-6 py-3 bg-slate-900 border-b border-slate-800">
      <a href="/" className="text-white font-semibold tracking-tight">ServerSmith</a>
      <div className="flex items-center gap-5">
        <a href="/servers" className="text-slate-400 hover:text-white text-sm transition">服务器管理</a>
        <a href="/probes" className="text-slate-400 hover:text-white text-sm transition">探针安装</a>
        <a href="/settings" className="text-slate-400 hover:text-white text-sm transition">系统设置</a>
        <a href={TOPEMBY_URL} target="_blank" rel="noopener noreferrer"
          className="text-slate-400 hover:text-white text-sm transition">TopEmby ↗</a>
        {showAlerts && <NotificationBadge />}
        {showAlerts && (
          <button onClick={logout}
            className="text-slate-500 hover:text-slate-300 text-xs border border-slate-700 px-2 py-0.5 rounded transition">
            退出
          </button>
        )}
      </div>
    </nav>
  );
}

export function AuthWrapper({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <NavBar />
      <AuthGuard>{children}</AuthGuard>
    </AuthProvider>
  );
}
