import { Building2, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { useBeginPlaneInstall, usePlaneInstallations } from '@/api/plane'
import { useCanManageWorkspace } from '@/hooks/use-workspace-role'
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
import { Skeleton } from '@/components/ui/skeleton'

const reasons: Record<string, string> = {
  denied: 'Installation was cancelled or denied in Plane.',
  installation_missing:
    'Plane did not return an installation ID. Check the App redirect URI and start again.',
  state_invalid:
    'The install link expired, was already used, or Plane did not return its state. Start the install again.',
  permission_changed:
    'Your workspace management permission changed. Ask an owner to install the App.',
  app_missing: 'This Plane App no longer exists.',
  app_mismatch:
    'The App identity changed during installation. Start the install again.',
  exchange_failed:
    'Plane did not issue a Bot token. Check the client ID, client secret and provider configuration.',
  identity_failed:
    'Plane did not confirm this installation and its workspace. Check the installation permissions and provider configuration.',
  storage_failed: 'The installation could not be saved. Try installing again.',
}

export function PlaneInstallResult({
  outcome,
  reason,
}: {
  outcome?: string
  reason?: string
}) {
  if (outcome === 'success') {
    return (
      <Alert>
        <AlertTitle>Installed in Plane</AlertTitle>
        <AlertDescription>
          The installation was verified and its Bot token was saved. Review
          inbound access below before accepting prompts.
        </AlertDescription>
      </Alert>
    )
  }
  if (outcome !== 'error') return null
  return (
    <Alert variant='destructive'>
      <AlertTitle>Plane installation failed</AlertTitle>
      <AlertDescription>
        {reasons[reason ?? ''] ?? 'Installation did not complete. Start again.'}
      </AlertDescription>
    </Alert>
  )
}

export function PlaneInstallationsCard({
  appId,
  canInstall,
  installationsEnabled = true,
}: {
  appId: string
  canInstall: boolean
  installationsEnabled?: boolean
}) {
  const { data, isLoading, error } = usePlaneInstallations(
    appId,
    installationsEnabled
  )
  const { canManage } = useCanManageWorkspace()
  const begin = useBeginPlaneInstall()

  function install() {
    begin.mutate(
      {
        appId,
        returnUrl: `${window.location.origin}/plane-apps/${appId}/edit`,
      },
      {
        onSuccess: (authorizeUrl) => window.location.assign(authorizeUrl),
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between gap-3 pb-2'>
        <div>
          <CardTitle className='flex items-center gap-2 text-base'>
            <Building2 className='h-4 w-4' />
            Plane installations
          </CardTitle>
          <CardDescription className='mt-1'>
            Choose a Plane workspace and approve this App.
          </CardDescription>
        </div>
        <Button
          size='sm'
          onClick={install}
          disabled={!canManage || !canInstall || begin.isPending}
        >
          <Plus className='h-4 w-4' />
          {begin.isPending ? 'Opening Plane…' : 'Install to Plane'}
        </Button>
      </CardHeader>
      <CardContent className='space-y-3 text-sm'>
        {!canManage ? (
          <p className='text-muted-foreground'>
            A workspace owner or administrator can install this App.
          </p>
        ) : !canInstall ? (
          <p className='text-muted-foreground'>
            Save the client secret and configure the Plane public base URL
            before installing.
          </p>
        ) : null}
        {isLoading ? (
          <Skeleton className='h-12' />
        ) : error ? (
          <p className='text-destructive'>{error.message}</p>
        ) : !data?.length ? (
          <p className='text-muted-foreground'>No verified installation yet.</p>
        ) : (
          <ul className='divide-y rounded-md border'>
            {data.map((installation) => (
              <li key={installation.id} className='space-y-1 p-3'>
                <div className='flex items-center gap-2 font-medium'>
                  {installation.workspaceSlug || installation.planeWorkspaceId}
                  <Badge variant='outline'>{installation.status}</Badge>
                </div>
                <p className='text-muted-foreground'>
                  Credentials: {installation.credentialState}
                </p>
                <p className='text-muted-foreground'>
                  Granted scopes:{' '}
                  {installation.grantedScopes.join(', ') || 'Not reported'}
                </p>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
