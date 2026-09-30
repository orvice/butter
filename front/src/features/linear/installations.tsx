import { useState } from 'react'
import { LinearInstallationCredentialState } from '@/gen/agents/v1/linear_pb'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Building2, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  useBeginLinearInstall,
  useDeleteLinearInstallation,
  useLinearInstallations,
} from '@/api/linear'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { DeleteDialog } from '@/components/delete-dialog'

const INSTALL_REASONS: Record<string, string> = {
  denied: 'The install was cancelled or denied in Linear.',
  state_invalid:
    'The install link expired or was already used. Start the install again.',
  app_missing: 'The Linear App no longer exists.',
  exchange_failed:
    'Linear did not accept the authorization. Check the client ID and secret, then try again.',
  identity_failed: 'Linear did not report which organization this is.',
  storage_failed: 'The installation could not be saved. Try again.',
}

/** The outcome of the OAuth redirect back from Linear, if any. */
export function InstallResultBanner({
  outcome,
  reason,
}: {
  outcome?: string
  reason?: string
}) {
  if (outcome === 'success') {
    return (
      <Alert>
        <AlertTitle>Installed</AlertTitle>
        <AlertDescription>
          The Linear App is installed. Delegate an issue to it or mention it in
          Linear.
        </AlertDescription>
      </Alert>
    )
  }
  if (outcome === 'error') {
    return (
      <Alert variant='destructive'>
        <AlertTitle>Install failed</AlertTitle>
        <AlertDescription>
          {INSTALL_REASONS[reason ?? ''] ?? 'The install did not complete.'}
        </AlertDescription>
      </Alert>
    )
  }
  return null
}

export function InstallationsCard({
  appId,
  canInstall,
}: {
  appId: string
  canInstall: boolean
}) {
  const { data, isLoading } = useLinearInstallations(appId)
  const begin = useBeginLinearInstall()
  const remove = useDeleteLinearInstallation()
  const [removeTarget, setRemoveTarget] = useState<string | null>(null)

  function install() {
    begin.mutate(
      {
        appId,
        returnUrl: `${window.location.origin}/linear-apps/${appId}/edit`,
      },
      {
        onSuccess: (authorizeUrl) => window.location.assign(authorizeUrl),
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between pb-2'>
        <CardTitle className='flex items-center gap-2 text-base'>
          <Building2 className='h-4 w-4' />
          Installations
        </CardTitle>
        <Button
          size='sm'
          onClick={install}
          disabled={!canInstall || begin.isPending}
        >
          <Plus className='size-4' />
          Install to Linear
        </Button>
      </CardHeader>
      <CardContent className='space-y-3'>
        {!canInstall && (
          <p className='text-sm text-muted-foreground'>
            Set the client secret and the Linear public base URL before
            installing.
          </p>
        )}
        {isLoading ? (
          <Skeleton className='h-12' />
        ) : (data ?? []).length === 0 ? (
          <p className='text-sm text-muted-foreground'>
            Not installed in any Linear organization yet.
          </p>
        ) : (
          <ul className='divide-y rounded-md border'>
            {(data ?? []).map((inst) => (
              <li
                key={inst.id}
                className='flex items-center justify-between gap-3 p-3'
              >
                <div className='min-w-0'>
                  <div className='flex items-center gap-2 font-medium'>
                    {inst.organizationName || inst.organizationId}
                    {inst.credentialState ===
                    LinearInstallationCredentialState.NEEDS_REINSTALL ? (
                      <Badge variant='destructive'>Needs reinstall</Badge>
                    ) : (
                      <Badge variant='outline'>Connected</Badge>
                    )}
                  </div>
                  <div className='text-xs text-muted-foreground'>
                    App user {inst.appUserId}
                    {inst.installedAt &&
                      ` · installed ${timestampDate(inst.installedAt).toLocaleString()}`}
                  </div>
                  {inst.lastCredentialError && (
                    <div className='text-xs text-destructive'>
                      {inst.lastCredentialError}
                    </div>
                  )}
                </div>
                <Button
                  variant='ghost'
                  size='icon'
                  aria-label={`Remove ${inst.organizationName}`}
                  onClick={() => setRemoveTarget(inst.id)}
                >
                  <Trash2 className='h-4 w-4' />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
      <DeleteDialog
        open={!!removeTarget}
        onOpenChange={(open) => !open && setRemoveTarget(null)}
        title='Remove installation'
        description='Remove this installation? The Agent stops answering in that Linear organization, and its token is revoked.'
        loading={remove.isPending}
        onConfirm={() => {
          if (!removeTarget) return
          remove.mutate(
            { appId, id: removeTarget },
            {
              onSuccess: () => {
                toast.success('Installation removed')
                setRemoveTarget(null)
              },
              onError: (err) => toast.error(err.message),
            }
          )
        }}
      />
    </Card>
  )
}
