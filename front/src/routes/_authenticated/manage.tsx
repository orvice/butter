import { createFileRoute } from '@tanstack/react-router'
import { ManagePage } from '@/features/manage'

export const Route = createFileRoute('/_authenticated/manage')({
  component: ManagePage,
})
