import { expect, test, type Page } from '@playwright/test'
import { setupAGUI, sse } from './support/agui'

// A2UI in AG-UI Chat. Every fixture below is the wire traffic the Butter
// server produces (internal/handler/http/agui_a2ui_test.go pins the server
// side): butter.a2ui CUSTOM events carrying one A2UI v0.9.1 envelope each,
// and the UI snapshot for a thread.

const V = 'v0.9.1'

interface Header {
  surfaceId: string
  kind: 'card' | 'form'
  revision: number
  seq: number
  runId?: string
  fallback?: string
  form?: Record<string, unknown>
}

function a2ui(h: Header, envelope: Record<string, unknown>) {
  return {
    type: 'CUSTOM',
    name: 'butter.a2ui',
    value: {
      version: V,
      threadId: 't',
      runId: h.runId ?? 'r1',
      messageId: 'a1',
      ...h,
      envelope: { version: V, ...envelope },
    },
  }
}

const cardComponents = (status: string, tone: string) => [
  { id: 'root', component: 'Card', child: 'col' },
  { id: 'col', component: 'Column', children: ['title', 'body', 'env', 'state'] },
  { id: 'title', component: 'Text', text: 'Deploy summary', variant: 'h3' },
  { id: 'body', component: 'Text', text: { path: '/summary' } },
  { id: 'env', component: 'KeyValue', label: 'Environment', value: 'production' },
  { id: 'state', component: 'Status', text: status, tone },
]

function cardCreated(surfaceId: string) {
  const h = { surfaceId, kind: 'card' as const, revision: 1, fallback: 'Deploy summary (text)' }
  return [
    a2ui({ ...h, seq: 0 }, { createSurface: { surfaceId, catalogId: 'butter-basic-v1' } }),
    a2ui({ ...h, seq: 1 }, { updateComponents: { surfaceId, components: cardComponents('Healthy', 'success') } }),
    a2ui({ ...h, seq: 2 }, { updateDataModel: { surfaceId, path: '/', value: { summary: '3 services rolled out.' } } }),
  ]
}

const runStarted = (runId: string) => ({ type: 'RUN_STARTED', threadId: 't', runId })
const runFinished = (runId: string, outcome?: Record<string, unknown>) => ({
  type: 'RUN_FINISHED',
  threadId: 't',
  runId,
  ...(outcome ? { outcome } : {}),
})
const text = (messageId: string, delta: string) => [
  { type: 'TEXT_MESSAGE_START', messageId, role: 'assistant' },
  { type: 'TEXT_MESSAGE_CONTENT', messageId, delta },
  { type: 'TEXT_MESSAGE_END', messageId },
]

async function send(page: Page, message: string) {
  const composer = page.getByPlaceholder(/Message the agent over AG-UI/)
  await composer.fill(message)
  await composer.press('Enter')
}

const card = (page: Page) => page.locator('[data-a2ui-surface]')

