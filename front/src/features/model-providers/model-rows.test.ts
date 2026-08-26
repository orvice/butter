import { describe, expect, it } from 'vitest'
import {
  AT_LEAST_ONE_MODEL_MESSAGE,
  CONTEXT_WINDOW_TOKENS_MESSAGE,
  MAX_CONTEXT_WINDOW_TOKENS,
  MODEL_ID_REQUIRED_MESSAGE,
  blankModelRow,
  modelRowSchema,
  modelsSchema,
  modelsToRows,
  rowsToModels,
} from './model-rows'

describe('modelsToRows', () => {
  it('maps model IDs, aliases, and configured capacities to editable strings', () => {
    expect(
      modelsToRows([
        {
          name: 'gpt-4o',
          alias: '4o',
          context_window_tokens: 128_000,
        },
        { name: 'gemini-2.5-pro' },
        { name: 'plain', alias: '', context_window_tokens: 0 },
      ])
    ).toEqual([
      { name: 'gpt-4o', alias: '4o', contextWindowTokens: '128000' },
      { name: 'gemini-2.5-pro', alias: '', contextWindowTokens: '' },
      { name: 'plain', alias: '', contextWindowTokens: '' },
    ])
  })

  it('treats missing models as no rows', () => {
    expect(modelsToRows(undefined)).toEqual([])
    expect(modelsToRows([])).toEqual([])
  })
})

describe('blankModelRow', () => {
  it('starts a new row with blank optional model metadata', () => {
    expect(blankModelRow()).toEqual({
      name: '',
      alias: '',
      contextWindowTokens: '',
    })
  })
})

describe('rowsToModels', () => {
  it('omits blank aliases and capacities while preserving row order', () => {
    expect(
      rowsToModels([
        {
          name: 'gemini-2.5-pro',
          alias: 'pro',
          contextWindowTokens: '1048576',
        },
        { name: 'gpt-4.1', alias: '', contextWindowTokens: '' },
        { name: 'gpt-4o', alias: '4o', contextWindowTokens: '0' },
      ])
    ).toEqual([
      {
        name: 'gemini-2.5-pro',
        alias: 'pro',
        context_window_tokens: 1_048_576,
      },
      { name: 'gpt-4.1' },
      { name: 'gpt-4o', alias: '4o' },
    ])
  })

  it('trims whitespace through the schema before mapping, like the real submit path', () => {
    const parsed = modelsSchema.parse([
      {
        name: '  gpt-4o  ',
        alias: '   ',
        contextWindowTokens: ' 128000 ',
      },
      { name: ' gpt-4o ', alias: ' 4o ', contextWindowTokens: ' ' },
    ])

    expect(rowsToModels(parsed)).toEqual([
      { name: 'gpt-4o', context_window_tokens: 128_000 },
      { name: 'gpt-4o', alias: '4o' },
    ])
  })

  it('maps row strings without altering model IDs or aliases', () => {
    const row = {
      name: ' gpt-4o ',
      alias: '',
      contextWindowTokens: '64000',
    }

    expect(rowsToModels([row])).toEqual([
      { name: ' gpt-4o ', context_window_tokens: 64_000 },
    ])
    expect(row).toEqual({
      name: ' gpt-4o ',
      alias: '',
      contextWindowTokens: '64000',
    })
  })
})

describe('existing-data round trips', () => {
  it('loads a stored provider into rows and back without changing models', () => {
    const stored = [
      {
        name: 'gpt-4o',
        alias: '4o',
        context_window_tokens: 128_000,
      },
      { name: 'gemini-2.5-pro' },
    ]

    expect(rowsToModels(modelsToRows(stored))).toEqual(stored)
  })

  it('keeps aliases and capacities distinct across repeated round trips', () => {
    const stored = [
      {
        name: 'shared-model',
        alias: 'a',
        context_window_tokens: 32_000,
      },
      {
        name: 'shared-model',
        alias: 'b',
        context_window_tokens: 64_000,
      },
    ]
    const once = modelsToRows(stored)

    expect(once.map((row) => row.alias)).toEqual(['a', 'b'])
    expect(once.map((row) => row.contextWindowTokens)).toEqual([
      '32000',
      '64000',
    ])
    expect(rowsToModels(modelsToRows(rowsToModels(once)))).toEqual(stored)
  })

  it("does not invent capacity for a provider's legacy model list", () => {
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
    expect(rowsToModels(rowsAfterAdd.slice(0, -1))).toEqual(existing)
  })

  it('attaches per-row errors to the specific row with the empty model ID', () => {
    const parsed = modelRowSchema.safeParse({
      name: '',
      alias: 'x',
      contextWindowTokens: '',
    })
    expect(parsed.success).toBe(false)
    if (!parsed.success) {
      expect(parsed.error.issues[0].message).toBe(MODEL_ID_REQUIRED_MESSAGE)
    }

    const filled = modelRowSchema.safeParse({
      name: 'm',
      alias: '',
      contextWindowTokens: '',
    })
    expect(filled.success).toBe(true)
  })

  it('rejects whitespace-only model IDs so every submitted ID is non-empty', () => {
    const parsed = modelsSchema.safeParse([
      { name: '   ', alias: '', contextWindowTokens: '' },
    ])
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
      {
        name: ' gemini-2.5-pro ',
        alias: ' pro ',
        contextWindowTokens: ' 1048576 ',
      },
    ])
    expect(parsed).toEqual([
      {
        name: 'gemini-2.5-pro',
        alias: 'pro',
        contextWindowTokens: '1048576',
      },
    ])
  })
})

describe('context window validation', () => {
  it.each(['', '0', '1', String(MAX_CONTEXT_WINDOW_TOKENS)])(
    'accepts %j as blank or unsigned uint32 input',
    (contextWindowTokens) => {
      expect(
        modelRowSchema.safeParse({
          name: 'model',
          alias: '',
          contextWindowTokens,
        }).success
      ).toBe(true)
    }
  )

  it.each(['-1', '1.5', '1e5', 'not-a-number', '4294967296'])(
    'rejects invalid capacity %j',
    (contextWindowTokens) => {
      const parsed = modelRowSchema.safeParse({
        name: 'model',
        alias: '',
        contextWindowTokens,
      })
      expect(parsed.success).toBe(false)
      if (!parsed.success) {
        expect(parsed.error.issues[0].message).toBe(
          CONTEXT_WINDOW_TOKENS_MESSAGE
        )
        expect(parsed.error.issues[0].path).toContain('contextWindowTokens')
      }
    }
  )
})
