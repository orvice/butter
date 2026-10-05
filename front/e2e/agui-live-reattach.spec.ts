import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage,
  setupAGUI,
  sse,
  stoppedEvent,
  type AGUIFixture,
  type AttachStream,
  type SnapshotResponse,
} from './support/agui'

// A thread a run holds when the page opens it, after a reload, from another
// tab or on coming back to it (ADR-0016 decision 8): the reads show the
// thread up to the turn that started the run, and the page attaches to the
// run's log (GET …/threads/:thread_id/run). The log replays the run from
// RUN_STARTED, then follows it, and the page streams the run's reply from it
// as it streams a run it started: text, tool calls, cards, the shared state,
// and the run's end. When the log cannot be followed (204, or the
// butter.fallback marker) the page waits the run out by reading the thread.

const V = 'v0.9.1'
const RUN = { runId: 'run-2', invocationId: 'inv-2' }
const THREAD = '/api/agui/streamer-id/threads/t-trip'
const MSG = 'msg-run-2'

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' }, 1)
const budget = () =>
  aguiSession('t-budget', 'Budget review', { agentId: 'streamer-id' }, 5)

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

const duringRun = () => history(before, { running: RUN })
const afterRun = () =>
  history([...before, { id: 'a2', role: 'assistant', content: PLAN }])

function snapshot(extra: Record<string, unknown> = {}): SnapshotResponse {
  return {
    body: {
      version: V,
      catalogId: 'butter-basic-v1',
      threadId: 't-trip',
      surfaces: [],
      ...extra,
    },
  }
}

// The run's log, as the server writes it: a detached run opens with
// RUN_STARTED and a STATE_SNAPSHOT, and gives its reply one message ID.
const runStarted = (state: Record<string, unknown> = {}) => [
  { type: 'RUN_STARTED', threadId: 't-trip', runId: RUN.runId },
  { type: 'STATE_SNAPSHOT', snapshot: state },
]
const textStart = {
  type: 'TEXT_MESSAGE_START',
  messageId: MSG,
  role: 'assistant',
}
const delta = (text: string) => ({
  type: 'TEXT_MESSAGE_CONTENT',
  messageId: MSG,
  delta: text,
})
const textEnd = { type: 'TEXT_MESSAGE_END', messageId: MSG }
const runFinished = (
  outcome: Record<string, unknown> = { type: 'success' }
) => ({
  type: 'RUN_FINISHED',
  threadId: 't-trip',
  runId: RUN.runId,
  outcome,
})
const toolCall = (id: string, name: string, args: unknown) => [
  { type: 'TOOL_CALL_START', toolCallId: id, toolCallName: name },
  { type: 'TOOL_CALL_ARGS', toolCallId: id, delta: JSON.stringify(args) },
  { type: 'TOOL_CALL_END', toolCallId: id },
]
const toolResult = (id: string, result: unknown) => ({
  type: 'TOOL_CALL_RESULT',
  messageId: MSG,
  toolCallId: id,
  content: JSON.stringify(result),
  role: 'tool',
})

function a2ui(
  surfaceId: string,
  kind: 'card' | 'form',
  revision: number,
  seq: number,
  envelope: Record<string, unknown>,
  extra: Record<string, unknown> = {}
) {
  return {
    type: 'CUSTOM',
    name: 'butter.a2ui',
    value: {
      version: V,
      surfaceId,
      kind,
      revision,
      seq,
      threadId: 't-trip',
      runId: RUN.runId,
      messageId: MSG,
      envelope: { version: V, ...envelope },
      ...extra,
    },
  }
}

// cardCreated is a Result Card's creation, as a render_ui write streams it:
// a card titled title, with one key and value.
const cardCreated = (
  surfaceId: string,
  title: string,
  label: string,
  value: string
) => [
  a2ui(surfaceId, 'card', 1, 0, {
    createSurface: { surfaceId, catalogId: 'butter-basic-v1' },
  }),
  a2ui(surfaceId, 'card', 1, 1, {
    updateComponents: {
      surfaceId,
      components: [
        { id: 'root', component: 'Card', child: 'col' },
        { id: 'col', component: 'Column', children: ['title', 'kv'] },
        { id: 'title', component: 'Text', text: title, variant: 'h3' },
        { id: 'kv', component: 'KeyValue', label, value },
      ],
    },
  }),
]