test.describe('A2UI result cards', () => {
  test('renders a card beside text and tool calls, updates it in place, ignores stale replays, and removes it', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        sse([
          runStarted('r1'),
          { type: 'TOOL_CALL_START', toolCallId: 'tc-s', toolCallName: 'search' },
          { type: 'TOOL_CALL_ARGS', toolCallId: 'tc-s', delta: '{"q":"deploys"}' },
          { type: 'TOOL_CALL_END', toolCallId: 'tc-s' },
          { type: 'TOOL_CALL_RESULT', messageId: 'a1', toolCallId: 'tc-s', content: '{"hits":1}' },
          { type: 'TOOL_CALL_START', toolCallId: 'tc-r', toolCallName: 'render_ui' },
          { type: 'TOOL_CALL_ARGS', toolCallId: 'tc-r', delta: '{"messages":[]}' },
          { type: 'TOOL_CALL_END', toolCallId: 'tc-r' },
          { type: 'TOOL_CALL_RESULT', messageId: 'a1', toolCallId: 'tc-r', content: '{"surface_id":"card-1","revision":1,"status":"created"}' },
          ...cardCreated('card-1'),
          ...text('a1', 'Deployed everything.'),
          runFinished('r1'),
        ]),
        sse([
          runStarted('r2'),
          a2ui({ surfaceId: 'card-1', kind: 'card', revision: 2, seq: 0, runId: 'r2' }, {
            updateComponents: { surfaceId: 'card-1', components: [{ id: 'state', component: 'Status', text: 'Degraded', tone: 'warning' }] },
          }),
          a2ui({ surfaceId: 'card-1', kind: 'card', revision: 2, seq: 1, runId: 'r2' }, {
            updateDataModel: { surfaceId: 'card-1', path: '/', value: { summary: '1 service rolled back.' } },
          }),
          // A replay of the first revision must not overwrite the second.
          a2ui({ surfaceId: 'card-1', kind: 'card', revision: 1, seq: 2, runId: 'r2' }, {
            updateDataModel: { surfaceId: 'card-1', path: '/', value: { summary: '3 services rolled out.' } },
          }),
          ...text('a2', 'Status changed.'),
          runFinished('r2'),
        ]),
        sse([
          runStarted('r3'),
          a2ui({ surfaceId: 'card-1', kind: 'card', revision: 3, seq: 0, runId: 'r3' }, { deleteSurface: { surfaceId: 'card-1' } }),
          // Nothing brings a deleted card back.
          a2ui({ surfaceId: 'card-1', kind: 'card', revision: 4, seq: 0, runId: 'r3' }, { createSurface: { surfaceId: 'card-1', catalogId: 'butter-basic-v1' } }),
          ...text('a3', 'Cleared.'),
          runFinished('r3'),
        ]),
      ],
    })

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'deploy')

    const deployCard = card(page)
    await expect(deployCard).toHaveCount(1)
    await expect(deployCard.getByRole('heading', { name: 'Deploy summary' })).toBeVisible()
    await expect(deployCard.getByText('3 services rolled out.')).toBeVisible()
    await expect(deployCard.getByText('Environment')).toBeVisible()
    await expect(deployCard.getByText('production')).toBeVisible()
    await expect(deployCard.getByText('Healthy')).toBeVisible()
    // Text and ordinary tool calls still render; the render_ui call itself
    // stays out of the way.
    await expect(page.getByText('Deployed everything.')).toBeVisible()
    await expect(page.getByRole('button', { name: 'search', exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: 'render_ui', exact: true })).toHaveCount(0)
    // The client declares the capability on every run.
    expect(fixture.requests[0].forwardedProps).toMatchObject({
      butterA2UI: { version: V, catalogs: ['butter-basic-v1'] },
    })

    await send(page, 'status?')
    await expect(page.getByText('Status changed.')).toBeVisible()
    await expect(card(page)).toHaveCount(1)
    await expect(card(page).getByText('Degraded')).toBeVisible()
    await expect(card(page).getByText('1 service rolled back.')).toBeVisible()
    await expect(card(page).getByText('Healthy')).toHaveCount(0)

    await send(page, 'clear')
    await expect(page.getByText('Cleared.')).toBeVisible()
    await expect(card(page)).toHaveCount(0)
  })

  test('shows the readable fallback for a surface it cannot render and keeps chatting', async ({ page }) => {
    await setupAGUI(page, {
      runs: [
        sse([
          runStarted('r1'),
          a2ui({ surfaceId: 'card-x', kind: 'card', revision: 1, seq: 0, fallback: 'Plain version of the foreign card.' }, {
            createSurface: { surfaceId: 'card-x', catalogId: 'https://example.com/other-catalog' },
          }),
          a2ui({ surfaceId: 'card-x', kind: 'card', revision: 1, seq: 1, fallback: 'Plain version of the foreign card.' }, {
            updateComponents: { surfaceId: 'card-x', components: [{ id: 'root', component: 'Text', text: 'never shown' }] },
          }),
          a2ui({ surfaceId: 'card-y', kind: 'card', revision: 1, seq: 0, fallback: 'Plain version of the odd card.' }, {
            createSurface: { surfaceId: 'card-y', catalogId: 'butter-basic-v1' },
          }),
          a2ui({ surfaceId: 'card-y', kind: 'card', revision: 1, seq: 1, fallback: 'Plain version of the odd card.' }, {
            updateComponents: { surfaceId: 'card-y', components: [{ id: 'root', component: 'Hologram', text: 'x' }] },
          }),
          ...cardCreated('card-ok'),
          ...text('a1', 'Still here.'),
          runFinished('r1'),
        ]),
        sse([runStarted('r2'), ...text('a2', 'Next answer.'), runFinished('r2')]),
      ],
    })

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'show me')

    await expect(page.getByText('Plain version of the foreign card.')).toBeVisible()
    await expect(page.getByText('never shown')).toHaveCount(0)
    await expect(page.getByText('Plain version of the odd card.')).toBeVisible()
    await expect(page.locator('[data-a2ui-surface="card-ok"]').getByText('Deploy summary')).toBeVisible()
    await expect(page.getByText('Still here.')).toBeVisible()

    await send(page, 'and then?')
    await expect(page.getByText('Next answer.')).toBeVisible()
  })
})

