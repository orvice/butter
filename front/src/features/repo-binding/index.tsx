import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { timestampDate, type Timestamp } from '@bufbuild/protobuf/wkt'
import {
  Check,
  ChevronRight,
  FileText,
  Folder,
  GitBranch,
  Info,
  KeyRound,
  RefreshCw,
  ShieldCheck,
  Trash2,
  X,
} from 'lucide-react'
import {
  useDeleteRepoBinding,
  usePutRepoBinding,
  useRepoBinding,
  useRepositoryEntries,
  useRepositoryFile,
  useSetRepoBindingCredential,
  useSyncRepository,
  useValidateRepoBinding,
} from '@/api/repo-binding'
import { Code, ConnectError } from '@/api/transport'
import { useGitHosts } from '@/api/git-hosts'
import {
  RepoBindingConnectionState,
  RepoBindingWriteMode,
  RepoCacheEntryKind,
  type RepoCacheEntry,
  type RepoBindingCheck,
  type RepoBindingOverlap,
  type WorkspaceRepoBinding,
} from '@/gen/agents/v1/repobinding_pb'
import { DeleteDialog } from '@/components/delete-dialog'
import { EmptyState } from '@/components/empty-state'
import { MarkdownContent } from '@/components/markdown-content'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const schema = z.object({
  git_host_id: z.string().min(1, 'Git host is required'),
  repository: z.string().min(1, 'Repository is required'),
  branch: z.string().min(1, 'Branch is required'),
  root_path: z.string().optional(),
  write_mode: z.enum(['direct_commit', 'change_request']),
})

type FormValues = z.infer<typeof schema>

function writeModeToForm(mode: RepoBindingWriteMode): FormValues['write_mode'] {
  return mode === RepoBindingWriteMode.CHANGE_REQUEST
    ? 'change_request'
    : 'direct_commit'
}

function writeModeToProto(mode: FormValues['write_mode']): RepoBindingWriteMode {
  return mode === 'change_request'
    ? RepoBindingWriteMode.CHANGE_REQUEST
    : RepoBindingWriteMode.DIRECT_COMMIT
}

function formatTs(ts: Timestamp | undefined): string {
  if (!ts) return '-'
  return timestampDate(ts).toLocaleString()
}

const STATE_PALETTE: Record<
  RepoBindingConnectionState,
  { cls: string; label: string }
> = {
  [RepoBindingConnectionState.UNSPECIFIED]: {
    cls: 'bg-muted text-muted-foreground',
    label: 'Unvalidated',
  },
  [RepoBindingConnectionState.UNVALIDATED]: {
    cls: 'bg-muted text-muted-foreground',
    label: 'Unvalidated',
  },
  [RepoBindingConnectionState.OK]: {
    cls: 'bg-success-muted text-success-foreground',
    label: 'Connected',
  },
  [RepoBindingConnectionState.FAILED]: {
    cls: 'bg-danger-muted text-danger-foreground',
    label: 'Failed',
  },
}

function ConnectionStateBadge({ state }: { state: RepoBindingConnectionState }) {
  const p = STATE_PALETTE[state] ?? STATE_PALETTE[RepoBindingConnectionState.UNSPECIFIED]
  return (
    <Badge className={p.cls}>
      <span className='h-1.5 w-1.5 rounded-full bg-current' />
      {p.label}
    </Badge>
  )
}

