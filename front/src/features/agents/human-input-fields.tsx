import { useId } from 'react'
import { ArrowDown, ArrowUp, ClipboardList, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  emptyField,
  MAX_CHOICE_OPTIONS,
  MAX_FORM_FIELDS,
  MAX_TEXT_LENGTH,
  type HumanInputFieldValues,
  type HumanInputNodeValues,
} from './human-input-config'

// Errors mirror the form values' shape, as react-hook-form reports them.
interface FieldMessage {
  message?: string
}
interface OptionErrors {
  value?: FieldMessage
  label?: FieldMessage
}
interface FieldErrors {
  name?: FieldMessage
  label?: FieldMessage
  max_length?: FieldMessage
  options?: FieldMessage & Record<number, OptionErrors | undefined>
}
export interface HumanInputNodeErrors {
  question?: FieldMessage
  fields?: FieldMessage & Record<number, FieldErrors | undefined>
}

function move<T>(list: T[], from: number, to: number): T[] {
  if (to < 0 || to >= list.length) return list
  const next = [...list]
  const [item] = next.splice(from, 1)
  next.splice(to, 0, item)
  return next
}

function ErrorText({ id, error }: { id?: string; error?: FieldMessage }) {
  if (!error?.message) return null
  return (
    <p id={id} role='alert' className='text-xs text-destructive'>
      {error.message}
    </p>
  )
}

// HumanInputConfigurationCard edits the question and form of every Human Input node
// of a Workflow Agent. The rest of the graph stays in JSON mode.
export function HumanInputConfigurationCard({
  value,
  onChange,
  errors,
}: {
  value: HumanInputNodeValues[]
  onChange: (value: HumanInputNodeValues[]) => void
  errors?: Record<number, HumanInputNodeErrors | undefined>
}) {
  if (value.length === 0) return null
  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <ClipboardList className='size-4' />
          Human Input nodes
        </CardTitle>
        <CardDescription>
          The question each node asks when it pauses the workflow. A form shows
          the question as fields in Chat; every other channel gets the
          question followed by the field list and may answer in text. The
          node&apos;s successor always receives text — a submitted form arrives
          as a JSON object with one member per field.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-6'>
        {value.map((node, i) => (
          <HumanInputNodeEditor
            key={node.node}
            value={node}
            errors={errors?.[i]}
            onChange={(next) =>
              onChange(value.map((v, j) => (j === i ? next : v)))
            }
          />
        ))}
      </CardContent>
    </Card>
  )
}

