import type { Attachment } from '@assistant-ui/react'

// attachmentImage is the image an attachment shows: the data URL of its
// image part. Chat reads an image when it is attached, so the
// composer's thumbnail and the sent message show the same image.
export function attachmentImage(attachment: Attachment): string | undefined {
  for (const part of attachment.content ?? []) {
    if (part.type === 'image') return part.image
  }
  return undefined
}
