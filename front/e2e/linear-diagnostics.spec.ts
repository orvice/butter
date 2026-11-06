import { expect, test, type Page, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import { TimestampSchema } from '@bufbuild/protobuf/wkt'
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
  GetLinearAppResponseSchema,
  LinearAppCredentialState,
  LinearAppSchema,
  LinearInstallationCredentialState,
  LinearInstallationSchema,
  LinearProcessingStatus,
  LinearProcessingRecordSchema,
  ListLinearAppsResponseSchema,
  ListLinearInstallationsRequestSchema,
  ListLinearInstallationsResponseSchema,
  ListLinearProcessingRecordsRequestSchema,
  ListLinearProcessingRecordsResponseSchema,
  ResendLinearReplyRequestSchema,
  ResendLinearReplyResponseSchema,
  type LinearApp,
  type LinearInstallation,
  type LinearProcessingRecord,
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

type LinearAppPatch = Partial<Omit<LinearApp, '$typeName' | '$unknown'>>
type InstallationPatch = Partial<
  Omit<LinearInstallation, '$typeName' | '$unknown'>
>
type RecordPatch = Partial<
  Omit<LinearProcessingRecord, '$typeName' | '$unknown'>
>

const BASE_SECONDS = 1_790_000_000

/** stamp returns a Timestamp offsetSeconds after the fixed fixture epoch. */
function stamp(offsetSeconds: number) {
  return create(TimestampSchema, {
    seconds: BigInt(BASE_SECONDS + offsetSeconds),
    nanos: 0,
  })
}

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

function installation(overrides: InstallationPatch = {}): LinearInstallation {
  return create(LinearInstallationSchema, {
    id: 'inst-1',
    appId: 'app-a',
    organizationId: 'org-1',
    organizationName: 'Linear Org One',
    appUserId: 'app-user-1',
    scopes: ['read', 'write'],
    credentialState: LinearInstallationCredentialState.VALID,
    installedAt: stamp(0),
    updatedAt: stamp(0),
    ...overrides,
  })
}

/** successRecord builds evidence of a real Agent run that reached Linear. */
function successRecord(overrides: RecordPatch = {}): LinearProcessingRecord {
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
    createdAt: stamp(100),
    updatedAt: stamp(100),
    ...overrides,
  })
}

function failedRecord(overrides: RecordPatch = {}): LinearProcessingRecord {
  return create(LinearProcessingRecordSchema, {
    id: 'rec-failed',
    appId: 'app-a',
    installationId: 'inst-1',
    deliveryId: 'delivery-failed',
    status: LinearProcessingStatus.FAILED,
    attempts: 3,
    appRevision: 4n,
    error: 'Linear API returned 401 while posting the reply.',
    action: 'prompted',
    agentId: 'agent-a',
    agentSessionId: 'session-1',
    issueIdentifier: 'KON-1',
    createdAt: stamp(200),
    updatedAt: stamp(200),
    ...overrides,
  })
}

type FixtureState = {
  apps: LinearApp[]
  installations: LinearInstallation[]
  records: LinearProcessingRecord[]
}

