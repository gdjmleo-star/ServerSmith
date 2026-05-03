"use client";
import { AgentInstallPanel } from "@/components/AgentInstallPanel";

export default function ProbesPage() {
  return (
    <div className="min-h-screen bg-slate-900 text-white p-6 space-y-4">
      <h1 className="text-xl font-bold">探针安装</h1>
      <AgentInstallPanel />
    </div>
  );
}
