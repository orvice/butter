import { expect, test, type Page } from '@playwright/test'
import { aguiSession, setupAGUI } from './support/agui'

// AG-UI Chat is Chat, at /chat (#409). Links to its old route, /agui-chat,
// land there with their query. The parameters of the Chat that came before
// land on a new-chat draft: its web-chat conversations are not carried over.
// Every way into Chat opens it, with the agent a link names preselected: the
// sidebar, the command palette, the Agents page and the dashboard.

const trip = () =>
  aguiSession('t-trip', 'Trip plan', { agentId: 'streamer-id' })

// draftComposer is the new-chat draft's composer, which names the agent the
// draft starts with.
const draftComposer = (page: Page) =>
  page.getByRole('textbox', { name: 'Message', exact: true })

async function expectDraftWith(page: Page, agentName: string) {
  await expect(draftComposer(page)).toHaveAttribute(
    'placeholder',
    `Message ${agentName}…`
  )
}

// expectEmptyDraft: a new chat no agent was picked for yet.
async function expectEmptyDraft(page: Page) {
  await expect(
    page.getByRole('heading', { name: 'Start a new chat' })
  ).toBeVisible()
  await expect(draftComposer(page)).toBeDisabled()
}

const sidebar = (page: Page) => page.locator('[data-sidebar="sidebar"]')

test.describe('Chat at /chat', () => {
  test('links to /agui-chat land on /chat with their query', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [], sessions: [trip()] })

    await page.goto('/agui-chat?thread=t-trip')
    await expect(page).toHaveURL(/\/chat\?thread=t-trip$/)
    await expect(
      page.getByRole('heading', { name: 'Trip plan', level: 1 })
    ).toBeVisible()

    await page.goto('/agui-chat?agent=second-id')
    await expect(page).toHaveURL(/\/chat\?agent=second-id$/)
    await expectDraftWith(page, 'Second')
  })

  test('the old Chat’s links land on a new-chat draft', async ({ page }) => {
    const fixture = await setupAGUI(page, { runs: [] })

    await page.goto(
      '/chat?session=chat-1&pending_message=hello&invocation=inv-1'
    )
    await expect(page).toHaveURL(/\/chat$/)
    await expectEmptyDraft(page)

    // One that named an agent starts the draft with it.
    await page.goto('/chat?new=1&agent=second-id')
    await expect(page).toHaveURL(/\/chat\?agent=second-id$/)
    await expectDraftWith(page, 'Second')

    // No web-chat session was read, nor listed.
    expect(fixture.sessionCalls.gets).toEqual([])
    expect(fixture.sessionCalls.lists.map((c) => c.appName)).not.toContain(
      'web-chat'
    )
  })

  test('the sidebar and the command palette have one Chat, which opens a new chat', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [] })
    await page.goto('/agents', { waitUntil: 'networkidle' })

    const chatLink = sidebar(page).getByRole('link', {
      name: 'Chat',
      exact: true,
    })
    await expect(chatLink).toHaveCount(1)
    await expect(
      sidebar(page).getByRole('link', { name: 'AG-UI Chat' })
    ).toHaveCount(0)
    await chatLink.click()
    await expect(page).toHaveURL(/\/chat$/)
    await expectEmptyDraft(page)

    await page.goto('/agents', { waitUntil: 'networkidle' })
    await page.keyboard.press('ControlOrMeta+k')
    const palette = page.getByRole('dialog')
    const chatCommand = palette.getByRole('option', {
      name: 'Chat',
      exact: true,
    })
    await expect(chatCommand).toHaveCount(1)
    await expect(
      palette.getByRole('option', { name: 'AG-UI Chat' })
    ).toHaveCount(0)
    await chatCommand.click()
    await expect(page).toHaveURL(/\/chat$/)
    await expectEmptyDraft(page)
  })

  test('Start chat on the Agents page opens a new chat with that agent', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [] })
    await page.goto('/agents', { waitUntil: 'networkidle' })

    // The card of the agent named Second: the innermost element holding its
    // name and its Start chat.
    const card = page
      .locator('div')
      .filter({ has: page.getByRole('heading', { name: 'Second', exact: true }) })
      .filter({ has: page.getByRole('link', { name: 'Start chat' }) })
      .last()
    // The AG-UI badge only tells the agent takes AG-UI from API tokens.
    await expect(card.getByText('AG-UI', { exact: true })).toBeVisible()
    await expect(card.getByRole('link', { name: 'AG-UI' })).toHaveCount(0)

    await card.getByRole('link', { name: 'Start chat' }).click()
    await expect(page).toHaveURL(/\/chat\?agent=second-id$/)
    await expectDraftWith(page, 'Second')
  })

  test('the dashboard’s New chat opens a new chat', async ({ page }) => {
    await setupAGUI(page, { runs: [] })
    await page.goto('/', { waitUntil: 'networkidle' })

    await page.getByRole('link', { name: 'New chat', exact: true }).click()
    await expect(page).toHaveURL(/\/chat$/)
    await expectEmptyDraft(page)
  })

  test('a dashboard quick start opens a new chat with its agent', async ({
    page,
  }) => {
    await setupAGUI(page, { runs: [] })
    await page.goto('/', { waitUntil: 'networkidle' })

    const quickStart = page
      .locator('section')
      .filter({ has: page.getByRole('heading', { name: 'Start a chat' }) })
    // Only agents a chat can start with: the deleted one is left out.
    await expect(quickStart.getByRole('link')).toHaveCount(3)
    await expect(
      quickStart.getByRole('link', { name: 'Retired' })
    ).toHaveCount(0)
    const second = quickStart.getByRole('link', {
      name: 'Second',
      exact: true,
    })
    // The link names the agent by its Agent ID, which Chat matches.
    await expect(second).toHaveAttribute('href', '/chat?agent=second-id')

    await second.click()
    await expect(page).toHaveURL(/\/chat\?agent=second-id$/)
    await expectDraftWith(page, 'Second')
  })
})
