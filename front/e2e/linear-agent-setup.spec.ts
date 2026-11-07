import { expect, test, type Page, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import {
  GetAgentResponseSchema,
  ListAgentsResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  AgentLifecycleStatus,
  AgentSchema,
  AgentType,
  type Agent,
} from '../src/gen/agents/v1/agent_pb'
import {
  CreateLinearAppRequestSchema,
  CreateLinearAppResponseSchema,
  GetLinearAppResponseSchema,
  LinearAppSchema,
  ListLinearAppsResponseSchema,
  ListLinearInstallationsResponseSchema,
  UpdateLinearAppRequestSchema,
  UpdateLinearAppResponseSchema,
  type CreateLinearAppRequest,
  type LinearApp,
  type UpdateLinearAppRequest,
} from '../src/gen/agents/v1/linear_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const AGENT_A: Agent = create(AgentSchema, {
  agentId: 'agent-a',
  name: 'Agent A',
  type: AgentType.LLM,
  lifecycleStatus: AgentLifecycleStatus.ACTIVE,
})

const AGENT_B: Agent = create(AgentSchema, {
  agentId: 'agent-b',
  name: 'Agent B',
  type: AgentType.LLM,
  lifecycleStatus: AgentLifecycleStatus.ACTIVE,
})

type LinearAppPatch = Partial<Omit<LinearApp, '$typeName' | '$unknown'>>

function linearApp(overrides: LinearAppPatch = {}): LinearApp {
  return create(LinearAppSchema, {
    id: 'app-a',
    displayName: 'Support A',
    clientId: 'lin-client-a',
    agentId: 'agent-a',
    revision: 2n,
    clientSecretSet: true,
    webhookSecretSet: true,
    callbackUrl: 'https://butter.example.com/api/linear/oauth/callback',
    webhookUrl: 'https://butter.example.com/api/linear/webhook/app-a',
    ...overrides,
  })
}

/** cloneApp returns a new App message carrying app plus the patch. */
function cloneApp(app: LinearApp, patch: LinearAppPatch = {}): LinearApp {
  const next = create(LinearAppSchema)
  Object.assign(next, app, patch)
  return next
}

function decodeRequest<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

async function setupLinearFixture(
  page: Page,
  options: { apps?: LinearApp[]; agents?: Agent[] } = {}
) {
  const state: {
    apps: LinearApp[]
    agents: Agent[]
    conflictOnNextUpdate: boolean
    externalDisplayName: string
  } = {
    apps: options.apps ?? [],
    agents: options.agents ?? [AGENT_A, AGENT_B],
    conflictOnNextUpdate: false,
    externalDisplayName: 'Renamed elsewhere',
  }
  const created: CreateLinearAppRequest[] = []
  const updates: UpdateLinearAppRequest[] = []

  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: state.agents,
      })
    }

    if (url.includes('AgentService/GetAgent')) {
      return fulfillProto(route, GetAgentResponseSchema, { agent: AGENT_A })
    }

    if (url.includes('LinearAppService/ListLinearApps')) {
      return fulfillProto(route, ListLinearAppsResponseSchema, {
        apps: state.apps,
      })
    }

    if (url.includes('LinearAppService/GetLinearApp')) {
      // The request carries only the App ID, so the fixture inspects the body.
      const body = route.request().postDataBuffer()
      const app = state.apps.find((candidate) =>
        body ? body.includes(candidate.id) : true
      )
      if (!app) return fulfillConnectError(route, 'not_found', 'not found')
      return fulfillProto(route, GetLinearAppResponseSchema, { app })
    }

    if (url.includes('LinearAppService/CreateLinearApp')) {
      const request = decodeRequest(CreateLinearAppRequestSchema, route)
      created.push(request)
      const app = linearApp({
        id: `app-${created.length}`,
        displayName: request.app?.displayName ?? '',
        clientId: request.app?.clientId ?? '',
        agentId: request.app?.agentId ?? '',
        revision: 1n,
        inboundEnabled: request.app?.inboundEnabled ?? false,
      })
      state.apps = [...state.apps, app]
      return fulfillProto(route, CreateLinearAppResponseSchema, { app })
    }

    if (url.includes('LinearAppService/UpdateLinearApp')) {
      const request = decodeRequest(UpdateLinearAppRequestSchema, route)
      updates.push(request)
      const stored = state.apps.find(
        (candidate) => candidate.id === request.app?.id
      )
      if (!stored) return fulfillConnectError(route, 'not_found', 'not found')

      if (state.conflictOnNextUpdate) {
        state.conflictOnNextUpdate = false
        // Another writer committed a newer revision in the meantime.
        const external = cloneApp(stored, {
          displayName: state.externalDisplayName,
          revision: stored.revision + 1n,
        })
        state.apps = state.apps.map((candidate) =>
          candidate.id === external.id ? external : candidate
        )
        return fulfillConnectError(route, 'aborted', 'revision conflict')
      }

      const updated = cloneApp(stored, {
        displayName: request.app?.displayName ?? stored.displayName,
        agentId: request.app?.agentId ?? stored.agentId,
        inboundEnabled: request.app?.inboundEnabled ?? stored.inboundEnabled,
        allowedUserIds: request.app?.allowedUserIds ?? stored.allowedUserIds,
        maxRunSeconds: request.app?.maxRunSeconds,
        revision: stored.revision + 1n,
      })
      state.apps = state.apps.map((candidate) =>
        candidate.id === updated.id ? updated : candidate
      )
      return fulfillProto(route, UpdateLinearAppResponseSchema, {
        app: updated,
      })
    }

    if (url.includes('LinearAppService/ListLinearInstallations')) {
      return fulfillProto(route, ListLinearInstallationsResponseSchema, {
        installations: [],
      })
    }

    return false
  })

  return {
    state,
    createdRequests: created,
    updateRequests: updates,
  }
}

