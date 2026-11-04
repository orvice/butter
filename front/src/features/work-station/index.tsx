import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { GitHostKind } from '@/gen/agents/v1/githost_pb'
import type {
  WorkStationConnection,
  WorkStationItem,
  WorkStationScope,
} from '@/gen/agents/v1/workstation_pb'
import { Route } from '@/routes/_authenticated/work-station'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { ConnectError } from '@connectrpc/connect'
import { ExternalLink, RefreshCw } from 'lucide-react'
import { useGitHosts } from '@/api/git-hosts'
import {
  workStationClient,
  workStationHostClient,
  type WorkStationClient,
  type WorkStationHostClient,
} from '@/api/work-station'
import { useWorkspaceMembers } from '@/api/workspaces'
import { useAuthStore, useIsAdmin } from '@/stores/auth-store'
import { useWorkspace } from '@/hooks/use-workspace'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { FacetFilter } from './facet-filter'

function errorText(error: unknown) {
  return error instanceof ConnectError
    ? error.rawMessage
    : 'Unable to complete the request. Try again.'
}
function when(value: WorkStationItem['sourceUpdatedAt']) {
  return value ? timestampDate(value).toLocaleString() : 'Not yet synced'
}
function providerName(provider: string) {
  return (
    (
      {
        github: 'GitHub',
        gitlab: 'GitLab',
        linear: 'Linear',
        plane: 'Plane',
      } as Record<string, string>
    )[provider] ?? provider
  )
}
function scopeName(provider: string) {
  return provider === 'linear'
    ? 'teams'
    : provider === 'github'
      ? 'repositories'
      : 'projects'
}
export function WorkStationPage() {
  const { selectedWorkspaceId } = useWorkspace()
  const user = useAuthStore((s) => s.auth.user)
  const token = useAuthStore((s) => s.auth.token)
  // Remount clears keys, candidate resources, detail and paging on identity change.
  return (
    <Station
      key={`${user?.id}:${selectedWorkspaceId}`}
      workspaceId={selectedWorkspaceId}
      token={token ?? ''}
      userId={user?.id ?? ''}
    />
  )
}
function Station({
  workspaceId,
  token,
  userId,
}: {
  workspaceId: string
  token: string
  userId: string
}) {
  const client = useMemo(
    () => workStationClient(workspaceId, token),
    [workspaceId, token]
  )
  const hostClient = useMemo(
    () => workStationHostClient(workspaceId, token),
    [workspaceId, token]
  )
  const prefix = useMemo(
    () => ['work-station', userId, workspaceId],
    [userId, workspaceId]
  )
  const qc = useQueryClient()
  const isAdmin = useIsAdmin()
  const { data: membership } = useWorkspaceMembers(isAdmin ? '' : workspaceId)
  const canManage =
    isAdmin ||
    membership?.members?.some(
      (m) => m.userId === userId && m.role === 'owner'
    ) === true
  const filters = Route.useSearch()
  const navigate = useNavigate({ from: Route.fullPath })
  const [pages, setPages] = useState([''])
  const [detailId, setDetailId] = useState('')
  const [manage, setManage] = useState(false)
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')
  const [refreshing, setRefreshing] = useState(false)
  const connections = useQuery({
    queryKey: [...prefix, 'connections'],
    queryFn: ({ signal }) => client.listConnections({}, { signal }),
    refetchInterval: 15000,
  })
  const pageToken = pages[pages.length - 1]
  const list = useQuery({
    queryKey: [...prefix, 'items', filters, pageToken],
    queryFn: ({ signal }) =>
      client.listItems(
        {
          kind: filters.kind,
          status: filters.status,
          keyword: filters.q,
          connectionId: filters.connection,
          scopeId: filters.scope,
          assignee: filters.assignee,
          provider: filters.provider,
          projectId: filters.project,
          pageSize: 50,
          pageToken,
        },
        { signal }
      ),
    refetchInterval: pages.length === 1 ? 15000 : false,
  })
  const detail = useQuery({
    queryKey: [...prefix, 'detail', detailId],
    queryFn: ({ signal }) => client.getItem({ id: detailId }, { signal }),
    enabled: !!detailId,
    refetchInterval: 15000,
    retry: false,
  })
  useEffect(
    () => () => {
      void qc.cancelQueries({ queryKey: prefix })
      qc.removeQueries({ queryKey: prefix })
    },
    [qc, prefix]
  )
  async function changed() {
    setDetailId('')
    setPages([''])
    await qc.cancelQueries({ queryKey: prefix })
    await qc.resetQueries({
      queryKey: prefix,
      predicate: (q) => q.queryKey[3] !== 'connections',
    })
    await qc.invalidateQueries({ queryKey: [...prefix, 'connections'] })
  }
  function filter(next: Partial<typeof filters>) {
    setPages([''])
    setDetailId('')
    void navigate({ search: { ...filters, ...next }, replace: true })
  }
  async function refresh() {
    setRefreshing(true)
    setError('')
    try {
      const responses = []
      for (let start = 0; start < refreshTargets.length; start += 4) {
        responses.push(
          ...(await Promise.all(
            refreshTargets
              .slice(start, start + 4)
              .map((c) => client.refresh({ id: c.id }))
          ))
        )
      }
      setNotice(
        responses.map((r) => r.status).join(', ') ||
          'Select scopes on an active connection to start syncing.'
      )
      // Accepted refresh is only an intent. Polling observes later publication.
      setPages([''])
      await qc.invalidateQueries({ queryKey: prefix })
    } catch (e) {
      setError(errorText(e))
    } finally {
      setRefreshing(false)
    }
  }
  const allConnections = connections.data?.connections ?? []
  const refreshTargets = allConnections.filter(
    (c) =>
      c.scopes.length > 0 &&
      !c.paused &&
      (!filters.provider || c.provider === filters.provider) &&
      (!filters.connection || c.id === filters.connection)
  )
  const hasScopes = allConnections.some((c) => c.scopes.length > 0)
  const initializing =
    hasScopes &&
    allConnections.some(
      (c) =>
        !c.paused &&
        c.scopes.some((s) => !s.initialized && !s.health.includes('DENIED'))
    )
  return (
    <>
      <Header />
      <Main className='space-y-6'>
        <div className='flex flex-wrap items-start justify-between gap-4'>
          <div>
            <h1 className='text-2xl font-semibold tracking-tight'>
              Work Station
            </h1>
            <p className='mt-1 max-w-2xl text-sm text-muted-foreground'>
              Read-only work shared with everyone in this workspace. Open items
              and items ended in the last 90 days are retained.
            </p>
          </div>
          <div className='flex flex-wrap gap-3'>
            <Button
              variant='outline'
              onClick={() => void refresh()}
              disabled={refreshing || !refreshTargets.length}
            >
              <RefreshCw className='size-4' aria-hidden />
              Request refresh
            </Button>
            {canManage && (
              <Button onClick={() => setManage(true)}>
                Manage connections
              </Button>
            )}
          </div>
        </div>
        <p className='text-sm text-muted-foreground'>
          Browse Linear and Plane issues, GitHub issues and pull requests, and
          GitLab issues and merge requests. Refresh applies to active
          connections matching your provider and connection filters.
        </p>
        <div role='status' className='text-sm'>
          {notice}
        </div>
        {(error || connections.error || list.error) && (
          <p role='alert' className='text-destructive'>
            {error || errorText(connections.error ?? list.error)}
          </p>
        )}
        {allConnections.length > 0 && (
          <details className='rounded-lg border p-4'>
            <summary className='cursor-pointer font-medium'>
              Sync health · {allConnections.length} connections
            </summary>
            <ul className='mt-4 space-y-4'>
              {allConnections.map((c) => (
                <li key={c.id}>
                  <p className='font-medium'>
                    {c.name}{' '}
                    {c.paused && <Badge variant='destructive'>Paused</Badge>}
                  </p>
                  {!c.scopes.length && (
                    <p className='text-sm text-muted-foreground'>
                      No scopes selected
                    </p>
                  )}
                  <ul className='mt-1 space-y-1'>
                    {c.scopes.map((s) => (
                      <li
                        key={s.id}
                        className='flex flex-wrap gap-x-3 text-sm text-muted-foreground'
                      >
                        <span>{s.name}</span>
                        {s.stale && <Badge variant='outline'>Stale</Badge>}
                        <span>
                          {s.health || 'QUEUED'} ·{' '}
                          {s.coverage || 'INITIALIZING'}
                        </span>
                        <span>Last success: {when(s.lastSuccessAt)}</span>
                        <span>Next attempt: {when(s.nextAttemptAt)}</span>
                        {s.scanStartedAt && (
                          <span>Scan started: {when(s.scanStartedAt)}</span>
                        )}
                        <span className='tabular-nums'>
                          {s.scannedItems} items scanned
                        </span>
                        <span>
                          Last full reconciliation:{' '}
                          {when(s.lastFullReconciliationAt)}
                        </span>
                        {s.unresolvedTerminalCount > 0 && (
                          <span>
                            {s.unresolvedTerminalCount} ended items need a
                            verified end time
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                </li>
              ))}
            </ul>
          </details>
        )}
        <Tabs
          value={filters.kind}
          onValueChange={(kind) => filter({ kind, status: 'open' })}
        >
          <TabsList aria-label='Work item type' className='h-auto flex-wrap'>
            <TabsTrigger value='ISSUE'>Issues</TabsTrigger>
            <TabsTrigger value='PULL_REQUEST'>Pull requests</TabsTrigger>
            <TabsTrigger value='MERGE_REQUEST'>Merge requests</TabsTrigger>
          </TabsList>
        </Tabs>
        <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
          <label className='space-y-1 text-sm'>
            Connection
            <select
              aria-label='Connection'
              className='block h-10 w-full rounded-md border bg-background px-3'
              value={filters.connection}
              onChange={(e) =>
                filter({ connection: e.target.value, scope: '', project: '' })
              }
            >
              <option value=''>All connections</option>
              {allConnections.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <FacetFilter
            client={client}
            prefix={prefix}
            filters={filters}
            dimension='provider'
            label='Provider'
            value={filters.provider}
            onChange={(provider) =>
              filter({ provider, connection: '', scope: '', project: '' })
            }
          />
          <FacetFilter
            client={client}
            prefix={prefix}
            filters={filters}
            dimension='scope'
            label='Scope'
            value={filters.scope}
            onChange={(scope) => filter({ scope, project: '' })}
          />
          <FacetFilter
            client={client}
            prefix={prefix}
            filters={filters}
            dimension='project'
            label='Project'
            value={filters.project}
            onChange={(project) => filter({ project })}
          />
          <FacetFilter
            client={client}
            prefix={prefix}
            filters={filters}
            dimension='status'
            label='Status'
            value={filters.status}
            onChange={(status) => filter({ status })}
            defaults={[
              { id: 'open', name: 'Open' },
              { id: 'all', name: 'All retained items' },
            ]}
          />
          <label className='space-y-1 text-sm'>
            Search title or identifier
            <Input
              value={filters.q}
              maxLength={256}
              onChange={(e) => filter({ q: e.target.value })}
            />
          </label>
          <FacetFilter
            client={client}
            prefix={prefix}
            filters={filters}
            dimension='assignee'
            label='Assignee'
            value={filters.assignee}
            onChange={(assignee) => filter({ assignee })}
          />
          <Button
            variant='ghost'
            className='self-end'
            onClick={() =>
              filter({
                connection: '',
                scope: '',
                provider: '',
                project: '',
                status: 'open',
                assignee: '',
                q: '',
              })
            }
          >
            Clear filters
          </Button>
        </div>
        {list.isPending || connections.isPending ? (
          <p role='status'>Loading work items…</p>
        ) : list.data?.items.length ? (
          <ul className='divide-y rounded-lg border'>
            {list.data.items.map((item) => (
              <li key={item.id} className='p-4'>
                <button
                  className='w-full space-y-2 text-start focus-visible:rounded focus-visible:outline-2 focus-visible:outline-offset-4'
                  onClick={() => setDetailId(item.id)}
                >
                  <span className='flex flex-wrap items-center gap-2'>
                    <span className='text-sm text-muted-foreground'>
                      {item.scopeName} {item.identifier}
                    </span>
                    <Badge variant='secondary'>
                      {item.status.toLowerCase().replace(/_/g, ' ')}
                    </Badge>
                    {item.draft && <Badge variant='outline'>Draft</Badge>}
                  </span>
                  <span className='block font-medium break-words'>
                    {item.title}
                  </span>
                  <span className='block text-sm text-muted-foreground'>
                    {when(item.sourceUpdatedAt)} · {item.sources.length} sources
                  </span>
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <div className='rounded-lg border border-dashed p-8 text-center'>
            <h2 className='font-medium'>
              {!allConnections.length
                ? 'No connections yet'
                : !hasScopes
                  ? 'Select scopes to import'
                  : allConnections.every((c) => c.paused)
                    ? 'Connections need attention'
                    : initializing
                      ? 'Initial sync is pending'
                      : 'No items match these filters'}
            </h2>
            <p className='mt-2 text-sm text-muted-foreground'>
              {!allConnections.length || !hasScopes
                ? canManage
                  ? 'Manage connections to choose what this workspace can read.'
                  : 'Ask a workspace owner to select the teams, projects, or repositories to share.'
                : initializing
                  ? 'Published items will appear after a complete scan.'
                  : 'Check sync health or clear filters to see more items.'}
            </p>
          </div>
        )}
        <div className='flex gap-3'>
          <Button
            variant='outline'
            disabled={pages.length === 1}
            onClick={() => setPages((p) => p.slice(0, -1))}
          >
            Previous page
          </Button>
          <Button
            variant='outline'
            disabled={!list.data?.nextPageToken}
            onClick={() => {
              setDetailId('')
              setPages((p) => [...p, list.data!.nextPageToken])
            }}
          >
            Next page
          </Button>
        </div>
        <Dialog
          open={!!detailId}
          onOpenChange={(open) => {
            if (!open) setDetailId('')
          }}
        >
          <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
            <DialogHeader>
              <DialogTitle>
                {detail.error
                  ? 'Item unavailable'
                  : (detail.data?.item?.title ?? 'Loading item…')}
              </DialogTitle>
              <DialogDescription>Read-only source details</DialogDescription>
            </DialogHeader>
            {detail.error ? (
              <p role='alert'>
                This item is no longer available. Close this dialog and refresh
                the list.
              </p>
            ) : (
              detail.data?.item && (
                <ItemDetail
                  item={detail.data.item}
                  connections={allConnections}
                />
              )
            )}
          </DialogContent>
        </Dialog>
        {canManage && (
          <Dialog open={manage} onOpenChange={setManage}>
            <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
              <DialogHeader>
                <DialogTitle>Manage connections</DialogTitle>
                <DialogDescription>
                  Selected scopes and their descriptions will be visible to
                  every member of this Butter workspace.
                </DialogDescription>
              </DialogHeader>
              <ConnectionManager
                client={client}
                hostClient={hostClient}
                prefix={prefix}
                connections={allConnections}
                changed={changed}
              />
            </DialogContent>
          </Dialog>
        )}
      </Main>
    </>
  )
}
function ItemDetail({
  item,
  connections,
}: {
  item: WorkStationItem
  connections: WorkStationConnection[]
}) {
  // Descriptions are plain text: no HTML, scripts, images or private attachments.
  const approvedRoots = connections
    .filter(
      (connection) =>
        connection.provider === item.provider &&
        item.sources.some((source) => source.connectionId === connection.id)
    )
    .map(
      (connection) =>
        connection.webBaseUrl ||
        (connection.provider === 'github' ? 'https://github.com' : '')
    )
  const safeSource = approvedRoots.some((root) => {
    try {
      if (!root || item.url.includes('\\')) return false
      const link = new URL(item.url)
      const approved = new URL(root)
      const path = approved.pathname.replace(/\/+$/, '')
      return (
        link.protocol === 'https:' &&
        !link.username &&
        !link.password &&
        link.origin === approved.origin &&
        (link.pathname === path || link.pathname.startsWith(`${path}/`))
      )
    } catch {
      return false
    }
  })
  return (
    <div className='space-y-4'>
      <p className='text-sm'>
        {item.scopeName} {item.identifier} · {item.rawState}
        {item.draft ? ' · Draft' : ''}
      </p>
      <dl className='grid gap-2 text-sm'>
        <div>
          <dt className='text-muted-foreground'>Assignees</dt>
          <dd>{item.assignees.join(', ') || 'Unassigned'}</dd>
        </div>
        <div>
          <dt className='text-muted-foreground'>Labels</dt>
          <dd>{item.labels.join(', ') || 'None'}</dd>
        </div>
        <div>
          <dt className='text-muted-foreground'>Updated</dt>
          <dd>{when(item.sourceUpdatedAt)}</dd>
        </div>
        {item.endedAt && (
          <div>
            <dt className='text-muted-foreground'>Ended</dt>
            <dd>{when(item.endedAt)}</dd>
          </div>
        )}
        <div>
          <dt className='text-muted-foreground'>Sources</dt>
          <dd>
            {item.sources
              .map(
                (s) =>
                  connections.find((c) => c.id === s.connectionId)?.name ??
                  s.connectionId
              )
              .join(', ')}
          </dd>
        </div>
      </dl>
      <p className='max-w-prose text-sm leading-relaxed break-words whitespace-pre-wrap'>
        {!item.descriptionAvailable
          ? 'Description is unavailable from the source.'
          : item.description || 'No description.'}
      </p>
      {safeSource && (
        <a
          className='inline-flex items-center gap-2 text-sm underline underline-offset-4'
          href={item.url}
          target='_blank'
          rel='noopener noreferrer'
        >
          Open on {providerName(item.provider)}
          <ExternalLink aria-hidden className='size-4' />
        </a>
      )}
    </div>
  )
}
function ConnectionManager({
  client,
  hostClient,
  prefix,
  connections,
  changed,
}: {
  client: WorkStationClient
  hostClient: WorkStationHostClient
  prefix: string[]
  connections: WorkStationConnection[]
  changed: () => Promise<void>
}) {
  const [id, setId] = useState('')
  const connection = connections.find((c) => c.id === id)
  return (
    <div className='space-y-5'>
      <label className='space-y-1 text-sm'>
        Connection to manage
        <select
          className='block h-10 w-full rounded-md border bg-background px-3'
          value={id}
          onChange={(e) => setId(e.target.value)}
        >
          <option value=''>Create a connection</option>
          {connections.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
      </label>
      <ConnectionForm
        key={connection?.id ?? 'new'}
        client={client}
        hostClient={hostClient}
        prefix={prefix}
        connection={connection}
        changed={changed}
        select={setId}
      />
    </div>
  )
}
function ConnectionForm({
  client,
  hostClient,
  prefix,
  connection,
  changed,
  select,
}: {
  client: WorkStationClient
  hostClient: WorkStationHostClient
  prefix: string[]
  connection?: WorkStationConnection
  changed: () => Promise<void>
  select: (id: string) => void
}) {
  const [name, setName] = useState(connection?.name ?? '')
  const [key, setKey] = useState('')
  const [hostId, setHostId] = useState(connection?.hostId ?? '')
  const [provider, setProvider] = useState(connection?.provider ?? 'github')
  const [workspaceSlug, setWorkspaceSlug] = useState(
    connection?.externalWorkspaceSlug ?? ''
  )
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [pending, setPending] = useState(false)
  const [candidates, setCandidates] = useState<WorkStationScope[]>(
    connection?.scopes ?? []
  )
  const [selected, setSelected] = useState(
    connection?.scopes.map((s) => s.id) ?? []
  )
  const [nextPage, setNextPage] = useState('')
  const [confirmDelete, setConfirmDelete] = useState(false)
  const hosts = useGitHosts()
  const planeHosts = useQuery({
    queryKey: [...prefix, 'plane-hosts'],
    queryFn: ({ signal }) => hostClient.listHosts({}, { signal }),
    enabled: provider === 'plane',
  })
  const scopeLabel = scopeName(provider)
  async function run(action: () => Promise<void>) {
    setPending(true)
    setError('')
    setNotice('')
    try {
      await action()
    } catch (e) {
      setError(errorText(e))
    } finally {
      setPending(false)
      setKey('')
    }
  }
  function save(e: FormEvent) {
    e.preventDefault()
    void run(async () => {
      const res = connection
        ? await client.updateConnection({
            id: connection.id,
            expectedRevision: connection.revision,
            name,
            apiKey: key || undefined,
          })
        : await client.createConnection({
            name,
            hostId: provider === 'linear' ? '' : hostId,
            provider,
            externalWorkspaceSlug: workspaceSlug,
            apiKey: key,
          })
      setKey('')
      await changed()
      select(res.connection?.id ?? '')
    })
  }
  return (
    <div className='space-y-5'>
      <form className='space-y-4' onSubmit={save}>
        {!connection && (
          <label className='block space-y-1 text-sm'>
            <span>Platform</span>
            <select
              className='block min-h-10 w-full rounded-md border bg-background px-3'
              value={provider}
              onChange={(e) => {
                setProvider(e.target.value)
                setHostId('')
                setWorkspaceSlug('')
                setKey('')
              }}
            >
              <option value='github'>GitHub</option>
              <option value='gitlab'>GitLab</option>
              <option value='linear'>Linear</option>
              <option value='plane'>Plane</option>
            </select>
          </label>
        )}
        {connection && (
          <p className='text-sm text-muted-foreground'>
            {providerName(provider)} · {connection.instanceUrl}
            {connection.externalWorkspaceSlug
              ? ` · ${connection.externalWorkspaceSlug}`
              : ''}
          </p>
        )}
        <div className='space-y-1'>
          <Label htmlFor='station-name'>Connection name</Label>
          <Input
            id='station-name'
            value={name}
            maxLength={128}
            required
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        {!connection && (provider === 'github' || provider === 'gitlab') && (
          <div className='block space-y-1 text-sm'>
            <Label htmlFor='station-host'>Git Host</Label>
            <select
              id='station-host'
              aria-describedby='station-host-help'
              className='h-10 w-full rounded-md border bg-background px-3'
              required
              value={hostId}
              onChange={(e) => setHostId(e.target.value)}
            >
              <option value=''>Select a {providerName(provider)} host</option>
              {hosts.data?.hosts
                ?.filter((h) =>
                  provider === 'gitlab'
                    ? h.kind === GitHostKind.GITLAB
                    : h.kind === GitHostKind.GITHUB &&
                      h.apiBaseUrl.replace(/\/+$/, '') ===
                        'https://api.github.com' &&
                      h.webBaseUrl.replace(/\/+$/, '') === 'https://github.com'
                )
                .map((h) => (
                  <option key={h.id} value={h.id}>
                    {h.name}
                  </option>
                ))}
            </select>
            <p id='station-host-help' className='block text-muted-foreground'>
              A global administrator manages approved endpoints in Admin → Git
              hosts.
            </p>
          </div>
        )}
        {!connection && provider === 'linear' && (
          <p className='text-sm text-muted-foreground'>
            Connect to the official Linear service at api.linear.app. Team
            selection is required after creation.
          </p>
        )}
        {!connection && provider === 'plane' && (
          <>
            <label className='block space-y-1 text-sm'>
              <span>Plane host</span>
              <select
                required
                className='block min-h-10 w-full rounded-md border bg-background px-3'
                value={hostId}
                onChange={(e) => setHostId(e.target.value)}
              >
                <option value=''>Select an approved Plane host</option>
                {planeHosts.data?.hosts.map((host) => (
                  <option key={host.id} value={host.id}>
                    {host.name}
                  </option>
                ))}
              </select>
            </label>
            <label className='block space-y-1 text-sm'>
              <span>Plane workspace slug</span>
              <Input
                required
                value={workspaceSlug}
                maxLength={128}
                placeholder='engineering'
                onChange={(e) => setWorkspaceSlug(e.target.value)}
              />
            </label>
            <p className='text-sm text-muted-foreground'>
              Plane Cloud and self-managed hosts use the approved v2 profile.
              API keys inherit the owner's source permissions; selected projects
              and descriptions are shared with this workspace.
            </p>
            {planeHosts.error && (
              <p role='alert' className='text-sm text-destructive'>
                Unable to load approved Plane hosts.
              </p>
            )}
          </>
        )}
        <div className='space-y-1'>
          <Label htmlFor='station-key'>
            {connection ? 'Replace API key' : 'API key'}
          </Label>
          <Input
            id='station-key'
            type='password'
            autoComplete='new-password'
            required={!connection}
            value={key}
            maxLength={16384}
            onChange={(e) => setKey(e.target.value)}
          />
          <p className='text-sm text-muted-foreground'>
            {connection?.credentialSet
              ? 'API key set. Leave blank to keep it.'
              : provider === 'github'
                ? 'Use a fine-grained token with Issues and Pull requests read permissions.'
                : provider === 'gitlab'
                  ? 'Use a personal, project, or group access token with read_api. Deploy and job tokens are unsupported.'
                  : provider === 'linear'
                    ? 'Use a read-capable Linear API key with access to the teams you want to share.'
                    : 'Use a Plane API key with access to the projects you want to share.'}
          </p>
        </div>
        <Button disabled={pending} type='submit'>
          {connection ? 'Save connection' : 'Create connection'}
        </Button>
      </form>
      {connection && (
        <>
          <div className='flex flex-wrap gap-3'>
            <Button
              disabled={pending || !connection.credentialSet}
              variant='outline'
              onClick={() =>
                void run(async () => {
                  const res = await client.testConnection({ id: connection.id })
                  await changed()
                  setNotice(
                    res.connection?.paused
                      ? 'The source rejected the key. Replace it to resume.'
                      : `Connection verified. Selected ${scopeLabel} are queued for a new scan.`
                  )
                })
              }
            >
              Test connection
            </Button>
            <Button
              disabled={pending || !connection.credentialSet}
              variant='outline'
              onClick={() =>
                void run(async () => {
                  await client.updateConnection({
                    id: connection.id,
                    expectedRevision: connection.revision,
                    name,
                    apiKey: '',
                  })
                  await changed()
                })
              }
            >
              Clear API key
            </Button>
          </div>
          <fieldset className='space-y-3' disabled={pending}>
            <legend className='font-medium'>Selected {scopeLabel}</legend>
            <p className='text-sm text-muted-foreground'>
              Select only {scopeLabel} you intend to share. Clearing every
              selection imports nothing. Saving rechecks access and retries
              unavailable scopes without resetting healthy ones.
            </p>
            <Button
              type='button'
              variant='outline'
              disabled={connection.paused || !connection.credentialSet}
              onClick={() =>
                void run(async () => {
                  const page = await client.discoverScopes({
                    id: connection.id,
                    pageToken: nextPage,
                  })
                  setCandidates((old) =>
                    Array.from(
                      new Map(
                        [...old, ...page.scopes].map((s) => [s.id, s])
                      ).values()
                    )
                  )
                  setNextPage(page.nextPageToken)
                })
              }
            >
              {nextPage ? `Load more ${scopeLabel}` : `Discover ${scopeLabel}`}
            </Button>
            <div className='max-h-64 space-y-2 overflow-y-auto'>
              {candidates.map((s) => (
                <label
                  key={s.id}
                  className='flex min-h-10 items-center gap-3 text-sm'
                >
                  <input
                    type='checkbox'
                    checked={selected.includes(s.id)}
                    onChange={(e) =>
                      setSelected((old) =>
                        e.target.checked
                          ? [...old, s.id]
                          : old.filter((id) => id !== s.id)
                      )
                    }
                  />
                  {s.name}
                </label>
              ))}
            </div>
            <Button
              onClick={() =>
                void run(async () => {
                  await client.setSelectedScopes({
                    id: connection.id,
                    expectedRevision: connection.revision,
                    scopeIds: selected,
                  })
                  await changed()
                })
              }
            >
              Save selected {scopeLabel}
            </Button>
          </fieldset>
          {!confirmDelete ? (
            <Button
              variant='destructive'
              disabled={pending}
              onClick={() => setConfirmDelete(true)}
            >
              Delete connection
            </Button>
          ) : (
            <div className='space-y-3 rounded-md border p-4'>
              <p className='text-sm'>
                Delete this connection and its saved key? Its imported sources
                will disappear immediately.
              </p>
              <div className='flex gap-3'>
                <Button
                  variant='destructive'
                  disabled={pending}
                  onClick={() =>
                    void run(async () => {
                      await client.deleteConnection({
                        id: connection.id,
                        expectedRevision: connection.revision,
                      })
                      await changed()
                      select('')
                    })
                  }
                >
                  Confirm deletion
                </Button>
                <Button
                  variant='outline'
                  onClick={() => setConfirmDelete(false)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}
        </>
      )}
      <p role='status' className='text-sm'>
        {notice}
      </p>
      {error && (
        <p role='alert' className='text-sm text-destructive'>
          {error}
        </p>
      )}
    </div>
  )
}
