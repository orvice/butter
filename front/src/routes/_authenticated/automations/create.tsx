import { createFileRoute } from '@tanstack/react-router'
import { AutomationCreatePage } from '@/features/automations/create'

export const Route = createFileRoute('/_authenticated/automations/create')({
  component: AutomationCreatePage,
})
