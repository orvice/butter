import { expect, test, type Page } from '@playwright/test'
import { ListAgentsResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

// The dashboard AG-UI chat uses the official assistant-ui AG-UI runtime with
// HttpAgent. Fixtures fulfill POST /api/agui/:agent_id with literal SSE event
// frames. The runtime handles parsing, message reconstruction, and state.

function sse(events: Array<Record<string, unknown>>): string {
  return events.map((ev) => `data: ${JSON.stringify(ev)}\n\n`).join('')
}

async function setupAGUI(
  page: Page,
  runs: string[],
  requests: Array<Record<string, unknown>>
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [
          {
            name: 'Streamer',
            agentId: 'streamer-id',
            description: 'AG-UI enabled',
            enableAgui: true,
            lifecycleStatus: 1,
          },
          {
            name: 'Plain',
            agentId: 'plain-id',
            description: 'not exposed',
            enableAgui: false,
            lifecycleStatus: 1,
          },
        ],
        total: 2,
      })
    }
    return false
  })

  await page.route('**/api/agui/**', async (route) => {
    requests.push(JSON.parse(route.request().postData() ?? '{}'))
    const body = runs.shift() ?? sse([])
    await route.fulfill({
      status: 200,
      contentType: 'text/event-stream',
      body,
    })
  })
}

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
