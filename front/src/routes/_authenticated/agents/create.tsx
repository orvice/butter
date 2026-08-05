import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { AgentCreate } from '@/features/agents/create'

export const Route = createFileRoute('/_authenticated/agents/create')({
  validateSearch: z.object({
    remote_agent_id: z.string().optional(),
  }),
  component: AgentCreate,
})
