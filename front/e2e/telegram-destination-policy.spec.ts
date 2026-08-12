import { expect, test, type Page } from '@playwright/test'
import {
  GetTelegramDestinationResponseSchema,
  TelegramReplyMode,
  TelegramSessionPolicy,
  TelegramTriggerMode,
  UpdateTelegramDestinationResponseSchema,
} from '../src/gen/agents/v1/telegram_pb'
import { ListAgentsResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const DESTINATION = {
  id: 'dest-1',
  key: 'incidents',
  name: 'Incidents',
  channelId: 'ch-1',
  chatId: '-1001234567890',
  messageThreadId: '42',
  inboundEnabled: true,
  outboundEnabled: true,
  revision: 3n,
  config: {
    agentId: 'support',
    triggerMode: TelegramTriggerMode.MENTION,
    sessionPolicy: TelegramSessionPolicy.USER,
    replyMode: TelegramReplyMode.NEW_MESSAGE,
    allowedUserIds: ['10', '20'],
    controllerUserIds: ['10'],
    debugDefault: true,
  },
}

async function setupPolicyRoutes(page: Page, onUpdate?: (route: never) => Promise<boolean>) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramDestinationService/UpdateTelegramDestination')) {
      if (onUpdate) return onUpdate(route as never)
      return fulfillProto(route, UpdateTelegramDestinationResponseSchema, {
        destination: DESTINATION,
      })
    }
    if (url.includes('TelegramDestinationService/GetTelegramDestination')) {
      return fulfillProto(route, GetTelegramDestinationResponseSchema, {
        destination: DESTINATION,
      })
    }
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [{ name: 'Support', agentId: 'support' }],
      })
    }
    return false
  })
}

test('the destination form shows the stored interaction policy', async ({ page }) => {
  await setupPolicyRoutes(page)

  await page.goto('/telegram-destinations/dest-1')

  await expect(page.getByRole('combobox', { name: 'Trigger' })).toContainText('Mention only')
  await expect(page.getByRole('combobox', { name: 'Conversation history' })).toContainText(
    'Separate per user'
  )
  await expect(page.getByRole('combobox', { name: 'Reply style' })).toContainText(
    'Send a new message'
  )
  await expect(page.getByLabel('Allowed user IDs')).toHaveValue('10, 20')
  await expect(page.getByLabel('Controller user IDs')).toHaveValue('10')
  await expect(page.getByLabel('Debug by default')).toBeChecked()

  // The address is immutable, so the form must not offer to change it.
  await expect(page.getByLabel('Chat ID')).toBeDisabled()
  await expect(page.getByLabel('Topic ID (optional)')).toBeDisabled()
})

// A controller outside the allow-list could never reach the destination, so
// the server refuses it and the form has to say so.
test('surfaces the controller-must-be-admitted rule', async ({ page }) => {
  await setupPolicyRoutes(page, (route) =>
    fulfillConnectError(
      route,
      'invalid_argument',
      'config.controller_user_ids user 30 must also appear in allowed_user_ids'
    )
  )

  await page.goto('/telegram-destinations/dest-1')
  await page.getByLabel('Controller user IDs').fill('30')
  await page.getByRole('button', { name: 'Save' }).click()

  await expect(page.getByText(/must also appear in allowed_user_ids/).first()).toBeVisible()
})
