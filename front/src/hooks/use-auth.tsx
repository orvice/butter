import {
  isAdmin as checkAdmin,
  type AuthUser,
  type LoginResponse,
} from '@/api/auth'
import { useAuthStore } from '@/stores/auth-store'

/**
 * Compatibility hook for pages migrated from the pre-shadcn-admin dashboard.
 * Presents the old AuthProvider context surface on top of the Zustand store.
 */
export function useAuth() {
  const auth = useAuthStore((state) => state.auth)
  return {
    token: auth.token,
    user: auth.user,
    loginWorkspaces: auth.loginWorkspaces,
    isAuthenticated: !!auth.token,
    isLoading: auth.restoring || (!!auth.token && !auth.user),
    isAdmin: checkAdmin(auth.user),
    login: auth.login,
    applyLoginResponse: (res: LoginResponse) => auth.applyLoginResponse(res),
    logout: auth.logout,
    refreshUser: (user: AuthUser) => auth.setUser(user),
  }
}
