import { describe, expect, it } from 'vitest'
import {
  buildResultCardsConfig,
  EMPTY_RESULT_CARDS_FORM_VALUES,
  resultCardsFormValuesFromConfig,
  supportsResultCards,
  type ResultCardsFormValues,
} from './result-cards-config'

function values(
  overrides: Partial<ResultCardsFormValues>
): ResultCardsFormValues {
  return { ...EMPTY_RESULT_CARDS_FORM_VALUES, ...overrides }
}

describe('result cards form mapping', () => {
  it('serializes the default as no config', () => {
    expect(
      buildResultCardsConfig(EMPTY_RESULT_CARDS_FORM_VALUES, 'AGENT_TYPE_LLM')
    ).toBeUndefined()
  })

  it('serializes cards off as a disabled generation', () => {
    expect(
      buildResultCardsConfig(values({ generation: 'off' }), 'AGENT_TYPE_LLM')
    ).toEqual({ generation: 'RESULT_CARD_GENERATION_DISABLED' })
  })

  it('serializes an explicit presentation', () => {
    expect(
      buildResultCardsConfig(
        values({ presentation: 'preferred' }),
        'AGENT_TYPE_LLM'
      )
    ).toEqual({ presentation: 'RESULT_CARD_PRESENTATION_PREFERRED' })
    expect(
      buildResultCardsConfig(values({ presentation: 'auto' }), 'AGENT_TYPE_LLM')
    ).toEqual({ presentation: 'RESULT_CARD_PRESENTATION_AUTO' })
  })

  it('keeps a chosen presentation while cards are off', () => {
    expect(
      buildResultCardsConfig(
        values({ generation: 'off', presentation: 'preferred' }),
        'AGENT_TYPE_LLM'
      )
    ).toEqual({
      generation: 'RESULT_CARD_GENERATION_DISABLED',
      presentation: 'RESULT_CARD_PRESENTATION_PREFERRED',
    })
  })

  it('keeps the policy on composite agents and drops it for box agents', () => {
    const off = values({ generation: 'off' })
    expect(buildResultCardsConfig(off, 'AGENT_TYPE_WORKFLOW')).toEqual({
      generation: 'RESULT_CARD_GENERATION_DISABLED',
    })
    const preferred = values({ presentation: 'preferred' })
    expect(buildResultCardsConfig(off, 'AGENT_TYPE_PI')).toBeUndefined()
    expect(
      buildResultCardsConfig(preferred, 'AGENT_TYPE_CURSOR')
    ).toBeUndefined()
  })

  it('reads a stored config back', () => {
    expect(
      resultCardsFormValuesFromConfig({
        generation: 'RESULT_CARD_GENERATION_DISABLED',
      })
    ).toEqual({ generation: 'off', presentation: 'inherit' })
    expect(
      resultCardsFormValuesFromConfig({
        presentation: 'RESULT_CARD_PRESENTATION_PREFERRED',
      })
    ).toEqual({ generation: 'allowed', presentation: 'preferred' })
    expect(
      resultCardsFormValuesFromConfig({
        presentation: 'RESULT_CARD_PRESENTATION_AUTO',
      })
    ).toEqual({ generation: 'allowed', presentation: 'auto' })
    expect(
      resultCardsFormValuesFromConfig({
        generation: 'RESULT_CARD_GENERATION_DISABLED',
        presentation: 'RESULT_CARD_PRESENTATION_PREFERRED',
      })
    ).toEqual({ generation: 'off', presentation: 'preferred' })
    expect(resultCardsFormValuesFromConfig({})).toEqual(
      EMPTY_RESULT_CARDS_FORM_VALUES
    )
    expect(resultCardsFormValuesFromConfig(undefined)).toEqual(
      EMPTY_RESULT_CARDS_FORM_VALUES
    )
  })

  it('knows which types support a Card Policy', () => {
    expect(supportsResultCards('AGENT_TYPE_LLM')).toBe(true)
    expect(supportsResultCards('AGENT_TYPE_SEQUENTIAL')).toBe(true)
    expect(supportsResultCards(undefined)).toBe(true)
    expect(supportsResultCards('AGENT_TYPE_PI')).toBe(false)
    expect(supportsResultCards('AGENT_TYPE_CURSOR')).toBe(false)
  })
})
