import { expect, test, type Page } from '@playwright/test'
import { create, toBinary } from '@bufbuild/protobuf'
import {
  MCPServerAuthType,
  MCPServerTransport,
} from '../src/gen/agents/v1/agent_pb'
import { GetMCPServerResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import { MeResponseSchema } from '../src/gen/agents/v1/auth_pb'
import { ListWorkspacesResponseSchema } from '../src/gen/agents/v1/workspace_pb'

async function setupMCPServerEdit(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('butter_token', 'fake-test-token')
  })

  await page.route('**/api/agents.v1.**', async (route) => {
    const url = route.request().url()

    if (url.includes('AuthService/Me')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/proto',
        body: Buffer.from(
          toBinary(
            MeResponseSchema,
            create(MeResponseSchema, {
              user: {
                id: 'test-user-1',
                username: 'testuser',
                displayName: 'Test User',
                email: 'test@example.com',
                role: 'admin',
              },
            })
          )
        ),
      })
    }

    if (url.includes('WorkspaceService')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/proto',
        body: Buffer.from(
          toBinary(
            ListWorkspacesResponseSchema,
            create(ListWorkspacesResponseSchema, {
              workspaces: [
                { id: 'default', name: 'Default', slug: 'default' },
              ],
            })
          )
        ),
      })
    }

    if (url.includes('MCPServerService/GetMCPServer')) {
      return route.fulfill({
        status: 200,
        contentType: 'application/proto',
        body: Buffer.from(
          toBinary(
            GetMCPServerResponseSchema,
            create(GetMCPServerResponseSchema, {
              mcpServer: {
                id: 'finance',
                name: 'Finance',
                transport:
                  MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
                url: 'https://example.com/mcp',
                auth: {
                  type:
                    MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
                },
              },
            })
          )
        ),
      })
    }

    return route.fulfill({
      status: 200,
      contentType: 'application/proto',
      body: Buffer.alloc(0),
    })
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
