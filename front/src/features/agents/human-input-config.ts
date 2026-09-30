import { z } from 'zod'
import type {
  HumanInputForm,
  HumanInputFormField,
  WorkflowConfig,
  WorkflowNode,
} from '@/types/api'

// Human Input node forms (issue #350): what an A2UI-capable AG-UI client
// renders for a Workflow Agent's HUMAN_INPUT node. The server validates the
// same rules on save (internal/a2ui/form.go); checking them here too lets
// the editor point at the offending field before a round trip.

export const MAX_FORM_FIELDS = 20
export const MAX_TEXT_LENGTH = 2000
export const MAX_CHOICE_OPTIONS = 50

const FIELD_NAME = /^[A-Za-z_][A-Za-z0-9_]*$/

export const humanInputOptionSchema = z.object({
  value: z.string(),
  label: z.string(),
})

export const humanInputFieldSchema = z.object({
  name: z.string(),
  label: z.string(),
  hint: z.string(),
  required: z.boolean(),
  type: z.enum(['text', 'single_choice']),
  max_length: z.string(),
  options: z.array(humanInputOptionSchema),
})

export const humanInputNodeSchema = z.object({
  node: z.string(),
  question: z.string(),
  form_enabled: z.boolean(),
  title: z.string(),
  fields: z.array(humanInputFieldSchema),
})

export type HumanInputFieldValues = z.infer<typeof humanInputFieldSchema>
export type HumanInputNodeValues = z.infer<typeof humanInputNodeSchema>

export function emptyField(): HumanInputFieldValues {
  return {
    name: '',
    label: '',
    hint: '',
    required: false,
    type: 'text',
    max_length: '',
    options: [],
  }
}

function fieldValues(f: HumanInputFormField): HumanInputFieldValues {
  const choice = f.type === 'HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE'
  return {
    name: f.name ?? '',
    label: f.label ?? '',
    hint: f.hint ?? '',
    required: !!f.required,
    type: choice ? 'single_choice' : 'text',
    max_length: !choice && f.max_length ? String(f.max_length) : '',
    options: (f.options ?? []).map((o) => ({
      value: o.value ?? '',
      label: o.label ?? '',
    })),
  }
}

function hasForm(form: HumanInputForm | undefined): boolean {
  return !!form && (!!form.title?.trim() || (form.fields?.length ?? 0) > 0)
}

// humanInputValuesFromWorkflow lists the graph's HUMAN_INPUT nodes in order.
export function humanInputValuesFromWorkflow(
  workflow: WorkflowConfig | undefined
): HumanInputNodeValues[] {
  return (workflow?.nodes ?? [])
    .filter((n) => n.kind === 'WORKFLOW_NODE_KIND_HUMAN_INPUT' && n.name)
    .map((n) => ({
      node: n.name ?? '',
      question: n.question ?? '',
      form_enabled: hasForm(n.form),
      title: n.form?.title ?? '',
      fields: (n.form?.fields ?? []).map(fieldValues),
    }))
}

function buildForm(values: HumanInputNodeValues): HumanInputForm | undefined {
  if (!values.form_enabled) return undefined
  return {
    title: values.title.trim() || undefined,
    fields: values.fields.map((f) => {
      const choice = f.type === 'single_choice'
      const field: HumanInputFormField = {
        name: f.name.trim(),
        label: f.label.trim(),
        hint: f.hint.trim() || undefined,
        required: f.required || undefined,
        type: choice
          ? 'HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE'
          : 'HUMAN_INPUT_FORM_FIELD_TYPE_TEXT',
      }
      if (choice) {
        field.options = f.options.map((o) => ({
          value: o.value.trim(),
          label: o.label.trim(),
        }))
      } else if (f.max_length.trim() !== '') {
        field.max_length = Number(f.max_length)
      }
      return field
    }),
  }
}

