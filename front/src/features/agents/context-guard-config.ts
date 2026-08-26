import { z, type RefinementCtx } from 'zod'
import type { AgentConfig, ContextGuardConfig } from '@/types/api'

export type ContextGuardMode = 'off' | 'threshold' | 'sliding_window'

export interface ContextGuardFormValues {
  mode: ContextGuardMode
  maxTokens: string
  maxTurns: string
}

export const EMPTY_CONTEXT_GUARD_FORM_VALUES: ContextGuardFormValues = {
  mode: 'off',
  maxTokens: '',
  maxTurns: '',
}

const MAX_INT32 = 2_147_483_647

function validateWholeNumber(value: string, path: string[], ctx: RefinementCtx) {
  const trimmed = value.trim()
  if (trimmed === '') return
  if (!/^\d+$/.test(trimmed) || Number(trimmed) > MAX_INT32) {
    ctx.addIssue({
      code: 'custom',
      path,
      message: 'Use a whole number from 0 to 2147483647.',
    })
  }
}

export const contextGuardFormSchema = z
  .object({
    mode: z.enum(['off', 'threshold', 'sliding_window']),
    maxTokens: z.string(),
    maxTurns: z.string(),
  })
  .superRefine((values, ctx) => {
    if (values.mode === 'threshold') {
      validateWholeNumber(values.maxTokens, ['maxTokens'], ctx)
    }
    if (values.mode === 'sliding_window') {
      validateWholeNumber(values.maxTurns, ['maxTurns'], ctx)
    }
  })

function inputValue(value: number | undefined): string {
  return value === undefined ? '' : String(value)
}

export function contextGuardFormValuesFromConfig(
  config?: ContextGuardConfig,
): ContextGuardFormValues {
  switch (config?.strategy) {
    case 'CONTEXT_GUARD_STRATEGY_THRESHOLD':
      return {
        mode: 'threshold',
        maxTokens: inputValue(config.max_tokens),
        maxTurns: '',
      }
    case 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW':
      return {
        mode: 'sliding_window',
        maxTokens: '',
        maxTurns: inputValue(config.max_turns),
      }
    default:
      return { ...EMPTY_CONTEXT_GUARD_FORM_VALUES }
  }
}

export function contextGuardValuesForMode(
  mode: ContextGuardMode,
  current: ContextGuardFormValues,
): ContextGuardFormValues {
  return {
    mode,
    maxTokens: mode === 'threshold' ? current.maxTokens : '',
    maxTurns: mode === 'sliding_window' ? current.maxTurns : '',
  }
}

function numberValue(value: string): number | undefined {
  const trimmed = value.trim()
  return trimmed === '' ? undefined : Number(trimmed)
}

export function buildContextGuardConfig(
  values: ContextGuardFormValues,
): AgentConfig['context_guard'] {
  switch (values.mode) {
    case 'threshold': {
      const maxTokens = numberValue(values.maxTokens)
      return {
        strategy: 'CONTEXT_GUARD_STRATEGY_THRESHOLD',
        ...(maxTokens === undefined ? {} : { max_tokens: maxTokens }),
      }
    }
    case 'sliding_window': {
      const maxTurns = numberValue(values.maxTurns)
      return {
        strategy: 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW',
        ...(maxTurns === undefined ? {} : { max_turns: maxTurns }),
      }
    }
    default:
      return undefined
  }
}

export function supportsContextGuard(type?: string): boolean {
  return type === 'AGENT_TYPE_LLM' || type === 'AGENT_TYPE_UNSPECIFIED'
}
