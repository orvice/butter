import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage,
  setupAGUI,
  sse,
  threadInURL,
} from './support/agui'

// AG-UI Chat takes its thread from the URL: ?thread=<id> opens that thread
// with the agent its binding names, and without it the page is a new-chat
// draft whose first message starts a thread and writes ?thread=.

const reply = (text: string) =>
  sse([
    { type: 'RUN_STARTED', threadId: 't', runId: 'r' },
    { type: 'TEXT_MESSAGE_START', messageId: 'a', role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a', delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: 'a' },
    { type: 'RUN_FINISHED', threadId: 't', runId: 'r' },
  ])

const tripHistory = {
  body: {
    threadId: 't-trip',
    messages: [
      { id: 'u1', role: 'user', content: 'Plan a trip to Tokyo' },
      { id: 'a1', role: 'assistant', content: 'Booked the 08:10 flight.' },
    ],
    interrupts: [],
    surfaces: [],
  },
}

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1)

// threadHeader is the bar above an open thread, found by its title.
const threadHeader = (page: Page, title: string) =>
  page
    .locator('header')
    .filter({ has: page.getByRole('heading', { name: title, level: 1 }) })

const notFound = (page: Page) => page.getByText('Thread not found.')

const pickAgent = async (page: Page, query: string, name: string) => {
  await page.getByTestId('agent-selector-trigger').click()
  await page.getByRole('textbox', { name: 'Search agents' }).fill(query)
  await page.getByRole('option', { name: new RegExp(name) }).click()
}

