import {
  WorkStationHostService,
  WorkStationService,
} from '@/gen/agents/v1/workstation_pb'
import { createClient, type Interceptor } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { BASE_URL } from './transport'

// Bind both identities when the client is created. A delayed query from the old
// workspace must never fetch with the new workspace's mutable storage header.
function identityTransport(workspaceId: string, token: string) {
  const identity: Interceptor = (next) => async (request) => {
    request.header.set('Authorization', `Bearer ${token}`)
    request.header.set('X-Workspace-ID', workspaceId)
    return next(request)
  }
  return createConnectTransport({
    baseUrl: `${BASE_URL}/api`,
    useBinaryFormat: true,
    interceptors: [identity],
  })
}
export function workStationClient(workspaceId: string, token: string) {
  return createClient(WorkStationService, identityTransport(workspaceId, token))
}
export function workStationHostClient(workspaceId: string, token: string) {
  return createClient(
    WorkStationHostService,
    identityTransport(workspaceId, token)
  )
}
export type WorkStationClient = ReturnType<typeof workStationClient>
export type WorkStationHostClient = ReturnType<typeof workStationHostClient>
