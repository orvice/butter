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

// Connect error codes and their protocol-defined HTTP statuses (subset used
// by the fixtures).
const CONNECT_ERROR_STATUS: Record<string, number> = {
  not_found: 404,
  failed_precondition: 412,
  permission_denied: 403,
  unauthenticated: 401,
  invalid_argument: 400,
  internal: 500,
}

// fulfillConnectError responds to a unary Connect request with an error in
// the JSON error envelope connect-web expects; the HTTP status is derived
// from the Connect code.
export async function fulfillConnectError(
  route: Route,
  code: keyof typeof CONNECT_ERROR_STATUS,
  message = ''
): Promise<true> {
  await route.fulfill({
    status: CONNECT_ERROR_STATUS[code],
    contentType: 'application/json',
    body: JSON.stringify({ code, message }),
  })
  return true
}

export interface ConnectFixtureOptions {
  // Workspaces the signed-in user belongs to; the first is selected.
  workspaces?: Array<{ id: string; name: string; slug: string }>
}

export async function setupAuthenticatedConnectRoutes(
  page: Page,
  handleRoute: ConnectRouteHandler,
  options: ConnectFixtureOptions = {}
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
        workspaces: options.workspaces ?? [
          { id: 'default', name: 'Default', slug: 'default' },
        ],
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
