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
  if (url.includes('GitHostService/ListGitHosts')) return fulfillProto(route, ListGitHostsResponseSchema, { hosts: [{ id: 'github', name: 'GitHub.com', kind: GitHostKind.GITHUB, apiBaseUrl: 'https://api.github.com', webBaseUrl: 'https://github.com' }, { id: 'gitlab', name: 'GitLab.com', kind: GitHostKind.GITLAB, apiBaseUrl: 'https://gitlab.com/api/v4', webBaseUrl: 'https://gitlab.com' }] })
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
  await expect(page.getByText('Selected scopes and their descriptions')).toBeVisible()
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
    await expect(page.getByRole('button', { name: 'Manage global Plane hosts' })).toHaveCount(0)
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

for (const provider of ['gitlab', 'linear', 'plane']) {
  test(`creates ${provider} connection with explicit sharing and a write-only key`, async ({ page }) => {
    let request: w.WorkStationServiceCreateConnectionRequest | undefined
    let saved: w.WorkStationConnection | undefined
    await setupAuthenticatedConnectRoutes(page, async (route, url) => {
      if (await hosts(route, url)) return true
      if (url.includes('WorkStationHostService/ListHosts')) return fulfillProto(route, w.WorkStationHostServiceListHostsResponseSchema, { hosts: [{ id: 'plane-cloud', name: 'Plane Cloud', apiBaseUrl: 'https://api.plane.so', webBaseUrl: 'https://app.plane.so', profile: 'plane-v2', revision: 1n }] })
      if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: saved ? [saved] : [] })
      if (url.includes('WorkStationService/CreateConnection')) {
        request = fromBinary(w.WorkStationServiceCreateConnectionRequestSchema, route.request().postDataBuffer()!)
        saved = create(w.WorkStationConnectionSchema, { id: provider, name: request.name, provider: request.provider, hostId: request.hostId, externalWorkspaceSlug: request.externalWorkspaceSlug, credentialSet: true, revision: 1n })
        return fulfillProto(route, w.WorkStationServiceCreateConnectionResponseSchema, { connection: saved })
      }
      return false
    })
    await page.goto('/work-station')
    await page.getByRole('button', { name: 'Manage connections', exact: true }).click()
    await page.getByRole('combobox', { name: 'Platform', exact: true }).selectOption(provider)
    await page.getByLabel('Connection name').fill(`${provider} planning`)
    if (provider === 'gitlab') await page.getByLabel('Git Host', { exact: true }).selectOption('gitlab')
    if (provider === 'plane') {
      await page.getByRole('combobox', { name: 'Plane host', exact: true }).selectOption('plane-cloud')
      await page.getByLabel('Plane workspace slug').fill('engineering')
    }
    await page.getByLabel('API key', { exact: true }).fill('source-private-key')
    await page.getByRole('button', { name: 'Create connection', exact: true }).click()
    await expect(page.getByLabel('Replace API key')).toHaveValue('')
    await expect(page.getByLabel('Connection name')).toHaveValue(`${provider} planning`)
    expect(request?.provider).toBe(provider)
    expect(request?.hostId).toBe(provider === 'linear' ? '' : provider === 'plane' ? 'plane-cloud' : 'gitlab')
    expect(request?.externalWorkspaceSlug).toBe(provider === 'plane' ? 'engineering' : '')
    expect(request?.apiKey).toBe('source-private-key')
    expect(saved?.scopes).toEqual([])
    await expect(page.getByRole('button', { name: provider === 'linear' ? 'Discover teams' : 'Discover projects', exact: true })).toBeVisible()
    expect(await page.evaluate(() => JSON.stringify(localStorage))).not.toContain('source-private-key')
    expect(page.url()).not.toContain('source-private-key')
  })
}

test('merge-request tab and paginated searched facets read the local projection', async ({ page }) => {
  const facetRequests: w.WorkStationServiceListFilterOptionsRequest[] = []
  let listRequest: w.WorkStationServiceListItemsRequest | undefined
  let discoveries = 0
  const gitlab = create(w.WorkStationConnectionSchema, { ...connection(), provider: 'gitlab', webBaseUrl: 'https://gitlab.example.com/group-root', scopes: [create(w.WorkStationScopeSchema, { id: '42', name: 'team/project', initialized: true, health: 'HEALTHY', coverage: 'COMPLETE', scannedItems: 24, stale: true })] })
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: [gitlab] })
    if (url.includes('WorkStationService/ListFilterOptions')) {
      const req = fromBinary(w.WorkStationServiceListFilterOptionsRequestSchema, route.request().postDataBuffer()!)
      facetRequests.push(req)
      const options = req.dimension === 'provider' ? [{ id: 'gitlab', name: 'GitLab' }] : req.dimension === 'assignee' ? req.search ? [{ id: 'u3', name: 'Morgan' }] : req.pageToken ? [{ id: 'u2', name: 'Sam' }] : [{ id: 'u1', name: 'Alex' }] : []
      return fulfillProto(route, w.WorkStationServiceListFilterOptionsResponseSchema, { options, nextPageToken: req.dimension === 'assignee' && !req.pageToken && !req.search ? 'assignee-page-2' : '' })
    }
    if (url.includes('WorkStationService/DiscoverScopes')) { discoveries++; return fulfillProto(route, w.WorkStationServiceDiscoverScopesResponseSchema, {}) }
    if (url.includes('WorkStationService/ListItems')) {
      listRequest = fromBinary(w.WorkStationServiceListItemsRequestSchema, route.request().postDataBuffer()!)
      return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, { items: [{ ...item(), provider: 'gitlab', kind: listRequest.kind, title: 'Review the merge', draft: true }] })
    }
    return false
  })
  await page.goto('/work-station')
  await page.getByRole('tab', { name: 'Merge requests' }).click()
  await expect.poll(() => listRequest?.kind).toBe('MERGE_REQUEST')
  await page.getByLabel('Provider', { exact: true }).selectOption('gitlab')
  await expect.poll(() => listRequest?.provider).toBe('gitlab')
  await page.getByText('Find more assignee options', { exact: true }).click()
  await page.getByRole('button', { name: 'Load more assignee options' }).click()
  await expect(page.getByLabel('Assignee', { exact: true }).getByRole('option', { name: 'Sam' })).toHaveCount(1)
  await page.getByLabel('Assignee', { exact: true }).selectOption('u2')
  await expect.poll(() => listRequest?.assignee).toBe('u2')
  await page.getByLabel('Search assignee options').fill('morgan')
  await expect(page.getByLabel('Assignee', { exact: true }).getByRole('option', { name: 'Morgan' })).toHaveCount(1)
  expect(facetRequests.some((request) => request.dimension === 'assignee' && request.pageToken === 'assignee-page-2')).toBe(true)
  expect(facetRequests.some((request) => request.search === 'morgan' && request.kind === 'MERGE_REQUEST' && request.provider === 'gitlab')).toBe(true)
  expect(discoveries).toBe(0)
  await page.getByText('Sync health · 1 connections').click()
  await expect(page.getByText('Stale', { exact: true })).toBeVisible()
  await expect(page.getByText('24 items scanned', { exact: true })).toBeVisible()
})

