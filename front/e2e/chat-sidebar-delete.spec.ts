import { expect, test, type Page } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import {
  DeleteSessionRequestSchema,
  DeleteSessionResponseSchema,
  GetSessionRequestSchema,
  GetSessionResponseSchema,
  ListSessionsRequestSchema,
  ListSessionsResponseSchema,
  SessionInfoSchema,
  type SessionInfo,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

function webChat(id: string, title: string, minutesAgo: number): SessionInfo {
  return create(SessionInfoSchema, {
    sessionId: id,
    appName: 'web-chat',
    userId: 'test-user-1',
    workspaceId: 'default',
    title,
    state: { agent_name: 'Helper', agent_id: 'helper', workspace_id: 'default' },
    lastUpdateTime: timestampFromDate(
      new Date(Date.now() - minutesAgo * 60_000)
    ),
  })
}

async function setupChats(page: Page, sessions: SessionInfo[]) {
  const deleted: string[] = []
  await page.addInitScript(() => {
    localStorage.setItem('butter_workspace_id', 'default')
  })
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
    if (url.includes('SessionService/ListSessions')) {
      const req = fromBinary(ListSessionsRequestSchema, body)
      return fulfillProto(route, ListSessionsResponseSchema, {
        sessions: sessions.filter((s) => s.appName === req.appName),
      })
    }
    if (url.includes('SessionService/GetSession')) {
      const req = fromBinary(GetSessionRequestSchema, body)
      const session = sessions.find((s) => s.sessionId === req.sessionId)
      if (!session) return fulfillConnectError(route, 'not_found')
      return fulfillProto(route, GetSessionResponseSchema, {
        sessionDetail: { session, events: [] },
      })
    }
    if (url.includes('SessionService/DeleteSession')) {
      const req = fromBinary(DeleteSessionRequestSchema, body)
      deleted.push(req.sessionId)
      sessions.splice(
        sessions.findIndex((s) => s.sessionId === req.sessionId),
        1
      )
      return fulfillProto(route, DeleteSessionResponseSchema, {})
    }
    return false
  })
  return deleted
}

async function deleteFromSidebar(page: Page, title: string) {
  const row = page
    .locator('[data-sidebar="menu-item"]')
    .filter({ has: page.getByRole('link', { name: title }) })
  await row.hover()
  await row.getByRole('button', { name: 'Chat actions' }).click()
  await page.getByRole('menuitem', { name: 'Delete' }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click()
}

test('deletes chats from the sidebar, leaving the open one only when it is deleted', async ({
  page,
}) => {
  const deleted = await setupChats(page, [
    webChat('chat-1', 'First chat', 1),
    webChat('chat-2', 'Second chat', 2),
  ])

  await page.goto('/chat?session=chat-1', { waitUntil: 'networkidle' })
  const sidebar = page.locator('[data-sidebar="sidebar"]')
  await expect(sidebar.getByRole('link', { name: 'Second chat' })).toBeVisible()

  // Deleting another chat keeps the open one.
  await deleteFromSidebar(page, 'Second chat')
  await expect(sidebar.getByRole('link', { name: 'Second chat' })).toHaveCount(0)
  expect(deleted).toEqual(['chat-2'])
  await expect(page).toHaveURL(/session=chat-1/)

  // Deleting the open chat leaves it for a new chat.
  await deleteFromSidebar(page, 'First chat')
  await expect(sidebar.getByRole('link', { name: 'First chat' })).toHaveCount(0)
  expect(deleted).toEqual(['chat-2', 'chat-1'])
  await expect(page).not.toHaveURL(/session=/)
})
