import { describe, expect, it } from 'vitest'
import {
  asPiAgent,
  buildPiAgentConfig,
  piFormValuesFromConfig,
  piModelCatalogKey,
  piModelFromCatalogKey,
} from './pi-config'

describe('buildPiAgentConfig', () => {
  it('serializes the box-owned fields and preserves an explicit unlimited timeout', () => {
    expect(
      buildPiAgentConfig({
        butterboxId: ' box-1 ',
        workingDir: ' /workspace/project ',
        provider: ' anthropic ',
        model: ' claude-sonnet ',
        thinkingLevel: ' high ',
        maxRunSeconds: '0',
      })
    ).toEqual({
      pi: {
        butterbox_id: 'box-1',
        working_dir: '/workspace/project',
        provider: 'anthropic',
        model: 'claude-sonnet',
        thinking_level: 'high',
        max_run_seconds: 0,
      },
    })
  })

  it('leaves the default timeout absent', () => {
    expect(
      buildPiAgentConfig({
        butterboxId: 'box-1',
        workingDir: '',
        provider: '',
        model: '',
        thinkingLevel: '',
        maxRunSeconds: '',
      })
    ).toEqual({
      pi: {
        butterbox_id: 'box-1',
        working_dir: '',
        provider: '',
        model: '',
        thinking_level: '',
      },
    })
  })
})

describe('piFormValuesFromConfig', () => {
  it('hydrates an existing Pi binding without inventing an optional timeout', () => {
    expect(
      piFormValuesFromConfig({
        butterbox_id: 'box-1',
        working_dir: 'repo',
        provider: 'openai',
        model: 'gpt-5',
        thinking_level: 'medium',
      })
    ).toEqual({
      butterboxId: 'box-1',
      workingDir: 'repo',
      provider: 'openai',
      model: 'gpt-5',
      thinkingLevel: 'medium',
      maxRunSeconds: '',
    })
  })
})

describe('asPiAgent', () => {
  it('clears composed-agent children when converting an existing Agent to Pi', () => {
    const converted = asPiAgent(
      {
        name: 'coordinator',
        type: 'AGENT_TYPE_SEQUENTIAL',
        child_agent_ids: ['researcher'],
        sub_agents: [{ name: 'researcher' }],
        config: { instruction: 'Coordinate the child.' },
      },
      {
        butterboxId: 'box-1',
        workingDir: 'repo',
        provider: '',
        model: '',
        thinkingLevel: '',
        maxRunSeconds: '',
      }
    )

    expect(converted.type).toBe('AGENT_TYPE_PI')
    expect(converted.child_agent_ids).toEqual([])
    expect(converted.sub_agents).toEqual([{ name: 'researcher' }])
    expect(converted.config).toEqual({
      pi: {
        butterbox_id: 'box-1',
        working_dir: 'repo',
        provider: '',
        model: '',
        thinking_level: '',
      },
    })
  })
})

describe('Pi model catalog selection', () => {
  it('distinguishes the same model ID exposed by different providers', () => {
    const models = [
      { id: 'shared-model', provider: 'provider-a' },
      { id: 'shared-model', provider: 'provider-b' },
    ]

    const secondKey = piModelCatalogKey(models[1])
    expect(secondKey).not.toBe(piModelCatalogKey(models[0]))
    expect(piModelFromCatalogKey(models, secondKey)).toEqual(models[1])
  })
})
