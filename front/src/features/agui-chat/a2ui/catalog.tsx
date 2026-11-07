/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useId, type ReactNode } from 'react'
import {
  Card,
  Column,
  Divider,
  Row as StandardRow,
  createComponentImplementation,
  type ReactComponentImplementation,
} from '@a2ui/react/v0_9'
import {
  ButtonApi,
  Catalog,
  ChoicePickerApi,
  CommonSchemas,
  TextApi,
  TextFieldApi,
} from '@a2ui/web_core/v0_9'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Textarea } from '@/components/ui/textarea'
import { CATALOG_ID } from './protocol'
import styles from './surface.module.css'

// The butter-basic-v1 catalog reuses A2UI's standard layout components, themed
// by surface.module.css. Butter keeps plain Text, KeyValue/Status and its form
// controls for the read-only card policy and server-owned submission state.
// Schemas share the renderer's zod instance. The server validates every model
// message (internal/a2ui/catalog.go); only server-built forms contain inputs
// or actions. Register implementations explicitly, not the whole basicCatalog,
// so importing a standard component never expands that allowlist or functions.

// FormState is what a form surface's inputs need beyond their A2UI props:
// the submission status and the per-field errors to show.
export interface FormState {
  disabled: boolean
  submitting: boolean
  fieldErrors: Record<string, string>
}

export const FormStateContext = createContext<FormState | null>(null)

const text = (v: unknown): string =>
  typeof v === 'string' ? v : v == null ? '' : String(v)

// Official Row deliberately does not wrap. Keep Butter's responsive wrapping
// without duplicating its bindings, child resolution or alignment rules. Both
// SDK render paths use the same presentation-only wrapper (display: contents).
const StandardRowView = StandardRow.view
const Row: ReactComponentImplementation = {
  ...StandardRow,
  render: (props) => (
    <div className={styles.row} data-a2ui-layout='row'>
      <StandardRow.render {...props} />
    </div>
  ),
  view: StandardRowView
    ? (props) => (
        <div className={styles.row} data-a2ui-layout='row'>
          <StandardRowView {...props} />
        </div>
      )
    : undefined,
}

const textVariantClass: Record<string, string> = {
  h1: 'text-lg font-semibold',
  h2: 'text-base font-semibold',
  h3: 'text-sm font-semibold',
  h4: 'text-sm font-medium',
  h5: 'text-xs font-medium uppercase tracking-wide text-muted-foreground',
  caption: 'text-xs text-muted-foreground',
  body: 'text-sm',
}

const headingLevel: Record<string, number> = { h1: 1, h2: 2, h3: 3, h4: 4 }

// Text renders as plain text: the catalog admits no markup, links or images.
// It is a span so it can also label the submit button.
const Text = createComponentImplementation(TextApi, ({ props }) => {
  const variant = props.variant ?? 'body'
  const level = headingLevel[variant]
  return (
    <span
      {...(level ? { role: 'heading', 'aria-level': level } : {})}
      className={cn(
        'block break-words whitespace-pre-wrap',
        textVariantClass[variant] ?? textVariantClass.body
      )}
    >
      {text(props.text)}
    </span>
  )
})

const KeyValueApi = {
  name: 'KeyValue',
  schema: TextApi.schema.pick({ weight: true }).extend({
    label: CommonSchemas.DynamicString,
    value: CommonSchemas.DynamicString,
  }),
}

const KeyValue = createComponentImplementation(KeyValueApi, ({ props }) => (
  <div className='flex items-baseline justify-between gap-4 text-sm'>
    <span className='text-muted-foreground'>{text(props.label)}</span>
    <span className='text-right font-medium break-words'>
      {text(props.value)}
    </span>
  </div>
))

const StatusApi = {
  name: 'Status',
  schema: TextApi.schema.pick({ weight: true }).extend({
    text: CommonSchemas.DynamicString,
    tone: CommonSchemas.ComponentId.optional(),
  }),
}

const toneClass: Record<string, string> = {
  neutral: 'border-border bg-muted text-foreground',
  info: 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300',
  success:
    'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
  warning:
    'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300',
  error: 'border-destructive/30 bg-destructive/10 text-destructive',
}

const Status = createComponentImplementation(StatusApi, ({ props }) => (
  <Badge
    variant='outline'
    className={cn(toneClass[props.tone ?? 'neutral'] ?? toneClass.neutral)}
  >
    {text(props.text)}
  </Badge>
))

// Form inputs. Their value is the renderer's local draft only: nothing is
// sent until the submit button's action fires.

const FormTextFieldApi = {
  name: 'TextField',
  schema: TextFieldApi.schema.extend({
    // The form field this input answers, as the submission names it.
    name: CommonSchemas.ComponentId,
    hint: CommonSchemas.ComponentId.optional(),
    required: CommonSchemas.DynamicBoolean.optional(),
    maxLength: CommonSchemas.DynamicNumber.optional(),
  }),
}

