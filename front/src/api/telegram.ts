import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  TelegramAdminService,
  TelegramProcessingService,
  TelegramChannelService,
  TelegramDestinationService,
  type TelegramChannel,
  type TelegramChannelStatus,
  type TelegramDestination,
  type TelegramProcessingRecord,
  type TelegramProcessingStatus,
  type TelegramSettings,
} from '@/gen/agents/v1/telegram_pb'
import { makeClient } from './transport'

const channelClient = makeClient(TelegramChannelService)
const destinationClient = makeClient(TelegramDestinationService)
const adminClient = makeClient(TelegramAdminService)
const processingClient = makeClient(TelegramProcessingService)

const CHANNELS_KEY = ['telegram-channels'] as const
const DESTINATIONS_KEY = ['telegram-destinations'] as const
const SETTINGS_KEY = ['telegram-settings'] as const

// --- Platform settings (global admin) ---------------------------------------

export function useTelegramSettings(enabled = true) {
  return useQuery({
    queryKey: SETTINGS_KEY,
    enabled,
    queryFn: async (): Promise<TelegramSettings> => {
      const res = await adminClient.getTelegramSettings({})
      if (!res.settings) throw new Error('settings unavailable')
      return res.settings
    },
  })
}

export function useUpdateTelegramSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (webhookBaseUrl: string) => {
      const res = await adminClient.updateTelegramSettings({
        settings: { webhookBaseUrl },
      })
      if (!res.settings) throw new Error('update returned no settings')
      return res.settings
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: SETTINGS_KEY })
      // The derived callback URL on every channel status depends on this.
      qc.invalidateQueries({ queryKey: CHANNELS_KEY })
    },
  })
}

// --- Channels ---------------------------------------------------------------

export function useTelegramChannels() {
  return useQuery({
    queryKey: CHANNELS_KEY,
    queryFn: async (): Promise<TelegramChannel[]> => {
      const res = await channelClient.listTelegramChannels({})
      return res.channels
    },
  })
}

export function useTelegramChannel(id: string | undefined) {
  return useQuery({
    queryKey: [...CHANNELS_KEY, id],
    enabled: Boolean(id),
    queryFn: async (): Promise<TelegramChannel> => {
      const res = await channelClient.getTelegramChannel({ id: id! })
      if (!res.channel) throw new Error('channel not found')
      return res.channel
    },
  })
}

export function useTelegramChannelStatus(id: string | undefined) {
  return useQuery({
    queryKey: [...CHANNELS_KEY, id, 'status'],
    enabled: Boolean(id),
    queryFn: async (): Promise<TelegramChannelStatus> => {
      const res = await channelClient.getTelegramChannelStatus({ channelId: id! })
      if (!res.status) throw new Error('status unavailable')
      return res.status
    },
  })
}

export interface CreateTelegramChannelInput {
  key: string
  name?: string
  receiveMode?: TelegramChannel['receiveMode']
  /** Write-only: validated with getMe, encrypted at rest, never read back. */
  botToken: string
}

export function useCreateTelegramChannel() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateTelegramChannelInput) => {
      const res = await channelClient.createTelegramChannel({
        channel: {
          key: input.key,
          name: input.name ?? '',
          receiveMode: input.receiveMode,
        },
        botToken: input.botToken,
      })
      if (!res.channel) throw new Error('create returned no channel')
      return res.channel
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: CHANNELS_KEY }),
  })
}

export interface UpdateTelegramChannelInput {
  id: string
  revision: bigint
  name: string
  receiveMode: TelegramChannel['receiveMode']
}

export function useUpdateTelegramChannel() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateTelegramChannelInput) => {
      const res = await channelClient.updateTelegramChannel({
        channel: {
          id: input.id,
          revision: input.revision,
          name: input.name,
          receiveMode: input.receiveMode,
        },
      })
      if (!res.channel) throw new Error('update returned no channel')
      return res.channel
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: CHANNELS_KEY }),
  })
}

