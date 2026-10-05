import { expect, test, type Page } from '@playwright/test'
import {
  aguiSession,
  sendMessage as send,
  setupAGUI,
  sse,
} from './support/agui'

// How Chat draws a conversation with the shared chat message
// components (src/components/chat): Markdown, the agent on every reply, tool
// calls, and the thread's loading, empty and failed states. A conversation
// starts from a new-chat draft with Streamer; a thread that exists opens by
// URL.

const NEW_CHAT = '/chat?agent=streamer-id'
const thread = (threadId: string) =>
  aguiSession(threadId, '', { agentId: 'streamer-id' })

const runStarted = (runId: string) => ({
  type: 'RUN_STARTED',
  threadId: 't',
  runId,
})
const runFinished = (runId: string) => ({
  type: 'RUN_FINISHED',
  threadId: 't',
  runId,
})
const text = (messageId: string, delta: string) => [
  { type: 'TEXT_MESSAGE_START', messageId, role: 'assistant' },
  { type: 'TEXT_MESSAGE_CONTENT', messageId, delta },
  { type: 'TEXT_MESSAGE_END', messageId },
]
const reply = (runId: string, delta: string) =>
  sse([runStarted(runId), ...text(`a-${runId}`, delta), runFinished(runId)])

function toolCall(id: string, name: string, args: string, result?: string) {
  return [
    { type: 'TOOL_CALL_START', toolCallId: id, toolCallName: name },
    { type: 'TOOL_CALL_ARGS', toolCallId: id, delta: args },
    { type: 'TOOL_CALL_END', toolCallId: id },
    ...(result === undefined
      ? []
      : [
          {
            type: 'TOOL_CALL_RESULT',
            messageId: `result-${id}`,
            toolCallId: id,
            content: result,
          },
        ]),
  ]
}

const replies = (page: Page) => page.locator('[data-message-role="assistant"]')
const sentMessages = (page: Page) =>
  page.locator('[data-message-role="user"]')
const hero = (page: Page) =>
  page.getByRole('heading', { name: 'Streamer', level: 2 })

