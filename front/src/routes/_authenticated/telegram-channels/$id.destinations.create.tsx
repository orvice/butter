import { createFileRoute } from '@tanstack/react-router'
import { TelegramDestinationForm } from '@/features/telegram/destination-form'

export const Route = createFileRoute(
  '/_authenticated/telegram-channels/$id/destinations/create'
)({
  component: () => <TelegramDestinationForm mode='create' />,
})
