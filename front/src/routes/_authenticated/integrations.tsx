import { createFileRoute } from '@tanstack/react-router'
import { IntegrationsPage } from '@/features/integrations'

export const Route = createFileRoute('/_authenticated/integrations')({
  component: IntegrationsPage,
})
