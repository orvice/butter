import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage as send,
  setupAGUI,
  sse,
  stoppedRun,
  type SnapshotResponse,
} from './support/agui'

// A run that fails or is stopped stays visible in Chat (ADR-0016
// decision 8): a notice under the conversation says how it ended and quotes
// the turn it started from, and Restore input puts that turn back in the
// composer. The notice comes from the run's RUN_ERROR, from a Stop, or,
// after a reload, from the thread history's lastRun. A later run that
// succeeds clears it.

// A 1×1 PNG.
const PNG =
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' })

const failureNotice = (page: Page) =>
  page.getByRole('alert').filter({ hasText: 'This run failed' })
const stoppedNotice = (page: Page) =>
  page.getByRole('status').filter({ hasText: 'Stopped' })
const restoreButton = (page: Page) =>
  page.getByRole('button', { name: 'Restore input' })
const composer = (page: Page) => page.getByRole('textbox', { name: /^Message/ })
const running = (page: Page) => page.getByText('Running…')
const stopButton = (page: Page) => page.getByRole('button', { name: 'Stop' })
const sentMessages = (page: Page) => page.locator('[data-message-role="user"]')
const toasts = (page: Page) => page.locator('[data-sonner-toast]')

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

// failed is a run that streams some text, then fails.
function failed(runId: string, error: string, text = 'Looking into it') {
  return sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    ...reply(runId, text),
    { type: 'RUN_ERROR', message: error, runId },
  ])
}

function pausedOn(runId: string, text: string, question: string) {
  return sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    ...reply(runId, text),
    {
      type: 'RUN_FINISHED',
      threadId: 't',
      runId,
      outcome: {
        type: 'interrupt',
        interrupts: [{ id: 'int-1', reason: 'human_input', message: question }],
      },
    },
  ])
}

// The parts of a user turn with images, as a run's request and the thread
// history carry them.
const textPart = (text: string) => ({ type: 'text', text })
const imagePart = () => ({
  type: 'image',
  source: { type: 'data', value: PNG, mimeType: 'image/png' },
})

// threadHistory is the thread t-trip as the server rebuilds it, with the
// lastRun it reports.
function threadHistory(
  messages: Array<Record<string, unknown>>,
  lastRun?: { status: string; error: string; input: string }
): SnapshotResponse {
  return {
    body: {
      threadId: 't-trip',
      messages,
      interrupts: [],
      surfaces: [],
      ...(lastRun ? { lastRun } : {}),
    },
  }
}

// pick attaches PNGs through the composer's file picker.
async function pick(page: Page, ...names: string[]) {
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Attach images' }).click()
  await (
    await chooser
  ).setFiles(
    names.map((name) => ({
      name,
      mimeType: 'image/png',
      buffer: Buffer.from(PNG, 'base64'),
    }))
  )
}

// deferred is a promise the test settles.
function deferred() {
  let resolve!: () => void
  const promise = new Promise<void>((r) => (resolve = r))
  return { promise, resolve }
}

