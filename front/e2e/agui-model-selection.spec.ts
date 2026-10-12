import { expect, test } from '@playwright/test'
import { ListModelProvidersResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import { fulfillProto } from './support/connect'
import { setupAGUI, sendMessage, sse, threadInURL } from './support/agui'

function finished(runId: string, text: string) {
  return sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    { type: 'TEXT_MESSAGE_START', messageId: runId, role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: runId, delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: runId },
    { type: 'RUN_FINISHED', threadId: 't', runId },
  ])
}

test('selects a model before the first turn, switches in the same thread and returns to default', async ({ page }) => {
  const requests: Array<Record<string, unknown>> = []
  await setupAGUI(page, { runs: [finished('r1', 'First reply'), finished('r2', 'Second reply'), finished('r3', 'Default reply')], requests })
  await page.route('**/api/agents.v1.ModelProviderService/ListModelProviders', (route) => fulfillProto(route, ListModelProvidersResponseSchema, {
    modelProviders: [{ name: 'Provider', type: 'openai', models: [{ name: 'model-fast', alias: 'fast' }, { name: 'model-quality', alias: 'quality' }] }],
  }))
  await page.goto('/chat?agent=streamer-id')
  await page.getByRole('combobox', { name: 'Chat model' }).click()
  await page.getByRole('option', { name: 'fast · Provider', exact: true }).click()
  await sendMessage(page, 'First turn')
  await expect(page.getByText('First reply', { exact: true })).toBeVisible()
  const thread = threadInURL(page)
  expect(requests[0].forwardedProps).toMatchObject({ butterModel: { override: 'fast' }, butterRun: { detach: true } })
  await expect(page.getByRole('combobox', { name: 'Chat model' })).toContainText('fast')
  await page.getByRole('combobox', { name: 'Chat model' }).click()
  await page.getByRole('option', { name: 'quality · Provider', exact: true }).click()
  await sendMessage(page, 'Second turn')
  await expect(page.getByText('Second reply', { exact: true })).toBeVisible()
  expect(requests[1].forwardedProps).toMatchObject({ butterModel: { override: 'quality' } })
  expect(threadInURL(page)).toBe(thread)
  await page.getByRole('combobox', { name: 'Chat model' }).click()
  await page.getByRole('option', { name: 'Agent default model', exact: true }).click()
  await sendMessage(page, 'Third turn')
  await expect(page.getByText('Default reply', { exact: true })).toBeVisible()
  expect(requests[2].forwardedProps).toMatchObject({ butterModel: { override: '' } })
  expect(requests.every((request) => request.threadId === thread)).toBe(true)
})
