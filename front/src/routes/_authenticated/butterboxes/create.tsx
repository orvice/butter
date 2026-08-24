import { createFileRoute } from '@tanstack/react-router'
import { ButterBoxCreate } from '@/features/butterboxes/create'

export const Route = createFileRoute('/_authenticated/butterboxes/create')({
  component: ButterBoxCreate,
})
