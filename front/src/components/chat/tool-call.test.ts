import { describe, expect, it } from 'vitest'
import {
  awaitsAnswer,
  formatJson,
  hasArgs,
  humanInputQuestion,
  renderUIFailed,
  toolCallState,
} from './tool-call'

describe('toolCallState', () => {
  it('reads a result as done, and an error result as failed', () => {
    expect(
      toolCallState({ status: { type: 'complete' }, result: { hits: 3 } })
    ).toBe('done')
    expect(toolCallState({ status: { type: 'running' }, result: '' })).toBe(
      'done'
    )
    expect(
      toolCallState({
        status: { type: 'complete' },
        result: 'x',
        isError: true,
      })
    ).toBe('failed')
  })

  it('reads a call without a result by its status', () => {
    expect(toolCallState({ status: { type: 'running' } })).toBe('running')
    expect(toolCallState({ status: { type: 'requires-action' } })).toBe(
      'pending'
    )
    expect(toolCallState({ status: { type: 'complete' } })).toBe('idle')
    expect(toolCallState({ status: { type: 'incomplete' } })).toBe('idle')
    expect(toolCallState({})).toBe('idle')
  })
})

describe('awaitsAnswer', () => {
  it('holds while the question has no answer and its run is going or paused', () => {
    expect(awaitsAnswer({ status: { type: 'running' } })).toBe(true)
    expect(awaitsAnswer({ status: { type: 'requires-action' } })).toBe(true)
    expect(
      awaitsAnswer({ status: { type: 'requires-action' }, result: 'yes' })
    ).toBe(false)
    expect(awaitsAnswer({ status: { type: 'complete' } })).toBe(false)
  })
})

describe('hasArgs', () => {
  it('holds only for an object with at least one key', () => {
    expect(hasArgs({ q: 'go' })).toBe(true)
    expect(hasArgs({})).toBe(false)
    expect(hasArgs(undefined)).toBe(false)
    expect(hasArgs(null)).toBe(false)
    expect(hasArgs('text')).toBe(false)
  })
})

describe('formatJson', () => {
  it('shows text as it is and anything else as indented JSON', () => {
    expect(formatJson('plain')).toBe('plain')
    expect(formatJson({ a: [1] })).toBe('{\n  "a": [\n    1\n  ]\n}')
    expect(formatJson(3)).toBe('3')
    expect(formatJson(false)).toBe('false')
  })

  it('shows nothing for no value, and never throws', () => {
    expect(formatJson(undefined)).toBe('')
    expect(formatJson(null)).toBe('')
    const cyclic: Record<string, unknown> = {}
    cyclic.self = cyclic
    expect(formatJson(cyclic)).toBe('[object Object]')
  })
})

describe('renderUIFailed', () => {
  it('holds for the error the tool answers when it drew no card', () => {
    expect(renderUIFailed({ error: 'card not rendered' })).toBe(true)
    expect(
      renderUIFailed({ surface_id: 'card-1', revision: 1, status: 'created' })
    ).toBe(false)
    expect(renderUIFailed(undefined)).toBe(false)
    expect(renderUIFailed('error')).toBe(false)
  })
})

describe('humanInputQuestion', () => {
  it('reads the question, then the message, then shows the arguments', () => {
    expect(humanInputQuestion({ question: 'Ship?', message: 'm' })).toBe(
      'Ship?'
    )
    expect(humanInputQuestion({ message: 'Approve?' })).toBe('Approve?')
    expect(humanInputQuestion({ question: 3 })).toBe('{\n  "question": 3\n}')
    expect(humanInputQuestion(undefined)).toBe('')
  })
})
