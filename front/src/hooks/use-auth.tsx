import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import { TOKEN_KEY } from "@/lib/constants";
import { login as loginRequest, logout as logoutRequest, type AuthUser } from "@/api/auth";

interface AuthContextValue {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(
    () => localStorage.getItem(TOKEN_KEY),
  );
  const [user, setUser] = useState<AuthUser | null>(null);

  const login = useCallback(async (username: string, password: string): Promise<boolean> => {
    try {
      const res = await loginRequest(username, password);
      if (!res.token) return false;
      localStorage.setItem(TOKEN_KEY, res.token);
      setToken(res.token);
      setUser(res.user ?? null);
      return true;
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      setToken(null);
      setUser(null);
      return false;
    }
  }, []);

  const logout = useCallback(() => {
    void logoutRequest().catch(() => undefined);
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, user, isAuthenticated: !!token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be inside AuthProvider");
  return ctx;
}
