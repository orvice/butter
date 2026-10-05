import { expect, test, type Page } from '@playwright/test'
import { create } from '@bufbuild/protobuf'
import { SessionInfoSchema } from '../src/gen/agents/v1/agent_service_pb'
import {
  USER_ID,
  aguiSession,
  sendMessage as send,
  setupAGUI,
  sse,
  threadInURL,
} from './support/agui'

// The sidebar lists the caller's threads in the selected workspace, across
// agents: the `agui` sessions whose A2UI binding is for this workspace. It is
// Chat's history: rows are grouped by when the thread was last updated, show
// their agent's avatar, and link to the thread with ?thread=. The web-chat
// sessions of the Chat before AG-UI are not listed (#409).

const reply = (text: string) =>
  sse([
    { type: 'RUN_STARTED', threadId: 't', runId: 'r' },
    { type: 'TEXT_MESSAGE_START', messageId: 'a', role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a', delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: 'a' },
    { type: 'RUN_FINISHED', threadId: 't', runId: 'r' },
  ])

// Each agent has its own icon, so a row's avatar tells its agent, and a row's
// text is its title alone (without an icon, the avatar is the name's initial).
const icon = (color: string) =>
  `data:image/svg+xml,${encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><rect width="1" height="1" fill="${color}"/></svg>`
  )}`
const ICONS = { 'streamer-id': icon('#c2410c'), 'second-id': icon('#1d4ed8') }

// dayStart is the local start of the day daysAgo days before today, plus
// minutes. The sidebar groups threads by the local day, so a time measured
// from the start of a day stays in its group whatever the hour the test runs.
function dayStart(daysAgo: number, minutes = 0): Date {
  const day = new Date()
  day.setHours(0, 0, 0, 0)
  day.setDate(day.getDate() - daysAgo)
  return new Date(day.getTime() + minutes * 60_000)
}

const threadList = (page: Page) =>
  page.getByRole('navigation', { name: 'Threads' })

// rowLinks are the thread links of the list, or of one of its date groups.
const rowLinks = (page: Page, group?: string) =>
  (group
    ? threadList(page).getByRole('group', { name: group })
    : threadList(page).getByRole('group')
  ).getByRole('link')

const threadRow = (page: Page, title: string) =>
  threadList(page)
    .locator('[data-sidebar="menu-item"]')
    .filter({ has: page.getByRole('link', { name: title, exact: true }) })

async function threadAction(
  page: Page,
  title: string,
  action: 'Rename' | 'Delete'
) {
  const row = threadRow(page, title)
  await row.hover()
  await row.getByRole('button', { name: 'Thread actions' }).click()
  await page.getByRole('menuitem', { name: action }).click()
}

// threadHeader is the bar above the open thread, found by its title.
const threadHeader = (page: Page, title: string) =>
  page
    .locator('header')
    .filter({ has: page.getByRole('heading', { name: title, level: 1 }) })

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, dayStart(0, 2))

test.describe('Threads in the sidebar', () => {
  test('lists every thread of the workspace across agents and pages, grouped by date', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      agentIcons: ICONS,
      sessions: [
        aguiSession(
          't-second',
          'Second agent chat',
          { agentId: 'second-id' },
          dayStart(0, 4)
        ),
        aguiSession('t-untitled', '', { agentId: 'second-id' }, dayStart(0, 3)),
        trip(),
        aguiSession(
          't-ws2',
          'Other workspace',
          { agentId: 'streamer-id', workspaceId: 'ws-2' },
          dayStart(0, 1)
        ),
        aguiSession('t-legacy', 'Before A2UI', null, dayStart(1)),
        aguiSession(
          't-budget',
          'Budget review',
          { agentId: 'streamer-id' },
          dayStart(2)
        ),
        // Enough older threads that the listing takes a second page.
        ...Array.from({ length: 100 }, (_, i) =>
          aguiSession(
            `t-old-${i}`,
            `Old thread ${i}`,
            { agentId: i % 2 ? 'second-id' : 'streamer-id' },
            dayStart(30 + i)
          )
        ),
        // A conversation of the Chat before AG-UI, which is not carried over.
        create(SessionInfoSchema, {
          sessionId: 'chat-1',
          appName: 'web-chat',
          userId: USER_ID,
          title: 'An old chat',
        }),
      ],
    })

    await page.goto('/chat', { waitUntil: 'networkidle' })

    await expect(rowLinks(page, 'Today')).toHaveText([
      'Second agent chat',
      // An untitled thread shows its agent's name.
      'Second',
      'Trip plan',
    ])
    await expect(rowLinks(page, 'Previous 7 days')).toHaveText([
      'Budget review',
    ])
    const older = rowLinks(page, 'Older')
    await expect(older).toHaveCount(100)
    await expect(older.last()).toHaveText('Old thread 99')
    // Threads of another workspace, and threads from before A2UI, which
    // cannot be attributed, are left out. No web-chat conversation shows
    // anywhere in the sidebar: nothing lists them (below).
    await expect(rowLinks(page)).toHaveCount(104)
    await expect(
      page.locator('[data-sidebar="sidebar"]').getByRole('link', {
        name: 'An old chat',
      })
    ).toHaveCount(0)

    // Each row shows its agent's avatar.
    for (const [title, agentId] of [
      ['Trip plan', 'streamer-id'],
      ['Second agent chat', 'second-id'],
      ['Old thread 98', 'streamer-id'],
      ['Old thread 99', 'second-id'],
    ] as const) {
      await expect(
        threadRow(page, title).locator('img'),
        title
      ).toHaveAttribute('src', ICONS[agentId])
    }

    // Every page was read, always scoped to the workspace.
    const threadLists = fixture.sessionCalls.lists.filter(
      (c) => c.appName === 'agui'
    )
    expect(threadLists.map((c) => c.pageToken)).toContain('100')
    expect(threadLists.every((c) => c.workspaceScoped)).toBe(true)
    expect(fixture.sessionCalls.lists.map((c) => c.appName)).not.toContain(
      'web-chat'
    )
  })

  test('search filters by title', async ({ page }) => {
    await setupAGUI(page, {
      runs: [],
      agentIcons: ICONS,
      sessions: [
        trip(),
        aguiSession('t-untitled', '', { agentId: 'second-id' }, dayStart(0, 1)),
        aguiSession(
          't-budget',
          'Budget review',
          { agentId: 'streamer-id' },
          dayStart(2)
        ),
      ],
    })
    await page.goto('/chat', { waitUntil: 'networkidle' })
    const search = threadList(page).getByRole('textbox', {
      name: 'Search threads',
    })

    await search.fill('BUDGET')
    await expect(rowLinks(page)).toHaveText(['Budget review'])
    await expect(
      threadList(page).getByRole('group', { name: 'Today' })
    ).toHaveCount(0)

    // An untitled thread is found by the agent name it shows.
    await search.fill('second')
    await expect(rowLinks(page)).toHaveText(['Second'])

    await search.fill('nothing like this')
    await expect(rowLinks(page)).toHaveCount(0)
    await expect(threadList(page).getByText('No threads found.')).toBeVisible()

    await search.fill('')
    await expect(rowLinks(page)).toHaveCount(3)
  })

  test('a row opens its thread and stays highlighted while it is open', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Back on the trip.')],
      sessions: [
        trip(),
        aguiSession(
          't-second',
          'Second agent chat',
          { agentId: 'second-id' },
          dayStart(0, 1)
        ),
      ],
    })
    await page.goto('/chat', { waitUntil: 'networkidle' })

    const tripLink = threadList(page).getByRole('link', { name: 'Trip plan' })
    const secondLink = threadList(page).getByRole('link', {
      name: 'Second agent chat',
    })
    await expect(tripLink).toHaveAttribute('data-active', 'false')

    await tripLink.click()
    await expect(page).toHaveURL(/\/chat\?thread=t-trip$/)
    await expect(threadHeader(page, 'Trip plan')).toContainText('Streamer')
    await expect(tripLink).toHaveAttribute('aria-current', 'page')
    await expect(tripLink).toHaveAttribute('data-active', 'true')
    await expect(secondLink).toHaveAttribute('data-active', 'false')
    await send(page, 'where were we?')
    await expect(page.getByText('Back on the trip.')).toBeVisible()
    expect(fixture.requests[0].threadId).toBe('t-trip')
    expect(new URL(fixture.runURLs[0]).pathname).toBe('/api/agui/streamer-id')

    // Another agent's thread opens with its own agent.
    await secondLink.click()
    await expect(page).toHaveURL(/\/chat\?thread=t-second$/)
    await expect(threadHeader(page, 'Second agent chat')).toContainText(
      'Second'
    )
    await expect(secondLink).toHaveAttribute('data-active', 'true')
    await expect(tripLink).toHaveAttribute('data-active', 'false')
    await expect
      .poll(() => fixture.historyRequests.at(-1))
      .toContain('/api/agui/second-id/threads/t-second/messages')
  })

  test('renaming from the sidebar updates the row and the page header', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, { runs: [], sessions: [trip()] })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await expect(threadHeader(page, 'Trip plan')).toBeVisible()

    await threadAction(page, 'Trip plan', 'Rename')
    const input = threadList(page).getByRole('textbox', { name: 'Chat title' })
    await expect(input).toHaveValue('Trip plan')
    await input.fill('Kyoto trip')
    await input.press('Enter')

    await expect(
      threadList(page).getByRole('link', { name: 'Kyoto trip' })
    ).toBeVisible()
    await expect(threadHeader(page, 'Kyoto trip')).toBeVisible()
    expect(fixture.sessionCalls.renames).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui', title: 'Kyoto trip' },
    ])
  })

  test('deleting another thread keeps the open one; deleting the open one lands on a new-chat draft', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Fresh start.')],
      agentIcons: ICONS,
      sessions: [
        trip(),
        aguiSession(
          't-budget',
          'Budget review',
          { agentId: 'second-id' },
          dayStart(2)
        ),
      ],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })

    await threadAction(page, 'Budget review', 'Delete')
    const dialog = page.getByRole('alertdialog')
    await expect(dialog).toContainText('Budget review')
    await dialog.getByRole('button', { name: 'Delete' }).click()
    await expect(rowLinks(page)).toHaveText(['Trip plan'])
    await expect(page).toHaveURL(/\/chat\?thread=t-trip$/)
    await expect(threadHeader(page, 'Trip plan')).toBeVisible()

    await threadAction(page, 'Trip plan', 'Delete')
    await dialog.getByRole('button', { name: 'Delete' }).click()
    await expect(page.getByText('Thread deleted').first()).toBeVisible()
    expect(fixture.sessionCalls.deletes).toEqual([
      { sessionId: 'agui-t-budget', appName: 'agui' },
      { sessionId: 'agui-t-trip', appName: 'agui' },
    ])
    await expect(threadList(page).getByText('No threads found.')).toBeVisible()

    // The open thread is gone: the page is a new draft with its agent, and
    // the next message starts a new thread.
    await expect(page).toHaveURL(/\/chat\?agent=streamer-id$/)
    await expect(
      page.getByRole('heading', { name: 'Streamer', exact: true })
    ).toBeVisible()
    await send(page, 'hello again')
    await expect(page.getByText('Fresh start.')).toBeVisible()
    const newThreadId = fixture.requests[0].threadId as string
    expect(newThreadId).not.toBe('t-trip')
    expect(threadInURL(page)).toBe(newThreadId)
  })

  test('New thread starts a draft with the open thread’s agent', async ({
    page,
  }) => {
    await setupAGUI(page, {
      runs: [],
      sessions: [
        aguiSession(
          't-second',
          'Second agent chat',
          { agentId: 'second-id' },
          dayStart(0, 1)
        ),
      ],
    })
    await page.goto('/chat?thread=t-second', { waitUntil: 'networkidle' })
    await expect(threadHeader(page, 'Second agent chat')).toBeVisible()

    await threadList(page).getByRole('link', { name: 'New thread' }).click()
    await expect(page).toHaveURL(/\/chat\?agent=second-id$/)
    await expect(
      page.getByRole('heading', { name: 'Second', exact: true })
    ).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Message' })).toBeEnabled()
  })

  test('shows the title the server gives a new thread, without asking for one', async ({
    page,
  }) => {
    // The run creates the thread's session, untitled, as the server does.
    const fixture = await setupAGUI(page, { runs: [reply('Here is a plan.')] })

    await page.goto('/chat?agent=streamer-id', {
      waitUntil: 'networkidle',
    })
    await expect(threadList(page).getByText('No threads found.')).toBeVisible()

    await send(page, 'Plan a trip to Kyoto')
    await expect(page.getByText('Here is a plan.')).toBeVisible()
    // The list read when the run ends shows the thread, under its agent's
    // name until it has a title.
    const row = threadList(page).getByRole('link', { name: 'Streamer' })
    await expect(row).toBeVisible()
    await expect(row).toHaveAttribute('aria-current', 'page')

    // The server stores the title after the run; the list and the header
    // catch up.
    fixture.sessions[0].title = 'Kyoto trip'
    await expect(
      threadList(page).getByRole('link', { name: 'Kyoto trip' })
    ).toBeVisible({ timeout: 15_000 })
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
      sessions: [trip()],
    })
    await page.goto('/chat?agent=streamer-id', {
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
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })

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

test.describe('Threads on a narrow screen', () => {
  test.use({ viewport: { width: 390, height: 844 } })

  test('threads are reached from the sidebar sheet, not from the page', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      historyByThread: { 't-trip': tripHistory() },
      sessions: [
        trip(),
        aguiSession(
          't-budget',
          'Budget review',
          { agentId: 'second-id' },
          dayStart(2)
        ),
      ],
    })
    await page.goto('/chat', { waitUntil: 'networkidle' })
    // The page has no thread list or drawer of its own.
    await expect(threadList(page)).toHaveCount(0)
    await expect(
      page.getByRole('button', { name: 'Show threads' })
    ).toHaveCount(0)

    await page.getByRole('button', { name: 'Toggle Sidebar' }).first().click()
    const sheet = page.getByRole('dialog', { name: 'Sidebar' })
    const list = sheet.getByRole('navigation', { name: 'Threads' })

    // A thread is deleted from the sheet too; the sheet stays open.
    await threadAction(page, 'Budget review', 'Delete')
    await page
      .getByRole('alertdialog')
      .getByRole('button', { name: 'Delete' })
      .click()
    await expect(list.getByRole('link', { name: 'Budget review' })).toHaveCount(
      0
    )
    expect(fixture.sessionCalls.deletes).toEqual([
      { sessionId: 'agui-t-budget', appName: 'agui' },
    ])

    await list.getByRole('link', { name: 'Trip plan' }).click()
    await expect(page).toHaveURL(/\/chat\?thread=t-trip$/)
    // Opening a thread closes the sheet on it.
    await expect(sheet).toHaveCount(0)
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
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
