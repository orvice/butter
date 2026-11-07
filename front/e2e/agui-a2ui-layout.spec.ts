import { expect, test } from '@playwright/test'
import { sendMessage, setupAGUI, sse } from './support/agui'

const version = 'v0.9.1'
const surfaceId = 'layout-card'
const longValue = 'build_artifact_'.repeat(12)

function ui(envelope: Record<string, unknown>, seq: number) {
  return {
    type: 'CUSTOM',
    name: 'butter.a2ui',
    value: {
      version,
      threadId: 't',
      runId: 'r1',
      messageId: 'a1',
      surfaceId,
      kind: 'card',
      revision: 1,
      seq,
      fallback: 'Release summary (text)',
      envelope: { version, ...envelope },
    },
  }
}

const run = sse([
  { type: 'RUN_STARTED', threadId: 't', runId: 'r1' },
  ui({ createSurface: { surfaceId, catalogId: 'butter-basic-v1' } }, 0),
  ui(
    {
      updateComponents: {
        surfaceId,
        components: [
          { id: 'root', component: 'Card', child: 'col' },
          {
            id: 'col', component: 'Column',
            children: ['title', 'row', 'divider', 'summary', 'status'],
          },
          { id: 'title', component: 'Text', text: 'Release summary', variant: 'h3' },
          {
            id: 'row', component: 'Row', align: 'start', justify: 'spaceBetween',
            children: ['left', 'right'],
          },
          { id: 'left', component: 'Column', children: ['env', 'value'] },
          { id: 'env', component: 'Text', text: 'Production' },
          { id: 'value', component: 'Text', text: longValue },
          { id: 'right', component: 'Column', children: ['region'] },
          { id: 'region', component: 'KeyValue', label: 'Region', value: 'EU' },
          { id: 'divider', component: 'Divider', axis: 'horizontal' },
          { id: 'summary', component: 'Text', text: { path: '/summary' } },
          { id: 'status', component: 'Status', text: 'Healthy', tone: 'success' },
        ],
      },
    },
    1
  ),
  ui(
    {
      updateDataModel: {
        surfaceId, path: '/', value: { summary: '**Plain text**, not a link: [details](https://example.com)' },
      },
    },
    2
  ),
  { type: 'RUN_FINISHED', threadId: 't', runId: 'r1' },
])

for (const width of [1100, 390]) {
  test(`official layouts stay themed and contained at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 844 })
    await setupAGUI(page, { runs: [run] })
    await page.goto('/chat?agent=streamer-id', { waitUntil: 'networkidle' })
    await sendMessage(page, 'release')

    const surface = page.locator(`[data-a2ui-surface="${surfaceId}"]`)
    const card = surface.locator('.a2ui-card')
    await expect(card.getByRole('heading', { name: 'Release summary' })).toBeVisible()
    await expect(card.getByText(longValue)).toBeVisible()
    await expect(card.getByText('Healthy')).toBeVisible()
    await expect(card.getByText('**Plain text**, not a link: [details](https://example.com)')).toBeVisible()
    await expect(card.locator('a, strong')).toHaveCount(0)

    // This is the official Column/Row structure, not a second Butter layout
    // implementation. The tokens supply the dashboard's compact spacing.
    const column = card.locator(':scope > div')
    const row = column.locator('[data-a2ui-layout="row"] > div')
    const divider = column.locator(':scope > div').nth(1)
    await expect(column).toHaveCSS('flex-direction', 'column')
    await expect(column).toHaveCSS('gap', '8px')
    await expect(row).toHaveCSS('flex-direction', 'row')
    await expect(row).toHaveCSS('gap', '12px')
    await expect(row).toHaveCSS('flex-wrap', 'wrap')
    await expect(row).toHaveCSS('justify-content', 'space-between')
    await expect(row).toHaveCSS('align-items', 'flex-start')
    await expect(divider).toHaveCSS('height', '1px')
    await expect(divider).toHaveCSS('margin', '0px')
    await expect(card).toHaveCSS('padding', '12px')
    await expect(card).toHaveCSS('margin', '0px')

    const backgrounds: string[] = []
    for (const dark of [false, true]) {
      await page.evaluate((dark) => {
        document.documentElement.classList.toggle('dark', dark)
      }, dark)
      const theme = await card.evaluate((element) => {
        const probe = document.createElement('div')
        probe.style.backgroundColor = 'var(--card)'
        probe.style.color = 'var(--card-foreground)'
        probe.style.border = '1px solid var(--border)'
        element.append(probe)
        const expected = getComputedStyle(probe)
        const actual = getComputedStyle(element)
        const result = {
          background: actual.backgroundColor,
          expectedBackground: expected.backgroundColor,
          foreground: actual.color,
          expectedForeground: expected.color,
          border: actual.borderTopColor,
          expectedBorder: expected.borderTopColor,
        }
        probe.remove()
        return result
      })
      expect(theme.background).toBe(theme.expectedBackground)
      expect(theme.foreground).toBe(theme.expectedForeground)
      expect(theme.border).toBe(theme.expectedBorder)
      await expect(divider).toHaveCSS('background-color', theme.expectedBorder)
      backgrounds.push(theme.background)
    }
    expect(backgrounds[0]).not.toBe(backgrounds[1])

    // A long unbroken value must not escape its unweighted column or card on
    // mobile. Checking the surface (not the page's hidden overflow) catches it.
    expect(await card.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(true)
    const bounds = await surface.boundingBox()
    expect(bounds).not.toBeNull()
    expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(width)
  })
}
