import { expect, test, type Page } from '@playwright/test'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import { AgentLifecycleStatus } from '../src/gen/agents/v1/agent_pb'
import {
  GetAgentInvocationResponseSchema,
  GetSessionResponseSchema,
  InvocationStatus,
  ListAgentsResponseSchema,
  ListSessionsResponseSchema,
  SubmitAgentInvocationResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

async function setupNewPiChat(page: Page) {
  let accepted = false
  let detailCalls = 0
  const session = {
    sessionId: 'chat-pi-1',
    appName: 'web-chat',
    userId: 'test-user-1',
    workspaceId: 'default',
    state: {
      agent_name: 'PiAgent',
      agent_id: 'pi-agent',
      workspace_id: 'default',
    },
    lastUpdateTime: timestampFromDate(new Date()),
  }

  await page.addInitScript(() => {
    localStorage.setItem('butter_workspace_id', 'default')
  })
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [
          {
            name: 'PiAgent',
            agentId: 'pi-agent',
            lifecycleStatus: AgentLifecycleStatus.ACTIVE,
          },
        ],
        total: 1,
      })
    }
    if (url.includes('AgentService/SubmitAgentInvocation')) {
      accepted = true
      return fulfillProto(route, SubmitAgentInvocationResponseSchema, {
        sessionId: session.sessionId,
        invocationId: 'inv-pi-1',
        status: InvocationStatus.QUEUED,
        sessionCreated: true,
      })
    }
    if (url.includes('SessionService/ListSessions')) {
      // Reproduce the list/detail race: the accepted session is not in the
      // recent-session snapshot yet.
      return fulfillProto(route, ListSessionsResponseSchema, {
        sessions: [],
        total: 0,
      })
    }
    if (url.includes('SessionService/GetSession')) {
      detailCalls++
      return fulfillProto(route, GetSessionResponseSchema, {
        sessionDetail: { session, events: [] },
      })
    }
    if (url.includes('AgentService/GetAgentInvocation')) {
      return fulfillProto(route, GetAgentInvocationResponseSchema, {
        invocation: {
          id: 'inv-pi-1',
          sessionId: session.sessionId,
          userId: session.userId,
          workspaceId: session.workspaceId,
          status: InvocationStatus.SUCCEEDED,
        },
      })
    }
    return false
  })

  return {
    wasAccepted: () => accepted,
    detailCalls: () => detailCalls,
  }
}

test('an accepted Pi chat opens by ID when the recent-session list is stale', async ({
  page,
}) => {
  const state = await setupNewPiChat(page)

  await page.goto('/chat?agent=pi-agent')
  const draft = page.getByTestId('draft-composer-input')
  await expect(draft).toBeEnabled()
  await draft.fill('hello pi')
  await page.getByTestId('draft-composer-send').click()

  await expect(page).toHaveURL(/session=chat-pi-1/)
  expect(state.wasAccepted()).toBe(true)
  await expect.poll(state.detailCalls).toBeGreaterThan(0)
  await expect(page.getByText('Session not found.', { exact: false })).toHaveCount(
    0
  )
  await expect(page.getByPlaceholder('Message PiAgent...')).toBeVisible()
})
