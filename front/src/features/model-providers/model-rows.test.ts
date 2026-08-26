import { describe, expect, it } from 'vitest'
import {
  AT_LEAST_ONE_MODEL_MESSAGE,
  MODEL_ID_REQUIRED_MESSAGE,
  blankModelRow,
  modelRowSchema,
  modelsSchema,
  modelsToRows,
  rowsToModels,
} from './model-rows'

describe('modelsToRows', () => {
  it('maps every model to an editable row, keeping empty aliases as empty strings', () => {
    expect(
      modelsToRows([
        { name: 'gpt-4o', alias: '4o' },
        { name: 'gemini-2.5-pro' },
        { name: 'plain', alias: '' },
      ])
    ).toEqual([
      { name: 'gpt-4o', alias: '4o' },
      { name: 'gemini-2.5-pro', alias: '' },
      { name: 'plain', alias: '' },
    ])
  })

  it('treats missing models as no rows', () => {
    expect(modelsToRows(undefined)).toEqual([])
    expect(modelsToRows([])).toEqual([])
  })
})

describe('blankModelRow', () => {
  it('starts a new row with neither a model ID nor an alias', () => {
    expect(blankModelRow()).toEqual({ name: '', alias: '' })
  })
})

describe('rowsToModels', () => {
  it('omits blank aliases and preserves the submitted row order', () => {
    expect(
      rowsToModels([
        { name: 'gemini-2.5-pro', alias: 'pro' },
        { name: 'gpt-4.1', alias: '' },
        { name: 'gpt-4o', alias: '4o' },
      ])
    ).toEqual([
      { name: 'gemini-2.5-pro', alias: 'pro' },
      { name: 'gpt-4.1' },
      { name: 'gpt-4o', alias: '4o' },
    ])
  })

  it('trims whitespace through the schema before mapping, like the real submit path', () => {
    // The form resolver parses rows through modelsSchema first, so its .trim()
    // owns normalization; rowsToModels maps the parsed values as-is.
    const parsed = modelsSchema.parse([
      { name: '  gpt-4o  ', alias: '   ' },
      { name: ' gpt-4o ', alias: ' 4o ' },
    ])

    expect(rowsToModels(parsed)).toEqual([
      { name: 'gpt-4o' },
      { name: 'gpt-4o', alias: '4o' },
    ])
  })

  it('maps each row to the payload shape without altering its exact values', () => {
    const row = { name: ' gpt-4o ', alias: '' }

    expect(rowsToModels([row])).toEqual([{ name: ' gpt-4o ' }])
    expect(row).toEqual({ name: ' gpt-4o ', alias: '' })
  })
})

describe('existing-data round trips', () => {
  it('loads a stored provider into rows and back without changing models', () => {
    const stored = [{ name: 'gpt-4o', alias: '4o' }, { name: 'gemini-2.5-pro' }]

    expect(rowsToModels(modelsToRows(stored))).toEqual(stored)
  })

  it('keeps aliases distinct per row across repeated round trips', () => {
    const stored = [
      { name: 'shared-model', alias: 'a' },
      { name: 'shared-model', alias: 'b' },
    ]
    const once = modelsToRows(stored)

    expect(once.map((row) => row.alias)).toEqual(['a', 'b'])
    expect(rowsToModels(modelsToRows(rowsToModels(once)))).toEqual(stored)
  })

  it("round-trips a provider's model list without inventing or reordering rows", () => {
    // Only the models array is mapped back and forth, so each stored model
    // comes back exactly once, in order.
    const providerModels = [{ name: 'ollama-local' }]

    expect(rowsToModels(modelsToRows(providerModels))).toEqual([
      { name: 'ollama-local' },
    ])
  })
})

describe('model validation (add/remove behavior)', () => {
  it('rejects a submitted payload whose appended row has an empty model ID', () => {
    const existing = rowsToModels(
      modelsToRows([{ name: 'gpt-4o', alias: '4o' }])
    )
    const rowsAfterAdd = [
      ...modelsToRows([{ name: 'gpt-4o', alias: '4o' }]),
      blankModelRow(),
    ]

    const parsed = modelsSchema.safeParse(rowsAfterAdd)
    expect(parsed.success).toBe(false)
    if (!parsed.success) {
      const issue = parsed.error.issues.find(
        (candidate) => candidate.message === MODEL_ID_REQUIRED_MESSAGE
      )
      expect(issue?.path).toContain(1)
      expect(issue?.path).toContain('name')
    }
    // Removing the appended row restores the exact previous configuration.
    expect(rowsToModels(rowsAfterAdd.slice(0, -1))).toEqual(existing)
  })

  it('attaches per-row errors to the specific row with the empty model ID', () => {
    const parsed = modelRowSchema.safeParse({ name: '', alias: 'x' })
    expect(parsed.success).toBe(false)
    if (!parsed.success) {
      expect(parsed.error.issues[0].message).toBe(MODEL_ID_REQUIRED_MESSAGE)
    }

    const filled = modelRowSchema.safeParse({ name: 'm', alias: '' })
    expect(filled.success).toBe(true)
  })

  it('rejects whitespace-only model IDs so every submitted ID is non-empty', () => {
    const parsed = modelsSchema.safeParse([{ name: '   ', alias: '' }])
    expect(parsed.success).toBe(false)
    if (!parsed.success) {
      expect(parsed.error.issues[0].message).toBe(MODEL_ID_REQUIRED_MESSAGE)
    }
  })

  it('requires at least one model when every row is removed', () => {
    const parsed = modelsSchema.safeParse([])
    expect(parsed.success).toBe(false)
    if (!parsed.success) {
      expect(parsed.error.issues[0].message).toBe(AT_LEAST_ONE_MODEL_MESSAGE)
      expect(parsed.error.issues[0].path).toEqual([])
    }
  })

  it('normalizes trimmed values through the schema like the resolver would on submit', () => {
    const parsed = modelsSchema.parse([
      { name: ' gemini-2.5-pro ', alias: ' pro ' },
    ])
    expect(parsed).toEqual([{ name: 'gemini-2.5-pro', alias: 'pro' }])
  })
})
