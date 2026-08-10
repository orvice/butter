import { expect, test, type Page } from '@playwright/test'
import { ListAgentsResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import {
  CronConcurrencyPolicy,
  CronDeliveryType,
  CronNotifyOn,
  GetCronJobResponseSchema,
} from '../src/gen/agents/v1/cron_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

async function setupAutomationEdit(page: Page) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('AgentService/ListAgents')) {
      await new Promise((resolve) => setTimeout(resolve, 150))
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [{ name: 'TicketManager', agentId: 'ticketmanager' }],
        total: 1,
      })
    }

    if (url.includes('CronJobService/GetCronJob')) {
      return fulfillProto(route, GetCronJobResponseSchema, {
        cronJob: {
          name: 'daily-ticket',
          schedule: '0 9 * * *',
          agentName: 'TicketManager',
          input: 'Review open tickets',
          timezone: 'UTC',
          enabled: true,
          delivery: {
            type: CronDeliveryType.WEBHOOK,
            webhookUrl: 'https://example.com/automation',
          },
          concurrencyPolicy: CronConcurrencyPolicy.UNSPECIFIED,
          notifyOn: CronNotifyOn.UNSPECIFIED,
          maxOutputBytes: 4096,
        },
      })
    }

    return false
  })
}

test('selects the values returned by the automation API', async ({ page }) => {
  await setupAutomationEdit(page)

  await page.goto('/automations/daily-ticket/edit')

  await expect(page.getByRole('combobox', { name: 'Agent' })).toHaveText(
    'TicketManager'
  )
  await expect(page.getByRole('combobox', { name: 'Concurrency' })).toHaveText(
    'Skip'
  )
  await expect(page.getByRole('combobox', { name: 'Notify On' })).toHaveText(
    'Always'
  )
  await expect(page.getByRole('combobox', { name: 'Type' })).toHaveText(
    'Webhook'
  )
})
