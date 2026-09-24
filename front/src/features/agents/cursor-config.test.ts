import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import {
  asCursorAgent,
  buildCursorAgentConfig,
  cursorAgentFormSchema,
  cursorFormValuesFromConfig,
  validateCursorAgentForm,
} from './cursor-config'

describe('buildCursorAgentConfig', () => {
  it('serializes the binding and preserves an explicit unlimited timeout', () => {
    expect(
      buildCursorAgentConfig({
        butterboxId: ' box-1 ',
        workingDir: ' /workspace/project ',
        model: ' composer-2.5 ',
        mode: 'plan',
        maxRunSeconds: '0',
      })
    ).toEqual({
      cursor: {
        butterbox_id: 'box-1',
        working_dir: '/workspace/project',
        model: 'composer-2.5',
        mode: 'plan',
        max_run_seconds: 0,
      },
    })
  })

  it('leaves the default timeout absent', () => {
    expect(
      buildCursorAgentConfig({
        butterboxId: 'box-1',
        workingDir: '',
        model: '',
        mode: 'agent',
        maxRunSeconds: '',
      })
    ).toEqual({
      cursor: { butterbox_id: 'box-1', working_dir: '', model: '', mode: 'agent' },
    })
  })
})

describe('cursorFormValuesFromConfig', () => {
  it('defaults the mode to agent and does not invent a timeout', () => {
    expect(cursorFormValuesFromConfig({ butterbox_id: 'box-1' })).toEqual({
      butterboxId: 'box-1',
      workingDir: '',
      model: '',
      mode: 'agent',
      maxRunSeconds: '',
    })
  })

  it('hydrates an existing binding', () => {
    expect(
      cursorFormValuesFromConfig({
        butterbox_id: 'box-1',
        working_dir: 'repo',
        model: 'auto-smart',
        mode: 'plan',
        max_run_seconds: 600,
      })
    ).toEqual({
      butterboxId: 'box-1',
      workingDir: 'repo',
      model: 'auto-smart',
      mode: 'plan',
      maxRunSeconds: '600',
    })
  })
})

describe('asCursorAgent', () => {
  it('drops children and butter-side behavior', () => {
    const agent = asCursorAgent(
      {
        name: 'coder',
        type: 'AGENT_TYPE_LLM',
        child_agent_ids: ['x'],
        config: { instruction: 'hi', mcp_server_ids: ['m'] },
      },
      { butterboxId: 'box-1', workingDir: '', model: '', mode: 'agent', maxRunSeconds: '' }
    )
    expect(agent.type).toBe('AGENT_TYPE_CURSOR')
    expect(agent.child_agent_ids).toEqual([])
    expect(agent.config).toEqual({
      cursor: { butterbox_id: 'box-1', working_dir: '', model: '', mode: 'agent' },
    })
  })
})

describe('validateCursorAgentForm', () => {
  const schema = z
    .object({ cursor: cursorAgentFormSchema })
    .superRefine((values, ctx) => validateCursorAgentForm(values.cursor, ctx))

  it('requires a ButterBox and a whole-number timeout', () => {
    const result = schema.safeParse({
      cursor: { butterboxId: ' ', workingDir: '', model: '', mode: 'agent', maxRunSeconds: '-1' },
    })
    expect(result.success).toBe(false)
    const paths = result.error?.issues.map((issue) => issue.path.join('.'))
    expect(paths).toEqual(['cursor.butterboxId', 'cursor.maxRunSeconds'])
  })
})
