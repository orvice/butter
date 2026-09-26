import { describe, expect, it } from 'vitest'
import {
  buildMemoryConfig,
  EMPTY_MEMORY_FORM_VALUES,
  memoryFormSchema,
  memoryFormValuesFromConfig,
  supportsMemory,
  supportsMemoryTools,
  type MemoryFormValues,
} from './memory-config'

function values(overrides: Partial<MemoryFormValues>): MemoryFormValues {
  return { ...EMPTY_MEMORY_FORM_VALUES, ...overrides }
}

describe('memory form mapping', () => {
  it('serializes disabled memory as no config', () => {
    expect(
      buildMemoryConfig(
        values({ enabled: false, tools: true }),
        'AGENT_TYPE_LLM'
      )
    ).toBeUndefined()
  })

  it('serializes the defaults as just enabled', () => {
    expect(
      buildMemoryConfig(values({ enabled: true }), 'AGENT_TYPE_LLM')
    ).toEqual({ enabled: true })
  })

  it('maps switches onto the zero-value-default proto fields', () => {
    expect(
      buildMemoryConfig(
        values({
          enabled: true,
          autoRecall: false,
          autoCapture: false,
          tools: true,
          agentScopeWrite: true,
          topK: ' 8 ',
          threshold: '0.45',
        }),
        'AGENT_TYPE_LLM'
      )
    ).toEqual({
      enabled: true,
      disable_auto_recall: true,
      disable_auto_capture: true,
      enable_tools: true,
      allow_agent_scope_write: true,
      top_k: 8,
      threshold: 0.45,
    })
  })

  it('drops tool switches for composite agents and all memory for box agents', () => {
    const v = values({ enabled: true, tools: true, agentScopeWrite: true })
    expect(buildMemoryConfig(v, 'AGENT_TYPE_SEQUENTIAL')).toEqual({
      enabled: true,
    })
    expect(buildMemoryConfig(v, 'AGENT_TYPE_PI')).toBeUndefined()
    expect(buildMemoryConfig(v, 'AGENT_TYPE_CURSOR')).toBeUndefined()
  })

  it('never writes agent scope without tools', () => {
    expect(
      buildMemoryConfig(
        values({ enabled: true, agentScopeWrite: true }),
        'AGENT_TYPE_LLM'
      )
    ).toEqual({ enabled: true })
  })

  it('round-trips a stored config', () => {
    const stored = {
      enabled: true,
      disable_auto_capture: true,
      enable_tools: true,
      top_k: 7,
      threshold: 0.5,
    }
    expect(
      buildMemoryConfig(memoryFormValuesFromConfig(stored), 'AGENT_TYPE_LLM')
    ).toEqual(stored)
    expect(memoryFormValuesFromConfig(undefined)).toEqual(
      EMPTY_MEMORY_FORM_VALUES
    )
  })

  it('validates bounds only when enabled', () => {
    expect(
      memoryFormSchema.safeParse(values({ enabled: true, topK: '0' })).success
    ).toBe(false)
    expect(
      memoryFormSchema.safeParse(values({ enabled: true, topK: '51' })).success
    ).toBe(false)
    expect(
      memoryFormSchema.safeParse(values({ enabled: true, topK: '2.5' })).success
    ).toBe(false)
    expect(
      memoryFormSchema.safeParse(values({ enabled: true, threshold: '1.5' }))
        .success
    ).toBe(false)
    expect(
      memoryFormSchema.safeParse(values({ enabled: true, threshold: 'abc' }))
        .success
    ).toBe(false)
    expect(
      memoryFormSchema.safeParse(
        values({ enabled: true, topK: '50', threshold: '0' })
      ).success
    ).toBe(true)
    expect(
      memoryFormSchema.safeParse(values({ enabled: false, topK: '0' })).success
    ).toBe(true)
  })

  it('knows which types support memory and tools', () => {
    expect(supportsMemory('AGENT_TYPE_LLM')).toBe(true)
    expect(supportsMemory('AGENT_TYPE_WORKFLOW')).toBe(true)
    expect(supportsMemory('AGENT_TYPE_PI')).toBe(false)
    expect(supportsMemoryTools('AGENT_TYPE_LLM')).toBe(true)
    expect(supportsMemoryTools('AGENT_TYPE_LOOP')).toBe(false)
  })
})
