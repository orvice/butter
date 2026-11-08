import { PlaneAppCredentialState } from '@/gen/agents/v1/plane_pb'
import { CircleSlash, Copy, ShieldCheck } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function PlaneCredentialBadge({
  state,
}: {
  state: PlaneAppCredentialState
}) {
  if (state === PlaneAppCredentialState.COMPLETE) {
    return <Badge variant='outline'>Secrets set</Badge>
  }
  return (
    <Badge variant='outline' className='text-muted-foreground'>
      Secrets missing
    </Badge>
  )
}

export function PlaneEndpointCard({
  callbackUrl,
  webhookUrl,
}: {
  callbackUrl: string
  webhookUrl: string
}) {
  async function copy(value: string) {
    if (value) await navigator.clipboard?.writeText(value)
  }
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-base'>Plane endpoints</CardTitle>
        <CardDescription>
          Register these URLs in the Plane App configuration.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-3 text-sm'>
        <EndpointRow
          label='OAuth callback URL'
          value={callbackUrl}
          onCopy={copy}
        />
        <EndpointRow label='Webhook URL' value={webhookUrl} onCopy={copy} />
        {!callbackUrl && (
          <p className='text-xs text-muted-foreground'>
            A global administrator must configure Plane public base URL under
            Public endpoints first.
          </p>
        )}
      </CardContent>
    </Card>
  )
}

function EndpointRow({
  label,
  value,
  onCopy,
}: {
  label: string
  value: string
  onCopy: (value: string) => void
}) {
  return (
    <div className='space-y-1'>
      <div className='font-medium'>{label}</div>
      <div className='flex items-center gap-2'>
        <code className='min-w-0 flex-1 rounded bg-muted px-2 py-1 text-xs break-all'>
          {value || 'Not configured'}
        </code>
        <button
          type='button'
          aria-label={`Copy ${label}`}
          className='rounded p-1 hover:bg-muted disabled:opacity-40'
          disabled={!value}
          onClick={() => onCopy(value)}
        >
          <Copy className='h-4 w-4' />
        </button>
      </div>
    </div>
  )
}

/**
 * New registrations stay off until an operator completes installation and
 * explicitly enables access on a verified provider profile.
 */
export function PlaneInboundNotice() {
  return (
    <Alert>
      <CircleSlash className='h-4 w-4' />
      <AlertTitle>Install before enabling inbound</AlertTitle>
      <AlertDescription>
        Register the App, install it to a Plane workspace, then review access
        and enable inbound on the App page. A verified provider profile and both
        secrets are required before the Agent can accept prompts.
      </AlertDescription>
    </Alert>
  )
}

/**
 * The provider profile is deployment configuration, never an App field.
 * Showing it read-only makes that visible instead of implying the operator
 * could point an App at another Plane.
 */
export function PlaneProviderNotice({
  name,
  apiBaseUrl,
  edition,
}: {
  name: string
  apiBaseUrl: string
  edition: string
}) {
  return (
    <Alert>
      <ShieldCheck className='h-4 w-4' />
      <AlertTitle>Provider profile: {name || 'unavailable'}</AlertTitle>
      <AlertDescription>
        Plane API <span className='font-mono text-xs'>{apiBaseUrl}</span>
        {edition && ` (${edition})`}. Fixed by this deployment's startup
        configuration; an App cannot choose or change it.
      </AlertDescription>
    </Alert>
  )
}