// --- forms ---

const deployFormView = {
  interruptId: 'ask-1',
  token: 'tok-1',
  revision: 1,
  title: 'Deploy approval',
  question: 'Approve this deploy?',
  fields: [
    {
      name: 'env', label: 'Environment', required: true, type: 'single_choice',
      options: [{ value: 'prod', label: 'Production' }, { value: 'staging', label: 'Staging' }],
    },
    { name: 'reason', label: 'Reason', hint: 'Why now?', required: true, type: 'text', maxLength: 20 },
    { name: 'note', label: 'Note', type: 'text', maxLength: 2000 },
  ],
}

function formEnvelopes(surfaceId: string) {
  return [
    { createSurface: { surfaceId, catalogId: 'butter-basic-v1' } },
    {
      updateComponents: {
        surfaceId,
        components: [
          { id: 'root', component: 'Card', child: 'form' },
          { id: 'form', component: 'Column', children: ['title', 'question', 'field_env', 'field_reason', 'field_note', 'submit'] },
          { id: 'title', component: 'Text', text: 'Deploy approval', variant: 'h3' },
          { id: 'question', component: 'Text', text: 'Approve this deploy?' },
          {
            id: 'field_env', component: 'ChoicePicker', name: 'env', label: 'Environment', value: { path: '/values/env' }, required: true,
            variant: 'mutuallyExclusive', options: [{ label: 'Production', value: 'prod' }, { label: 'Staging', value: 'staging' }],
          },
          { id: 'field_reason', component: 'TextField', name: 'reason', label: 'Reason', value: { path: '/values/reason' }, hint: 'Why now?', required: true, variant: 'shortText', maxLength: 20 },
          { id: 'field_note', component: 'TextField', name: 'note', label: 'Note', value: { path: '/values/note' }, variant: 'longText', maxLength: 2000 },
          {
            id: 'submit', component: 'Button', child: 'submit_label', variant: 'primary',
            action: { event: { name: 'butter.submitForm', context: { values: { path: '/values' } } } },
          },
          { id: 'submit_label', component: 'Text', text: 'Submit' },
        ],
      },
    },
    { updateDataModel: { surfaceId, path: '/', value: { values: { env: [], reason: '', note: '' }, status: 'pending' } } },
  ]
}

function formShown(surfaceId = 'form-1') {
  return formEnvelopes(surfaceId).map((env, seq) =>
    a2ui({ surfaceId, kind: 'form', revision: 1, seq, form: deployFormView, fallback: 'Approve this deploy? (text)' }, env)
  )
}

