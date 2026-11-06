import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { LinearAgentSetup } from '@/features/linear/agent-setup'

const searchSchema = z.object({
  step: z.number().int().min(1).max(4).optional(),
  app: z.string().optional(),
})

export const Route = createFileRoute('/_authenticated/agents/$name/linear')({
  validateSearch: searchSchema,
  component: LinearAgentSetup,
})
