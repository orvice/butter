import type { ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { Brain, Info } from 'lucide-react'
import { useWorkspaceMemoryConfig } from '@/api/workspace-memory'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { supportsMemoryTools, type MemoryFormValues } from './memory-config'

export interface MemoryFormErrors {
  topK?: string
  threshold?: string
}

interface MemoryConfigurationCardProps {
  value: MemoryFormValues
  onChange: (value: MemoryFormValues) => void
  agentType?: string
  errors?: MemoryFormErrors
}

export function MemoryConfigurationCard({
  value,
  onChange,
  agentType,
  errors,
}: MemoryConfigurationCardProps) {
  const { data: workspaceConfig, isSuccess } = useWorkspaceMemoryConfig()
  const workspaceReady = !!workspaceConfig?.enabled
  const toolsSupported = supportsMemoryTools(agentType)

  function set<K extends keyof MemoryFormValues>(
    key: K,
    next: MemoryFormValues[K]
  ) {
    onChange({ ...value, [key]: next })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <Brain className='size-4' />
          Memory
        </CardTitle>
        <CardDescription>
          Recall and save long-term memories through the workspace&apos;s mem0
          server. Applies when this agent is invoked directly; as a sub-agent it
          follows its root agent&apos;s settings.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {isSuccess && !workspaceReady && (
          <Alert>
            <Info className='size-4' />
            <AlertDescription>
              <span>
                This workspace has no enabled mem0 connection, so memory has no
                effect yet. An owner or admin can set one up in{' '}
                <Link
                  to='/memory'
                  className='font-medium underline underline-offset-4'
                >
                  Memory settings
                </Link>
                .
              </span>
            </AlertDescription>
          </Alert>
        )}

        <SwitchRow
          id='memory-enabled'
          label='Enable memory'
          description='Workspace Memory is shared by all members and agents of this workspace.'
          checked={value.enabled}
          onCheckedChange={(checked) => set('enabled', checked)}
        />

        {value.enabled && (
          <div className='space-y-4'>
            <SwitchRow
              id='memory-auto-recall'
              label='Auto recall'
              description='Before each turn, recall relevant memories and give them to every model call. Never saved to the session.'
              checked={value.autoRecall}
              onCheckedChange={(checked) => set('autoRecall', checked)}
            />
            <SwitchRow
              id='memory-auto-capture'
              label='Auto capture'
              description='After each turn, send the exchange to Workspace Memory for extraction. Common secrets are redacted first.'
              checked={value.autoCapture}
              onCheckedChange={(checked) => set('autoCapture', checked)}
            />
            {toolsSupported ? (
              <>
                <SwitchRow
                  id='memory-tools'
                  label='Memory tools'
                  description='Let the model call search_memory and add_memory.'
                  checked={value.tools}
                  onCheckedChange={(checked) => set('tools', checked)}
                />
                {value.tools && (
                  <SwitchRow
                    id='memory-agent-scope-write'
                    label='Allow agent-scope writes'
                    description='Let add_memory save memory specific to this agent, not only to the shared workspace pool.'
                    checked={value.agentScopeWrite}
                    onCheckedChange={(checked) =>
                      set('agentScopeWrite', checked)
                    }
                  />
                )}
              </>
            ) : (
              <p className='text-sm text-muted-foreground'>
                Memory tools are only available on LLM agents. This agent&apos;s
                LLM sub-agents still receive recalled memories.
              </p>
            )}

            <div className='grid gap-4 sm:grid-cols-2'>
              <NumberField
                id='memory-top-k'
                label='Memories per turn'
                placeholder='5'
                hint='Across both scopes, 1-50. Default 5.'
                value={value.topK}
                error={errors?.topK}
                onChange={(next) => set('topK', next)}
              />
              <NumberField
                id='memory-threshold'
                label='Relevance threshold'
                placeholder='0.3'
                hint='Minimum mem0 score, 0-1. Default 0.3.'
                value={value.threshold}
                error={errors?.threshold}
                onChange={(next) => set('threshold', next)}
              />
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function SwitchRow({
  id,
  label,
  description,
  checked,
  onCheckedChange,
}: {
  id: string
  label: string
  description: ReactNode
  checked: boolean
  onCheckedChange: (checked: boolean) => void
}) {
  return (
    <div className='flex flex-row items-center justify-between gap-4 rounded-md border p-3'>
      <div className='space-y-0.5'>
        <Label htmlFor={id}>{label}</Label>
        <p className='text-xs text-muted-foreground'>{description}</p>
      </div>
      <Switch id={id} checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  )
}

function NumberField({
  id,
  label,
  placeholder,
  hint,
  value,
  error,
  onChange,
}: {
  id: string
  label: string
  placeholder: string
  hint: string
  value: string
  error?: string
  onChange: (value: string) => void
}) {
  return (
    <div className='space-y-2'>
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        inputMode='decimal'
        placeholder={placeholder}
        value={value}
        aria-invalid={!!error}
        onChange={(e) => onChange(e.target.value)}
      />
      <p
        className={
          error ? 'text-xs text-destructive' : 'text-xs text-muted-foreground'
        }
      >
        {error ?? hint}
      </p>
    </div>
  )
}
