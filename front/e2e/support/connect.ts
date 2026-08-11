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

// connectStreamBody frames messages using the Connect streaming envelope
// (1 flag byte + 4-byte big-endian length per message, then an EndStream
// frame with flag 0x02), matching what connect-web expects from a
// server-stream RPC over the binary transport.
export function connectStreamBody<T extends DescMessage>(
  schema: T,
  messages: MessageInitShape<T>[]
): Buffer {
  const chunks: Buffer[] = []
  for (const value of messages) {
    const bin = toBinary(schema, create(schema, value))
    const head = Buffer.alloc(5)
    head.writeUInt8(0, 0)
    head.writeUInt32BE(bin.length, 1)
    chunks.push(head, Buffer.from(bin))
  }
  const end = Buffer.from(JSON.stringify({}))
  const endHead = Buffer.alloc(5)
  endHead.writeUInt8(2, 0)
  endHead.writeUInt32BE(end.length, 1)
  chunks.push(endHead, end)
  return Buffer.concat(chunks)
}

// fulfillConnectStream responds to a Connect server-stream request with the
// given ordered messages followed by a clean end-of-stream frame.
export async function fulfillConnectStream<T extends DescMessage>(
  route: Route,
  schema: T,
  messages: MessageInitShape<T>[]
): Promise<true> {
  await route.fulfill({
    status: 200,
    contentType: 'application/connect+proto',
    body: connectStreamBody(schema, messages),
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
