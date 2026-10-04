// What a chat shows about one tool call, read from its message part.

export interface ToolCallPart {
  args?: unknown
  result?: unknown
  isError?: boolean
  status?: { type: string }
}

// ToolCallState is what a tool call's header says about it.
export type ToolCallState = 'running' | 'pending' | 'done' | 'failed' | 'idle'

export function toolCallState(call: ToolCallPart): ToolCallState {
  if (call.isError) return 'failed'
  if (call.result !== undefined) return 'done'
  switch (call.status?.type) {
    case 'running':
      return 'running'
    case 'requires-action':
      return 'pending'
    default:
      return 'idle'
  }
}

// hasArgs reports whether a call carries arguments worth showing.
export function hasArgs(args: unknown): boolean {
  return (
    !!args && typeof args === 'object' && Object.keys(args as object).length > 0
  )
}

// formatJson shows a tool's arguments or result: text as it is, anything
// else as indented JSON.
export function formatJson(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string') return value
  try {
    return JSON.stringify(value, null, 2) ?? String(value)
  } catch {
    return String(value)
  }
}

// renderUIFailed reports whether a render_ui call drew no card: the tool
// answers {"error": "..."} then, and the call is the only trace of it.
export function renderUIFailed(result: unknown): boolean {
  return !!result && typeof result === 'object' && 'error' in result
}

// humanInputQuestion is the question an adk_request_input call asks.
export function humanInputQuestion(args: unknown): string {
  if (args && typeof args === 'object') {
    const { question, message } = args as Record<string, unknown>
    if (typeof question === 'string') return question
    if (typeof message === 'string') return message
  }
  return formatJson(args)
}

// awaitsAnswer reports whether a Human Input question is still open: it has
// no answer, and its run is going or paused on it.
export function awaitsAnswer(call: ToolCallPart): boolean {
  const state = toolCallState(call)
  return state === 'running' || state === 'pending'
}
