import { createFileRoute } from '@tanstack/react-router'
import { TelegramProcessingList } from '@/features/telegram/processing-list'

export const Route = createFileRoute('/_authenticated/telegram-updates')({
  component: TelegramProcessingList,
})
