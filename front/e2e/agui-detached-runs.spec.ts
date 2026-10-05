import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage as send,
  setupAGUI,
  sse,
  stoppedRun,
  threadInURL,
  type AGUIFixture,
} from './support/agui'

// Every Chat run asks to detach (ADR-0016), so it outlives its
// request. Leaving a thread only detaches the page from the run. Stop asks
// the server to stop the run, then ends it here. Deleting the open thread
// leaves the stopping to the server. An open run stays in flight until a
// Stop or the test ends it, so the page shows it running meanwhile.

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' })
const budget = () =>
  aguiSession('t-budget', 'Budget review', { agentId: 'streamer-id' }, 5)

const running = (page: Page) => page.getByText('Running…')
const stopButton = (page: Page) => page.getByRole('button', { name: 'Stop' })
const sendButton = (page: Page) => page.getByRole('button', { name: 'Send' })
const composer = (page: Page) => page.getByRole('textbox', { name: /^Message/ })
const toasts = (page: Page) => page.locator('[data-sonner-toast]')
// The message the user sent; the stopped run's notice quotes it too.
const sent = (page: Page, text: string) =>
  page
    .locator('[data-message-role="user"]')
    .getByText(text, { exact: true })
const stoppedNotice = (page: Page) =>
  page.getByRole('status').filter({ hasText: 'Stopped' })
const threadLink = (page: Page, title: string) =>
  page
    .getByRole('navigation', { name: 'Threads' })
    .getByRole('link', { name: title, exact: true })

const stopPaths = (urls: string[]) => urls.map((u) => new URL(u).pathname)

const question = (id: string, message: string) => ({
  id,
  reason: 'human_input',
  message,
})

const reply = (runId: string, text: string) => [
  { type: 'TEXT_MESSAGE_START', messageId: `a-${runId}`, role: 'assistant' },
  { type: 'TEXT_MESSAGE_CONTENT', messageId: `a-${runId}`, delta: text },
  { type: 'TEXT_MESSAGE_END', messageId: `a-${runId}` },
]

function finished(runId: string, text: string) {
  return sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    ...reply(runId, text),
    { type: 'RUN_FINISHED', threadId: 't', runId },
  ])
}

function pausedOn(
  runId: string,
  text: string,
  ...interrupts: Array<ReturnType<typeof question>>
) {
  return sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    ...reply(runId, text),
    {
      type: 'RUN_FINISHED',
      threadId: 't',
      runId,
      outcome: { type: 'interrupt', interrupts },
    },
  ])
}

// deferred is a promise the test settles.
function deferred() {
  let resolve!: () => void
  const promise = new Promise<void>((r) => (resolve = r))
  return { promise, resolve }
}

