import { createFileRoute } from '@tanstack/react-router'
import { ModelProviderEdit } from '@/features/model-providers/edit'

export const Route = createFileRoute('/_authenticated/model-providers/$name/edit')({
  component: ModelProviderEdit,
})