export function RepoBindingPage() {
  const { data, isLoading } = useRepoBinding()
  const binding = data?.binding
  const overlaps = data?.overlaps ?? []

  return (
    <Page>
      <PageHeader
        title='Repository Binding'
        subtitle='Bind this workspace to a Git repository so agent content is read from and written to source control.'
      />
      <PageScroll>
        {isLoading ? (
          <Card>
            <CardContent className='py-10 text-center text-sm text-muted-foreground'>
              Loading…
            </CardContent>
          </Card>
        ) : (
          <div className='max-w-6xl space-y-6'>
            <div className='grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]'>
              <div className='space-y-6'>
                {!binding && (
                  <Alert>
                    <Info className='h-4 w-4' />
                    <AlertTitle>No repository bound yet</AlertTitle>
                    <AlertDescription>
                      Configure a binding below, then set a personal access token
                      and validate the connection.
                    </AlertDescription>
                  </Alert>
                )}
                {overlaps.length > 0 && <OverlapsAlert overlaps={overlaps} />}
                <BindingForm binding={binding} />
              </div>
              {binding && (
                <div className='space-y-6'>
                  <ConnectionCard binding={binding} />
                  <CredentialCard binding={binding} />
                  <DangerZoneCard binding={binding} />
                </div>
              )}
            </div>
            {binding && <RepositoryBrowser binding={binding} />}
          </div>
        )}
      </PageScroll>
    </Page>
  )
}

