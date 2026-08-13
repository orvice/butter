import { expect, test, type Page } from '@playwright/test'
import { fromBinary } from '@bufbuild/protobuf'
import {
  GetTelegramDestinationResponseSchema,
  TelegramSessionPolicy,
  UpdateTelegramDestinationRequestSchema,
  UpdateTelegramDestinationResponseSchema,
} from '../src/gen/agents/v1/telegram_pb'
import {
  ListAgentsResponseSchema,
  ListModelProvidersResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

const DESTINATION = {
  id: 'dest-1',
  key: 'incidents',
  name: 'Incidents',
  channelId: 'ch-1',
  chatId: '-1001234567890',
  messageThreadId: '42',
  inboundEnabled: true,
  outboundEnabled: true,
  revision: 2n,
  config: {
    agentId: 'support',
    model: 'fast',
    selectableAgentIds: ['support', 'research'],
    selectableModels: ['fast', 'pro'],
    controllerUserIds: ['7'],
    sessionPolicy: TelegramSessionPolicy.USER,
  },
}

async function setupSelectionRoutes(
  page: Page,
  onUpdate?: (route: never) => Promise<boolean>
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramDestinationService/UpdateTelegramDestination')) {
      if (onUpdate) return onUpdate(route as never)
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(UpdateTelegramDestinationRequestSchema, body)
      return fulfillProto(route, UpdateTelegramDestinationResponseSchema, {
        destination: req.destination,
      })
    }
    if (url.includes('TelegramDestinationService/GetTelegramDestination')) {
      return fulfillProto(route, GetTelegramDestinationResponseSchema, {
        destination: DESTINATION,
      })
    }
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [
          { name: 'Support', agentId: 'support' },
          { name: 'Research', agentId: 'research' },
        ],
      })
    }
    if (url.includes('ModelProviderService/ListModelProviders')) {
      return fulfillProto(route, ListModelProvidersResponseSchema, {
        modelProviders: [
          { name: 'openai', models: [{ name: 'gpt-5', alias: 'fast' }, { name: 'gpt-5-pro', alias: 'pro' }] },
        ],
      })
    }
    return false
  })
}

test('the destination form shows the configured candidate lists', async ({ page }) => {
  await setupSelectionRoutes(page)

  await page.goto('/telegram-destinations/dest-1')

  await expect(page.getByLabel('Selectable agents')).toHaveValue('support, research')
  await expect(page.getByLabel('Selectable models')).toHaveValue('fast, pro')
  await expect(page.getByRole('combobox', { name: 'Conversation history' })).toContainText(
    'Separate per user'
  )
})

test('candidate lists round-trip through a save', async ({ page }) => {
  let sentAgents: string[] = []
  await setupSelectionRoutes(page, async (route) => {
    const body = (route as never as { request(): { postDataBuffer(): Buffer | null } })
      .request()
      .postDataBuffer() ?? Buffer.alloc(0)
    const req = fromBinary(UpdateTelegramDestinationRequestSchema, body)
    sentAgents = req.destination?.config?.selectableAgentIds ?? []
    return fulfillProto(route, UpdateTelegramDestinationResponseSchema, {
      destination: req.destination,
    })
  })

  await page.goto('/telegram-destinations/dest-1')
  await page.getByLabel('Selectable agents').fill('support, research, triage')
  await page.getByRole('button', { name: 'Save' }).click()

  await expect.poll(() => sentAgents).toEqual(['support', 'research', 'triage'])
})
