import { expect, test, type Page } from '@playwright/test'

async function setupChat(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('butter_token', 'fake-test-token')
    localStorage.setItem('butter_workspace_id', 'default')
  })

  await page.route('**/api/agents.v1.**', async (route) => {
    const url = route.request().url()

    if (url.includes('AuthService/Me') || url.includes('auth.AuthService/Me')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          user: {
            id: 'test-user-1',
            username: 'testuser',
            displayName: 'Test User',
            email: 'test@example.com',
            role: 'admin',
            isAdmin: true,
          },
        }),
      })
    }

    if (url.includes('AuthService/IsAdmin')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ isAdmin: true }),
      })
    }

    if (url.includes('WorkspaceService')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          workspaces: [
            { id: 'default', name: 'default', displayName: 'Default' },
          ],
          workspace: { id: 'default', name: 'default', displayName: 'Default' },
        }),
      })
    }

    if (url.includes('AgentService/ListAgents')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          agents: [
            {
              name: 'ChatBot',
              agentId: 'chatbot-id',
              description: 'General assistant',
              lifecycleStatus: 1,
            },
          ],
          total: 1,
        }),
      })
    }

    if (url.includes('SessionService')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ sessions: [], total: 0 }),
      })
    }

    if (url.includes('DashboardService')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      })
    }

    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({}),
    })
  })
}

test.describe('AUI Chat', () => {
  test.beforeEach(async ({ page }) => {
    await setupChat(page)
  })

  test('/chat loads without JS errors', async ({ page }) => {
    const jsErrors: string[] = []
    page.on('pageerror', (err) => {
      if (
        !err.message.includes('net::ERR_') &&
        !err.message.includes('ConnectError')
      ) {
        jsErrors.push(err.message)
      }
    })

    await page.goto('/chat', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)

    const body = await page.locator('body').innerHTML()
    expect(body.length).toBeGreaterThan(0)
    expect(jsErrors).toHaveLength(0)
  })

  test('/chat?session=x passes session to the chat view', async ({
    page,
  }) => {
    const jsErrors: string[] = []
    page.on('pageerror', (err) => {
      if (
        !err.message.includes('net::ERR_') &&
        !err.message.includes('ConnectError')
      ) {
        jsErrors.push(err.message)
      }
    })

    await page.goto('/chat?session=test-session', {
      waitUntil: 'networkidle',
    })
    await page.waitForTimeout(1000)

    const parsed = new URL(page.url())
    expect(parsed.searchParams.get('session')).toBe('test-session')
    expect(jsErrors).toHaveLength(0)
  })

  test('no SessionService/CreateSession call on draft view', async ({
    page,
  }) => {
    let sessionCreated = false
    page.on('request', (req) => {
      if (req.url().includes('SessionService/CreateSession')) {
        sessionCreated = true
      }
    })

    await page.goto('/chat', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)
    expect(sessionCreated).toBe(false)
  })
})
