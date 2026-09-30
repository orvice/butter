import { useState } from 'react'
import type { LinearApp } from '@/gen/agents/v1/linear_pb'
import { useAgents } from '@/api/agents'
import type { LinearAppInput } from '@/api/linear'
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
import { Textarea } from '@/components/ui/textarea'
import { PageActions } from '@/components/butter/page-parts'
import { BoxAgentWarning } from './shared'
import { formatUserIds, parseUserIds } from './user-ids'

export type LinearAppFormValues = LinearAppInput & {
  clientSecret: string
  webhookSecret: string
}

type FormState = {
  displayName: string
  clientId: string
  agentId: string
  inboundEnabled: boolean
  allowedUserIds: string
  maxRunSeconds: string
  clientSecret: string
  webhookSecret: string
}

function toState(app: LinearApp | undefined): FormState {
  return {
    displayName: app?.displayName ?? '',
    clientId: app?.clientId ?? '',
    agentId: app?.agentId ?? '',
    inboundEnabled: app?.inboundEnabled ?? false,
    allowedUserIds: formatUserIds(app?.allowedUserIds),
    maxRunSeconds:
      app?.maxRunSeconds === undefined ? '' : String(app.maxRunSeconds),
    clientSecret: '',
    webhookSecret: '',
  }
}

type LinearAppFormProps = {
  mode: 'create' | 'edit'
  initialValue?: LinearApp
  loading?: boolean
  submitLabel: string
  onCancel: () => void
  onSubmit: (values: LinearAppFormValues) => void
}

/**
 * One form for registering and editing a Linear App. The client ID is
 * immutable once registered, and the secrets are write-only: they are only
 * entered here on create and managed separately on the edit page.
 */
export function LinearAppForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: LinearAppFormProps) {
  const { data: agentsData } = useAgents()
  const [form, setForm] = useState<FormState>(() => toState(initialValue))
  const [error, setError] = useState<string | null>(null)

  // Re-seed on every new server revision so the next save carries it.
  const [syncedRevision, setSyncedRevision] = useState<string | null>(
    initialValue ? `${initialValue.id}:${initialValue.revision}` : null
  )
  const revisionKey = initialValue
    ? `${initialValue.id}:${initialValue.revision}`
    : null
  if (initialValue && revisionKey !== syncedRevision) {
    setSyncedRevision(revisionKey)
    setForm(toState(initialValue))
  }

  const agents = (agentsData?.agents ?? []).filter((agent) =>
    Boolean(agent.agent_id)
  )
  const selectedAgent = agents.find((agent) => agent.agent_id === form.agentId)
  const boxAgent =
    selectedAgent?.type === 'AGENT_TYPE_PI' ||
    selectedAgent?.type === 'AGENT_TYPE_CURSOR'

  function set<K extends keyof FormState>(field: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!form.displayName.trim()) return setError('Display name is required')
    if (!form.clientId.trim()) return setError('Client ID is required')
    if (!form.agentId) return setError('Choose the Agent this App routes to')
    let maxRunSeconds: number | undefined
    if (form.maxRunSeconds.trim() !== '') {
      const parsed = Number(form.maxRunSeconds)
      if (!Number.isInteger(parsed) || parsed < 0) {
        return setError(
          'Max run time must be a whole number of seconds (0 for unlimited)'
        )
      }
      maxRunSeconds = parsed
    }
    setError(null)
    onSubmit({
      displayName: form.displayName.trim(),
      clientId: form.clientId.trim(),
      agentId: form.agentId,
      inboundEnabled: form.inboundEnabled,
      allowedUserIds: parseUserIds(form.allowedUserIds),
      maxRunSeconds,
      clientSecret: form.clientSecret.trim(),
      webhookSecret: form.webhookSecret.trim(),
    })
  }

  return (
    <form onSubmit={submit} className='space-y-6'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Linear App</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='linear-display-name'>Display name</Label>
            <Input
              id='linear-display-name'
              value={form.displayName}
              placeholder='Support agent'
              onChange={(e) => set('displayName', e.target.value)}
            />
          </div>
          <div className='space-y-2'>
            <Label htmlFor='linear-client-id'>Client ID</Label>
            <Input
              id='linear-client-id'
              value={form.clientId}
              disabled={mode === 'edit'}
              onChange={(e) => set('clientId', e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              From the OAuth app in Linear (Settings → API → OAuth
              applications). It cannot be changed after registration.
            </p>
          </div>
          <div className='space-y-2'>
            <Label htmlFor='linear-agent'>Agent</Label>
            <Select
              value={form.agentId || undefined}
              onValueChange={(v) => set('agentId', v)}
            >
              <SelectTrigger id='linear-agent' aria-label='Agent'>
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
              Every Linear Agent Session opened on this App is answered by this
              Agent.
            </p>
          </div>
          {boxAgent && <BoxAgentWarning />}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Access</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='linear-allowlist'>Allowed Linear user IDs</Label>
            <Textarea
              id='linear-allowlist'
              rows={3}
              value={form.allowedUserIds}
              placeholder='One Linear user ID per line'
              onChange={(e) => set('allowedUserIds', e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              Leave empty to admit every member of the installed Linear
              organizations.
            </p>
          </div>
          <div className='space-y-2'>
            <Label htmlFor='linear-max-run'>Max run time (seconds)</Label>
            <Input
              id='linear-max-run'
              inputMode='numeric'
              value={form.maxRunSeconds}
              placeholder='1800'
              onChange={(e) => set('maxRunSeconds', e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              Empty uses the default of 1800 seconds; 0 means unlimited.
            </p>
          </div>
          <div className='flex flex-row items-center justify-between rounded-md border p-3'>
            <div className='space-y-0.5'>
              <Label htmlFor='linear-inbound'>Inbound enabled</Label>
              <p className='text-xs text-muted-foreground'>
                Accept Linear deliveries. Requires both secrets.
              </p>
            </div>
            <Switch
              id='linear-inbound'
              checked={form.inboundEnabled}
              onCheckedChange={(v) => set('inboundEnabled', v)}
            />
          </div>
        </CardContent>
      </Card>

      {mode === 'create' && (
        <Card>
          <CardHeader>
            <CardTitle className='text-base'>Secrets</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='space-y-2'>
              <Label htmlFor='linear-client-secret'>Client secret</Label>
              <Input
                id='linear-client-secret'
                type='password'
                autoComplete='new-password'
                value={form.clientSecret}
                onChange={(e) => set('clientSecret', e.target.value)}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='linear-webhook-secret'>
                Webhook signing secret
              </Label>
              <Input
                id='linear-webhook-secret'
                type='password'
                autoComplete='new-password'
                value={form.webhookSecret}
                onChange={(e) => set('webhookSecret', e.target.value)}
              />
            </div>
            <p className='text-xs text-muted-foreground'>
              Write-only: encrypted at rest and never displayed again. You can
              also set them later.
            </p>
          </CardContent>
        </Card>
      )}

      {error && <p className='text-sm text-destructive'>{error}</p>}
      <PageActions>
        <Button type='button' variant='outline' onClick={onCancel}>
          Cancel
        </Button>
        <Button type='submit' disabled={loading}>
          {loading ? 'Saving...' : submitLabel}
        </Button>
      </PageActions>
    </form>
  )
}
