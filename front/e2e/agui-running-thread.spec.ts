import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage,
  setupAGUI,
  sse,
  type SnapshotResponse,
} from './support/agui'

// A thread can be held by a run this page did not start: one started before
// a reload, or in another tab (ADR-0016 decision 8). The thread's reads then
// answer at once with the run (`running`) and the thread up to the turn that
// started it. AG-UI Chat attaches to the run's log and streams the run
// (agui-live-reattach.spec.ts); here attaching answers 204, as it does for a
// run whose log is gone, so the page falls back to waiting the run out. It
// shows that run as running under its turn, with the composer disabled,
// reads the thread again with backoff until no run holds it, then shows the
// thread as the run left it, cards included. A read is never tried again on
// a 409.

const V = 'v0.9.1'
const RUN = { runId: 'run-2', invocationId: 'inv-2' }
const THREAD = '/api/agui/streamer-id/threads/t-trip'

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1)

// before is the thread as the run found it: the conversation so far, and the
// turn that started the run.
const before = [
  { id: 'u1', role: 'user', content: 'Book a flight to Lisbon' },
  { id: 'a1', role: 'assistant', content: 'Booked the 08:10 flight.' },
  { id: 'u2', role: 'user', content: 'Now plan the trip' },
]

const PLAN = 'Three days in Lisbon: Alfama, Belém and Sintra.'

function history(
  messages: Array<Record<string, unknown>>,
  extra: Record<string, unknown> = {}
): SnapshotResponse {
  return {
    body: {
      threadId: 't-trip',
      messages,
      interrupts: [],
      surfaces: [],
      ...extra,
    },
  }
}

// duringRun is what the history answers while the run goes on.
const duringRun = () => history(before, { running: RUN })

// afterRun is the history once the run ended, with its reply.
const afterRun = (extra: Record<string, unknown> = {}) =>
  history(
    [...before, { id: 'a2', role: 'assistant', content: PLAN }],
    extra
  )

function snapshot(
  surfaces: Array<Record<string, unknown>> = [],
  extra: Record<string, unknown> = {}
): SnapshotResponse {
  return {
    body: {
      version: V,
      catalogId: 'butter-basic-v1',
      threadId: 't-trip',
      surfaces,
      ...extra,
    },
  }
}

// planCard is the Result Card the run created, in its reply a2.
const planCard = {
  surfaceId: 'card-plan',
  kind: 'card',
  revision: 1,
  runId: 'run-2',
  messageId: 'a2',
  fallback: 'Lisbon itinerary (text)',
  envelopes: [
    {
      version: V,
      createSurface: { surfaceId: 'card-plan', catalogId: 'butter-basic-v1' },
    },
    {
      version: V,
      updateComponents: {
        surfaceId: 'card-plan',
        components: [
          { id: 'root', component: 'Card', child: 'col' },
          { id: 'col', component: 'Column', children: ['title', 'days'] },
          {
            id: 'title',
            component: 'Text',
            text: 'Lisbon itinerary',
            variant: 'h3',
          },
          { id: 'days', component: 'KeyValue', label: 'Days', value: '3' },
        ],
      },
    },
  ],
}

const running = (page: Page) => page.getByText('Running…')
const composer = (page: Page) => page.getByRole('textbox', { name: /^Message/ })
const stopButton = (page: Page) => page.getByRole('button', { name: 'Stop' })
const sendButton = (page: Page) => page.getByRole('button', { name: 'Send' })
const attachButton = (page: Page) =>
  page.getByRole('button', { name: 'Attach images' })
const replies = (page: Page) =>
  page.locator('[data-message-role="assistant"]')
const toasts = (page: Page) => page.locator('[data-sonner-toast]')
const loadFailed = (page: Page) => page.getByText('Failed to load this thread.')
const loadingSkeleton = (page: Page) =>
  page.getByRole('status', { name: 'Loading conversation' })
// The notice of a run that failed or was stopped (#408).
const failureNotice = (page: Page) =>
  page.getByRole('alert').filter({ hasText: 'This run failed' })
const stoppedNotice = (page: Page) =>
  page.getByRole('status').filter({ hasText: 'Stopped' })

