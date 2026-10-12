import { useMemo } from 'react'
import type { Agent, ModelProvider } from '@/types/api'
import { useModelProviders } from '@/api/model-providers'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const DEFAULT_MODEL = '__agent_default__'

export function supportsChatModelOverride(agent: Agent | null | undefined) {
  return (
    !!agent &&
    (!agent.type ||
      agent.type === 'AGENT_TYPE_UNSPECIFIED' ||
      agent.type === 'AGENT_TYPE_LLM')
  )
}

export function chatModelOptions(providers: ModelProvider[] = []) {
  const options = new Map<
    string,
    { value: string; label: string; provider: string }
  >()
  for (const provider of providers) {
    for (const model of provider.models ?? []) {
      const value = (model.alias || model.name || '').trim()
      if (value && !options.has(value)) {
        options.set(value, {
          value,
          label: model.alias || model.name,
          provider: provider.name,
        })
      }
    }
  }
  return [...options.values()].sort((a, b) => a.label.localeCompare(b.label))
}

export function ChatModelSelector({
  value,
  onChange,
  disabled = false,
}: {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}) {
  const { data, isLoading, isError } = useModelProviders()
  const options = useMemo(
    () => chatModelOptions(data?.model_providers),
    [data?.model_providers]
  )
  return (
    <Select
      value={value ? `model:${value}` : DEFAULT_MODEL}
      onValueChange={(next) => {
        if (next)
          onChange(next === DEFAULT_MODEL ? '' : next.slice('model:'.length))
      }}
      disabled={disabled || isLoading || isError}
    >
      <SelectTrigger aria-label='Chat model' className='w-64'>
        <SelectValue
          placeholder={isLoading ? 'Loading models…' : 'Agent default model'}
        />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={DEFAULT_MODEL}>Agent default model</SelectItem>
        {value && !options.some((option) => option.value === value) && (
          <SelectItem value={`model:${value}`} disabled>
            {value} (unavailable)
          </SelectItem>
        )}
        {options.map((option) => (
          <SelectItem key={option.value} value={`model:${option.value}`}>
            {option.label} · {option.provider}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
