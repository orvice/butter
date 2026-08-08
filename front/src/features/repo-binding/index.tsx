import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { timestampDate, type Timestamp } from '@bufbuild/protobuf/wkt'
import {
  Check,
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
  useSetRepoBindingCredential,
  useValidateRepoBinding,
} from '@/api/repo-binding'
import { useGitHosts } from '@/api/git-hosts'
import {
  RepoBindingConnectionState,
  RepoBindingWriteMode,
  type RepoBindingCheck,
  type RepoBindingOverlap,
  type WorkspaceRepoBinding,
} from '@/gen/agents/v1/repobinding_pb'
import { DeleteDialog } from '@/components/delete-dialog'
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
          <div className='grid max-w-6xl gap-6 xl:grid-cols-[minmax(0,1fr)_420px]'>
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
        )}
      </PageScroll>
    </Page>
  )
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
