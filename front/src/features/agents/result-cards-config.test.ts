import { describe, expect, it } from 'vitest'
import {
  buildResultCardsConfig,
  EMPTY_RESULT_CARDS_FORM_VALUES,
  resultCardsFormValuesFromConfig,
  supportsResultCards,
} from './result-cards-config'

describe('result cards form mapping', () => {
  it('serializes the default as no config', () => {
    expect(
      buildResultCardsConfig(EMPTY_RESULT_CARDS_FORM_VALUES, 'AGENT_TYPE_LLM')
    ).toBeUndefined()
  })

  it('serializes cards off as a disabled generation', () => {
    expect(
      buildResultCardsConfig({ generation: 'off' }, 'AGENT_TYPE_LLM')
    ).toEqual({ generation: 'RESULT_CARD_GENERATION_DISABLED' })
  })

  it('keeps the policy on composite agents and drops it for box agents', () => {
    const off = { generation: 'off' as const }
    expect(buildResultCardsConfig(off, 'AGENT_TYPE_WORKFLOW')).toEqual({
      generation: 'RESULT_CARD_GENERATION_DISABLED',
    })
    expect(buildResultCardsConfig(off, 'AGENT_TYPE_PI')).toBeUndefined()
    expect(buildResultCardsConfig(off, 'AGENT_TYPE_CURSOR')).toBeUndefined()
  })

  it('reads a stored config back', () => {
    expect(
      resultCardsFormValuesFromConfig({
        generation: 'RESULT_CARD_GENERATION_DISABLED',
      })
    ).toEqual({ generation: 'off' })
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