function HumanInputNodeEditor({
  value,
  onChange,
  errors,
}: {
  value: HumanInputNodeValues
  onChange: (value: HumanInputNodeValues) => void
  errors?: HumanInputNodeErrors
}) {
  const id = useId()
  const set = <K extends keyof HumanInputNodeValues>(
    key: K,
    next: HumanInputNodeValues[K]
  ) => onChange({ ...value, [key]: next })
  const setField = (i: number, next: HumanInputFieldValues) =>
    set(
      'fields',
      value.fields.map((f, j) => (j === i ? next : f))
    )

  return (
    <section
      aria-labelledby={`${id}-node`}
      className='space-y-4 rounded-lg border border-border p-4'
    >
      <h3 id={`${id}-node`} className='font-mono text-sm font-medium'>
        {value.node}
      </h3>
      <div className='space-y-1.5'>
        <Label htmlFor={`${id}-question`}>Question</Label>
        <Textarea
          id={`${id}-question`}
          value={value.question}
          rows={2}
          aria-invalid={!!errors?.question}
          onChange={(e) => set('question', e.target.value)}
        />
        <ErrorText error={errors?.question} />
      </div>
      <div className='flex items-center justify-between gap-4'>
        <Label
          htmlFor={`${id}-form`}
          className='flex flex-col items-start gap-1'
        >
          <span>Show as a form</span>
          <span className='text-xs font-normal text-muted-foreground'>
            Fields with fixed options and checks, submitted as one answer.
          </span>
        </Label>
        <Switch
          id={`${id}-form`}
          checked={value.form_enabled}
          onCheckedChange={(checked) =>
            onChange({
              ...value,
              form_enabled: checked,
              fields:
                checked && value.fields.length === 0
                  ? [emptyField()]
                  : value.fields,
            })
          }
        />
      </div>
      {value.form_enabled && (
        <div className='space-y-4'>
          <div className='space-y-1.5'>
            <Label htmlFor={`${id}-title`}>Form title</Label>
            <Input
              id={`${id}-title`}
              value={value.title}
              placeholder='Defaults to the question'
              onChange={(e) => set('title', e.target.value)}
            />
          </div>
          <ErrorText error={errors?.fields} />
          <ol className='space-y-3'>
            {value.fields.map((field, i) => (
              <li key={i}>
                <FieldEditor
                  index={i}
                  count={value.fields.length}
                  value={field}
                  errors={errors?.fields?.[i]}
                  onChange={(next) => setField(i, next)}
                  onMove={(to) => set('fields', move(value.fields, i, to))}
                  onRemove={() =>
                    set(
                      'fields',
                      value.fields.filter((_, j) => j !== i)
                    )
                  }
                />
              </li>
            ))}
          </ol>
          <Button
            type='button'
            variant='outline'
            size='sm'
            disabled={value.fields.length >= MAX_FORM_FIELDS}
            onClick={() => set('fields', [...value.fields, emptyField()])}
          >
            <Plus className='size-4' />
            Add field
          </Button>
        </div>
      )}
    </section>
  )
}