test.describe('Chat detached runs', () => {
  test('every run asks to detach', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        finished('r1', 'Hello.'),
        pausedOn(
          'r2',
          'Two things first.',
          question('int-1', 'Which region?'),
          question('int-2', 'Which version?')
        ),
        pausedOn('r3', 'Version noted.', question('int-1', 'Which region?')),
        finished('r4', 'Deploying to eu-west.'),
      ],
    })
    await page.goto('/chat?agent=streamer-id', {
      waitUntil: 'networkidle',
    })

    // The draft's first message starts the thread.
    await send(page, 'hi')
    await expect(page.getByText('Hello.')).toBeVisible()
    // A message on the open thread.
    await send(page, 'deploy')
    const version = page.getByRole('group', { name: 'Which version?' })
    await expect(version).toBeVisible()
    // An answer to one question, from its prompt.
    await version.getByPlaceholder('Type your answer…').fill('2.4.1')
    await version.getByRole('button', { name: 'Answer' }).click()
    await expect(page.getByText('Version noted.')).toBeVisible()
    // A message the server takes as the answer to the oldest question.
    await send(page, 'eu-west')
    await expect(page.getByText('Deploying to eu-west.')).toBeVisible()

    expect(fixture.requests).toHaveLength(4)
    expect(fixture.requests[2].resume).toHaveLength(1)
    expect(fixture.requests[3]).not.toHaveProperty('resume')
    for (const request of fixture.requests) {
      expect(request.forwardedProps).toMatchObject({
        butterA2UI: { catalogs: ['butter-basic-v1'] },
        butterRun: { detach: true },
      })
    }
  })

  test('Stop sends one stop request, and the run ends here once the server took it', async ({
    page,
  }) => {
    const accepted = deferred()
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
      stops: [{ status: 202, until: accepted.promise }],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).dblclick()
    await expect(stopButton(page)).toBeDisabled()
    // Until the server takes the Stop, the run goes on here.
    await expect(running(page)).toBeVisible()
    expect(stopPaths(fixture.stopRequests)).toEqual([
      '/api/agui/streamer-id/threads/t-trip/stop',
    ])

    // The server takes it (202); the run's stream then ends as stopped.
    accepted.resolve()
    await expect(running(page)).toHaveCount(0)
    await expect(sendButton(page)).toBeVisible()
    // The page shows the run as stopped, quoting the turn it started from.
    // Whether the turn also stays in the thread depends on which ends the
    // run here first: its stopped stream, or the local cancel, which takes
    // a turn with no reply yet back into the composer.
    await expect(stoppedNotice(page)).toContainText('plan the trip')
    // A stopped run is not a failure.
    await expect(toasts(page)).toHaveCount(0)
    expect(fixture.stopRequests).toHaveLength(1)
    expect(fixture.requests).toHaveLength(1)
  })

  test('the page settles on a run whose stopped stream ends before the Stop is answered', async ({
    page,
  }) => {
    const accepted = deferred()
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
      stops: [{ status: 202, until: accepted.promise }],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).click()
    await expect.poll(() => fixture.stopRequests.length).toBe(1)
    // The stopped run's RUN_ERROR reaches the page ahead of the Stop's 202.
    const runId = String(fixture.requests[0].runId)
    fixture.endOpenRun('t-trip', stoppedRun('t-trip', runId))
    await expect(running(page)).toHaveCount(0)
    await expect(sendButton(page)).toBeVisible()

    const answered = page.waitForResponse((r) => r.url().endsWith('/stop'))
    accepted.resolve()
    expect((await answered).status()).toBe(202)
    // Nothing is left to cancel: the turn stays as it ended.
    await page.waitForTimeout(250)
    await expect(sent(page, 'plan the trip')).toBeVisible()
    await expect(composer(page)).toHaveValue('')
    await expect(toasts(page)).toHaveCount(0)
    expect(fixture.stopRequests).toHaveLength(1)
  })

  test('Stop still ends the run here when the server has none in flight', async ({
    page,
  }) => {
    // 204: nothing to stop on the server, and nothing ends the stream but
    // the page.
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
      stops: [{ status: 204 }],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).click()
    await expect(running(page)).toHaveCount(0)
    await expect(sendButton(page)).toBeVisible()
    await expect.poll(() => fixture.abortedRuns).toEqual(['t-trip'])
    expect(fixture.stopRequests).toHaveLength(1)
    await expect(toasts(page)).toHaveCount(0)
  })

  test('a Stop the server refuses leaves the run going, and Stop can be tried again', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
      stops: [
        { status: 503, body: { error: 'stop unavailable, retry later' } },
      ],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).click()
    await expect(
      toasts(page).filter({
        hasText: 'Could not stop the run: stop unavailable, retry later',
      })
    ).toBeVisible()
    await expect(running(page)).toBeVisible()
    await expect(stopButton(page)).toBeEnabled()
    expect(fixture.abortedRuns).toEqual([])

    await stopButton(page).click()
    await expect(running(page)).toHaveCount(0)
    expect(fixture.stopRequests).toHaveLength(2)
  })

  test('switching thread, New thread and leaving the page only detach the run', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }, { open: true }, { open: true }],
      sessions: [trip(), budget()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    // Another thread.
    await threadLink(page, 'Budget review').click()
    await expect(page).toHaveURL(/\/chat\?thread=t-budget$/)
    await expect.poll(() => fixture.abortedRuns).toEqual(['t-trip'])
    await expect(running(page)).toHaveCount(0)

    // New thread.
    await send(page, 'review the budget')
    await expect(running(page)).toBeVisible()
    await page.getByRole('link', { name: 'New thread' }).click()
    await expect(page).toHaveURL(/\/chat\?agent=streamer-id$/)
    await expect.poll(() => fixture.abortedRuns).toEqual(['t-trip', 't-budget'])

    // Another page.
    await send(page, 'hello')
    await expect(running(page)).toBeVisible()
    const started = threadInURL(page)
    expect(started).toBe(fixture.requests[2].threadId)
    await page
      .getByRole('link', { name: 'Agents', exact: true })
      .first()
      .click()
    await expect(page).toHaveURL(/\/agents/)
    await expect
      .poll(() => fixture.abortedRuns)
      .toEqual(['t-trip', 't-budget', started])

    expect(fixture.stopRequests).toEqual([])
  })

  test('reloading during a run sends no stop request', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await page.reload({ waitUntil: 'networkidle' })
    await expect(composer(page)).toBeVisible()
    expect(fixture.stopRequests).toEqual([])
  })

  test('deleting the open thread during a run sends only the delete', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await page.getByRole('button', { name: 'Thread options' }).click()
    await page.getByRole('menuitem', { name: 'Delete thread' }).click()
    await page
      .getByRole('alertdialog')
      .getByRole('button', { name: 'Delete' })
      .click()

    await expect(page).toHaveURL(/\/chat\?agent=streamer-id$/)
    await expect(toasts(page)).toHaveText(['Thread deleted'])
    expect(fixture.sessionCalls.deletes).toEqual([
      { sessionId: 'agui-t-trip', appName: 'agui' },
    ])
    // The page only detached from the run; the server's delete stops it.
    await expect.poll(() => fixture.abortedRuns).toEqual(['t-trip'])
    expect(fixture.stopRequests).toEqual([])
    expect(fixture.requests).toHaveLength(1)
  })
})

