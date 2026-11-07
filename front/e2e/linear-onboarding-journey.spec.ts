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
  BeginLinearInstallRequestSchema,
  BeginLinearInstallResponseSchema,
  GetLinearAppRequestSchema,
  GetLinearAppResponseSchema,
  LinearAppCredentialState,
  LinearAppSchema,
  LinearInstallationCredentialState,
  LinearInstallationSchema,
  LinearProcessingStatus,
  LinearProcessingRecordSchema,
  ListLinearAppsResponseSchema,
  ListLinearInstallationsResponseSchema,
  ListLinearProcessingRecordsResponseSchema,
  UpdateLinearAppRequestSchema,
  UpdateLinearAppResponseSchema,
  type LinearApp,
  type LinearInstallation,
  type LinearProcessingRecord,
} from '../src/gen/agents/v1/linear_pb'
import {
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const AGENT_A: Agent = create(AgentSchema, {
  agentId: 'agent-a',
  name: 'Agent A',
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
    revision: 4n,
    clientSecretSet: true,
    webhookSecretSet: true,
    callbackUrl: 'https://butter.example.com/api/linear/oauth/callback',
    webhookUrl: 'https://butter.example.com/api/linear/webhook/app-a',
    credentialState: LinearAppCredentialState.COMPLETE,
    inboundEnabled: true,
    ...overrides,
  })
}

function installation(
  overrides: Partial<Omit<LinearInstallation, '$typeName' | '$unknown'>> = {}
): LinearInstallation {
  return create(LinearInstallationSchema, {
    id: 'inst-1',
    appId: 'app-a',
    organizationId: 'org-1',
    organizationName: 'Linear Org One',
    appUserId: 'app-user-1',
    scopes: ['read', 'write'],
    credentialState: LinearInstallationCredentialState.VALID,
    ...overrides,
  })
}

function successRecord(
  overrides: Partial<Omit<LinearProcessingRecord, '$typeName' | '$unknown'>> = {}
): LinearProcessingRecord {
  return create(LinearProcessingRecordSchema, {
    id: 'rec-success',
    appId: 'app-a',
    installationId: 'inst-1',
    deliveryId: 'delivery-success',
    status: LinearProcessingStatus.SUCCEEDED,
    invocationId: 'inv-success',
    attempts: 1,
    appRevision: 4n,
    output: 'Done: created the issue.',
    outputType: 'response',
    delivered: true,
    action: 'created',
    agentId: 'agent-a',
    agentSessionId: 'session-1',
    issueIdentifier: 'KON-1',
    ...overrides,
  })
}

type FixtureState = {
  apps: LinearApp[]
  installations: LinearInstallation[]
  records: LinearProcessingRecord[]
}

function cloneApp(app: LinearApp, patch: LinearAppPatch): LinearApp {
  const next = create(LinearAppSchema)
  Object.assign(next, app, patch)
  return next
}

function setupLinearFixture(page: Page, state: FixtureState) {
  const updates: Array<{ app?: LinearApp }> = []
  const installs: Array<{ appId: string; returnUrl: string }> = []

  return setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, { agents: [AGENT_A] })
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
      const request = fromBinary(
        GetLinearAppRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      const app = state.apps.find((candidate) => candidate.id === request.id)
      if (!app) return fulfillProto(route, GetLinearAppResponseSchema, {})
      return fulfillProto(route, GetLinearAppResponseSchema, { app })
    }

    if (url.includes('LinearAppService/UpdateLinearApp')) {
      const request = fromBinary(
        UpdateLinearAppRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      updates.push(request)
      const stored = state.apps.find(
        (candidate) => candidate.id === request.app?.id
      )
      if (!stored) throw new Error('fixture has no App to update')
      const updated = cloneApp(stored, {
        displayName: request.app?.displayName ?? stored.displayName,
        inboundEnabled: request.app?.inboundEnabled ?? false,
        allowedUserIds: request.app?.allowedUserIds ?? [],
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
        installations: state.installations,
      })
    }

    if (url.includes('LinearProcessingService/ListLinearProcessingRecords')) {
      return fulfillProto(route, ListLinearProcessingRecordsResponseSchema, {
        records: state.records,
      })
    }

    if (url.includes('LinearAppService/BeginLinearInstall')) {
      const request = fromBinary(
        BeginLinearInstallRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      installs.push(request)
      return fulfillProto(route, BeginLinearInstallResponseSchema, {
        authorizeUrl: 'https://linear.test/oauth/authorize?actor=app',
      })
    }

    return false
  }).then(() => ({ state, updates, installs }))
}

