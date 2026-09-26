import { createFileRoute } from '@tanstack/react-router'
import { MemorySettingsPage } from '@/features/memory'

export const Route = createFileRoute('/_authenticated/memory')({
  component: MemorySettingsPage,
})
