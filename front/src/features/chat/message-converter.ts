import type { SessionEvent } from '@/types/api'
import type { ThreadMessageLike } from '@assistant-ui/react'
import {
  parseSessionEventsFull,
  type FullParsedEvent,
  type ParsedEvent,
  type ParsedFunctionResponse,
} from '@/lib/session-events'

type ContentPart = NonNullable<
  Exclude<ThreadMessageLike['content'], string>
>[number]

interface ToolCallPart {
  readonly type: 'tool-call'
  readonly toolCallId: string
  readonly toolName: string
  readonly args?: Record<string, unknown>
  readonly result?: unknown
  readonly isError?: boolean
}

/**
 * Convert raw session events into `assistant-ui` ThreadMessageLike messages.
 *
 * Tool calls are matched with their corresponding responses by scanning
 * forward through subsequent events. A function call that has a matching
 * response in a later event produces a single "completed" tool-call part;
 * unmatched calls stay as "running". Pure function-response events that
 * were already consumed by a forward match are dropped to avoid duplicates.
 */
export function convertSessionEvents(
  parsedEvents: ParsedEvent[]
): ThreadMessageLike[] {
  const rawEvents: SessionEvent[] = parsedEvents.map((e) => e.raw)
  const fullEvents = parseSessionEventsFull(rawEvents)

  const responseMap = buildResponseMap(fullEvents)
  const consumedResponseKeys = new Set<string>()
  const messages: ThreadMessageLike[] = []

  for (const event of fullEvents) {
    const content: ContentPart[] = []

    for (const tp of event.textParts) {
      if (tp.text.trim()) {
        content.push({ type: 'text' as const, text: tp.text })
      }
    }

    for (let i = 0; i < event.functionCalls.length; i++) {
      const call = event.functionCalls[i]
      const key = responseKey(call.name, event.eventId, i)
      const response = responseMap.get(key)
      const part: ToolCallPart = {
        type: 'tool-call',
        toolCallId: `${event.eventId}-call-${i}`,
        toolName: call.name,
        args: normalizeArgs(call.args),
        ...(response ? { result: response.resp.response } : {}),
      }
      content.push(part as ContentPart)
      if (response) consumedResponseKeys.add(response.consumeKey)
    }

    for (let i = 0; i < event.functionResponses.length; i++) {
      const resp = event.functionResponses[i]
      const consumeKey = `${event.eventId}-resp-${i}`
      if (consumedResponseKeys.has(consumeKey)) continue

      const part: ToolCallPart = {
        type: 'tool-call',
        toolCallId: `${event.eventId}-resp-${i}`,
        toolName: resp.name,
        result: resp.response,
      }
      content.push(part as ContentPart)
    }

    if (content.length === 0) continue

    messages.push({
      role: event.role,
      content,
      id: event.eventId,
      createdAt: event.timestamp ? new Date(event.timestamp) : undefined,
    })
  }

  return messages
}

/**
 * Build a map from each function call (identified by name + source event)
 * to its matching function response in a later event. Uses FIFO matching
 * per tool name so concurrent calls to the same tool resolve correctly.
 */
function buildResponseMap(
  events: FullParsedEvent[]
): Map<string, { resp: ParsedFunctionResponse; consumeKey: string }> {
  const result = new Map<
    string,
    { resp: ParsedFunctionResponse; consumeKey: string }
  >()
  const pending: { name: string; key: string }[] = []

  for (const event of events) {
    for (let i = 0; i < event.functionCalls.length; i++) {
      pending.push({
        name: event.functionCalls[i].name,
        key: responseKey(event.functionCalls[i].name, event.eventId, i),
      })
    }
    for (let i = 0; i < event.functionResponses.length; i++) {
      const resp = event.functionResponses[i]
      const idx = pending.findIndex((p) => p.name === resp.name)
      if (idx !== -1) {
        const matched = pending.splice(idx, 1)[0]
        result.set(matched.key, {
          resp,
          consumeKey: `${event.eventId}-resp-${i}`,
        })
      }
    }
  }
  return result
}

function responseKey(name: string, eventId: string, index: number): string {
  return `${eventId}-call-${index}-${name}`
}

function normalizeArgs(args: unknown): Record<string, unknown> | undefined {
  if (args === null || args === undefined) return undefined
  if (typeof args === 'object' && !Array.isArray(args))
    return args as Record<string, unknown>
  return { _: args }
}

export function makeOptimisticUserMessage(
  id: string,
  text: string
): ThreadMessageLike {
  return {
    role: 'user',
    content: text,
    id,
    createdAt: new Date(),
  }
}

export function makeStreamingAssistantMessage(
  id: string,
  text: string
): ThreadMessageLike {
  return {
    role: 'assistant',
    content: text,
    id,
    status: { type: 'running' },
  }
}