export function useRotateTelegramChannelCredential() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { channelId: string; botToken: string }) => {
      const res = await channelClient.rotateTelegramChannelCredential(input)
      if (!res.channel) throw new Error('rotation returned no channel')
      return res.channel
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: CHANNELS_KEY }),
  })
}

export function useSetTelegramChannelEnabled() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      channelId: string
      revision: bigint
      inboundEnabled: boolean
      outboundEnabled: boolean
    }) => {
      const res = await channelClient.setTelegramChannelEnabled(input)
      return { channel: res.channel, warnings: res.warnings }
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: CHANNELS_KEY }),
  })
}

export function useDeleteTelegramChannel() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await channelClient.deleteTelegramChannel({ id })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: CHANNELS_KEY }),
  })
}

// --- Destinations -----------------------------------------------------------

export function useTelegramDestinations(channelId?: string) {
  return useQuery({
    queryKey: [...DESTINATIONS_KEY, channelId ?? 'all'],
    queryFn: async (): Promise<TelegramDestination[]> => {
      const res = await destinationClient.listTelegramDestinations({
        channelId: channelId ?? '',
      })
      return res.destinations
    },
  })
}

export function useTelegramDestination(id: string | undefined) {
  return useQuery({
    queryKey: [...DESTINATIONS_KEY, 'detail', id],
    enabled: Boolean(id),
    queryFn: async (): Promise<TelegramDestination> => {
      const res = await destinationClient.getTelegramDestination({ id: id! })
      if (!res.destination) throw new Error('destination not found')
      return res.destination
    },
  })
}

export function useCreateTelegramDestination() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (destination: Partial<TelegramDestination>) => {
      const res = await destinationClient.createTelegramDestination({
        destination: destination as TelegramDestination,
      })
      if (!res.destination) throw new Error('create returned no destination')
      return res.destination
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: DESTINATIONS_KEY }),
  })
}

export function useUpdateTelegramDestination() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (destination: Partial<TelegramDestination>) => {
      const res = await destinationClient.updateTelegramDestination({
        destination: destination as TelegramDestination,
      })
      if (!res.destination) throw new Error('update returned no destination')
      return res.destination
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: DESTINATIONS_KEY }),
  })
}

export function useSendTelegramTestMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { destinationId: string; text?: string }) => {
      const res = await destinationClient.sendTelegramTestMessage({
        destinationId: input.destinationId,
        text: input.text ?? '',
      })
      return res
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: DESTINATIONS_KEY }),
  })
}

export function useDeleteTelegramDestination() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await destinationClient.deleteTelegramDestination({ id })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: DESTINATIONS_KEY }),
  })
}


// --- Processing records -----------------------------------------------------

const PROCESSING_KEY = ['telegram-processing'] as const

export function useTelegramProcessingRecords(filter: {
  channelId?: string
  destinationId?: string
  status?: TelegramProcessingStatus
} = {}) {
  return useQuery({
    queryKey: [...PROCESSING_KEY, filter],
    queryFn: async (): Promise<TelegramProcessingRecord[]> => {
      const res = await processingClient.listTelegramProcessingRecords({
        channelId: filter.channelId ?? '',
        destinationId: filter.destinationId ?? '',
        status: filter.status,
      })
      return res.records
    },
  })
}

export function useTelegramProcessingRecord(id: string | undefined) {
  return useQuery({
    queryKey: [...PROCESSING_KEY, 'detail', id],
    enabled: Boolean(id),
    queryFn: async (): Promise<TelegramProcessingRecord> => {
      const res = await processingClient.getTelegramProcessingRecord({ id: id! })
      if (!res.record) throw new Error('record not found')
      return res.record
    },
  })
}

/**
 * Resends the segments of an already-produced reply that never landed. There
 * is deliberately no rerun action: once agent work may have had side effects,
 * repeating it is an operator decision the dashboard does not make.
 */
export function useResendTelegramReply() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await processingClient.resendTelegramReply({ id })
      return res
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: PROCESSING_KEY }),
  })
}
