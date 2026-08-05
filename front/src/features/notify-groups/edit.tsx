import { useNavigate, useParams } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useNotifyGroup, useUpdateNotifyGroup } from '@/api/notify-groups'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Skeleton } from '@/components/ui/skeleton'
import type { NotifyGroup } from '@/types/api'
import { NotifyGroupForm } from './form'

export function NotifyGroupEdit() {
  const { name } = useParams({ from: '/_authenticated/notify-groups/$name/edit' })
  const navigate = useNavigate()
  const { data, isLoading } = useNotifyGroup(name ?? '')
  const updateMutation = useUpdateNotifyGroup()

  function onSubmit(group: NotifyGroup) {
    updateMutation.mutate(group, {
      onSuccess: () => {
        toast.success('Notify group updated')
        navigate({ to: '/notify-groups' })
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
      <PageHeader className='max-w-3xl' title='Edit Notify Group' subtitle={name} />
      <PageScroll className='max-w-3xl'>
        <NotifyGroupForm
          initialValue={data?.notify_group}
          submitLabel='Save'
          submitting={updateMutation.isPending}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  )
}
