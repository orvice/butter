import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  PlaneAdminService,
  PlaneAppService,
  type PlaneApp,
  type GetPlaneSettingsResponse,
  type PlaneProviderProfile,
} from '@/gen/agents/v1/plane_pb'
import { makeClient } from './transport'

const appClient = makeClient(PlaneAppService)
const adminClient = makeClient(PlaneAdminService)

export const PLANE_APPS_KEY = ['plane-apps'] as const
const PLANE_SETTINGS_KEY = ['plane-settings'] as const
const PLANE_PROFILE_KEY = ['plane-provider-profile'] as const

/** The operator-editable fields of a Plane App. */
export interface PlaneAppInput {
  displayName: string
  /** Immutable after registration; the server preserves the stored value. */
  clientId: string
  agentId: string
  allowedUserIds: string[]
  /** Undefined follows the server default; 0 means unlimited. */
  maxRunSeconds?: number
}

export interface CreatePlaneAppInput extends PlaneAppInput {
  /** Write-only: encrypted at rest, never read back. */
  clientSecret?: string
  /** Write-only: encrypted at rest, never read back. */
  webhookSecret?: string
}

export interface UpdatePlaneAppInput extends PlaneAppInput {
  id: string
  /** Required. A stale revision is refused instead of overwriting. */
  revision: bigint
}

// --- Provider profile (platform-level, fixed by startup config) -------------

/**
 * The deployment's Plane provider profile. This is read-only by design: an
 * App cannot submit its own origin, so the browser never offers one either.
 */
export function usePlaneProviderProfile(enabled = true) {
  return useQuery({
    queryKey: PLANE_PROFILE_KEY,
    enabled,
    staleTime: 5 * 60 * 1000,
    queryFn: async (): Promise<PlaneProviderProfile> => {
      const res = await appClient.getPlaneProviderProfile({})
      if (!res.profile) throw new Error('provider profile unavailable')
      return res.profile
    },
  })
}

// --- Platform settings (global admin) ---------------------------------------

export function usePlaneSettings(enabled = true) {
  return useQuery({
    queryKey: PLANE_SETTINGS_KEY,
    enabled,
    queryFn: async (): Promise<GetPlaneSettingsResponse> => {
      const res = await adminClient.getPlaneSettings({})
      if (!res.settings) throw new Error('settings unavailable')
      return res
    },
  })
}

export function useUpdatePlaneSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (publicBaseUrl: string) => {
      const res = await adminClient.updatePlaneSettings({
        settings: { publicBaseUrl },
      })
      if (!res.settings) throw new Error('update returned no settings')
      return res.settings
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: PLANE_SETTINGS_KEY })
      // Derived callback and webhook URLs depend on this.
      qc.invalidateQueries({ queryKey: PLANE_APPS_KEY })
    },
  })
}

// --- Plane Apps -------------------------------------------------------------

export function usePlaneApps() {
  return useQuery({
    queryKey: PLANE_APPS_KEY,
    queryFn: async (): Promise<PlaneApp[]> => {
      const res = await appClient.listPlaneApps({})
      return res.apps
    },
  })
}

export function usePlaneApp(id: string | undefined) {
  return useQuery({
    queryKey: [...PLANE_APPS_KEY, id],
    enabled: Boolean(id),
    queryFn: async (): Promise<PlaneApp> => {
      const res = await appClient.getPlaneApp({ id: id! })
      if (!res.app) throw new Error('plane app not found')
      return res.app
    },
  })
}

export function useCreatePlaneApp() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreatePlaneAppInput) => {
      const { clientSecret, webhookSecret, ...app } = input
      const res = await appClient.createPlaneApp({
        app,
        clientSecret: clientSecret || undefined,
        webhookSecret: webhookSecret || undefined,
      })
      if (!res.app) throw new Error('create returned no app')
      return res.app
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: PLANE_APPS_KEY }),
  })
}

export function useUpdatePlaneApp() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdatePlaneAppInput) => {
      const res = await appClient.updatePlaneApp({ app: input })
      if (!res.app) throw new Error('update returned no app')
      return res.app
    },
    onSuccess: (app) => {
      qc.invalidateQueries({ queryKey: PLANE_APPS_KEY })
      qc.setQueryData([...PLANE_APPS_KEY, app.id], app)
    },
  })
}

export interface PutPlaneAppCredentialsInput {
  appId: string
  /** Undefined keeps the stored secret; '' clears it. */
  clientSecret?: string
  /** Undefined keeps the stored secret; '' clears it. */
  webhookSecret?: string
}

export function usePutPlaneAppCredentials() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: PutPlaneAppCredentialsInput) => {
      const res = await appClient.putPlaneAppCredentials(input)
      if (!res.app) throw new Error('credential update returned no app')
      return res.app
    },
    onSuccess: (app) => {
      qc.invalidateQueries({ queryKey: PLANE_APPS_KEY })
      qc.setQueryData([...PLANE_APPS_KEY, app.id], app)
    },
  })
}

export function useDeletePlaneApp() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await appClient.deletePlaneApp({ id })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: PLANE_APPS_KEY }),
  })
}
