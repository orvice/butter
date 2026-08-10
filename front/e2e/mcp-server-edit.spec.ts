import { expect, test, type Page } from '@playwright/test'
import {
  MCPServerAuthType,
  MCPServerTransport,
} from '../src/gen/agents/v1/agent_pb'
import { GetMCPServerResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

async function setupMCPServerEdit(page: Page) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('MCPServerService/GetMCPServer')) {
      return fulfillProto(route, GetMCPServerResponseSchema, {
        mcpServer: {
          id: 'finance',
          name: 'Finance',
          transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
          url: 'https://example.com/mcp',
          auth: {
            type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
          },
        },
      })
    }

    return false
  })
}

test('selects the authentication type returned by the MCP server API', async ({
  page,
}) => {
  await setupMCPServerEdit(page)

  await page.goto('/mcp-servers/finance/edit')

  await expect(page.getByRole('combobox', { name: 'Authentication' })).toHaveText(
    'Static headers'
  )
})
