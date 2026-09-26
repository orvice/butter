import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  WorkspaceMemoryConfigService,
  type PutWorkspaceMemoryConfigResponse,
  type TestWorkspaceMemoryConnectionResponse,
  type WorkspaceMemoryConfig,
} from '@/gen/agents/v1/workspace_memory_pb'
import { makeClient } from './transport'

const client = makeClient(WorkspaceMemoryConfigService)

const MEMORY_CONFIG_KEY = ['workspace-memory-config'] as const

/**
 * The workspace's mem0 connection (ADR-0013). Resolves to `null` when the
 * workspace has none — memory-enabled agents then run without memory.
 */
export function useWorkspaceMemoryConfig() {
  return useQuery({
    queryKey: MEMORY_CONFIG_KEY,
    queryFn: async (): Promise<WorkspaceMemoryConfig | null> => {
      const res = await client.getWorkspaceMemoryConfig({})
      return res.config ?? null
    },
  })
}

export interface PutWorkspaceMemoryConfigInput {
  baseUrl: string
  enabled: boolean
  /**
   * Write-only mem0 API key: `undefined` keeps the stored key, `''` clears
   * it, any other value sets or rotates it.
   */
  apiKey?: string
}

/**
 * Creates or replaces the config. An enabled save is probed first: a key
 * the server rejects fails the mutation, while an unreachable server saves
 * and returns a `warning`.
 */
export function usePutWorkspaceMemoryConfig() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (
      input: PutWorkspaceMemoryConfigInput
    ): Promise<PutWorkspaceMemoryConfigResponse> => {
      return client.putWorkspaceMemoryConfig({
        baseUrl: input.baseUrl,
        enabled: input.enabled,
        apiKey: input.apiKey,
      })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: MEMORY_CONFIG_KEY }),
  })
}

export function useDeleteWorkspaceMemoryConfig() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      await client.deleteWorkspaceMemoryConfig({})
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: MEMORY_CONFIG_KEY }),
  })
}

/** Probes the stored config's server with the stored key. */
export function useTestWorkspaceMemoryConnection() {
  return useMutation({
    mutationFn: async (): Promise<TestWorkspaceMemoryConnectionResponse> => {
      return client.testWorkspaceMemoryConnection({})
    },
  })
}
