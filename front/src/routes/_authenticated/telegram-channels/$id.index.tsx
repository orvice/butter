import { createFileRoute } from '@tanstack/react-router'
import { TelegramChannelDetail } from '@/features/telegram/channel-detail'

export const Route = createFileRoute('/_authenticated/telegram-channels/$id/')({
  component: TelegramChannelDetail,
})