test.describe('AG-UI Chat threads by URL', () => {
  test('opening ?thread= shows that thread with its agent', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('Hotels are next.')],
      sessions: [trip()],
      historyByThread: { 't-trip': tripHistory },
    })
    // Reading the thread (and the thread list) is slow enough to see.
    await page.route('**/agents.v1.SessionService/*', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 600))
      await route.fallback()
    })

    await page.goto('/agui-chat?thread=t-trip')
    await expect(page.getByText('Loading thread…')).toBeVisible()
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
    await expect(threadHeader(page, 'Trip plan')).toContainText('Streamer')
    // The thread's agent comes from its binding.
    expect(
      fixture.historyRequests.every((u) =>
        u.includes('/api/agui/streamer-id/threads/t-trip/messages')
      )
    ).toBe(true)

    await sendMessage(page, 'And hotels?')
    await expect(page.getByText('Hotels are next.')).toBeVisible()
    expect(fixture.requests.map((r) => r.threadId)).toEqual(['t-trip'])
    expect(new URL(fixture.runURLs[0]).pathname).toBe('/api/agui/streamer-id')
  })

  test('a thread that does not exist, or is not this workspace’s or agent’s, is not found', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [
        trip(),
        aguiSession('t-ws2', 'Elsewhere', {
          agentId: 'streamer-id',
          workspaceId: 'ws-2',
        }),
        aguiSession('t-gone', 'Retired agent', { agentId: 'retired-id' }),
        aguiSession('t-legacy', 'Before A2UI', null),
      ],
    })

    for (const url of [
      '/agui-chat?thread=missing',
      '/agui-chat?thread=t-ws2',
      '/agui-chat?thread=t-gone',
      '/agui-chat?thread=t-legacy',
      '/agui-chat?thread=t-trip&agent=second-id',
    ]) {
      await page.goto(url)
      await expect(notFound(page), url).toBeVisible()
      await expect(page.getByRole('textbox', { name: /^Message/ })).toHaveCount(
        0
      )
    }
    // Nothing was read or run for a thread that is not this one.
    expect(fixture.historyRequests).toEqual([])
    expect(fixture.requests).toEqual([])

    await page.getByRole('button', { name: 'Start a new chat' }).click()
    await expect(page).toHaveURL(/\/agui-chat$/)
    await expect(
      page.getByRole('heading', { name: 'Start a new chat' })
    ).toBeVisible()
  })

  test('a failed history load offers Retry', async ({ page }) => {
    await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      histories: [
        { status: 500, body: { error: 'session store unavailable' } },
        tripHistory,
      ],
    })

    await page.goto('/agui-chat?thread=t-trip')
    await expect(page.getByText('Failed to load this thread.')).toBeVisible()
    await expect(page.getByText('session store unavailable')).toBeVisible()

    await page.getByRole('button', { name: 'Retry' }).click()
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
    await expect(page.getByText('Failed to load this thread.')).toHaveCount(0)
  })

  test('a failed thread lookup offers Retry', async ({ page }) => {
    await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': tripHistory },
    })
    let down = true
    await page.route('**/agents.v1.SessionService/GetSession', async (route) => {
      if (!down) return route.fallback()
      await route.fulfill({
        status: 503,
        contentType: 'application/json',
        body: JSON.stringify({ code: 'unavailable', message: 'try later' }),
      })
    })
    // The thread list would stand in for the lookup; keep it empty.
    await page.route('**/agents.v1.SessionService/ListSessions', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/proto',
        body: Buffer.alloc(0),
      })
    )

    await page.goto('/agui-chat?thread=t-trip')
    await expect(page.getByText('Failed to load this thread.')).toBeVisible()
    await expect(page.getByText('try later')).toBeVisible()
    down = false
    await page.getByRole('button', { name: 'Retry' }).click()
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
  })

  test('a send the server refuses for the thread shows it as not found', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        {
          status: 403,
          body: { error: 'threadId is not available; start a new thread' },
        },
      ],
      sessions: [trip()],
      historyByThread: { 't-trip': tripHistory },
    })

    await page.goto('/agui-chat?thread=t-trip')
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
    await sendMessage(page, 'still there?')

    await expect(notFound(page)).toBeVisible()
    await expect(
      page.getByText('threadId is not available; start a new thread')
    ).toBeVisible()
    expect(fixture.requests).toHaveLength(1)

    await page.getByRole('button', { name: 'Start a new chat' }).click()
    await expect(page).toHaveURL(/\/agui-chat\?agent=streamer-id$/)
    await expect(page.getByRole('heading', { name: 'Streamer' })).toBeVisible()
  })

  test('a draft picks an agent by search, and its first message writes ?thread=', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, { runs: [reply('Hi from Second.')] })
    await page.goto('/agui-chat')

    // No agent is chosen for a person who never picked one.
    await expect(
      page.getByRole('heading', { name: 'Start a new chat' })
    ).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Message' })).toBeDisabled()

    await pickAgent(page, 'sec', 'Second')
    await expect(page.getByRole('heading', { name: 'Second' })).toBeVisible()
    await expect(page.getByTestId('agent-selector-trigger')).toContainText(
      'Second'
    )

    const entries = await page.evaluate(() => history.length)
    await sendMessage(page, 'Hello there')
    await expect(page.getByText('Hi from Second.')).toBeVisible()
    await expect(page.getByText('Hello there', { exact: true })).toBeVisible()
    const threadId = threadInURL(page)
    expect(threadId).toBeTruthy()
    // The URL is replaced, not pushed.
    expect(await page.evaluate(() => history.length)).toBe(entries)
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.requests[0].threadId).toBe(threadId)
    expect(new URL(fixture.runURLs[0]).pathname).toBe('/api/agui/second-id')
    // A new thread has nothing to read.
    expect(fixture.historyRequests).toEqual([])

    // A reload keeps the thread and reads its conversation.
    fixture.historyByThread = {
      [threadId!]: {
        body: {
          threadId,
          messages: [
            { id: 'u1', role: 'user', content: 'Hello there' },
            { id: 'a1', role: 'assistant', content: 'Hi from Second.' },
          ],
          interrupts: [],
          surfaces: [],
        },
      },
    }
    await page.reload()
    await expect(page.getByText('Hi from Second.')).toBeVisible()
    expect(threadInURL(page)).toBe(threadId)
    await expect(threadHeader(page, 'Second')).toBeVisible()
    expect(fixture.historyRequests.at(-1)).toContain(
      `/api/agui/second-id/threads/${threadId}/messages`
    )
  })

  test('the agent picked last is preselected next time in the same workspace', async ({
    page,
  }) => {
    await setupAGUI(
      page,
      { runs: [] },
      {
        workspaces: [
          { id: 'default', name: 'Default', slug: 'default' },
          { id: 'team-b', name: 'Team B', slug: 'team-b' },
        ],
      }
    )
    await page.goto('/agui-chat')
    await pickAgent(page, 'plain', 'Plain')
    await expect(page.getByRole('heading', { name: 'Plain' })).toBeVisible()

    await page.goto('/agui-chat')
    await expect(page.getByRole('heading', { name: 'Plain' })).toBeVisible()
    await expect(page.getByTestId('agent-selector-trigger')).toContainText(
      'Plain'
    )
    // ?agent= wins over the remembered agent.
    await page.goto('/agui-chat?agent=second-id')
    await expect(page.getByRole('heading', { name: 'Second' })).toBeVisible()

    // Another workspace remembers its own.
    await page.goto('/agui-chat')
    await page.getByRole('button', { name: /Default/ }).first().click()
    await page.getByRole('menuitem', { name: 'Team B' }).click()
    await expect(
      page.getByRole('heading', { name: 'Start a new chat' })
    ).toBeVisible()
  })

  test('the header renames the thread and deletes it into a new draft', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': tripHistory },
    })
    await page.goto('/agui-chat?thread=t-trip')
    const header = threadHeader(page, 'Trip plan')
    await expect(header).toContainText('Streamer')

    await header.getByRole('button', { name: 'Rename thread' }).click()
    const title = page.getByRole('textbox', { name: 'Chat title' })
    await title.fill('Kyoto trip')
    await title.press('Enter')
    await expect(threadHeader(page, 'Kyoto trip')).toBeVisible()
    await expect(
      page
        .getByRole('navigation', { name: 'AG-UI threads' })
        .getByRole('link', { name: 'Kyoto trip' })
    ).toBeVisible()
    expect(fixture.sessionCalls.renames).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui', title: 'Kyoto trip' },
    ])

    await page.getByRole('button', { name: 'Thread options' }).click()
    await page.getByRole('menuitem', { name: 'Delete thread' }).click()
    const dialog = page.getByRole('alertdialog')
    await expect(dialog).toContainText('Kyoto trip')
    await dialog.getByRole('button', { name: 'Delete' }).click()

    await expect(page.getByText('Thread deleted')).toBeVisible()
    expect(fixture.sessionCalls.deletes).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui' },
    ])
    await expect(page).toHaveURL(/\/agui-chat\?agent=streamer-id$/)
    await expect(page.getByRole('heading', { name: 'Streamer' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Message' })).toBeEnabled()
  })
})
