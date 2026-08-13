import { createFileRoute } from '@tanstack/react-router'
import { AdminTelegramSettingsPage } from '@/features/admin/telegram-settings'

export const Route = createFileRoute('/_authenticated/admin/telegram')({
  component: AdminTelegramSettingsPage,
})
