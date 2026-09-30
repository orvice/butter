import { A2UI_VERSION, type FormField, type FormView } from './protocol'

// Form answers as the server expects them: one string per configured field.
// A single-choice field's draft is the A2UI ChoicePicker's string list.
export function formValues(
  form: FormView,
  draft: Record<string, unknown> | undefined
): Record<string, string> {
  const out: Record<string, string> = {}
  for (const field of form.fields) {
    const raw = draft?.[field.name]
    if (Array.isArray(raw)) {
      out[field.name] = typeof raw[0] === 'string' ? raw[0] : ''
    } else if (typeof raw === 'string') {
      out[field.name] = field.type === 'text' ? raw.trim() : raw
    } else {
      out[field.name] = ''
    }
  }
  return out
}

// validateForm mirrors the server's field rules so most mistakes are caught
// before a submission is sent. The server validates independently.
export function validateForm(
  form: FormView,
  values: Record<string, string>
): Record<string, string> {
  const errors: Record<string, string> = {}
  for (const field of form.fields) {
    const value = values[field.name] ?? ''
    if (value === '') {
      if (field.required) errors[field.name] = 'Required'
      continue
    }
    if (field.type === 'text' && field.maxLength) {
      // The server counts characters (code points), not UTF-16 units.
      const length = Array.from(value).length
      if (length > field.maxLength) {
        errors[field.name] =
          `At most ${field.maxLength} characters (${length} now)`
      }
    }
    if (
      field.type === 'single_choice' &&
      !(field.options ?? []).some((o) => o.value === value)
    ) {
      errors[field.name] = 'Choose one of the options'
    }
  }
  return errors
}

// submissionPayload is the AG-UI resume payload of a form submission.
export function submissionPayload(
  surfaceId: string,
  form: FormView,
  values: Record<string, string>
) {
  return {
    butterForm: {
      version: A2UI_VERSION,
      surfaceId,
      revision: form.revision,
      token: form.token,
      values,
    },
  }
}

function displayValue(field: FormField, value: string): string {
  if (value === '') return '—'
  if (field.type === 'single_choice') {
    return field.options?.find((o) => o.value === value)?.label ?? value
  }
  return value
}

// readableReply is how a submission appears in the conversation.
export function readableReply(
  form: FormView,
  values: Record<string, string>
): string {
  const lines = [form.title]
  for (const field of form.fields) {
    lines.push(
      `${field.label}: ${displayValue(field, values[field.name] ?? '')}`
    )
  }
  return lines.join('\n')
}
