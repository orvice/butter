import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { KeyRound, Plane } from 'lucide-react'
import { toast } from 'sonner'
import {
  PLANE_APPS_KEY,
  usePlaneApp,
  usePlaneInstallations,
  usePlaneProviderProfile,
  usePutPlaneAppCredentials,
  useUpdatePlaneApp,
} from '@/api/plane'
import { Code, ConnectError } from '@/api/transport'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { PlaneAppForm, type PlaneAppFormValues } from './form'
import {
  PlaneCredentialBadge,
  PlaneInboundNotice,
  PlaneProviderNotice,
} from './shared'

export function PlaneAppEdit() {
  const { id } = useParams({ from: '/_authenticated/plane-apps/$id/edit' })
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data: app, isLoading } = usePlaneApp(id)
  const { data: installations = [] } = usePlaneInstallations(id)
  const updateMutation = useUpdatePlaneApp()
  const [conflict, setConflict] = useState(false)
  // Bumped after a conflict (or a successful save) so the form re-seeds from
  // the newest server revision even while it was dirty.
  const [reloadToken, setReloadToken] = useState(0)

  function onSubmit(values: PlaneAppFormValues) {
    if (!app) return
    const {
      clientSecret: _clientSecret,
      webhookSecret: _webhookSecret,
      ...fields
    } = values
    updateMutation.mutate(
      { ...fields, id, revision: app.revision },
      {
        onSuccess: () => {
          setConflict(false)
          setReloadToken((n) => n + 1)
          toast.success('Plane App updated')
        },
        onError: (err) => {
          // An aborted update means someone else saved a newer revision. Pull
          // it so the form re-seeds with it and the next save can succeed
          // instead of failing until the user reloads by hand.
          if (err instanceof ConnectError && err.code === Code.Aborted) {
            setConflict(true)
            setReloadToken((n) => n + 1)
            toast.error(
              'This Plane App changed elsewhere. The latest values were reloaded.'
            )
            queryClient.invalidateQueries({ queryKey: [...PLANE_APPS_KEY, id] })
            return
          }
          toast.error(err.message)
        },
      }
    )
  }

  if (isLoading || !app) {
    return (
      <div className='p-6'>
        <Skeleton className='h-96 w-full' />
      </div>
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title={app.displayName}
        subtitle={<PlaneCredentialBadge state={app.credentialState} />}
        actions={
          <Button variant='outline' asChild>
            <Link to='/agents/$name/plane' params={{ name: app.agentId }}>
              Agent setup
            </Link>
          </Button>
        }
      />
      <PageScroll className='max-w-3xl'>
        <div className='space-y-6'>
          {conflict && (
            <Alert variant='destructive'>
              <AlertTitle>The App changed elsewhere</AlertTitle>
              <AlertDescription>
                A newer revision is now loaded below. Review the current values
                and save again to apply your change.
              </AlertDescription>
            </Alert>
          )}
          {profile && (
            <PlaneProviderNotice
              name={profile.name}
              apiBaseUrl={profile.apiBaseUrl}
              edition={profile.edition}
            />
          )}
          <Card>
            <CardHeader className='pb-2'><CardTitle className='text-base'>Plane installations</CardTitle></CardHeader>
            <CardContent className='space-y-2 text-sm'>
              {installations.length === 0 ? <p className='text-muted-foreground'>No verified installation.</p> : installations.map((installation) => <div key={installation.id} className='rounded border p-2'><div className='font-medium'>{installation.workspaceSlug || installation.planeWorkspaceId}</div><div className='text-muted-foreground'>Granted scopes: {installation.grantedScopes.join(', ') || 'none'}</div><div>Status: {installation.status} · {installation.credentialState}</div></div>)}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Plane className='h-4 w-4' />
                Registration
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-1 text-sm'>
              <p>
                Provider profile{' '}
                <span className='font-mono'>{app.providerProfile}</span>
              </p>
              <p className='text-muted-foreground'>
                Server-assigned and immutable. Register a separate App against a
                different profile if the deployment offers one.
              </p>
            </CardContent>
          </Card>

          <PlaneAppForm
            mode='edit'
            submitLabel='Save'
            loading={updateMutation.isPending}
            initialValue={app}
            reloadToken={reloadToken}
            onCancel={() => navigate({ to: '/plane-apps' })}
            onSubmit={onSubmit}
          />

          <SecretsCard
            appId={app.id}
            clientSecretSet={app.clientSecretSet}
            webhookSecretSet={app.webhookSecretSet}
          />
        </div>
      </PageScroll>
    </Page>
  )
}

function SecretsCard({
  appId,
  clientSecretSet,
  webhookSecretSet,
}: {
  appId: string
  clientSecretSet: boolean
  webhookSecretSet: boolean
}) {
  const put = usePutPlaneAppCredentials()
  const [clientSecret, setClientSecret] = useState('')
  const [webhookSecret, setWebhookSecret] = useState('')

  function save(
    input: { clientSecret?: string; webhookSecret?: string },
    done: string
  ) {
    put.mutate(
      { appId, ...input },
      {
        onSuccess: () => {
          toast.success(done)
          setClientSecret('')
          setWebhookSecret('')
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='flex items-center gap-2 text-base'>
          <KeyRound className='h-4 w-4' />
          Secrets
        </CardTitle>
      </CardHeader>
      <CardContent className='space-y-4'>
        <p className='text-sm text-muted-foreground'>
          Write-only: encrypted at rest and never displayed again. Clearing a
          secret is allowed while inbound is off, which it always is in this
          version.
        </p>
        <SecretRow
          id='plane-client-secret-rotate'
          label='Client secret'
          isSet={clientSecretSet}
          value={clientSecret}
          onChange={setClientSecret}
          pending={put.isPending}
          onSave={() =>
            save({ clientSecret: clientSecret.trim() }, 'Client secret updated')
          }
          onClear={() => save({ clientSecret: '' }, 'Client secret cleared')}
        />
        <SecretRow
          id='plane-webhook-secret-rotate'
          label='Webhook signing secret'
          isSet={webhookSecretSet}
          value={webhookSecret}
          onChange={setWebhookSecret}
          pending={put.isPending}
          onSave={() =>
            save(
              { webhookSecret: webhookSecret.trim() },
              'Webhook secret updated'
            )
          }
          onClear={() => save({ webhookSecret: '' }, 'Webhook secret cleared')}
        />
      </CardContent>
    </Card>
  )
}

function SecretRow({
  id,
  label,
  isSet,
  value,
  onChange,
  pending,
  onSave,
  onClear,
}: {
  id: string
  label: string
  isSet: boolean
  value: string
  onChange: (value: string) => void
  pending: boolean
  onSave: () => void
  onClear: () => void
}) {
  return (
    <div className='space-y-2'>
      <Label htmlFor={id}>
        {label}{' '}
        <span className='font-normal text-muted-foreground'>
          ({isSet ? 'set' : 'not set'})
        </span>
      </Label>
      <div className='flex gap-2'>
        <Input
          id={id}
          type='password'
          autoComplete='new-password'
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        <Button
          variant='outline'
          onClick={onSave}
          disabled={!value.trim() || pending}
        >
          {isSet ? 'Replace' : 'Set'}
        </Button>
        <Button
          variant='outline'
          onClick={onClear}
          disabled={!isSet || pending}
        >
          Clear
        </Button>
      </div>
    </div>
  )
}