function FieldChrome({
  id,
  label,
  required,
  hint,
  error,
  children,
}: {
  id: string
  label: string
  required?: boolean
  hint?: string
  error?: string
  children: ReactNode
}) {
  return (
    <div className='space-y-1.5'>
      <Label htmlFor={id}>
        {label}
        {required && (
          <span className='text-destructive' aria-hidden='true'>
            *
          </span>
        )}
      </Label>
      {children}
      {hint && (
        <p id={`${id}-hint`} className='text-xs text-muted-foreground'>
          {hint}
        </p>
      )}
      {error && (
        <p id={`${id}-error`} role='alert' className='text-xs text-destructive'>
          {error}
        </p>
      )}
    </div>
  )
}

function describedBy(id: string, hint?: string, error?: string) {
  return (
    [hint ? `${id}-hint` : '', error ? `${id}-error` : '']
      .filter(Boolean)
      .join(' ') || undefined
  )
}

const TextField = createComponentImplementation(
  FormTextFieldApi,
  ({ props }) => {
    const form = useContext(FormStateContext)
    const id = useId()
    const name = props.name
    const error = form?.fieldErrors[name]
    const hint = props.hint ? text(props.hint) : undefined
    const common = {
      id,
      name,
      value: text(props.value),
      disabled: form?.disabled,
      required: !!props.required,
      'aria-required': !!props.required,
      'aria-invalid': !!error,
      'aria-describedby': describedBy(id, hint, error),
      maxLength:
        typeof props.maxLength === 'number' ? props.maxLength : undefined,
    }
    return (
      <FieldChrome
        id={id}
        label={text(props.label)}
        required={!!props.required}
        hint={hint}
        error={error}
      >
        {props.variant === 'longText' ? (
          <Textarea
            {...common}
            rows={3}
            onChange={(e) => props.setValue(e.target.value)}
          />
        ) : (
          <Input {...common} onChange={(e) => props.setValue(e.target.value)} />
        )}
      </FieldChrome>
    )
  }
)

const FormChoicePickerApi = {
  name: 'ChoicePicker',
  schema: ChoicePickerApi.schema.extend({
    name: CommonSchemas.ComponentId,
    hint: CommonSchemas.ComponentId.optional(),
    required: CommonSchemas.DynamicBoolean.optional(),
  }),
}

const ChoicePicker = createComponentImplementation(
  FormChoicePickerApi,
  ({ props }) => {
    const form = useContext(FormStateContext)
    const id = useId()
    const name = props.name
    const error = form?.fieldErrors[name]
    const hint = props.hint ? text(props.hint) : undefined
    const selected = Array.isArray(props.value) ? (props.value[0] ?? '') : ''
    return (
      <div className='space-y-1.5' role='group' aria-labelledby={`${id}-label`}>
        <p id={`${id}-label`} className='text-sm leading-none font-medium'>
          {text(props.label)}
          {props.required && (
            <span className='text-destructive' aria-hidden='true'>
              *
            </span>
          )}
        </p>
        <RadioGroup
          value={selected}
          onValueChange={(v) => props.setValue([v])}
          disabled={form?.disabled}
          aria-labelledby={`${id}-label`}
          aria-required={!!props.required}
          aria-invalid={!!error}
          aria-describedby={describedBy(id, hint, error)}
          className='gap-2'
        >
          {(props.options ?? []).map((opt, i) => {
            const optId = `${id}-opt-${i}`
            return (
              <div key={opt.value} className='flex items-center gap-2'>
                <RadioGroupItem id={optId} value={opt.value} />
                <Label htmlFor={optId} className='font-normal'>
                  {text(opt.label)}
                </Label>
              </div>
            )
          })}
        </RadioGroup>
        {hint && (
          <p id={`${id}-hint`} className='text-xs text-muted-foreground'>
            {hint}
          </p>
        )}
        {error && (
          <p
            id={`${id}-error`}
            role='alert'
            className='text-xs text-destructive'
          >
            {error}
          </p>
        )}
      </div>
    )
  }
)

const SubmitButton = createComponentImplementation(
  ButtonApi,
  ({ props, buildChild }) => {
    const form = useContext(FormStateContext)
    return (
      <div>
        <Button
          type='button'
          size='sm'
          variant={props.variant === 'primary' ? 'default' : 'outline'}
          disabled={form?.disabled}
          aria-busy={form?.submitting}
          onClick={() => props.action()}
        >
          {form?.submitting && <Loader2 className='size-3.5 animate-spin' />}
          {props.child ? buildChild(props.child) : null}
        </Button>
      </div>
    )
  }
)

const components: ReactComponentImplementation[] = [
  Column,
  Row,
  Text,
  Card,
  Divider,
  KeyValue,
  Status,
  TextField,
  ChoicePicker,
  SubmitButton,
]

export const butterCatalog = new Catalog<ReactComponentImplementation>(
  CATALOG_ID,
  '0.9.1',
  components,
  []
)
