import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  WorkspaceMemoryConfigService,
  WorkspaceMemoryScope,
  WorkspaceMemoryService,
  type ListWorkspaceMemoriesResponse,
  type PutWorkspaceMemoryConfigResponse,
  type TestWorkspaceMemoryConnectionResponse,
  type WorkspaceMemory,
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

// --- Memory management (issue #339) -----------------------------------------

const memoriesClient = makeClient(WorkspaceMemoryService)

const MEMORIES_KEY = ['workspace-memories'] as const

export type MemoryScopeSelection =
  { scope: 'workspace' } | { scope: 'agent'; agentId: string }

function scopeRequest(selection: MemoryScopeSelection) {
  return selection.scope === 'agent'
    ? { scope: WorkspaceMemoryScope.AGENT, agentId: selection.agentId }
    : { scope: WorkspaceMemoryScope.WORKSPACE, agentId: '' }
}

function selectionKey(selection: MemoryScopeSelection) {
  return selection.scope === 'agent'
    ? ['agent', selection.agentId]
    : ['workspace']
}

/**
 * The newest memories of one scope. The mem0 OSS listing has no
 * pagination: the response is capped (`limit`) and `truncated` says more
 * exist — search reaches them.
 */
export function useWorkspaceMemories(
  selection: MemoryScopeSelection,
  enabled = true
) {
  return useQuery({
    queryKey: [...MEMORIES_KEY, ...selectionKey(selection)],
    enabled:
      enabled && (selection.scope === 'workspace' || !!selection.agentId),
    retry: false,
    queryFn: async (): Promise<ListWorkspaceMemoriesResponse> =>
      memoriesClient.listWorkspaceMemories(scopeRequest(selection)),
  })
}

/** Semantic search within one scope; `query` empty disables the query. */
export function useSearchWorkspaceMemories(
  selection: MemoryScopeSelection,
  query: string
) {
  const trimmed = query.trim()
  return useQuery({
    queryKey: [...MEMORIES_KEY, ...selectionKey(selection), 'search', trimmed],
    enabled:
      trimmed !== '' &&
      (selection.scope === 'workspace' || !!selection.agentId),
    retry: false,
    queryFn: async (): Promise<WorkspaceMemory[]> => {
      const res = await memoriesClient.searchWorkspaceMemories({
        ...scopeRequest(selection),
        query: trimmed,
      })
      return res.memories
    },
  })
}

export function useDeleteWorkspaceMemory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (memoryId: string) => {
      await memoriesClient.deleteWorkspaceMemory({ memoryId })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: MEMORIES_KEY }),
  })
}
