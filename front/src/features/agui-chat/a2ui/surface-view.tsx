import { Component, useEffect, useRef, type ReactNode } from 'react'
import { A2uiSurface } from '@a2ui/react/v0_9'
import { CheckCircle2, TriangleAlert } from 'lucide-react'
import { cn } from '@/lib/utils'
import { FormStateContext, type FormState } from './catalog'
import { useA2UIStore, type A2UIStore, type SurfaceEntry } from './store'

// SurfaceBoundary confines a renderer failure to its own surface: the
// surface shows its fallback text and the rest of the conversation keeps
// working. It retries whenever the surface receives a newer revision.
class SurfaceBoundary extends Component<
  { fallback: ReactNode; children: ReactNode },
  { failed: boolean }
> {
  state = { failed: false }

  static getDerivedStateFromError() {
    return { failed: true }
  }

  render() {
    return this.state.failed ? this.props.fallback : this.props.children
  }
}

function Fallback({ entry }: { entry: SurfaceEntry }) {
  return (
    <div
      className='rounded-lg border border-dashed border-border bg-muted/40 px-3 py-2 text-sm'
      data-a2ui-fallback={entry.id}
    >
      <p className='mb-1 flex items-center gap-1.5 text-xs text-muted-foreground'>
        <TriangleAlert className='size-3.5' />
        {entry.kind === 'form'
          ? 'This form cannot be displayed here; answer in text instead.'
          : 'This card cannot be displayed; showing its text version.'}
      </p>
      {entry.fallback && (
        <p className='whitespace-pre-wrap'>{entry.fallback}</p>
      )}
    </div>
  )
}

function FormStatusLine({ entry }: { entry: SurfaceEntry }) {
  switch (entry.formStatus) {
    case 'submitting':
      return (
        <p className='mt-2 text-xs text-muted-foreground' aria-live='polite'>
          Submitting…
        </p>
      )
    case 'answered':
      return (
        <p
          className='mt-2 flex items-center gap-1.5 text-xs text-emerald-700 dark:text-emerald-300'
          aria-live='polite'
        >
          <CheckCircle2 className='size-3.5' />
          {entry.formMessage ?? 'Submitted'}
        </p>
      )
    case 'error':
      return entry.formMessage ? (
        <p role='alert' className='mt-2 text-xs text-destructive'>
          {entry.formMessage}
        </p>
      ) : null
    default:
      return null
  }
}

// A2UISurfaceView renders one surface with the butter-basic-v1 catalog. A
// form adds its submission state; `locked` disables it while another run is
// in progress on the thread.
export function A2UISurfaceView({
  store,
  surfaceId,
  locked = false,
  className,
}: {
  store: A2UIStore
  surfaceId: string
  locked?: boolean
  className?: string
}) {
  useA2UIStore(store)
  const entry = store.entry(surfaceId)
  const sectionRef = useRef<HTMLElement>(null)
  const fieldErrors = entry?.fieldErrors

  // Move focus to the first invalid field so keyboard and screen-reader
  // users land on what needs fixing.
  useEffect(() => {
    if (!fieldErrors || Object.keys(fieldErrors).length === 0) return
    const target = sectionRef.current?.querySelector<HTMLElement>(
      'input[aria-invalid="true"], textarea[aria-invalid="true"], [aria-invalid="true"] button'
    )
    target?.focus()
  }, [fieldErrors])

  if (!entry || entry.deleted) return null
  const surface = store.processor.getSurface(surfaceId)
  const fallback = <Fallback entry={entry} />
  if (entry.error || !surface) {
    return <div className={cn('max-w-[85%]', className)}>{fallback}</div>
  }
  const rendered = (
    <SurfaceBoundary key={entry.revision} fallback={fallback}>
      <A2uiSurface surface={surface} />
    </SurfaceBoundary>
  )

  if (entry.kind !== 'form') {
    return (
      <div
        className={cn('max-w-[85%]', className)}
        data-a2ui-surface={entry.id}
      >
        {rendered}
      </div>
    )
  }
  const state: FormState = {
    disabled:
      locked ||
      entry.formStatus === 'submitting' ||
      entry.formStatus === 'answered',
    submitting: entry.formStatus === 'submitting',
    fieldErrors: entry.fieldErrors,
  }
  return (
    <section
      ref={sectionRef}
      aria-label={entry.form?.title ?? 'Form'}
      className={cn('max-w-[85%]', className)}
      data-a2ui-surface={entry.id}
      data-form-status={entry.formStatus}
    >
      <FormStateContext.Provider value={state}>
        {rendered}
      </FormStateContext.Provider>
      <FormStatusLine entry={entry} />
    </section>
  )
}
