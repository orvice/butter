import { z, type RefinementCtx } from 'zod'
import type { Agent, AgentConfig, PiAgentConfig } from '@/types/api'

export type PiAgentFormValues = {
  butterboxId: string
  workingDir: string
  provider: string
  model: string
  thinkingLevel: string
  maxRunSeconds: string
}

export const EMPTY_PI_AGENT_FORM_VALUES: PiAgentFormValues = {
  butterboxId: '',
  workingDir: '',
  provider: '',
  model: '',
  thinkingLevel: '',
  maxRunSeconds: '',
}

export const piAgentFormSchema = z.object({
  butterboxId: z.string(),
  workingDir: z.string(),
  provider: z.string(),
  model: z.string(),
  thinkingLevel: z.string(),
  maxRunSeconds: z.string(),
})

export function validatePiAgentForm(
  values: PiAgentFormValues,
  ctx: RefinementCtx
) {
  if (values.butterboxId.trim() === '') {
    ctx.addIssue({
      code: 'custom',
      path: ['pi', 'butterboxId'],
      message: 'Select a ButterBox.',
    })
  }
  const maxRunSeconds = values.maxRunSeconds.trim()
  if (
    maxRunSeconds !== '' &&
    (!/^\d+$/.test(maxRunSeconds) || Number(maxRunSeconds) > 2_147_483_647)
  ) {
    ctx.addIssue({
      code: 'custom',
      path: ['pi', 'maxRunSeconds'],
      message: 'Use a whole number from 0 to 2147483647.',
    })
  }
}

export function buildPiAgentConfig(values: PiAgentFormValues): AgentConfig {
  const maxRunSeconds = values.maxRunSeconds.trim()
  return {
    pi: {
      butterbox_id: values.butterboxId.trim(),
      working_dir: values.workingDir.trim(),
      provider: values.provider.trim(),
      model: values.model.trim(),
      thinking_level: values.thinkingLevel.trim(),
      ...(maxRunSeconds === ''
        ? {}
        : { max_run_seconds: Number(maxRunSeconds) }),
    },
  }
}

export function asPiAgent(
  agent: Agent,
  values: PiAgentFormValues
): Agent {
  return {
    ...agent,
    type: 'AGENT_TYPE_PI',
    child_agent_ids: [],
    config: buildPiAgentConfig(values),
  }
}

export type PiModelCatalogEntry = {
  id: string
  provider: string
}

export function piModelCatalogKey(model: PiModelCatalogEntry): string {
  return JSON.stringify([model.provider, model.id])
}

export function piModelFromCatalogKey<T extends PiModelCatalogEntry>(
  models: readonly T[],
  key: string
): T | undefined {
  return models.find((model) => piModelCatalogKey(model) === key)
}

export function piFormValuesFromConfig(
  config?: PiAgentConfig
): PiAgentFormValues {
  return {
    butterboxId: config?.butterbox_id ?? '',
    workingDir: config?.working_dir ?? '',
    provider: config?.provider ?? '',
    model: config?.model ?? '',
    thinkingLevel: config?.thinking_level ?? '',
    maxRunSeconds:
      config?.max_run_seconds === undefined
        ? ''
        : String(config.max_run_seconds),
  }
}