test.describe('AG-UI chat messages', () => {
  test('a reply is Markdown: links open in a new tab, code and tables are styled', async ({
    page,
  }) => {
    const markdown = [
      'See [the runbook](https://example.com/runbook) and run `make build`.',
      '',
      '```go',
      'fmt.Println("deployed")',
      '```',
      '',
      '```',
      'plain block',
      '```',
      '',
      '| Service | State |',
      '| --- | --- |',
      '| api | healthy |',
    ].join('\n')
    await setupAGUI(page, { runs: [reply('r1', markdown)] })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'status?')

    const answer = replies(page).last()
    const link = answer.getByRole('link', { name: 'the runbook' })
    await expect(link).toHaveAttribute('href', 'https://example.com/runbook')
    await expect(link).toHaveAttribute('target', '_blank')
    await expect(link).toHaveAttribute('rel', 'noopener noreferrer')
    // Inline code stays in the line; fenced code is a block, with or
    // without a language, and is not styled as inline code.
    await expect(answer.locator('p > code')).toHaveText('make build')
    await expect(answer.locator('pre > code')).toHaveText([
      'fmt.Println("deployed")',
      'plain block',
    ])
    await expect(answer.locator('pre > code.bg-muted')).toHaveCount(0)
    const table = answer.getByRole('table')
    await expect(table.getByRole('columnheader')).toHaveText([
      'Service',
      'State',
    ])
    await expect(table.getByRole('cell')).toHaveText(['api', 'healthy'])
  })

  test('what the user sends is Markdown that keeps their lines', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [reply('r1', 'Noted.')] })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'Ship **only** `api`\nthen tell me')

    await expect(page.getByText('Noted.')).toBeVisible()
    const sent = sentMessages(page).last()
    await expect(sent.locator('strong')).toHaveText('only')
    await expect(sent.locator('code')).toHaveText('api')
    await expect(sent.locator('br')).toHaveCount(1)
  })

  test('every reply shows its agent', async ({ page }) => {
    await setupAGUI(page, {
      runs: [reply('r1', 'First answer.'), reply('r2', 'Second answer.')],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'one')
    await expect(page.getByText('First answer.')).toBeVisible()
    await send(page, 'two')
    await expect(page.getByText('Second answer.')).toBeVisible()

    await expect(replies(page)).toHaveCount(2)
    for (const answer of await replies(page).all()) {
      await expect(
        answer.getByText('Streamer', { exact: true })
      ).toBeVisible()
    }
    // A message the user sent carries no agent.
    await expect(sentMessages(page).getByText('Streamer')).toHaveCount(0)
  })

  test('a tool call opens to its arguments and result; render_ui stays hidden unless it failed', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        sse([
          runStarted('r1'),
          ...toolCall('tc-1', 'search', '{"q":"deploys"}', '{"hits":3}'),
          ...toolCall(
            'tc-2',
            'render_ui',
            '{"messages":[]}',
            '{"surface_id":"card-1","revision":1,"status":"created"}'
          ),
          ...text('a1', 'Found three.'),
          runFinished('r1'),
        ]),
        sse([
          runStarted('r2'),
          ...toolCall(
            'tc-3',
            'render_ui',
            '{"messages":[]}',
            '{"error":"card not rendered, nothing was changed"}'
          ),
          ...text('a2', 'No card this time.'),
          runFinished('r2'),
        ]),
      ],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'find deploys')
    await expect(page.getByText('Found three.')).toBeVisible()

    // Collapsed: the call is its name and state.
    const call = page.getByRole('button', { name: 'search', exact: true })
    await expect(call).toHaveAttribute('aria-expanded', 'false')
    await expect(call).toHaveAccessibleDescription('done')
    const view = page.locator('[data-tool-call="search"]')
    await expect(view.getByText('Arguments')).toHaveCount(0)

    await call.click()
    await expect(call).toHaveAttribute('aria-expanded', 'true')
    await expect(view.getByText('Arguments')).toBeVisible()
    await expect(view.locator('pre').first()).toContainText('"q": "deploys"')
    await expect(view.getByText('Result')).toBeVisible()
    await expect(view.locator('pre').last()).toContainText('"hits": 3')
    await call.click()
    await expect(view.getByText('Arguments')).toHaveCount(0)

    // render_ui drew a card: the reply shows the card, not the call.
    await expect(
      page.getByRole('button', { name: 'render_ui', exact: true })
    ).toHaveCount(0)
    // The toolkit's entries only draw calls: nothing is offered to the
    // agent as a tool the browser runs.
    expect(fixture.requests[0].tools ?? []).toEqual([])

    // A render_ui call that drew nothing shows, with its error.
    await send(page, 'and a card?')
    await expect(page.getByText('No card this time.')).toBeVisible()
    const failed = page.getByRole('button', { name: 'render_ui', exact: true })
    await expect(failed).toHaveCount(1)
    await failed.click()
    await expect(
      page.locator('[data-tool-call="render_ui"]').locator('pre').last()
    ).toContainText('card not rendered')
  })

  test('the toolkit draws a Human Input call as its question', async ({
    page,
  }) => {
    // Butter's AG-UI server never streams this call: the question arrives as
    // an Interrupt. Only the Chat before AG-UI (#409) received it; the shared
    // toolkit still draws it.
    await setupAGUI(page, {
      runs: [
        sse([
          runStarted('r1'),
          ...toolCall(
            'ask-1',
            'adk_request_input',
            '{"message":"Approve the deploy?"}'
          ),
          runFinished('r1'),
        ]),
      ],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'deploy')

    await expect(page.getByText('Approve the deploy?')).toBeVisible()
    await expect(page.getByText('Waiting for input')).toBeVisible()
    await expect(
      page.getByRole('button', { name: 'adk_request_input', exact: true })
    ).toHaveCount(0)
  })

  test('a new chat introduces its agent until the first message', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [reply('r1', 'Hello there.')] })
    // The new-chat draft introduces the agent it starts with.
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })

    await expect(hero(page)).toBeVisible()
    await expect(
      page.getByText('Send a message below to start the conversation.')
    ).toBeVisible()
    await expect(
      page.getByText(
        'Butter can make mistakes. Verify important actions before running them.'
      )
    ).toBeVisible()

    await send(page, 'hi')
    await expect(page.getByText('Hello there.')).toBeVisible()
    await expect(hero(page)).toHaveCount(0)
  })

  test('a thread shows a skeleton while its history loads, then introduces its agent when empty', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [], sessions: [thread('t-empty')] })
    let release = () => {}
    const held = new Promise<void>((resolve) => {
      release = resolve
    })
    await page.route('**/api/agui/*/threads/*/messages', async (route) => {
      await held
      await route.fallback()
    })
    await page.goto('/chat?thread=t-empty')

    const loading = page.getByRole('status', { name: 'Loading conversation' })
    await expect(loading).toBeVisible()
    await expect(hero(page)).toHaveCount(0)

    release()
    await expect(loading).toHaveCount(0)
    await expect(hero(page)).toBeVisible()
  })

  test('a reply that cannot be drawn says so, and the chat keeps working', async ({
    page,
  }) => {
    const pageErrors: string[] = []
    page.on('pageerror', (err) => pageErrors.push(err.message))
    await setupAGUI(page, {
      runs: [
        sse([
          runStarted('r1'),
          // A malformed card: its text version is not text, so drawing the
          // reply that holds it throws.
          {
            type: 'CUSTOM',
            name: 'butter.a2ui',
            value: {
              version: 'v0.9.1',
              threadId: 't',
              runId: 'r1',
              messageId: 'a1',
              surfaceId: 'card-x',
              kind: 'card',
              revision: 1,
              seq: 0,
              fallback: { not: 'text' },
              envelope: {
                version: 'v0.9.1',
                createSurface: {
                  surfaceId: 'card-x',
                  catalogId: 'https://example.com/other-catalog',
                },
              },
            },
          },
          ...text('a1', 'Here is the card.'),
          runFinished('r1'),
        ]),
        reply('r2', 'Still here.'),
      ],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'show me')

    const broken = replies(page).first()
    await expect(
      broken.getByText('This message could not be displayed.')
    ).toBeVisible()
    await expect(broken.getByText('Streamer', { exact: true })).toBeVisible()
    await expect(sentMessages(page).getByText('show me')).toBeVisible()

    await send(page, 'and now?')
    await expect(page.getByText('Still here.')).toBeVisible()
    expect(pageErrors).toEqual([])
  })

  test('a thread that cannot be drawn offers to try again, and the composer keeps working', async ({
    page,
  }) => {
    const pageErrors: string[] = []
    page.on('pageerror', (err) => pageErrors.push(err.message))
    await setupAGUI(page, { runs: [], sessions: [thread('t-broken')] })
    // The thread holds a malformed open question: its message is not text.
    await page.route('**/api/agui/*/threads/*/messages', (route) =>
      route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          threadId: 't-broken',
          messages: [
            { id: 'u1', role: 'user', content: 'deploy' },
            { id: 'a1', role: 'assistant', content: 'One question first.' },
          ],
          interrupts: [
            { id: 'int-1', reason: 'human_input', message: { not: 'text' } },
          ],
          surfaces: [],
        }),
      })
    )
    await page.goto('/chat?thread=t-broken', { waitUntil: 'networkidle' })

    await expect(
      page.getByRole('alert').filter({
        hasText: 'This conversation could not be displayed.',
      })
    ).toBeVisible()
    await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible()
    await expect(page.getByPlaceholder('Message the agent…')).toBeEditable()
    expect(pageErrors).toEqual([])
  })
})
