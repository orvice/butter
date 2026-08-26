import { z } from 'zod'
import type { ModelConfig } from '@/types/api'

// Pure seam between the structured model rows in the provider form and the
// wire-level ModelConfig[] (issue #321). No DOM, no react-hook-form, so the
// mapping and validation can be unit-tested directly.

export const MODEL_ID_REQUIRED_MESSAGE = 'Model ID is required'
export const AT_LEAST_ONE_MODEL_MESSAGE = 'At least one model is required'
export const CONTEXT_WINDOW_TOKENS_MESSAGE =
  'Context Window must be a whole number from 0 to 4,294,967,295'
export const MAX_CONTEXT_WINDOW_TOKENS = 4_294_967_295

// Editable strings preserve a genuinely blank optional capacity. The mapper
// emits only positive capacities because zero has the same wire meaning as
// unset.
export interface ModelRowValues {
  name: string
  alias: string
  contextWindowTokens: string
}

export const modelRowSchema = z.object({
  name: z.string().trim().min(1, MODEL_ID_REQUIRED_MESSAGE),
  alias: z.string().trim(),
  contextWindowTokens: z
    .string()
    .trim()
    .refine(
      (value) =>
        value === '' ||
        (/^\d+$/.test(value) && Number(value) <= MAX_CONTEXT_WINDOW_TOKENS),
      CONTEXT_WINDOW_TOKENS_MESSAGE
    ),
})

export const modelsSchema = z
  .array(modelRowSchema)
  .min(1, AT_LEAST_ONE_MODEL_MESSAGE)

export function blankModelRow(): ModelRowValues {
  return { name: '', alias: '', contextWindowTokens: '' }
}

export function modelsToRows(models?: ModelConfig[]): ModelRowValues[] {
  return (models ?? []).map((model) => ({
    name: model.name ?? '',
    alias: model.alias ?? '',
    contextWindowTokens: model.context_window_tokens
      ? String(model.context_window_tokens)
      : '',
  }))
}

// Pure shape mapper: values arrive already trimmed because the form resolver
// parses rows through modelRowSchema, which owns whitespace normalization.
export function rowsToModels(rows: readonly ModelRowValues[]): ModelConfig[] {
  return rows.map((row) => {
    const model: ModelConfig = row.alias
      ? { name: row.name, alias: row.alias }
      : { name: row.name }
    const contextWindowTokens = Number(row.contextWindowTokens)
    if (contextWindowTokens > 0) {
      model.context_window_tokens = contextWindowTokens
    }
    return model
  })
}
