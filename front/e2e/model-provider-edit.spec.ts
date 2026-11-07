import { expect, test, type Page, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import {
  CreateModelProviderRequestSchema,
  CreateModelProviderResponseSchema,
  type CreateModelProviderRequest,
  GetModelProviderResponseSchema,
  ListModelProvidersResponseSchema,
  type UpdateModelProviderRequest,
  UpdateModelProviderRequestSchema,
  UpdateModelProviderResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  ModelProviderSchema,
  type ModelProvider,
} from '../src/gen/agents/v1/agent_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

// Seed provider: one aliased model plus one plain model (issue #321).
// Built with create() so nested ModelConfig values carry their $typeName.
const SEED_PROVIDER: ModelProvider = create(ModelProviderSchema, {
  name: 'openai',
  type: 'openai',
  apiKey: 'sk-test-key',
  baseUrl: 'https://api.example.com/v1',
  models: [
    { name: 'gpt-4o', alias: '4o', contextWindowTokens: 128_000 },
    { name: 'gemini-2.5-pro', alias: '', contextWindowTokens: 0 },
  ],
  workspaceId: '',
})

function decodeRequest<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

async function setupProviderEdit(page: Page, provider = SEED_PROVIDER) {
  const state = { provider }
  let submitted: UpdateModelProviderRequest | undefined

  let resolveSaved!: () => void
  const saved = new Promise<void>((resolve) => {
    resolveSaved = resolve
  })

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('ModelProviderService/ListModelProviders')) {
      return fulfillProto(route, ListModelProvidersResponseSchema, {
        modelProviders: [state.provider],
      })
    }
    if (url.includes('ModelProviderService/GetModelProvider')) {
      return fulfillProto(route, GetModelProviderResponseSchema, {
        modelProvider: state.provider,
      })
    }

    if (url.includes('ModelProviderService/UpdateModelProvider')) {
      submitted = decodeRequest(UpdateModelProviderRequestSchema, route)
      if (submitted.modelProvider) state.provider = submitted.modelProvider
      await fulfillProto(route, UpdateModelProviderResponseSchema, {
        modelProvider: state.provider,
      })
      resolveSaved()
      return true
    }

    return false
  })

  return {
    saved,
    submittedRequest: () => submitted,
  }
}

