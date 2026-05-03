"use client";
import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import { API_BASE } from "@/lib/config";

interface AuthContextType {
  isLoggedIn: boolean;
  username: string | null;
  initialized: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  getToken: () => string | null;
}

const AuthContext = createContext<AuthContextType | null>(null);

const TOKEN_KEY = "serversmith_token";
const USER_KEY = "serversmith_user";

function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState<string | null>(null);
  const [initialized, setInitialized] = useState(false);

  const getToken = useCallback((): string | null => {
    return getStoredToken();
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setIsLoggedIn(false);
    setUsername(null);
  }, []);

  // Verify stored token on mount — runs once, does NOT block rendering
  useEffect(() => {
    const token = getStoredToken();
    const storedUser = typeof window !== "undefined" ? localStorage.getItem(USER_KEY) : null;
    if (!token) {
      setInitialized(true);
      return;
    }

    fetch(`${API_BASE}/api/verify`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => {
        if (res.ok) {
          setIsLoggedIn(true);
          setUsername(storedUser);
        } else {
          localStorage.removeItem(TOKEN_KEY);
          localStorage.removeItem(USER_KEY);
        }
      })
      .catch(() => {})
      .finally(() => setInitialized(true));
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    const res = await fetch(`${API_BASE}/api/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });
    if (!res.ok) {
      const text = await res.text().catch(() => "Login failed");
      throw new Error(text);
    }
    const data = await res.json();
    localStorage.setItem(TOKEN_KEY, data.token);
    localStorage.setItem(USER_KEY, data.username);
    setIsLoggedIn(true);
    setUsername(data.username);
  }, []);

  return (
    <AuthContext.Provider value={{ isLoggedIn, username, initialized, login, logout, getToken }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

// ---- API fetch with auth token ----

export async function authFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getStoredToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  });

  if (res.status === 401) {
    // Token expired — clean up and redirect
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
    throw new Error("Session expired");
  }

  if (!res.ok) {
    const text = await res.text().catch(() => "Unknown error");
    throw new Error(text || `HTTP ${res.status}`);
  }

  return res.json() as Promise<T>;
}
