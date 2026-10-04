import type { ComponentProps } from 'react'
import {
  AttachmentPrimitive,
  AuiIf,
  ComposerPrimitive,
  type Attachment,
} from '@assistant-ui/react'
import { Paperclip, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { attachmentImage } from '@/components/chat/images'

// The images a composer holds show as thumbnails above it, each with a
// button that removes it: in an open thread's composer (ComposerImages) and
// in the new-chat draft's.

// ComposerImages shows the images the thread's composer holds.
export function ComposerImages({ className }: { className?: string }) {
  return (
    <AuiIf condition={(s) => s.composer.attachments.length > 0}>
      <div className={cn('flex flex-wrap gap-1.5', className)}>
        <ComposerPrimitive.Attachments>
          {({ attachment }) => <ComposerImage attachment={attachment} />}
        </ComposerPrimitive.Attachments>
      </div>
    </AuiIf>
  )
}

function ComposerImage({ attachment }: { attachment: Attachment }) {
  return (
    <AttachmentPrimitive.Root className='relative inline-flex'>
      <ImageThumb src={attachmentImage(attachment)} name={attachment.name} />
      <AttachmentPrimitive.Remove asChild>
        <RemoveImageButton name={attachment.name} />
      </AttachmentPrimitive.Remove>
    </AttachmentPrimitive.Root>
  )
}

// AttachImagesButton opens the file picker of the thread's composer. The
// composer also takes images pasted into it or dropped on the chat.
export function AttachImagesButton({ className }: { className?: string }) {
  return (
    <ComposerPrimitive.AddAttachment asChild>
      <AttachButton className={className} />
    </ComposerPrimitive.AddAttachment>
  )
}

export function ImageThumb({ src, name }: { src?: string; name: string }) {
  return (
    <img
      src={src}
      alt={name}
      title={name}
      className='size-14 rounded-md object-cover outline outline-1 -outline-offset-1 outline-black/10 dark:outline-white/10'
    />
  )
}

export function RemoveImageButton({
  name,
  className,
  ...props
}: ComponentProps<'button'> & { name: string }) {
  return (
    <button
      type='button'
      aria-label={`Remove ${name}`}
      className={cn(
        'absolute -top-2 -right-2 inline-flex size-6 touch-manipulation items-center justify-center rounded-full border border-border bg-background text-muted-foreground shadow-sm transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100',
        className
      )}
      {...props}
    >
      <X className='size-3' />
    </button>
  )
}

export function AttachButton({
  className,
  ...props
}: ComponentProps<'button'>) {
  return (
    <button
      type='button'
      aria-label='Attach images'
      title='Attach images'
      className={cn(
        'inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] disabled:pointer-events-none disabled:opacity-50 motion-reduce:active:scale-100',
        className
      )}
      {...props}
    >
      <Paperclip className='size-4' />
    </button>
  )
}
