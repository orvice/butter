import {
  generateId,
  type Attachment,
  type AttachmentAdapter,
  type CompleteAttachment,
  type PendingAttachment,
} from '@assistant-ui/react'
import { IMAGE_FILE_ACCEPT, acceptImageFiles } from '@/lib/image-attachments'

// ImageAttachmentAdapter is how AG-UI Chat's composer takes images. It
// accepts what the server accepts in one message (lib/image-attachments):
// JPEG, PNG, GIF and WebP, at most 10 MiB each, 10 of them and 20 MiB in
// all. A file is checked when it is added, and one over a limit is refused
// with a message that names it and the limit; the composer reports the
// refusal as composer.attachmentAddError and never shows the file. A type
// outside `accept` is refused by the composer before it gets here.
//
// The limits count what the composer holds (serve) and the files still being
// added: the picker, paste and drop add several files at once, and the
// composer shows each one only once it has been read.
//
// An image is read once, when it is added. Its data URL is the composer's
// thumbnail and, once sent, the message's image, which the AG-UI runtime
// sends as an image part with a data source.
export class ImageAttachmentAdapter implements AttachmentAdapter {
  readonly accept = IMAGE_FILE_ACCEPT
  private held: () => readonly Attachment[] = () => []
  private readonly adding = new Map<symbol, File>()

  // serve makes the limits count the attachments `held` reads: the ones the
  // composer this adapter serves holds. That composer exists only once the
  // runtime the adapter is given to is up.
  serve(held: () => readonly Attachment[]) {
    this.held = held
  }

  async add({ file }: { file: File }): Promise<PendingAttachment> {
    const existing = [
      ...this.held().flatMap((a) => (a.file ? [a.file] : [])),
      ...this.adding.values(),
    ]
    const [refusal] = acceptImageFiles(existing, [file]).errors
    if (refusal) throw new Error(refusal)
    const key = Symbol(file.name)
    this.adding.set(key, file)
    try {
      return {
        id: generateId(),
        type: 'image',
        name: file.name,
        contentType: file.type,
        file,
        status: { type: 'requires-action', reason: 'composer-send' },
        content: [{ type: 'image', image: await readDataURL(file) }],
      }
    } finally {
      this.adding.delete(key)
    }
  }

  async send(attachment: PendingAttachment): Promise<CompleteAttachment> {
    return {
      ...attachment,
      status: { type: 'complete' },
      content: attachment.content ?? [
        { type: 'image', image: await readDataURL(attachment.file) },
      ],
    }
  }

  async remove() {
    // Nothing outside the composer holds an image.
  }
}

// sendAll finishes a composer's attachments, as its own send would, for a
// message sent around it. An answer to an open question is one: the runtime
// refuses an ordinary send while questions are open.
export function sendAll(
  adapter: AttachmentAdapter,
  attachments: readonly Attachment[]
): Promise<CompleteAttachment[]> {
  return Promise.all(
    attachments.map((a) =>
      a.status.type === 'complete'
        ? (a as CompleteAttachment)
        : adapter.send(a as PendingAttachment)
    )
  )
}

// imageAttachment is an image as a sent message holds it, for a message the
// composer did not send: a thread's first one, written in the new-chat
// draft.
export async function imageAttachment(file: File): Promise<CompleteAttachment> {
  return {
    id: generateId(),
    type: 'image',
    name: file.name,
    contentType: file.type,
    status: { type: 'complete' },
    content: [{ type: 'image', image: await readDataURL(file) }],
  }
}

// attachmentRefusal is what to tell the user about a file the composer
// refused (composer.attachmentAddError). The adapter's refusals name the
// file and the limit it broke; a type the adapter does not accept is refused
// by the composer itself, which knows only the type.
export function attachmentRefusal(event: {
  reason: string
  message: string
  contentType?: string
}): string {
  if (event.reason === 'not-accepted') {
    return `Unsupported type ${event.contentType || 'unknown'}; accepted: JPEG, PNG, GIF, WebP`
  }
  return event.message
}

// readDataURL reads a file as a base64 data URL. Browsers read it with a
// FileReader; elsewhere, as in unit tests, its bytes are encoded here.
export async function readDataURL(file: Blob): Promise<string> {
  if (typeof FileReader !== 'undefined') {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = () =>
        reject(reader.error ?? new Error('could not read the image'))
      reader.readAsDataURL(file)
    })
  }
  const bytes = new Uint8Array(await file.arrayBuffer())
  let binary = ''
  for (let i = 0; i < bytes.length; i += 0x8000) {
    binary += String.fromCharCode(...bytes.subarray(i, i + 0x8000))
  }
  return `data:${file.type || 'application/octet-stream'};base64,${btoa(binary)}`
}
