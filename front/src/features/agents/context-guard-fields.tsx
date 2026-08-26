import { ShieldCheck } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  contextGuardValuesForMode,
  type ContextGuardFormValues,
  type ContextGuardMode,
} from './context-guard-config'

export interface ContextGuardFormErrors {
  mode?: string
  maxTokens?: string
  maxTurns?: string
}

interface ContextGuardConfigurationCardProps {
  value: ContextGuardFormValues
  onChange: (value: ContextGuardFormValues) => void
  errors?: ContextGuardFormErrors
}

const MODES: readonly {
  value: ContextGuardMode
  label: string
  description: string
}[] = [
  {
    value: 'off',
    label: 'Off',
    description: 'Keep the full available conversation history.',
  },
  {
    value: 'threshold',
    label: 'Token Threshold',
    description: 'Compact when the estimated context approaches its window.',
  },
  {
    value: 'sliding_window',
    label: 'Sliding Window',
    description: 'Compact after a configurable number of content entries.',
  },
]

export function ContextGuardConfigurationCard({
  value,
  onChange,
  errors,
}: ContextGuardConfigurationCardProps) {
  function changeMode(mode: string) {
    if (mode !== 'off' && mode !== 'threshold' && mode !== 'sliding_window') return
    onChange(contextGuardValuesForMode(mode, value))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <ShieldCheck className='size-4' />
          Context Guard
        </CardTitle>
        <CardDescription>
          Optional input context management for this LLM Agent. Model context capacity is not an output-token limit.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <RadioGroup
          aria-label='Context guard strategy'
          value={value.mode}
          onValueChange={changeMode}
          className='grid gap-2 md:grid-cols-3'
        >
          {MODES.map((mode) => {
            const id = `context-guard-${mode.value}`
            return (
              <Label
                key={mode.value}
                htmlFor={id}
                className={`flex cursor-pointer items-start gap-3 rounded-md border p-3 transition-colors ${
                  value.mode === mode.value ? 'border-primary bg-primary/10' : 'hover:bg-muted'
                }`}
              >
                <RadioGroupItem id={id} value={mode.value} className='mt-0.5' />
                <span className='space-y-1'>
                  <span className='block text-sm font-medium'>{mode.label}</span>
                  <span className='block text-xs text-muted-foreground'>{mode.description}</span>
                </span>
              </Label>
            )
          })}
        </RadioGroup>
        {errors?.mode && <p className='text-sm text-destructive'>{errors.mode}</p>}

        {value.mode === 'threshold' && (
          <div className='space-y-2'>
            <Label htmlFor='context-guard-max-tokens'>Context window override (tokens)</Label>
            <Input
              id='context-guard-max-tokens'
              aria-describedby='context-guard-max-tokens-help'
              type='number'
              min={0}
              max={2_147_483_647}
              step={1}
              inputMode='numeric'
              placeholder='Use model capacity'
              value={value.maxTokens}
              onChange={(event) => onChange({ ...value, maxTokens: event.target.value })}
            />
            <p id='context-guard-max-tokens-help' className='text-xs text-muted-foreground'>
              Optional Agent Context Override. Blank inherits model metadata, then the embedded registry or the 128,000-token fallback.
            </p>
            {errors?.maxTokens && <p className='text-sm text-destructive'>{errors.maxTokens}</p>}
          </div>
        )}

        {value.mode === 'sliding_window' && (
          <div className='space-y-2'>
            <Label htmlFor='context-guard-max-turns'>Maximum content entries</Label>
            <Input
              id='context-guard-max-turns'
              aria-describedby='context-guard-max-turns-help'
              type='number'
              min={0}
              max={2_147_483_647}
              step={1}
              inputMode='numeric'
              placeholder='20'
              value={value.maxTurns}
              onChange={(event) => onChange({ ...value, maxTurns: event.target.value })}
            />
            <p id='context-guard-max-turns-help' className='text-xs text-muted-foreground'>
              Blank or 0 keeps the existing default of 20 content entries. Model capacity is used only for safety checks.
            </p>
            {errors?.maxTurns && <p className='text-sm text-destructive'>{errors.maxTurns}</p>}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
