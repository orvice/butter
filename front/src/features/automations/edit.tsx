import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { toast } from 'sonner'
import { ArrowLeft, ChevronRight } from 'lucide-react'
import { useCronJob, useUpdateCronJob } from '@/api/cron'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Skeleton } from '@/components/ui/skeleton'
import { AutomationForm } from './form'
import type { CronJob } from '@/types/api'

export function AutomationEditPage() {
  const { name } = useParams({ from: '/_authenticated/automations/$name/edit' })
  const navigate = useNavigate()
  const { data, isLoading } = useCronJob(name ?? '')
  const updateMutation = useUpdateCronJob()

  const job = data?.cron_job

  function handleSubmit(next: CronJob) {
    updateMutation.mutate(next, {
      onSuccess: () => {
        toast.success('Automation updated')
        navigate({ to: '/automations/$name', params: { name: next.name } })
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

  if (!job) {
    return (
      <div className='flex h-full items-center justify-center'>
        <p className='text-sm text-muted-foreground'>Automation not found.</p>
      </div>
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title={`Edit ${job.name}`}
        subtitle='Update the schedule, agent, reliability, and delivery settings.'
        breadcrumb={
          <Link
            to='/automations/$name'
            params={{ name: job.name }}
            className='inline-flex min-w-0 items-center gap-1.5 hover:text-foreground'
          >
            <ArrowLeft className='size-3.5 shrink-0' />
            <span>Automations</span>
            <ChevronRight className='size-3 shrink-0' />
            <span className='truncate'>{job.name}</span>
            <ChevronRight className='size-3 shrink-0' />
            <span className='text-foreground'>Edit</span>
          </Link>
        }
      />
      <PageScroll className='max-w-3xl'>
        <AutomationForm
          mode='edit'
          initialValue={job}
          loading={updateMutation.isPending}
          submitLabel='Save Changes'
          onCancel={() => navigate({ to: '/automations/$name', params: { name: job.name } })}
          onSubmit={handleSubmit}
        />
      </PageScroll>
    </Page>
  )
}
