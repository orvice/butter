import { Undo2 } from 'lucide-react'
import { cn } from '@/lib/utils'

// RunNoticeContent is how a run that did not succeed ended, as a chat shows
// it under the conversation: it failed, with the error it reported, or a
// person stopped it. input is the text of the turn it started from, and
// images how many images that turn carried.
export interface RunNoticeContent {
  status: 'failed' | 'stopped'
  error?: string
  input?: string
  images?: number
}

// RunNotice keeps a failed or stopped turn visible, in Chat and in AG-UI
// Chat: it quotes the turn, says how the run ended, and offers Restore
// input, which puts the turn back in the composer. Sending it again starts a
// new run, as the notice warns. Without onRestore there is nothing to put
// back, and the notice offers nothing.
export function RunNotice({
  notice,
  onRestore,
}: {
  notice: RunNoticeContent
  onRestore?: () => void
}) {
  const failed = notice.status === 'failed'
  const quote = quoteOf(notice)
  return (
    <div
      role={failed ? 'alert' : 'status'}
      className={cn(
        'mb-4 rounded-lg border px-3.5 py-3 text-sm',
        failed
          ? 'border-destructive/35 bg-destructive/5'
          : 'border-border/70 bg-muted/30'
      )}
    >
      {quote && (
        <p className='mb-2 truncate border-l-2 border-border pl-2 text-[0.85rem] text-muted-foreground italic'>
          {quote}
        </p>
      )}
      <p
        className={cn(
          'font-medium',
          failed ? 'text-destructive' : 'text-foreground'
        )}
      >
        {failed ? 'This run failed' : 'Stopped'}
      </p>
      <p className='mt-1 text-[0.85rem] leading-5 text-muted-foreground'>
        {failed
          ? notice.error || 'The run ended with an error.'
          : 'You stopped this response before it finished.'}
      </p>
      {onRestore && (
        <div className='mt-2 flex flex-wrap items-center gap-x-3 gap-y-1.5'>
          <button
            type='button'
            onClick={onRestore}
            className='inline-flex touch-manipulation items-center gap-1.5 rounded-md border border-border bg-background px-2.5 py-1.5 text-xs font-medium transition-[background-color,scale] hover:bg-muted active:scale-[0.97] motion-reduce:active:scale-100'
          >
            <Undo2 className='size-3.5' />
            Restore input
          </button>
          <span className='text-[0.75rem] leading-4 text-muted-foreground/90'>
            Sending again starts a new run and may repeat external tool actions.
          </span>
        </div>
      )}
    </div>
  )
}

// quoteOf is the turn as the notice quotes it: its text, and how many images
// went with it.
function quoteOf({ input, images = 0 }: RunNoticeContent): string {
  const imageCount =
    images === 0 ? '' : images === 1 ? '1 image' : `${images} images`
  return [input, imageCount].filter(Boolean).join(' · ')
}
