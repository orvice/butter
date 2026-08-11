import { fromBinary } from '@bufbuild/protobuf'
import { expect, test, type Page } from '@playwright/test'
import { AgentLifecycleStatus } from '../src/gen/agents/v1/agent_pb'
import {
  CancelAgentInvocationRequestSchema,
  CancelAgentInvocationResponseSchema,
  GetAgentInvocationRequestSchema,
  GetAgentInvocationResponseSchema,
  GetSessionRequestSchema,
  GetSessionResponseSchema,
  InvocationStatus,
  ListAgentsResponseSchema,
  ListSessionsResponseSchema,
  SubmitAgentInvocationRequestSchema,
  SubmitAgentInvocationResponseSchema,
  WatchAgentInvocationRequestSchema,
  WatchAgentInvocationResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  decodeConnectStreamRequest,
  fulfillConnectStream,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const sessions = ['session-a', 'session-b'].map((sessionId) => ({
  sessionId,
  appName: 'web-chat',
  userId: 'test-user-1',
  workspaceId: 'default',
  title: sessionId === 'session-a' ? 'Chat A' : 'Chat B',
  state: {
    agent_name: 'ChatBot',
    agent_id: 'chatbot-id',
    workspace_id: 'default',
  },
}))

interface MockState {
  cancellations: string[]
  statuses: Map<string, number>
  submittedBySession: Map<string, string>
}

async function setupChat(page: Page): Promise<MockState> {
  const state: MockState = {
    cancellations: [],
    statuses: new Map(),
    submittedBySession: new Map(),
  }

  await page.addInitScript(() => {
    localStorage.setItem('butter_workspace_id', 'default')
  })

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [{
          name: 'ChatBot',
          agentId: 'chatbot-id',
          lifecycleStatus: AgentLifecycleStatus.ACTIVE,
        }],
        total: 1,
      })
    }
    if (url.includes('SessionService/ListSessions')) {
      return fulfillProto(route, ListSessionsResponseSchema, {
        sessions,
        total: sessions.length,
      })
    }
    if (url.includes('SessionService/GetSession')) {
      const request = fromBinary(
        GetSessionRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      const selected = sessions.find(
        (session) => session.sessionId === request.sessionId
      )
      return fulfillProto(route, GetSessionResponseSchema, {
        sessionDetail: { session: selected, events: [] },
      })
    }
    if (url.includes('AgentService/SubmitAgentInvocation')) {
      const request = fromBinary(
        SubmitAgentInvocationRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      const sessionId = request.sessionId
      const invocationId = sessionId === 'session-a' ? 'inv-a' : 'inv-b'
      state.submittedBySession.set(sessionId, invocationId)
      state.statuses.set(invocationId, InvocationStatus.RUNNING)
      return fulfillProto(route, SubmitAgentInvocationResponseSchema, {
        sessionId,
        invocationId,
        status: InvocationStatus.QUEUED,
        sessionCreated: false,
      })
    }
    if (url.includes('AgentService/GetAgentInvocation')) {
      const request = fromBinary(
        GetAgentInvocationRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      const invocationId = request.invocationId
      return fulfillProto(route, GetAgentInvocationResponseSchema, {
        invocation: {
          id: invocationId,
          appName: 'web-chat',
          userId: 'test-user-1',
          sessionId: invocationId === 'inv-a' ? 'session-a' : 'session-b',
          workspaceId: 'default',
          source: 'dashboard-async',
          status: state.statuses.get(invocationId) ?? InvocationStatus.RUNNING,
        },
      })
    }
    if (url.includes('AgentService/WatchAgentInvocation')) {
      const request = decodeConnectStreamRequest(
        WatchAgentInvocationRequestSchema,
        route.request().postDataBuffer()
      )
      const invocationId = request.invocationId
      // Hold the observer stream open while the run is active, then deliver
      // the single terminal state frame.
      const deadline = Date.now() + 15_000
      while (
        (state.statuses.get(invocationId) ?? InvocationStatus.RUNNING) ===
          InvocationStatus.RUNNING &&
        Date.now() < deadline
      ) {
        await new Promise((resolve) => setTimeout(resolve, 100))
      }
      try {
        return await fulfillConnectStream(route, WatchAgentInvocationResponseSchema, [
          {
            event: {
              case: 'state',
              value: {
                invocation: {
                  id: invocationId,
                  appName: 'web-chat',
                  userId: 'test-user-1',
                  sessionId: invocationId === 'inv-a' ? 'session-a' : 'session-b',
                  workspaceId: 'default',
                  source: 'dashboard-async',
                  status: state.statuses.get(invocationId) ?? InvocationStatus.RUNNING,
                },
              },
            },
          },
        ])
      } catch {
        return true // observer detached (request aborted by navigation)
      }
    }
    if (url.includes('AgentService/CancelAgentInvocation')) {
      const request = fromBinary(
        CancelAgentInvocationRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      const invocationId = request.invocationId
      state.cancellations.push(invocationId)
      state.statuses.set(invocationId, InvocationStatus.CANCELLED)
      return fulfillProto(route, CancelAgentInvocationResponseSchema, {
        cancelled: true,
      })
    }
    return false
  })

  return state
}

async function sendMessage(page: Page, message: string) {
  const composer = page.getByPlaceholder('Message ChatBot...')
  await composer.fill(message)
  await page.getByRole('button', { name: 'Send message' }).click()
  await expect(page.getByRole('button', { name: 'Stop generating' })).toBeVisible()
}

test('switching chats leaves the active Invocation running', async ({ page }) => {
  const state = await setupChat(page)
  await page.goto('/chat?session=session-a')
  await sendMessage(page, 'keep working')

  await page.goto('/chat?session=session-b')

  expect(state.submittedBySession.get('session-a')).toBe('inv-a')
  expect(state.statuses.get('inv-a')).toBe(InvocationStatus.RUNNING)
  expect(state.cancellations).toEqual([])
})

test('Stop cancels only the Invocation selected in the current chat', async ({ page }) => {
  const state = await setupChat(page)
  await page.goto('/chat?session=session-a')
  await sendMessage(page, 'run A')

  await page.goto('/chat?session=session-b')
  await sendMessage(page, 'run B')
  await page.getByRole('button', { name: 'Stop generating' }).click()

  await expect.poll(() => state.cancellations).toEqual(['inv-b'])
  await expect(page.getByRole('button', { name: 'Stop generating' })).toBeHidden()
  expect(state.statuses.get('inv-a')).toBe(InvocationStatus.RUNNING)
  expect(state.statuses.get('inv-b')).toBe(InvocationStatus.CANCELLED)
})
