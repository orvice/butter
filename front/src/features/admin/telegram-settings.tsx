import { useState } from 'react'
import { toast } from 'sonner'
import { Send } from 'lucide-react'
import { useTelegramSettings, useUpdateTelegramSettings } from '@/api/telegram'
import { useAuth } from '@/hooks/use-auth'
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
 * The Telegram webhook base URL is platform-level, not per-workspace: it
 * names the public address of this deployment behind its load balancer.
 * Only global admins may see or change it.
 */
export function TelegramWebhookBaseUrlCard() {
  const { isAdmin } = useAuth()
  const { data: settings, isLoading } = useTelegramSettings(isAdmin)
  const update = useUpdateTelegramSettings()
  const [baseUrl, setBaseUrl] = useState<string | null>(null)

  const value = baseUrl ?? settings?.webhookBaseUrl ?? ''

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <Send className='h-4 w-4' />
          Telegram
        </CardTitle>
        <CardDescription>Bot webhook callbacks.</CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {isLoading ? (
          <Skeleton className='h-10' />
        ) : (
          <>
            <div className='space-y-2'>
              <Label htmlFor='webhook-base-url'>Telegram public base URL</Label>
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
                aria-label='Save Telegram settings'
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
  )
}