test('the Connections entry, the Agent entry and the App page are one resource', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [successRecord()],
  })

  // Connections (Manage → Linear Apps) → the App page.
  await page.goto('/linear-apps')
  await expect(page.getByRole('table').getByText('Support A')).toBeVisible()
  await page.getByRole('button', { name: 'Actions for Support A' }).click()
  await page.getByRole('menuitem', { name: 'Edit' }).click()
  await expect(page).toHaveURL(/\/linear-apps\/app-a\/edit/)
  await expect(page.getByTestId('linear-recent-success')).toBeVisible()

  // The App page's guided setup opens the same App on its verification step.
  await page.getByRole('link', { name: 'Guided setup' }).click()
  await expect(page).toHaveURL(/\/agents\/agent-a\/linear\?step=4&app=app-a/)
  await expect(page.getByText('Enable and verify', { exact: true })).toBeVisible()
  await expect(page.getByTestId('linear-recent-success')).toBeVisible()
  await page.reload()
  await expect(page).toHaveURL(/step=4&app=app-a/)
  await expect(page.getByTestId('linear-recent-success')).toBeVisible()

  // The Agent entry lists the same App and hands back to the same App page.
  await page.goto('/agents/agent-a/edit')
  await page.getByRole('link', { name: 'Connect Linear' }).click()
  await expect(page).toHaveURL(/\/agents\/agent-a\/linear/)
  await expect(page.getByText('Before you start')).toBeVisible()
  await page.getByRole('button', { name: 'Continue', exact: true }).click()
  await expect(page.getByText('Apps already routed to this Agent')).toBeVisible()
  await page.getByRole('button', { name: 'Continue', exact: true }).click()
  await expect(page.getByText('Continue Support A')).toBeVisible()
  await expect(
    page.getByRole('link', { name: /Open App configuration/ })
  ).toHaveAttribute('href', '/linear-apps/app-a/edit')
})

test('enabling inbound is an explicit save that keeps the revision and stays write-only', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page, {
    apps: [linearApp({ inboundEnabled: false })],
    installations: [installation()],
    records: [],
  })

  await page.goto('/linear-apps/app-a/edit')
  await expect(page.getByText('Configured, inbound disabled')).toBeVisible()
  await expect(
    page.getByText('No run evidence in the current record range')
  ).toBeVisible()

  // A secret typed into the rotation field is never persisted or echoed.
  const clientSecretRotation = page.getByLabel(/^Client secret/)
  await clientSecretRotation.fill('super-secret-value')
  await expect(clientSecretRotation).toHaveAttribute('type', 'password')

  await page.getByLabel('Allowed Linear user IDs').fill('User-1\nuser-2')
  await page.getByRole('switch', { name: 'Inbound enabled' }).click()
  await page.getByRole('button', { name: 'Save', exact: true }).click()

  await expect.poll(() => fixture.updates.length).toBe(1)
  const update = fixture.updates[0]
  expect(update.app?.id).toBe('app-a')
  expect(update.app?.revision).toBe(4n)
  expect(update.app?.agentId).toBe('agent-a')
  expect(update.app?.inboundEnabled).toBe(true)
  expect(update.app?.allowedUserIds).toEqual(['User-1', 'user-2'])

  // The saved App is now shown as enabled — a desired state, not health.
  await expect(page.getByText('Inbound enabled (desired state)')).toBeVisible()
  await expect(
    page.getByText('No run evidence in the current record range')
  ).toBeVisible()

  // A reload keeps the fact and never refills or exposes a secret.
  await page.reload()
  await expect(page.getByText('Inbound enabled (desired state)')).toBeVisible()
  await expect(page.getByRole('switch', { name: 'Inbound enabled' })).toBeChecked()
  await expect(page.getByLabel(/^Client secret/)).toHaveValue('')
  await expect(page.getByLabel(/^Webhook signing secret/)).toHaveValue('')
  expect(new URL(page.url()).search).toBe('')

  const stored = await page.evaluate(() => {
    const values = Object.keys(localStorage).map(
      (key) => localStorage.getItem(key) ?? ''
    )
    return values.join('|')
  })
  expect(stored).not.toContain('super-secret-value')
})

test('Install to Linear starts the OAuth flow for this App and returns to this page', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [],
    records: [],
  })

  await page.route('https://linear.test/**', (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'text/html',
      body: '<html><body><h1>Linear authorize</h1></body></html>',
    })
  )

  await page.goto('/linear-apps/app-a/edit')
  await expect(
    page.getByText('Not installed in any Linear organization yet.')
  ).toBeVisible()
  await page.getByRole('button', { name: 'Install to Linear' }).click()

  await expect.poll(() => fixture.installs.length).toBe(1)
  const request = fixture.installs[0]
  expect(request.appId).toBe('app-a')
  expect(request.returnUrl).toMatch(/\/linear-apps\/app-a\/edit$/)

  await expect(page).toHaveURL(/linear\.test\/oauth\/authorize\?actor=app/)
  await expect(
    page.getByRole('heading', { name: 'Linear authorize' })
  ).toBeVisible()
})
