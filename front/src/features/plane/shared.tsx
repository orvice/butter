import { PlaneAppCredentialState } from '@/gen/agents/v1/plane_pb'
import { CircleSlash, ShieldCheck } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'

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

/**
 * Inbound is off in this version of the integration and there is no switch
 * for it. Saying so is better than a disabled toggle the operator would keep
 * trying to flip.
 */
export function PlaneInboundNotice() {
  return (
    <Alert>
      <CircleSlash className='h-4 w-4' />
      <AlertTitle>Inbound is off</AlertTitle>
      <AlertDescription>
        This version only registers the App. Delivering Plane webhooks to an
        Agent — and therefore turning inbound on — ships with the execution
        path, so a registered App does not accept prompts yet.
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
