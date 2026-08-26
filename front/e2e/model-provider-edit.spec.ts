import { expect, test, type Page, type Route } from '@playwright/test'
import { fromBinary } from '@bufbuild/protobuf'
import {
  GetModelProviderResponseSchema,
  UpdateModelProviderRequestSchema,
  UpdateModelProviderResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import type { ModelProvider } from '../src/gen/agents/v1/agent_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

// Seed provider: one aliased model plus one plain model (issue #321).
const SEED_PROVIDER: ModelProvider = {
  name: 'openai',
  type: 'openai',
  apiKey: 'sk-test-key',
  baseUrl: 'https://api.example.com/v1',
  models: [
    { name: 'gpt-4o', alias: '4o' },
    { name: 'gemini-2.5-pro', alias: '' },
  ],
  workspaceId: '',
}

function decodeRequest<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

async function setupProviderEdit(page: Page) {
  const state = { provider: SEED_PROVIDER }
  let submitted: UpdateModelProviderRequest | undefined

  let resolveSaved!: () => void
  const saved = new Promise<void>((resolve) => {
    resolveSaved = resolve
  })

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
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

  const rows = page.getByRole('list', { name: 'Models' }).getByRole('listitem')

  // Existing data loads into rows without changing IDs, aliases, or ordering.
  await expect(rows).toHaveCount(2)
  await expect(
    rows.nth(0).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gpt-4o')
  await expect(rows.nth(0).getByPlaceholder('Optional alias')).toHaveValue('4o')
  await expect(
    rows.nth(1).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gemini-2.5-pro')
  await expect(rows.nth(1).getByPlaceholder('Optional alias')).toHaveValue('')
  await expect(page.getByRole('textbox', { name: 'Name' })).toHaveValue(
    'openai'
  )
  await expect(page.getByRole('textbox', { name: 'Name' })).toBeDisabled()

  // Edit the rows: give the plain model an alias...
  await rows.nth(1).getByPlaceholder('Optional alias').fill('pro')
  // ...append a new row with a model ID...
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

  // The submitted payload carries exactly the edited rows, and the untouched
  // provider fields (name, type, credentials) are preserved.
  const request = ctx.submittedRequest()
  const provider = request?.modelProvider
  expect(provider?.name).toBe('openai')
  expect(provider?.type).toBe('openai')
  expect(provider?.apiKey).toBe('sk-test-key')
  expect(provider?.baseUrl).toBe('https://api.example.com/v1')
  expect(provider?.models.map((m) => ({ name: m.name, alias: m.alias }))).toEqual([
    { name: 'gemini-2.5-pro', alias: 'pro' },
    { name: 'gpt-4.1', alias: '' },
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
    rows.nth(1).getByRole('textbox', { name: 'Model ID' })
  ).toHaveValue('gpt-4.1')
  await expect(rows.nth(1).getByPlaceholder('Optional alias')).toHaveValue('')
})

test('blocks submission without at least one model and flags the offending row', async ({
  page,
}) => {
  await setupProviderEdit(page)

  await page.goto('/model-providers/openai/edit')

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
})
