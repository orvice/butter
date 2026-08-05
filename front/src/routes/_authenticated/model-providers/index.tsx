import { createFileRoute } from '@tanstack/react-router'
import { ModelProviderList } from '@/features/model-providers/list'

export const Route = createFileRoute('/_authenticated/model-providers/')({
  component: ModelProviderList,
})
