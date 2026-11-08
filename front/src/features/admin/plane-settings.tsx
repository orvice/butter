import { useState } from 'react'
import { Plane } from 'lucide-react'
import { toast } from 'sonner'
import { usePlaneSettings, useUpdatePlaneSettings } from '@/api/plane'
import { useAuth } from '@/hooks/use-auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'

/**
 * The Plane public base URL is platform-level, not per-workspace. The
 * provider profile is read-only here too: it comes from the deployment's
 * startup configuration and no RPC can change it.
 */
export function PlaneSettingsCard() {
  const { isAdmin } = useAuth()
  const { data, isLoading } = usePlaneSettings(isAdmin)
  const update = useUpdatePlaneSettings()
  const [baseUrl, setBaseUrl] = useState<string | null>(null)

  const settings = data?.settings
  const profile = data?.provider
  const value = baseUrl ?? settings?.publicBaseUrl ?? ''

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <Plane className='h-4 w-4' />
          Plane
        </CardTitle>
        <CardDescription>
          Callbacks and the fixed provider profile.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {isLoading ? (
          <Skeleton className='h-10' />
        ) : (
          <>
            <div className='space-y-2'>
              <Label htmlFor='plane-base-url'>Plane public base URL</Label>
              <Input
                id='plane-base-url'
                value={value}
                placeholder='https://butter.example.com'
                onChange={(e) => setBaseUrl(e.target.value)}
              />
              <p className='text-xs text-muted-foreground'>
                HTTPS with no path (plain HTTP only for localhost). Future
                callback and webhook URLs derive from it and the App's immutable
                ID.
              </p>
            </div>
            {profile && (
              <div className='flex items-center gap-2 text-sm'>
                <span className='text-muted-foreground'>Provider profile</span>
                <Badge variant='outline' className='font-mono'>
                  {profile.name}
                </Badge>
                <span className='font-mono text-xs text-muted-foreground'>
                  {profile.apiBaseUrl}
                </span>
                {!profile.verified && (
                  <span className='text-xs text-muted-foreground'>
                    (not verified against a live Plane instance)
                  </span>
                )}
              </div>
            )}
            <div className='flex justify-end'>
              <Button
                aria-label='Save Plane settings'
                disabled={update.isPending}
                onClick={async () => {
                  try {
                    await update.mutateAsync(value.trim())
                    toast.success('Plane settings updated')
                  } catch (err) {
                    toast.error(
                      err instanceof Error ? err.message : 'Update failed'
                    )
                  }
                }}
              >
                Save
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
