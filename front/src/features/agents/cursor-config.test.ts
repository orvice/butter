import { describe, expect, it } from 'vitest'
import {
  asCursorAgent,
  buildCursorAgentConfig,
  cursorFormValuesFromConfig,
  validateCursorAgentForm,
} from './cursor-config'

describe('buildCursorAgentConfig', () => {
  it('serializes the box-owned fields and preserves an explicit unlimited timeout', () => {
    expect(
      buildCursorAgentConfig({
        butterboxId: ' box-1 ',
        workingDir: ' /workspace/project ',
        model: ' composer-2.5 ',
        mode: ' agent ',
        maxRunSeconds: '0',
      })
    ).toEqual({
      cursor: {
        butterbox_id: 'box-1',
        working_dir: '/workspace/project',
        model: 'composer-2.5',
        mode: 'agent',
        max_run_seconds: 0,
      },
    })
  })

  it('leaves the default timeout absent and the mode empty', () => {
    expect(
      buildCursorAgentConfig({
        butterboxId: 'box-1',
        workingDir: '',
        model: '',
        mode: '',
        maxRunSeconds: '',
      })
    ).toEqual({
      cursor: {
        butterbox_id: 'box-1',
        working_dir: '',
        model: '',
        mode: '',
      },
    })
  })
})

describe('cursorFormValuesFromConfig', () => {
  it('hydrates an existing Cursor binding without inventing an optional timeout', () => {
    expect(
      cursorFormValuesFromConfig({
        butterbox_id: 'box-1',
        working_dir: 'projects/demo',
        model: 'auto-smart',
        mode: 'plan',
      })
    ).toEqual({
      butterboxId: 'box-1',
      workingDir: 'projects/demo',
      model: 'auto-smart',
      mode: 'plan',
      maxRunSeconds: '',
    })
  })
})

describe('validateCursorAgentForm', () => {
  it('requires a ButterBox and rejects a bogus mode or max run', () => {
    const issues: Array<{ path: string[]; message: string }> = []
    validateCursorAgentForm(
      {
        butterboxId: ' ',
        workingDir: '',
        model: '',
        mode: 'execute',
        maxRunSeconds: 'not-a-number',
      },
      {
        addIssue: (issue) => issues.push(issue as { path: string[]; message: string }),
      } as never
    )
    const messages = issues.map((i) => i.message).join(' | ')
    expect(messages).toContain('Select a ButterBox')
    expect(messages).toContain('Mode must be empty')
    expect(messages).toContain('2147483647')
  })

  it('accepts an empty model (box default) and an explicit 0 timeout', () => {
    const issues: Array<{ path: string[]; message: string }> = []
    validateCursorAgentForm(
      {
        butterboxId: 'box-1',
        workingDir: '',
        model: '',
        mode: '',
        maxRunSeconds: '0',
      },
      {
        addIssue: (issue) => issues.push(issue as { path: string[]; message: string }),
      } as never
    )
    expect(issues).toHaveLength(0)
  })
})

describe('asCursorAgent', () => {
  it('clears composed-agent children when converting an existing Agent to Cursor', () => {
    const converted = asCursorAgent(
      {
        name: 'coordinator',
        type: 'AGENT_TYPE_SEQUENTIAL',
        child_agent_ids: ['researcher'],
        config: { instruction: 'Coordinate the child.' },
      },
      {
        butterboxId: 'box-1',
        workingDir: 'projects/demo',
        model: '',
        mode: '',
        maxRunSeconds: '',
      }
    )

    expect(converted.type).toBe('AGENT_TYPE_CURSOR')
    expect(converted.child_agent_ids).toEqual([])
    expect(converted.config).toEqual({
      cursor: {
        butterbox_id: 'box-1',
        working_dir: 'projects/demo',
        model: '',
        mode: '',
      },
    })
  })
})
