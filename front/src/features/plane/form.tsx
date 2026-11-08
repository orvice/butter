import { useState } from 'react'
import type { PlaneApp } from '@/gen/agents/v1/plane_pb'
import { useAgents } from '@/api/agents'
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
import { Textarea } from '@/components/ui/textarea'
import { PageActions } from '@/components/butter/page-parts'
import {
  planeRevisionKey,
  shouldReseedPlaneForm,
  toPlaneAppInput,
  toPlaneFormState,
  type PlaneAppDraft,
  type PlaneFormState,
} from './form-state'
import { PlaneInboundNotice } from './shared'

export type PlaneAppFormValues = PlaneAppDraft & {
  clientSecret: string
  webhookSecret: string
}

type Props = {
  mode: 'create' | 'edit'
  initialValue?: PlaneApp
  /** Set by the Agent-scoped entry: the Agent is fixed by the route. */
  lockedAgentId?: string
  loading?: boolean
  submitLabel: string
  onCancel: () => void
  onSubmit: (values: PlaneAppFormValues) => void
  /**
   * Bumped by the edit page after an aborted save. A dirty form otherwise
   * keeps the operator's input even when a newer revision arrives.
   */
  reloadToken?: number
}

/**
 * One form for registering and editing a Plane App.
 *
 * The client ID is immutable once registered, inbound cannot be enabled at
 * all, and secrets are only entered on create — the edit page has its own
 * write-only secret card.
 */
export function PlaneAppForm({
  mode,
  initialValue,
  lockedAgentId,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
  reloadToken = 0,
}: Props) {
  const { data: agentsData } = useAgents()
  const [form, setForm] = useState<PlaneFormState>(() =>
    toPlaneFormState(initialValue, lockedAgentId)
  )
  const [error, setError] = useState<string | null>(null)
  const [syncedKey, setSyncedKey] = useState<string | null>(
    planeRevisionKey(initialValue)
  )
  const [syncedToken, setSyncedToken] = useState(reloadToken)
  const [dirty, setDirty] = useState(false)

  // Re-seed on a newer server revision, but never over unsaved input unless
  // the page explicitly asks for it after a conflict.
  const nextKey = planeRevisionKey(initialValue)
  const forced = reloadToken !== syncedToken
  if (shouldReseedPlaneForm({ dirty, syncedKey, nextKey, force: forced })) {
    setSyncedKey(nextKey)
    setSyncedToken(reloadToken)
    setDirty(false)
    setForm(toPlaneFormState(initialValue, lockedAgentId))
    setError(null)
  }

  const agents = (agentsData?.agents ?? []).filter((agent) =>
    Boolean(agent.agent_id)
  )
  const selectedAgent = agents.find((agent) => agent.agent_id === form.agentId)

  function set<K extends keyof PlaneFormState>(
    field: K,
    value: PlaneFormState[K]
  ) {
    setDirty(true)
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!form.displayName.trim()) return setError('Display name is required')
    if (!form.clientId.trim()) return setError('Client ID is required')
    if (!(lockedAgentId ?? form.agentId)) {
      return setError('Choose the Agent this App routes to')
    }
    const rawMax = form.maxRunSeconds.trim()
    if (
      rawMax !== '' &&
      (!Number.isInteger(Number(rawMax)) || Number(rawMax) < 0)
    ) {
      return setError('Max run time must be a whole number of seconds')
    }
    setError(null)
    const draft = toPlaneAppInput(form, {
      id: initialValue?.id,
      revision: initialValue?.revision,
      lockedAgentId,
    })
    onSubmit({
      ...draft,
      clientSecret: form.clientSecret.trim(),
      webhookSecret: form.webhookSecret.trim(),
    })
  }

  return (
    <form onSubmit={submit} className='space-y-6'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Plane App</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='plane-display-name'>Display name</Label>
            <Input
              id='plane-display-name'
              value={form.displayName}
              placeholder='Support agent'
              onChange={(e) => set('displayName', e.target.value)}
            />
          </div>
          <div className='space-y-2'>
            <Label htmlFor='plane-client-id'>Client ID</Label>
            <Input
              id='plane-client-id'
              value={form.clientId}
              disabled={mode === 'edit'}
              onChange={(e) => set('clientId', e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              The public client ID of the Plane app. It must be unique across
              every workspace and cannot be changed after registration.
            </p>
          </div>
          <div className='space-y-2'>
            <Label htmlFor='plane-agent'>Agent</Label>
            <Select
              value={form.agentId || undefined}
              disabled={Boolean(lockedAgentId)}
              onValueChange={(v) => set('agentId', v)}
            >
              <SelectTrigger id='plane-agent' aria-label='Agent'>
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
              {lockedAgentId
                ? 'Fixed by the Agent entry you came from. Changing the binding is an explicit edit on the App.'
                : 'Every Plane request accepted by this App is answered by this Agent.'}
            </p>
          </div>
          {selectedAgent?.type === 'AGENT_TYPE_PI' ||
          selectedAgent?.type === 'AGENT_TYPE_CURSOR' ? (
            <p className='text-xs text-muted-foreground'>
              This Agent runs commands on a ButterBox. Once inbound ships,
              anyone admitted to this App can drive those commands.
            </p>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>Access</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='plane-allowlist'>Allowed Plane user IDs</Label>
            <Textarea
              id='plane-allowlist'
              rows={3}
              value={form.allowedUserIds}
              placeholder='One Plane user ID per line'
              onChange={(e) => set('allowedUserIds', e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              Leave empty to admit every member of the installed Plane
              workspace. Invalid IDs are rejected, not silently dropped.
            </p>
          </div>
          <div className='space-y-2'>
            <Label htmlFor='plane-max-run'>Max run time (seconds)</Label>
            <Input
              id='plane-max-run'
              inputMode='numeric'
              value={form.maxRunSeconds}
              placeholder='Use the server default'
              onChange={(e) => set('maxRunSeconds', e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              Empty follows the server default; 0 means unlimited.
            </p>
          </div>
          <PlaneInboundNotice />
        </CardContent>
      </Card>

      {mode === 'create' && (
        <Card>
          <CardHeader>
            <CardTitle className='text-base'>Secrets</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='space-y-2'>
              <Label htmlFor='plane-client-secret'>Client secret</Label>
              <Input
                id='plane-client-secret'
                type='password'
                autoComplete='new-password'
                value={form.clientSecret}
                onChange={(e) => set('clientSecret', e.target.value)}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='plane-webhook-secret'>
                Webhook signing secret
              </Label>
              <Input
                id='plane-webhook-secret'
                type='password'
                autoComplete='new-password'
                value={form.webhookSecret}
                onChange={(e) => set('webhookSecret', e.target.value)}
              />
            </div>
            <p className='text-xs text-muted-foreground'>
              Write-only: encrypted at rest and never displayed again. You can
              also set them later from the App's page.
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
