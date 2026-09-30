import { createFileRoute } from '@tanstack/react-router'
import { LinearProcessingList } from '@/features/linear/processing-list'

export const Route = createFileRoute('/_authenticated/linear-deliveries')({
  component: LinearProcessingList,
})
