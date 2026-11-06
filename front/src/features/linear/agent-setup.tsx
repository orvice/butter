import { Link, useNavigate, useParams, useSearch } from '@tanstack/react-router'
import {
  Check,
  CircleAlert,
  ExternalLink,
  Plus,
  SquareKanban,
} from 'lucide-react'
import { toast } from 'sonner'
import { useAgent } from '@/api/agents'
import {
  useCreateLinearApp,
  useLinearApps,
  type CreateLinearAppInput,
} from '@/api/linear'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { LinearDiagnosticsCard } from './diagnostics'
import { LinearAppForm, type LinearAppFormValues } from './form'

const steps = [
  'Check prerequisites',
  'Choose or register an App',
  'Install and configure',
  'Enable and verify',
]

export function LinearAgentSetup() {
  const { name: agentId } = useParams({
    from: '/_authenticated/agents/$name/linear',
  })
  const search = useSearch({ from: '/_authenticated/agents/$name/linear' })
  const navigate = useNavigate()
  const { data: agent, isLoading: agentLoading } = useAgent(agentId)
  const { data: apps, isLoading: appsLoading } = useLinearApps()
  const create = useCreateLinearApp()
  const step = Math.min(Math.max(search.step ?? 1, 1), 4)
  const selectedApp = apps?.find(
    (app) => app.id === search.app && app.agentId === agentId
  )
  const agentApps = (apps ?? []).filter((app) => app.agentId === agentId)
  const canManage =
    agent?.agent?.lifecycle_status === undefined ||
    agent?.agent?.lifecycle_status === 'AGENT_LIFECYCLE_STATUS_ACTIVE' ||
    agent?.agent?.lifecycle_status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED'

  function go(nextStep: number, app?: string) {
    navigate({
      to: '/agents/$name/linear',
      params: { name: agentId },
      search: { step: nextStep, ...(app ? { app } : {}) },
    })
  }

  function createApp(values: LinearAppFormValues) {
    const input: CreateLinearAppInput = {
      displayName: values.displayName,
      clientId: values.clientId,
      agentId,
      inboundEnabled: false,
      allowedUserIds: values.allowedUserIds,
      maxRunSeconds: values.maxRunSeconds,
      clientSecret: values.clientSecret,
      webhookSecret: values.webhookSecret,
    }
    create.mutate(input, {
      onSuccess: (app) => {
        toast.success('Linear App registered')
        go(3, app.id)
      },
      onError: (error) => toast.error(error.message),
    })
  }

  if (agentLoading || appsLoading || !agent?.agent) {
    return <div className='p-6'>Loading Linear setup…</div>
  }

  return (
    <Page>
      <PageHeader
        className='max-w-4xl'
        title={`Connect Linear to ${agent.agent.name}`}
        subtitle='A resumable setup using the existing Linear App and Installation resources.'
        actions={
          <Button variant='outline' asChild>
            <Link to='/agents/$name/edit' params={{ name: agentId }}>
              Back to Agent
            </Link>
          </Button>
        }
      />
      <PageScroll className='max-w-4xl'>
        <div className='space-y-6'>
          <nav
            aria-label='Linear setup progress'
            className='grid gap-2 sm:grid-cols-4'
          >
            {steps.map((label, index) => {
              const number = index + 1
              return (
                <button
                  key={label}
                  type='button'
                  className={`rounded-md border p-3 text-left text-sm ${number === step ? 'border-primary bg-primary/5' : number < step ? 'border-emerald-500/40' : 'text-muted-foreground'}`}
                  onClick={() => number <= step && go(number, search.app)}
                  disabled={number > step}
                >
                  <span className='flex items-center gap-2 font-medium'>
                    {number < step ? (
                      <Check className='size-4 text-emerald-600' />
                    ) : (
                      <span className='flex size-4 items-center justify-center rounded-full border text-[10px]'>
                        {number}
                      </span>
                    )}
                    {label}
                  </span>
                </button>
              )
            })}
          </nav>

          {step === 1 && (
            <Card>
              <CardHeader>
                <CardTitle>Before you start</CardTitle>
                <CardDescription>
                  These deployment prerequisites are checked by the server when
                  you save or enable the App.
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-3'>
                <Prerequisite
                  ok={canManage}
                  text='You can manage this ACTIVE Agent.'
                />
                <Prerequisite
                  ok
                  text='Use an administrator configured public base URL and durable Redis deployment.'
                />
                <Prerequisite
                  ok
                  text='Create or use an OAuth App in Linear with Agent session events enabled.'
                />
                {!canManage && (
                  <p className='text-sm text-destructive'>
                    This Agent is not active. Activate it before configuring
                    Linear.
                  </p>
                )}
                <Button disabled={!canManage} onClick={() => go(2, search.app)}>
                  Continue
                </Button>
              </CardContent>
            </Card>
          )}

          {step === 2 && (
            <div className='space-y-4'>
              <Card>
                <CardHeader>
                  <CardTitle>Apps already routed to this Agent</CardTitle>
                  <CardDescription>
                    Continue an existing App to preserve its installation and
                    saved progress, or explicitly register another App.
                  </CardDescription>
                </CardHeader>
                <CardContent className='space-y-3'>
                  {agentApps.length === 0 && (
                    <p className='text-sm text-muted-foreground'>
                      No Linear App is associated with this Agent yet.
                    </p>
                  )}
                  {agentApps.map((app) => (
                    <div
                      key={app.id}
                      className='flex items-center justify-between gap-3 rounded-md border p-3'
                    >
                      <div className='flex min-w-0 items-center gap-2'>
                        <SquareKanban className='size-4 text-muted-foreground' />
                        <div>
                          <div className='font-medium'>{app.displayName}</div>
                          <div className='text-xs text-muted-foreground'>
                            {app.clientId}
                          </div>
                        </div>
                      </div>
                      <Button size='sm' onClick={() => go(3, app.id)}>
                        Continue
                      </Button>
                    </div>
                  ))}
                  <Button variant='outline' onClick={() => go(3)}>
                    <Plus className='size-4' /> Register another App
                  </Button>
                </CardContent>
              </Card>
              {apps?.some((app) => app.agentId !== agentId) && (
                <p className='text-xs text-muted-foreground'>
                  Apps routed to another Agent are not selectable here. Changing
                  an App's Agent is an explicit edit and never happens silently.
                </p>
              )}
            </div>
          )}

          {step === 3 && selectedApp && (
            <Card>
              <CardHeader>
                <CardTitle>Continue {selectedApp.displayName}</CardTitle>
                <CardDescription>
                  Resume registration, secret entry, OAuth installation,
                  allowlist, and runtime settings.
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-4'>
                <p className='text-sm text-muted-foreground'>
                  The App is still routed to the selected Agent. OAuth
                  cancellation or failure returns here without marking the
                  installation successful.
                </p>
                <Button asChild>
                  <Link
                    to='/linear-apps/$id/edit'
                    params={{ id: selectedApp.id }}
                  >
                    Open App configuration <ExternalLink className='size-4' />
                  </Link>
                </Button>
                <Button variant='outline' onClick={() => go(2)}>
                  Choose another App
                </Button>
              </CardContent>
            </Card>
          )}

          {step === 3 && !selectedApp && (
            <Card>
              <CardHeader>
                <CardTitle>Register a Linear App</CardTitle>
                <CardDescription>
                  Enter the OAuth client details and write-only secrets. The
                  Agent is preselected and cannot be changed by this flow.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <LinearAppForm
                  mode='create'
                  lockedAgentId={agentId}
                  submitLabel='Register App'
                  loading={create.isPending}
                  onCancel={() => go(2)}
                  onSubmit={createApp}
                />
              </CardContent>
            </Card>
          )}

          {step === 4 && selectedApp && (
            <Card>
              <CardHeader>
                <CardTitle>Enable and verify</CardTitle>
                <CardDescription>
                  Finish in the App configuration page after OAuth returns.
                  Enabling inbound is separate from installation.
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-4'>
                <p className='text-sm text-muted-foreground'>
                  An empty allowlist means every usable user in the installed
                  organization can invoke the Agent. Review the Pi/Cursor risk
                  warning before enabling.
                </p>
                <Button asChild>
                  <Link
                    to='/linear-apps/$id/edit'
                    params={{ id: selectedApp.id }}
                  >
                    Open verification page
                  </Link>
                </Button>
              </CardContent>
            </Card>
          )}

          {step === 4 && selectedApp && (
            <LinearDiagnosticsCard appId={selectedApp.id} live />
          )}
        </div>
      </PageScroll>
    </Page>
  )
}

function Prerequisite({ ok, text }: { ok: boolean; text: string }) {
  return (
    <div className='flex items-start gap-2 text-sm'>
      {ok ? (
        <Check className='mt-0.5 size-4 text-emerald-600' />
      ) : (
        <CircleAlert className='mt-0.5 size-4 text-destructive' />
      )}
      {text}
    </div>
  )
}
