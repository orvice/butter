import { expect, test, type Page } from '@playwright/test'
import {
  resumeEntries,
  setupAGUI as setupAGUIFixture,
  sse,
} from './support/agui'

// The dashboard AG-UI chat uses the official assistant-ui AG-UI runtime with
// HttpAgent. Fixtures fulfill POST /api/agui/:agent_id with literal SSE event
// frames. The runtime handles parsing, message reconstruction, and state.

async function setupAGUI(
  page: Parameters<typeof setupAGUIFixture>[0],
  runs: string[],
  requests: Array<Record<string, unknown>>
) {
  await setupAGUIFixture(page, { runs, requests })
}

const COMPOSER_HINT = 'Sending answers the earliest open question.'

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

// pausedOn is a run that replies, then pauses on open questions.
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

function finished(runId: string, text: string) {
  return sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    ...reply(runId, text),
    { type: 'RUN_FINISHED', threadId: 't', runId },
  ])
}

async function send(page: Page, message: string) {
  const composer = page.getByPlaceholder(/Message the agent over AG-UI/)
  await composer.fill(message)
  await composer.press('Enter')
}

// prompt is the text prompt of the open question asking `message`.
const prompt = (page: Page, message: string) =>
  page.getByRole('group', { name: message })

test.describe('AG-UI chat', () => {
  test('streams text and tool calls, then resumes an interrupt by id', async ({
    page,
  }) => {
    const requests: Array<Record<string, unknown>> = []
    const firstRun = sse([
      { type: 'RUN_STARTED', threadId: 't', runId: 'r1' },
      { type: 'STATE_SNAPSHOT', snapshot: { draft: 'v1' } },
      { type: 'TOOL_CALL_START', toolCallId: 'tc-1', toolCallName: 'search' },
      { type: 'TOOL_CALL_ARGS', toolCallId: 'tc-1', delta: '{"q":"go"}' },
      { type: 'TOOL_CALL_END', toolCallId: 'tc-1' },
      {
        type: 'TOOL_CALL_RESULT',
        messageId: 'm1',
        toolCallId: 'tc-1',
        content: '{"hits":3}',
      },
      { type: 'TEXT_MESSAGE_START', messageId: 'a1', role: 'assistant' },
      { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a1', delta: 'Found it. ' },
      { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a1', delta: 'Deploying…' },
      { type: 'TEXT_MESSAGE_END', messageId: 'a1' },
      {
        type: 'RUN_FINISHED',
        threadId: 't',
        runId: 'r1',
        outcome: {
          type: 'interrupt',
          interrupts: [
            { id: 'int-1', reason: 'human_input', message: 'Approve deploy?' },
          ],
        },
      },
    ])
    const secondRun = sse([
      { type: 'RUN_STARTED', threadId: 't', runId: 'r2' },
      { type: 'TEXT_MESSAGE_START', messageId: 'a2', role: 'assistant' },
      { type: 'TEXT_MESSAGE_CONTENT', messageId: 'a2', delta: 'Deployed.' },
      { type: 'TEXT_MESSAGE_END', messageId: 'a2' },
      { type: 'RUN_FINISHED', threadId: 't', runId: 'r2' },
    ])
    await setupAGUI(page, [firstRun, secondRun], requests)

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })

    const composer = page.getByPlaceholder(/Message the agent over AG-UI/)
    await composer.fill('ship it')
    await composer.press('Enter')

    // Streamed text renders.
    await expect(page.getByText('Found it. Deploying…')).toBeVisible()
    // Tool call renders with name visible.
    await expect(
      page.getByRole('button', { name: 'search', exact: true })
    ).toBeVisible()
    // Shared state panel appears.
    await expect(page.getByText('Shared state')).toBeVisible()

    // The interrupt becomes an addressed prompt.
    await expect(page.getByText('Approve deploy?')).toBeVisible()
    const answer = page.getByPlaceholder('Type your answer…')
    await answer.fill('yes')
    await page.getByRole('button', { name: 'Answer' }).click()

    await expect(page.getByText('Deployed.')).toBeVisible()

    expect(requests).toHaveLength(2)
    // First run sends the user message to the enabled agent.
    expect(JSON.stringify(requests[0])).toContain('ship it')
    // The resume addresses the interrupt by id.
    const resume = requests[1].resume as Array<Record<string, unknown>>
    expect(resume).toHaveLength(1)
    expect(resume[0].interruptId).toBe('int-1')
    expect(resume[0].status).toBe('resolved')
    expect(resume[0].payload).toBe('yes')
    // The state mirror travels back so the server can validate it.
    expect(requests[1].state).toEqual({ draft: 'v1' })
    // Both runs stay on one AG-UI thread.
    expect(requests[1].threadId).toBe(requests[0].threadId)
  })

  test('answers one of several open questions from its prompt', async ({
    page,
  }) => {
    const fixture = await setupAGUIFixture(page, {
      runs: [
        pausedOn(
          'r1',
          'Two things first.',
          question('int-1', 'Which region?'),
          question('int-2', 'Which version?')
        ),
        // The server lists every question still open.
        pausedOn('r2', 'Version noted.', question('int-1', 'Which region?')),
      ],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'deploy')

    await expect(prompt(page, 'Which region?')).toBeVisible()
    await prompt(page, 'Which version?')
      .getByPlaceholder('Type your answer…')
      .fill('2.4.1')
    await prompt(page, 'Which version?')
      .getByRole('button', { name: 'Answer' })
      .click()

    await expect(page.getByText('Version noted.')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
    // Exactly one addressed entry: the other question stays open.
    expect(fixture.requests[1].resume).toEqual([
      { interruptId: 'int-2', status: 'resolved', payload: '2.4.1' },
    ])
    // The answer reads as the user's reply.
    await expect(page.getByText('2.4.1', { exact: true })).toBeVisible()
    await expect(prompt(page, 'Which version?')).toHaveCount(0)
    await expect(prompt(page, 'Which region?')).toBeVisible()
    expect(resumeEntries(fixture.requests).map((e) => e.status)).not.toContain(
      'cancelled'
    )
  })

  test('the composer answers the earliest open question with a plain message', async ({
    page,
  }) => {
    const fixture = await setupAGUIFixture(page, {
      runs: [
        pausedOn('r1', 'Ready to ship.', question('int-1', 'Approve deploy?')),
        finished('r2', 'Deployed.'),
      ],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'ship it')

    await expect(prompt(page, 'Approve deploy?')).toBeVisible()
    await expect(page.getByText(COMPOSER_HINT)).toBeVisible()
    const composer = page.getByPlaceholder(/Message the agent over AG-UI/)
    await composer.fill('yes, go ahead')
    await composer.press('Enter')

    await expect(page.getByText('Deployed.')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
    // No resume: the server answers its oldest open question (ADR-0002).
    expect(fixture.requests[1]).not.toHaveProperty('resume')
    const messages = fixture.requests[1].messages as Array<
      Record<string, unknown>
    >
    expect(messages.at(-1)).toMatchObject({
      role: 'user',
      content: 'yes, go ahead',
    })
    await expect(page.getByText('yes, go ahead', { exact: true })).toBeVisible()
    await expect(composer).toHaveValue('')
    await expect(prompt(page, 'Approve deploy?')).toHaveCount(0)
    await expect(page.getByText(COMPOSER_HINT)).toHaveCount(0)
  })

  test('the Send button answers too, and questions left open come back', async ({
    page,
  }) => {
    const fixture = await setupAGUIFixture(page, {
      runs: [
        pausedOn(
          'r1',
          'Two things first.',
          question('int-1', 'Which region?'),
          question('int-2', 'Which version?')
        ),
        pausedOn('r2', 'Region noted.', question('int-2', 'Which version?')),
      ],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'deploy')

    await expect(prompt(page, 'Which version?')).toBeVisible()
    await page.getByPlaceholder(/Message the agent over AG-UI/).fill('eu-west')
    // The dev server's router devtools badge covers the button's corner, so
    // it is activated from the keyboard.
    await page.getByRole('button', { name: 'Send' }).focus()
    await page.keyboard.press('Enter')

    await expect(page.getByText('Region noted.')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
    expect(fixture.requests[1]).not.toHaveProperty('resume')
    const messages = fixture.requests[1].messages as Array<
      Record<string, unknown>
    >
    expect(messages.at(-1)).toMatchObject({ role: 'user', content: 'eu-west' })
    await expect(page.getByText('eu-west', { exact: true })).toBeVisible()
    // What the server still reports open is asked again.
    await expect(prompt(page, 'Which version?')).toBeVisible()
    await expect(prompt(page, 'Which region?')).toHaveCount(0)
    await expect(page.getByText(COMPOSER_HINT)).toBeVisible()
  })

  test('a refused answer is reported once, and the message stays', async ({
    page,
  }) => {
    const fixture = await setupAGUIFixture(page, {
      runs: [
        pausedOn('r1', 'Ready to ship.', question('int-1', 'Approve deploy?')),
        { status: 409, body: { error: 'a run is in progress on this thread' } },
      ],
    })
    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    await send(page, 'ship it')

    await expect(prompt(page, 'Approve deploy?')).toBeVisible()
    await send(page, 'yes')

    const toast = page
      .locator('[data-sonner-toast]')
      .filter({ hasText: 'a run is in progress on this thread' })
    await expect(toast).toBeVisible()
    await expect(toast).toHaveCount(1)
    expect(fixture.requests).toHaveLength(2)
    expect(fixture.requests[1]).not.toHaveProperty('resume')
    await expect(page.getByText('yes', { exact: true })).toBeVisible()
  })

  test('renders RUN_ERROR in-band', async ({ page }) => {
    const requests: Array<Record<string, unknown>> = []
    const run = sse([
      { type: 'RUN_STARTED', threadId: 't', runId: 'r1' },
      { type: 'RUN_ERROR', message: 'model exploded', runId: 'r1' },
    ])
    await setupAGUI(page, [run], requests)

    await page.goto('/agui-chat', { waitUntil: 'networkidle' })
    const composer = page.getByPlaceholder(/Message the agent over AG-UI/)
    await composer.fill('hi')
    await composer.press('Enter')

    await expect(page.getByText('model exploded')).toBeVisible()
  })
})
