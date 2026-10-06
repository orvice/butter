import { z } from 'zod'
import type {
  AgentConfig,
  ResultCardConfig,
  ResultCardPresentation,
} from '@/types/api'

/** 'allowed' leaves the generation unset, so the agent inherits. */
export type CardGenerationChoice = 'allowed' | 'off'

/** 'inherit' leaves the presentation unset: the nearest parent decides. */
export type CardPresentationChoice = 'inherit' | 'auto' | 'preferred'

export interface ResultCardsFormValues {
  generation: CardGenerationChoice
  presentation: CardPresentationChoice
}

export const EMPTY_RESULT_CARDS_FORM_VALUES: ResultCardsFormValues = {
  generation: 'allowed',
  presentation: 'inherit',
}

export const resultCardsFormSchema = z.object({
  generation: z.enum(['allowed', 'off']),
  presentation: z.enum(['inherit', 'auto', 'preferred']),
})

/** The explicit presentations, read in both directions. */
const PRESENTATION_VALUES = {
  auto: 'RESULT_CARD_PRESENTATION_AUTO',
  preferred: 'RESULT_CARD_PRESENTATION_PREFERRED',
} as const satisfies Record<
  Exclude<CardPresentationChoice, 'inherit'>,
  ResultCardPresentation
>

const EXPLICIT_PRESENTATIONS = Object.keys(PRESENTATION_VALUES) as Array<
  keyof typeof PRESENTATION_VALUES
>

export function resultCardsFormValuesFromConfig(
  config?: ResultCardConfig
): ResultCardsFormValues {
  return {
    generation:
      config?.generation === 'RESULT_CARD_GENERATION_DISABLED'
        ? 'off'
        : 'allowed',
    presentation:
      EXPLICIT_PRESENTATIONS.find(
        (choice) => PRESENTATION_VALUES[choice] === config?.presentation
      ) ?? 'inherit',
  }
}

/**
 * Serializes the form. The default serializes to no config at all, and the
 * field is dropped for types that reject it. A presentation chosen before
 * cards were turned off is kept: it does nothing while they are off, and it
 * is still there when they are turned back on.
 */
export function buildResultCardsConfig(
  values: ResultCardsFormValues,
  type?: string
): AgentConfig['result_cards'] {
  if (!supportsResultCards(type)) return undefined
  const config: ResultCardConfig = {}
  if (values.generation === 'off')
    config.generation = 'RESULT_CARD_GENERATION_DISABLED'
  if (values.presentation !== 'inherit')
    config.presentation = PRESENTATION_VALUES[values.presentation]
  return Object.keys(config).length > 0 ? config : undefined
}

/**
 * PI and CURSOR agents keep their behavior on the ButterBox and reject a
 * Card Policy on write.
 */
export function supportsResultCards(type?: string): boolean {
  return type !== 'AGENT_TYPE_PI' && type !== 'AGENT_TYPE_CURSOR'
}
