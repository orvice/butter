import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import { TOKEN_KEY } from "@/lib/constants";
import { validateToken } from "@/api/client";

interface AuthContextValue {
  token: string | null;
  isAuthenticated: boolean;
  login: (token: string) => Promise<boolean>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(
    () => localStorage.getItem(TOKEN_KEY),
  );

  const login = useCallback(async (newToken: string): Promise<boolean> => {
    localStorage.setItem(TOKEN_KEY, newToken);
    const valid = await validateToken(newToken);
    if (valid) {
      setToken(newToken);
      return true;
    }
    localStorage.removeItem(TOKEN_KEY);
    return false;
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, isAuthenticated: !!token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be inside AuthProvider");
  return ctx;
}
