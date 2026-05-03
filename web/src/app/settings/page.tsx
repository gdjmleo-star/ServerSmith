"use client";
import { useState } from "react";
import { ChangePasswordForm } from "@/components/ChangePasswordForm";
import { UserManagement } from "@/components/UserManagement";

function Toast({ msg, type }: { msg: string; type: "ok" | "err" }) {
  return (
    <div className={`fixed top-4 right-4 z-50 px-4 py-3 rounded-lg text-sm font-medium shadow-lg ${
      type === "ok" ? "bg-green-600 text-white" : "bg-red-600 text-white"
    }`}>
      {msg}
    </div>
  );
}

export default function SettingsPage() {
  const [toast, setToast] = useState<{ msg: string; type: "ok" | "err" } | null>(null);

  const showToast = (msg: string, type: "ok" | "err") => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  return (
    <div className="space-y-8 max-w-2xl">
      {toast && <Toast msg={toast.msg} type={toast.type} />}
      <h1 className="text-xl font-bold text-white">系统设置</h1>
      <ChangePasswordForm onSuccess={() => showToast("密码修改成功，下次登录生效", "ok")} />
      <UserManagement onToast={showToast} />
    </div>
  );
}
