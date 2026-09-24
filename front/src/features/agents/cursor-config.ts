import { z, type RefinementCtx } from 'zod'
import type {
  Agent,
  AgentConfig,
  CursorAgentConfig,
  CursorAgentMode,
} from '@/types/api'

export const CURSOR_AGENT_MODES: readonly CursorAgentMode[] = ['agent', 'plan']

export type CursorAgentFormValues = {
  butterboxId: string
  workingDir: string
  model: string
  mode: CursorAgentMode
  maxRunSeconds: string
}

export const EMPTY_CURSOR_AGENT_FORM_VALUES: CursorAgentFormValues = {
  butterboxId: '',
  workingDir: '',
  model: '',
  mode: 'agent',
  maxRunSeconds: '',
}

export const cursorAgentFormSchema = z.object({
  butterboxId: z.string(),
  workingDir: z.string(),
  model: z.string(),
  mode: z.enum(['agent', 'plan']),
  maxRunSeconds: z.string(),
})

export function validateCursorAgentForm(
  values: CursorAgentFormValues,
  ctx: RefinementCtx
) {
  if (values.butterboxId.trim() === '') {
    ctx.addIssue({
      code: 'custom',
      path: ['cursor', 'butterboxId'],
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
      path: ['cursor', 'maxRunSeconds'],
      message: 'Use a whole number from 0 to 2147483647.',
    })
  }
}

export function buildCursorAgentConfig(
  values: CursorAgentFormValues
): AgentConfig {
  const maxRunSeconds = values.maxRunSeconds.trim()
  return {
    cursor: {
      butterbox_id: values.butterboxId.trim(),
      working_dir: values.workingDir.trim(),
      model: values.model.trim(),
      mode: values.mode,
      ...(maxRunSeconds === ''
        ? {}
        : { max_run_seconds: Number(maxRunSeconds) }),
    },
  }
}

export function asCursorAgent(
  agent: Agent,
  values: CursorAgentFormValues
): Agent {
  return {
    ...agent,
    type: 'AGENT_TYPE_CURSOR',
    child_agent_ids: [],
    config: buildCursorAgentConfig(values),
  }
}

export function cursorFormValuesFromConfig(
  config?: CursorAgentConfig
): CursorAgentFormValues {
  return {
    butterboxId: config?.butterbox_id ?? '',
    workingDir: config?.working_dir ?? '',
    model: config?.model ?? '',
    mode: config?.mode === 'plan' ? 'plan' : 'agent',
    maxRunSeconds:
      config?.max_run_seconds === undefined
        ? ''
        : String(config.max_run_seconds),
  }
}
