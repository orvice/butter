import { fromBinary } from '@bufbuild/protobuf'
import { expect, test, type Page } from '@playwright/test'
import { AgentLifecycleStatus } from '../src/gen/agents/v1/agent_pb'
import {
  GetAgentInvocationRequestSchema,
  GetAgentInvocationResponseSchema,
  GetSessionRequestSchema,
  GetSessionResponseSchema,
  InvocationStatus,
  ListAgentsResponseSchema,
  ListSessionsResponseSchema,
  SubmitAgentInvocationRequestSchema,
  SubmitAgentInvocationResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

// session-a's last turn FAILED (process restart); session-b's was stopped.
const sessions = ['session-a', 'session-b'].map((sessionId) => ({
  sessionId,
  appName: 'web-chat',
  userId: 'test-user-1',
  workspaceId: 'default',
  title: sessionId === 'session-a' ? 'Failed chat' : 'Stopped chat',
  state: {
    agent_name: 'ChatBot',
    agent_id: 'chatbot-id',
    workspace_id: 'default',
  },
}))

const RESTART_FAILURE =
  'interrupted by a service restart before it could finish; no work was replayed automatically. Review your message and resubmit to retry.'

// 1×1 transparent PNG — a valid restored image attachment.
const PNG_BYTES = new Uint8Array(
  Buffer.from(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
    'base64'
  )
)

interface SubmittedRequest {
  requestId: string
  sessionId: string
  message: string
  partsCount: number
}

interface MockState {
  submits: SubmittedRequest[]
}

function invocationFor(id: string, sessionId: string, status: InvocationStatus, error = '') {
  return {
    id,
    appName: 'web-chat',
    userId: 'test-user-1',
    sessionId,
    workspaceId: 'default',
    source: 'dashboard-async',
    status,
    input: 'original message',
    error,
  }
}

async function setupChat(page: Page): Promise<MockState> {
  const state: MockState = { submits: [] }
  const latestBySession = new Map([
    ['session-a', invocationFor('inv-failed', 'session-a', InvocationStatus.FAILED, RESTART_FAILURE)],
    ['session-b', invocationFor('inv-stopped', 'session-b', InvocationStatus.CANCELLED)],
  ])

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
      state.submits.push({
        requestId: request.requestId,
        sessionId: request.sessionId,
        message: request.message,
        partsCount: request.parts.length,
      })
      const invocationId = `inv-new-${state.submits.length}`
      // The resubmitted run completes immediately so the observer resolves
      // without a watch stream.
      latestBySession.set(
        request.sessionId,
        invocationFor(invocationId, request.sessionId, InvocationStatus.SUCCEEDED)
      )
      return fulfillProto(route, SubmitAgentInvocationResponseSchema, {
        sessionId: request.sessionId,
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
      if (request.invocationId) {
        const known = [...latestBySession.values()].find((inv) => inv.id === request.invocationId)
        if (!known) {
          return fulfillConnectError(route, 'not_found', 'invocation not found')
        }
        return fulfillProto(route, GetAgentInvocationResponseSchema, {
          invocation: known,
          inputParts: request.includeInputParts
            ? [
                { part: { case: 'text', value: 'original message' } },
                { part: { case: 'inlineData', value: { mimeType: 'image/png', data: PNG_BYTES } } },
              ]
            : [],
        })
      }
      if (request.latest) {
        const latest = latestBySession.get(request.sessionId)
        if (!latest) {
          return fulfillConnectError(route, 'not_found', 'invocation not found')
        }
        return fulfillProto(route, GetAgentInvocationResponseSchema, { invocation: latest })
      }
      // Active-invocation lookup: nothing is running in these fixtures.
      return fulfillConnectError(route, 'not_found', 'no active invocation')
    }
    return false
  })

  return state
}

test('a failed invocation renders inline with the turn after reload', async ({ page }) => {
  await setupChat(page)
  await page.goto('/chat?session=session-a')

  const notice = page.getByRole('alert')
  await expect(notice).toBeVisible()
  await expect(notice).toContainText('This run failed')
  await expect(notice).toContainText('interrupted by a service restart')
  await expect(notice).toContainText('may repeat external tool actions')
  // The turn never persisted as a session event (orphaned while QUEUED), so
  // the notice itself shows the submitted message.
  await expect(notice).toContainText('original message')
  await expect(notice.getByRole('button', { name: 'Restore input' })).toBeVisible()
})

test('a cancelled invocation renders as stopped, not as a failure', async ({ page }) => {
  await setupChat(page)
  await page.goto('/chat?session=session-b')

  const notice = page.getByRole('status').filter({ hasText: 'Stopped' })
  await expect(notice).toBeVisible()
  await expect(notice).toContainText('You stopped this response')
  await expect(page.getByText('This run failed')).toBeHidden()
  await expect(notice.getByRole('button', { name: 'Restore input' })).toBeVisible()
})

test('restore repopulates text and image input; resubmission is a new invocation with a fresh request ID', async ({ page }) => {
  const state = await setupChat(page)
  await page.goto('/chat?session=session-a')

  await page.getByRole('button', { name: 'Restore input' }).click()

  const composer = page.getByPlaceholder('Message ChatBot...')
  await expect(composer).toHaveValue('original message')
  await expect(page.getByRole('img', { name: 'restored-2.png' })).toBeVisible()

  await page.getByRole('button', { name: 'Send message' }).click()

  // Resubmission created a new invocation; the stale failure notice is gone.
  await expect(page.getByRole('alert')).toBeHidden()
  await expect.poll(() => state.submits.length).toBe(1)
  expect(state.submits[0].sessionId).toBe('session-a')
  expect(state.submits[0].requestId).not.toBe('')
  expect(state.submits[0].message).toBe('original message')
  expect(state.submits[0].partsCount).toBe(2)

  // A second explicit send is another new invocation under another request ID.
  await composer.fill('try once more')
  await page.getByRole('button', { name: 'Send message' }).click()
  await expect.poll(() => state.submits.length).toBe(2)
  expect(state.submits[1].requestId).not.toBe('')
  expect(state.submits[1].requestId).not.toBe(state.submits[0].requestId)
})
