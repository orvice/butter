import type { Agent, SessionInfo } from '@/types/api'

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

// newThreadId names a new thread. The URL carries it (?thread=) from the
// first message on, and the server creates its session on the first run.
export function newThreadId(): string {
  return crypto.randomUUID()
}

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

// sessionIdOf is the session ID of the AG-UI thread threadId.
export function sessionIdOf(threadId: string): string {
  return SESSION_PREFIX + threadId
}

// boundAgentId is the agent session holds the thread threadId with in
// workspaceId: the one its binding names, when that binding is for this
// workspace and this thread. Anything else is null: another app's session, a
// thread without a binding (it predates A2UI and cannot be attributed),
// another workspace's thread, or a binding for another thread.
export function boundAgentId(
  session: SessionInfo,
  workspaceId: string,
  threadId: string
): string | null {
  if (session.app_name !== AGUI_APP_NAME) return null
  if (threadIdOf(session) !== threadId) return null
  const binding = threadBinding(session)
  if (
    !binding ||
    binding.workspaceId !== workspaceId ||
    binding.threadId !== threadId
  ) {
    return null
  }
  return binding.agentId
}

// WorkspaceThread is one of the caller's threads in a workspace, with the
// agent its binding names.
export interface WorkspaceThread {
  session: SessionInfo
  threadId: string
  agentId: string
}

// workspaceThreads keeps the caller's threads in one workspace, with every
// agent, in the order given (ListSessions answers newest first). A thread is
// listed only when its binding is for this workspace and for the thread its
// session holds. Threads without a binding predate A2UI and cannot be
// attributed, so they are left out.
export function workspaceThreads(
  sessions: readonly SessionInfo[],
  workspaceId: string
): WorkspaceThread[] {
  const threads: WorkspaceThread[] = []
  for (const session of sessions) {
    const threadId = threadIdOf(session)
    if (threadId === null) continue
    const agentId = boundAgentId(session, workspaceId, threadId)
    if (agentId) threads.push({ session, threadId, agentId })
  }
  return threads
}

// threadTitle is the title a thread shows: its own, else the name of its
// agent, else (the agent is unknown) a placeholder.
export function threadTitle(session: SessionInfo, agentName?: string): string {
  return session.title?.trim() || agentName?.trim() || 'Untitled thread'
}

// ThreadRow is a thread as the sidebar lists it: with its agent as the
// workspace's agent list has it (undefined for one the list lacks), and the
// title it shows.
export interface ThreadRow extends WorkspaceThread {
  agent?: Agent
  title: string
}

// threadRows lists the caller's threads in one workspace, each with the agent
// its binding names, taken from agents.
export function threadRows(
  sessions: readonly SessionInfo[],
  workspaceId: string,
  agents: readonly Agent[]
): ThreadRow[] {
  const byId = new Map<string, Agent>()
  for (const agent of agents) {
    if (agent.agent_id) byId.set(agent.agent_id, agent)
  }
  return workspaceThreads(sessions, workspaceId).map((thread) => {
    const agent = byId.get(thread.agentId)
    return { ...thread, agent, title: threadTitle(thread.session, agent?.name) }
  })
}

// searchThreads keeps the rows whose title contains query, ignoring case.
export function searchThreads(
  rows: readonly ThreadRow[],
  query: string
): readonly ThreadRow[] {
  const q = query.trim().toLowerCase()
  if (!q) return rows
  return rows.filter((row) => row.title.toLowerCase().includes(q))
}

// ThreadView is what the page shows for the thread the URL names.
export type ThreadView =
  | { kind: 'loading' }
  | { kind: 'failed'; error: unknown }
  | { kind: 'not-found' }
  | { kind: 'open'; agentId: string; session: SessionInfo | null }

// ThreadLookup is everything known about the thread the URL names.
export interface ThreadLookup {
  threadId: string
  workspaceId: string
  // requestedAgentId is the agent the URL names next to the thread, if any.
  requestedAgentId?: string
  // startedAgentId is set for a thread this page started: it opens with that
  // agent, before its session exists and whatever a read of it says.
  startedAgentId?: string
  // session is the thread's session as GetSession read it: undefined until
  // read, null when there is none the caller may see.
  session?: SessionInfo | null
  sessionError?: unknown
  // listed is the thread as the thread list holds it, used until the
  // session is read.
  listed?: SessionInfo
  // agentIds are the agents this workspace can open; undefined until read.
  agentIds?: readonly string[]
  agentsError?: unknown
}

// resolveThreadView decides what opening a thread shows. A thread opens
// with the agent its binding names: an agent is fixed once a thread exists.
// Not found covers a missing session (another user's thread reads as
// missing too) and a session that is not this workspace's thread with an
// agent it can open: no binding, a binding for another workspace or thread,
// an agent the workspace cannot run, or another agent than the URL names.
export function resolveThreadView(lookup: ThreadLookup): ThreadView {
  if (lookup.startedAgentId) {
    return {
      kind: 'open',
      agentId: lookup.startedAgentId,
      session: lookup.session ?? lookup.listed ?? null,
    }
  }
  const session = lookup.session !== undefined ? lookup.session : lookup.listed
  if (session === undefined) {
    return lookup.sessionError != null
      ? { kind: 'failed', error: lookup.sessionError }
      : { kind: 'loading' }
  }
  if (session === null) return { kind: 'not-found' }
  const agentId = boundAgentId(session, lookup.workspaceId, lookup.threadId)
  if (
    !agentId ||
    (lookup.requestedAgentId && lookup.requestedAgentId !== agentId)
  ) {
    return { kind: 'not-found' }
  }
  if (lookup.agentIds === undefined) {
    return lookup.agentsError != null
      ? { kind: 'failed', error: lookup.agentsError }
      : { kind: 'loading' }
  }
  if (!lookup.agentIds.includes(agentId)) return { kind: 'not-found' }
  return { kind: 'open', agentId, session }
}