async function openCreateStep(page: Page) {
  await page.getByRole('button', { name: 'Continue' }).click()
  await expect(
    page.getByText('No Linear App is associated with this Agent yet.')
  ).toBeVisible()
  await page.getByRole('button', { name: 'Register another App' }).click()
  await expect(page.getByText('Register a Linear App')).toBeVisible()
}

test('the Agent entry links to the Agent-scoped setup and locks the submitted agent_id', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page)

  await page.goto('/agents/agent-a/edit')
  await page.getByRole('link', { name: 'Connect Linear' }).click()
  await expect(page).toHaveURL(/\/agents\/agent-a\/linear/)
  await expect(page.getByText('Connect Linear to Agent A')).toBeVisible()

  await openCreateStep(page)

  // The Agent is preselected from the entry and cannot be changed here.
  const agentSelect = page.getByRole('combobox', { name: 'Agent' })
  await expect(agentSelect).toBeDisabled()
  await expect(agentSelect).toHaveText('Agent A')
  await expect(
    page.getByText('Fixed by the Agent entry you came from.')
  ).toBeVisible()

  await page.getByLabel('Display name').fill('Support Agent A')
  await page.getByLabel('Client ID').fill('lin-client-new')
  await page.getByLabel('Client secret').fill('client-secret-value')
  await page.getByLabel('Webhook signing secret').fill('webhook-secret-value')
  await page.getByRole('button', { name: 'Register App' }).click()

  await expect.poll(() => fixture.createdRequests.length).toBe(1)
  expect(fixture.createdRequests[0].app?.agentId).toBe('agent-a')
  expect(fixture.createdRequests[0].app?.clientId).toBe('lin-client-new')
  await expect(page.getByText('Continue Support Agent A')).toBeVisible()
})

test('a crafted app param for another Agent cannot rebind the new App', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page, {
    apps: [linearApp({ id: 'app-b', agentId: 'agent-b' })],
  })

  // The URL tries to resume an App routed to Agent B on Agent A's entry.
  await page.goto('/agents/agent-a/linear?step=3&app=app-b')
  await expect(page.getByText('Register a Linear App')).toBeVisible()
  await expect(page.getByText('Continue Support A')).toHaveCount(0)

  const agentSelect = page.getByRole('combobox', { name: 'Agent' })
  await expect(agentSelect).toBeDisabled()
  await expect(agentSelect).toHaveText('Agent A')

  await page.getByLabel('Display name').fill('Support Agent A')
  await page.getByLabel('Client ID').fill('lin-client-new')
  await page.getByRole('button', { name: 'Register App' }).click()

  await expect.poll(() => fixture.createdRequests.length).toBe(1)
  expect(fixture.createdRequests[0].app?.agentId).toBe('agent-a')
})

