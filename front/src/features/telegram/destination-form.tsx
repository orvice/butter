import { useState } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { useModelProviders } from '@/api/model-providers'
import {
  useCreateTelegramDestination,
  useTelegramDestination,
  useUpdateTelegramDestination,
} from '@/api/telegram'
import {
  TelegramReplyMode,
  TelegramSessionPolicy,
  TelegramTriggerMode,
} from '@/gen/agents/v1/telegram_pb'
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
import { Switch } from '@/components/ui/switch'
import {
  REPLY_MODE_OPTIONS,
  SESSION_POLICY_OPTIONS,
  TRIGGER_MODE_OPTIONS,
  formatIdList,
  parseIdList,
} from './labels'

type FormState = {
  key: string
  name: string
  chatId: string
  messageThreadId: string
  inboundEnabled: boolean
  outboundEnabled: boolean
  agentId: string
  model: string
  selectableAgentIds: string
  selectableModels: string
  triggerMode: TelegramTriggerMode
  sessionPolicy: TelegramSessionPolicy
  allowedUserIds: string
  controllerUserIds: string
  replyMode: TelegramReplyMode
  debugDefault: boolean
}

const EMPTY: FormState = {
  key: '',
  name: '',
  chatId: '',
  messageThreadId: '',
  inboundEnabled: true,
  outboundEnabled: true,
  agentId: '',
  model: '',
  selectableAgentIds: '',
  selectableModels: '',
  triggerMode: TelegramTriggerMode.ALL,
  sessionPolicy: TelegramSessionPolicy.DESTINATION,
  allowedUserIds: '',
  controllerUserIds: '',
  replyMode: TelegramReplyMode.REPLY,
  debugDefault: true,
}

/**
 * One form for creating and editing a Destination. The address fields are
 * disabled in edit mode because they are immutable server-side: a Cron job or
 * Notify Group already persists this Destination's ID, so changing where it
 * points would silently redirect them.
 */
