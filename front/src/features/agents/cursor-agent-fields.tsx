import { useId } from 'react'
import { Info } from 'lucide-react'
import { useButterBoxes, useCursorModels } from '@/api/butterboxes'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { CursorAgentFormValues } from './cursor-config'

const DEFAULT_MODE = '__default__'
const CUSTOM_MODEL = '__custom_model__'

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
  const modelsQuery = useCursorModels(
    value.butterboxId,
    Boolean(value.butterboxId)
  )
  const boxes = boxesQuery.data ?? []
  const models = modelsQuery.data ?? []
  const boxErrorId = `${id}-butterbox-error`
  const maxRunErrorId = `${id}-max-run-error`
  const modeErrorId = `${id}-mode-error`
  const modelStatusId = `${id}-model-status`

  function selectModel(selection: string) {
    onChange('model', selection === CUSTOM_MODEL ? '' : selection)
  }

  const usesCatalogModel = models.some((model) => model.id === value.model)
  const modelSelection = usesCatalogModel ? value.model : CUSTOM_MODEL

  let modelStatus = ''
  if (!value.butterboxId) modelStatus = 'Select a ButterBox to load its models.'
  else if (modelsQuery.isFetching) modelStatus = 'Loading models...'
  else if (modelsQuery.isError)
    modelStatus = 'Unable to load models. Enter a model ID manually.'
  else if (models.length === 0)
    modelStatus = 'No models reported. Enter a model ID manually.'
  else if (value.model && !usesCatalogModel)
    modelStatus = 'Custom model ID — not in the box catalog.'

  return (
    <div className='space-y-5'>
      <div className='flex items-start gap-2 text-sm text-muted-foreground'>
        <Info aria-hidden='true' className='mt-0.5 size-4 shrink-0' />
        <p>
          Butter chooses where this Agent runs. Cursor manages its tools and
          instructions on the box via <code>.cursor/rules</code> and{' '}
          <code>mcp.json</code>.
        </p>
      </div>

      <div className='space-y-2'>
        <Label htmlFor={`${id}-butterbox`}>ButterBox</Label>
        <Select
          value={value.butterboxId || undefined}
          onValueChange={(butterboxId) => onChange('butterboxId', butterboxId)}
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

      <div className='grid gap-4 sm:grid-cols-2'>
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
                  <SelectItem key={model.id} value={model.id}>
                    {model.name || model.id}
                  </SelectItem>
                ))}
                <SelectItem value={CUSTOM_MODEL}>Custom model ID...</SelectItem>
              </SelectContent>
            </Select>
          ) : null}
          {(models.length === 0 || !usesCatalogModel) && (
            <Input
              id={models.length === 0 ? `${id}-model` : `${id}-custom-model`}
              value={value.model}
              onChange={(event) => onChange('model', event.target.value)}
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

        <div className='space-y-2'>
          <Label htmlFor={`${id}-mode`}>Mode</Label>
          <Select
            value={value.mode || DEFAULT_MODE}
            onValueChange={(next) =>
              onChange('mode', next === DEFAULT_MODE ? '' : next)
            }
          >
            <SelectTrigger
              id={`${id}-mode`}
              aria-invalid={Boolean(errors.mode)}
              aria-describedby={errors.mode ? modeErrorId : undefined}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={DEFAULT_MODE}>Box default</SelectItem>
              <SelectItem value='agent'>Agent</SelectItem>
              <SelectItem value='plan'>Plan</SelectItem>
            </SelectContent>
          </Select>
          {errors.mode && (
            <p id={modeErrorId} className='text-sm text-destructive'>
              {errors.mode}
            </p>
          )}
        </div>
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
  )
}

export function CursorAgentConfigurationCard(props: CursorAgentFieldsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Cursor configuration</CardTitle>
        <CardDescription>
          Bind this Agent to a ButterBox and its Cursor runtime.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <CursorAgentFields {...props} />
      </CardContent>
    </Card>
  )
}