const pathsOf = (urls: string[]) => urls.map((u) => new URL(u).pathname)

// The page reads the thread again a second after it found the run, then two
// seconds later: the end of a run that answers twice as running shows about
// three seconds in.
const AFTER_TWO_READS = { timeout: 10_000 }

// deferred is a promise the test settles.
function deferred() {
  let resolve!: () => void
  const promise = new Promise<void>((r) => (resolve = r))
  return { promise, resolve }
}

test.describe('AG-UI Chat opening a thread during a run', () => {
  test('shows the run running with the composer disabled, then its reply with the composer enabled', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        sse([
          { type: 'RUN_STARTED', threadId: 't-trip', runId: 'r3' },
          { type: 'TEXT_MESSAGE_START', messageId: 'a3', role: 'assistant' },
          {
            type: 'TEXT_MESSAGE_CONTENT',
            messageId: 'a3',
            delta: 'Hotels are next.',
          },
          { type: 'TEXT_MESSAGE_END', messageId: 'a3' },
          { type: 'RUN_FINISHED', threadId: 't-trip', runId: 'r3' },
        ]),
      ],
      sessions: [trip()],
      histories: [duringRun(), duringRun(), afterRun()],
      snapshots: [snapshot([], { running: RUN }), snapshot()],
    })
    await page.goto('/agui-chat?thread=t-trip')

    // The thread as the run found it, ending with the turn that started it.
    await expect(page.getByText('Booked the 08:10 flight.')).toBeVisible()
    await expect(
      page.getByText('Now plan the trip', { exact: true })
    ).toBeVisible()
    // The run shows as running where its reply will appear, and nothing can
    // be sent meanwhile.
    await expect(running(page)).toBeVisible()
    await expect(replies(page)).toHaveCount(2)
    await expect(loadingSkeleton(page)).toHaveCount(0)
    await expect(composer(page)).toBeDisabled()
    await expect(attachButton(page)).toBeDisabled()
    await expect(stopButton(page)).toBeEnabled()
    await expect(sendButton(page)).toHaveCount(0)
    await expect.poll(() => fixture.historyRequests.length).toBe(2)
    await expect(running(page)).toBeVisible()
    await expect(composer(page)).toBeDisabled()

    // The third read finds the run ended: its reply takes the running one's
    // place.
    await expect(page.getByText(PLAN)).toBeVisible(AFTER_TWO_READS)
    await expect(running(page)).toHaveCount(0)
    await expect(replies(page)).toHaveCount(2)
    await expect(composer(page)).toBeEnabled()
    await expect(attachButton(page)).toBeEnabled()
    await expect(sendButton(page)).toBeVisible()
    expect(pathsOf(fixture.historyRequests)).toEqual([
      `${THREAD}/messages`,
      `${THREAD}/messages`,
      `${THREAD}/messages`,
    ])
    expect(pathsOf(fixture.snapshotRequests)).toEqual([
      `${THREAD}/ui`,
      `${THREAD}/ui`,
    ])
    // Waiting a run out neither runs the agent nor stops the run.
    expect(fixture.requests).toEqual([])
    expect(fixture.stopRequests).toEqual([])
    await expect(toasts(page)).toHaveCount(0)

    // The thread goes on as usual.
    await sendMessage(page, 'And hotels?')
    await expect(page.getByText('Hotels are next.')).toBeVisible()
    expect(fixture.requests).toHaveLength(1)
    expect(
      (fixture.requests[0].messages as Array<Record<string, unknown>>).at(-1)
    ).toMatchObject({ role: 'user', content: 'And hotels?' })
    await expect(replies(page)).toHaveCount(3)
  })

  test('a Result Card the run created shows in its reply once the run ended', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      histories: [
        duringRun(),
        afterRun({ surfaces: [{ surfaceId: 'card-plan', messageId: 'a2' }] }),
      ],
      snapshots: [snapshot([], { running: RUN }), snapshot([planCard])],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()
    // The card is the run's: the reads during the run leave it out.
    await expect(page.locator('[data-a2ui-surface]')).toHaveCount(0)

    const reply = replies(page).filter({ hasText: PLAN })
    await expect(reply).toBeVisible(AFTER_TWO_READS)
    await expect(
      reply.locator('[data-a2ui-surface="card-plan"]')
    ).toBeVisible()
    await expect(reply.getByText('Lisbon itinerary')).toBeVisible()
    // It shows in the reply that produced it, not on its own.
    await expect(
      page.getByRole('region', { name: 'Restored from this conversation' })
    ).toHaveCount(0)
    await expect(running(page)).toHaveCount(0)
    expect(fixture.snapshotRequests).toHaveLength(2)
  })

  test('a run that ends on a question shows it once the run ended, and it can be answered', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        sse([
          { type: 'RUN_STARTED', threadId: 't-trip', runId: 'r3' },
          { type: 'TEXT_MESSAGE_START', messageId: 'a3', role: 'assistant' },
          {
            type: 'TEXT_MESSAGE_CONTENT',
            messageId: 'a3',
            delta: 'Booked the Lumiares.',
          },
          { type: 'TEXT_MESSAGE_END', messageId: 'a3' },
          { type: 'RUN_FINISHED', threadId: 't-trip', runId: 'r3' },
        ]),
      ],
      sessions: [trip()],
      histories: [
        duringRun(),
        afterRun({
          interrupts: [
            { id: 'int-1', reason: 'human_input', message: 'Which hotel?' },
          ],
        }),
      ],
      snapshots: [snapshot([], { running: RUN }), snapshot()],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()
    const question = page.getByRole('group', { name: 'Which hotel?' })
    await expect(question).toHaveCount(0)

    await expect(question).toBeVisible(AFTER_TWO_READS)
    await expect(page.getByText(PLAN)).toBeVisible()
    await expect(
      page.getByText('Sending answers the earliest open question.')
    ).toBeVisible()
    await question.getByPlaceholder('Type your answer…').fill('The Lumiares')
    await question.getByRole('button', { name: 'Answer' }).click()
    await expect(page.getByText('Booked the Lumiares.')).toBeVisible()
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.requests[0].resume).toEqual([
      { interruptId: 'int-1', status: 'resolved', payload: 'The Lumiares' },
    ])
  })

  test('Stop sends one stop request, and the page settles on the stopped run', async ({
    page,
  }) => {
    const accepted = deferred()
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot([], { running: RUN })],
      stops: [
        {
          status: 202,
          body: { threadId: 't-trip', ...RUN },
          until: accepted.promise,
        },
      ],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).dblclick()
    await expect(stopButton(page)).toBeDisabled()
    await expect
      .poll(() => pathsOf(fixture.stopRequests))
      .toEqual([`${THREAD}/stop`])
    // Until the history shows the run stopped, it goes on showing here.
    await expect(running(page)).toBeVisible()
    await expect(stoppedNotice(page)).toHaveCount(0)

    // The Stop reaches the run, which ends with what it stored so far.
    fixture.historyByThread!['t-trip'] = history(
      [...before, { id: 'a2', role: 'assistant', content: 'Day one: Alfama.' }],
      {
        lastRun: {
          status: 'cancelled',
          error: 'stopped by user',
          input: 'Now plan the trip',
        },
      }
    )
    const reads = fixture.historyRequests.length
    accepted.resolve()

    // The page reads the thread again at once, not after its backoff.
    await expect(page.getByText('Day one: Alfama.')).toBeVisible({
      timeout: 1_500,
    })
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    await expect(sendButton(page)).toBeVisible()
    expect(fixture.historyRequests.length).toBeGreaterThan(reads)
    expect(fixture.stopRequests).toHaveLength(1)
    // The read that showed the end says the run was stopped (#408), with
    // the turn it started from: no reload needed.
    await expect(stoppedNotice(page)).toBeVisible()
    await expect(stoppedNotice(page)).toContainText('Now plan the trip')
    // A stopped run is not a failure, and the page started no run.
    await expect(toasts(page)).toHaveCount(0)
    await expect(failureNotice(page)).toHaveCount(0)
    expect(fixture.requests).toEqual([])
  })

  test('a run that fails while the page waits it out shows the failure once it ended', async ({
    page,
  }) => {
    await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      histories: [
        duringRun(),
        history(before, {
          lastRun: {
            status: 'failed',
            error: 'model exploded',
            input: 'Now plan the trip',
          },
        }),
      ],
      snapshots: [snapshot([], { running: RUN })],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()
    await expect(failureNotice(page)).toHaveCount(0)

    const notice = failureNotice(page)
    await expect(notice).toBeVisible(AFTER_TWO_READS)
    await expect(notice).toContainText('model exploded')
    await expect(notice).toContainText('Now plan the trip')
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    // Restore input puts the turn back in the composer.
    await notice.getByRole('button', { name: 'Restore input' }).click()
    await expect(composer(page)).toHaveValue('Now plan the trip')
  })

  test('a Stop the server refuses leaves the run going, and the page still settles on its end', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot([], { running: RUN })],
      stops: [
        { status: 503, body: { error: 'stop unavailable, retry later' } },
      ],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).click()
    await expect(
      toasts(page).filter({
        hasText: 'Could not stop the run: stop unavailable, retry later',
      })
    ).toBeVisible()
    await expect(stopButton(page)).toBeEnabled()
    await expect(running(page)).toBeVisible()
    await expect(composer(page)).toBeDisabled()

    // The run ends on its own.
    fixture.historyByThread!['t-trip'] = afterRun()
    await expect(page.getByText(PLAN)).toBeVisible(AFTER_TWO_READS)
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    expect(fixture.stopRequests).toHaveLength(1)
  })

  test('leaving a thread the page waits out stops reading it', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [
        trip(),
        aguiSession('t-budget', 'Budget review', { agentId: 'streamer-id' }, 5),
      ],
      historyByThread: {
        't-trip': duringRun(),
        't-budget': {
          body: {
            threadId: 't-budget',
            messages: [
              { id: 'b1', role: 'user', content: 'Review the budget' },
              { id: 'b2', role: 'assistant', content: 'The budget holds.' },
            ],
            interrupts: [],
            surfaces: [],
          },
        },
      },
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()
    await expect.poll(() => fixture.historyRequests.length).toBe(2)

    await page
      .getByRole('navigation', { name: 'AG-UI threads' })
      .getByRole('link', { name: 'Budget review', exact: true })
      .click()
    await expect(page.getByText('The budget holds.')).toBeVisible()
    await expect(running(page)).toHaveCount(0)
    await expect(composer(page)).toBeEnabled()
    const tripReads = () =>
      pathsOf(fixture.historyRequests).filter((p) => p.includes('/t-trip/'))
        .length
    const left = tripReads()
    // The next read of the left thread would have gone out two seconds
    // after the last one.
    await page.waitForTimeout(3_000)
    expect(tripReads()).toBe(left)
    expect(fixture.stopRequests).toEqual([])
  })

  test('a read refused with 409 is not tried again', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      histories: [
        { status: 409, body: { error: 'a run is in progress on this thread' } },
        afterRun(),
      ],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(loadFailed(page)).toBeVisible()
    await expect(
      page.getByText('a run is in progress on this thread')
    ).toBeVisible()
    // Nothing reads the thread again on its own: the page used to, with
    // backoff from half a second.
    await page.waitForTimeout(1_500)
    expect(fixture.historyRequests).toHaveLength(1)

    // Retry reads it again.
    await page.getByRole('button', { name: 'Retry' }).click()
    await expect(page.getByText(PLAN)).toBeVisible()
    expect(fixture.historyRequests).toHaveLength(2)
  })

  test('a 409 while the page waits out a run is not tried again', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      histories: [
        duringRun(),
        { status: 409, body: { error: 'a run is in progress on this thread' } },
        afterRun(),
      ],
      snapshots: [snapshot([], { running: RUN })],
    })
    await page.goto('/agui-chat?thread=t-trip')
    await expect(running(page)).toBeVisible()
    await expect(loadFailed(page)).toBeVisible(AFTER_TWO_READS)
    await expect(
      page.getByText('a run is in progress on this thread')
    ).toBeVisible()
    await page.waitForTimeout(2_500)
    expect(fixture.historyRequests).toHaveLength(2)
  })
})
