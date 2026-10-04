import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage as send,
  setupAGUI,
  sse,
  stoppedRun,
  threadInURL,
} from './support/agui'

// Every AG-UI Chat run asks to detach (ADR-0016), so it outlives its
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
    .getByRole('navigation', { name: 'AG-UI threads' })
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

test.describe('AG-UI Chat detached runs', () => {
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
    await page.goto('/agui-chat?agent=streamer-id', {
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    // Another thread.
    await threadLink(page, 'Budget review').click()
    await expect(page).toHaveURL(/\/agui-chat\?thread=t-budget$/)
    await expect.poll(() => fixture.abortedRuns).toEqual(['t-trip'])
    await expect(running(page)).toHaveCount(0)

    // New thread.
    await send(page, 'review the budget')
    await expect(running(page)).toBeVisible()
    await page.getByRole('link', { name: 'New thread' }).click()
    await expect(page).toHaveURL(/\/agui-chat\?agent=streamer-id$/)
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
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
    await page.goto('/agui-chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await page.getByRole('button', { name: 'Thread options' }).click()
    await page.getByRole('menuitem', { name: 'Delete thread' }).click()
    await page
      .getByRole('alertdialog')
      .getByRole('button', { name: 'Delete' })
      .click()

    await expect(page).toHaveURL(/\/agui-chat\?agent=streamer-id$/)
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
