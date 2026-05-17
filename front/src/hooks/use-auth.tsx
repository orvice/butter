/* eslint-disable react-refresh/only-export-components */

import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from "react";
import type { Workspace } from "@/gen/agents/v1/workspace_pb";
import { TOKEN_KEY } from "@/lib/constants";
import { login as loginRequest, logout as logoutRequest, me as meRequest, type AuthUser } from "@/api/auth";

interface AuthContextValue {
  token: string | null;
  user: AuthUser | null;
  loginWorkspaces: Workspace[];
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY));
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loginWorkspaces, setLoginWorkspaces] = useState<Workspace[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(() => !!localStorage.getItem(TOKEN_KEY));

  useEffect(() => {
    let cancelled = false;

    if (!token) {
      setUser(null);
      setLoginWorkspaces([]);
      setIsLoading(false);
      return;
    }

    if (user?.id) {
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    void meRequest()
      .then((res) => {
        if (cancelled) return;
        if (!res.user) {
          localStorage.removeItem(TOKEN_KEY);
          setToken(null);
          setUser(null);
          setLoginWorkspaces([]);
          return;
        }
        setUser(res.user);
      })
      .catch(() => {
        if (cancelled) return;
        localStorage.removeItem(TOKEN_KEY);
        setToken(null);
        setUser(null);
        setLoginWorkspaces([]);
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [token, user?.id]);

  const login = useCallback(async (username: string, password: string): Promise<boolean> => {
    try {
      setIsLoading(true);
      const res = await loginRequest(username, password);
      if (!res.token) return false;
      localStorage.setItem(TOKEN_KEY, res.token);
      setToken(res.token);
      setUser(res.user ?? null);
      setLoginWorkspaces(res.workspaces ?? []);
      return true;
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      setToken(null);
      setUser(null);
      setLoginWorkspaces([]);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    void logoutRequest().catch(() => undefined);
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
    setLoginWorkspaces([]);
    setIsLoading(false);
  }, []);

  return (
    <AuthContext.Provider value={{ token, user, loginWorkspaces, isAuthenticated: !!token, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be inside AuthProvider");
  return ctx;
}
