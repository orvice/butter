import { expect, test, type Page } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import {
  DeleteSessionRequestSchema,
  DeleteSessionResponseSchema,
  GenerateSessionTitleRequestSchema,
  GenerateSessionTitleResponseSchema,
  ListSessionsRequestSchema,
  ListSessionsResponseSchema,
  SessionInfoSchema,
  UpdateSessionTitleRequestSchema,
  UpdateSessionTitleResponseSchema,
  type SessionInfo,
} from '../src/gen/agents/v1/agent_service_pb'
import { fulfillProto } from './support/connect'
import { setupAGUI, sse } from './support/agui'

// The AG-UI thread list reads the caller's `agui` sessions and keeps those
// whose A2UI binding names the selected agent in the selected workspace.

function aguiThread(
  threadId: string,
  title: string,
  binding: { agentId: string; workspaceId?: string } | null,
  minutesAgo: number
): SessionInfo {
  return create(SessionInfoSchema, {
    sessionId: `agui-${threadId}`,
    appName: 'agui',
    userId: 'test-user-1',
    title,
    state: binding
      ? {
          'butter:a2ui:binding': JSON.stringify({
            principal: 'test-user-1',
            workspace_id: binding.workspaceId ?? 'default',
            agent_id: binding.agentId,
            thread_id: threadId,
          }),
        }
      : {},
    lastUpdateTime: timestampFromDate(
      new Date(Date.now() - minutesAgo * 60_000)
    ),
  })
}

interface SessionCalls {
  renames: Array<{ sessionId: string; appName: string; title: string }>
  deletes: Array<{ sessionId: string; appName: string }>
  generated: string[]
}

async function setupThreads(page: Page, sessions: SessionInfo[]) {
  const calls: SessionCalls = { renames: [], deletes: [], generated: [] }
  // Registered after setupAGUI's catch-all, so it answers SessionService.
  await page.route('**/api/agents.v1.SessionService/**', async (route) => {
    const url = route.request().url()
    const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
    if (url.endsWith('/ListSessions')) {
      const req = fromBinary(ListSessionsRequestSchema, body)
      return fulfillProto(route, ListSessionsResponseSchema, {
        sessions: sessions.filter((s) => s.appName === req.appName),
      })
    }
    if (url.endsWith('/UpdateSessionTitle')) {
      const req = fromBinary(UpdateSessionTitleRequestSchema, body)
      calls.renames.push({
        sessionId: req.sessionId,
        appName: req.appName,
        title: req.title,
      })
      const s = sessions.find((x) => x.sessionId === req.sessionId)!
      s.title = req.title
      return fulfillProto(route, UpdateSessionTitleResponseSchema, {
        session: s,
      })
    }
    if (url.endsWith('/DeleteSession')) {
      const req = fromBinary(DeleteSessionRequestSchema, body)
      calls.deletes.push({ sessionId: req.sessionId, appName: req.appName })
      sessions.splice(
        sessions.findIndex((x) => x.sessionId === req.sessionId),
        1
      )
      return fulfillProto(route, DeleteSessionResponseSchema, {})
    }
    if (url.endsWith('/GenerateSessionTitle')) {
      const req = fromBinary(GenerateSessionTitleRequestSchema, body)
      calls.generated.push(req.sessionId)
      return fulfillProto(route, GenerateSessionTitleResponseSchema, {
        session: { sessionId: req.sessionId, appName: 'agui' },
        generated: false,
      })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/proto',
      body: Buffer.alloc(0),
    })
  })
  return calls
}

const reply = (text: string) =>
  sse([
    { type: 'RUN_STARTED', threadId: 't', runId: 'r' },
    { type: 'TEXT_MESSAGE_START', messageId: 'a', role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a', delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: 'a' },
    { type: 'RUN_FINISHED', threadId: 't', runId: 'r' },
  ])

async function send(page: Page, text: string) {
  const composer = page.getByPlaceholder(/Message the agent over AG-UI/)
  await composer.fill(text)
  await composer.press('Enter')
}

function threadActions(page: Page, title: string) {
  const row = page
    .getByRole('complementary', { name: 'Threads' })
    .getByRole('listitem')
    .filter({ hasText: title })
  return row.getByRole('button', { name: 'Thread actions' })
}

test.describe('AG-UI threads', () => {
  test('lists only this agent’s threads and switches between them', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Back on the trip.')],
    })
    await setupThreads(page, [
      aguiThread('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
      aguiThread('t-budget', 'Budget', { agentId: 'streamer-id' }, 5),
      aguiThread('t-second', 'Second agent chat', { agentId: 'second-id' }, 2),
      aguiThread(
        't-ws2',
        'Other workspace',
        { agentId: 'streamer-id', workspaceId: 'ws-2' },
        3
      ),
      aguiThread('t-legacy', 'Before A2UI', null, 4),
    ])

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })

    const list = page.getByRole('complementary', { name: 'Threads' })
    await expect(list.getByRole('listitem')).toHaveText([
      'Trip plan',
      'Budget',
    ])

    await list.getByRole('button', { name: 'Trip plan' }).click()
    await expect(
      list.getByRole('button', { name: 'Trip plan' })
    ).toHaveAttribute('aria-current', 'true')
    await send(page, 'where were we?')
    await expect(page.getByText('Back on the trip.')).toBeVisible()
    expect(fixture.requests[0].threadId).toBe('t-trip')
  })

  test('renames a thread and deletes the open one', async ({ page }) => {
    const fixture = await setupAGUI(page, { runs: [reply('Fresh start.')] })
    const calls = await setupThreads(page, [
      aguiThread('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
    ])

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    const list = page.getByRole('complementary', { name: 'Threads' })
    await list.getByRole('button', { name: 'Trip plan' }).click()

    await threadActions(page, 'Trip plan').click()
    await page.getByRole('menuitem', { name: 'Rename' }).click()
    const input = list.getByRole('textbox')
    await input.fill('Kyoto trip')
    await input.press('Enter')
    await expect(list.getByRole('button', { name: 'Kyoto trip' })).toBeVisible()
    expect(calls.renames).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui', title: 'Kyoto trip' },
    ])

    await threadActions(page, 'Kyoto trip').click()
    await page.getByRole('menuitem', { name: 'Delete' }).click()
    await page
      .getByRole('alertdialog')
      .getByRole('button', { name: 'Delete' })
      .click()
    await expect(page.getByText('Thread deleted')).toBeVisible()
    expect(calls.deletes).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui' },
    ])
    await expect(list.getByText('No threads with this agent yet.')).toBeVisible()

    // The deleted thread is gone: the next message starts a new one, and its
    // first run asks the server to title it.
    await send(page, 'hello again')
    await expect(page.getByText('Fresh start.')).toBeVisible()
    const newThreadId = fixture.requests[0].threadId as string
    expect(newThreadId).not.toBe('t-trip')
    await expect
      .poll(() => calls.generated)
      .toEqual([`agui-${newThreadId}`])
  })
})
