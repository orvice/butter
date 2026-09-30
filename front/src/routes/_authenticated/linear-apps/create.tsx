import { createFileRoute } from '@tanstack/react-router'
import { LinearAppCreate } from '@/features/linear/create'

export const Route = createFileRoute('/_authenticated/linear-apps/create')({
  component: LinearAppCreate,
})
