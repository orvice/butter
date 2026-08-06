import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { RemoteAgentCreate } from '@/features/remote-agents/create'

export const Route = createFileRoute('/_authenticated/remote-agents/create')({
  validateSearch: z.object({
    daemon_runtime_id: z.string().optional(),
    acp_runtime: z.string().optional(),
  }),
  component: RemoteAgentCreate,
})
