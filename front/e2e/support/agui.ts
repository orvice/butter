import type { Page } from '@playwright/test'
import { ListAgentsResponseSchema } from '../../src/gen/agents/v1/agent_service_pb'
import {
  fulfillProto,
  setupAuthenticatedConnectRoutes,
  type ConnectFixtureOptions,
} from './connect'

// AG-UI fixtures: POST /api/agui/:agent_id is answered from a queue of
// literal SSE bodies (or HTTP errors), GET .../threads/:id/ui from a queue of
// UI snapshots, and GET .../threads/:id/messages from a queue of thread
// histories (an empty history by default). The dashboard's real AG-UI client
// parses them, so what a test asserts is what a user would see for that wire
// traffic.

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

export interface AGUIFixture {
  runs: RunResponse[]
  snapshots: SnapshotResponse[]
  histories: SnapshotResponse[]
  // historyByThread answers a thread's history by its ID, ahead of the queue.
  historyByThread?: Record<string, SnapshotResponse>
  requests: Array<Record<string, unknown>>
  snapshotRequests: string[]
  historyRequests: string[]
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

// resumeEntries lists every resume entry the page sent, in request order.
export function resumeEntries(
  requests: Array<Record<string, unknown>>
): Array<Record<string, unknown>> {
  return requests.flatMap(
    (r) => (r.resume as Array<Record<string, unknown>> | undefined) ?? []
  )
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
    snapshotRequests: fixture.snapshotRequests ?? [],
    historyRequests: fixture.historyRequests ?? [],
  }
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
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
          {
            name: 'Plain',
            agentId: 'plain-id',
            description: 'not exposed',
            enableAgui: false,
            lifecycleStatus: 1,
          },
        ],
        total: 3,
      })
    }
    return false
  }, options)

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
    state.requests.push(JSON.parse(request.postData() ?? '{}'))
    const next = state.runs.shift() ?? sse([])
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