for (const approved of [true, false]) {
  test(`self-managed source link ${approved ? 'matches' : 'rejects'} the approved web-root boundary`, async ({ page }) => {
    const source = create(w.WorkStationItemSchema, { ...item(), provider: 'gitlab', url: `https://gitlab.example.com/${approved ? 'gitlab' : 'gitlab-other'}/org/repo/-/issues/1` })
    await setupAuthenticatedConnectRoutes(page, async (route, url) => {
      if (url.includes('WorkStationService/ListConnections')) return fulfillProto(route, w.WorkStationServiceListConnectionsResponseSchema, { connections: [{ ...connection(), provider: 'gitlab', webBaseUrl: 'https://gitlab.example.com/gitlab' }] })
      if (url.includes('WorkStationService/ListItems')) return fulfillProto(route, w.WorkStationServiceListItemsResponseSchema, { items: [source] })
      if (url.includes('WorkStationService/GetItem')) return fulfillProto(route, w.WorkStationServiceGetItemResponseSchema, { item: source })
      return false
    })
    await page.goto('/work-station')
    await page.getByRole('button', { name: /Fix the queue/ }).click()
    await expect(page.getByRole('link', { name: 'Open on GitLab' })).toHaveCount(approved ? 1 : 0)
  })
}

test('global Plane endpoint administration preserves revisions and needs explicit deletion', async ({ page }) => {
  let saved: w.WorkStationHost | undefined
  let deletes = 0
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (await hosts(route, url)) return true
    if (url.includes('WorkStationHostService/ListHosts')) return fulfillProto(route, w.WorkStationHostServiceListHostsResponseSchema, { hosts: saved ? [saved] : [] })
    if (url.includes('WorkStationHostService/CreateHost')) {
      const req = fromBinary(w.WorkStationHostServiceCreateHostRequestSchema, route.request().postDataBuffer()!)
      expect(req.host?.profile).toBe('plane-v2')
      saved = create(w.WorkStationHostSchema, req.host); saved.id = 'selfhost'; saved.revision = 1n
      return fulfillProto(route, w.WorkStationHostServiceCreateHostResponseSchema, { host: saved })
    }
    if (url.includes('WorkStationHostService/UpdateHost')) {
      const req = fromBinary(w.WorkStationHostServiceUpdateHostRequestSchema, route.request().postDataBuffer()!)
      expect(req.expectedRevision).toBe(1n)
      saved = create(w.WorkStationHostSchema, req.host); saved.revision = 2n
      return fulfillProto(route, w.WorkStationHostServiceUpdateHostResponseSchema, { host: saved })
    }
    if (url.includes('WorkStationHostService/DeleteHost')) {
      const req = fromBinary(w.WorkStationHostServiceDeleteHostRequestSchema, route.request().postDataBuffer()!)
      expect(req.expectedRevision).toBe(2n); deletes++; saved = undefined
      return fulfillProto(route, w.WorkStationHostServiceDeleteHostResponseSchema, {})
    }
    return false
  })
  await page.goto('/work-station')
  await expect(page.getByRole('button', { name: 'Manage global Plane hosts' })).toHaveCount(0)
  await page.goto('/admin/git-hosts')
  await page.getByRole('button', { name: 'Manage global Plane hosts' }).click()
  await expect(page.getByRole('dialog').getByText('Approved endpoints are available to every workspace.', { exact: false })).toBeVisible()
  await page.getByLabel('Host name').fill('Company Plane')
  await page.getByLabel('API base URL').fill('https://plane.example.com')
  await page.getByLabel('Web base URL').fill('https://plane.example.com')
  await page.getByRole('button', { name: 'Register host', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Save host', exact: true })).toBeVisible()
  await page.getByLabel('Host name').fill('Renamed Plane')
  await page.getByRole('button', { name: 'Save host', exact: true }).click()
  await expect.poll(() => saved?.revision).toBe(2n)
  await page.getByRole('button', { name: 'Delete host', exact: true }).click()
  expect(deletes).toBe(0)
  await page.getByRole('button', { name: 'Confirm host deletion' }).click()
  await expect.poll(() => deletes).toBe(1)
  await expect(page.getByRole('button', { name: 'Register host', exact: true })).toBeVisible()
})
