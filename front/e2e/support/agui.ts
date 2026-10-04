import type { Page, Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import {
  DeleteSessionRequestSchema,
  DeleteSessionResponseSchema,
  GenerateSessionTitleRequestSchema,
  GenerateSessionTitleResponseSchema,
  GetSessionRequestSchema,
  GetSessionResponseSchema,
  ListAgentsResponseSchema,
  ListSessionsRequestSchema,
  ListSessionsResponseSchema,
  SessionInfoSchema,
  UpdateSessionTitleRequestSchema,
  UpdateSessionTitleResponseSchema,
  type SessionInfo,
} from '../../src/gen/agents/v1/agent_service_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
  type ConnectFixtureOptions,
} from './connect'

// AG-UI fixtures: POST /api/agui/:agent_id is answered from a queue of
// literal SSE bodies (or HTTP errors), GET .../threads/:id/ui from a queue of
// UI snapshots, and GET .../threads/:id/messages from a queue of thread
// histories (an empty history by default). The dashboard's real AG-UI client
// parses them, so what a test asserts is what a user would see for that wire
// traffic. SessionService holds the caller's `agui` sessions: a run on a
// thread without one creates it, bound to the run's workspace and agent, as
// the server does.

export const USER_ID = 'test-user-1'

export function sse(events: Array<Record<string, unknown>>): string {
  return events.map((ev) => `data: ${JSON.stringify(ev)}\n\n`).join('')
}

// A run answered with an HTTP error before any stream opens.
export interface RunRejection {
  status: number
  body: Record<string, unknown>
}

// A run whose stream starts only after a delay, to observe in-flight UI.
export interface DelayedRun {
  delayMs: number
  sse: string
}

export type RunResponse = string | RunRejection | DelayedRun

export interface SnapshotResponse {
  status?: number
  body?: Record<string, unknown>
}

// SessionCalls records what the page asked SessionService to do.
export interface SessionCalls {
  lists: Array<{ appName: string; workspaceScoped: boolean; pageToken: string }>
  gets: string[]
  renames: Array<{ sessionId: string; appName: string; title: string }>
  deletes: Array<{ sessionId: string; appName: string }>
  generated: string[]
}

export interface AGUIFixture {
  runs: RunResponse[]
  snapshots: SnapshotResponse[]
  histories: SnapshotResponse[]
  // historyByThread answers a thread's history by its ID, ahead of the queue.
  historyByThread?: Record<string, SnapshotResponse>
  requests: Array<Record<string, unknown>>
  // runURLs holds the URL of each run request, in the order of requests.
  runURLs: string[]
  snapshotRequests: string[]
  historyRequests: string[]
  // sessions are the caller's sessions, newest first.
  sessions: SessionInfo[]
  sessionCalls: SessionCalls
  // agentIcons are icon URLs by agent ID, set as those agents' icon_url.
  agentIcons?: Record<string, string>
}

export const emptySnapshot = (threadId = 't') => ({
  body: {
    version: 'v0.9.1',
    catalogId: 'butter-basic-v1',
    threadId,
    surfaces: [],
  },
})

export const emptyHistory = (threadId = 't') => ({
  body: { threadId, messages: [], interrupts: [], surfaces: [] },
})

// aguiSession is the session of the AG-UI thread threadId. binding is what
// the server records when it creates the session (null for a thread from
// before A2UI, which has none). updated is when it was last updated: a number
// of minutes ago, or a time.
export function aguiSession(
  threadId: string,
  title: string,
  binding: { agentId: string; workspaceId?: string } | null,
  updated: number | Date = 0
): SessionInfo {
  return create(SessionInfoSchema, {
    sessionId: `agui-${threadId}`,
    appName: 'agui',
    userId: USER_ID,
    title,
    state: binding
      ? {
          'butter:a2ui:binding': JSON.stringify({
            principal: USER_ID,
            workspace_id: binding.workspaceId ?? 'default',
            agent_id: binding.agentId,
            thread_id: threadId,
          }),
        }
      : {},
    lastUpdateTime: timestampFromDate(
      updated instanceof Date
        ? updated
        : new Date(Date.now() - updated * 60_000)
    ),
  })
}

