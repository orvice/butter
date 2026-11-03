import { defineConfig } from '@playwright/test'

const port = Number(process.env.PLAYWRIGHT_PORT ?? 5173)

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: `http://localhost:${port}`,
    headless: true,
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: `npm run dev -- --port ${port} --strictPort`,
    port,
    reuseExistingServer: !process.env.PLAYWRIGHT_PORT,
  },
})
