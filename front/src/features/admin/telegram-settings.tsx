import { useState } from 'react'
import { Navigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { Send } from 'lucide-react'
import { useTelegramSettings, useUpdateTelegramSettings } from '@/api/telegram'
import { useAuth } from '@/hooks/use-auth'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'

/**
 * The Telegram webhook base URL is platform-level, not per-workspace: it
 * names the public address of this deployment behind its load balancer.
 * Only global admins may see or change it.
 */
export function AdminTelegramSettingsPage() {
  const { isAdmin, isLoading: isAuthLoading } = useAuth()
  const { data: settings, isLoading } = useTelegramSettings(isAdmin)
  const update = useUpdateTelegramSettings()
  const [baseUrl, setBaseUrl] = useState<string | null>(null)

  if (!isAuthLoading && !isAdmin) return <Navigate to='/403' />

  const value = baseUrl ?? settings?.webhookBaseUrl ?? ''

  return (
    <Page>
      <PageHeader
        title='Telegram platform settings'
        subtitle='Where Telegram delivers webhook callbacks for every workspace.'
      />
      <PageScroll>
        <Card className='max-w-2xl'>
          <CardHeader>
            <CardTitle className='flex items-center gap-2 text-base'>
              <Send className='h-4 w-4' />
              Webhook base URL
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            {isLoading ? (
              <Skeleton className='h-10' />
            ) : (
              <>
                <div className='space-y-2'>
                  <Label htmlFor='webhook-base-url'>Public base URL</Label>
                  <Input
                    id='webhook-base-url'
                    value={value}
                    placeholder='https://butter.example.com'
                    onChange={(e) => setBaseUrl(e.target.value)}
                  />
                  <p className='text-xs text-muted-foreground'>
                    Must be HTTPS with no path — Telegram only delivers over TLS, and each
                    channel's callback path is derived from its immutable ID. Leave empty
                    to stop registering webhooks.
                  </p>
                </div>
                <div className='flex justify-end'>
                  <Button
                    disabled={update.isPending}
                    onClick={async () => {
                      try {
                        await update.mutateAsync(value.trim())
                        toast.success('Telegram settings updated')
                      } catch (err) {
                        toast.error(err instanceof Error ? err.message : 'Update failed')
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
