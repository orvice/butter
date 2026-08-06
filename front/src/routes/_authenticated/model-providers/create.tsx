import { createFileRoute } from '@tanstack/react-router'
import { ModelProviderCreate } from '@/features/model-providers/create'

export const Route = createFileRoute('/_authenticated/model-providers/create')({
  component: ModelProviderCreate,
})
