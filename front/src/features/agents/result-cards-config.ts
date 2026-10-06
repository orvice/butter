import { z } from 'zod'
import type {
  AgentConfig,
  ResultCardConfig,
  ResultCardPresentation,
} from '@/types/api'

/** 'allowed' leaves the generation unset, so the agent inherits. */
export type ResultCardsGeneration = 'allowed' | 'off'

/** 'inherit' leaves the presentation unset: the nearest parent decides. */
export type ResultCardsPresentation = 'inherit' | 'auto' | 'preferred'

export interface ResultCardsFormValues {
  generation: ResultCardsGeneration
  presentation: ResultCardsPresentation
}

export const EMPTY_RESULT_CARDS_FORM_VALUES: ResultCardsFormValues = {
  generation: 'allowed',
  presentation: 'inherit',
}

export const resultCardsFormSchema = z.object({
  generation: z.enum(['allowed', 'off']),
  presentation: z.enum(['inherit', 'auto', 'preferred']),
})

const PRESENTATIONS: Record<
  Exclude<ResultCardsPresentation, 'inherit'>,
  ResultCardPresentation
> = {
  auto: 'RESULT_CARD_PRESENTATION_AUTO',
  preferred: 'RESULT_CARD_PRESENTATION_PREFERRED',
}

export function resultCardsFormValuesFromConfig(
  config?: ResultCardConfig
): ResultCardsFormValues {
  let presentation: ResultCardsPresentation = 'inherit'
  if (config?.presentation === 'RESULT_CARD_PRESENTATION_AUTO')
    presentation = 'auto'
  if (config?.presentation === 'RESULT_CARD_PRESENTATION_PREFERRED')
    presentation = 'preferred'
  return {
    generation:
      config?.generation === 'RESULT_CARD_GENERATION_DISABLED'
        ? 'off'
        : 'allowed',
    presentation,
  }
}

/**
 * Serializes the form. The default serializes to no config at all, and the
 * field is dropped for types that reject it. With cards off the
 * presentation means nothing anywhere below, so it is dropped too.
 */
export function buildResultCardsConfig(
  values: ResultCardsFormValues,
  type?: string
): AgentConfig['result_cards'] {
  if (!supportsResultCards(type)) return undefined
  if (values.generation === 'off')
    return { generation: 'RESULT_CARD_GENERATION_DISABLED' }
  if (values.presentation === 'inherit') return undefined
  return { presentation: PRESENTATIONS[values.presentation] }
}

/**
 * PI and CURSOR agents keep their behavior on the ButterBox and reject a
 * Card Policy on write.
 */
export function supportsResultCards(type?: string): boolean {
  return type !== 'AGENT_TYPE_PI' && type !== 'AGENT_TYPE_CURSOR'
}
