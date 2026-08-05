import { createFileRoute } from '@tanstack/react-router'
import { AutomationEditPage } from '@/features/automations/edit'

export const Route = createFileRoute('/_authenticated/automations/$name/edit')({
  component: AutomationEditPage,
})
