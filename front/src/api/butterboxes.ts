import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ButterBoxService,
  type ButterBox,
  type ButterBoxModel,
  type GetButterBoxStatusResponse,
} from '@/gen/agents/v1/butterbox_pb'
import { makeClient } from './transport'

const client = makeClient(ButterBoxService)

const BOXES_KEY = ['butterboxes'] as const

export function useButterBoxes() {
  return useQuery({
    queryKey: BOXES_KEY,
    queryFn: async (): Promise<ButterBox[]> => {
      const res = await client.listButterBoxes({})
      return res.butterBoxes
    },
  })
}

export function useButterBox(id: string | undefined) {
  return useQuery({
    queryKey: [...BOXES_KEY, id],
    enabled: Boolean(id),
    queryFn: async (): Promise<ButterBox> => {
      const res = await client.getButterBox({ id: id! })
      if (!res.butterBox) throw new Error('butterbox not found')
      return res.butterBox
    },
  })
}

export function useButterBoxStatus(id: string | undefined) {
  return useQuery({
    queryKey: [...BOXES_KEY, id, 'status'],
    enabled: Boolean(id),
    refetchInterval: 30_000,
    queryFn: async (): Promise<GetButterBoxStatusResponse> => {
      return client.getButterBoxStatus({ id: id! })
    },
  })
}

/**
 * The box's pi model catalog. Fetched on demand (pass `enabled: false`
 * until a consumer — e.g. the agent form — actually needs it).
 */
export function useButterBoxModels(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: [...BOXES_KEY, id, 'models'],
    enabled: Boolean(id) && enabled,
    queryFn: async (): Promise<ButterBoxModel[]> => {
      const res = await client.listButterBoxModels({ id: id! })
      return res.models
    },
  })
}

export interface CreateButterBoxInput {
  name: string
  baseUrl: string
  enabled: boolean
  /** Optional write-only access token: encrypted at rest, never read back. */
  token?: string
}

export function useCreateButterBox() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateButterBoxInput) => {
      const res = await client.createButterBox({
        name: input.name,
        baseUrl: input.baseUrl,
        enabled: input.enabled,
        token: input.token ?? '',
      })
      if (!res.butterBox) throw new Error('create returned no butterbox')
      return res.butterBox
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: BOXES_KEY }),
  })
}

export interface UpdateButterBoxInput {
  id: string
  name: string
  baseUrl: string
  enabled: boolean
}

export function useUpdateButterBox() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: UpdateButterBoxInput) => {
      const res = await client.updateButterBox(input)
      if (!res.butterBox) throw new Error('update returned no butterbox')
      return res.butterBox
    },
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: BOXES_KEY })
      qc.invalidateQueries({ queryKey: [...BOXES_KEY, input.id] })
    },
  })
}

export function useDeleteButterBox() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await client.deleteButterBox({ id })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: BOXES_KEY }),
  })
}

export function useSetButterBoxToken() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { id: string; token: string }) => {
      const res = await client.setButterBoxToken(input)
      if (!res.butterBox) throw new Error('token update returned no butterbox')
      return res.butterBox
    },
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: BOXES_KEY })
      qc.invalidateQueries({ queryKey: [...BOXES_KEY, input.id] })
    },
  })
}