// renderUI is a render_ui call that created surfaceId.
const renderUI = (callId: string, surfaceId: string) => [
  ...toolCall(callId, 'render_ui', { messages: [] }),
  toolResult(callId, { surface_id: surfaceId, revision: 1, status: 'created' }),
]

const hotelForm = {
  interruptId: 'ask-1',
  token: 'tok-1',
  revision: 1,
  title: 'Hotel booking',
  question: 'Which hotel?',
  fields: [
    {
      name: 'hotel',
      label: 'Hotel',
      required: true,
      type: 'single_choice',
      options: [
        { value: 'lumiares', label: 'Lumiares' },
        { value: 'memmo', label: 'Memmo Alfama' },
      ],
    },
  ],
}

const formShown = () =>
  [
    { createSurface: { surfaceId: 'form-1', catalogId: 'butter-basic-v1' } },
    {
      updateComponents: {
        surfaceId: 'form-1',
        components: [
          { id: 'root', component: 'Card', child: 'form' },
          {
            id: 'form',
            component: 'Column',
            children: ['title', 'question', 'field_hotel', 'submit'],
          },
          {
            id: 'title',
            component: 'Text',
            text: 'Hotel booking',
            variant: 'h3',
          },
          { id: 'question', component: 'Text', text: 'Which hotel?' },
          {
            id: 'field_hotel',
            component: 'ChoicePicker',
            name: 'hotel',
            label: 'Hotel',
            value: { path: '/values/hotel' },
            required: true,
            variant: 'mutuallyExclusive',
            options: [
              { label: 'Lumiares', value: 'lumiares' },
              { label: 'Memmo Alfama', value: 'memmo' },
            ],
          },
          {
            id: 'submit',
            component: 'Button',
            child: 'submit_label',
            variant: 'primary',
            action: {
              event: {
                name: 'butter.submitForm',
                context: { values: { path: '/values' } },
              },
            },
          },
          { id: 'submit_label', component: 'Text', text: 'Submit' },
        ],
      },
    },
    {
      updateDataModel: {
        surfaceId: 'form-1',
        path: '/',
        value: { values: { hotel: [] }, status: 'pending' },
      },
    },
  ].map((envelope, seq) =>
    a2ui('form-1', 'form', 1, seq, envelope, {
      form: hotelForm,
      fallback: 'Which hotel? (text)',
    })
  )

const running = (page: Page) => page.getByText('Running…')
const composer = (page: Page) => page.getByRole('textbox', { name: /^Message/ })
const stopButton = (page: Page) => page.getByRole('button', { name: 'Stop' })
const sendButton = (page: Page) => page.getByRole('button', { name: 'Send' })
const replies = (page: Page) => page.locator('[data-message-role="assistant"]')
const sentTurns = (page: Page) => page.locator('[data-message-role="user"]')
const toasts = (page: Page) => page.locator('[data-sonner-toast]')
const stoppedNotice = (page: Page) =>
  page.getByRole('status').filter({ hasText: 'Stopped' })
const failureNotice = (page: Page) =>
  page.getByRole('alert').filter({ hasText: 'This run failed' })

const pathsOf = (urls: string[]) => urls.map((u) => new URL(u).pathname)

// The page reads the thread again a second after it found the run, then two
// seconds later.
const AFTER_TWO_READS = { timeout: 10_000 }

// openThreadDuringRun opens t-trip while run-2 holds it, and returns the
// stream of the run's log the page attached to.
async function openThreadDuringRun(
  page: Page,
  fixture: AGUIFixture
): Promise<AttachStream> {
  await page.goto('/chat?thread=t-trip')
  await expect.poll(() => fixture.liveAttaches.length).toBe(1)
  return fixture.liveAttaches[0]
}

// occurrences counts how often text shows in a locator's text.
async function occurrences(
  locator: ReturnType<Page['locator']>,
  text: string
): Promise<number> {
  return (await locator.innerText()).split(text).length - 1
}

