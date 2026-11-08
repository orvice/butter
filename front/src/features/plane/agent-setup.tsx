import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { Plane, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { useAgent } from '@/api/agents'
import { useCreatePlaneApp, usePlaneApps } from '@/api/plane'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { PlaneAppForm, type PlaneAppFormValues } from './form'
import { PlaneCredentialBadge, PlaneInboundNotice } from './shared'

/**
 * Agent-scoped entry point: the App is registered for this Agent, and the
 * Agent is fixed by the route rather than chosen in the form.
 */
export function PlaneAgentSetup() {
  const { name: agentId } = useParams({
    from: '/_authenticated/agents/$name/plane',
  })
  const navigate = useNavigate()
  const { data: agent, isLoading: agentLoading } = useAgent(agentId)
  const { data: apps, isLoading: appsLoading } = usePlaneApps()
  const create = useCreatePlaneApp()
  const agentApps = (apps ?? []).filter((app) => app.agentId === agentId)

  function onSubmit(values: PlaneAppFormValues) {
    create.mutate(
      {
        displayName: values.displayName,
        clientId: values.clientId,
        // The locked Agent wins; the URL cannot route the App elsewhere.
        agentId,
        allowedUserIds: values.allowedUserIds,
        maxRunSeconds: values.maxRunSeconds,
        clientSecret: values.clientSecret,
        webhookSecret: values.webhookSecret,
      },
      {
        onSuccess: (app) => {
          toast.success('Plane App registered')
          navigate({ to: '/plane-apps/$id/edit', params: { id: app.id } })
        },
        onError: (error) => toast.error(error.message),
      }
    )
  }

  if (agentLoading || appsLoading || !agent?.agent) {
    return <div className='p-6'>Loading Plane setup…</div>
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title={`Connect Plane to ${agent.agent.name}`}
        subtitle='Register a Plane App bound to this Agent, then open Manage to install it and configure inbound access.'
        actions={
          <Button variant='outline' asChild>
            <Link to='/agents/$name/edit' params={{ name: agentId }}>
              Back to Agent
            </Link>
          </Button>
        }
      />
      <PageScroll className='max-w-3xl'>
        <div className='space-y-6'>
          <PlaneInboundNotice />
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Plane className='h-4 w-4' />
                Registered Apps
              </CardTitle>
              <CardDescription>
                Apps already bound to this Agent.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-2'>
              {agentApps.length === 0 ? (
                <p className='text-sm text-muted-foreground'>
                  No Plane App is bound to this Agent yet.
                </p>
              ) : (
                agentApps.map((app) => (
                  <div
                    key={app.id}
                    className='flex items-center justify-between rounded-md border p-3'
                  >
                    <div>
                      <div className='font-medium'>{app.displayName}</div>
                      <div className='font-mono text-xs text-muted-foreground'>
                        {app.clientId}
                      </div>
                    </div>
                    <div className='flex items-center gap-2'>
                      <Badge
                        variant='outline'
                        className='text-muted-foreground'
                      >
                        {app.inboundEnabled ? 'Inbound on' : 'Inbound off'}
                      </Badge>
                      <PlaneCredentialBadge state={app.credentialState} />
                      <Button size='sm' variant='outline' asChild>
                        <Link to='/plane-apps/$id/edit' params={{ id: app.id }}>
                          Manage
                        </Link>
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Plus className='h-4 w-4' />
                Register another App
              </CardTitle>
              <CardDescription>
                The Agent is fixed to {agent.agent.name} for this entry point.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <PlaneAppForm
                mode='create'
                lockedAgentId={agentId}
                submitLabel='Register'
                loading={create.isPending}
                onCancel={() =>
                  navigate({
                    to: '/agents/$name/edit',
                    params: { name: agentId },
                  })
                }
                onSubmit={onSubmit}
              />
            </CardContent>
          </Card>
        </div>
      </PageScroll>
    </Page>
  )
}
