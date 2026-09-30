import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useCreateLinearApp } from '@/api/linear'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { LinearAppForm, type LinearAppFormValues } from './form'

export function LinearAppCreate() {
  const navigate = useNavigate()
  const createMutation = useCreateLinearApp()

  function onSubmit(values: LinearAppFormValues) {
    createMutation.mutate(values, {
      onSuccess: (app) => {
        toast.success('Linear App registered')
        navigate({ to: '/linear-apps/$id/edit', params: { id: app.id } })
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Register Linear App'
        subtitle='Route a Linear OAuth app to an Agent, so people can delegate issues to it or mention it in Linear.'
      />
      <PageScroll className='max-w-3xl'>
        <LinearAppForm
          mode='create'
          submitLabel='Register'
          loading={createMutation.isPending}
          onCancel={() => navigate({ to: '/linear-apps' })}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  )
}
