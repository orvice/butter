import { expect, test, type Page } from '@playwright/test'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import {
  GetAgentInvocationResponseSchema,
  GetSessionResponseSchema,
  InvocationStatus,
  ListAgentsResponseSchema,
  ListSessionsResponseSchema,
  WatchAgentInvocationResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  fulfillConnectStream,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const SESSION_ID = 'sess-watch-1'
const INVOCATION_ID = 'inv-watch-1'

const sessionInfo = {
  sessionId: SESSION_ID,
  appName: 'web-chat',
  userId: 'test-user-1',
  title: 'Long-running chat',
  state: { agent_name: 'ChatBot', agent_id: 'chatbot-id' },
  lastUpdateTime: timestampFromDate(new Date()),
  turnCount: 1,
}

function userEvent(text: string) {
  return {
    eventId: 'evt-user-1',
    invocationId: INVOCATION_ID,
    author: 'user',
    contentJson: JSON.stringify({ role: 'user', parts: [{ text }] }),
    timestamp: timestampFromDate(new Date(Date.now() - 60_000)),
  }
}

function assistantEvent(text: string) {
  return {
    eventId: 'evt-assistant-1',
    invocationId: INVOCATION_ID,
    author: 'ChatBot',
    contentJson: JSON.stringify({ role: 'model', parts: [{ text }] }),
    timestamp: timestampFromDate(new Date()),
  }
}

interface WatchMockState {
  watchCalls: number
  cancelCalls: number
  invocationTerminal: boolean
}

// Mocks the chat backend around one session that has a RUNNING invocation.
// The first WatchAgentInvocation attach stalls (simulating a still-running
// stream the user navigates away from); later attaches deliver live deltas
// and the terminal state. Persisted GetSession includes the assistant turn
// only once the invocation is terminal, mirroring the real backend.
async function setupRunningChat(page: Page): Promise<WatchMockState> {
  const state: WatchMockState = {
    watchCalls: 0,
    cancelCalls: 0,
    invocationTerminal: false,
  }

  await page.addInitScript(() => {
    localStorage.setItem('butter_workspace_id', 'default')
  })

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [
          {
            name: 'ChatBot',
            agentId: 'chatbot-id',
            description: 'General assistant',
            lifecycleStatus: 1, // ACTIVE
          },
        ],
        total: 1,
      })
    }

    if (url.includes('SessionService/ListSessions')) {
      return fulfillProto(route, ListSessionsResponseSchema, {
        sessions: [sessionInfo],
        total: 1,
      })
    }

    if (url.includes('SessionService/GetSession')) {
      return fulfillProto(route, GetSessionResponseSchema, {
        sessionDetail: {
          session: sessionInfo,
          events: state.invocationTerminal
            ? [userEvent('Summarize the report'), assistantEvent('Hello world, the report is done.')]
            : [userEvent('Summarize the report')],
        },
      })
    }

    if (url.includes('AgentService/GetAgentInvocation')) {
      return fulfillProto(route, GetAgentInvocationResponseSchema, {
        invocation: {
          id: INVOCATION_ID,
          sessionId: SESSION_ID,
          userId: 'test-user-1',
          workspaceId: 'default',
          source: 'dashboard-async',
          status: state.invocationTerminal
            ? InvocationStatus.SUCCEEDED
            : InvocationStatus.RUNNING,
          output: state.invocationTerminal ? 'Hello world, the report is done.' : '',
        },
      })
    }

    if (url.includes('AgentService/CancelAgentInvocation')) {
      state.cancelCalls++
      return false // fall through to the generic empty fulfillment
    }

    if (url.includes('AgentService/WatchAgentInvocation')) {
      state.watchCalls++
      if (state.watchCalls === 1) {
        // First observer: the run is mid-flight and produces nothing yet.
        // Hold the stream open; navigating away simply aborts this request
        // without cancelling the run.
        await new Promise((resolve) => setTimeout(resolve, 30_000))
        try {
          await fulfillConnectStream(route, WatchAgentInvocationResponseSchema, [])
        } catch {
          // Observer already detached (request aborted by navigation).
        }
        return true
      }
      // Re-attached observer: authoritative state first, then live deltas,
      // then exactly one terminal state.
      state.invocationTerminal = true
      return fulfillConnectStream(route, WatchAgentInvocationResponseSchema, [
        {
          event: {
            case: 'state',
            value: {
              invocation: {
                id: INVOCATION_ID,
                sessionId: SESSION_ID,
                status: InvocationStatus.RUNNING,
              },
            },
          },
        },
        {
          event: {
            case: 'textDelta',
            value: {
              invocationId: INVOCATION_ID,
              sessionId: SESSION_ID,
              agentName: 'ChatBot',
              text: 'Hello world, ',
            },
          },
        },
        {
          event: {
            case: 'textDelta',
            value: {
              invocationId: INVOCATION_ID,
              sessionId: SESSION_ID,
              agentName: 'ChatBot',
              text: 'the report is done.',
            },
          },
        },
        {
          event: {
            case: 'state',
            value: {
              invocation: {
                id: INVOCATION_ID,
                sessionId: SESSION_ID,
                status: InvocationStatus.SUCCEEDED,
                output: 'Hello world, the report is done.',
              },
            },
          },
        },
      ])
    }

    return false
  })

  return state
}

test.describe('Chat watch stream', () => {
  test('navigating away from a running chat never cancels it, and returning observes live output', async ({
    page,
  }) => {
    const state = await setupRunningChat(page)

    // Open the session with a run in flight: the client loads persisted
    // events and the invocation status, then attaches a watch stream.
    await page.goto(`/chat?session=${SESSION_ID}`)
    await expect(page.getByText('Summarize the report')).toBeVisible()
    await expect(page.getByText('Thinking…')).toBeVisible()
    await expect.poll(() => state.watchCalls).toBe(1)

    // Navigate away while the run is still active.
    await page.goto('/chat')
    await expect(page.getByText('Thinking…')).not.toBeVisible()

    // Observation is read-only: leaving must not cancel the invocation.
    expect(state.cancelCalls).toBe(0)

    // Return to the running chat: persisted state loads first, then a fresh
    // observer receives the live deltas and the terminal state.
    await page.goto(`/chat?session=${SESSION_ID}`)
    await expect(page.getByText('Hello world, the report is done.')).toBeVisible()
    expect(state.watchCalls).toBeGreaterThanOrEqual(2)

    // The turn completes and the composer returns to idle.
    await expect(page.getByText('Thinking…')).not.toBeVisible()
    expect(state.cancelCalls).toBe(0)
  })

  test('an idle session attaches no observer stream', async ({ page }) => {
    const state = await setupRunningChat(page)
    state.invocationTerminal = true // lookup reports SUCCEEDED — nothing to watch

    await page.goto(`/chat?session=${SESSION_ID}`)
    await expect(page.getByText('Summarize the report')).toBeVisible()
    // Give the reconnect lookup time to complete before asserting no attach.
    await page.waitForTimeout(750)
    await expect(page.getByText('Thinking…')).not.toBeVisible()
    expect(state.watchCalls).toBe(0)
  })
})
