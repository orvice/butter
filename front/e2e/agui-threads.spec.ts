import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage as send,
  setupAGUI,
  sse,
  threadInURL,
} from './support/agui'

// The AG-UI thread list reads the caller's `agui` sessions and keeps those
// whose A2UI binding names the listed agent in the selected workspace. Each
// row links to its thread with ?thread=.

const reply = (text: string) =>
  sse([
    { type: 'RUN_STARTED', threadId: 't', runId: 'r' },
    { type: 'TEXT_MESSAGE_START', messageId: 'a', role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a', delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: 'a' },
    { type: 'RUN_FINISHED', threadId: 't', runId: 'r' },
  ])

const threadList = (page: Page) =>
  page.getByRole('complementary', { name: 'Threads' })

function threadActions(page: Page, title: string) {
  const row = threadList(page).getByRole('listitem').filter({ hasText: title })
  return row.getByRole('button', { name: 'Thread actions' })
}

test.describe('AG-UI threads', () => {
  test('lists only this agent’s threads and opens one by URL', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Back on the trip.')],
      sessions: [
        aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
        aguiSession('t-second', 'Second agent chat', { agentId: 'second-id' }, 2),
        aguiSession(
          't-ws2',
          'Other workspace',
          { agentId: 'streamer-id', workspaceId: 'ws-2' },
          3
        ),
        aguiSession('t-legacy', 'Before A2UI', null, 4),
        aguiSession('t-budget', 'Budget', { agentId: 'streamer-id' }, 5),
      ],
    })

    await page.goto('/agui-chat?agent=streamer-id', {
      waitUntil: 'networkidle',
    })

    const list = threadList(page)
    await expect(list.getByRole('listitem')).toHaveText([
      'Trip plan',
      'Budget',
    ])

    await list.getByRole('link', { name: 'Trip plan' }).click()
    await expect(page).toHaveURL(/\/agui-chat\?thread=t-trip$/)
    await expect(
      list.getByRole('link', { name: 'Trip plan' })
    ).toHaveAttribute('aria-current', 'page')
    await send(page, 'where were we?')
    await expect(page.getByText('Back on the trip.')).toBeVisible()
    expect(fixture.requests[0].threadId).toBe('t-trip')
  })

  test('lists threads past the first page', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: Array.from({ length: 101 }, (_, i) =>
        aguiSession(`t-${i}`, `Thread ${i}`, { agentId: 'streamer-id' }, i + 1)
      ),
    })

    await page.goto('/agui-chat?agent=streamer-id', {
      waitUntil: 'networkidle',
    })

    const list = threadList(page)
    await expect(list.getByRole('listitem')).toHaveCount(101)
    await expect(
      list.getByRole('link', { name: 'Thread 100', exact: true })
    ).toBeVisible()
    // The sidebar's chat history lists web-chat too; only the thread list
    // reads agui. It reached the second page (a dev-mode remount may walk
    // twice), always scoped to the workspace.
    const threadLists = fixture.sessionCalls.lists.filter(
      (c) => c.appName === 'agui'
    )
    expect(threadLists.map((c) => c.pageToken)).toContain('100')
    expect(threadLists.every((c) => c.workspaceScoped)).toBe(true)
  })

  test('renames a thread and deletes the open one', async ({ page }) => {
    const fixture = await setupAGUI(page, { runs: [reply('Fresh start.')] })
    fixture.sessions.push(
      aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1)
    )

    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
    const list = threadList(page)

    await threadActions(page, 'Trip plan').click()
    await page.getByRole('menuitem', { name: 'Rename' }).click()
    const input = list.getByRole('textbox')
    await input.fill('Kyoto trip')
    await input.press('Enter')
    await expect(list.getByRole('link', { name: 'Kyoto trip' })).toBeVisible()
    expect(fixture.sessionCalls.renames).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui', title: 'Kyoto trip' },
    ])

    await threadActions(page, 'Kyoto trip').click()
    await page.getByRole('menuitem', { name: 'Delete' }).click()
    await page
      .getByRole('alertdialog')
      .getByRole('button', { name: 'Delete' })
      .click()
    await expect(page.getByText('Thread deleted')).toBeVisible()
    expect(fixture.sessionCalls.deletes).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui' },
    ])
    await expect(list.getByText('No threads with this agent yet.')).toBeVisible()

    // The deleted thread is gone: the page is a new draft with its agent,
    // and the next message starts a new thread.
    expect(threadInURL(page)).toBeNull()
    await send(page, 'hello again')
    await expect(page.getByText('Fresh start.')).toBeVisible()
    const newThreadId = fixture.requests[0].threadId as string
    expect(newThreadId).not.toBe('t-trip')
    expect(threadInURL(page)).toBe(newThreadId)
  })

  test('shows the title the server gives a new thread, without asking for one', async ({
    page,
  }) => {
    // The run creates the thread's session, untitled, as the server does.
    const fixture = await setupAGUI(page, { runs: [reply('Here is a plan.')] })

    await page.goto('/agui-chat?agent=streamer-id', {
      waitUntil: 'networkidle',
    })
    const list = threadList(page)
    await expect(list.getByText('No threads with this agent yet.')).toBeVisible()

    await send(page, 'Plan a trip to Kyoto')
    await expect(page.getByText('Here is a plan.')).toBeVisible()
    // The list read when the run ends shows the thread before its title.
    await expect(
      list.getByRole('link', { name: 'Untitled thread' })
    ).toBeVisible()

    // The server stores the title after the run; the list and the header
    // catch up.
    fixture.sessions[0].title = 'Kyoto trip'
    await expect(list.getByRole('link', { name: 'Kyoto trip' })).toBeVisible({
      timeout: 15_000,
    })
    await expect(
      page.getByRole('heading', { name: 'Kyoto trip', level: 1 })
    ).toBeVisible()
    expect(fixture.sessionCalls.generated).toEqual([])
    expect(fixture.requests).toHaveLength(1)
  })

  test('restores a thread’s conversation when it opens, and sends only the new message', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Hotels are next.')],
      historyByThread: { 't-trip': tripHistory() },
      sessions: [
        aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
      ],
    })
    await page.goto('/agui-chat?agent=streamer-id', {
      waitUntil: 'networkidle',
    })

    await threadList(page).getByRole('link', { name: 'Trip plan' }).click()
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
            {
              id: 'int-1',
              reason: 'human_input',
              message: 'Approve the booking?',
            },
          ],
        }),
      },
      sessions: [
        aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1),
      ],
    })
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })

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
        {
          id: 'result:c1',
          role: 'tool',
          toolCallId: 'c1',
          content: '{"flights":2}',
        },
        { id: 'u2', role: 'user', content: 'Book the morning one' },
        { id: 'a2', role: 'assistant', content: 'Booked the 08:10 flight.' },
      ],
      interrupts: [],
      surfaces: [],
      ...extra,
    },
  }
}
