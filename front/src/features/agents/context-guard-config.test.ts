import { describe, expect, it } from 'vitest'
import {
  buildContextGuardConfig,
  contextGuardFormSchema,
  contextGuardFormValuesFromConfig,
  contextGuardValuesForMode,
  EMPTY_CONTEXT_GUARD_FORM_VALUES,
} from './context-guard-config'

describe('context guard form mapping', () => {
  it('leaves ContextGuard off without serializing a config', () => {
    expect(buildContextGuardConfig({
      mode: 'off',
      maxTokens: '32000',
      maxTurns: '10',
    })).toBeUndefined()
  })

  it('round-trips a threshold Agent Context Override', () => {
    const values = contextGuardFormValuesFromConfig({
      strategy: 'CONTEXT_GUARD_STRATEGY_THRESHOLD',
      max_tokens: 32000,
    })

    expect(values).toEqual({ mode: 'threshold', maxTokens: '32000', maxTurns: '' })
    expect(buildContextGuardConfig(values)).toEqual({
      strategy: 'CONTEXT_GUARD_STRATEGY_THRESHOLD',
      max_tokens: 32000,
    })
  })

  it('round-trips a sliding-window content-entry limit', () => {
    const values = contextGuardFormValuesFromConfig({
      strategy: 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW',
      max_turns: 7,
    })

    expect(values).toEqual({ mode: 'sliding_window', maxTokens: '', maxTurns: '7' })
    expect(buildContextGuardConfig(values)).toEqual({
      strategy: 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW',
      max_turns: 7,
    })
  })

  it('preserves explicit zero values with their documented default semantics', () => {
    expect(contextGuardFormValuesFromConfig({
      strategy: 'CONTEXT_GUARD_STRATEGY_THRESHOLD',
      max_tokens: 0,
    })).toEqual({ mode: 'threshold', maxTokens: '0', maxTurns: '' })
    expect(contextGuardFormValuesFromConfig({
      strategy: 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW',
      max_turns: 0,
    })).toEqual({ mode: 'sliding_window', maxTokens: '', maxTurns: '0' })
  })

  it('clears fields that do not apply to the selected strategy', () => {
    const threshold = contextGuardValuesForMode('threshold', {
      mode: 'sliding_window',
      maxTokens: '32000',
      maxTurns: '7',
    })
    expect(threshold).toEqual({ mode: 'threshold', maxTokens: '32000', maxTurns: '' })

    expect(contextGuardValuesForMode('sliding_window', threshold)).toEqual({
      mode: 'sliding_window',
      maxTokens: '',
      maxTurns: '',
    })
    expect(contextGuardValuesForMode('off', threshold)).toEqual(EMPTY_CONTEXT_GUARD_FORM_VALUES)
  })
})

describe('contextGuardFormSchema', () => {
  it('accepts blank optional values and zero', () => {
    expect(contextGuardFormSchema.safeParse({
      mode: 'threshold',
      maxTokens: '',
      maxTurns: '',
    }).success).toBe(true)
    expect(contextGuardFormSchema.safeParse({
      mode: 'sliding_window',
      maxTokens: '',
      maxTurns: '0',
    }).success).toBe(true)
  })

  it('rejects negative, fractional, and out-of-range numeric values', () => {
    for (const value of ['-1', '1.5', '2147483648']) {
      const result = contextGuardFormSchema.safeParse({
        mode: 'threshold',
        maxTokens: value,
        maxTurns: '',
      })
      expect(result.success, `threshold maxTokens ${value}`).toBe(false)
    }

    const result = contextGuardFormSchema.safeParse({
      mode: 'sliding_window',
      maxTokens: '',
      maxTurns: '-1',
    })
    expect(result.success).toBe(false)
  })
})