function FieldEditor({
  index,
  count,
  value,
  onChange,
  onMove,
  onRemove,
  errors,
}: {
  index: number
  count: number
  value: HumanInputFieldValues
  onChange: (value: HumanInputFieldValues) => void
  onMove: (to: number) => void
  onRemove: () => void
  errors?: FieldErrors
}) {
  const id = useId()
  const n = index + 1
  const set = <K extends keyof HumanInputFieldValues>(
    key: K,
    next: HumanInputFieldValues[K]
  ) => onChange({ ...value, [key]: next })
  const setOption = (j: number, key: 'value' | 'label', next: string) =>
    set(
      'options',
      value.options.map((o, k) => (k === j ? { ...o, [key]: next } : o))
    )

  return (
    <fieldset className='space-y-3 rounded-md border border-dashed border-border p-3'>
      <div className='flex items-center justify-between gap-2'>
        <legend className='text-sm font-medium'>Field {n}</legend>
        <div className='flex items-center gap-1'>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            className='size-7'
            aria-label={`Move field ${n} up`}
            disabled={index === 0}
            onClick={() => onMove(index - 1)}
          >
            <ArrowUp className='size-3.5' />
          </Button>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            className='size-7'
            aria-label={`Move field ${n} down`}
            disabled={index === count - 1}
            onClick={() => onMove(index + 1)}
          >
            <ArrowDown className='size-3.5' />
          </Button>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            className='size-7'
            aria-label={`Remove field ${n}`}
            onClick={onRemove}
          >
            <Trash2 className='size-3.5' />
          </Button>
        </div>
      </div>
      <div className='grid gap-3 sm:grid-cols-2'>
        <div className='space-y-1.5'>
          <Label htmlFor={`${id}-name`}>Field {n} name</Label>
          <Input
            id={`${id}-name`}
            className='font-mono'
            value={value.name}
            placeholder='environment'
            aria-invalid={!!errors?.name}
            onChange={(e) => set('name', e.target.value)}
          />
          <ErrorText error={errors?.name} />
        </div>
        <div className='space-y-1.5'>
          <Label htmlFor={`${id}-label`}>Field {n} label</Label>
          <Input
            id={`${id}-label`}
            value={value.label}
            placeholder='Environment'
            aria-invalid={!!errors?.label}
            onChange={(e) => set('label', e.target.value)}
          />
          <ErrorText error={errors?.label} />
        </div>
        <div className='space-y-1.5'>
          <Label htmlFor={`${id}-type`}>Field {n} type</Label>
          <Select
            value={value.type}
            onValueChange={(t) =>
              onChange({
                ...value,
                type: t as HumanInputFieldValues['type'],
                max_length: t === 'text' ? value.max_length : '',
                options:
                  t === 'single_choice' && value.options.length === 0
                    ? [{ value: '', label: '' }]
                    : value.options,
              })
            }
          >
            <SelectTrigger id={`${id}-type`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='text'>Text</SelectItem>
              <SelectItem value='single_choice'>Single choice</SelectItem>
            </SelectContent>
          </Select>
        </div>
        {value.type === 'text' && (
          <div className='space-y-1.5'>
            <Label htmlFor={`${id}-max`}>Field {n} max length</Label>
            <Input
              id={`${id}-max`}
              inputMode='numeric'
              value={value.max_length}
              placeholder={String(MAX_TEXT_LENGTH)}
              aria-invalid={!!errors?.max_length}
              onChange={(e) => set('max_length', e.target.value)}
            />
            <ErrorText error={errors?.max_length} />
          </div>
        )}
        <div className='space-y-1.5 sm:col-span-2'>
          <Label htmlFor={`${id}-hint`}>Field {n} hint</Label>
          <Input
            id={`${id}-hint`}
            value={value.hint}
            placeholder='Optional help shown under the input'
            onChange={(e) => set('hint', e.target.value)}
          />
        </div>
      </div>
      <div className='flex items-center gap-2'>
        <Switch
          id={`${id}-required`}
          checked={value.required}
          onCheckedChange={(checked) => set('required', checked)}
        />
        <Label htmlFor={`${id}-required`}>Field {n} required</Label>
      </div>
      {value.type === 'single_choice' && (
        <div className='space-y-2'>
          <p className='text-sm font-medium'>Field {n} options</p>
          <ErrorText error={errors?.options} />
          {value.options.map((option, j) => (
            <div key={j} className='flex items-start gap-2'>
              <div className='flex-1 space-y-1'>
                <Input
                  aria-label={`Field ${n} option ${j + 1} value`}
                  className='font-mono'
                  value={option.value}
                  placeholder='value'
                  aria-invalid={!!errors?.options?.[j]?.value}
                  onChange={(e) => setOption(j, 'value', e.target.value)}
                />
                <ErrorText error={errors?.options?.[j]?.value} />
              </div>
              <div className='flex-1 space-y-1'>
                <Input
                  aria-label={`Field ${n} option ${j + 1} label`}
                  value={option.label}
                  placeholder='Label'
                  aria-invalid={!!errors?.options?.[j]?.label}
                  onChange={(e) => setOption(j, 'label', e.target.value)}
                />
                <ErrorText error={errors?.options?.[j]?.label} />
              </div>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='size-9'
                aria-label={`Move field ${n} option ${j + 1} up`}
                disabled={j === 0}
                onClick={() => set('options', move(value.options, j, j - 1))}
              >
                <ArrowUp className='size-3.5' />
              </Button>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='size-9'
                aria-label={`Remove field ${n} option ${j + 1}`}
                onClick={() =>
                  set(
                    'options',
                    value.options.filter((_, k) => k !== j)
                  )
                }
              >
                <Trash2 className='size-3.5' />
              </Button>
            </div>
          ))}
          <Button
            type='button'
            variant='outline'
            size='sm'
            disabled={value.options.length >= MAX_CHOICE_OPTIONS}
            onClick={() =>
              set('options', [...value.options, { value: '', label: '' }])
            }
          >
            <Plus className='size-4' />
            Add option
          </Button>
        </div>
      )}
    </fieldset>
  )
}