function setupLinearFixture(
  page: Page,
  state: FixtureState,
  options: { readError?: 'permission_denied' } = {}
) {
  let listRecordCalls = 0

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
      const body = route.request().postDataBuffer()
      const app = state.apps.find((candidate) =>
        body ? body.includes(candidate.id) : true
      )
      if (!app) return fulfillConnectError(route, 'not_found', 'not found')
      return fulfillProto(route, GetLinearAppResponseSchema, { app })
    }

    if (url.includes('LinearAppService/ListLinearInstallations')) {
      const request = decode(ListLinearInstallationsRequestSchema, route)
      return fulfillProto(route, ListLinearInstallationsResponseSchema, {
        installations: state.installations.filter(
          (candidate) =>
            !request.appId || candidate.appId === request.appId
        ),
      })
    }

    if (url.includes('LinearProcessingService/ListLinearProcessingRecords')) {
      if (options.readError) {
        return fulfillConnectError(
          route,
          options.readError,
          'missing Linear read access'
        )
      }
      listRecordCalls += 1
      const request = decode(ListLinearProcessingRecordsRequestSchema, route)
      let records = state.records
      if (request.appId) {
        records = records.filter((record) => record.appId === request.appId)
      }
      if (request.status) {
        records = records.filter((record) => record.status === request.status)
      }
      records = [...records].sort(
        (a, b) =>
          Number((b.createdAt?.seconds ?? 0n) - (a.createdAt?.seconds ?? 0n))
      )
      const limit = request.pageSize > 0 ? request.pageSize : 200
      return fulfillProto(route, ListLinearProcessingRecordsResponseSchema, {
        records: records.slice(0, limit),
      })
    }

    if (url.includes('LinearProcessingService/ResendLinearReply')) {
      const request = decode(ResendLinearReplyRequestSchema, route)
      const record = state.records.find((item) => item.id === request.id)
      if (!record) return fulfillConnectError(route, 'not_found', 'not found')
      record.delivered = true
      record.status = LinearProcessingStatus.SUCCEEDED
      return fulfillProto(route, ResendLinearReplyResponseSchema, { record })
    }

    return false
  }).then(() => ({
    state,
    get listRecordCalls() {
      return listRecordCalls
    },
  }))
}

function decode<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

test('configuration gaps are named, and a missing installation is not a failure', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [
      linearApp({
        clientSecretSet: false,
        webhookSecretSet: false,
        callbackUrl: '',
        webhookUrl: '',
        credentialState: LinearAppCredentialState.INCOMPLETE,
        inboundEnabled: false,
      }),
    ],
    installations: [],
    records: [],
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(page.getByText('Configuration incomplete')).toBeVisible()
  await expect(
    page.getByText(
      'Missing the OAuth client secret, the webhook signing secret, the public base URL (callback URL).'
    )
  ).toBeVisible()
  await expect(page.getByText('Inbound enabled (desired state)')).toHaveCount(0)
  await expect(
    page.getByText(
      'No Linear organization is installed yet. Install the App before expecting any delivery evidence.'
    )
  ).toBeVisible()

  // The verification view is read-only: no install, verify, or enable action.
  await expect(
    page.getByTestId('linear-diagnostics').getByRole('button')
  ).toHaveCount(0)
})

test('configured but inbound disabled is worded as desired state, not health', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp({ inboundEnabled: false })],
    installations: [installation()],
    records: [successRecord()],
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(page.getByText('Configured, inbound disabled')).toBeVisible()
  await expect(
    page.getByText(
      'Linear deliveries are not accepted. Existing evidence stays visible.'
    )
  ).toBeVisible()
  // Historical evidence is still shown even while inbound is off.
  await expect(page.getByTestId('linear-recent-success')).toBeVisible()
  // Turning inbound off must not promise to stop an executing turn.
  await expect(
    page.getByText(
      'it does not interrupt a turn that is already running'
    )
  ).toBeVisible()
})

test('inbound enabled is shown as the desired state, with historical success and no health claim', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [successRecord()],
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(page.getByText('Inbound enabled (desired state)')).toBeVisible()
  await expect(
    page.getByText(
      'This is the desired state, not a health signal: check the evidence below for what actually ran.'
    )
  ).toBeVisible()

  const success = page.getByTestId('linear-recent-success')
  await expect(success).toContainText('Recent successful run')
  await expect(success).toContainText('invocation inv-success')
  await expect(success).toContainText('KON-1')
  await expect(success).toContainText(
    'Historical evidence from the read window, not current health. A later run may still fail.'
  )
  // The page states the installation fact, never a connected/health claim.
  await expect(page.getByText('Connected', { exact: true })).toHaveCount(0)
  await expect(page.getByText('Healthy', { exact: true })).toHaveCount(0)
  await expect(
    page.getByText('Refreshes every 5 seconds while this page is visible')
  ).toBeVisible()
})

