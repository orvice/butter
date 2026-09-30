import { useState } from 'react'
import { Navigate } from '@tanstack/react-router'
import { SquareKanban } from 'lucide-react'
import { toast } from 'sonner'
import { useLinearSettings, useUpdateLinearSettings } from '@/api/linear'
import { useAuth } from '@/hooks/use-auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'

/**
 * The Linear public base URL is platform-level, not per-workspace: every
 * Linear App's callback and webhook URLs derive from it. Only global admins
 * may see or change it.
 */
export function AdminLinearSettingsPage() {
  const { isAdmin, isLoading: isAuthLoading } = useAuth()
  const { data: settings, isLoading } = useLinearSettings(isAdmin)
  const update = useUpdateLinearSettings()
  const [baseUrl, setBaseUrl] = useState<string | null>(null)

  if (!isAuthLoading && !isAdmin) return <Navigate to='/403' />

  const value = baseUrl ?? settings?.publicBaseUrl ?? ''

  return (
    <Page>
      <PageHeader
        title='Linear platform settings'
        subtitle='Where Linear delivers webhooks and OAuth callbacks for every workspace.'
      />
      <PageScroll>
        <Card className='max-w-2xl'>
          <CardHeader>
            <CardTitle className='flex items-center gap-2 text-base'>
              <SquareKanban className='h-4 w-4' />
              Public base URL
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            {isLoading ? (
              <Skeleton className='h-10' />
            ) : (
              <>
                <div className='space-y-2'>
                  <Label htmlFor='linear-base-url'>Public base URL</Label>
                  <Input
                    id='linear-base-url'
                    value={value}
                    placeholder='https://butter.example.com'
                    onChange={(e) => setBaseUrl(e.target.value)}
                  />
                  <p className='text-xs text-muted-foreground'>
                    HTTPS with no path (plain HTTP only for localhost). Each
                    Linear App's callback and webhook URLs are derived from it
                    and the App's immutable ID.
                  </p>
                </div>
                <div className='flex justify-end'>
                  <Button
                    disabled={update.isPending}
                    onClick={async () => {
                      try {
                        await update.mutateAsync(value.trim())
                        toast.success('Linear settings updated')
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
      </PageScroll>
    </Page>
  )
}
