import { createFileRoute } from '@tanstack/react-router'
import { AdminLinearSettingsPage } from '@/features/admin/linear-settings'

export const Route = createFileRoute('/_authenticated/admin/linear')({
  component: AdminLinearSettingsPage,
})