test('installed but no records says no event was observed in the range', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [],
  })

  await page.goto('/linear-apps/app-a/edit')

  const block = page.getByTestId('linear-installation-diagnostic-inst-1')
  await expect(block).toContainText('No run evidence in the current record range')
  await expect(block).toContainText(
    'No events were observed for this installation in the current record range.'
  )
  await expect(block).toContainText('Delegate an issue to the App or @mention it')
})

test('multiple installations report success only where a run actually happened', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [
      installation(),
      installation({
        id: 'inst-2',
        organizationId: 'org-2',
        organizationName: 'Linear Org Two',
        appUserId: 'app-user-2',
      }),
    ],
    records: [successRecord()],
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(
    page
      .getByTestId('linear-installation-diagnostic-inst-1')
      .getByTestId('linear-recent-success')
  ).toBeVisible()
  const second = page.getByTestId('linear-installation-diagnostic-inst-2')
  await expect(second.getByTestId('linear-no-success')).toBeVisible()
  await expect(second).toContainText(
    'No events were observed for this installation in the current record range.'
  )
  await expect(second).toContainText('Linear Org Two')
})

test('a historical success followed by a recent failure shows both, and the failure wins', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [
      failedRecord(),
      successRecord({ createdAt: stamp(100), updatedAt: stamp(100) }),
    ],
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(page.getByTestId('linear-recent-success')).toBeVisible()
  await expect(
    page.getByText('The latest delivery did not succeed')
  ).toBeVisible()
  const issues = page.getByTestId('linear-recent-issues')
  await expect(issues).toContainText('Failed before the Agent ran')
  await expect(issues).toContainText(
    'Linear API returned 401 while posting the reply.'
  )
  await expect(issues).toContainText(
    'Next: Fix the cause, then trigger the event again from Linear. This delivery is safe to retry.'
  )
})

test('old Agent records, a changed configuration, removed installations, and uncertain outcomes are explained', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp({ allowedUserIds: ['user-allowed'] })],
    installations: [
      installation({
        credentialState: LinearInstallationCredentialState.NEEDS_REINSTALL,
        lastCredentialError: 'Linear revoked the refresh token.',
      }),
    ],
    records: [
      create(LinearProcessingRecordSchema, {
        id: 'rec-uncertain',
        appId: 'app-a',
        installationId: 'inst-1',
        status: LinearProcessingStatus.FAILED_UNCERTAIN,
        error: 'Pod restarted while the Agent was running.',
        action: 'prompted',
        agentId: 'agent-a',
        issueIdentifier: 'KON-9',
        createdAt: stamp(500),
        updatedAt: stamp(500),
      }),
      create(LinearProcessingRecordSchema, {
        id: 'rec-cancelled',
        appId: 'app-a',
        installationId: 'inst-1',
        status: LinearProcessingStatus.CANCELLED,
        error: 'stopped by the user',
        action: 'prompted',
        agentId: 'agent-a',
        issueIdentifier: 'KON-8',
        createdAt: stamp(400),
        updatedAt: stamp(400),
      }),
      create(LinearProcessingRecordSchema, {
        id: 'rec-rejected',
        appId: 'app-a',
        installationId: 'inst-1',
        status: LinearProcessingStatus.SUCCEEDED,
        outputType: 'error',
        output: 'This Linear user is not allowed to invoke the Agent.',
        delivered: true,
        action: 'created',
        agentId: 'agent-a',
        promptingUserId: 'user-not-allowed',
        issueIdentifier: 'KON-7',
        createdAt: stamp(300),
        updatedAt: stamp(300),
      }),
      successRecord({
        id: 'rec-old-agent',
        invocationId: 'inv-old',
        agentId: 'agent-old',
        issueIdentifier: 'KON-6',
        createdAt: stamp(200),
        updatedAt: stamp(200),
      }),
      successRecord({
        id: 'rec-old-revision',
        invocationId: 'inv-old-revision',
        appRevision: 3n,
        issueIdentifier: 'KON-5',
        createdAt: stamp(150),
        updatedAt: stamp(150),
      }),
      successRecord({
        id: 'rec-removed-installation',
        installationId: 'inst-removed',
        invocationId: 'inv-removed',
        issueIdentifier: 'KON-4',
        createdAt: stamp(120),
        updatedAt: stamp(120),
      }),
    ],
  })

  await page.goto('/linear-apps/app-a/edit')

  const block = page.getByTestId('linear-installation-diagnostic-inst-1')
  await expect(block).toContainText('Reinstall needed')
  await expect(block).toContainText('Linear revoked the refresh token.')
  await expect(block).toContainText('Stopped mid-run, result uncertain')
  await expect(block).toContainText(
    'This is never retried automatically. Check the likely side effects'
  )
  await expect(block).toContainText('Stopped')
  await expect(block).toContainText(
    'Observed and answered without running the Agent'
  )
  await expect(block).toContainText(
    "is not in this App's allowlist"
  )
  await expect(block).toContainText(
    'from an earlier Agent binding (or without an Agent stamp) are not counted as evidence for this Agent'
  )
  await expect(block).toContainText(
    'This run was accepted under App revision 3; the current configuration is revision 4.'
  )
  // The record that did run under the current Agent and revision is evidence.
  await expect(block).toContainText('inv-old-revision')
  await expect(
    page.getByText(
      'belong to an installation that is no longer listed'
    )
  ).toBeVisible()
})