// resumeEntries lists every resume entry the page sent, in request order.
export function resumeEntries(
  requests: Array<Record<string, unknown>>
): Array<Record<string, unknown>> {
  return requests.flatMap(
    (r) => (r.resume as Array<Record<string, unknown>> | undefined) ?? []
  )
}

// sendMessage types a message into whichever composer the page shows (the
// new-chat draft or the open thread's) and sends it with Enter.
export async function sendMessage(page: Page, text: string) {
  const composer = page.getByRole('textbox', { name: /^Message/ })
  await composer.fill(text)
  await composer.press('Enter')
}

// threadInURL is the thread the page's URL names.
export function threadInURL(page: Page): string | null {
  return new URL(page.url()).searchParams.get('thread')
}

export async function setupAGUI(
  page: Page,
  fixture: Partial<AGUIFixture> & { runs: RunResponse[] },
  options: ConnectFixtureOptions = {}
): Promise<AGUIFixture> {
  const state: AGUIFixture = {
    runs: fixture.runs,
    snapshots: fixture.snapshots ?? [],
    histories: fixture.histories ?? [],
    historyByThread: fixture.historyByThread,
    requests: fixture.requests ?? [],
    runURLs: fixture.runURLs ?? [],
    snapshotRequests: fixture.snapshotRequests ?? [],
    historyRequests: fixture.historyRequests ?? [],
    sessions: fixture.sessions ?? [],
    sessionCalls: {
      lists: [],
      gets: [],
      renames: [],
      deletes: [],
      generated: [],
    },
    agentIcons: fixture.agentIcons,
  }
  const withIcon = (agentId: string) => {
    const icon = state.agentIcons?.[agentId]
    return icon ? { metadata: { icon_url: icon } } : {}
  }

  await setupAuthenticatedConnectRoutes(
    page,
    async (route, url) => {
      if (url.includes('AgentService/ListAgents')) {
        return fulfillProto(route, ListAgentsResponseSchema, {
          agents: [
            {
              name: 'Streamer',
              agentId: 'streamer-id',
              description: 'AG-UI enabled',
              enableAgui: true,
              lifecycleStatus: 1,
            },
            {
              name: 'Second',
              agentId: 'second-id',
              description: 'another AG-UI agent',
              enableAgui: true,
              lifecycleStatus: 1,
            },
            // AG-UI Chat lists it too: enable_agui only gates API tokens.
            {
              name: 'Plain',
              agentId: 'plain-id',
              description: 'no programmatic AG-UI access',
              enableAgui: false,
              lifecycleStatus: 1,
            },
            // Deleted, so it cannot run and is not listed.
            {
              name: 'Retired',
              agentId: 'retired-id',
              description: 'deleted',
              enableAgui: true,
              lifecycleStatus: 6,
            },
          ].map((agent) => ({ ...agent, ...withIcon(agent.agentId) })),
          total: 4,
        })
      }
      if (url.includes('SessionService/')) {
        return answerSessions(route, url, state)
      }
      return false
    },
    options
  )

  await page.route('**/api/agui/**', async (route) => {
    const request = route.request()
    const path = new URL(request.url()).pathname
    if (request.method() === 'GET' && path.endsWith('/messages')) {
      state.historyRequests.push(request.url())
      const threadId = decodeURIComponent(path.split('/').at(-2) ?? '')
      const history: SnapshotResponse =
        state.historyByThread?.[threadId] ??
        state.histories.shift() ??
        emptyHistory(threadId)
      await route.fulfill({
        status: history.status ?? 200,
        contentType: 'application/json',
        body: JSON.stringify(history.body ?? { error: 'busy' }),
      })
      return
    }
    if (request.method() === 'GET') {
      state.snapshotRequests.push(request.url())
      const snap: SnapshotResponse = state.snapshots.shift() ?? emptySnapshot()
      await route.fulfill({
        status: snap.status ?? 200,
        contentType: 'application/json',
        body: JSON.stringify(snap.body ?? { error: 'busy' }),
      })
      return
    }
    const input = JSON.parse(request.postData() ?? '{}')
    state.requests.push(input)
    state.runURLs.push(request.url())
    const next = state.runs.shift() ?? sse([])
    if (typeof next === 'string' || 'delayMs' in next) {
      // The server creates a new thread's session before the stream opens.
      const agentId = decodeURIComponent(path.split('/').at(-1) ?? '')
      const workspaceId =
        (await request.headerValue('x-workspace-id')) ?? 'default'
      const threadId = String(input.threadId ?? '')
      if (!state.sessions.some((s) => s.sessionId === `agui-${threadId}`)) {
        state.sessions.unshift(
          aguiSession(threadId, '', { agentId, workspaceId })
        )
      }
    }
    if (typeof next !== 'string' && 'delayMs' in next) {
      await new Promise((resolve) => setTimeout(resolve, next.delayMs))
      await route.fulfill({
        status: 200,
        contentType: 'text/event-stream',
        body: next.sse,
      })
      return
    }
    if (typeof next === 'string') {
      await route.fulfill({
        status: 200,
        contentType: 'text/event-stream',
        body: next,
      })
      return
    }
    await route.fulfill({
      status: next.status,
      contentType: 'application/json',
      body: JSON.stringify(next.body),
    })
  })
  return state
}

