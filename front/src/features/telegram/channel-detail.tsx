import { useState } from 'react'
import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { AlertTriangle, KeyRound, Plus, Send, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  useDeleteTelegramChannel,
  useDeleteTelegramDestination,
  useRotateTelegramChannelCredential,
  useSetTelegramChannelEnabled,
  useTelegramChannel,
  useTelegramChannelStatus,
  useSendTelegramTestMessage,
  useTelegramDestinations,
  useUpdateTelegramChannel,
} from '@/api/telegram'
import { TelegramReceiveMode } from '@/gen/agents/v1/telegram_pb'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Badge } from '@/components/ui/badge'
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
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { RECEIVE_MODE_LABELS, WEBHOOK_STATE_LABELS } from './labels'
import { AddressLabel, CredentialStateBadge } from './shared'

export function TelegramChannelDetail() {
  const { id } = useParams({ from: '/_authenticated/telegram-channels/$id/' })
  const navigate = useNavigate()
  const { data: channel, isLoading } = useTelegramChannel(id)
  const { data: status } = useTelegramChannelStatus(id)
  const { data: destinations } = useTelegramDestinations(id)
  const update = useUpdateTelegramChannel()
  const rotate = useRotateTelegramChannelCredential()
  const setEnabled = useSetTelegramChannelEnabled()
  const removeChannel = useDeleteTelegramChannel()
  const removeDestination = useDeleteTelegramDestination()
  const sendTest = useSendTelegramTestMessage()

  const [name, setName] = useState('')
  const [receiveMode, setReceiveMode] = useState<TelegramReceiveMode>(
    TelegramReceiveMode.WEBHOOK
  )
  const [rotationToken, setRotationToken] = useState('')

  // Re-seed the form whenever a new server revision arrives — including the
  // one produced by our own save — so the editable fields always reflect what
  // the next optimistic update will be checked against. Adjusting state
  // during render (rather than in an effect) avoids a cascading second pass.
  const [syncedRevision, setSyncedRevision] = useState<string | null>(null)
  const revisionKey = channel ? `${channel.id}:${channel.revision}` : null
  if (channel && revisionKey !== syncedRevision) {
    setSyncedRevision(revisionKey)
    setName(channel.name)
    setReceiveMode(channel.receiveMode)
  }

  if (isLoading || !channel) {
    return (
      <Page>
        <PageHeader title='Telegram channel' />
        <PageScroll>
          <Skeleton className='h-64' />
        </PageScroll>
      </Page>
    )
  }

  async function saveSettings() {
    if (!channel) return
    try {
      await update.mutateAsync({
        id: channel.id,
        revision: channel.revision,
        name,
        receiveMode,
      })
      toast.success('Channel updated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Update failed')
    }
  }

  async function applyEnablement(inbound: boolean, outbound: boolean) {
    if (!channel) return
    try {
      const res = await setEnabled.mutateAsync({
        channelId: channel.id,
        revision: channel.revision,
        inboundEnabled: inbound,
        outboundEnabled: outbound,
      })
      res.warnings.forEach((warning) => toast.warning(warning))
      toast.success('Channel state updated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not change channel state')
    }
  }

  async function rotateCredential() {
    if (!channel) return
    try {
      await rotate.mutateAsync({ channelId: channel.id, botToken: rotationToken })
      setRotationToken('')
      toast.success('Bot token rotated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Rotation failed')
    }
  }

  const blockers = status?.blockers ?? []
  const warnings = status?.warnings ?? []

  return (
    <Page>
      <PageHeader
        title={channel.name || channel.key}
        subtitle={`@${channel.botUsername} · bot id ${channel.botId}`}
        actions={
          <Button
            variant='outline'
            onClick={async () => {
              try {
                await removeChannel.mutateAsync(channel.id)
                toast.success('Channel deleted')
                navigate({ to: '/telegram-channels' })
              } catch (err) {
                toast.error(err instanceof Error ? err.message : 'Delete failed')
              }
            }}
          >
            <Trash2 className='h-4 w-4' />
            Delete
          </Button>
        }
      />
      <PageScroll>
        <div className='space-y-6'>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-base'>Status</CardTitle>
            </CardHeader>
            <CardContent className='space-y-3 text-sm'>
              <div className='flex flex-wrap items-center gap-2'>
                <CredentialStateBadge state={channel.credentialState} />
                <Badge variant='outline'>
                  {RECEIVE_MODE_LABELS[channel.receiveMode] ?? 'Unset'}
                </Badge>
                <Badge variant='outline'>
                  {status?.inboundDestinationCount ?? 0} inbound destinations
                </Badge>
                <Badge variant='outline'>
                  {status?.outboundDestinationCount ?? 0} outbound destinations
                </Badge>
                {channel.receiveMode === TelegramReceiveMode.WEBHOOK && (
                  <Badge variant='outline' data-testid='webhook-state'>
                    {WEBHOOK_STATE_LABELS[status?.webhookState ?? 0] ?? 'Unknown'}
                  </Badge>
                )}
              </div>
              {status?.webhookUrl && (
                <p className='font-mono text-xs break-all text-muted-foreground'>
                  Callback: {status.webhookUrl}
                </p>
              )}
              {status?.lastWebhookError && (
                <p className='text-destructive' data-testid='webhook-error'>
                  {status.lastWebhookError}
                </p>
              )}
              {blockers.length > 0 && (
                <div data-testid='channel-blockers' className='space-y-1'>
                  {blockers.map((blocker) => (
                    <p
                      key={blocker}
                      className='flex items-start gap-2 text-destructive'
                    >
                      <AlertTriangle className='mt-0.5 h-3.5 w-3.5 shrink-0' />
                      {blocker}
                    </p>
                  ))}
                </div>
              )}
              {warnings.map((warning) => (
                <p key={warning} className='text-muted-foreground'>
                  {warning}
                </p>
              ))}
              <div className='flex flex-col gap-3 pt-2 sm:flex-row sm:items-center sm:gap-6'>
                <label className='flex items-center gap-2'>
                  <Switch
                    checked={channel.inboundEnabled}
                    aria-label='Inbound enabled'
                    onCheckedChange={(checked) =>
                      applyEnablement(checked, checked ? true : channel.outboundEnabled)
                    }
                  />
                  <span>Receive updates</span>
                </label>
                <label className='flex items-center gap-2'>
                  <Switch
                    checked={channel.outboundEnabled}
                    aria-label='Outbound enabled'
                    onCheckedChange={(checked) =>
                      applyEnablement(checked ? channel.inboundEnabled : false, checked)
                    }
                  />
                  <span>Send messages</span>
                </label>
              </div>
              <p className='text-xs text-muted-foreground'>
                Turning off sending also turns off receiving: every accepted interaction
                must be able to reply.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-base'>Settings</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='space-y-2'>
                <Label htmlFor='channel-key'>Key</Label>
                <Input id='channel-key' value={channel.key} disabled />
                <p className='text-xs text-muted-foreground'>Immutable.</p>
              </div>
              <div className='space-y-2'>
                <Label htmlFor='channel-name'>Display name</Label>
                <Input
                  id='channel-name'
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='channel-receive-mode'>Receive mode</Label>
                <Select
                  value={String(receiveMode)}
                  onValueChange={(value) =>
                    setReceiveMode(Number(value) as TelegramReceiveMode)
                  }
                >
                  <SelectTrigger id='channel-receive-mode' aria-label='Receive mode'>
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
              </div>
              <div className='flex justify-end'>
                <Button onClick={saveSettings} disabled={update.isPending}>
                  Save
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <KeyRound className='h-4 w-4' />
                Rotate bot token
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                The new token must resolve to the same bot ({channel.botId}). A token for
                a different bot is rejected, and the current one keeps working.
              </p>
              <Input
                type='password'
                autoComplete='off'
                aria-label='New bot token'
                value={rotationToken}
                onChange={(e) => setRotationToken(e.target.value)}
                placeholder='123456:ABC-DEF…'
              />
              <div className='flex justify-end'>
                <Button
                  variant='outline'
                  onClick={rotateCredential}
                  disabled={!rotationToken.trim() || rotate.isPending}
                >
                  Rotate
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='flex flex-row items-center justify-between pb-2'>
              <CardTitle className='text-base'>Destinations</CardTitle>
              <Button size='sm' asChild>
                <Link
                  to='/telegram-channels/$id/destinations/create'
                  params={{ id: channel.id }}
                >
                  <Plus className='h-4 w-4' />
                  Add destination
                </Link>
              </Button>
            </CardHeader>
            <CardContent className='space-y-2'>
              {!destinations?.length ? (
                <p className='py-6 text-center text-sm text-muted-foreground'>
                  No destinations yet. Send <code>/where</code> in the chat or topic you
                  want to bind, then add it here.
                </p>
              ) : (
                destinations.map((destination) => (
                  <div
                    key={destination.id}
                    data-testid={`telegram-destination-${destination.key}`}
                    className='flex flex-wrap items-center justify-between gap-2 rounded-md border p-3'
                  >
                    <div className='min-w-0 space-y-1'>
                      <div className='flex flex-wrap items-center gap-2'>
                        <span className='font-medium'>
                          {destination.name || destination.key}
                        </span>
                        {destination.inboundEnabled && (
                          <Badge variant='outline'>Inbound</Badge>
                        )}
                        {destination.outboundEnabled && (
                          <Badge variant='outline'>Outbound</Badge>
                        )}
                        {destination.verification?.verified && (
                          <Badge className='bg-success-muted text-success-foreground'>
                            Verified
                          </Badge>
                        )}
                      </div>
                      <AddressLabel destination={destination} />
                    </div>
                    <div className='flex items-center gap-2'>
                      <Button
                        variant='outline'
                        size='sm'
                        aria-label={`Send test message to ${destination.key}`}
                        disabled={!destination.outboundEnabled || sendTest.isPending}
                        onClick={async () => {
                          try {
                            await sendTest.mutateAsync({ destinationId: destination.id })
                            toast.success('Test message delivered')
                          } catch (err) {
                            toast.error(
                              err instanceof Error ? err.message : 'Test message failed'
                            )
                          }
                        }}
                      >
                        <Send className='h-4 w-4' />
                        Test
                      </Button>
                      <Button variant='outline' size='sm' asChild>
                        <Link
                          to='/telegram-destinations/$id'
                          params={{ id: destination.id }}
                        >
                          Edit
                        </Link>
                      </Button>
                      <Button
                        variant='ghost'
                        size='sm'
                        onClick={async () => {
                          try {
                            await removeDestination.mutateAsync(destination.id)
                            toast.success('Destination deleted')
                          } catch (err) {
                            toast.error(
                              err instanceof Error ? err.message : 'Delete failed'
                            )
                          }
                        }}
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </PageScroll>
    </Page>
  )
}
