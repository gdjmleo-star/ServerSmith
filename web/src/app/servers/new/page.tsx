"use client";
import { ServerForm } from "@/components/ServerForm";
import { useServers } from "@/hooks/useServers";
import { ServerFormData } from "@/hooks/useServers";

export default function NewServerPage() {
  const { createServer } = useServers();
  async function handleSubmit(data: ServerFormData) {
    await createServer(data);
  }
  return <ServerForm onSubmit={handleSubmit} submitLabel="创建服务器" mode="create" />;
}
