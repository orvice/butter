import { test, expect, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import * as w from '../src/gen/agents/v1/workstation_pb'
import { GitHostKind, ListGitHostsResponseSchema } from '../src/gen/agents/v1/githost_pb'
import { MeResponseSchema } from '../src/gen/agents/v1/auth_pb'
import { ListWorkspaceMembersResponseSchema } from '../src/gen/agents/v1/workspace_pb'
import { fulfillConnectError, fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

const connection = () => create(w.WorkStationConnectionSchema, { id: 'c1', name: 'Engineering', provider: 'github', hostId: 'github', principalId: '123', revision: 1n, credentialSet: true, scopes: [{ id: '42', name: 'org/repo', initialized: true, health: 'HEALTHY', coverage: 'COMPLETE' }] })
const item = (id = 'item1', title = 'Fix the queue') => create(w.WorkStationItemSchema, { id, title, provider: 'github', kind: 'ISSUE', remoteId: '1', identifier: '#1', scopeId: '42', scopeName: 'org/repo', status: 'NOT_STARTED', rawState: 'open', description: '<img src="https://evil.invalid/track" onerror="alert(1)"> [unsafe](javascript:alert(1))', descriptionAvailable: true, url: 'https://github.com/org/repo/issues/1', sources: [{ connectionId: 'c1', scopeId: '42' }] })
async function hosts(route: Route, url: string) {
  if (url.includes('GitHostService/ListGitHosts')) return fulfillProto(route, ListGitHostsResponseSchema, { hosts: [{ id: 'github', name: 'GitHub.com', kind: GitHostKind.GITHUB, apiBaseUrl: 'https://api.github.com', webBaseUrl: 'https://github.com' }] })
  return false
}

test('creates with empty selection, explicitly shares repositories, clears the key and deletes', async ({ page }) => {
  let c: w.WorkStationConnection | undefined
  const keys: string[] = []
  let selected: string[] = []
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (await hosts(route, url)) return true
    if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: c ? [c] : [] })
    if (url.includes('WorkStationService/ListItems')) return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, {})
    if (url.includes('WorkStationService/CreateConnection')) {
      const req = fromBinary(w.WorkStationServiceCreateConnectionRequestSchema, route.request().postDataBuffer()!)
      keys.push(req.apiKey); c = connection(); c.name = req.name; c.scopes = []
      return fulfillProto(route, w.WorkStationServiceCreateConnectionResponseSchema, { connection: c })
    }
    if (url.includes('WorkStationService/DiscoverScopes')) return fulfillProto(route, w.WorkStationServiceDiscoverScopesResponseSchema, { scopes: [{ id: '42', name: 'org/repo' }, { id: '43', name: 'org/private-unselected' }] })
    if (url.includes('WorkStationService/SetSelectedScopes')) {
      const req = fromBinary(w.WorkStationServiceSetSelectedScopesRequestSchema, route.request().postDataBuffer()!)
      expect(req.expectedRevision).toBe(c!.revision); selected = req.scopeIds
      c!.revision++; c!.scopes = create(w.WorkStationConnectionSchema, { scopes: req.scopeIds.map((id) => ({ id, name: 'org/repo', health: 'QUEUED' })) }).scopes
      return fulfillProto(route, w.WorkStationServiceCreateConnectionResponseSchema, { connection: c })
    }
    if (url.includes('WorkStationService/UpdateConnection')) {
      const req = fromBinary(w.WorkStationServiceUpdateConnectionRequestSchema, route.request().postDataBuffer()!)
      expect(req.expectedRevision).toBe(c!.revision); c!.revision++
      if (req.apiKey !== undefined) { c!.credentialSet = req.apiKey !== ''; c!.paused = req.apiKey === '' }
      return fulfillProto(route, w.WorkStationServiceCreateConnectionResponseSchema, { connection: c })
    }
    if (url.includes('WorkStationService/DeleteConnection')) { c = undefined; return fulfillProto(route, w.WorkStationServiceDeleteConnectionResponseSchema, {}) }
    return false
  })
  await page.goto('/work-station')
  await expect(page.getByText('No connections yet')).toBeVisible()
  await page.getByRole('button', { name: 'Manage connections' }).click()
  await expect(page.getByText('Selected repositories and their descriptions')).toBeVisible()
  await page.getByLabel('Connection name').fill('Engineering')
  await page.getByLabel('Git Host', { exact: true }).selectOption('github')
  await page.getByLabel('API key', { exact: true }).fill('private-token-for-test')
  await page.getByRole('button', { name: 'Create connection', exact: true }).click()
  await expect(page.getByLabel('Replace API key')).toHaveValue('')
  expect(c!.scopes).toHaveLength(0)
  await page.getByRole('button', { name: 'Discover repositories' }).click()
  await expect(page.getByLabel('org/private-unselected')).toBeVisible()
  await expect(page.getByLabel('org/repo', { exact: true })).not.toBeChecked()
  await page.getByLabel('org/repo', { exact: true }).check()
  await page.getByRole('button', { name: 'Save selected repositories' }).click()
  await expect.poll(() => selected).toEqual(['42'])
  await page.getByRole('button', { name: 'Clear API key' }).click()
  await expect.poll(() => c!.credentialSet).toBe(false)
  await page.getByRole('button', { name: 'Delete connection', exact: true }).click()
  await page.getByRole('button', { name: 'Confirm deletion' }).click()
  await expect(page.getByRole('button', { name: 'Create connection', exact: true })).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.getByText('No connections yet')).toBeVisible()
  expect(keys).toEqual(['private-token-for-test'])
  expect(await page.evaluate(() => JSON.stringify(localStorage))).not.toContain('private-token-for-test')
  expect(page.url()).not.toContain('private-token-for-test')
})

