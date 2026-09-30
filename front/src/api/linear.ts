import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  LinearAdminService,
  LinearAppService,
  LinearProcessingService,
  type LinearApp,
  type LinearProcessingRecord,
  type LinearProcessingStatus,
  type LinearInstallation,
  type LinearSettings,
} from '@/gen/agents/v1/linear_pb'
import { makeClient } from './transport'

const appClient = makeClient(LinearAppService)
const adminClient = makeClient(LinearAdminService)
const processingClient = makeClient(LinearProcessingService)

export const LINEAR_APPS_KEY = ['linear-apps'] as const
const SETTINGS_KEY = ['linear-settings'] as const
const PROCESSING_KEY = ['linear-processing'] as const

// --- Platform settings (global admin) ---------------------------------------

export function useLinearSettings(enabled = true) {
  return useQuery({
    queryKey: SETTINGS_KEY,
    enabled,
    queryFn: async (): Promise<LinearSettings> => {
      const res = await adminClient.getLinearSettings({})
      if (!res.settings) throw new Error('settings unavailable')
      return res.settings
    },
  })
}

export function useUpdateLinearSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (publicBaseUrl: string) => {
      const res = await adminClient.updateLinearSettings({
        settings: { publicBaseUrl },
      })
      if (!res.settings) throw new Error('update returned no settings')
      return res.settings
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: SETTINGS_KEY })
      // Every App's derived callback and webhook URLs depend on this.
      qc.invalidateQueries({ queryKey: LINEAR_APPS_KEY })
    },
  })
}

// --- Linear Apps --------------------------------------------------------------

export function useLinearApps() {
  return useQuery({
    queryKey: LINEAR_APPS_KEY,
    queryFn: async (): Promise<LinearApp[]> => {
      const res = await appClient.listLinearApps({})
      return res.apps
    },
  })
}

export function useLinearApp(id: string | undefined) {
  return useQuery({
    queryKey: [...LINEAR_APPS_KEY, id],
    enabled: Boolean(id),
    queryFn: async (): Promise<LinearApp> => {
      const res = await appClient.getLinearApp({ id: id! })
      if (!res.app) throw new Error('linear app not found')
      return res.app
    },
  })
}

/** The operator-editable fields of a Linear App. */
export interface LinearAppInput {
  displayName: string
  clientId: string
  agentId: string
  inboundEnabled: boolean
  allowedUserIds: string[]
  /** Undefined keeps the default (1800s); 0 means unlimited. */
  maxRunSeconds?: number
}

export interface CreateLinearAppInput extends LinearAppInput {
  /** Write-only: encrypted at rest, never read back. */
  clientSecret?: string
  /** Write-only: encrypted at rest, never read back. */
  webhookSecret?: string
}

export function useCreateLinearApp() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateLinearAppInput) => {
      const { clientSecret, webhookSecret, ...app } = input
      const res = await appClient.createLinearApp({
        app,
        clientSecret: clientSecret || undefined,
        webhookSecret: webhookSecret || undefined,
      })
      if (!res.app) throw new Error('create returned no app')
      return res.app
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: LINEAR_APPS_KEY }),
  })
}

export function useUpdateLinearApp() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (
      input: LinearAppInput & { id: string; revision: bigint }
    ) => {
      const res = await appClient.updateLinearApp({ app: input })
      if (!res.app) throw new Error('update returned no app')
      return res.app
    },
    onSuccess: (app) => {
      qc.invalidateQueries({ queryKey: LINEAR_APPS_KEY })
      qc.setQueryData([...LINEAR_APPS_KEY, app.id], app)
    },
  })
}

export interface PutLinearAppCredentialsInput {
  appId: string
  /** Undefined keeps the stored secret; '' clears it. */
  clientSecret?: string
  /** Undefined keeps the stored secret; '' clears it. */
  webhookSecret?: string
}

export function usePutLinearAppCredentials() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: PutLinearAppCredentialsInput) => {
      const res = await appClient.putLinearAppCredentials(input)
      if (!res.app) throw new Error('credential update returned no app')
      return res.app
    },
    onSuccess: (app) => {
      qc.invalidateQueries({ queryKey: LINEAR_APPS_KEY })
      qc.setQueryData([...LINEAR_APPS_KEY, app.id], app)
    },
  })
}

export function useDeleteLinearApp() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await appClient.deleteLinearApp({ id })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: LINEAR_APPS_KEY }),
  })
}

// --- Installations ------------------------------------------------------------

export function useLinearInstallations(appId: string | undefined) {
  return useQuery({
    queryKey: [...LINEAR_APPS_KEY, appId, 'installations'],
    enabled: Boolean(appId),
    queryFn: async (): Promise<LinearInstallation[]> => {
      const res = await appClient.listLinearInstallations({ appId: appId! })
      return res.installations
    },
  })
}

/** Starts an install; the caller sends the browser to the returned URL. */
export function useBeginLinearInstall() {
  return useMutation({
    mutationFn: async (input: { appId: string; returnUrl: string }) => {
      const res = await appClient.beginLinearInstall(input)
      return res.authorizeUrl
    },
  })
}

export function useDeleteLinearInstallation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { appId: string; id: string }) => {
      await appClient.deleteLinearInstallation(input)
    },
    onSuccess: (_data, input) =>
      qc.invalidateQueries({
        queryKey: [...LINEAR_APPS_KEY, input.appId, 'installations'],
      }),
  })
}

// --- Processing records -------------------------------------------------------

export function useLinearProcessingRecords(
  filter: { appId?: string; status?: LinearProcessingStatus } = {}
) {
  return useQuery({
    queryKey: [...PROCESSING_KEY, filter],
    queryFn: async (): Promise<LinearProcessingRecord[]> => {
      const res = await processingClient.listLinearProcessingRecords({
        appId: filter.appId ?? '',
        status: filter.status,
      })
      return res.records
    },
  })
}

export function useResendLinearReply() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await processingClient.resendLinearReply({ id })
      return res.record
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: PROCESSING_KEY }),
  })
}
