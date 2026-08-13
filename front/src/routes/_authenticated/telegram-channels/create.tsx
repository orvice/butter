import { createFileRoute } from '@tanstack/react-router'
import { TelegramChannelCreate } from '@/features/telegram/channel-create'

export const Route = createFileRoute('/_authenticated/telegram-channels/create')({
  component: TelegramChannelCreate,
})
