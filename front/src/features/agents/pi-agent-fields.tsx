import { useId } from 'react'
import { Info } from 'lucide-react'
import { useButterBoxes, useButterBoxModels } from '@/api/butterboxes'
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
import {
  piModelCatalogKey,
  piModelFromCatalogKey,
  type PiAgentFormValues,
} from './pi-config'

const THINKING_LEVELS = [
  'off',
  'minimal',
  'low',
  'medium',
  'high',
  'xhigh',
  'max',
] as const
const DEFAULT_THINKING_LEVEL = '__default__'
const CUSTOM_MODEL = '__custom_model__'

export type PiAgentFieldsProps = {
  value: PiAgentFormValues
  onChange: <K extends keyof PiAgentFormValues>(
    field: K,
    value: PiAgentFormValues[K]
  ) => void
  errors?: Partial<Record<keyof PiAgentFormValues, string>>
}

export function PiAgentFields({
  value,
  onChange,
  errors = {},
}: PiAgentFieldsProps) {
  const id = useId()
  const boxesQuery = useButterBoxes()
  const modelsQuery = useButterBoxModels(
    value.butterboxId,
    Boolean(value.butterboxId)
  )
  const boxes = boxesQuery.data ?? []
  const models = modelsQuery.data ?? []
  const boxErrorId = `${id}-butterbox-error`
  const maxRunErrorId = `${id}-max-run-error`
  const modelStatusId = `${id}-model-status`

  function selectButterBox(butterboxId: string) {
    onChange('butterboxId', butterboxId)
    onChange('provider', '')
    onChange('model', '')
  }

  function selectModel(selection: string) {
    const catalogModel = piModelFromCatalogKey(models, selection)
    onChange('model', catalogModel?.id ?? '')
    onChange('provider', catalogModel?.provider ?? '')
  }

  const selectedCatalogModel = models.find(
    (model) => model.id === value.model && model.provider === value.provider
  )
  const modelSelection = selectedCatalogModel
    ? piModelCatalogKey(selectedCatalogModel)
    : CUSTOM_MODEL
  const usesCustomModel = modelSelection === CUSTOM_MODEL

  let modelStatus = ''
  if (!value.butterboxId) modelStatus = 'Select a ButterBox to load its models.'
  else if (modelsQuery.isFetching) modelStatus = 'Loading models...'
  else if (modelsQuery.isError)
    modelStatus = 'Unable to load models. Enter a model ID manually.'
  else if (models.length === 0)
    modelStatus = 'No models reported. Enter a model ID manually.'
  else if (value.provider) modelStatus = `Provider: ${value.provider}`

  return (
    <div className='space-y-5'>
      <div className='flex items-start gap-2 text-sm text-muted-foreground'>
        <Info aria-hidden='true' className='mt-0.5 size-4 shrink-0' />
        <p>
          Butter chooses where this Agent runs. Pi manages its tools and
          instructions on the box, using the working directory&apos;s AGENTS.md
          and .pi configuration.
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
        {models.length > 0 ? (
          <Select value={modelSelection} onValueChange={selectModel}>
            <SelectTrigger
              id={`${id}-model`}
              disabled={!value.butterboxId}
              aria-describedby={modelStatusId}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {models.map((model) => (
                <SelectItem
                  key={piModelCatalogKey(model)}
                  value={piModelCatalogKey(model)}
                >
                  {model.name || model.id} ({model.provider} / {model.id})
                </SelectItem>
              ))}
              <SelectItem value={CUSTOM_MODEL}>Custom model ID...</SelectItem>
            </SelectContent>
          </Select>
        ) : null}
        {(models.length === 0 || usesCustomModel) && (
          <Input
            id={models.length === 0 ? `${id}-model` : `${id}-custom-model`}
            value={value.model}
            onChange={(event) => {
              onChange('model', event.target.value)
              onChange('provider', '')
            }}
            placeholder='Model ID'
            autoComplete='off'
            spellCheck={false}
            disabled={!value.butterboxId}
            aria-label={models.length > 0 ? 'Custom model ID' : undefined}
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
          <Label htmlFor={`${id}-thinking-level`}>Thinking level</Label>
          <Select
            value={value.thinkingLevel || DEFAULT_THINKING_LEVEL}
            onValueChange={(next) =>
              onChange(
                'thinkingLevel',
                next === DEFAULT_THINKING_LEVEL ? '' : next
              )
            }
          >
            <SelectTrigger id={`${id}-thinking-level`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={DEFAULT_THINKING_LEVEL}>
                Box default
              </SelectItem>
              {THINKING_LEVELS.map((level) => (
                <SelectItem key={level} value={level}>
                  {level}
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

export function PiAgentConfigurationCard(props: PiAgentFieldsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Pi configuration</CardTitle>
        <CardDescription>
          Bind this Agent to a ButterBox and its pi runtime.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <PiAgentFields {...props} />
      </CardContent>
    </Card>
  )
}