test('a full read window is called out instead of claiming there was never a run', async ({
  page,
}) => {
  const records = Array.from({ length: 200 }, (_, index) =>
    failedRecord({
      id: `rec-${index}`,
      issueIdentifier: `KON-${index}`,
      error: 'Webhook signature rejected.',
      createdAt: stamp(1000 + index),
      updatedAt: stamp(1000 + index),
    })
  )
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records,
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(page.getByText('The read window is full (200 records)')).toBeVisible()
  await expect(
    page.getByText(
      'A lack of success here may only mean the evidence is out of range.'
    )
  ).toBeVisible()
  await expect(page.getByTestId('linear-read-window')).toContainText(
    'Read window: 200 records'
  )
  await expect(page.getByTestId('linear-no-success')).toBeVisible()
})

test('the verification view refreshes every 5 seconds only while it is visible', async ({
  page,
}) => {
  const fixture = await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [],
  })

  await page.goto('/linear-apps/app-a/edit')
  await expect(page.getByTestId('linear-no-success')).toBeVisible()

  // A new delivery arrives while the page stays open: the next poll shows it.
  fixture.state.records = [successRecord()]
  await expect(page.getByTestId('linear-recent-success')).toBeVisible({
    timeout: 12_000,
  })

  // Leaving the page stops the refresh: no further list requests arrive.
  await page.goto('/linear-apps')
  await expect(page).toHaveURL(/\/linear-apps$/)
  const callsAfterLeaving = fixture.listRecordCalls
  fixture.state.records = [
    successRecord({ id: 'rec-later', invocationId: 'inv-later' }),
    ...fixture.state.records,
  ]
  await page.waitForTimeout(6_000)
  expect(fixture.listRecordCalls).toBe(callsAfterLeaving)
})

