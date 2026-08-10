import type { Page, Route } from '@playwright/test'
import {
  create,
  toBinary,
  type DescMessage,
  type MessageInitShape,
} from '@bufbuild/protobuf'
import { MeResponseSchema } from '../../src/gen/agents/v1/auth_pb'
import { ListWorkspacesResponseSchema } from '../../src/gen/agents/v1/workspace_pb'

type ConnectRouteHandler = (route: Route, url: string) => Promise<boolean>

export async function fulfillProto<T extends DescMessage>(
  route: Route,
  schema: T,
  value: MessageInitShape<T>
): Promise<true> {
  await route.fulfill({
    status: 200,
    contentType: 'application/proto',
    body: Buffer.from(toBinary(schema, create(schema, value))),
  })
  return true
}

export async function setupAuthenticatedConnectRoutes(
  page: Page,
  handleRoute: ConnectRouteHandler
) {
  await page.addInitScript(() => {
    localStorage.setItem('butter_token', 'fake-test-token')
  })

  await page.route('**/api/agents.v1.**', async (route) => {
    const url = route.request().url()

    if (url.includes('AuthService/Me')) {
      await fulfillProto(route, MeResponseSchema, {
        user: {
          id: 'test-user-1',
          username: 'testuser',
          displayName: 'Test User',
          email: 'test@example.com',
          role: 'admin',
        },
      })
      return
    }

    if (url.includes('WorkspaceService')) {
      await fulfillProto(route, ListWorkspacesResponseSchema, {
        workspaces: [{ id: 'default', name: 'Default', slug: 'default' }],
      })
      return
    }

    if (await handleRoute(route, url)) return

    await route.fulfill({
      status: 200,
      contentType: 'application/proto',
      body: Buffer.alloc(0),
    })
  })
}
