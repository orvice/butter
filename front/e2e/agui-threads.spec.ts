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
  lists: Array<{ appName: string; workspaceScoped: boolean; pageToken: string }>
  renames: Array<{ sessionId: string; appName: string; title: string }>
  deletes: Array<{ sessionId: string; appName: string }>
  generated: string[]
}

async function setupThreads(page: Page, sessions: SessionInfo[]) {
  const calls: SessionCalls = {
    lists: [],
    renames: [],
    deletes: [],
    generated: [],
  }
  // Registered after setupAGUI's catch-all, so it answers SessionService.
  await page.route('**/api/agents.v1.SessionService/**', async (route) => {
    const url = route.request().url()
    const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
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

  test('lists threads past the first page', async ({ page }) => {
    await setupAGUI(page, { runs: [] })
    const calls = await setupThreads(
      page,
      Array.from({ length: 101 }, (_, i) =>
        aguiThread(`t-${i}`, `Thread ${i}`, { agentId: 'streamer-id' }, i + 1)
      )
    )

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })

    const list = page.getByRole('complementary', { name: 'Threads' })
    await expect(list.getByRole('listitem')).toHaveCount(101)
    await expect(
      list.getByRole('button', { name: 'Thread 100', exact: true })
    ).toBeVisible()
    // The sidebar's chat history lists web-chat too; only the thread list
    // reads agui. It reached the second page (a dev-mode remount may walk
    // twice), always scoped to the workspace.
    const threadLists = calls.lists.filter((c) => c.appName === 'agui')
    expect(threadLists.map((c) => c.pageToken)).toContain('100')
    expect(threadLists.every((c) => c.workspaceScoped)).toBe(true)
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

  test('restores a thread’s conversation when it opens, and sends only the new message', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Hotels are next.')],
      historyByThread: { 't-trip': tripHistory() },
    })
    await setupThreads(page, [
      aguiThread('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
    ])
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })

    await page
      .getByRole('complementary', { name: 'Threads' })
      .getByRole('button', { name: 'Trip plan' })
      .click()
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
    await expect(
      page.getByRole('button', { name: 'searchFlights', exact: true })
    ).toBeVisible()
    const shown = await page.locator('main').innerText()
    const order = [
      'Plan a trip to Tokyo',
      'Here is the plan.',
      'Book the morning one',
      'Booked the 08:10 flight.',
    ].map((text) => shown.indexOf(text))
    expect(order).toEqual([...order].sort((a, b) => a - b))
    expect(order.every((i) => i >= 0)).toBe(true)

    // The restored tool call has its result: sending starts exactly one run
    // that ends on the new message, never a run of tool results first.
    await send(page, 'And hotels?')
    await expect(page.getByText('Hotels are next.')).toBeVisible()
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.requests[0].threadId).toBe('t-trip')
    const messages = fixture.requests[0].messages as Array<{
      role: string
      content?: unknown
    }>
    expect(messages.at(-1)?.role).toBe('user')
    expect(JSON.stringify(messages.at(-1))).toContain('And hotels?')
  })

  test('restores an open question so it can be answered', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Approved and deployed.')],
      historyByThread: {
        't-trip': tripHistory({
          interrupts: [
            { id: 'int-1', reason: 'human_input', message: 'Approve the booking?' },
          ],
        }),
      },
    })
    await setupThreads(page, [
      aguiThread('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
    ])
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await page
      .getByRole('complementary', { name: 'Threads' })
      .getByRole('button', { name: 'Trip plan' })
      .click()

    await expect(page.getByText('Approve the booking?')).toBeVisible()
    await page.getByPlaceholder('Type your answer…').fill('yes')
    await page.getByRole('button', { name: 'Answer' }).click()
    await expect(page.getByText('Approved and deployed.')).toBeVisible()
    const resume = fixture.requests[0].resume as Array<Record<string, unknown>>
    expect(resume).toEqual([
      { interruptId: 'int-1', status: 'resolved', payload: 'yes' },
    ])
  })
})

// tripHistory is the history of the thread t-trip: two turns, the first with
// a tool call and its result.
function tripHistory(extra: Record<string, unknown> = {}) {
  return {
    body: {
      threadId: 't-trip',
      messages: [
        { id: 'u1', role: 'user', content: 'Plan a trip to Tokyo' },
        {
          id: 'a1',
          role: 'assistant',
          content: 'Here is the plan.',
          toolCalls: [
            {
              id: 'c1',
              type: 'function',
              function: { name: 'searchFlights', arguments: '{"to":"NRT"}' },
            },
          ],
        },
        { id: 'result:c1', role: 'tool', toolCallId: 'c1', content: '{"flights":2}' },
        { id: 'u2', role: 'user', content: 'Book the morning one' },
        { id: 'a2', role: 'assistant', content: 'Booked the 08:10 flight.' },
      ],
      interrupts: [],
      surfaces: [],
      ...extra,
    },
  }
}