test('an existing App is resumed from the URL and survives reload', async ({
  page,
}) => {
  await setupLinearFixture(page, { apps: [linearApp()] })

  await page.goto('/agents/agent-a/linear?step=2')
  await expect(page.getByText('Apps already routed to this Agent')).toBeVisible()
  await expect(page.getByText('Support A')).toBeVisible()

  await page.getByRole('button', { name: 'Continue' }).click()
  await expect(page).toHaveURL(/step=3/)
  await expect(page).toHaveURL(/app=app-a/)
  await expect(page.getByText('Continue Support A')).toBeVisible()

  // A refresh keeps the same step and App from the URL.
  await page.reload()
  await expect(page.getByText('Continue Support A')).toBeVisible()

  // A later step also restores directly from the URL.
  await page.goto('/agents/agent-a/linear?step=4&app=app-a')
  await expect(page.getByText('Enable and verify', { exact: true })).toBeVisible()
  await expect(
    page.getByText(
      'An empty allowlist means every usable user in the installed organization can invoke the Agent.'
    )
  ).toBeVisible()
})

test('the shared create page still lets an operator choose the Agent', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page)

  await page.goto('/linear-apps/create')
  const agentSelect = page.getByRole('combobox', { name: 'Agent' })
  await expect(agentSelect).toBeEnabled()
  await agentSelect.click()
  await page.getByRole('option', { name: 'Agent B' }).click()
  await expect(agentSelect).toHaveText('Agent B')

  await page.getByLabel('Display name').fill('Support B')
  await page.getByLabel('Client ID').fill('lin-client-b')
  await page.getByRole('button', { name: 'Register' }).click()

  await expect.poll(() => fixture.createdRequests.length).toBe(1)
  expect(fixture.createdRequests[0].app?.agentId).toBe('agent-b')
  await expect(page).toHaveURL(/\/linear-apps\/app-1\/edit/)
})

test('OAuth success, cancellation, failure, and expiry each show the right result', async ({
  page,
}) => {
  await setupLinearFixture(page, { apps: [linearApp()] })

  await page.goto('/linear-apps/app-a/edit?linear_install=success')
  await expect(page.getByText('Installed', { exact: true })).toBeVisible()
  await expect(
    page.getByText('The Linear App is installed. Delegate an issue to it')
  ).toBeVisible()

  // Cancelling at Linear returns an error outcome, never a success banner.
  await page.goto('/linear-apps/app-a/edit?linear_install=error&reason=denied')
  await expect(page.getByText('Install failed')).toBeVisible()
  await expect(
    page.getByText('The install was cancelled or denied in Linear.')
  ).toBeVisible()
  await expect(page.getByText('Installed', { exact: true })).toHaveCount(0)

  await page.goto(
    '/linear-apps/app-a/edit?linear_install=error&reason=exchange_failed'
  )
  await expect(
    page.getByText(
      'Linear did not accept the authorization. Check the client ID and secret, then try again.'
    )
  ).toBeVisible()

  // An expired or replayed state must not look installed either.
  await page.goto(
    '/linear-apps/app-a/edit?linear_install=error&reason=state_invalid'
  )
  await expect(
    page.getByText(
      'The install link expired or was already used. Start the install again.'
    )
  ).toBeVisible()
  await expect(page.getByText('Installed', { exact: true })).toHaveCount(0)
})

test('a stale revision reloads the newer App and the next save succeeds', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page, { apps: [linearApp()] })

  await page.goto('/linear-apps/app-a/edit')
  const displayName = page.getByLabel('Display name')
  await expect(displayName).toHaveValue('Support A')

  await displayName.fill('Renamed by me')
  fixture.state.conflictOnNextUpdate = true
  await page.getByRole('button', { name: 'Save', exact: true }).click()

  await expect(page.getByText('The App changed elsewhere')).toBeVisible()
  await expect(displayName).toHaveValue('Renamed elsewhere')
  await expect.poll(() => fixture.updateRequests.length).toBe(1)
  expect(fixture.updateRequests[0].app?.revision).toBe(2n)
  expect(fixture.updateRequests[0].app?.agentId).toBe('agent-a')

  // The reloaded revision is now in the form, so saving again recovers.
  await displayName.fill('Recovered name')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect.poll(() => fixture.updateRequests.length).toBe(2)
  expect(fixture.updateRequests[1].app?.revision).toBe(3n)
  expect(fixture.updateRequests[1].app?.displayName).toBe('Recovered name')
  expect(fixture.updateRequests[1].app?.agentId).toBe('agent-a')
  await expect(page.getByText('The App changed elsewhere')).toHaveCount(0)
})
