import { expect, test, type Page } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import {
  ListAgentsResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import { AgentLifecycleStatus, AgentType } from '../src/gen/agents/v1/agent_pb'
import * as plane from '../src/gen/agents/v1/plane_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

async function setup(
  page: Page,
  options: { secret?: boolean; verified?: boolean; failInstall?: boolean; installed?: boolean; inbound?: boolean } = {}
) {
  let app = create(plane.PlaneAppSchema, {
    id: 'app-1',
    displayName: 'Plane Support',
    clientId: 'plane-client',
    agentId: 'support',
    revision: 2n,
    providerProfile: 'cloud',
    callbackUrl: 'https://butter.example/api/plane/oauth/callback',
    webhookUrl: 'https://butter.example/api/plane/webhook/app-1',
    clientSecretSet: options.secret ?? true,
    webhookSecretSet: true,
    inboundEnabled: options.inbound ?? false,
  })
  const installs: plane.BeginPlaneInstallRequest[] = []
  const updates: plane.UpdatePlaneAppRequest[] = []
  let installed = options.installed ?? false
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('PlaneAppService/GetPlaneApp')) {
      return fulfillProto(route, plane.GetPlaneAppResponseSchema, { app })
    }
    if (url.includes('PlaneAppService/GetPlaneProviderProfile')) {
      return fulfillProto(route, plane.GetPlaneProviderProfileResponseSchema, {
        profile: {
          name: 'cloud',
          apiBaseUrl: 'https://api.plane.so',
          edition: 'cloud',
          verified: options.verified ?? true,
        },
      })
    }
    if (url.includes('PlaneAppService/ListPlaneInstallations')) {
      return fulfillProto(route, plane.ListPlaneInstallationsResponseSchema, {
        installations: installed
          ? [{
              id: 'local-inst',
              appId: 'app-1',
              workspaceSlug: 'engineering',
              planeWorkspaceId: 'plane-ws',
              appInstallationId: 'external-inst',
              credentialState: 'valid',
              status: 'installed',
              grantedScopes: ['read', 'write'],
            }]
          : [],
      })
    }
    if (url.includes('PlaneAppService/BeginPlaneInstall')) {
      installs.push(fromBinary(plane.BeginPlaneInstallRequestSchema, route.request().postDataBuffer()!))
      if (options.failInstall) {
        return fulfillConnectError(route, 'failed_precondition', 'Set the Plane App client secret before installing it')
      }
      return fulfillProto(route, plane.BeginPlaneInstallResponseSchema, {
        authorizeUrl: 'https://consent.plane.test/auth/o/authorize-app/?state=one-use',
      })
    }
    if (url.includes('PlaneAppService/UpdatePlaneApp')) {
      const request = fromBinary(plane.UpdatePlaneAppRequestSchema, route.request().postDataBuffer()!)
      updates.push(request)
      app = create(plane.PlaneAppSchema, { ...app, ...request.app, revision: app.revision + 1n })
      return fulfillProto(route, plane.UpdatePlaneAppResponseSchema, { app })
    }
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [{ agentId: 'support', name: 'Support', type: AgentType.LLM, lifecycleStatus: AgentLifecycleStatus.ACTIVE }],
      })
    }
    return false
  })
  await page.route('https://consent.plane.test/**', async (route) => {
    installed = true
    const returnUrl = installs.at(-1)!.returnUrl
    await route.fulfill({
      status: 302,
      headers: { location: `${returnUrl}?plane_install=success&installation_id=local-inst` },
      body: '',
    })
  })
  return { installs, updates }
}

test('App page starts installation and displays verified callback result', async ({ page }) => {
  const { installs } = await setup(page)
  await page.goto('/plane-apps/app-1/edit')
  await expect(page.getByText('No verified installation yet.')).toBeVisible()
  await page.getByRole('button', { name: 'Install to Plane' }).click()
  await expect(page.getByText('Installed in Plane', { exact: true })).toBeVisible()
  await expect(page.getByText('engineering', { exact: false })).toBeVisible()
  await expect(page.getByText('Granted scopes: read, write')).toBeVisible()
  expect(installs).toHaveLength(1)
  expect(installs[0].appId).toBe('app-1')
  expect(new URL(installs[0].returnUrl).pathname).toBe('/plane-apps/app-1/edit')
  await page.screenshot({ path: '/tmp/butter-plane-install-page.png', fullPage: true })
})

test('missing secret disables install; unverified profile blocks only inbound', async ({ page }) => {
  await setup(page, { secret: false, verified: false })
  await page.goto('/plane-apps/app-1/edit')
  await expect(page.getByRole('button', { name: 'Install to Plane' })).toBeDisabled()
  await expect(page.getByRole('button', { name: 'Enable inbound' })).toBeDisabled()
  await expect(page.getByText(/must verify this deployment/)).toBeVisible()
})

test('installation errors and callback failures are actionable', async ({ page }) => {
  await setup(page, { failInstall: true, verified: false })
  await page.goto('/plane-apps/app-1/edit?plane_install=error&reason=installation_missing')
  await expect(page.getByText('Plane installation failed', { exact: true })).toBeVisible()
  await expect(page.getByText(/Plane did not return an installation ID/)).toBeVisible()
  await page.getByRole('button', { name: 'Install to Plane' }).click()
  await expect(
    page.getByText('Set the Plane App client secret before installing it', {
      exact: true,
    })
  ).toBeVisible()
})

test('inbound requires explicit access confirmation and normal editing preserves it', async ({ page }) => {
  const { updates } = await setup(page, { installed: true })
  await page.goto('/plane-apps/app-1/edit')
  const enable = page.getByRole('button', { name: 'Enable inbound' })
  await expect(enable).toBeDisabled()
  await page.getByRole('checkbox', { name: /I allow all members/ }).check()
  await enable.click()
  await expect(page.getByRole('button', { name: 'Disable inbound' })).toBeVisible()
  expect(updates[0].app?.inboundEnabled).toBe(true)
  await page.getByLabel('Display name').fill('Renamed support')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page.getByText('Plane App updated', { exact: true })).toBeVisible()
  expect(updates.at(-1)?.app?.inboundEnabled).toBe(true)
})