export function TelegramDestinationForm({ mode }: { mode: 'create' | 'edit' }) {
  const navigate = useNavigate()
  const params = useParams({ strict: false })
  const destinationId = mode === 'edit' ? (params.id as string) : undefined
  const channelIdParam = mode === 'create' ? (params.id as string) : undefined

  const { data: existing } = useTelegramDestination(destinationId)
  const { data: agentsData } = useAgents()
  const { data: providersData } = useModelProviders()
  const create = useCreateTelegramDestination()
  const update = useUpdateTelegramDestination()

  const [form, setForm] = useState<FormState>(EMPTY)

  // Re-seed from the server on every new revision — including the one our own
  // save produced — so the next update carries a current revision. Adjusting
  // state during render rather than in an effect avoids a cascading pass.
  const [syncedRevision, setSyncedRevision] = useState<string | null>(null)
  const revisionKey = existing ? `${existing.id}:${existing.revision}` : null
  if (existing && revisionKey !== syncedRevision) {
    setSyncedRevision(revisionKey)
    const config = existing.config
    setForm({
      key: existing.key,
      name: existing.name,
      chatId: existing.chatId,
      messageThreadId: existing.messageThreadId,
      inboundEnabled: existing.inboundEnabled,
      outboundEnabled: existing.outboundEnabled,
      agentId: config?.agentId ?? '',
      model: config?.model ?? '',
      selectableAgentIds: formatIdList(config?.selectableAgentIds),
      selectableModels: formatIdList(config?.selectableModels),
      triggerMode: config?.triggerMode ?? TelegramTriggerMode.ALL,
      sessionPolicy: config?.sessionPolicy ?? TelegramSessionPolicy.DESTINATION,
      allowedUserIds: formatIdList(config?.allowedUserIds),
      controllerUserIds: formatIdList(config?.controllerUserIds),
      replyMode: config?.replyMode ?? TelegramReplyMode.REPLY,
      debugDefault: config?.debugDefault ?? true,
    })
  }

  const agents = (agentsData?.agents ?? []).filter((agent) => Boolean(agent.agent_id))
  const selectedAgent = agents.find((agent) => agent.agent_id === form.agentId)
  const selectedAgentIsBoxBacked =
    selectedAgent?.type === 'AGENT_TYPE_PI' ||
    selectedAgent?.type === 'AGENT_TYPE_CURSOR'
  const modelAliases = (providersData?.model_providers ?? []).flatMap((provider) =>
    (provider.models ?? []).map((model) => model.alias || model.name)
  )

  function set<K extends keyof FormState>(field: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  function setAgent(agentId: string) {
    const agent = agents.find((item) => item.agent_id === agentId)
    const isBoxBacked =
      agent?.type === 'AGENT_TYPE_PI' || agent?.type === 'AGENT_TYPE_CURSOR'
    setForm((prev) => ({
      ...prev,
      agentId,
      ...(isBoxBacked ? { model: '', selectableModels: '' } : {}),
    }))
  }

  async function submit() {
    const config = {
      agentId: form.agentId,
      model: selectedAgentIsBoxBacked ? '' : form.model,
      selectableAgentIds: parseIdList(form.selectableAgentIds),
      selectableModels: selectedAgentIsBoxBacked ? [] : parseIdList(form.selectableModels),
      triggerMode: form.triggerMode,
      sessionPolicy: form.sessionPolicy,
      allowedUserIds: parseIdList(form.allowedUserIds),
      controllerUserIds: parseIdList(form.controllerUserIds),
      replyMode: form.replyMode,
      debugDefault: form.debugDefault,
    }
    try {
      if (mode === 'create') {
        const created = await create.mutateAsync({
          key: form.key,
          name: form.name,
          channelId: channelIdParam,
          chatId: form.chatId,
          messageThreadId: form.messageThreadId,
          inboundEnabled: form.inboundEnabled,
          outboundEnabled: form.outboundEnabled,
          config,
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } as any)
        toast.success('Destination created')
        navigate({ to: '/telegram-destinations/$id', params: { id: created.id } })
      } else {
        await update.mutateAsync({
          id: destinationId,
          revision: existing?.revision,
          name: form.name,
          inboundEnabled: form.inboundEnabled,
          outboundEnabled: form.outboundEnabled,
          config,
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } as any)
        toast.success('Destination updated')
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Save failed')
    }
  }

  const addressDisabled = mode === 'edit'

  return (
    <Page>
      <PageHeader
        title={mode === 'create' ? 'Add destination' : form.name || form.key}
        subtitle='One exact Telegram address: a private chat, a group, or a single forum topic.'
      />
      <PageScroll>
        <div className='max-w-3xl space-y-6'>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-base'>Address</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='space-y-2'>
                <Label htmlFor='destination-key'>Key</Label>
                <Input
                  id='destination-key'
                  value={form.key}
                  disabled={addressDisabled}
                  onChange={(e) => set('key', e.target.value)}
                  placeholder='ops-alerts'
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='destination-name'>Display name</Label>
                <Input
                  id='destination-name'
                  value={form.name}
                  onChange={(e) => set('name', e.target.value)}
                />
              </div>
              <div className='grid gap-4 sm:grid-cols-2'>
                <div className='space-y-2'>
                  <Label htmlFor='destination-chat-id'>Chat ID</Label>
                  <Input
                    id='destination-chat-id'
                    value={form.chatId}
                    disabled={addressDisabled}
                    onChange={(e) => set('chatId', e.target.value)}
                    placeholder='-1001234567890'
                  />
                </div>
                <div className='space-y-2'>
                  <Label htmlFor='destination-thread-id'>Topic ID (optional)</Label>
                  <Input
                    id='destination-thread-id'
                    value={form.messageThreadId}
                    disabled={addressDisabled}
                    onChange={(e) => set('messageThreadId', e.target.value)}
                    placeholder='leave empty for a non-topic chat'
                  />
                </div>
              </div>
              <p className='text-xs text-muted-foreground'>
                Run <code>/where</code> in the target chat or topic to read these
                identifiers.
                {addressDisabled &&
                  ' The address is immutable — create a new destination to point somewhere else.'}
              </p>
              <div className='flex flex-col gap-3 sm:flex-row sm:gap-6'>
                <label className='flex items-center gap-2'>
                  <Switch
                    checked={form.inboundEnabled}
                    aria-label='Inbound enabled'
                    onCheckedChange={(checked) => set('inboundEnabled', checked)}
                  />
                  <span className='text-sm'>Handle incoming messages</span>
                </label>
                <label className='flex items-center gap-2'>
                  <Switch
                    checked={form.outboundEnabled}
                    aria-label='Outbound enabled'
                    onCheckedChange={(checked) => set('outboundEnabled', checked)}
                  />
                  <span className='text-sm'>Allow proactive sends</span>
                </label>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-base'>Routing</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='space-y-2'>
                <Label htmlFor='destination-agent'>Default agent</Label>
                <Select
                  value={form.agentId || undefined}
                  onValueChange={setAgent}
                >
                  <SelectTrigger id='destination-agent' aria-label='Default agent'>
                    <SelectValue placeholder='Select an agent' />
                  </SelectTrigger>
                  <SelectContent>
                    {agents.map((agent) => (
                      <SelectItem key={agent.agent_id} value={agent.agent_id!}>
                        {agent.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className='text-xs text-muted-foreground'>
                  Required when incoming messages are handled.
                </p>
              </div>
              <div className='space-y-2'>
                <Label htmlFor='destination-model'>Model override (optional)</Label>
                <Select
                  value={form.model || undefined}
                  onValueChange={(value) => set('model', value)}
                  disabled={selectedAgentIsBoxBacked}
                >
                  <SelectTrigger id='destination-model' aria-label='Model override'>
                    <SelectValue placeholder="Inherit the agent's model" />
                  </SelectTrigger>
                  <SelectContent>
                    {modelAliases.map((alias) => (
                      <SelectItem key={alias} value={alias}>
                        {alias}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='grid gap-4 sm:grid-cols-2'>
                <div className='space-y-2'>
                  <Label htmlFor='destination-selectable-agents'>
                    Selectable agents
                  </Label>
                  <Input
                    id='destination-selectable-agents'
                    value={form.selectableAgentIds}
                    onChange={(e) => set('selectableAgentIds', e.target.value)}
                    placeholder='leave empty to lock the agent'
                  />
                </div>
                <div className='space-y-2'>
                  <Label htmlFor='destination-selectable-models'>
                    Selectable models
                  </Label>
                  <Input
                    id='destination-selectable-models'
                    value={form.selectableModels}
                    onChange={(e) => set('selectableModels', e.target.value)}
                    placeholder='leave empty to lock the model'
                    disabled={selectedAgentIsBoxBacked}
                  />
                </div>
              </div>
              {selectedAgentIsBoxBacked && (
                <p className='text-sm text-muted-foreground'>
                  {selectedAgent?.type === 'AGENT_TYPE_PI'
                    ? 'Pi uses the model in its ButterBox binding, so Telegram model switching is locked while this Agent is active.'
                    : 'Cursor uses the model in its ButterBox binding, so Telegram model switching is locked while this Agent is active.'}
                </p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-base'>Interaction policy</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='space-y-2'>
                <Label htmlFor='destination-trigger'>Trigger</Label>
                <Select
                  value={String(form.triggerMode)}
                  onValueChange={(value) =>
                    set('triggerMode', Number(value) as TelegramTriggerMode)
                  }
                >
                  <SelectTrigger id='destination-trigger' aria-label='Trigger'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {TRIGGER_MODE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={String(option.value)}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='space-y-2'>
                <Label htmlFor='destination-session'>Conversation history</Label>
                <Select
                  value={String(form.sessionPolicy)}
                  onValueChange={(value) =>
                    set('sessionPolicy', Number(value) as TelegramSessionPolicy)
                  }
                >
                  <SelectTrigger id='destination-session' aria-label='Conversation history'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {SESSION_POLICY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={String(option.value)}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='space-y-2'>
                <Label htmlFor='destination-reply'>Reply style</Label>
                <Select
                  value={String(form.replyMode)}
                  onValueChange={(value) =>
                    set('replyMode', Number(value) as TelegramReplyMode)
                  }
                >
                  <SelectTrigger id='destination-reply' aria-label='Reply style'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {REPLY_MODE_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={String(option.value)}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='grid gap-4 sm:grid-cols-2'>
                <div className='space-y-2'>
                  <Label htmlFor='destination-allowed-users'>Allowed user IDs</Label>
                  <Input
                    id='destination-allowed-users'
                    value={form.allowedUserIds}
                    onChange={(e) => set('allowedUserIds', e.target.value)}
                    placeholder='empty admits everyone at this address'
                  />
                </div>
                <div className='space-y-2'>
                  <Label htmlFor='destination-controller-users'>
                    Controller user IDs
                  </Label>
                  <Input
                    id='destination-controller-users'
                    value={form.controllerUserIds}
                    onChange={(e) => set('controllerUserIds', e.target.value)}
                    placeholder='may switch agent/model, toggle debug'
                  />
                </div>
              </div>
              <p className='text-xs text-muted-foreground'>
                Controllers must also appear in the allowed list when one is set.
              </p>
              <label className='flex items-center gap-2'>
                <Switch
                  checked={form.debugDefault}
                  aria-label='Debug by default'
                  onCheckedChange={(checked) => set('debugDefault', checked)}
                />
                <span className='text-sm'>Start new sessions with debug output</span>
              </label>
            </CardContent>
          </Card>

          <div className='flex justify-end gap-2'>
            <Button variant='outline' onClick={() => navigate({ to: '/telegram-channels' })}>
              Cancel
            </Button>
            <Button onClick={submit} disabled={create.isPending || update.isPending}>
              Save
            </Button>
          </div>
        </div>
      </PageScroll>
    </Page>
  )
}
