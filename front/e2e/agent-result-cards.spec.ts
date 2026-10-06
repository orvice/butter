import { expect, test, type Page, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import {
  CreateAgentRequestSchema,
  CreateAgentResponseSchema,
  GetAgentResponseSchema,
  ListModelProvidersResponseSchema,
  UpdateAgentRequestSchema,
  UpdateAgentResponseSchema,
  type CreateAgentRequest,
  type UpdateAgentRequest,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  AgentSchema,
  AgentType,
  ModelProviderSchema,
  ResultCardGeneration,
  ResultCardPresentation,
  type Agent,
  type ModelProvider,
} from '../src/gen/agents/v1/agent_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

const SEED_AGENT: Agent = create(AgentSchema, {
  name: 'cards-agent',
  agentId: 'cards-agent',
  description: 'Result cards test agent',
  type: AgentType.LLM,
  config: { model: 'gpt-4o', instruction: 'Be concise.' },
})

const SEED_PROVIDER: ModelProvider = create(ModelProviderSchema, {
  name: 'openai',
  type: 'openai',
  models: [{ name: 'gpt-4o', alias: '4o' }],
})

const ALLOWED = /^Allowed/
const OFF = /^Off for this agent and its sub-agents/

function decodeRequest<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route,
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

async function setupAgentRoutes(page: Page) {
  const state: { agent: Agent } = { agent: SEED_AGENT }
  let updated: UpdateAgentRequest | undefined
  let created: CreateAgentRequest | undefined

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
      updated = decodeRequest(UpdateAgentRequestSchema, route)
      if (updated.agent) state.agent = updated.agent
      await fulfillProto(route, UpdateAgentResponseSchema, { agent: state.agent })
      resolveSaved()
      return true
    }

    if (url.includes('AgentService/CreateAgent')) {
      created = decodeRequest(CreateAgentRequestSchema, route)
      await fulfillProto(route, CreateAgentResponseSchema, { agent: created.agent })
      resolveSaved()
      return true
    }

    return false
  })

  return {
    saved,
    updateRequest: () => updated,
    createRequest: () => created,
  }
}

test('turns result cards off for an agent, saves the policy, and reloads it', async ({ page }) => {
  const ctx = await setupAgentRoutes(page)

  await page.goto('/agents/cards-agent/edit')

  const policy = page.getByRole('radiogroup', { name: 'Result cards' })
  await expect(policy.getByRole('radio', { name: ALLOWED })).toBeChecked()

  await policy.getByRole('radio', { name: OFF }).click()
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await ctx.saved
  await expect(page).toHaveURL(/\/agents$/)

  expect(ctx.updateRequest()?.agent?.config?.resultCards?.generation).toBe(
    ResultCardGeneration.DISABLED,
  )

  // The mocked persisted response is used on the next page load, proving
  // the edit form maps the stored policy back to its control.
  await page.goto('/agents/cards-agent/edit')
  await expect(
    page.getByRole('radiogroup', { name: 'Result cards' }).getByRole('radio', { name: OFF }),
  ).toBeChecked()
})

test('asks the models to prefer cards, and locks the choice while cards are off', async ({ page }) => {
  const ctx = await setupAgentRoutes(page)

  await page.goto('/agents/cards-agent/edit')

  const policy = page.getByRole('radiogroup', { name: 'Result cards' })
  const presentation = page.getByRole('combobox', { name: 'Presentation' })
  await expect(presentation).toHaveText('Inherit')

  await policy.getByRole('radio', { name: OFF }).click()
  await expect(presentation).toBeDisabled()
  await policy.getByRole('radio', { name: ALLOWED }).click()
  await expect(presentation).toBeEnabled()

  await presentation.click()
  await page.getByRole('option', { name: 'Prefer cards' }).click()
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await ctx.saved

  const cards = ctx.updateRequest()?.agent?.config?.resultCards
  expect(cards?.presentation).toBe(ResultCardPresentation.PREFERRED)
  expect(cards?.generation).toBe(ResultCardGeneration.UNSPECIFIED)

  await page.goto('/agents/cards-agent/edit')
  await expect(page.getByRole('combobox', { name: 'Presentation' })).toHaveText('Prefer cards')
})

test('creates an agent with result cards off, and hides the card for box agents', async ({ page }) => {
  const ctx = await setupAgentRoutes(page)

  await page.goto('/agents/create')
  await page.getByLabel('Name', { exact: true }).fill('No cards')

  const policy = page.getByRole('radiogroup', { name: 'Result cards' })
  await expect(policy.getByRole('radio', { name: ALLOWED })).toBeChecked()

  // PI and CURSOR agents reject a Card Policy, so the card is not offered.
  for (const type of ['Pi', 'Cursor', 'LLM']) {
    await page.getByRole('combobox', { name: 'Type' }).click()
    await page.getByRole('option', { name: type, exact: true }).click()
    await expect(policy).toBeVisible({ visible: type === 'LLM' })
  }

  await policy.getByRole('radio', { name: OFF }).click()
  await page.getByRole('button', { name: 'Create Agent' }).click()
  await ctx.saved

  const agent = ctx.createRequest()?.agent
  expect(agent?.agentId).toBe('no-cards')
  expect(agent?.config?.resultCards?.generation).toBe(ResultCardGeneration.DISABLED)
})
