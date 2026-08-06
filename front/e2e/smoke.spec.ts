import { test, expect, type Page } from '@playwright/test'

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

const emptyList = { items: [], total: 0 }

async function setupAuth(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('butter_token', 'fake-test-token')
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

    if (url.includes('DashboardService')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          agentCount: 0,
          channelCount: 0,
          sessionCount: 0,
          cronJobCount: 0,
        }),
      })
    }

    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [],
        agents: [],
        channels: [],
        sessions: [],
        mcpServers: [],
        modelProviders: [],
        remoteAgents: [],
        notifyGroups: [],
        cronJobs: [],
        executions: [],
        threads: [],
        posts: [],
        tokens: [],
        files: [],
        fileSpaces: [],
        users: [],
        invocations: [],
        total: 0,
      }),
    })
  })
}

async function checkPageLoads(page: Page, path: string, name: string) {
  const errors: string[] = []

  page.on('pageerror', (err) => {
    errors.push(`${err.name}: ${err.message}`)
  })

  const consoleErrors: string[] = []
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      const text = msg.text()
      if (
        !text.includes('net::ERR_') &&
        !text.includes('Failed to fetch') &&
        !text.includes('NetworkError') &&
        !text.includes('AbortError') &&
        !text.includes('ConnectError')
      ) {
        consoleErrors.push(text)
      }
    }
  })

  const response = await page.goto(path, { waitUntil: 'networkidle' })
  expect(response?.status(), `${name}: HTTP status`).toBeLessThan(500)

  await page.waitForTimeout(1000)

  const body = await page.locator('body').innerHTML()
  expect(body.length, `${name}: page has content`).toBeGreaterThan(0)

  if (errors.length > 0) {
    console.warn(`  ⚠️  JS errors on ${name} (${path}):`)
    errors.forEach((e) => console.warn(`    - ${e}`))
  }

  return { path, name, jsErrors: errors, consoleErrors }
}

test.describe('Public pages smoke test', () => {
  test('sign-in page loads', async ({ page }) => {
    const result = await checkPageLoads(page, '/sign-in', 'Sign In')
    expect(result.jsErrors, 'No JS errors on sign-in').toHaveLength(0)
  })

  test('401 error page loads', async ({ page }) => {
    const result = await checkPageLoads(page, '/401', '401 Page')
    expect(result.jsErrors, 'No JS errors on 401').toHaveLength(0)
  })

  test('403 error page loads', async ({ page }) => {
    const result = await checkPageLoads(page, '/403', '403 Page')
    expect(result.jsErrors, 'No JS errors on 403').toHaveLength(0)
  })

  test('404 error page loads', async ({ page }) => {
    const result = await checkPageLoads(page, '/404', '404 Page')
    expect(result.jsErrors, 'No JS errors on 404').toHaveLength(0)
  })

  test('500 error page loads', async ({ page }) => {
    const result = await checkPageLoads(page, '/500', '500 Page')
    expect(result.jsErrors, 'No JS errors on 500').toHaveLength(0)
  })

  test('503 error page loads', async ({ page }) => {
    const result = await checkPageLoads(page, '/503', '503 Page')
    expect(result.jsErrors, 'No JS errors on 503').toHaveLength(0)
  })
})

test.describe('Authenticated pages smoke test', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page)
  })

  const authenticatedPages = [
    { path: '/', name: 'Dashboard' },
    { path: '/agents', name: 'Agent List' },
    { path: '/agents/create', name: 'Create Agent' },
    { path: '/automations', name: 'Automations' },
    { path: '/automations/create', name: 'Create Automation' },
    { path: '/sessions', name: 'Sessions' },
    { path: '/chat', name: 'Chat' },
    { path: '/forum', name: 'Forum' },
    { path: '/mcp-servers', name: 'MCP Servers' },
    { path: '/mcp-servers/create', name: 'Create MCP Server' },
    { path: '/model-providers', name: 'Model Providers' },
    { path: '/model-providers/create', name: 'Create Model Provider' },
    { path: '/remote-agents', name: 'Remote Agents' },
    { path: '/remote-agents/create', name: 'Create Remote Agent' },
    { path: '/channels', name: 'Channels' },
    { path: '/channels/create', name: 'Create Channel' },
    { path: '/notify-groups', name: 'Notify Groups' },
    { path: '/notify-groups/create', name: 'Create Notify Group' },
    { path: '/agent-files', name: 'Agent Files' },
    { path: '/daemons', name: 'Daemons' },
    { path: '/integrations', name: 'Integrations' },
    { path: '/api-tokens', name: 'API Tokens' },
    { path: '/users', name: 'Users' },
    { path: '/workspaces', name: 'Workspaces' },
    { path: '/profile', name: 'Profile' },
    { path: '/settings', name: 'Settings' },
    { path: '/settings/appearance', name: 'Settings Appearance' },
    { path: '/settings/display', name: 'Settings Display' },
    { path: '/admin/users', name: 'Admin Users' },
    { path: '/admin/global-mcp-servers', name: 'Admin Global MCP' },
    { path: '/manage', name: 'Manage' },
  ]

  for (const { path, name } of authenticatedPages) {
    test(`${name} (${path}) loads without errors`, async ({ page }) => {
      const result = await checkPageLoads(page, path, name)
      expect(result.jsErrors, `No JS errors on ${name}`).toHaveLength(0)
    })
  }
})
