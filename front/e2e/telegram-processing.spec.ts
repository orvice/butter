import { expect, test, type Page, type Route } from '@playwright/test'
import { fromBinary } from '@bufbuild/protobuf'
import {
  ListTelegramProcessingRecordsRequestSchema,
  ListTelegramProcessingRecordsResponseSchema,
  ResendTelegramReplyRequestSchema,
  ResendTelegramReplyResponseSchema,
  TelegramProcessingStatus,
} from '../src/gen/agents/v1/telegram_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const RECORDS = [
  {
    id: 'rec-succeeded',
    workspaceId: 'default',
    channelId: 'ch-1',
    destinationId: 'dest-1',
    updateId: 101n,
    status: TelegramProcessingStatus.SUCCEEDED,
    invocationId: 'inv-1',
    attempts: 1,
    segments: [{ index: 0, text: 'done', status: 'sent', messageId: '900' }],
  },
  {
    id: 'rec-partial',
    workspaceId: 'default',
    channelId: 'ch-1',
    destinationId: 'dest-1',
    updateId: 102n,
    status: TelegramProcessingStatus.FAILED,
    invocationId: 'inv-2',
    attempts: 2,
    error: 'telegram error 500: Internal Server Error',
    output: 'part one\npart two',
    segments: [
      { index: 0, text: 'part one', status: 'sent', messageId: '901' },
      { index: 1, text: 'part two', status: 'failed', error: 'telegram error 500' },
    ],
  },
  {
    id: 'rec-uncertain',
    workspaceId: 'default',
    channelId: 'ch-1',
    destinationId: 'dest-1',
    updateId: 103n,
    status: TelegramProcessingStatus.FAILED_UNCERTAIN,
    invocationId: 'inv-3',
    attempts: 1,
    error: 'model timed out mid-tool-call',
    deadLettered: true,
  },
]

async function setupProcessingRoutes(
  page: Page,
  overrides: { onResend?: (route: Route) => Promise<boolean> } = {}
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramProcessingService/ResendTelegramReply')) {
      if (overrides.onResend) return overrides.onResend(route)
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(ResendTelegramReplyRequestSchema, body)
      return fulfillProto(route, ResendTelegramReplyResponseSchema, {
        record: { ...RECORDS[1], id: req.id, status: TelegramProcessingStatus.SUCCEEDED },
        messageIds: ['902'],
      })
    }
    if (url.includes('TelegramProcessingService/ListTelegramProcessingRecords')) {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(ListTelegramProcessingRecordsRequestSchema, body)
      const filtered =
        req.status === TelegramProcessingStatus.UNSPECIFIED
          ? RECORDS
          : RECORDS.filter((record) => record.status === req.status)
      return fulfillProto(route, ListTelegramProcessingRecordsResponseSchema, {
        records: filtered,
      })
    }
    return false
  })
}

test('lists processing records with their status and failure detail', async ({ page }) => {
  await setupProcessingRoutes(page)

  await page.goto('/telegram-updates')

  await expect(page.getByTestId('processing-rec-succeeded')).toContainText('Succeeded')
  await expect(page.getByTestId('processing-error-rec-partial')).toContainText(
    'Internal Server Error'
  )
  await expect(page.getByTestId('processing-rec-uncertain')).toContainText('Uncertain')
  await expect(page.getByTestId('processing-rec-uncertain')).toContainText(
    'not retried automatically'
  )
})

test('filters by status', async ({ page }) => {
  await setupProcessingRoutes(page)

  await page.goto('/telegram-updates')
  await page.getByRole('combobox', { name: 'Status filter' }).click()
  await page.getByRole('option', { name: 'Needs review' }).click()

  await expect(page.getByTestId('processing-rec-uncertain')).toBeVisible()
  await expect(page.getByTestId('processing-rec-succeeded')).toHaveCount(0)
})

// Resend is offered only when segments are still outstanding, and there is
// deliberately no rerun action anywhere on the page.
test('offers resend only for an incomplete delivery', async ({ page }) => {
  let resentID = ''
  await setupProcessingRoutes(page, {
    onResend: async (route) => {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      resentID = fromBinary(ResendTelegramReplyRequestSchema, body).id
      return fulfillProto(route, ResendTelegramReplyResponseSchema, {
        record: { ...RECORDS[1], status: TelegramProcessingStatus.SUCCEEDED },
        messageIds: ['902'],
      })
    },
  })

  await page.goto('/telegram-updates')

  await expect(page.getByLabel('Resend reply for update 101')).toBeDisabled()
  await expect(page.getByLabel('Resend reply for update 103')).toBeDisabled()
  await expect(page.getByRole('button', { name: /rerun/i })).toHaveCount(0)

  await page.getByLabel('Resend reply for update 102').click()
  await expect(page.getByText('Reply resent').first()).toBeVisible()
  expect(resentID).toBe('rec-partial')
})

test('reports why a resend is refused', async ({ page }) => {
  await setupProcessingRoutes(page, {
    onResend: (route) =>
      fulfillConnectError(
        route,
        'failed_precondition',
        'this record has no complete response to resend'
      ),
  })

  await page.goto('/telegram-updates')
  await page.getByLabel('Resend reply for update 102').click()

  await expect(page.getByText(/no complete response to resend/).first()).toBeVisible()
})
