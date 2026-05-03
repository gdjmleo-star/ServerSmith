import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthWrapper } from "@/components/AuthWrapper";
import { NotificationBadge } from "@/components/NotificationBadge";
import { TOPEMBY_URL } from "@/lib/config";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "ServerSmith — 运维驾驶舱",
  description: "轻量流量额度管理与服务器运维面板",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background text-foreground`}
        style={{ fontFamily: geistSans.style.fontFamily }}
      >
        <AuthWrapper>
          <nav className="flex items-center justify-between px-6 py-3 bg-slate-900 border-b border-slate-800">
            <a href="/" className="text-white font-semibold tracking-tight">ServerSmith</a>
            <div className="flex items-center gap-5">
              <a href="/servers" className="text-slate-400 hover:text-white text-sm transition">服务器管理</a>
              <a href="/probes" className="text-slate-400 hover:text-white text-sm transition">探针管理</a>
              <a href={TOPEMBY_URL} target="_blank" rel="noopener noreferrer"
                className="text-slate-400 hover:text-white text-sm transition">TopEmby ↗</a>
              <NotificationBadge />
            </div>
          </nav>
          {children}
        </AuthWrapper>
      </body>
    </html>
  );
}