function RepositoryBrowser({ binding }: { binding: WorkspaceRepoBinding }) {
  const [directoryPath, setDirectoryPath] = useState('')
  const [selectedFile, setSelectedFile] = useState('')
  const entriesQuery = useRepositoryEntries(directoryPath)
  const fileQuery = useRepositoryFile(selectedFile)
  const syncMutation = useSyncRepository()
  const entries = useMemo(
    () =>
      [...(entriesQuery.data?.entries ?? [])].sort((a, b) => {
        if (a.kind !== b.kind) {
          if (a.kind === RepoCacheEntryKind.DIRECTORY) return -1
          if (b.kind === RepoCacheEntryKind.DIRECTORY) return 1
        }
        return a.path.localeCompare(b.path)
      }),
    [entriesQuery.data?.entries]
  )
  const cacheMissing =
    entriesQuery.error instanceof ConnectError &&
    entriesQuery.error.code === Code.NotFound

  function handleSync() {
    syncMutation.mutate(undefined, {
      onSuccess: () => toast.success('Repository cache refreshed'),
      onError: (error) => toast.error(error.message),
    })
  }

  function openEntry(entry: RepoCacheEntry) {
    if (entry.kind === RepoCacheEntryKind.DIRECTORY) {
      setDirectoryPath(entry.path)
      setSelectedFile('')
      return
    }
    if (entry.kind === RepoCacheEntryKind.FILE) {
      setSelectedFile(entry.path)
    }
  }

  function navigateTo(path: string) {
    setDirectoryPath(path)
    setSelectedFile('')
  }

  const crumbs = directoryPath ? directoryPath.split('/') : []
  const cachedSHA = entriesQuery.data?.commitSha ?? ''

  return (
    <Card>
      <CardHeader className='gap-4 sm:grid-cols-[minmax(0,1fr)_auto]'>
        <div className='min-w-0 space-y-1.5'>
          <CardTitle className='flex items-center gap-2 text-base'>
            <Folder className='h-4 w-4' />
            Cached content
          </CardTitle>
          <div className='flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground'>
            <RevisionLabel label='Cached' sha={cachedSHA} empty='Not synced' />
            <RevisionLabel
              label='Observed'
              sha={binding.observedCommitSha}
              empty='Not observed'
            />
            <RevisionLabel
              label='Active'
              sha={binding.activeCommitSha}
              empty='Not published'
            />
            <span>Synced {formatTs(binding.lastSyncedAt)}</span>
          </div>
        </div>
        <Button
          type='button'
          variant='outline'
          onClick={handleSync}
          disabled={syncMutation.isPending}
        >
          <RefreshCw
            className={syncMutation.isPending ? 'animate-spin' : undefined}
            aria-hidden='true'
          />
          {syncMutation.isPending ? 'Syncing…' : 'Refresh cache'}
        </Button>
        <p className='sr-only' role='status' aria-live='polite'>
          {syncMutation.isPending ? 'Repository synchronization in progress' : ''}
        </p>
      </CardHeader>
      <CardContent>
        <div className='overflow-hidden rounded-md border lg:grid lg:min-h-[30rem] lg:grid-cols-[minmax(15rem,20rem)_minmax(0,1fr)]'>
          <section
            className='min-w-0 border-b lg:border-e lg:border-b-0'
            aria-labelledby='repository-directory-heading'
          >
            <div className='border-b px-4 py-3'>
              <h3 id='repository-directory-heading' className='sr-only'>
                Repository directory
              </h3>
              <nav aria-label='Repository path'>
                <ol className='flex min-w-0 flex-wrap items-center gap-1 text-sm'>
                  <li>
                    <Button
                      type='button'
                      variant='ghost'
                      size='sm'
                      className='h-7 px-2 font-mono text-xs'
                      onClick={() => navigateTo('')}
                    >
                      Root
                    </Button>
                  </li>
                  {crumbs.map((crumb, index) => {
                    const path = crumbs.slice(0, index + 1).join('/')
                    return (
                      <li key={path} className='flex min-w-0 items-center gap-1'>
                        <ChevronRight
                          className='h-3.5 w-3.5 shrink-0 text-muted-foreground'
                          aria-hidden='true'
                        />
                        <Button
                          type='button'
                          variant='ghost'
                          size='sm'
                          className='h-7 min-w-0 px-2 font-mono text-xs'
                          onClick={() => navigateTo(path)}
                          aria-current={index === crumbs.length - 1 ? 'page' : undefined}
                        >
                          <span className='truncate'>{crumb}</span>
                        </Button>
                      </li>
                    )
                  })}
                </ol>
              </nav>
            </div>
            <div className='max-h-[22rem] overflow-y-auto p-2 lg:max-h-[26rem]'>
              {entriesQuery.isLoading ? (
                <div className='space-y-2 p-2'>
                  {Array.from({ length: 6 }).map((_, index) => (
                    <Skeleton key={index} className='h-10 w-full' />
                  ))}
                </div>
              ) : cacheMissing ? (
                <EmptyState
                  className='m-2 border-0 bg-transparent p-6 shadow-none'
                  title='No cached content'
                  icon={<Folder className='h-6 w-6' />}
                  action={
                    <Button
                      type='button'
                      size='sm'
                      onClick={handleSync}
                      disabled={syncMutation.isPending}
                    >
                      <RefreshCw
                        className={syncMutation.isPending ? 'animate-spin' : undefined}
                        aria-hidden='true'
                      />
                      Refresh cache
                    </Button>
                  }
                />
              ) : entriesQuery.error ? (
                <p className='p-4 text-sm text-destructive' role='alert'>
                  {entriesQuery.error.message}
                </p>
              ) : entries.length === 0 ? (
                <p className='p-4 text-sm text-muted-foreground'>
                  This directory is empty.
                </p>
              ) : (
                <ul className='space-y-1'>
                  {entries.map((entry) => (
                    <RepositoryEntryRow
                      key={entry.path}
                      entry={entry}
                      selected={entry.path === selectedFile}
                      onOpen={() => openEntry(entry)}
                    />
                  ))}
                </ul>
              )}
            </div>
          </section>
          <section
            className='min-w-0 p-4 sm:p-6'
            aria-labelledby='repository-preview-heading'
          >
            <h3 id='repository-preview-heading' className='sr-only'>
              Markdown preview
            </h3>
            {!selectedFile ? (
              <EmptyState
                className='min-h-[18rem] border-0 bg-transparent shadow-none lg:min-h-full'
                title='No file selected'
                icon={<FileText className='h-6 w-6' />}
              />
            ) : fileQuery.isLoading ? (
              <div className='space-y-3'>
                <Skeleton className='h-6 w-2/3' />
                <Skeleton className='h-4 w-full' />
                <Skeleton className='h-4 w-5/6' />
                <Skeleton className='h-28 w-full' />
              </div>
            ) : fileQuery.error ? (
              <div className='space-y-2' role='alert'>
                <h4 className='break-all font-mono text-sm font-medium'>
                  {selectedFile}
                </h4>
                <p className='text-sm text-destructive'>
                  {fileQuery.error.message}
                </p>
              </div>
            ) : (
              <div className='min-w-0 space-y-5'>
                <div className='flex flex-wrap items-center justify-between gap-2'>
                  <h4 className='min-w-0 break-all font-mono text-sm font-medium'>
                    {selectedFile}
                  </h4>
                  <span className='shrink-0 text-xs text-muted-foreground'>
                    {formatBytes(fileQuery.data?.entry?.size)}
                  </span>
                </div>
                <MarkdownContent content={fileQuery.data?.content ?? ''} />
              </div>
            )}
          </section>
        </div>
      </CardContent>
    </Card>
  )
}