// answerSessions answers SessionService from state.sessions the way the
// server does for the caller's own sessions.
async function answerSessions(
  route: Route,
  url: string,
  state: AGUIFixture
): Promise<boolean> {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  const { sessions, sessionCalls: calls } = state
  if (url.endsWith('/ListSessions')) {
    const req = fromBinary(ListSessionsRequestSchema, body)
    calls.lists.push({
      appName: req.appName,
      workspaceScoped: req.workspaceScoped,
      pageToken: req.pageToken,
    })
    // Pages like the server: an offset cursor, empty on the last page.
    const matching = sessions.filter((s) => s.appName === req.appName)
    const offset = Number(req.pageToken || '0')
    const end = req.pageSize > 0 ? offset + req.pageSize : matching.length
    return fulfillProto(route, ListSessionsResponseSchema, {
      sessions: matching.slice(offset, end),
      nextPageToken: end < matching.length ? String(end) : '',
    })
  }
  if (url.endsWith('/GetSession')) {
    const req = fromBinary(GetSessionRequestSchema, body)
    calls.gets.push(req.sessionId)
    const session = sessions.find(
      (s) =>
        s.appName === req.appName &&
        s.userId === req.userId &&
        s.sessionId === req.sessionId
    )
    if (!session) {
      return fulfillConnectError(route, 'not_found', 'session not found')
    }
    return fulfillProto(route, GetSessionResponseSchema, {
      sessionDetail: { session, events: [] },
    })
  }
  if (url.endsWith('/UpdateSessionTitle')) {
    const req = fromBinary(UpdateSessionTitleRequestSchema, body)
    calls.renames.push({
      sessionId: req.sessionId,
      appName: req.appName,
      title: req.title,
    })
    const session = sessions.find((s) => s.sessionId === req.sessionId)
    if (!session) {
      return fulfillConnectError(route, 'not_found', 'session not found')
    }
    session.title = req.title
    return fulfillProto(route, UpdateSessionTitleResponseSchema, { session })
  }
  if (url.endsWith('/DeleteSession')) {
    const req = fromBinary(DeleteSessionRequestSchema, body)
    calls.deletes.push({ sessionId: req.sessionId, appName: req.appName })
    const at = sessions.findIndex((s) => s.sessionId === req.sessionId)
    if (at >= 0) sessions.splice(at, 1)
    return fulfillProto(route, DeleteSessionResponseSchema, {})
  }
  if (url.endsWith('/GenerateSessionTitle')) {
    const req = fromBinary(GenerateSessionTitleRequestSchema, body)
    calls.generated.push(req.sessionId)
    return fulfillProto(route, GenerateSessionTitleResponseSchema, {
      session: { sessionId: req.sessionId, appName: req.appName },
      generated: false,
    })
  }
  return false
}