test('edits model rows, submits them to UpdateModelProvider, and reflects persisted rows after reload', async ({
  page,
}) => {
  const ctx = await setupProviderEdit(page)

  await page.goto('/model-providers/openai/edit')
  await expect(page.getByRole('combobox', { name: 'API Protocol' })).toHaveText('Chat Completions')

  const rows = page.getByRole('list', { name: 'Models' }).getByRole('listitem')

  // Existing data loads into rows without changing IDs, aliases, or ordering.
  await expect(rows).toHaveCount(2)
  await expect(
    rows.nth(0).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gpt-4o')
  await expect(rows.nth(0).getByPlaceholder('Optional alias')).toHaveValue('4o')
  await expect(
    rows.nth(0).getByRole('spinbutton', { name: 'Context Window' })
  ).toHaveValue('128000')
  await expect(
    rows.nth(1).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gemini-2.5-pro')
  await expect(rows.nth(1).getByPlaceholder('Optional alias')).toHaveValue('')
  await expect(
    rows.nth(1).getByRole('spinbutton', { name: 'Context Window' })
  ).toHaveValue('')
  await expect(page.getByRole('textbox', { name: 'Name' })).toHaveValue(
    'openai'
  )
  await expect(page.getByRole('textbox', { name: 'Name' })).toBeDisabled()

  // Edit the rows: give the plain model an alias and context capacity...
  await rows.nth(1).getByPlaceholder('Optional alias').fill('pro')
  await rows
    .nth(1)
    .getByRole('spinbutton', { name: 'Context Window' })
    .fill('1048576')
  // ...append a new row with a model ID and no configured capacity...
  await page.getByRole('button', { name: 'Add model' }).click()
  await expect(rows).toHaveCount(3)
  await rows
    .nth(2)
    .getByRole('textbox', { name: 'Model ID' })
    .fill('gpt-4.1')
  // ...and remove the first row; the remaining order must be stable.
  await page.getByRole('button', { name: 'Remove model 1' }).click()
  await expect(rows).toHaveCount(2)
  await expect(
    rows.nth(0).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gemini-2.5-pro')
  await expect(
    rows.nth(1).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gpt-4.1')

  await page.getByRole('button', { name: 'Save' }).click()
  await ctx.saved
  await expect(page).toHaveURL(/\/model-providers$/)

  // The submitted payload carries exactly the edited rows, and the untouched
  // provider fields (name, type, credentials) are preserved.
  const request = ctx.submittedRequest()
  const provider = request?.modelProvider
  expect(provider?.name).toBe('openai')
  expect(provider?.type).toBe('openai')
  expect(provider?.apiKey).toBe('sk-test-key')
  expect(provider?.baseUrl).toBe('https://api.example.com/v1')
  expect(
    provider?.models.map((m) => ({
      name: m.name,
      alias: m.alias,
      contextWindowTokens: m.contextWindowTokens,
    }))
  ).toEqual([
    {
      name: 'gemini-2.5-pro',
      alias: 'pro',
      contextWindowTokens: 1_048_576,
    },
    { name: 'gpt-4.1', alias: '', contextWindowTokens: 0 },
  ])

  // Reload: the persisted models come back as the same structured rows.
  await page.goto('/model-providers/openai/edit')
  await expect(rows).toHaveCount(2)
  await expect(
    rows.nth(0).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gemini-2.5-pro')
  await expect(rows.nth(0).getByPlaceholder('Optional alias')).toHaveValue(
    'pro'
  )
  await expect(
    rows.nth(0).getByRole('spinbutton', { name: 'Context Window' })
  ).toHaveValue('1048576')
  await expect(
    rows.nth(1).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gpt-4.1')
  await expect(rows.nth(1).getByPlaceholder('Optional alias')).toHaveValue('')
  await expect(
    rows.nth(1).getByRole('spinbutton', { name: 'Context Window' })
  ).toHaveValue('')
})

test('selects and persists both OpenAI protocols, including after reload', async ({ page }) => {
  const ctx = await setupProviderEdit(page)
  await page.goto('/model-providers/openai/edit')
  const protocol = page.getByRole('combobox', { name: 'API Protocol' })
  await expect(protocol).toHaveText('Chat Completions')
  await protocol.click()
  await page.getByRole('option', { name: 'Responses API', exact: true }).click()
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await ctx.saved
  expect(ctx.submittedRequest()?.modelProvider?.openaiApi).toBe('responses')
  await expect(page).toHaveURL(/\/model-providers$/)
  await expect(page.getByRole('table').getByText('Responses API', { exact: true })).toBeVisible()

  await page.goto('/model-providers/openai/edit')
  await expect(protocol).toHaveText('Responses API')
  // An unrelated edit must not reset a persisted Responses provider.
  await page.getByRole('textbox', { name: 'Base URL' }).fill('https://changed.example/v1')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => ctx.submittedRequest()?.modelProvider?.baseUrl).toBe('https://changed.example/v1')
  expect(ctx.submittedRequest()?.modelProvider?.openaiApi).toBe('responses')
  await expect(page).toHaveURL(/\/model-providers$/)

  await page.goto('/model-providers/openai/edit')
  await expect(protocol).toHaveText('Responses API')
  await protocol.click()
  await page.getByRole('option', { name: 'Chat Completions', exact: true }).click()
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => ctx.submittedRequest()?.modelProvider?.openaiApi).toBe('chat_completions')
  await expect(page).toHaveURL(/\/model-providers$/)
  await page.goto('/model-providers/openai/edit')
  await expect(protocol).toHaveText('Chat Completions')
})

test('hides and clears the OpenAI protocol when switching to Gemini', async ({ page }) => {
  const ctx = await setupProviderEdit(page, create(ModelProviderSchema, { ...SEED_PROVIDER, openaiApi: 'responses' }))
  await page.goto('/model-providers/openai/edit')
  await expect(page.getByRole('combobox', { name: 'API Protocol' })).toHaveText('Responses API')
  await page.getByRole('combobox', { name: 'Type', exact: true }).click()
  await page.getByRole('option', { name: 'Gemini', exact: true }).click()
  await expect(page.getByRole('combobox', { name: 'API Protocol' })).toHaveCount(0)
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await ctx.saved
  expect(ctx.submittedRequest()?.modelProvider?.type).toBe('gemini')
  expect(ctx.submittedRequest()?.modelProvider?.openaiApi).toBe('')
  await expect(page).toHaveURL(/\/model-providers$/)
  await page.goto('/model-providers/openai/edit')
  await expect(page.getByRole('combobox', { name: 'Type', exact: true })).toHaveText('Gemini')
  await expect(page.getByRole('combobox', { name: 'API Protocol' })).toHaveCount(0)
})

test('creates an OpenAI provider with the selected protocol', async ({ page }) => {
  let submitted: CreateModelProviderRequest | undefined
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('ModelProviderService/CreateModelProvider')) {
      submitted = decodeRequest(CreateModelProviderRequestSchema, route)
      return fulfillProto(route, CreateModelProviderResponseSchema, { modelProvider: submitted.modelProvider })
    }
    return false
  })
  await page.goto('/model-providers/create')
  await page.getByRole('textbox', { name: 'Name', exact: true }).fill('responses-provider')
  const protocol = page.getByRole('combobox', { name: 'API Protocol' })
  await expect(protocol).toHaveText('Chat Completions')
  await protocol.click()
  await page.getByRole('option', { name: 'Responses API', exact: true }).click()
  await page.getByRole('button', { name: 'Add model' }).click()
  await page.getByRole('textbox', { name: 'Model ID' }).fill('gpt-4o')
  await page.getByRole('button', { name: 'Create', exact: true }).click()
  await expect.poll(() => submitted?.modelProvider?.openaiApi).toBe('responses')
  expect(submitted?.modelProvider?.name).toBe('responses-provider')
  expect(submitted?.modelProvider?.models[0]?.name).toBe('gpt-4o')
  await expect(page).toHaveURL(/\/model-providers$/)
})

test('blocks submission without at least one model and flags the offending row', async ({
  page,
}) => {
  await setupProviderEdit(page)

  await page.goto('/model-providers/openai/edit')

  const rows = page.getByRole('list', { name: 'Models' }).getByRole('listitem')
  const removeFirst = page.getByRole('button', { name: 'Remove model 1' })
  await removeFirst.click()
  await removeFirst.click() // indices shift; row 2 is now row 1

  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByRole('alert')).toHaveText(
    'At least one model is required'
  )

  // Adding an empty row back moves the error onto that specific row.
  await page.getByRole('button', { name: 'Add model' }).click()
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByText('At least one model is required')).toBeHidden()
  await expect(page.getByText('Model ID is required')).toBeVisible()

  await rows
    .nth(0)
    .getByRole('textbox', { name: 'Model ID' })
    .fill('invalid-capacity')
  await rows
    .nth(0)
    .getByRole('spinbutton', { name: 'Context Window' })
    .fill('-1')
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(
    page.getByText(
      'Context Window must be a whole number from 0 to 4,294,967,295'
    )
  ).toBeVisible()
})