// A run's own stream observes its log as an attach does, and ends with the
// butter.fallback marker when the log cannot carry the run to its end
// (docs/api.md "The fallback marker"). The run goes on, so the page waits it
// out by reading the thread, and the read after it brings the reply.
test.describe('Chat runs whose own stream falls back', () => {
  const RUN = { runId: 'r1', invocationId: 'inv-1' }
  const THREAD = '/api/agui/streamer-id/threads/t-trip'
  const before = [
    { id: 'u1', role: 'user', content: 'Book a flight to Lisbon' },
    { id: 'a1', role: 'assistant', content: 'Booked the 08:10 flight.' },
  ]
  const turn = { id: 'u2', role: 'user', content: 'Now plan the trip' }
  const PLAN = 'Three days in Lisbon: Alfama, Belém and Sintra.'

  // thread is t-trip's history as the server reads it.
  const thread = (
    messages: Array<Record<string, unknown>>,
    extra: Record<string, unknown> = {}
  ) => ({
    body: {
      threadId: 't-trip',
      messages,
      interrupts: [],
      surfaces: [],
      ...extra,
    },
  })

  // fellBack is the stream of a run whose log stopped short: the start of
  // its reply, then the marker in place of the run's end.
  const fellBack = (text: string) =>
    sse([
      { type: 'RUN_STARTED', threadId: 't-trip', runId: RUN.runId },
      { type: 'STATE_SNAPSHOT', snapshot: {} },
      ...reply(RUN.runId, text).slice(0, 2),
      {
        type: 'CUSTOM',
        name: 'butter.fallback',
        value: { threadId: 't-trip', runId: RUN.runId, reason: 'truncated' },
      },
    ])

  const replies = (page: Page) =>
    page.locator('[data-message-role="assistant"]')
  const sentTurns = (page: Page) => page.locator('[data-message-role="user"]')
  // threadReads counts the reads of t-trip's history.
  const threadReads = (fixture: AGUIFixture) =>
    fixture.historyRequests.filter((url) => url.includes(THREAD)).length

  // openTripAndSend opens t-trip, then sends the turn that starts the run.
  // From then on the server's reads show the run holding the thread.
  async function openTripAndSend(page: Page, fixture: AGUIFixture) {
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
    fixture.historyByThread!['t-trip'] = thread([...before, turn], {
      running: RUN,
    })
    await send(page, 'Now plan the trip')
  }

  test('the page waits the run out, showing what streamed, and the run’s reply arrives', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [fellBack('Three days')],
      sessions: [trip()],
      historyByThread: { 't-trip': thread(before) },
    })
    await openTripAndSend(page, fixture)

    // The reply keeps what streamed, as running, and the composer waits.
    await expect(replies(page).nth(1)).toContainText('Three days')
    await expect(running(page)).toBeVisible()
    await expect(stopButton(page)).toBeEnabled()
    await expect(composer(page)).toBeDisabled()
    // The page reads the thread, which the run still holds.
    await expect.poll(() => threadReads(fixture)).toBe(2)
    await expect(running(page)).toBeVisible()

    // The run ends; the read after it shows its reply in place of what
    // streamed.
    fixture.historyByThread!['t-trip'] = thread([
      ...before,
      turn,
      { id: 'a2', role: 'assistant', content: PLAN },
    ])
    await expect(page.getByText(PLAN)).toBeVisible({ timeout: 10_000 })
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    await expect(replies(page)).toHaveCount(2)
    await expect(sentTurns(page)).toHaveCount(2)
    expect(
      (await replies(page).nth(1).innerText()).split('Three days')
    ).toHaveLength(2)

    // One run, waited out by reading the thread: the log is not attached
    // again, and nothing failed.
    expect(threadReads(fixture)).toBe(3)
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.attachRequests).toEqual([])
    expect(fixture.stopRequests).toEqual([])
    await expect(toasts(page)).toHaveCount(0)
  })

  test('the first run of a thread started from a new chat is waited out too', async ({
    page,
  }) => {
    const first = { id: 'u1', role: 'user', content: 'Plan a trip to Lisbon' }
    const fixture = await setupAGUI(page, {
      runs: [fellBack('Three days')],
      histories: [
        thread([first], { running: RUN }),
        thread([first, { id: 'a1', role: 'assistant', content: PLAN }]),
      ],
    })
    await page.goto('/chat?agent=streamer-id', { waitUntil: 'networkidle' })
    await send(page, 'Plan a trip to Lisbon')

    await expect(replies(page)).toContainText(['Three days'])
    await expect(running(page)).toBeVisible()
    await expect(composer(page)).toBeDisabled()

    await expect(page.getByText(PLAN)).toBeVisible({ timeout: 10_000 })
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    await expect(replies(page)).toHaveCount(1)
    await expect(sentTurns(page)).toHaveCount(1)
    expect(threadInURL(page)).toBe(fixture.requests[0].threadId)
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.historyRequests).toHaveLength(2)
    expect(fixture.attachRequests).toEqual([])
    await expect(toasts(page)).toHaveCount(0)
  })

  test('Stop while the page waits stops the run, and the stopped run shows', async ({
    page,
  }) => {
    const accepted = deferred()
    const fixture = await setupAGUI(page, {
      runs: [fellBack('Day one')],
      sessions: [trip()],
      historyByThread: { 't-trip': thread(before) },
      stops: [
        {
          status: 202,
          body: { threadId: 't-trip', ...RUN },
          until: accepted.promise,
        },
      ],
    })
    await openTripAndSend(page, fixture)
    await expect(running(page)).toBeVisible()
    await expect(composer(page)).toBeDisabled()

    await stopButton(page).click()
    await expect
      .poll(() => stopPaths(fixture.stopRequests))
      .toEqual([`${THREAD}/stop`])
    // Until the thread shows the run stopped, it goes on showing here.
    await expect(running(page)).toBeVisible()

    // The Stop reaches the run, which ends with what it stored so far.
    fixture.historyByThread!['t-trip'] = thread(
      [...before, turn, { id: 'a2', role: 'assistant', content: 'Day one.' }],
      {
        lastRun: {
          status: 'cancelled',
          error: 'stopped by user',
          input: 'Now plan the trip',
        },
      }
    )
    accepted.resolve()

    await expect(stoppedNotice(page)).toBeVisible({ timeout: 2_500 })
    await expect(stoppedNotice(page)).toContainText('Now plan the trip')
    await expect(page.getByText('Day one.')).toBeVisible()
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    await expect(toasts(page)).toHaveCount(0)
    expect(fixture.stopRequests).toHaveLength(1)
    expect(fixture.requests).toHaveLength(1)
  })
})