// applyHumanInputValues writes the edited questions and forms back onto the
// graph's HUMAN_INPUT nodes, leaving every other part of the graph as is.
export function applyHumanInputValues(
  workflow: WorkflowConfig | undefined,
  values: HumanInputNodeValues[]
): WorkflowConfig | undefined {
  if (!workflow) return workflow
  const byNode = new Map(values.map((v) => [v.node, v]))
  return {
    ...workflow,
    nodes: (workflow.nodes ?? []).map((n): WorkflowNode => {
      const edited =
        n.kind === 'WORKFLOW_NODE_KIND_HUMAN_INPUT' && n.name
          ? byNode.get(n.name)
          : undefined
      if (!edited) return n
      return { ...n, question: edited.question, form: buildForm(edited) }
    }),
  }
}

// validateHumanInputNodes reports each broken rule at the input it concerns.
export function validateHumanInputNodes(
  nodes: HumanInputNodeValues[],
  ctx: z.RefinementCtx,
  path: (string | number)[]
) {
  nodes.forEach((node, n) => {
    const at = (...rest: (string | number)[]) => [...path, n, ...rest]
    if (node.question.trim() === '') {
      ctx.addIssue({
        code: 'custom',
        path: at('question'),
        message: 'A Human Input node needs a question.',
      })
    }
    if (!node.form_enabled) return
    if (node.fields.length === 0) {
      ctx.addIssue({
        code: 'custom',
        path: at('fields'),
        message: 'Add at least one field, or turn the form off.',
      })
    }
    if (node.fields.length > MAX_FORM_FIELDS) {
      ctx.addIssue({
        code: 'custom',
        path: at('fields'),
        message: `A form may have at most ${MAX_FORM_FIELDS} fields.`,
      })
    }
    const seen = new Set<string>()
    node.fields.forEach((f, i) => {
      const name = f.name.trim()
      if (!FIELD_NAME.test(name)) {
        ctx.addIssue({
          code: 'custom',
          path: at('fields', i, 'name'),
          message:
            'Use letters, digits or underscores, starting with a letter or underscore.',
        })
      } else if (seen.has(name)) {
        ctx.addIssue({
          code: 'custom',
          path: at('fields', i, 'name'),
          message: `Field name "${name}" is used twice.`,
        })
      }
      seen.add(name)
      if (f.label.trim() === '') {
        ctx.addIssue({
          code: 'custom',
          path: at('fields', i, 'label'),
          message: 'A label is required.',
        })
      }
      if (f.type === 'text' && f.max_length.trim() !== '') {
        const max = Number(f.max_length)
        if (!Number.isInteger(max) || max < 0 || max > MAX_TEXT_LENGTH) {
          ctx.addIssue({
            code: 'custom',
            path: at('fields', i, 'max_length'),
            message: `Use a whole number from 0 to ${MAX_TEXT_LENGTH} (empty means ${MAX_TEXT_LENGTH}).`,
          })
        }
      }
      if (f.type === 'single_choice') {
        if (f.options.length === 0) {
          ctx.addIssue({
            code: 'custom',
            path: at('fields', i, 'options'),
            message: 'Add at least one option.',
          })
        }
        if (f.options.length > MAX_CHOICE_OPTIONS) {
          ctx.addIssue({
            code: 'custom',
            path: at('fields', i, 'options'),
            message: `A choice may have at most ${MAX_CHOICE_OPTIONS} options.`,
          })
        }
        const values = new Set<string>()
        f.options.forEach((o, j) => {
          const value = o.value.trim()
          if (value === '') {
            ctx.addIssue({
              code: 'custom',
              path: at('fields', i, 'options', j, 'value'),
              message: 'A value is required.',
            })
          } else if (values.has(value)) {
            ctx.addIssue({
              code: 'custom',
              path: at('fields', i, 'options', j, 'value'),
              message: `Value "${value}" is used twice.`,
            })
          }
          values.add(value)
          if (o.label.trim() === '') {
            ctx.addIssue({
              code: 'custom',
              path: at('fields', i, 'options', j, 'label'),
              message: 'A label is required.',
            })
          }
        })
      }
    })
  })
}
