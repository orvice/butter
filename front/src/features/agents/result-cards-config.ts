import { z } from 'zod'
import type { AgentConfig, ResultCardConfig } from '@/types/api'

/** 'allowed' leaves the generation unset, so the agent inherits. */
export type ResultCardsGeneration = 'allowed' | 'off'

export interface ResultCardsFormValues {
  generation: ResultCardsGeneration
}

export const EMPTY_RESULT_CARDS_FORM_VALUES: ResultCardsFormValues = {
  generation: 'allowed',
}

export const resultCardsFormSchema = z.object({
  generation: z.enum(['allowed', 'off']),
})

export function resultCardsFormValuesFromConfig(
  config?: ResultCardConfig
): ResultCardsFormValues {
  return {
    generation:
      config?.generation === 'RESULT_CARD_GENERATION_DISABLED'
        ? 'off'
        : 'allowed',
  }
}

/**
 * Serializes the form. The default serializes to no config at all, and the
 * field is dropped for types that reject it.
 */
export function buildResultCardsConfig(
  values: ResultCardsFormValues,
  type?: string
): AgentConfig['result_cards'] {
  if (values.generation !== 'off' || !supportsResultCards(type))
    return undefined
  return { generation: 'RESULT_CARD_GENERATION_DISABLED' }
}

/**
 * PI and CURSOR agents keep their behavior on the ButterBox and reject a
 * Card Policy on write.
 */
export function supportsResultCards(type?: string): boolean {
  return type !== 'AGENT_TYPE_PI' && type !== 'AGENT_TYPE_CURSOR'
}