test.describe('Chat re-attaching to a run', () => {
  test('reloading mid-run streams the same reply on, live, with no text twice', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        { open: true },
        sse([
          { type: 'RUN_STARTED', threadId: 't-trip', runId: 'r3' },
          { type: 'RUN_FINISHED', threadId: 't-trip', runId: 'r3' },
        ]),
      ],
      sessions: [trip()],
      historyByThread: { 't-trip': history(before.slice(0, 2)) },
    })
    await page.goto('/chat?thread=t-trip', { waitUntil: 'networkidle' })
    await sendMessage(page, 'Now plan the trip')
    await expect(running(page)).toBeVisible()
    const runId = String(fixture.requests[0].runId)

    // The page reloads while the run goes on: the reads show the thread as
    // the run found it, and the run's log can be followed.
    fixture.historyByThread!['t-trip'] = duringRun()
    fixture.snapshots.push(snapshot({ running: RUN }))
    fixture.attaches.push({ live: true, runId })
    await page.reload()
    await expect.poll(() => fixture.liveAttaches.length).toBe(1)
    const log = fixture.liveAttaches[0]
    expect(pathsOf(fixture.attachRequests)).toEqual([`${THREAD}/run`])

    // The replay: what the run sent before the reload.
    log.send([...runStarted({ plan: 'draft' }), textStart, delta('Three days')])
    const reply = replies(page).nth(1)
    await expect(reply).toContainText('Three days')
    await expect(running(page)).toBeVisible()
    await expect(stopButton(page)).toBeEnabled()
    // A run the page streams leaves the composer to type in, as one it
    // started does.
    await expect(composer(page)).toBeEnabled()
    await expect(
      page.getByText('Now plan the trip', { exact: true })
    ).toHaveCount(1)

    // The run goes on, and its reply streams in as it does.
    log.send([delta(' in Lisbon:')])
    await expect(reply).toContainText('Three days in Lisbon:')
    log.send([
      delta(' Alfama, Belém and Sintra.'),
      textEnd,
      {
        type: 'STATE_DELTA',
        delta: [{ op: 'replace', path: '/plan', value: 'final' }],
      },
    ])
    await expect(reply).toContainText(PLAN)
    log.end([runFinished()])
    await expect(running(page)).toHaveCount(0)
    await expect(sendButton(page)).toBeVisible()

    // One reply, with every delta once.
    await expect(replies(page)).toHaveCount(2)
    expect(await occurrences(reply, 'Three days')).toBe(1)
    expect(await occurrences(reply, 'Alfama')).toBe(1)
    await expect(sentTurns(page)).toHaveCount(2)
    // Followed live: the thread was read once, and nothing ran again.
    expect(fixture.historyRequests).toHaveLength(2)
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.stopRequests).toEqual([])
    await expect(toasts(page)).toHaveCount(0)

    // The run moved the shared state, which the next run sends back.
    await page.getByRole('button', { name: 'Shared state' }).click()
    await expect(page.getByText('"plan": "final"')).toBeVisible()
    await sendMessage(page, 'And hotels?')
    await expect.poll(() => fixture.requests.length).toBe(2)
    expect(fixture.requests[1].state).toEqual({ plan: 'final' })
  })

  test('Result Cards show in the reply where the run made them, from the replay and after it, and update in place', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot({ running: RUN })],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    const log = await openThreadDuringRun(page, fixture)
    const reply = replies(page).nth(1)
    const surfaces = page.locator('[data-a2ui-surface]')

    // The replay carries a card the run made before the page attached.
    log.send([
      ...runStarted(),
      ...renderUI('call-plan', 'card-plan'),
      ...cardCreated('card-plan', 'Lisbon itinerary', 'Days', '3'),
      textStart,
      delta('Here is the plan.'),
    ])
    const plan = reply.locator('[data-a2ui-surface="card-plan"]')
    await expect(plan).toBeVisible()
    await expect(plan.getByText('Lisbon itinerary')).toBeVisible()
    await expect(plan.getByText('3', { exact: true })).toBeVisible()
    await expect(reply).toContainText('Here is the plan.')
    await expect(running(page)).toBeVisible()
    // The render_ui call itself stays out of the way, and the card is the
    // reply's, not one restored on its own.
    await expect(
      page.getByRole('button', { name: 'render_ui', exact: true })
    ).toHaveCount(0)
    await expect(
      page.getByRole('region', { name: 'Restored from this conversation' })
    ).toHaveCount(0)

    // The run goes on: it changes that card, which re-renders where it is,
    // and makes another one.
    log.send([
      a2ui('card-plan', 'card', 2, 0, {
        updateComponents: {
          surfaceId: 'card-plan',
          components: [
            { id: 'kv', component: 'KeyValue', label: 'Days', value: '4' },
          ],
        },
      }),
    ])
    await expect(plan.getByText('4', { exact: true })).toBeVisible()
    await expect(surfaces).toHaveCount(1)
    log.send([
      ...renderUI('call-hotel', 'card-hotel'),
      ...cardCreated('card-hotel', 'Hotel', 'Name', 'Lumiares'),
    ])
    const hotel = reply.locator('[data-a2ui-surface="card-hotel"]')
    await expect(hotel).toBeVisible()
    await expect(hotel.getByText('Lumiares')).toBeVisible()
    await expect(surfaces).toHaveCount(2)
    // Each card sits where its creation arrived in the reply.
    expect(
      await surfaces.evaluateAll((els) =>
        els.map((el) => el.getAttribute('data-a2ui-surface'))
      )
    ).toEqual(['card-plan', 'card-hotel'])

    log.end([textEnd, runFinished()])
    await expect(running(page)).toHaveCount(0)
    await expect(plan.getByText('4', { exact: true })).toBeVisible()
    await expect(hotel).toBeVisible()
    expect(fixture.historyRequests).toHaveLength(1)
  })

  test('a run that ends on a question shows its prompt, which answers it', async ({
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
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot({ running: RUN })],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    const log = await openThreadDuringRun(page, fixture)
    log.send([...runStarted(), textStart, delta('Two hotels fit.')])
    await expect(replies(page).nth(1)).toContainText('Two hotels fit.')
    const question = page.getByRole('group', { name: 'Which hotel?' })
    await expect(question).toHaveCount(0)

    log.end([
      textEnd,
      runFinished({
        type: 'interrupt',
        interrupts: [
          { id: 'int-1', reason: 'human_input', message: 'Which hotel?' },
        ],
      }),
    ])
    await expect(question).toBeVisible()
    await expect(running(page)).toHaveCount(0)
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

  test('a run that ends on a form shows the form, which answers it', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        sse([
          { type: 'RUN_STARTED', threadId: 't-trip', runId: 'r3' },
          {
            type: 'RUN_FINISHED',
            threadId: 't-trip',
            runId: 'r3',
            outcome: { type: 'success' },
          },
        ]),
      ],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot({ running: RUN })],
      attaches: [
        sse([
          ...runStarted(),
          textStart,
          delta('Two hotels fit.'),
          textEnd,
          ...formShown(),
          runFinished({
            type: 'interrupt',
            interrupts: [
              {
                id: 'ask-1',
                reason: 'human_input',
                message: 'Which hotel?\n\nPlease answer these fields…',
              },
            ],
          }),
        ]),
      ],
    })
    await page.goto('/chat?thread=t-trip')
    const form = page.getByRole('region', { name: 'Hotel booking' })
    await expect(form).toBeVisible()
    await expect(replies(page).nth(1)).toContainText('Two hotels fit.')
    // The form takes the place of the text prompt for its question.
    await expect(page.getByPlaceholder('Type your answer…')).toHaveCount(0)
    await expect(running(page)).toHaveCount(0)

    await form.getByRole('radio', { name: 'Lumiares' }).check()
    await form.getByRole('button', { name: 'Submit' }).click()
    await expect.poll(() => fixture.requests.length).toBe(1)
    const resume = fixture.requests[0].resume as Array<Record<string, unknown>>
    expect(resume).toHaveLength(1)
    expect(resume[0]).toMatchObject({
      interruptId: 'ask-1',
      status: 'resolved',
      payload: {
        butterForm: {
          surfaceId: 'form-1',
          token: 'tok-1',
          revision: 1,
          values: { hotel: 'lumiares' },
        },
      },
    })
    expect(fixture.historyRequests).toHaveLength(1)
  })

  test('a run that fails shows the failure with its turn, and a toast', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot({ running: RUN })],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    const log = await openThreadDuringRun(page, fixture)
    log.send([...runStarted(), textStart, delta('Day one: Alfama.')])
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')

    log.end([
      textEnd,
      { type: 'RUN_ERROR', message: 'model exploded', runId: RUN.runId },
    ])
    const notice = failureNotice(page)
    await expect(notice).toBeVisible()
    await expect(notice).toContainText('model exploded')
    await expect(notice).toContainText('Now plan the trip')
    await expect(
      toasts(page).filter({ hasText: 'model exploded' })
    ).toBeVisible()
    // What the run streamed before it failed stays.
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')
    await notice.getByRole('button', { name: 'Restore input' }).click()
    await expect(composer(page)).toHaveValue('Now plan the trip')
    expect(fixture.historyRequests).toHaveLength(1)
  })

  test('a run stopped from elsewhere shows as stopped, not failed', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot({ running: RUN })],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    const log = await openThreadDuringRun(page, fixture)
    log.send([...runStarted(), textStart, delta('Day one: Alfama.')])
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')

    // A Stop from another tab, or a delete elsewhere, ends the run.
    log.end([textEnd, stoppedEvent(RUN.runId)])
    await expect(running(page)).toHaveCount(0)
    await expect(stoppedNotice(page)).toContainText('Now plan the trip')
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')
    await expect(failureNotice(page)).toHaveCount(0)
    await expect(toasts(page)).toHaveCount(0)
    // This page sent no Stop.
    expect(fixture.stopRequests).toEqual([])
    expect(fixture.historyRequests).toHaveLength(1)
  })

  test('Stop during the re-attach stops the run, and ends its stream here', async ({
    page,
  }) => {
    let accept!: () => void
    const accepted = new Promise<void>((resolve) => (accept = resolve))
    const fixture = await setupAGUI(page, {
      runs: [
        sse([
          { type: 'RUN_STARTED', threadId: 't-trip', runId: 'r3' },
          { type: 'TEXT_MESSAGE_START', messageId: 'a3', role: 'assistant' },
          {
            type: 'TEXT_MESSAGE_CONTENT',
            messageId: 'a3',
            delta: 'Day two: Belém.',
          },
          { type: 'TEXT_MESSAGE_END', messageId: 'a3' },
          { type: 'RUN_FINISHED', threadId: 't-trip', runId: 'r3' },
        ]),
      ],
      sessions: [trip()],
      historyByThread: { 't-trip': duringRun() },
      snapshots: [snapshot({ running: RUN })],
      attaches: [{ live: true, runId: RUN.runId }],
      stops: [{ status: 202, until: accepted }],
    })
    const log = await openThreadDuringRun(page, fixture)
    log.send([...runStarted(), textStart, delta('Day one: Alfama.')])
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')

    await stopButton(page).dblclick()
    await expect(stopButton(page)).toBeDisabled()
    await expect
      .poll(() => pathsOf(fixture.stopRequests))
      .toEqual([`${THREAD}/stop`])
    // Until the server takes the Stop, the run goes on here.
    await expect(running(page)).toBeVisible()

    accept()
    await expect(running(page)).toHaveCount(0)
    await expect(sendButton(page)).toBeVisible()
    await expect(stoppedNotice(page)).toContainText('Now plan the trip')
    // The reply keeps what the run streamed before it stopped.
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')
    await expect.poll(() => log.aborted || log.ended).toBe(true)
    // A stopped run is not a failure, and nothing ran or was read again.
    await expect(toasts(page)).toHaveCount(0)
    expect(fixture.stopRequests).toHaveLength(1)
    expect(fixture.requests).toEqual([])
    expect(fixture.historyRequests).toHaveLength(1)

    // The thread goes on as usual.
    await sendMessage(page, 'Then day two?')
    await expect(page.getByText('Day two: Belém.')).toBeVisible()
    expect(fixture.requests).toHaveLength(1)
    await expect(stoppedNotice(page)).toHaveCount(0)
  })

  test('leaving the thread mid-attach only stops following the run', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip(), budget()],
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
      snapshots: [snapshot({ running: RUN })],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    const log = await openThreadDuringRun(page, fixture)
    log.send([...runStarted(), textStart, delta('Day one: Alfama.')])
    await expect(replies(page).nth(1)).toContainText('Day one: Alfama.')

    await page
      .getByRole('navigation', { name: 'Threads' })
      .getByRole('link', { name: 'Budget review', exact: true })
      .click()
    await expect(page.getByText('The budget holds.')).toBeVisible()
    await expect.poll(() => log.aborted).toBe(true)
    await expect(running(page)).toHaveCount(0)
    // The run goes on without the page, which shows none of it.
    log.end([delta(' Day two: Belém.'), textEnd, runFinished()])
    await page.waitForTimeout(500)
    await expect(page.getByText('Day one: Alfama.')).toHaveCount(0)
    await expect(page.getByText('Day two: Belém.')).toHaveCount(0)
    expect(fixture.stopRequests).toEqual([])
    expect(fixture.requests).toEqual([])
    expect(
      pathsOf(fixture.historyRequests).filter((p) => p.includes('/t-trip/'))
    ).toHaveLength(1)
  })

  for (const [what, attach] of [
    ['answers 204, with no log to follow', { status: 204 }],
    [
      'fails (503)',
      { status: 503, body: { error: 'run log unavailable, retry later' } },
    ],
  ] as const) {
    test(`attaching ${what}: the page waits the run out, then shows its reply`, async ({
      page,
    }) => {
      const fixture = await setupAGUI(page, {
        runs: [],
        sessions: [trip()],
        histories: [duringRun(), duringRun(), afterRun()],
        snapshots: [snapshot({ running: RUN }), snapshot()],
        attaches: [attach],
      })
      await page.goto('/chat?thread=t-trip')
      await expect(running(page)).toBeVisible()
      await expect.poll(() => fixture.attachRequests.length).toBe(1)
      // Waiting the run out, the page keeps the composer to itself.
      await expect(composer(page)).toBeDisabled()

      await expect(page.getByText(PLAN)).toBeVisible(AFTER_TWO_READS)
      await expect(running(page)).toHaveCount(0)
      await expect(replies(page)).toHaveCount(2)
      await expect(composer(page)).toBeEnabled()
      expect(fixture.historyRequests).toHaveLength(3)
      expect(fixture.attachRequests).toHaveLength(1)
      expect(fixture.requests).toEqual([])
      await expect(toasts(page)).toHaveCount(0)
    })
  }

  test('a run that took the thread after its history was read is waited out, and an earlier run’s notice does not outlive it', async ({
    page,
  }) => {
    const earlier = [
      { id: 'u1', role: 'user', content: 'Book a flight to Lisbon' },
      { id: 'a1', role: 'assistant', content: 'Looking for flights' },
    ]
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      // The history was read just before run-2 took the thread: it shows
      // how the run before it failed. The snapshot, read after, names run-2.
      histories: [
        history(earlier, {
          lastRun: {
            status: 'failed',
            error: 'model exploded',
            input: 'Book a flight to Lisbon',
          },
        }),
        history([
          ...earlier,
          { id: 'u2', role: 'user', content: 'Now plan the trip' },
          { id: 'a2', role: 'assistant', content: PLAN },
        ]),
      ],
      snapshots: [snapshot({ running: RUN }), snapshot()],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    await page.goto('/chat?thread=t-trip')
    await expect(page.getByText('Looking for flights')).toBeVisible()
    await expect(running(page)).toBeVisible()
    // That history lacks run-2's turn, so the page waits the run out rather
    // than stream its reply under the wrong turn.
    await expect(composer(page)).toBeDisabled()

    await expect(page.getByText(PLAN)).toBeVisible(AFTER_TWO_READS)
    await expect(running(page)).toHaveCount(0)
    await expect(
      page.getByText('Now plan the trip', { exact: true })
    ).toBeVisible()
    await expect(replies(page)).toHaveCount(2)
    // The earlier run's failure is not run-2's.
    await expect(failureNotice(page)).toHaveCount(0)
    expect(fixture.attachRequests).toEqual([])
    expect(fixture.historyRequests).toHaveLength(2)
  })

  test('the stream ends with the fallback marker: the page waits the run out, and its reply replaces what streamed', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [],
      sessions: [trip()],
      histories: [duringRun(), afterRun()],
      snapshots: [snapshot({ running: RUN }), snapshot()],
      attaches: [{ live: true, runId: RUN.runId }],
    })
    const log = await openThreadDuringRun(page, fixture)
    log.send([...runStarted(), textStart, delta('Three days')])
    await expect(replies(page).nth(1)).toContainText('Three days')
    await expect(composer(page)).toBeEnabled()

    // The log fell behind the run, which goes on.
    log.end([
      {
        type: 'CUSTOM',
        name: 'butter.fallback',
        value: { threadId: 't-trip', runId: RUN.runId, reason: 'truncated' },
      },
    ])
    await expect(composer(page)).toBeDisabled()
    await expect(running(page)).toBeVisible()

    await expect(page.getByText(PLAN)).toBeVisible(AFTER_TWO_READS)
    await expect(running(page)).toHaveCount(0)
    // The read's reply replaced the one that streamed: no part of it twice.
    await expect(replies(page)).toHaveCount(2)
    expect(await occurrences(replies(page).nth(1), 'Three days')).toBe(1)
    await expect(composer(page)).toBeEnabled()
    expect(fixture.historyRequests).toHaveLength(2)
    expect(fixture.requests).toEqual([])
  })
})