const pausedRun = sse([
  runStarted('r1'),
  ...formShown(),
  runFinished('r1', {
    type: 'interrupt',
    interrupts: [{ id: 'ask-1', reason: 'human_input', message: 'Approve this deploy?\n\nPlease answer these fields…' }],
  }),
])

const answeredRun = sse([
  runStarted('r2'),
  a2ui({ surfaceId: 'form-1', kind: 'form', revision: 2, seq: 0, runId: 'r2', form: deployFormView }, {
    updateDataModel: { surfaceId: 'form-1', path: '/status', value: 'answered' },
  }),
  ...text('a2', 'Published to production.'),
  runFinished('r2', { type: 'success' }),
])

const form = (page: Page) => page.getByRole('region', { name: 'Deploy approval' })

function butterForm(request: Record<string, unknown>) {
  const resume = request.resume as Array<Record<string, unknown>>
  return { resume, payload: (resume?.[0]?.payload as { butterForm: Record<string, unknown> })?.butterForm }
}

test.describe('A2UI Human Input forms', () => {
  test('validates fields, submits a structured answer to its own interrupt, then locks', async ({ page }) => {
    const fixture = await setupAGUI(page, { runs: [pausedRun, answeredRun] })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')

    await expect(form(page)).toBeVisible()
    await expect(form(page).getByText('Approve this deploy?')).toBeVisible()
    // The form replaces the free-text prompt for its interrupt.
    await expect(page.getByPlaceholder('Type your answer…')).toHaveCount(0)
    await expect(form(page).getByRole('radio', { name: 'Production' })).toBeVisible()
    await expect(form(page).getByText('Why now?')).toBeVisible()

    // Client-side validation: nothing is sent, the errors are announced,
    // and focus lands on the first invalid field.
    await form(page).getByRole('button', { name: 'Submit' }).click()
    await expect(form(page).getByRole('alert').filter({ hasText: 'Required' }).first()).toBeVisible()
    await expect(form(page).getByRole('radio', { name: 'Production' })).toBeFocused()
    expect(fixture.requests).toHaveLength(1)

    await form(page).getByRole('radio', { name: 'Production' }).check()
    // The input itself stops at the field's limit.
    await form(page).getByLabel('Reason').fill('too long for twenty characters')
    await expect(form(page).getByLabel('Reason')).toHaveValue('too long for twenty ')
    expect(fixture.requests).toHaveLength(1)

    await form(page).getByLabel('Reason').fill('  hotfix ')
    await form(page).getByRole('button', { name: 'Submit' }).click()

    await expect(page.getByText('Published to production.')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
    const { resume, payload } = butterForm(fixture.requests[1])
    expect(resume).toHaveLength(1)
    expect(resume[0]).toMatchObject({ interruptId: 'ask-1', status: 'resolved' })
    expect(payload).toEqual({
      version: V,
      surfaceId: 'form-1',
      revision: 1,
      token: 'tok-1',
      values: { env: 'prod', reason: 'hotfix', note: '' },
    })
    expect(fixture.requests[1].threadId).toBe(fixture.requests[0].threadId)

    // The submission reads as a reply in the conversation, and the form
    // cannot resume the workflow twice.
    await expect(page.getByText('Environment: Production')).toBeVisible()
    await expect(form(page).getByText('Submitted')).toBeVisible()
    await expect(form(page).getByRole('button', { name: 'Submit' })).toBeDisabled()
    await expect(form(page).getByLabel('Reason')).toBeDisabled()
  })

  test('keeps the draft and shows server field errors when a submission is refused', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        pausedRun,
        {
          status: 422,
          body: {
            error: 'some fields are invalid',
            code: 'form_invalid',
            fieldErrors: { reason: 'must be at most 20 characters' },
          },
        },
        answeredRun,
      ],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')

    await form(page).getByRole('radio', { name: 'Staging' }).check()
    await form(page).getByLabel('Reason').fill('weekly train')
    await form(page).getByLabel('Note').fill('ping on-call')
    await form(page).getByRole('button', { name: 'Submit' }).click()

    await expect(form(page).getByText('must be at most 20 characters')).toBeVisible()
    await expect(form(page).getByText('some fields are invalid')).toBeVisible()
    // Nothing typed was lost.
    await expect(form(page).getByRole('radio', { name: 'Staging' })).toBeChecked()
    await expect(form(page).getByLabel('Reason')).toHaveValue('weekly train')
    await expect(form(page).getByLabel('Note')).toHaveValue('ping on-call')
    await expect(form(page).getByRole('button', { name: 'Submit' })).toBeEnabled()

    await form(page).getByLabel('Reason').fill('weekly')
    await form(page).getByRole('button', { name: 'Submit' }).click()
    await expect(page.getByText('Published to production.')).toBeVisible()
    // The retry still addresses only this form's interrupt.
    expect(butterForm(fixture.requests[2]).resume).toHaveLength(1)
    expect(butterForm(fixture.requests[2]).payload.values).toEqual({ env: 'staging', reason: 'weekly', note: 'ping on-call' })
  })

  test('tells the user when the form was already submitted', async ({ page }) => {
    await setupAGUI(page, {
      runs: [pausedRun, { status: 409, body: { error: 'this form was already submitted', code: 'form_answered' } }],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')
    await form(page).getByRole('radio', { name: 'Production' }).check()
    await form(page).getByLabel('Reason').fill('hotfix')
    await form(page).getByRole('button', { name: 'Submit' }).click()
    await expect(form(page).getByText('This form was already submitted.')).toBeVisible()
    await expect(form(page).getByRole('button', { name: 'Submit' })).toBeDisabled()
  })

  test('shows the submission in flight and sends it only once', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [pausedRun, { delayMs: 1500, sse: answeredRun }],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')
    await form(page).getByRole('radio', { name: 'Production' }).check()
    await form(page).getByLabel('Reason').fill('hotfix')

    await form(page).getByRole('button', { name: 'Submit' }).dblclick()
    await expect(form(page).getByText('Submitting…')).toBeVisible()
    await expect(form(page).getByRole('button', { name: 'Submit' })).toBeDisabled()
    await expect(form(page).getByLabel('Reason')).toBeDisabled()

    await expect(page.getByText('Published to production.')).toBeVisible()
    await expect(form(page).getByText('Submitted')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
  })

  test('can be filled and submitted with the keyboard alone', async ({ page }) => {
    const fixture = await setupAGUI(page, { runs: [pausedRun, answeredRun] })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')

    // Start keyboard navigation at the form, as clicking its heading would.
    await form(page).getByRole('heading', { name: 'Deploy approval' }).click()
    await page.keyboard.press('Tab')
    await expect(form(page).getByRole('radio', { name: 'Production' })).toBeFocused()
    // Radix selects the radio focus moves to while the arrow key is held.
    await page.keyboard.down('ArrowDown')
    await expect(form(page).getByRole('radio', { name: 'Staging' })).toBeFocused()
    await page.keyboard.up('ArrowDown')
    await expect(form(page).getByRole('radio', { name: 'Staging' })).toBeChecked()
    await page.keyboard.press('Tab')
    await expect(form(page).getByLabel('Reason')).toBeFocused()
    await page.keyboard.type('weekly')
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')
    await expect(form(page).getByRole('button', { name: 'Submit' })).toBeFocused()
    await page.keyboard.press('Enter')

    await expect(page.getByText('Published to production.')).toBeVisible()
    expect(butterForm(fixture.requests[1]).payload.values).toMatchObject({ env: 'staging', reason: 'weekly' })
  })
})

// --- recovery ---

test.describe('A2UI recovery', () => {
  test('restores cards and the unanswered form after a refresh, and the form still submits', async ({ page }) => {
    const restored = {
      body: {
        version: V,
        catalogId: 'butter-basic-v1',
        threadId: 't',
        surfaces: [
          {
            surfaceId: 'card-1',
            kind: 'card',
            revision: 2,
            runId: 'r1',
            messageId: 'a1',
            fallback: 'Deploy summary (text)',
            envelopes: [
              { version: V, createSurface: { surfaceId: 'card-1', catalogId: 'butter-basic-v1' } },
              { version: V, updateComponents: { surfaceId: 'card-1', components: cardComponents('Degraded', 'warning') } },
              { version: V, updateDataModel: { surfaceId: 'card-1', path: '/', value: { summary: 'Restored summary.' } } },
            ],
          },
          {
            surfaceId: 'form-1',
            kind: 'form',
            revision: 1,
            fallback: 'Approve this deploy? (text)',
            form: deployFormView,
            envelopes: formEnvelopes('form-1').map((env) => ({ version: V, ...env })),
          },
        ],
      },
    }
    const fixture = await setupAGUI(page, { runs: [pausedRun, answeredRun] })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')
    await expect(form(page)).toBeVisible()
    const threadId = fixture.requests[0].threadId as string

    // After the refresh the thread is busy once (409), then the snapshot
    // arrives.
    fixture.snapshots.push({ status: 409, body: { error: 'a run is in progress' } }, restored, restored)
    await page.reload({ waitUntil: 'networkidle' })

    const region = page.getByRole('region', { name: 'Restored from this conversation' })
    await expect(region).toBeVisible()
    await expect(region.getByText('Restored summary.')).toBeVisible()
    await expect(region.getByText('Degraded')).toBeVisible()
    await expect(region.getByText('Waiting for your answer')).toBeVisible()
    expect(fixture.snapshotRequests.length).toBeGreaterThanOrEqual(3)
    for (const url of fixture.snapshotRequests) {
      expect(url).toContain(`/api/agui/streamer-id/threads/${threadId}/ui`)
    }

    await form(page).getByRole('radio', { name: 'Production' }).check()
    await form(page).getByLabel('Reason').fill('after refresh')
    await form(page).getByRole('button', { name: 'Submit' }).click()
    await expect(page.getByText('Published to production.')).toBeVisible()
    const submit = fixture.requests[fixture.requests.length - 1]
    expect(submit.threadId).toBe(threadId)
    expect(butterForm(submit).payload).toMatchObject({ surfaceId: 'form-1', token: 'tok-1' })
    await expect(form(page).getByText('Submitted')).toBeVisible()
  })

  test('switching agent or starting a new thread clears the previous UI', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [sse([runStarted('r1'), ...cardCreated('card-1'), ...text('a1', 'Here.'), runFinished('r1')])],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'deploy')
    await expect(card(page)).toHaveCount(1)
    const firstThread = fixture.requests[0].threadId as string

    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: 'Second' }).click()
    await expect(card(page)).toHaveCount(0)
    await expect(page.getByText('Here.')).toHaveCount(0)
    await expect.poll(() => fixture.snapshotRequests.some((u) => u.includes('/api/agui/second-id/threads/'))).toBe(true)
    const secondURL = fixture.snapshotRequests.find((u) => u.includes('/second-id/'))!
    expect(secondURL).not.toContain(firstThread)

    // Back on the first agent the same thread is resumed (its snapshot is
    // read again); a new thread starts empty with a fresh id.
    const firstThreadReads = () => fixture.snapshotRequests.filter((u) => u.includes(firstThread)).length
    const readsBefore = firstThreadReads()
    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: 'Streamer' }).click()
    await expect.poll(firstThreadReads).toBeGreaterThan(readsBefore)
    await page.getByRole('button', { name: 'New thread' }).click()
    await expect
      .poll(() => fixture.snapshotRequests.some((u) => u.includes('/streamer-id/') && !u.includes(firstThread)))
      .toBe(true)
    await expect(card(page)).toHaveCount(0)
  })

  test('a refresh restores the conversation with the card in the reply that produced it', async ({ page }) => {
    const fixture = await setupAGUI(page, {
      runs: [sse([runStarted('r1'), ...cardCreated('card-1'), ...text('a1', 'Here.'), runFinished('r1')])],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'deploy')
    await expect(card(page)).toHaveCount(1)
    const threadId = fixture.requests[0].threadId as string

    fixture.historyByThread = {
      [threadId]: {
        body: {
          threadId,
          messages: [
            { id: 'u1', role: 'user', content: 'deploy' },
            { id: 'a1', role: 'assistant', content: 'Here.' },
          ],
          interrupts: [],
          surfaces: [{ surfaceId: 'card-1', messageId: 'a1' }],
        },
      },
    }
    fixture.snapshots.push({
      body: {
        version: V,
        catalogId: 'butter-basic-v1',
        threadId,
        surfaces: [
          {
            surfaceId: 'card-1',
            kind: 'card',
            revision: 1,
            fallback: 'Deploy summary (text)',
            envelopes: [
              { version: V, createSurface: { surfaceId: 'card-1', catalogId: 'butter-basic-v1' } },
              { version: V, updateComponents: { surfaceId: 'card-1', components: cardComponents('Healthy', 'success') } },
              { version: V, updateDataModel: { surfaceId: 'card-1', path: '/', value: { summary: '3 services rolled out.' } } },
            ],
          },
        ],
      },
    })
    await page.reload({ waitUntil: 'networkidle' })

    await expect(page.getByText('deploy', { exact: true })).toBeVisible()
    await expect(page.getByText('Here.')).toBeVisible()
    await expect(card(page)).toHaveCount(1)
    await expect(card(page).getByText('3 services rolled out.')).toBeVisible()
    await expect(page.getByRole('region', { name: 'Restored from this conversation' })).toHaveCount(0)
  })

  test('an answered form does not come back after a refresh', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, { runs: [pausedRun, answeredRun] })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'release it')
    await form(page).getByRole('radio', { name: 'Production' }).check()
    await form(page).getByLabel('Reason').fill('hotfix')
    await form(page).getByRole('button', { name: 'Submit' }).click()
    await expect(page.getByText('Published to production.')).toBeVisible()

    // The server no longer lists the answered form.
    const reads = fixture.snapshotRequests.length
    await page.reload({ waitUntil: 'networkidle' })
    await expect.poll(() => fixture.snapshotRequests.length).toBeGreaterThan(reads)
    await expect(form(page)).toHaveCount(0)
    await expect(
      page.getByRole('region', { name: 'Restored from this conversation' })
    ).toHaveCount(0)
    expect(fixture.snapshotRequests.at(-1)).toContain(
      fixture.requests[0].threadId as string
    )
  })

  test('switching workspace shows that workspace\'s own thread', async ({
    page,
  }) => {
    const fixture = await setupAGUI(
      page,
      {
        runs: [
          sse([
            runStarted('r1'),
            ...cardCreated('card-1'),
            ...text('a1', 'Here.'),
            runFinished('r1'),
          ]),
        ],
      },
      {
        workspaces: [
          { id: 'default', name: 'Default', slug: 'default' },
          { id: 'team-b', name: 'Team B', slug: 'team-b' },
        ],
      }
    )
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'deploy')
    await expect(card(page)).toHaveCount(1)
    const firstThread = fixture.requests[0].threadId as string

    await page.getByRole('button', { name: /Default/ }).first().click()
    await page.getByRole('menuitem', { name: 'Team B' }).click()
    await expect(card(page)).toHaveCount(0)
    await expect(page.getByText('Here.')).toHaveCount(0)
    await expect
      .poll(() =>
        fixture.snapshotRequests.some(
          (u) => u.includes('/streamer-id/') && !u.includes(firstThread)
        )
      )
      .toBe(true)
  })
})
