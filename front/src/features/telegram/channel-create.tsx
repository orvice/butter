import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useCreateTelegramChannel } from '@/api/telegram'
import { TelegramReceiveMode } from '@/gen/agents/v1/telegram_pb'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { RECEIVE_MODE_LABELS } from './labels'

export function TelegramChannelCreate() {
  const navigate = useNavigate()
  const create = useCreateTelegramChannel()
  const [key, setKey] = useState('')
  const [name, setName] = useState('')
  const [botToken, setBotToken] = useState('')
  const [receiveMode, setReceiveMode] = useState<TelegramReceiveMode>(
    TelegramReceiveMode.WEBHOOK
  )

  async function submit() {
    try {
      const channel = await create.mutateAsync({ key, name, botToken, receiveMode })
      toast.success(`Registered @${channel.botUsername}`)
      navigate({ to: '/telegram-channels/$id', params: { id: channel.id } })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create channel')
    }
  }

  return (
    <Page>
      <PageHeader
        title='Add Telegram bot'
        subtitle='The token is validated with Telegram getMe before anything is saved.'
      />
      <PageScroll>
        <Card className='max-w-2xl'>
          <CardHeader>
            <CardTitle className='text-base'>Bot transport</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='space-y-2'>
              <Label htmlFor='key'>Key</Label>
              <Input
                id='key'
                value={key}
                onChange={(e) => setKey(e.target.value)}
                placeholder='ops-bot'
              />
              <p className='text-xs text-muted-foreground'>
                Immutable, unique in this workspace. Used in logs and configuration.
              </p>
            </div>
            <div className='space-y-2'>
              <Label htmlFor='name'>Display name</Label>
              <Input
                id='name'
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder='Ops bot'
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='bot-token'>Bot token</Label>
              <Input
                id='bot-token'
                type='password'
                autoComplete='off'
                value={botToken}
                onChange={(e) => setBotToken(e.target.value)}
                placeholder='123456:ABC-DEF…'
              />
              <p className='text-xs text-muted-foreground'>
                Write-only. It is encrypted at rest and never shown again — rotate it if
                you lose it.
              </p>
            </div>
            <div className='space-y-2'>
              <Label htmlFor='receive-mode'>Receive mode</Label>
              <Select
                value={String(receiveMode)}
                onValueChange={(value) => setReceiveMode(Number(value) as TelegramReceiveMode)}
              >
                <SelectTrigger id='receive-mode' aria-label='Receive mode'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={String(TelegramReceiveMode.WEBHOOK)}>
                    {RECEIVE_MODE_LABELS[TelegramReceiveMode.WEBHOOK]}
                  </SelectItem>
                  <SelectItem value={String(TelegramReceiveMode.LONG_POLLING)}>
                    {RECEIVE_MODE_LABELS[TelegramReceiveMode.LONG_POLLING]}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className='text-xs text-muted-foreground'>
                Webhook is the multi-pod default. Long polling is intended for local
                development.
              </p>
            </div>
            <div className='flex justify-end gap-2'>
              <Button variant='outline' onClick={() => navigate({ to: '/telegram-channels' })}>
                Cancel
              </Button>
              <Button
                onClick={submit}
                disabled={!key.trim() || !botToken.trim() || create.isPending}
              >
                Validate and save
              </Button>
            </div>
          </CardContent>
        </Card>
      </PageScroll>
    </Page>
  )
}
