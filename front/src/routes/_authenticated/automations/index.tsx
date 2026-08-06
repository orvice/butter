import { createFileRoute } from '@tanstack/react-router'
import { AutomationListPage } from '@/features/automations/list'

export const Route = createFileRoute('/_authenticated/automations/')({
  component: AutomationListPage,
})
