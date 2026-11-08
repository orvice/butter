import { useState } from 'react'
import type { PlaneApp } from '@/gen/agents/v1/plane_pb'
import { toast } from 'sonner'
import { usePlaneInstallations, useUpdatePlaneApp } from '@/api/plane'
import { useCanManageWorkspace } from '@/hooks/use-workspace-role'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'

export function PlaneInboundCard({
  app,
  profileVerified,
}: {
  app: PlaneApp
  profileVerified: boolean
}) {
  const { data: installations } = usePlaneInstallations(app.id)
  const { canManage } = useCanManageWorkspace()
  const update = useUpdatePlaneApp()
  const [confirmedAccess, setConfirmedAccess] = useState(false)
  const hasInstallation = installations?.some(
    (installation) =>
      installation.status === 'installed' &&
      installation.credentialState === 'valid'
  )
  const ready =
    app.clientSecretSet &&
    app.webhookSecretSet &&
    profileVerified &&
    hasInstallation

  async function toggle() {
    try {
      await update.mutateAsync({
        id: app.id,
        revision: app.revision,
        displayName: app.displayName,
        clientId: app.clientId,
        agentId: app.agentId,
        allowedUserIds: app.allowedUserIds,
        maxRunSeconds: app.maxRunSeconds,
        inboundEnabled: !app.inboundEnabled,
      })
      setConfirmedAccess(false)
      toast.success(
        app.inboundEnabled ? 'Plane inbound disabled' : 'Plane inbound enabled'
      )
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Update failed')
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          Inbound prompts
          <Badge variant='outline'>{app.inboundEnabled ? 'On' : 'Off'}</Badge>
        </CardTitle>
        <CardDescription>
          Allow the installed Plane App to run the bound Butter Agent.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-3 text-sm'>
        {!profileVerified && (
          <p className='text-muted-foreground'>
            An administrator must verify this deployment’s Plane provider
            profile before inbound can be enabled. Installation remains
            available for protocol verification.
          </p>
        )}
        {!hasInstallation && (
          <p className='text-muted-foreground'>
            Install to a Plane workspace first.
          </p>
        )}
        {(!app.clientSecretSet || !app.webhookSecretSet) && (
          <p className='text-muted-foreground'>
            Save both the client secret and webhook secret before enabling
            inbound.
          </p>
        )}
        {!app.inboundEnabled && (
          <div className='flex items-start gap-2'>
            <Checkbox
              id='plane-confirm-access'
              checked={confirmedAccess}
              onCheckedChange={(checked) =>
                setConfirmedAccess(checked === true)
              }
              disabled={!canManage || !ready}
            />
            <Label htmlFor='plane-confirm-access' className='leading-relaxed'>
              {app.allowedUserIds.length
                ? 'I allow the configured Plane users to prompt and stop this Agent.'
                : 'I allow all members who can use this installed Plane App to prompt and stop this Agent.'}{' '}
              This is a shared conversation. Users may see earlier participants’
              context and use this Agent’s tools, including commands on a
              ButterBox when configured.
            </Label>
          </div>
        )}
        <Button
          variant={app.inboundEnabled ? 'outline' : 'default'}
          onClick={toggle}
          disabled={
            !canManage ||
            update.isPending ||
            (!app.inboundEnabled && (!ready || !confirmedAccess))
          }
        >
          {app.inboundEnabled ? 'Disable inbound' : 'Enable inbound'}
        </Button>
      </CardContent>
    </Card>
  )
}
