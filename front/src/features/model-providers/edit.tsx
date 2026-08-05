import { useNavigate, useParams } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useModelProvider, useUpdateModelProvider } from '@/api/model-providers'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Skeleton } from '@/components/ui/skeleton'
import { ModelProviderForm } from './form'
import type { ModelProvider } from '@/types/api'

export function ModelProviderEdit() {
  const { name } = useParams({ from: '/_authenticated/model-providers/$name/edit' })
  const navigate = useNavigate()
  const { data, isLoading } = useModelProvider(name ?? '')
  const updateMutation = useUpdateModelProvider()

  function onSubmit(provider: ModelProvider) {
    updateMutation.mutate(provider, {
      onSuccess: () => {
        toast.success('Model provider updated')
        navigate({ to: '/model-providers' })
      },
      onError: (err) => toast.error(err.message),
    })
  }

  if (isLoading) {
    return (
      <div className='p-6'>
        <Skeleton className='h-96 w-full' />
      </div>
    )
  }

  return (
    <Page>
      <PageHeader className='max-w-3xl' title='Edit Model Provider' subtitle={name} />
      <PageScroll className='max-w-3xl'>
        <ModelProviderForm
          mode='edit'
          initialValue={data?.model_provider}
          submitLabel='Save'
          loading={updateMutation.isPending}
          onCancel={() => navigate({ to: '/model-providers' })}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  )
}
