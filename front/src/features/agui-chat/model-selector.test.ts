import type { Agent } from '@/types/api'
import { describe, expect, it } from 'vitest'
import { chatModelOptions, supportsChatModelOverride } from './model-selector'

describe('chat model selection', () => {
  it('uses aliases, falls back to names, deduplicates and sorts the catalog', () => {
    expect(
      chatModelOptions([
        {
          name: 'Provider',
          type: 'openai',
          models: [
            { name: 'z-model', alias: 'fast' },
            { name: 'a-model' },
            { name: '' },
          ],
        },
        {
          name: 'Duplicate',
          type: 'openai',
          models: [{ name: 'other', alias: 'fast' }],
        },
      ])
    ).toEqual([
      { value: 'a-model', label: 'a-model', provider: 'Provider' },
      { value: 'fast', label: 'fast', provider: 'Provider' },
    ])
    expect(chatModelOptions()).toEqual([])
  })

  it('offers overrides only for LLM agents, including legacy unspecified types', () => {
    const agent: Agent = { name: 'Agent' }
    expect(supportsChatModelOverride(undefined)).toBe(false)
    expect(supportsChatModelOverride(agent)).toBe(true)
    for (const type of ['AGENT_TYPE_LLM', 'AGENT_TYPE_UNSPECIFIED'] as const) {
      expect(supportsChatModelOverride({ ...agent, type })).toBe(true)
    }
    for (const type of [
      'AGENT_TYPE_PI',
      'AGENT_TYPE_CURSOR',
      'AGENT_TYPE_WORKFLOW',
      'AGENT_TYPE_SEQUENTIAL',
      'AGENT_TYPE_PARALLEL',
      'AGENT_TYPE_LOOP',
    ] as const) {
      expect(supportsChatModelOverride({ ...agent, type })).toBe(false)
    }
  })
})
