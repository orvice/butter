import { z } from 'zod'
import type { ModelConfig } from '@/types/api'

// Pure seam between the structured model rows in the provider form and the
// wire-level ModelConfig[] (issue #321). No DOM, no react-hook-form, so the
// mapping and validation can be unit-tested directly.

export const MODEL_ID_REQUIRED_MESSAGE = 'Model ID is required'
export const AT_LEAST_ONE_MODEL_MESSAGE = 'At least one model is required'

// One editable row in the Models section. Alias is held as '' while editing
// and omitted from the submitted payload when blank (see rowsToModels).
export interface ModelRowValues {
  name: string
  alias: string
}

export const modelRowSchema = z.object({
  name: z.string().trim().min(1, MODEL_ID_REQUIRED_MESSAGE),
  alias: z.string().trim(),
})

export const modelsSchema = z
  .array(modelRowSchema)
  .min(1, AT_LEAST_ONE_MODEL_MESSAGE)

export function blankModelRow(): ModelRowValues {
  return { name: '', alias: '' }
}

export function modelsToRows(models?: ModelConfig[] | null): ModelRowValues[] {
  return (models ?? []).map((model) => ({
    name: model.name ?? '',
    alias: model.alias ?? '',
  }))
}

export function rowsToModels(rows: readonly ModelRowValues[]): ModelConfig[] {
  return rows.map((row) => {
    const name = row.name.trim()
    const alias = row.alias.trim()
    return alias ? { name, alias } : { name }
  })
}
