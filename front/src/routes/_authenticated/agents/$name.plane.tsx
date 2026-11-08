import { createFileRoute } from '@tanstack/react-router'
import { PlaneAgentSetup } from '@/features/plane/agent-setup'

export const Route = createFileRoute('/_authenticated/agents/$name/plane')({
  component: PlaneAgentSetup,
})
