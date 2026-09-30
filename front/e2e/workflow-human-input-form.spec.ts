import { expect, test, type Page, type Route } from '@playwright/test'
import { create, fromBinary } from '@bufbuild/protobuf'
import {
  GetAgentResponseSchema,
  UpdateAgentRequestSchema,
  UpdateAgentResponseSchema,
  type UpdateAgentRequest,
} from '../src/gen/agents/v1/agent_service_pb'
import {
  AgentSchema,
  AgentType,
  HumanInputFormFieldType,
  WorkflowNodeKind,
  type Agent,
} from '../src/gen/agents/v1/agent_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

// Configuring a Human Input node's form in the agent editor: the questions
// and forms of HUMAN_INPUT nodes are edited in place, everything else in the
// graph is kept as is, and the saved config reads back into the editor.

const SEED: Agent = create(AgentSchema, {
  name: 'approval',
  agentId: 'approval',
  type: AgentType.WORKFLOW,
  childAgentIds: ['draft', 'publish'],
  config: {
    workflow: {
      nodes: [
        { name: 'draft', kind: WorkflowNodeKind.AGENT, agentId: 'draft' },
        { name: 'ask', kind: WorkflowNodeKind.HUMAN_INPUT, question: 'Approve this deploy?' },
        { name: 'publish', kind: WorkflowNodeKind.AGENT, agentId: 'publish' },
      ],
      edges: [
        { from: 'START', to: 'draft' },
        { from: 'draft', to: 'ask' },
        { from: 'ask', to: 'publish' },
      ],
    },
  },
})

async function setupEditor(page: Page) {
  const state: {
    agent: Agent
    submitted: UpdateAgentRequest[]
    rejectNext?: string
  } = { agent: SEED, submitted: [] }
  await setupAuthenticatedConnectRoutes(page, async (route: Route, url) => {
    if (url.includes('AgentService/GetAgent')) {
      return fulfillProto(route, GetAgentResponseSchema, { agent: state.agent })
    }
    if (url.includes('AgentService/UpdateAgent')) {
      const req = fromBinary(
        UpdateAgentRequestSchema,
        route.request().postDataBuffer() ?? Buffer.alloc(0)
      )
      state.submitted.push(req)
      if (state.rejectNext) {
        const message = state.rejectNext
        state.rejectNext = undefined
        return fulfillConnectError(route, 'invalid_argument', message)
      }
      if (req.agent) state.agent = req.agent
      return fulfillProto(route, UpdateAgentResponseSchema, { agent: state.agent })
    }
    return false
  })
  return state
}

test('configures, validates, saves and reads back a Human Input form', async ({ page }) => {
  const state = await setupEditor(page)
  await page.goto('/agents/approval/edit')

  const node = page.getByRole('region', { name: 'ask' })
  await expect(node.getByLabel('Question')).toHaveValue('Approve this deploy?')
  await node.getByLabel('Show as a form').click()
  await node.getByLabel('Form title').fill('Deploy approval')

  // Field 1: a required single choice with two options.
  await node.getByLabel('Field 1 name').fill('env')
  await node.getByLabel('Field 1 label').fill('Environment')
  await node.getByLabel('Field 1 type').click()
  await page.getByRole('option', { name: 'Single choice' }).click()
  await node.getByLabel('Field 1 required').click()
  await node.getByLabel('Field 1 option 1 value').fill('prod')
  await node.getByLabel('Field 1 option 1 label').fill('Production')
  await node.getByRole('button', { name: 'Add option' }).click()
  await node.getByLabel('Field 1 option 2 value').fill('prod')
  await node.getByLabel('Field 1 option 2 label').fill('Staging')

  // Field 2: bounded text, then moved above field 1.
  await node.getByRole('button', { name: 'Add field' }).click()
  await node.getByLabel('Field 2 name').fill('reason')
  await node.getByLabel('Field 2 label').fill('Reason')
  await node.getByLabel('Field 2 hint').fill('Why now?')
  await node.getByLabel('Field 2 max length').fill('5000')
  await node.getByRole('button', { name: 'Move field 2 up' }).click()
  await expect(node.getByLabel('Field 1 name')).toHaveValue('reason')

  // Broken rules are named at their inputs and nothing is saved.
  await page.getByRole('button', { name: 'Save', exact: true }).first().click()
  await expect(node.getByText('Value "prod" is used twice.')).toBeVisible()
  await expect(node.getByText('Use a whole number from 0 to 2000')).toBeVisible()
  expect(state.submitted).toHaveLength(0)

  await node.getByLabel('Field 2 option 2 value').fill('staging')
  await node.getByLabel('Field 1 max length').fill('200')

  // A server-side refusal is shown as is and keeps the editor's content.
  state.rejectNext = 'agent "approval": workflow node "ask": form field "reason": label is too long'
  await page.getByRole('button', { name: 'Save', exact: true }).first().click()
  await expect(page.getByText('form field "reason": label is too long').first()).toBeVisible()
  await expect(node.getByLabel('Field 1 name')).toHaveValue('reason')

  await page.getByRole('button', { name: 'Save', exact: true }).first().click()
  await expect(page).toHaveURL(/\/agents$/)

  const saved = state.submitted[state.submitted.length - 1].agent!
  const nodes = saved.config!.workflow!.nodes
  expect(nodes.map((n) => n.name)).toEqual(['draft', 'ask', 'publish'])
  expect(nodes[0].agentId).toBe('draft')
  expect(saved.config!.workflow!.edges).toHaveLength(3)
  const form = nodes[1].form!
  expect(form.title).toBe('Deploy approval')
  expect(form.fields.map((f) => f.name)).toEqual(['reason', 'env'])
  expect(form.fields[0]).toMatchObject({
    label: 'Reason',
    hint: 'Why now?',
    type: HumanInputFormFieldType.TEXT,
    maxLength: 200,
    required: false,
  })
  expect(form.fields[1]).toMatchObject({
    label: 'Environment',
    required: true,
    type: HumanInputFormFieldType.SINGLE_CHOICE,
  })
  expect(form.fields[1].options.map((o) => [o.value, o.label])).toEqual([
    ['prod', 'Production'],
    ['staging', 'Staging'],
  ])

  // The saved config reads back into the editor.
  await page.goto('/agents/approval/edit')
  const reloaded = page.getByRole('region', { name: 'ask' })
  await expect(reloaded.getByLabel('Show as a form')).toBeChecked()
  await expect(reloaded.getByLabel('Form title')).toHaveValue('Deploy approval')
  await expect(reloaded.getByLabel('Field 1 name')).toHaveValue('reason')
  await expect(reloaded.getByLabel('Field 1 max length')).toHaveValue('200')
  await expect(reloaded.getByLabel('Field 2 option 2 value')).toHaveValue('staging')

  // Turning the form off keeps the plain question.
  await reloaded.getByLabel('Show as a form').click()
  await page.getByRole('button', { name: 'Save', exact: true }).first().click()
  await expect(page).toHaveURL(/\/agents$/)
  const off = state.submitted[state.submitted.length - 1].agent!.config!.workflow!.nodes[1]
  expect(off.form).toBeUndefined()
  expect(off.question).toBe('Approve this deploy?')
})
