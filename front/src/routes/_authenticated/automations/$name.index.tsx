import { createFileRoute } from '@tanstack/react-router'
import { AutomationDetailPage } from '@/features/automations/detail'

export const Route = createFileRoute('/_authenticated/automations/$name/')({
  component: AutomationDetailPage,
})
