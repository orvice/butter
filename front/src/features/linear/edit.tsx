import { useState } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import { KeyRound, Link2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  useLinearApp,
  usePutLinearAppCredentials,
  useUpdateLinearApp,
} from '@/api/linear'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { LinearAppForm, type LinearAppFormValues } from './form'
import { CopyField, CredentialStateBadge } from './shared'

export function LinearAppEdit() {
  const { id } = useParams({ from: '/_authenticated/linear-apps/$id/edit' })
  const navigate = useNavigate()
  const { data: app, isLoading } = useLinearApp(id)
  const updateMutation = useUpdateLinearApp()

  function onSubmit(values: LinearAppFormValues) {
    if (!app) return
    const {
      clientSecret: _clientSecret,
      webhookSecret: _webhookSecret,
      ...fields
    } = values
    updateMutation.mutate(
      { ...fields, id, revision: app.revision },
      {
        onSuccess: () => toast.success('Linear App updated'),
        onError: (err) => toast.error(err.message),
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
        subtitle={<CredentialStateBadge state={app.credentialState} />}
      />
      <PageScroll className='max-w-3xl'>
        <div className='space-y-6'>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Link2 className='h-4 w-4' />
                Register in Linear
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                In the Linear OAuth app, set the callback URL and the webhook
                URL below, and enable Agent session events. A platform admin
                configures the public base URL these are derived from.
              </p>
              <CopyField
                id='linear-callback-url'
                label='Callback URL'
                value={app.callbackUrl}
                emptyHint='Set the Linear public base URL first'
              />
              <CopyField
                id='linear-webhook-url'
                label='Webhook URL'
                value={app.webhookUrl}
                emptyHint='Set the Linear public base URL first'
              />
            </CardContent>
          </Card>

          <LinearAppForm
            mode='edit'
            submitLabel='Save'
            loading={updateMutation.isPending}
            initialValue={app}
            onCancel={() => navigate({ to: '/linear-apps' })}
            onSubmit={onSubmit}
          />

          <SecretsCard
            appId={app.id}
            clientSecretSet={app.clientSecretSet}
            webhookSecretSet={app.webhookSecretSet}
            inboundEnabled={app.inboundEnabled}
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
  inboundEnabled,
}: {
  appId: string
  clientSecretSet: boolean
  webhookSecretSet: boolean
  inboundEnabled: boolean
}) {
  const put = usePutLinearAppCredentials()
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
          Write-only: encrypted at rest and never displayed again.
          {inboundEnabled && ' Disable inbound before clearing a secret.'}
        </p>
        <SecretRow
          id='linear-client-secret-rotate'
          label='Client secret'
          isSet={clientSecretSet}
          value={clientSecret}
          onChange={setClientSecret}
          pending={put.isPending}
          canClear={!inboundEnabled}
          onSave={() =>
            save({ clientSecret: clientSecret.trim() }, 'Client secret updated')
          }
          onClear={() => save({ clientSecret: '' }, 'Client secret cleared')}
        />
        <SecretRow
          id='linear-webhook-secret-rotate'
          label='Webhook signing secret'
          isSet={webhookSecretSet}
          value={webhookSecret}
          onChange={setWebhookSecret}
          pending={put.isPending}
          canClear={!inboundEnabled}
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
  canClear,
  onSave,
  onClear,
}: {
  id: string
  label: string
  isSet: boolean
  value: string
  onChange: (value: string) => void
  pending: boolean
  canClear: boolean
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
          disabled={!isSet || !canClear || pending}
        >
          Clear
        </Button>
      </div>
    </div>
  )
}
