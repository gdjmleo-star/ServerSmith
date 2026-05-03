"use client";
import { useEffect } from "react";

interface Props {
  message: string;
  type: "ok" | "err";
  onClose: () => void;
  durationMs?: number;
}

export function Toast({ message, type, onClose, durationMs = 3000 }: Props) {
  useEffect(() => {
    const t = setTimeout(onClose, durationMs);
    return () => clearTimeout(t);
  }, [onClose, durationMs]);

  const bg = type === "ok" ? "bg-green-800 border-green-600" : "bg-red-900 border-red-600";

  return (
    <div
      className={`fixed top-5 right-5 z-[100] flex items-center gap-3 px-4 py-3 rounded-xl border text-white text-sm shadow-xl ${bg}`}
    >
      <span>{type === "ok" ? "✓" : "✗"}</span>
      <span>{message}</span>
      <button onClick={onClose} className="ml-2 opacity-60 hover:opacity-100 text-lg leading-none">
        ×
      </button>
    </div>
  );
}
