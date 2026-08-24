import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useCreateButterBox } from '@/api/butterboxes'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { ButterBoxForm, type ButterBoxFormValues } from './form'

export function ButterBoxCreate() {
  const navigate = useNavigate()
  const createMutation = useCreateButterBox()

  function onSubmit(values: ButterBoxFormValues) {
    createMutation.mutate(
      {
        name: values.name,
        baseUrl: values.baseUrl,
        enabled: values.enabled,
        token: values.token ?? '',
      },
      {
        onSuccess: () => {
          toast.success('ButterBox registered')
          navigate({ to: '/butterboxes' })
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Register ButterBox'
        subtitle='Register an agent VM running butter-box; tools and credentials for pi live on the box itself.'
      />
      <PageScroll className='max-w-3xl'>
        <ButterBoxForm
          mode='create'
          submitLabel='Register'
          loading={createMutation.isPending}
          onCancel={() => navigate({ to: '/butterboxes' })}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  )
}
