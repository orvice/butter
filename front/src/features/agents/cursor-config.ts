import { z, type RefinementCtx } from 'zod'
import type { Agent, AgentConfig, CursorAgentConfig } from '@/types/api'

export type CursorAgentFormValues = {
  butterboxId: string
  workingDir: string
  model: string
  mode: string
  maxRunSeconds: string
}

export const EMPTY_CURSOR_AGENT_FORM_VALUES: CursorAgentFormValues = {
  butterboxId: '',
  workingDir: '',
  model: '',
  mode: '',
  maxRunSeconds: '',
}

export const cursorAgentFormSchema = z.object({
  butterboxId: z.string(),
  workingDir: z.string(),
  model: z.string(),
  mode: z.string(),
  maxRunSeconds: z.string(),
})

const VALID_CURSOR_MODES = new Set(['', 'agent', 'plan'])

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
  const mode = values.mode.trim()
  if (!VALID_CURSOR_MODES.has(mode)) {
    ctx.addIssue({
      code: 'custom',
      path: ['cursor', 'mode'],
      message: 'Mode must be empty (box default), agent, or plan.',
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

export function buildCursorAgentConfig(values: CursorAgentFormValues): AgentConfig {
  const maxRunSeconds = values.maxRunSeconds.trim()
  return {
    cursor: {
      butterbox_id: values.butterboxId.trim(),
      working_dir: values.workingDir.trim(),
      model: values.model.trim(),
      mode: values.mode.trim(),
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
    mode: config?.mode ?? '',
    maxRunSeconds:
      config?.max_run_seconds === undefined
        ? ''
        : String(config.max_run_seconds),
  }
}
