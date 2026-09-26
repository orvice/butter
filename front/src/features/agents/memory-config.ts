import { z } from 'zod'
import type { AgentConfig, MemoryConfig } from '@/types/api'

/** Mirrors internalagent.MaxMemoryTopK. */
export const MAX_MEMORY_TOP_K = 50

export interface MemoryFormValues {
  enabled: boolean
  autoRecall: boolean
  autoCapture: boolean
  tools: boolean
  agentScopeWrite: boolean
  topK: string
  threshold: string
}

export const EMPTY_MEMORY_FORM_VALUES: MemoryFormValues = {
  enabled: false,
  autoRecall: true,
  autoCapture: true,
  tools: false,
  agentScopeWrite: false,
  topK: '',
  threshold: '',
}

export const memoryFormSchema = z
  .object({
    enabled: z.boolean(),
    autoRecall: z.boolean(),
    autoCapture: z.boolean(),
    tools: z.boolean(),
    agentScopeWrite: z.boolean(),
    topK: z.string(),
    threshold: z.string(),
  })
  .superRefine((values, ctx) => {
    if (!values.enabled) return
    const topK = values.topK.trim()
    if (
      topK !== '' &&
      (!/^\d+$/.test(topK) ||
        Number(topK) < 1 ||
        Number(topK) > MAX_MEMORY_TOP_K)
    ) {
      ctx.addIssue({
        code: 'custom',
        path: ['topK'],
        message: `Use a whole number from 1 to ${MAX_MEMORY_TOP_K}.`,
      })
    }
    const threshold = values.threshold.trim()
    if (threshold !== '') {
      const n = Number(threshold)
      if (!Number.isFinite(n) || n < 0 || n > 1) {
        ctx.addIssue({
          code: 'custom',
          path: ['threshold'],
          message: 'Use a number from 0 to 1.',
        })
      }
    }
  })

export function memoryFormValuesFromConfig(
  config?: MemoryConfig
): MemoryFormValues {
  if (!config) return { ...EMPTY_MEMORY_FORM_VALUES }
  return {
    enabled: config.enabled ?? false,
    autoRecall: !config.disable_auto_recall,
    autoCapture: !config.disable_auto_capture,
    tools: config.enable_tools ?? false,
    agentScopeWrite: config.allow_agent_scope_write ?? false,
    topK: config.top_k === undefined ? '' : String(config.top_k),
    threshold: config.threshold === undefined ? '' : String(config.threshold),
  }
}

/**
 * Serializes the form. Disabled memory serializes to no config at all, and
 * the tool switches are dropped for types that cannot mount tools.
 */
export function buildMemoryConfig(
  values: MemoryFormValues,
  type?: string
): AgentConfig['memory'] {
  if (!values.enabled || !supportsMemory(type)) return undefined
  const tools = supportsMemoryTools(type) && values.tools
  const topK = values.topK.trim()
  const threshold = values.threshold.trim()
  return {
    enabled: true,
    ...(values.autoRecall ? {} : { disable_auto_recall: true }),
    ...(values.autoCapture ? {} : { disable_auto_capture: true }),
    ...(tools ? { enable_tools: true } : {}),
    ...(tools && values.agentScopeWrite
      ? { allow_agent_scope_write: true }
      : {}),
    ...(topK === '' ? {} : { top_k: Number(topK) }),
    ...(threshold === '' ? {} : { threshold: Number(threshold) }),
  }
}

/**
 * PI and CURSOR agents keep their behavior on the ButterBox and reject a
 * memory config on write.
 */
export function supportsMemory(type?: string): boolean {
  return type !== 'AGENT_TYPE_PI' && type !== 'AGENT_TYPE_CURSOR'
}

/** Memory tools are only mounted on an LLM root agent (ADR-0013 §7). */
export function supportsMemoryTools(type?: string): boolean {
  return (
    type === 'AGENT_TYPE_LLM' ||
    type === 'AGENT_TYPE_UNSPECIFIED' ||
    type === undefined
  )
}
