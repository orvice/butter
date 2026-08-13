import { Link } from '@tanstack/react-router'
import { Plus, Send } from 'lucide-react'
import { useTelegramChannels } from '@/api/telegram'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { RECEIVE_MODE_LABELS } from './labels'
import { CredentialStateBadge } from './shared'

export function TelegramChannelList() {
  const { data: channels, isLoading } = useTelegramChannels()

  return (
    <Page>
      <PageHeader
        title='Telegram Channels'
        subtitle='Each channel is one Telegram bot transport. Addresses live in its destinations.'
        actions={
          <Button asChild>
            <Link to='/telegram-channels/create'>
              <Plus className='h-4 w-4' />
              Add bot
            </Link>
          </Button>
        }
      />
      <PageScroll>
        {isLoading ? (
          <div className='grid gap-4 md:grid-cols-2'>
            <Skeleton className='h-32' />
            <Skeleton className='h-32' />
          </div>
        ) : !channels?.length ? (
          <Card>
            <CardContent className='py-10 text-center text-sm text-muted-foreground'>
              No Telegram channels yet. Add a bot token to register one.
            </CardContent>
          </Card>
        ) : (
          <div className='grid gap-4 md:grid-cols-2'>
            {channels.map((channel) => (
              <Card key={channel.id} data-testid={`telegram-channel-${channel.key}`}>
                <CardHeader className='flex flex-row items-start justify-between gap-2 pb-2'>
                  <div className='flex min-w-0 flex-wrap items-center gap-2'>
                    <Send className='h-4 w-4 text-sky-500' />
                    <CardTitle className='text-base'>{channel.name || channel.key}</CardTitle>
                    {channel.inboundEnabled && <Badge variant='outline'>Inbound</Badge>}
                    {channel.outboundEnabled && <Badge variant='outline'>Outbound</Badge>}
                    {!channel.inboundEnabled && !channel.outboundEnabled && (
                      <Badge className='bg-muted text-muted-foreground'>Disabled</Badge>
                    )}
                  </div>
                  <Button variant='outline' size='sm' asChild>
                    <Link to='/telegram-channels/$id' params={{ id: channel.id }}>
                      Manage
                    </Link>
                  </Button>
                </CardHeader>
                <CardContent className='space-y-1 text-sm text-muted-foreground'>
                  <div className='font-mono text-xs'>
                    {channel.botUsername ? `@${channel.botUsername}` : 'unknown bot'} · id{' '}
                    {channel.botId}
                  </div>
                  <div className='flex flex-wrap items-center gap-2 pt-1'>
                    <Badge variant='outline'>
                      {RECEIVE_MODE_LABELS[channel.receiveMode] ?? 'Unset'}
                    </Badge>
                    <CredentialStateBadge state={channel.credentialState} />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </PageScroll>
    </Page>
  )
}
