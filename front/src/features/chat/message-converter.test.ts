import { describe, expect, it } from 'vitest'
import type { ParsedEvent } from '@/lib/session-events'
import {
  convertSessionEvents,
  makeOptimisticUserMessage,
  makeStreamingAssistantMessage,
} from './message-converter'

function makeTextEvent(overrides: Partial<ParsedEvent> = {}): ParsedEvent {
  return {
    eventId: 'evt-1',
    author: 'agent',
    role: 'assistant',
    text: 'Hello world',
    toolCalls: [],
    toolResponses: [],
    timestamp: '2026-01-01T00:00:00Z',
    raw: {
      event_id: overrides.eventId ?? 'evt-1',
      author: overrides.author ?? 'agent',
      content_json: JSON.stringify({
        role: overrides.role ?? 'assistant',
        parts: [{ text: overrides.text ?? 'Hello world' }],
      }),
      timestamp: overrides.timestamp ?? '2026-01-01T00:00:00Z',
    },
    ...overrides,
  }
}

function makeToolCallEvent(
  eventId: string,
  callName: string,
  args: unknown
): ParsedEvent {
  return {
    eventId,
    author: 'agent',
    role: 'assistant',
    text: '',
    toolCalls: [{ name: callName, argsPreview: JSON.stringify(args) }],
    toolResponses: [],
    timestamp: '2026-01-01T00:00:00Z',
    raw: {
      event_id: eventId,
      author: 'agent',
      content_json: JSON.stringify({
        role: 'model',
        parts: [{ functionCall: { name: callName, args } }],
      }),
      timestamp: '2026-01-01T00:00:00Z',
    },
  }
}

function makeToolResponseEvent(
  eventId: string,
  name: string,
  response: unknown
): ParsedEvent {
  return {
    eventId,
    author: 'agent',
    role: 'assistant',
    text: '',
    toolCalls: [],
    toolResponses: [{ name, responsePreview: JSON.stringify(response) }],
    timestamp: '2026-01-01T00:00:00Z',
    raw: {
      event_id: eventId,
      author: 'agent',
      content_json: JSON.stringify({
        role: 'model',
        parts: [{ functionResponse: { name, response } }],
      }),
      timestamp: '2026-01-01T00:00:00Z',
    },
  }
}

describe('convertSessionEvents', () => {
  it('converts text events', () => {
    const events = [
      makeTextEvent({ eventId: 'e1', text: 'Hello' }),
      makeTextEvent({
        eventId: 'e2',
        role: 'user',
        author: 'user',
        text: 'Hi',
      }),
    ]
    const result = convertSessionEvents(events)
    expect(result).toHaveLength(2)
    expect(result[0].role).toBe('assistant')
    expect(result[0].id).toBe('e1')
    expect(result[1].role).toBe('user')
  })

  it('filters out empty-text events with no tool calls', () => {
    const events = [
      makeTextEvent({ eventId: 'e1', text: 'visible' }),
      makeTextEvent({ eventId: 'e2', text: '' }),
      makeTextEvent({ eventId: 'e3', text: '   ' }),
    ]
    const result = convertSessionEvents(events)
    expect(result).toHaveLength(1)
    expect(result[0].id).toBe('e1')
  })

  it('includes tool-call parts for function calls', () => {
    const events = [makeToolCallEvent('tc1', 'search', { query: 'test' })]
    const result = convertSessionEvents(events)
    expect(result).toHaveLength(1)
    const content = result[0].content as readonly { type: string }[]
    const toolCall = content.find((p) => p.type === 'tool-call') as unknown as {
      toolName: string
      args: Record<string, unknown>
    }
    expect(toolCall).toBeTruthy()
    expect(toolCall.toolName).toBe('search')
    expect(toolCall.args).toEqual({ query: 'test' })
  })

  it('matches function calls with responses', () => {
    const events = [
      makeToolCallEvent('tc1', 'search', { query: 'test' }),
      makeToolResponseEvent('tr1', 'search', { results: [1, 2, 3] }),
    ]
    const result = convertSessionEvents(events)
    expect(result).toHaveLength(1)
    const content = result[0].content as readonly { type: string }[]
    const toolCall = content.find((p) => p.type === 'tool-call') as unknown as {
      toolName: string
      result: unknown
    }
    expect(toolCall.toolName).toBe('search')
    expect(toolCall.result).toEqual({ results: [1, 2, 3] })
  })

  it('handles unmatched function responses', () => {
    const events = [
      makeToolResponseEvent('tr1', 'unknown_tool', { data: 'result' }),
    ]
    const result = convertSessionEvents(events)
    expect(result).toHaveLength(1)
    const content = result[0].content as readonly { type: string }[]
    const part = content.find((p) => p.type === 'tool-call') as unknown as {
      toolName: string
      result: unknown
    }
    expect(part.toolName).toBe('unknown_tool')
    expect(part.result).toEqual({ data: 'result' })
  })

  it('handles text + tool call in the same event', () => {
    const event: ParsedEvent = {
      eventId: 'mixed1',
      author: 'agent',
      role: 'assistant',
      text: 'Searching...',
      toolCalls: [{ name: 'search', argsPreview: '{}' }],
      toolResponses: [],
      timestamp: '2026-01-01T00:00:00Z',
      raw: {
        event_id: 'mixed1',
        author: 'agent',
        content_json: JSON.stringify({
          role: 'model',
          parts: [
            { text: 'Searching...' },
            { functionCall: { name: 'search', args: {} } },
          ],
        }),
        timestamp: '2026-01-01T00:00:00Z',
      },
    }
    const result = convertSessionEvents([event])
    expect(result).toHaveLength(1)
    const content = result[0].content as readonly { type: string }[]
    expect(content.some((p) => p.type === 'text')).toBe(true)
    expect(content.some((p) => p.type === 'tool-call')).toBe(true)
  })

  it('returns empty array for no events', () => {
    expect(convertSessionEvents([])).toEqual([])
  })

  it('uses FIFO matching for multiple calls to same tool', () => {
    const events = [
      makeToolCallEvent('c1', 'fetch', { url: '/a' }),
      makeToolCallEvent('c2', 'fetch', { url: '/b' }),
      makeToolResponseEvent('r1', 'fetch', { body: 'response-a' }),
      makeToolResponseEvent('r2', 'fetch', { body: 'response-b' }),
    ]
    const result = convertSessionEvents(events)
    const toolCalls = result
      .flatMap((m) => m.content as readonly { type: string }[])
      .filter((p) => p.type === 'tool-call') as {
      args?: { url?: string }
      result?: { body?: string }
    }[]

    const first = toolCalls.find((tc) => tc.args?.url === '/a')
    expect(first?.result).toEqual({ body: 'response-a' })

    const second = toolCalls.find((tc) => tc.args?.url === '/b')
    expect(second?.result).toEqual({ body: 'response-b' })
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
