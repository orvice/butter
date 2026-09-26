import { expect, test, type Route } from '@playwright/test'
import { create, fromBinary, type DescMessage } from '@bufbuild/protobuf'
import {
  DeleteWorkspaceMemoryConfigResponseSchema,
  GetWorkspaceMemoryConfigResponseSchema,
  PutWorkspaceMemoryConfigRequestSchema,
  PutWorkspaceMemoryConfigResponseSchema,
  TestWorkspaceMemoryConnectionResponseSchema,
  WorkspaceMemoryConfigSchema,
  type PutWorkspaceMemoryConfigRequest,
  type WorkspaceMemoryConfig,
} from '../src/gen/agents/v1/workspace_memory_pb'
import {
  GetAgentResponseSchema,
  ListModelProvidersResponseSchema,
  UpdateAgentRequestSchema,
  UpdateAgentResponseSchema,
  type UpdateAgentRequest,
} from '../src/gen/agents/v1/agent_service_pb'
import { AgentSchema, AgentType, type Agent } from '../src/gen/agents/v1/agent_pb'
import { fulfillConnectError, fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

function decode<T extends DescMessage>(schema: T, route: Route) {
  return fromBinary(schema, route.request().postDataBuffer() ?? Buffer.alloc(0))
}

test('connects, tests, and removes the workspace mem0 server', async ({ page }) => {
  let config: WorkspaceMemoryConfig | undefined
  const puts: PutWorkspaceMemoryConfigRequest[] = []
  let deleted = false

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('WorkspaceMemoryConfigService/GetWorkspaceMemoryConfig')) {
      return fulfillProto(route, GetWorkspaceMemoryConfigResponseSchema, { config })
    }
    if (url.includes('WorkspaceMemoryConfigService/PutWorkspaceMemoryConfig')) {
      const req = decode(PutWorkspaceMemoryConfigRequestSchema, route)
      puts.push(req)
      if (req.apiKey === 'm0sk_rejected') {
        return fulfillConnectError(route, 'failed_precondition', 'mem0 server rejected the API key')
      }
      config = create(WorkspaceMemoryConfigSchema, {
        workspaceId: 'default',
        baseUrl: req.baseUrl.replace(/\/+$/, ''),
        enabled: req.enabled,
        credentialSet: req.apiKey === undefined ? !!config?.credentialSet : req.apiKey !== '',
      })
      return fulfillProto(route, PutWorkspaceMemoryConfigResponseSchema, {
        config,
        warning: req.baseUrl.includes('offline') ? 'mem0: /search: connection refused' : '',
      })
    }
    if (url.includes('WorkspaceMemoryConfigService/TestWorkspaceMemoryConnection')) {
      return fulfillProto(route, TestWorkspaceMemoryConnectionResponseSchema, { ok: true })
    }
    if (url.includes('WorkspaceMemoryConfigService/DeleteWorkspaceMemoryConfig')) {
      deleted = true
      config = undefined
      return fulfillProto(route, DeleteWorkspaceMemoryConfigResponseSchema, {})
    }
    return false
  })

  await page.goto('/memory')
  await expect(page.getByRole('heading', { name: 'Memory' })).toBeVisible()
  await expect(page.getByText('Workspace Memory is shared')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Test connection' })).toBeDisabled()

  // A key the server refuses is reported and nothing is saved.
  await page.getByLabel('Base URL').fill('https://mem0.example.com')
  await page.getByLabel('API key').fill('m0sk_rejected')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByText('mem0 server rejected the API key', { exact: true })).toBeVisible()

  // A good key saves; the page reports the stored key without echoing it.
  await page.getByLabel('API key').fill('m0sk_good')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByText('API key set')).toBeVisible()
  expect(puts.at(-1)).toMatchObject({ baseUrl: 'https://mem0.example.com', enabled: true, apiKey: 'm0sk_good' })
  await expect(page.getByLabel('Replace API key')).toHaveValue('')

  // Saving without touching the key keeps it (api_key absent), and an
  // unreachable server saves with a visible warning.
  await page.getByLabel('Base URL').fill('https://offline.example.com')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByText('Saved, but the connection check failed')).toBeVisible()
  expect(puts.at(-1)?.apiKey).toBeUndefined()

  await page.getByRole('button', { name: 'Test connection' }).click()
  await expect(page.getByText('The mem0 server answered and accepted the API key.')).toBeVisible()

  // Clearing the key sends an explicit empty key.
  await page.getByLabel('Clear the stored API key on save').check()
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByText('No API key')).toBeVisible()
  expect(puts.at(-1)?.apiKey).toBe('')

  await page.getByRole('button', { name: 'Remove' }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Remove' }).click()
  await expect(page.getByRole('button', { name: 'Test connection' })).toBeDisabled()
  expect(deleted).toBe(true)
})

const MEMORY_AGENT: Agent = create(AgentSchema, {
  name: 'memory-agent',
  agentId: 'memory-agent',
  type: AgentType.LLM,
  config: {
    model: 'gpt-4o',
    instruction: 'Help.',
    memory: { enabled: true, enableTools: true, topK: 8 },
  },
})

test('edits an agent memory config and flags an unconfigured workspace', async ({ page }) => {
  let agent = MEMORY_AGENT
  let submitted: UpdateAgentRequest | undefined
  let resolveSaved!: () => void
  const saved = new Promise<void>((resolve) => {
    resolveSaved = resolve
  })

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('WorkspaceMemoryConfigService/GetWorkspaceMemoryConfig')) {
      return fulfillProto(route, GetWorkspaceMemoryConfigResponseSchema, {})
    }
    if (url.includes('AgentService/GetAgent')) {
      return fulfillProto(route, GetAgentResponseSchema, { agent })
    }
    if (url.includes('ModelProviderService/ListModelProviders')) {
      return fulfillProto(route, ListModelProvidersResponseSchema, {
        modelProviders: [{ name: 'openai', type: 'openai', models: [{ name: 'gpt-4o', alias: '4o' }] }],
      })
    }
    if (url.includes('AgentService/UpdateAgent')) {
      submitted = decode(UpdateAgentRequestSchema, route)
      if (submitted.agent) agent = submitted.agent
      await fulfillProto(route, UpdateAgentResponseSchema, { agent })
      resolveSaved()
      return true
    }
    return false
  })

  await page.goto('/agents/memory-agent/edit')
  await expect(page.getByText('This workspace has no enabled mem0 connection')).toBeVisible()
  await expect(page.getByLabel('Enable memory')).toBeChecked()
  await expect(page.getByLabel('Memory tools')).toBeChecked()
  await expect(page.getByLabel('Memories per turn')).toHaveValue('8')

  // Out-of-range values block the save.
  await page.getByLabel('Relevance threshold').fill('1.5')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByText('Use a number from 0 to 1.')).toBeVisible()
  expect(submitted).toBeUndefined()

  await page.getByLabel('Relevance threshold').fill('0.4')
  await page.getByLabel('Auto capture').click()
  await page.getByLabel('Allow agent-scope writes').click()
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await saved

  const memory = submitted?.agent?.config?.memory
  expect(memory?.enabled).toBe(true)
  expect(memory?.disableAutoCapture).toBe(true)
  expect(memory?.disableAutoRecall).toBe(false)
  expect(memory?.enableTools).toBe(true)
  expect(memory?.allowAgentScopeWrite).toBe(true)
  expect(memory?.topK).toBe(8)
  expect(memory?.threshold).toBeCloseTo(0.4)
})