function RepositoryEntryRow({
  entry,
  selected,
  onOpen,
}: {
  entry: RepoCacheEntry
  selected: boolean
  onOpen: () => void
}) {
  const isDirectory = entry.kind === RepoCacheEntryKind.DIRECTORY
  const parentPath = entry.path.split('/').slice(0, -1).join('/')
  const showClaimState = isDirectory && parentPath === 'agents'
  const pathParts = entry.path.split('/')
  const name = pathParts[pathParts.length - 1] ?? entry.path

  return (
    <li>
      <button
        type='button'
        className={`flex min-h-10 w-full items-center gap-3 rounded-md px-3 py-2 text-start text-sm transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 ${
          selected ? 'bg-accent text-accent-foreground' : ''
        }`}
        onClick={onOpen}
        aria-current={selected ? 'true' : undefined}
      >
        {isDirectory ? (
          <Folder className='h-4 w-4 shrink-0 text-muted-foreground' aria-hidden='true' />
        ) : (
          <FileText className='h-4 w-4 shrink-0 text-muted-foreground' aria-hidden='true' />
        )}
        <span className='min-w-0 flex-1 truncate font-mono text-xs'>{name}</span>
        {showClaimState ? (
          <Badge variant={entry.claimed ? 'secondary' : 'outline'}>
            {entry.claimed ? 'Claimed' : 'Unclaimed'}
          </Badge>
        ) : !isDirectory ? (
          <span className='shrink-0 text-xs tabular-nums text-muted-foreground'>
            {formatBytes(entry.size)}
          </span>
        ) : null}
      </button>
    </li>
  )
}

function RevisionLabel({
  label,
  sha,
  empty,
}: {
  label: string
  sha: string
  empty: string
}) {
  return (
    <span>
      {label}:{' '}
      {sha ? (
        <code className='font-mono text-foreground' title={sha}>
          {sha.slice(0, 8)}
        </code>
      ) : (
        empty
      )}
    </span>
  )
}

