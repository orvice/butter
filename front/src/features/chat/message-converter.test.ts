import { describe, expect, it } from 'vitest'
import type { ParsedEvent } from '@/lib/session-events'
import {
  convertSessionEvent,
  convertSessionEvents,
  makeOptimisticUserMessage,
  makeStreamingAssistantMessage,
} from './message-converter'

function makeEvent(overrides: Partial<ParsedEvent> = {}): ParsedEvent {
  return {
    eventId: 'evt-1',
    author: 'agent',
    role: 'assistant',
    text: 'Hello world',
    toolCalls: [],
    toolResponses: [],
    timestamp: '2026-01-01T00:00:00Z',
    raw: { event_id: 'evt-1' },
    ...overrides,
  }
}

describe('convertSessionEvent', () => {
  it('maps assistant event to assistant role', () => {
    const result = convertSessionEvent(makeEvent({ role: 'assistant' }))
    expect(result.role).toBe('assistant')
    expect(result.content).toBe('Hello world')
    expect(result.id).toBe('evt-1')
  })

  it('maps user event to user role', () => {
    const result = convertSessionEvent(
      makeEvent({ role: 'user', author: 'user' })
    )
    expect(result.role).toBe('user')
  })

  it('maps system event to system role', () => {
    const result = convertSessionEvent(makeEvent({ role: 'system' }))
    expect(result.role).toBe('system')
  })

  it('sets createdAt from timestamp', () => {
    const result = convertSessionEvent(makeEvent())
    expect(result.createdAt).toEqual(new Date('2026-01-01T00:00:00Z'))
  })

  it('handles missing timestamp', () => {
    const result = convertSessionEvent(makeEvent({ timestamp: undefined }))
    expect(result.createdAt).toBeUndefined()
  })
})

describe('convertSessionEvents', () => {
  it('filters out empty-text events', () => {
    const events = [
      makeEvent({ eventId: 'e1', text: 'visible' }),
      makeEvent({ eventId: 'e2', text: '' }),
      makeEvent({ eventId: 'e3', text: '   ' }),
      makeEvent({ eventId: 'e4', text: 'also visible' }),
    ]
    const result = convertSessionEvents(events)
    expect(result).toHaveLength(2)
    expect(result[0].id).toBe('e1')
    expect(result[1].id).toBe('e4')
  })

  it('returns empty array for no events', () => {
    expect(convertSessionEvents([])).toEqual([])
  })
})

describe('makeOptimisticUserMessage', () => {
  it('creates a user message with correct fields', () => {
    const result = makeOptimisticUserMessage('opt-1', 'Hello')
    expect(result.role).toBe('user')
    expect(result.content).toBe('Hello')
    expect(result.id).toBe('opt-1')
    expect(result.createdAt).toBeInstanceOf(Date)
  })
})

describe('makeStreamingAssistantMessage', () => {
  it('creates a running assistant message', () => {
    const result = makeStreamingAssistantMessage('stream-1', 'partial response')
    expect(result.role).toBe('assistant')
    expect(result.content).toBe('partial response')
    expect(result.id).toBe('stream-1')
    expect(result.status).toEqual({ type: 'running' })
  })
})