test('plain-text detail blocks HTML and links; paging, tabs, filters and refresh work', async ({ page }) => {
  let kind = ''
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: [connection()] })
    if (url.includes('WorkStationService/ListItems')) {
      const req = fromBinary(w.WorkStationServiceListItemsRequestSchema, route.request().postDataBuffer()!); kind = req.kind
      return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, { items: req.pageToken ? [item('second', 'Second page')] : [item()], nextPageToken: req.pageToken ? '' : 'next-page' })
    }
    if (url.includes('WorkStationService/GetItem')) return fulfillProto(route, w.WorkStationServiceGetItemResponseSchema, { item: item() })
    if (url.includes('WorkStationService/Refresh')) return fulfillProto(route, w.WorkStationServiceRefreshResponseSchema, { status: 'QUEUED' })
    return false
  })
  await page.goto('/work-station')
  await page.getByRole('button', { name: /Fix the queue/ }).click()
  await expect(page.getByRole('dialog').getByText('<img src=', { exact: false })).toBeVisible()
  await expect(page.getByRole('dialog').locator('img')).toHaveCount(0)
  await expect(page.getByRole('dialog').locator('a[href^="javascript:"]')).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'Open on GitHub' })).toHaveAttribute('href', 'https://github.com/org/repo/issues/1')
  await page.keyboard.press('Escape')
  await page.getByRole('button', { name: 'Next page' }).click()
  await expect(page.getByRole('button', { name: /Second page/ })).toBeVisible()
  await page.getByRole('tab', { name: 'Pull requests' }).click()
  await expect.poll(() => kind).toBe('PULL_REQUEST')
  await expect(page.getByRole('button', { name: /Fix the queue/ })).toBeVisible()
  await page.getByRole('button', { name: 'Request refresh' }).click()
  await expect(page.getByRole('status').filter({ hasText: 'QUEUED' })).toBeVisible()
  await page.getByLabel('Search title or identifier').fill('queue')
  await expect(page).toHaveURL(/q=queue/)
})

for (const role of ['member', 'admin']) {
  test(`workspace ${role} can browse and refresh without management`, async ({ page }) => {
    await setupAuthenticatedConnectRoutes(page, async (route, url) => {
      if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: [connection()] })
      if (url.includes('WorkStationService/ListItems')) return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, { items: [item()] })
      return false
    })
    await page.route('**/api/agents.v1.AuthService/Me', (route) => fulfillProto(route, MeResponseSchema, { user: { id: 'test-user-1', username: 'member', role: 'user' } }))
    await page.route('**/api/agents.v1.WorkspaceService/ListWorkspaceMembers', (route) => fulfillProto(route, ListWorkspaceMembersResponseSchema, { members: [{ workspaceId: 'default', userId: 'test-user-1', role }] }))
    await page.goto('/work-station')
    await expect(page.getByRole('button', { name: /Fix the queue/ })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Request refresh' })).toBeEnabled()
    await expect(page.getByRole('button', { name: 'Manage connections' })).toHaveCount(0)
  })
}

