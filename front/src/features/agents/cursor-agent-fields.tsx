import { useId, useState } from 'react'
import { Info } from 'lucide-react'
import { useButterBoxCursorModels, useButterBoxes } from '@/api/butterboxes'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { CursorAgentMode } from '@/types/api'
import { CURSOR_AGENT_MODES, type CursorAgentFormValues } from './cursor-config'

const DEFAULT_MODEL = '__default_model__'
const CUSTOM_MODEL = '__custom_model__'

const MODE_LABELS: Record<CursorAgentMode, string> = {
  agent: 'Agent (default)',
  plan: 'Plan',
}

export type CursorAgentFieldsProps = {
  value: CursorAgentFormValues
  onChange: <K extends keyof CursorAgentFormValues>(
    field: K,
    value: CursorAgentFormValues[K]
  ) => void
  errors?: Partial<Record<keyof CursorAgentFormValues, string>>
}

export function CursorAgentFields({
  value,
  onChange,
  errors = {},
}: CursorAgentFieldsProps) {
  const id = useId()
  const boxesQuery = useButterBoxes()
  const modelsQuery = useButterBoxCursorModels(
    value.butterboxId,
    Boolean(value.butterboxId)
  )
  const boxes = boxesQuery.data ?? []
  const models = modelsQuery.data ?? []
  const boxErrorId = `${id}-butterbox-error`
  const maxRunErrorId = `${id}-max-run-error`
  const modelStatusId = `${id}-model-status`

  // Picking "Custom model ID..." must show the text input even while the
  // field is still empty or holds a catalog ID, so the choice is local state.
  const [customRequested, setCustomRequested] = useState(false)

  function selectButterBox(butterboxId: string) {
    onChange('butterboxId', butterboxId)
    onChange('model', '')
    setCustomRequested(false)
  }

  function selectModel(selection: string) {
    setCustomRequested(selection === CUSTOM_MODEL)
    if (selection === DEFAULT_MODEL) onChange('model', '')
    else if (selection !== CUSTOM_MODEL) onChange('model', selection)
  }

  let modelSelection = CUSTOM_MODEL
  if (!customRequested) {
    if (value.model === '') modelSelection = DEFAULT_MODEL
    else if (models.some((model) => model.id === value.model))
      modelSelection = value.model
  }
  const usesCustomModel = modelSelection === CUSTOM_MODEL

  let modelStatus = ''
  if (!value.butterboxId) modelStatus = 'Select a ButterBox to load its Cursor models.'
  else if (modelsQuery.isFetching) modelStatus = 'Loading Cursor models...'
  else if (modelsQuery.isError)
    modelStatus = `Unable to load Cursor models (${modelsQuery.error.message}). Enter a model ID manually.`
  else if (models.length === 0)
    modelStatus = 'No models reported. Enter a model ID manually.'

  return (
    <div className='space-y-5'>
      <div className='flex items-start gap-2 text-sm text-muted-foreground'>
        <Info aria-hidden='true' className='mt-0.5 size-4 shrink-0' />
        <p>
          Cursor agents are configured on the box via .cursor/rules and
          mcp.json. Butter chooses where this Agent runs; the Cursor API key
          lives in the box&apos;s CURSOR_API_KEY environment variable.
        </p>
      </div>

      <div className='space-y-2'>
        <Label htmlFor={`${id}-butterbox`}>ButterBox</Label>
        <Select
          value={value.butterboxId || undefined}
          onValueChange={selectButterBox}
          disabled={boxesQuery.isLoading || boxes.length === 0}
        >
          <SelectTrigger
            id={`${id}-butterbox`}
            aria-invalid={Boolean(errors.butterboxId)}
            aria-describedby={errors.butterboxId ? boxErrorId : undefined}
          >
            <SelectValue
              placeholder={
                boxesQuery.isLoading
                  ? 'Loading ButterBoxes...'
                  : 'Select a ButterBox'
              }
            />
          </SelectTrigger>
          <SelectContent>
            {boxes.map((box) => (
              <SelectItem key={box.id} value={box.id} disabled={!box.enabled}>
                {box.name}
                {box.enabled ? '' : ' (disabled)'}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {boxesQuery.isError && (
          <p className='text-sm text-destructive'>
            Unable to load ButterBoxes. Try again.
          </p>
        )}
        {!boxesQuery.isLoading && !boxesQuery.isError && boxes.length === 0 && (
          <p className='text-sm text-muted-foreground'>
            No ButterBoxes are available.
          </p>
        )}
        {errors.butterboxId && (
          <p id={boxErrorId} className='text-sm text-destructive'>
            {errors.butterboxId}
          </p>
        )}
      </div>

      <div className='space-y-2'>
        <Label htmlFor={`${id}-working-dir`}>Working directory</Label>
        <Input
          id={`${id}-working-dir`}
          value={value.workingDir}
          onChange={(event) => onChange('workingDir', event.target.value)}
          placeholder='/workspace/project'
          autoComplete='off'
          spellCheck={false}
        />
      </div>

      <div className='space-y-2'>
        <Label htmlFor={`${id}-model`}>Model</Label>
        <Select
          value={modelSelection}
          onValueChange={selectModel}
          disabled={!value.butterboxId}
        >
          <SelectTrigger id={`${id}-model`} aria-describedby={modelStatusId}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={DEFAULT_MODEL}>Cursor default</SelectItem>
            {models.map((model) => (
              <SelectItem key={model.id} value={model.id}>
                {model.name && model.name !== model.id
                  ? `${model.name} (${model.id})`
                  : model.id}
              </SelectItem>
            ))}
            <SelectItem value={CUSTOM_MODEL}>Custom model ID...</SelectItem>
          </SelectContent>
        </Select>
        {usesCustomModel && (
          <Input
            id={`${id}-custom-model`}
            value={value.model}
            onChange={(event) => onChange('model', event.target.value)}
            placeholder='Model ID, e.g. composer-2.5'
            autoComplete='off'
            spellCheck={false}
            aria-label='Custom model ID'
            aria-describedby={modelStatusId}
          />
        )}
        <p
          id={modelStatusId}
          role='status'
          className='min-h-5 text-xs text-muted-foreground'
        >
          {modelStatus}
        </p>
      </div>

      <div className='grid gap-4 sm:grid-cols-2'>
        <div className='space-y-2'>
          <Label htmlFor={`${id}-mode`}>Mode</Label>
          <Select
            value={value.mode}
            onValueChange={(next) => onChange('mode', next as CursorAgentMode)}
          >
            <SelectTrigger id={`${id}-mode`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {CURSOR_AGENT_MODES.map((mode) => (
                <SelectItem key={mode} value={mode}>
                  {MODE_LABELS[mode]}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className='space-y-2'>
          <Label htmlFor={`${id}-max-run-seconds`}>
            Maximum run time (seconds)
          </Label>
          <Input
            id={`${id}-max-run-seconds`}
            type='number'
            min={0}
            step={1}
            value={value.maxRunSeconds}
            onChange={(event) => onChange('maxRunSeconds', event.target.value)}
            placeholder='1800'
            aria-invalid={Boolean(errors.maxRunSeconds)}
            aria-describedby={errors.maxRunSeconds ? maxRunErrorId : undefined}
          />
          {errors.maxRunSeconds && (
            <p id={maxRunErrorId} className='text-sm text-destructive'>
              {errors.maxRunSeconds}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

export function CursorAgentConfigurationCard(props: CursorAgentFieldsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Cursor configuration</CardTitle>
        <CardDescription>
          Bind this Agent to a ButterBox and its Cursor SDK Bridge.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <CursorAgentFields {...props} />
      </CardContent>
    </Card>
  )
}
