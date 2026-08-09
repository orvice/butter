import { createFileRoute } from '@tanstack/react-router'
import { OperationsPage } from '@/features/operations/list'

export const Route = createFileRoute('/_authenticated/operations')({
  component: OperationsPage,
})