test('switching workspace clears detail and ignores a late old response', async ({ page }) => {
  let release: (() => void) | undefined
  let delay = false; let requested = false
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    const workspace = route.request().headers()['x-workspace-id']
    if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: workspace === 'A' ? [connection()] : [] })
    if (url.includes('WorkStationService/ListItems')) {
      if (workspace === 'A' && delay) { requested = true; await new Promise<void>((resolve) => { release = resolve }) }
      return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, { items: workspace === 'A' ? [item()] : [] })
    }
    if (url.includes('WorkStationService/GetItem')) return fulfillProto(route, w.WorkStationServiceGetItemResponseSchema, { item: item() })
    return false
  }, { workspaces: [{ id: 'A', name: 'Workspace Alpha', slug: 'alpha' }, { id: 'B', name: 'Workspace Beta', slug: 'beta' }] })
  await page.goto('/work-station')
  await page.getByRole('button', { name: /Fix the queue/ }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.keyboard.press('Escape')
  delay = true
  await page.getByLabel('Search title or identifier').fill('delayed')
  await expect.poll(() => requested).toBe(true)
  await page.getByRole('button', { name: /Workspace Alpha/ }).click()
  await page.getByRole('menuitem', { name: 'Workspace Beta' }).click()
  release?.()
  await expect(page.getByText('No connections yet')).toBeVisible()
  await expect(page.getByRole('button', { name: /Fix the queue/ })).toHaveCount(0)
  await expect(page.getByRole('dialog')).toHaveCount(0)
})

test('revoked detail renders no cached source body', async ({ page }) => {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: [connection()] })
    if (url.includes('WorkStationService/ListItems')) return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, { items: [item()] })
    if (url.includes('WorkStationService/GetItem')) return fulfillConnectError(route, 'not_found', 'work station resource not found')
    return false
  })
  await page.goto('/work-station')
  await page.getByRole('button', { name: /Fix the queue/ }).click()
  await expect(page.getByRole('heading', { name: 'Item unavailable' })).toBeVisible()
  await expect(page.getByRole('dialog').getByRole('link')).toHaveCount(0)
})

test('refresh targets the connection filter and skips empty or paused connections', async ({ page }) => {
  const refreshed: string[] = []
  const one = connection()
  const two = create(w.WorkStationConnectionSchema, { ...one, id: 'c2', name: 'Second connection' })
  const empty = create(w.WorkStationConnectionSchema, { ...one, id: 'empty', name: 'No selection', scopes: [] })
  const paused = create(w.WorkStationConnectionSchema, { ...one, id: 'paused', name: 'Paused connection', paused: true })
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: [one, two, empty, paused] })
    if (url.includes('WorkStationService/ListItems')) return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, {})
    if (url.includes('WorkStationService/Refresh')) {
      const request = fromBinary(w.WorkStationServiceRefreshRequestSchema, route.request().postDataBuffer()!)
      refreshed.push(request.id)
      return fulfillProto(route, w.WorkStationServiceRefreshResponseSchema, { status: 'QUEUED' })
    }
    return false
  })
  await page.goto('/work-station')
  await page.getByLabel('Connection', { exact: true }).selectOption('c2')
  await page.getByRole('button', { name: 'Request refresh' }).click()
  await expect.poll(() => refreshed).toEqual(['c2'])
  await page.getByLabel('Connection', { exact: true }).selectOption('empty')
  await expect(page.getByRole('button', { name: 'Request refresh' })).toBeDisabled()
  await page.getByLabel('Connection', { exact: true }).selectOption('')
  await page.getByRole('button', { name: 'Request refresh' }).click()
  await expect.poll(() => refreshed).toEqual(['c2', 'c1', 'c2'])
})
