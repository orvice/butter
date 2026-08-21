import type { ThreadMessageLike } from '@assistant-ui/react'
import type { ParsedEvent } from '@/lib/session-events'

export function convertSessionEvent(event: ParsedEvent): ThreadMessageLike {
  return {
    role: event.role === 'system' ? 'system' : event.role,
    content: event.text || '',
    id: event.eventId,
    createdAt: event.timestamp ? new Date(event.timestamp) : undefined,
  }
}

export function convertSessionEvents(
  events: ParsedEvent[]
): ThreadMessageLike[] {
  return events.filter(isRenderableEvent).map(convertSessionEvent)
}

function isRenderableEvent(event: ParsedEvent): boolean {
  return event.text.trim().length > 0
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
