import { createFileRoute } from '@tanstack/react-router'
import { TelegramChannelList } from '@/features/telegram/channel-list'

export const Route = createFileRoute('/_authenticated/telegram-channels/')({
  component: TelegramChannelList,
})
