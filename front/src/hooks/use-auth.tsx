/* eslint-disable react-refresh/only-export-components */

import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from "react";
import type { Workspace } from "@/gen/agents/v1/workspace_pb";
import { TOKEN_KEY } from "@/lib/constants";
import {
  isAdmin as checkAdmin,
  login as loginRequest,
  logout as logoutRequest,
  me as meRequest,
  type AuthUser,
  type LoginResponse,
} from "@/api/auth";

interface AuthContextValue {
  token: string | null;
  user: AuthUser | null;
  loginWorkspaces: Workspace[];
  isAuthenticated: boolean;
  isLoading: boolean;
  isAdmin: boolean;
  login: (username: string, password: string) => Promise<boolean>;
  applyLoginResponse: (response: LoginResponse) => void;
  logout: () => void;
  refreshUser: (user: AuthUser) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY));
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loginWorkspaces, setLoginWorkspaces] = useState<Workspace[]>([]);
  const [loginPending, setLoginPending] = useState(false);
  // Loading means restoring a session (token present, user not yet fetched)
  // or an explicit login() in flight; derived so the effect below never has
  // to sync it back into state.
  const isLoading = loginPending || (!!token && !user);

  useEffect(() => {
    let cancelled = false;

    // Every path that clears the token (logout, login failure, meRequest
    // failure) already clears user/workspaces, so a missing token — or an
    // already-restored user — needs no state sync here.
    if (!token || user?.id) return;

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
      });

    return () => {
      cancelled = true;
    };
  }, [token, user?.id]);

  const applyLoginResponse = useCallback((res: LoginResponse) => {
    if (!res.token) return;
    localStorage.setItem(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user ?? null);
    setLoginWorkspaces(res.workspaces ?? []);
  }, []);

  const login = useCallback(async (username: string, password: string): Promise<boolean> => {
    try {
      setLoginPending(true);
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
      setLoginPending(false);
    }
  }, []);

  const logout = useCallback(() => {
    void logoutRequest().catch(() => undefined);
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
    setLoginWorkspaces([]);
  }, []);

  const refreshUser = useCallback((next: AuthUser) => {
    setUser(next);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        token,
        user,
        loginWorkspaces,
        isAuthenticated: !!token,
        isLoading,
        isAdmin: checkAdmin(user),
        login,
        applyLoginResponse,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be inside AuthProvider");
  return ctx;
}