function formatBytes(value: bigint | undefined): string {
  const bytes = Number(value ?? 0n)
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MiB`
}

function BindingForm({ binding }: { binding?: WorkspaceRepoBinding }) {
  const { data: hostsData, isLoading: hostsLoading } = useGitHosts()
  const putMutation = usePutRepoBinding()
  const hosts = hostsData?.hosts ?? []

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      git_host_id: '',
      repository: '',
      branch: 'main',
      root_path: '',
      write_mode: 'direct_commit',
    },
  })

  useEffect(() => {
    if (binding) {
      form.reset({
        git_host_id: binding.gitHostId,
        repository: binding.repository,
        branch: binding.branch,
        root_path: binding.rootPath,
        write_mode: writeModeToForm(binding.writeMode),
      })
    }
  }, [form, binding])

  function handleSubmit(values: FormValues) {
    putMutation.mutate(
      {
        gitHostId: values.git_host_id,
        repository: values.repository.trim(),
        branch: values.branch.trim(),
        rootPath: values.root_path?.trim() || undefined,
        writeMode: writeModeToProto(values.write_mode),
      },
      {
        onSuccess: () =>
          toast.success(binding ? 'Binding updated' : 'Binding created'),
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <GitBranch className='h-4 w-4' />
          Configuration
        </CardTitle>
        <CardDescription>
          Where this workspace's agent content lives and how changes are
          written back.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className='space-y-4'
          >
            <FormField
              control={form.control}
              name='git_host_id'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Git host</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue
                          placeholder={
                            hostsLoading ? 'Loading hosts…' : 'Select a Git host'
                          }
                        />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {hosts.map((host) => (
                        <SelectItem key={host.id} value={host.id}>
                          {host.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {!hostsLoading && hosts.length === 0 && (
                    <p className='text-xs text-muted-foreground'>
                      No Git hosts are configured. Ask a platform administrator
                      to add one under Admin → Git Hosts.
                    </p>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='repository'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Repository</FormLabel>
                  <FormControl>
                    <Input placeholder='owner/repo' {...field} />
                  </FormControl>
                  <p className='text-xs text-muted-foreground'>
                    Host-relative path: <span className='font-mono'>owner/repo</span>{' '}
                    on GitHub, <span className='font-mono'>group/project</span> on GitLab.
                  </p>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className='grid gap-4 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='branch'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Branch</FormLabel>
                    <FormControl>
                      <Input placeholder='main' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='root_path'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Root path (optional)</FormLabel>
                    <FormControl>
                      <Input placeholder='Repository root' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <FormField
              control={form.control}
              name='write_mode'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Write mode</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value='direct_commit'>
                        Direct commit
                      </SelectItem>
                      <SelectItem value='change_request'>
                        Pull / Merge request
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <p className='text-xs text-muted-foreground'>
                    Direct commit writes to the bound branch; Pull / Merge
                    request opens a change request instead.
                  </p>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className='flex gap-2'>
              <Button type='submit' disabled={putMutation.isPending}>
                {putMutation.isPending ? (
                  <RefreshCw className='mr-2 h-4 w-4 animate-spin' />
                ) : null}
                {binding ? 'Save Binding' : 'Create Binding'}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  )
}

function ConnectionCard({ binding }: { binding: WorkspaceRepoBinding }) {
  const validateMutation = useValidateRepoBinding()
  const status = binding.status
  const state = status?.state ?? RepoBindingConnectionState.UNVALIDATED
  const checks = status?.checks ?? []

  function handleValidate() {
    validateMutation.mutate(undefined, {
      onSuccess: (updated) => {
        const nextState = updated.status?.state
        if (nextState === RepoBindingConnectionState.OK) {
          toast.success('Connection validated')
        } else {
          toast.error(updated.status?.error || 'Validation failed')
        }
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <ShieldCheck className='h-4 w-4' />
          Connection
        </CardTitle>
        <CardDescription>
          Probes the repository with the stored credential and records the
          result.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='flex flex-wrap items-center gap-3'>
          <ConnectionStateBadge state={state} />
          <span className='text-xs text-muted-foreground'>
            Last validated: {formatTs(status?.lastValidatedAt)}
          </span>
        </div>
        {status?.error ? (
          <p className='rounded-md border border-destructive/30 bg-danger-muted px-3 py-2 text-xs text-danger-foreground'>
            {status.error}
          </p>
        ) : null}
        {checks.length > 0 && (
          <ul className='space-y-2'>
            {checks.map((check) => (
              <CheckRow key={check.name} check={check} />
            ))}
          </ul>
        )}
        <Button
          type='button'
          variant='outline'
          onClick={handleValidate}
          disabled={validateMutation.isPending}
        >
          {validateMutation.isPending ? (
            <RefreshCw className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <ShieldCheck className='mr-2 h-4 w-4' />
          )}
          Validate connection
        </Button>
      </CardContent>
    </Card>
  )
}

function CheckRow({ check }: { check: RepoBindingCheck }) {
  return (
    <li className='flex items-start gap-2 text-sm'>
      {check.ok ? (
        <Check className='mt-0.5 h-4 w-4 shrink-0 text-success-foreground' />
      ) : (
        <X className='mt-0.5 h-4 w-4 shrink-0 text-danger-foreground' />
      )}
      <div className='min-w-0'>
        <div className='flex flex-wrap items-center gap-2'>
          <span className='font-mono text-xs'>{check.name}</span>
          {check.required && (
            <Badge variant='outline' className='text-[10px]'>
              required
            </Badge>
          )}
        </div>
        {check.detail && (
          <p className='text-xs text-muted-foreground'>{check.detail}</p>
        )}
      </div>
    </li>
  )
}

function CredentialCard({ binding }: { binding: WorkspaceRepoBinding }) {
  const setCredential = useSetRepoBindingCredential()
  const [open, setOpen] = useState(false)
  const [pat, setPat] = useState('')

  function handleSubmit() {
    if (!pat) {
      toast.error('Personal access token is required')
      return
    }
    setCredential.mutate(pat, {
      onSuccess: () => {
        toast.success('Credential saved')
        setOpen(false)
        setPat('')
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <KeyRound className='h-4 w-4' />
          Credential
        </CardTitle>
        <CardDescription>
          Personal access token used to talk to the repository.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {binding.credentialSet ? (
          <div className='flex flex-wrap items-center gap-2 text-sm'>
            <Badge className='bg-success-muted text-success-foreground'>
              Credential set
            </Badge>
            <span className='text-xs text-muted-foreground'>
              Updated {formatTs(binding.credentialUpdatedAt)}
            </span>
          </div>
        ) : (
          <div className='text-sm text-muted-foreground'>No credential</div>
        )}
        <Button type='button' variant='outline' onClick={() => setOpen(true)}>
          <KeyRound className='mr-2 h-4 w-4' />
          {binding.credentialSet ? 'Replace credential' : 'Set credential'}
        </Button>
      </CardContent>

      <Dialog
        open={open}
        onOpenChange={(next) => {
          setOpen(next)
          if (!next) setPat('')
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {binding.credentialSet ? 'Replace credential' : 'Set credential'}
            </DialogTitle>
            <DialogDescription>
              The personal access token is stored encrypted and is never shown
              again. Setting a new token replaces the previous one.
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-2'>
            <Label htmlFor='repo-binding-pat'>Personal access token</Label>
            <Input
              id='repo-binding-pat'
              type='password'
              autoComplete='new-password'
              value={pat}
              onChange={(e) => setPat(e.target.value)}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={setCredential.isPending}>
              {setCredential.isPending ? 'Saving…' : 'Save credential'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}

function DangerZoneCard({ binding }: { binding: WorkspaceRepoBinding }) {
  const deleteMutation = useDeleteRepoBinding()
  const [confirmOpen, setConfirmOpen] = useState(false)

  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-base'>Remove binding</CardTitle>
        <CardDescription>
          Removes the binding and its stored credential. Repository content is
          not touched.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button
          type='button'
          variant='outline'
          className='text-destructive hover:text-destructive'
          onClick={() => setConfirmOpen(true)}
        >
          <Trash2 className='mr-2 h-4 w-4' />
          Remove binding
        </Button>
      </CardContent>

      <DeleteDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title='Remove repository binding'
        description={`Remove the binding to ${binding.repository}? The stored credential is deleted with it.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          deleteMutation.mutate(undefined, {
            onSuccess: () => {
              toast.success('Binding removed')
              setConfirmOpen(false)
            },
            onError: (err) => toast.error(err.message),
          })
        }}
      />
    </Card>
  )
}

function OverlapsAlert({ overlaps }: { overlaps: RepoBindingOverlap[] }) {
  const names = overlaps
    .map((o) => o.workspaceName || o.workspaceId)
    .join(', ')
  return (
    <Alert>
      <Info className='h-4 w-4' />
      <AlertTitle>Shared repository location</AlertTitle>
      <AlertDescription>
        This repository location is also bound by workspace
        {overlaps.length > 1 ? 's' : ''}: {names}. Agent content at this
        location is intentionally shared between these workspaces.
      </AlertDescription>
    </Alert>
  )
}
