import { expect, test, type Page } from '@playwright/test'

const mockUser = {
  user: {
    id: 'test-user-1',
    username: 'testuser',
    displayName: 'Test User',
    email: 'test@example.com',
    role: 'admin',
    isAdmin: true,
  },
}

const AGENTS = {
  agents: [
    {
      name: 'CodeReviewer',
      agentId: 'code-reviewer-id',
      description: 'Reviews pull requests',
      lifecycleStatus: 1, // ACTIVE
    },
    {
      name: 'ChatBot',
      agentId: 'chatbot-id',
      description: 'General assistant',
      lifecycleStatus: 1, // ACTIVE
    },
    {
      name: 'Retired',
      agentId: 'retired-id',
      description: 'No longer used',
      lifecycleStatus: 6, // DELETED
    },
  ],
  total: 3,
}

async function setupChat(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('butter_token', 'fake-test-token')
    localStorage.setItem('butter_workspace_id', 'default')
  })

  await page.route('**/api/**', async (route) => {
    const url = route.request().url()

    if (url.includes('AuthService/Me') || url.includes('auth.AuthService/Me')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockUser),
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
        body: JSON.stringify(AGENTS),
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

test.describe('Chat draft view', () => {
  test.beforeEach(async ({ page }) => {
    await setupChat(page)
  })

  test('/chat shows the draft view and does not auto-activate a session', async ({
    page,
  }) => {
    await page.goto('/chat', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)

    const body = await page.locator('body').innerHTML()
    expect(body.length).toBeGreaterThan(0)
    expect(page.url()).not.toContain('session=')

    const jsErrors: string[] = []
    page.on('pageerror', (err) => jsErrors.push(err.message))
    expect(jsErrors).toHaveLength(0)
  })

  test('/chat?new=1 remains on /chat (backward compat)', async ({ page }) => {
    const jsErrors: string[] = []
    page.on('pageerror', (err) => {
      if (!err.message.includes('net::ERR_') && !err.message.includes('ConnectError')) {
        jsErrors.push(err.message)
      }
    })

    await page.goto('/chat?new=1', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)

    const url = new URL(page.url())
    expect(url.pathname).toBe('/chat')
    expect(jsErrors).toHaveLength(0)
  })

  test('/chat?agent=code-reviewer-id loads without error', async ({
    page,
  }) => {
    const jsErrors: string[] = []
    page.on('pageerror', (err) => {
      if (!err.message.includes('net::ERR_') && !err.message.includes('ConnectError')) {
        jsErrors.push(err.message)
      }
    })

    await page.goto('/chat?agent=code-reviewer-id', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)

    expect(page.url()).not.toContain('session=')
    expect(jsErrors).toHaveLength(0)
  })

  test('/chat?session=nonexistent loads without error', async ({ page }) => {
    const jsErrors: string[] = []
    page.on('pageerror', (err) => {
      if (!err.message.includes('net::ERR_') && !err.message.includes('ConnectError')) {
        jsErrors.push(err.message)
      }
    })

    await page.goto('/chat?session=nonexistent-session-id', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)

    expect(jsErrors).toHaveLength(0)
  })

  test('no SessionService/CreateSession call on /chat', async ({ page }) => {
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

  test('no SessionService/CreateSession call on /chat?agent=x', async ({ page }) => {
    let sessionCreated = false
    page.on('request', (req) => {
      if (req.url().includes('SessionService/CreateSession')) {
        sessionCreated = true
      }
    })

    await page.goto('/chat?agent=code-reviewer-id', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)
    expect(sessionCreated).toBe(false)
  })
})
