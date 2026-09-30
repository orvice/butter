import type { Page } from '@playwright/test'
import { ListAgentsResponseSchema } from '../../src/gen/agents/v1/agent_service_pb'
import {
  fulfillProto,
  setupAuthenticatedConnectRoutes,
  type ConnectFixtureOptions,
} from './connect'

// AG-UI fixtures: POST /api/agui/:agent_id is answered from a queue of
// literal SSE bodies (or HTTP errors), and GET .../threads/:id/ui from a
// queue of UI snapshots. The dashboard's real AG-UI client parses them, so
// what a test asserts is what a user would see for that wire traffic.

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
  requests: Array<Record<string, unknown>>
  snapshotRequests: string[]
}

export const emptySnapshot = (threadId = 't') => ({
  body: {
    version: 'v0.9.1',
    catalogId: 'butter-basic-v1',
    threadId,
    surfaces: [],
  },
})

export async function setupAGUI(
  page: Page,
  fixture: Partial<AGUIFixture> & { runs: RunResponse[] },
  options: ConnectFixtureOptions = {}
): Promise<AGUIFixture> {
  const state: AGUIFixture = {
    runs: fixture.runs,
    snapshots: fixture.snapshots ?? [],
    requests: fixture.requests ?? [],
    snapshotRequests: fixture.snapshotRequests ?? [],
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
