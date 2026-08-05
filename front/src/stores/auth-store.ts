import { create } from 'zustand'
import type { Workspace } from '@/gen/agents/v1/workspace_pb'
import {
  isAdmin as checkAdmin,
  login as loginRequest,
  logout as logoutRequest,
  me as meRequest,
  type AuthUser,
  type LoginResponse,
} from '@/api/auth'
import { TOKEN_KEY } from '@/lib/constants'

interface AuthState {
  auth: {
    token: string | null
    user: AuthUser | null
    loginWorkspaces: Workspace[]
    /** Session restore (token present, user not yet fetched) in flight. */
    restoring: boolean
    setUser: (user: AuthUser | null) => void
    applyLoginResponse: (res: LoginResponse) => void
    login: (username: string, password: string) => Promise<boolean>
    /** Fetch `Me` when a stored token exists but the user is unknown. */
    restore: () => Promise<void>
    logout: () => void
    reset: () => void
  }
}

export const useAuthStore = create<AuthState>()((set, get) => ({
  auth: {
    token: localStorage.getItem(TOKEN_KEY),
    user: null,
    loginWorkspaces: [],
    restoring: false,
    setUser: (user) => set((state) => ({ auth: { ...state.auth, user } })),
    applyLoginResponse: (res) => {
      if (!res.token) return
      localStorage.setItem(TOKEN_KEY, res.token)
      set((state) => ({
        auth: {
          ...state.auth,
          token: res.token,
          user: res.user ?? null,
          loginWorkspaces: res.workspaces ?? [],
        },
      }))
    },
    login: async (username, password) => {
      try {
        const res = await loginRequest(username, password)
        if (!res.token) return false
        get().auth.applyLoginResponse(res)
        return true
      } catch {
        get().auth.reset()
        return false
      }
    },
    restore: async () => {
      const { token, user, restoring } = get().auth
      if (!token || user || restoring) return
      set((state) => ({ auth: { ...state.auth, restoring: true } }))
      try {
        const res = await meRequest()
        if (!res.user) {
          get().auth.reset()
          return
        }
        set((state) => ({
          auth: { ...state.auth, user: res.user ?? null },
        }))
      } catch {
        get().auth.reset()
      } finally {
        set((state) => ({ auth: { ...state.auth, restoring: false } }))
      }
    },
    logout: () => {
      void logoutRequest().catch(() => undefined)
      get().auth.reset()
    },
    reset: () => {
      localStorage.removeItem(TOKEN_KEY)
      set((state) => ({
        auth: { ...state.auth, token: null, user: null, loginWorkspaces: [] },
      }))
    },
  },
}))

export function useIsAdmin() {
  return useAuthStore((state) => checkAdmin(state.auth.user))
}