test.describe('Chat failed and stopped runs', () => {
  test('a failed run shows the failure with its input, and Restore input refills the composer', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [failed('r1', 'model exploded'), { open: true }],
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await pick(page, 'map.png')
    await send(page, 'plan the trip')

    const notice = failureNotice(page)
    await expect(notice).toBeVisible()
    await expect(notice).toContainText('model exploded')
    await expect(notice).toContainText('plan the trip · 1 image')
    await expect(notice).toContainText(
      'Sending again starts a new run and may repeat external tool actions.'
    )
    // What the run streamed before it failed stays.
    await expect(page.getByText('Looking into it')).toBeVisible()
    // The composer is left alone until the input is restored.
    await expect(composer(page)).toHaveValue('')

    await notice.getByRole('button', { name: 'Restore input' }).click()
    await expect(composer(page)).toHaveValue('plan the trip')
    await expect(
      page.getByRole('button', { name: 'Remove image-1.png' })
    ).toBeVisible()

    // Sending it again starts a new run with the same turn. The notice is
    // gone while that run goes on, and stays gone once it succeeds.
    await composer(page).press('Enter')
    await expect(running(page)).toBeVisible()
    await expect(failureNotice(page)).toHaveCount(0)
    expect(fixture.requests).toHaveLength(2)
    expect(fixture.requests[1].messages).toEqual([
      {
        id: expect.any(String),
        role: 'user',
        content: [textPart('plan the trip'), imagePart()],
      },
    ])
    fixture.endOpenRun('t-trip', finished('r2', 'Here is the plan.'))
    await expect(page.getByText('Here is the plan.')).toBeVisible()
    await expect(running(page)).toHaveCount(0)
    await expect(failureNotice(page)).toHaveCount(0)
  })

  test('after a reload the failure comes back from lastRun, with the turn the history keeps', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [failed('r1', 'model exploded')],
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await pick(page, 'map.png')
    await send(page, 'plan the trip')
    await expect(failureNotice(page)).toBeVisible()

    // The thread as the server rebuilds it: the turn the run stored, what it
    // replied before it failed, and how it ended.
    fixture.historyByThread = {
      't-trip': threadHistory(
        [
          {
            id: 'u1',
            role: 'user',
            content: [textPart('plan the trip'), imagePart()],
          },
          { id: 'a1', role: 'assistant', content: 'Looking into it' },
        ],
        { status: 'failed', error: 'model exploded', input: 'plan the trip' }
      ),
    }
    await page.reload({ waitUntil: 'networkidle' })

    const notice = failureNotice(page)
    await expect(notice).toBeVisible()
    await expect(notice).toContainText('model exploded')
    await expect(notice).toContainText('plan the trip · 1 image')
    await expect(page.getByText('Looking into it')).toBeVisible()

    // The images come back with the history, so they are restored too.
    await restoreButton(page).click()
    await expect(composer(page)).toHaveValue('plan the trip')
    await expect(page.getByRole('img', { name: 'image-1.png' })).toBeVisible()
  })

  test('a run that never stored its turn restores the text alone', async ({
    page,
  }) => {
    await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: {
        't-trip': threadHistory([], {
          status: 'failed',
          error:
            'interrupted by a service shutdown; no work was replayed automatically.',
          input: 'plan the trip',
        }),
      },
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })

    const notice = failureNotice(page)
    await expect(notice).toContainText('interrupted by a service shutdown')
    await expect(notice).toContainText('plan the trip')
    await expect(notice).not.toContainText('image')
    await restoreButton(page).click()
    await expect(composer(page)).toHaveValue('plan the trip')
    await expect(page.getByRole('button', { name: /^Remove / })).toHaveCount(0)
  })

  test('Stop shows the stopped notice, and Restore input refills the composer', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [{ open: true }], sessions: [trip()] })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).click()
    const notice = stoppedNotice(page)
    await expect(notice).toBeVisible()
    await expect(notice).toContainText(
      'You stopped this response before it finished.'
    )
    await expect(notice).toContainText('plan the trip')
    // A stopped run did not fail.
    await expect(failureNotice(page)).toHaveCount(0)
    await expect(toasts(page)).toHaveCount(0)

    // When the local cancel ends the run before its stopped stream does,
    // the runtime takes a turn with no reply yet back into the composer.
    // Restore input refills the composer whatever it holds.
    await composer(page).fill('something else')
    await notice.getByRole('button', { name: 'Restore input' }).click()
    await expect(composer(page)).toHaveValue('plan the trip')
  })

  test('a Stop that ends the run here shows it stopped, though its stream never says so', async ({
    page,
  }) => {
    // 204: the server has no run in flight to stop, so only the page ends
    // the stream, and no stopped RUN_ERROR arrives.
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
      stops: [{ status: 204 }],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    await stopButton(page).click()
    await expect(stoppedNotice(page)).toContainText('plan the trip')
    await expect.poll(() => fixture.abortedRuns).toEqual(['t-trip'])
    await expect(failureNotice(page)).toHaveCount(0)
  })

  test('a stopped stream that ends before the Stop is answered shows the stopped notice', async ({
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
    const runId = String(fixture.requests[0].runId)
    fixture.endOpenRun('t-trip', stoppedRun('t-trip', runId))
    await expect(stoppedNotice(page)).toContainText('plan the trip')

    const answered = page.waitForResponse((r) => r.url().endsWith('/stop'))
    accepted.resolve()
    expect((await answered).status()).toBe(202)
    await expect(stoppedNotice(page)).toBeVisible()
    await expect(failureNotice(page)).toHaveCount(0)
  })

  test('a run stopped elsewhere shows the stopped notice, and so does its lastRun after a reload', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [{ open: true }],
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'plan the trip')
    await expect(running(page)).toBeVisible()

    // Another tab's Stop: the stream ends with the stop code.
    const runId = String(fixture.requests[0].runId)
    fixture.endOpenRun('t-trip', stoppedRun('t-trip', runId))
    await expect(stoppedNotice(page)).toContainText('plan the trip')
    expect(fixture.stopRequests).toEqual([])

    fixture.historyByThread = {
      't-trip': threadHistory(
        [{ id: 'u1', role: 'user', content: 'plan the trip' }],
        {
          status: 'cancelled',
          error: 'stopped by user',
          input: 'plan the trip',
        }
      ),
    }
    await page.reload({ waitUntil: 'networkidle' })
    await expect(stoppedNotice(page)).toContainText(
      'You stopped this response before it finished.'
    )
    await expect(failureNotice(page)).toHaveCount(0)
    await restoreButton(page).click()
    await expect(composer(page)).toHaveValue('plan the trip')
  })

  test('a later run that succeeds clears a notice from lastRun', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [finished('r2', 'Here is the plan.')],
      sessions: [trip()],
      historyByThread: {
        't-trip': threadHistory(
          [{ id: 'u1', role: 'user', content: 'plan the trip' }],
          { status: 'failed', error: 'model exploded', input: 'plan the trip' }
        ),
      },
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await expect(failureNotice(page)).toBeVisible()

    await send(page, 'try again')
    await expect(page.getByText('Here is the plan.')).toBeVisible()
    await expect(failureNotice(page)).toHaveCount(0)

    // The server no longer reports the failure, and a reload shows none.
    fixture.historyByThread = {
      't-trip': threadHistory([
        { id: 'u1', role: 'user', content: 'plan the trip' },
        { id: 'u2', role: 'user', content: 'try again' },
        { id: 'a2', role: 'assistant', content: 'Here is the plan.' },
      ]),
    }
    await page.reload({ waitUntil: 'networkidle' })
    await expect(page.getByText('Here is the plan.')).toBeVisible()
    await expect(failureNotice(page)).toHaveCount(0)
    await expect(sentMessages(page)).toHaveCount(2)
  })

  test('a failed answer to a question has no input to restore', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        pausedOn('r1', 'One thing first.', 'Which version?'),
        failed('r2', 'deploy tool unavailable', 'Deploying'),
      ],
      sessions: [trip()],
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await send(page, 'deploy')
    const question = page.getByRole('group', { name: 'Which version?' })
    await question.getByPlaceholder('Type your answer…').fill('2.4.1')
    await question.getByRole('button', { name: 'Answer' }).click()

    // The answer went as a resume entry, which the server records no input
    // for: the notice quotes nothing and offers nothing to put back.
    const notice = failureNotice(page)
    await expect(notice).toContainText('deploy tool unavailable')
    expect(fixture.requests[1].resume).toHaveLength(1)
    await expect(notice).not.toContainText('2.4.1')
    await expect(restoreButton(page)).toHaveCount(0)
  })
})