test('Deliveries explains each record and still resends a persisted reply without rerunning', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [
      successRecord({ id: 'rec-success', issueIdentifier: 'KON-1' }),
      successRecord({
        id: 'rec-error-run',
        invocationId: 'inv-error',
        outputType: 'error',
        output: 'Run failed: model quota exhausted.',
        issueIdentifier: 'KON-3',
        createdAt: stamp(450),
      }),
      failedRecord({
        id: 'rec-failed-reply',
        issueIdentifier: 'KON-2',
        output: 'Persisted reply',
        delivered: false,
        error: 'Posting the reply failed.',
        createdAt: stamp(400),
      }),
    ],
  })

  await page.goto('/linear-deliveries')
  await page.getByRole('combobox', { name: 'Linear App filter' }).click()
  await page.getByRole('option', { name: 'Support A' }).click()

  await expect(page.getByTestId('linear-diagnostics')).toBeVisible()
  await expect(page.getByTestId('linear-recent-success')).toBeVisible()

  await expect(
    page.getByTestId('linear-record-diagnosis-rec-failed-reply')
  ).toContainText('Delivery failed after the reply was produced')
  await expect(
    page.getByTestId('linear-record-diagnosis-rec-failed-reply')
  ).toContainText(
    'Next: Resend the persisted reply from Deliveries. Resending never reruns the Agent.'
  )
  await expect(
    page.getByTestId('linear-record-diagnosis-rec-success')
  ).toContainText('Ran and replied')
  await expect(
    page.getByTestId('linear-record-diagnosis-rec-error-run')
  ).toContainText('Ran, and replied with an error')
  await expect(
    page.getByTestId('linear-record-diagnosis-rec-error-run')
  ).toContainText(
    'Next: Read the error reply in Linear, fix the cause, then prompt again.'
  )

  const resend = page.getByRole('button', {
    name: 'Resend reply for KON-2',
  })
  await expect(resend).toBeEnabled()
  await resend.click()
  await expect(page.getByText('Reply resent')).toBeVisible()

  // The other Deliveries entry point still renders its own list.
  await page.goto('/linear-apps')
  await expect(page.getByRole('table').getByText('Support A')).toBeVisible()
})

test('the guided setup step 4 shows the same diagnostics for the bound App', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation()],
    records: [successRecord()],
  })

  await page.goto('/agents/agent-a/linear?step=4&app=app-a')

  await expect(page.getByText('Enable and verify', { exact: true })).toBeVisible()
  await expect(page.getByTestId('linear-diagnostics')).toBeVisible()
  await expect(page.getByTestId('linear-recent-success')).toContainText(
    'invocation inv-success'
  )
})

test('a read-scope failure is explained and never reads as "no events"', async ({
  page,
}) => {
  await setupLinearFixture(
    page,
    {
      apps: [linearApp()],
      installations: [installation()],
      records: [],
    },
    { readError: 'permission_denied' }
  )

  await page.goto('/linear-apps/app-a/edit')

  await expect(
    page.getByText('Delivery records are not visible to this account')
  ).toBeVisible()
  await expect(
    page.getByText(
      'This is a read-scope limit, not evidence that nothing ran.'
    )
  ).toBeVisible()
  // The evidence section is not rendered at all, so no false "no events" claim.
  await expect(page.getByTestId('linear-read-window')).toHaveCount(0)
  await expect(page.getByTestId('linear-no-success')).toHaveCount(0)
  await expect(
    page.getByText('No events were observed for this installation')
  ).toHaveCount(0)
  // The credential fact is still shown, since installations were readable.
  await expect(page.getByText('Credential: last known usable')).toBeVisible()

  await page.goto('/linear-deliveries')
  await expect(
    page.getByText(/Delivery records are not visible to this account/)
  ).toBeVisible()
  await expect(
    page.getByText('No Linear deliveries have been processed yet.')
  ).toHaveCount(0)
})

test('limited Linear scopes are stated as a fact and do not hide the read window', async ({
  page,
}) => {
  await setupLinearFixture(page, {
    apps: [linearApp()],
    installations: [installation({ scopes: ['write'] })],
    records: [successRecord()],
  })

  await page.goto('/linear-apps/app-a/edit')

  await expect(page.getByText('Granted Linear scopes: write')).toBeVisible()
  await expect(
    page.getByText(
      "This installation does not include Linear's read scope. The evidence here only reflects what Butter recorded; some Linear data cannot be read."
    )
  ).toBeVisible()
  await expect(page.getByTestId('linear-recent-success')).toBeVisible()
})
