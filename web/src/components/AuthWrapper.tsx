"use client";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { AuthProvider, useAuth } from "@/hooks/useAuth";

// Public paths that don't require login
const PUBLIC_PATHS = ["/login", "/public"];

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isLoggedIn } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (isLoggedIn) return;
    // Allow public paths without redirect
    if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) return;
    // Redirect to login
    router.push("/login");
  }, [isLoggedIn, pathname, router]);

  return <>{children}</>;
}

export function AuthWrapper({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <AuthGuard>{children}</AuthGuard>
    </AuthProvider>
  );
}
