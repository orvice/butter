import { expect, test, type Page, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import {
  GetAgentResponseSchema,
  ListModelProvidersResponseSchema,
  UpdateAgentRequestSchema,
  UpdateAgentResponseSchema,
  type UpdateAgentRequest,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  AgentSchema,
  AgentType,
  ContextGuardStrategy,
  ModelProviderSchema,
  type Agent,
  type ModelProvider,
} from '../src/gen/agents/v1/agent_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

const SEED_AGENT: Agent = create(AgentSchema, {
  name: 'context-agent',
  agentId: 'context-agent',
  description: 'Context test agent',
  type: AgentType.LLM,
  config: {
    model: 'gpt-4o',
    instruction: 'Be concise.',
    contextGuard: {
      strategy: ContextGuardStrategy.THRESHOLD,
      maxTokens: 32000,
    },
  },
})

const SEED_PROVIDER: ModelProvider = create(ModelProviderSchema, {
  name: 'openai',
  type: 'openai',
  models: [{ name: 'gpt-4o', alias: '4o' }],
})

function decodeRequest<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route,
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

async function setupAgentEdit(page: Page) {
  const state: { agent: Agent } = { agent: SEED_AGENT }
  let submitted: UpdateAgentRequest | undefined

  let resolveSaved!: () => void
  const saved = new Promise<void>((resolve) => {
    resolveSaved = resolve
  })

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/GetAgent')) {
      return fulfillProto(route, GetAgentResponseSchema, { agent: state.agent })
    }

    if (url.includes('ModelProviderService/ListModelProviders')) {
      return fulfillProto(route, ListModelProvidersResponseSchema, {
        modelProviders: [SEED_PROVIDER],
      })
    }

    if (url.includes('AgentService/UpdateAgent')) {
      submitted = decodeRequest(UpdateAgentRequestSchema, route)
      if (submitted.agent) state.agent = submitted.agent
      await fulfillProto(route, UpdateAgentResponseSchema, { agent: state.agent })
      resolveSaved()
      return true
    }

    return false
  })

  return {
    saved,
    submittedRequest: () => submitted,
    currentAgent: () => state.agent,
  }
}

test('edits ContextGuard modes, persists the valid fields, and reloads them', async ({ page }) => {
  const ctx = await setupAgentEdit(page)

  await page.goto('/agents/context-agent/edit')

  await expect(page.getByText('Context Guard')).toBeVisible()
  await expect(page.getByRole('radio', { name: /Token Threshold/ })).toBeChecked()
  await expect(page.getByLabel('Context window override (tokens)')).toHaveValue('32000')

  // Switching strategy hides and clears the threshold-only override.
  await page.getByRole('radio', { name: /Sliding Window/ }).click()
  await expect(page.getByLabel('Context window override (tokens)')).toBeHidden()
  await page.getByLabel('Maximum content entries').fill('6')

  // Switching back starts with a clean threshold field, then saves a new
  // Agent Context Override through the normal Agent update RPC.
  await page.getByRole('radio', { name: /Token Threshold/ }).click()
  await expect(page.getByLabel('Maximum content entries')).toBeHidden()
  await page.getByLabel('Context window override (tokens)').fill('64000')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await ctx.saved

  const request = ctx.submittedRequest()
  const guard = request?.agent?.config?.contextGuard
  expect(guard?.strategy).toBe(ContextGuardStrategy.THRESHOLD)
  expect(guard?.maxTokens).toBe(64000)
  expect(guard?.maxTurns).toBe(0)

  // The mocked persisted response is used on the next page load, proving the
  // edit form maps the nested wire value back to the threshold input.
  await page.goto('/agents/context-agent/edit')
  await expect(page.getByRole('radio', { name: /Token Threshold/ })).toBeChecked()
  await expect(page.getByLabel('Context window override (tokens)')).toHaveValue('64000')
  await expect(page.getByLabel('Maximum content entries')).toBeHidden()
  expect(ctx.currentAgent().config?.contextGuard?.maxTokens).toBe(64000)
})
