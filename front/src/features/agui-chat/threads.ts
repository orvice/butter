import type { SessionInfo } from '@/types/api'

// An AG-UI thread is the session `agui-{threadId}` under the ADK app `agui`,
// keyed by (caller, thread) only. Which workspace and agent a thread belongs
// to is recorded once, when the handler creates the session, in the hidden
// A2UI binding (ADR-0014) — the only place the dashboard can read it from.

export const AGUI_APP_NAME = 'agui'
// THREAD_PAGE_SIZE is one page of the thread listing, which reads every page.
export const THREAD_PAGE_SIZE = 100
// TITLE_REFRESH_DELAYS_MS are when the thread list is read again after a run
// on an untitled thread. The server titles the thread in the background once
// the run succeeds: within moments, or after a title model call that gives up
// after 10 seconds.
export const TITLE_REFRESH_DELAYS_MS = [3_000, 12_000]
const SESSION_PREFIX = 'agui-'
const BINDING_KEY = 'butter:a2ui:binding'

export interface ThreadBinding {
  workspaceId: string
  agentId: string
  threadId: string
}

// threadBinding reads a session's binding. The value is a JSON string (the
// backend never stores nested documents in UI state); a session created
// before A2UI has none.
export function threadBinding(session: SessionInfo): ThreadBinding | null {
  let raw: unknown = session.state?.[BINDING_KEY]
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      return null
    }
  }
  if (!raw || typeof raw !== 'object') return null
  const b = raw as Record<string, unknown>
  const workspaceId = typeof b.workspace_id === 'string' ? b.workspace_id : ''
  const agentId = typeof b.agent_id === 'string' ? b.agent_id : ''
  const threadId = typeof b.thread_id === 'string' ? b.thread_id : ''
  if (!workspaceId || !agentId || !threadId) return null
  return { workspaceId, agentId, threadId }
}

export function threadIdOf(session: SessionInfo): string | null {
  return session.session_id.startsWith(SESSION_PREFIX)
    ? session.session_id.slice(SESSION_PREFIX.length)
    : null
}

// agentThreads keeps the caller's threads with one agent in one workspace,
// in the order given (ListSessions answers newest first). A thread is listed
// only when its binding matches its own session ID: reusing a threadId under
// another agent lands on the same session, and that session is not this
// agent's conversation. Threads without a binding predate A2UI and cannot be
// attributed, so they are left out.
export function agentThreads(
  sessions: SessionInfo[],
  workspaceId: string,
  agentId: string
): SessionInfo[] {
  return sessions.filter((s) => {
    if (s.app_name !== AGUI_APP_NAME) return false
    const binding = threadBinding(s)
    return (
      !!binding &&
      binding.workspaceId === workspaceId &&
      binding.agentId === agentId &&
      binding.threadId === threadIdOf(s)
    )
  })
}

export function threadTitle(session: SessionInfo): string {
  return session.title?.trim() || 'Untitled thread'
}
