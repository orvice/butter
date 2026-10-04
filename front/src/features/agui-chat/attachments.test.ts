import type { Attachment, PendingAttachment } from '@assistant-ui/react'
import { describe, expect, it } from 'vitest'
import {
  MAX_IMAGE_BYTES,
  MAX_IMAGE_COUNT,
  MAX_TOTAL_IMAGE_BYTES,
} from '@/lib/image-attachments'
import { attachmentImage } from '@/components/chat/images'
import {
  ImageAttachmentAdapter,
  attachmentRefusal,
  imageAttachment,
  readDataURL,
  sendAll,
} from './attachments'

const MiB = 1024 * 1024

function image(name: string, size = 3, type = 'image/png'): File {
  return new File([new Uint8Array(size).fill(7)], name, { type })
}

// composer stands in for the composer the adapter serves: what it holds,
// and an add that shows the attachment once the adapter has read it.
function composer() {
  const held: Attachment[] = []
  const adapter = new ImageAttachmentAdapter()
  adapter.serve(() => held)
  return {
    adapter,
    held,
    async add(file: File) {
      const attachment = await adapter.add({ file })
      held.push(attachment)
      return attachment
    },
  }
}

const refusals = (results: PromiseSettledResult<unknown>[]) =>
  results.flatMap((r) =>
    r.status === 'rejected' ? [(r.reason as Error).message] : []
  )

describe('ImageAttachmentAdapter', () => {
  it('accepts the image types the server accepts', () => {
    expect(composer().adapter.accept).toBe(
      'image/jpeg,image/png,image/gif,image/webp'
    )
  })

  it('reads an image once, when it is added, and sends what it read', async () => {
    const { adapter } = composer()
    const pending = await adapter.add({ file: image('cat.png') })
    expect(pending).toMatchObject({
      type: 'image',
      name: 'cat.png',
      contentType: 'image/png',
      status: { type: 'requires-action', reason: 'composer-send' },
    })
    // The thumbnail is the image itself.
    expect(attachmentImage(pending)).toBe('data:image/png;base64,BwcH')

    const sent = await adapter.send(pending)
    expect(sent.status).toEqual({ type: 'complete' })
    expect(sent.content).toEqual([
      { type: 'image', image: 'data:image/png;base64,BwcH' },
    ])
  })

  it('refuses an image over the per-image limit', async () => {
    const { adapter } = composer()
    await expect(
      adapter.add({ file: image('huge.png', MAX_IMAGE_BYTES + 1) })
    ).rejects.toThrow('huge.png: exceeds the 10 MiB per-image limit')
  })

  it('refuses a type the server does not accept', async () => {
    const { adapter } = composer()
    await expect(
      adapter.add({ file: image('scan.bmp', 3, 'image/bmp') })
    ).rejects.toThrow('scan.bmp: unsupported type image/bmp')
  })

  it('counts the images the composer holds', async () => {
    const c = composer()
    for (let i = 0; i < MAX_IMAGE_COUNT; i++) await c.add(image(`${i}.png`))
    await expect(c.add(image('one-more.png'))).rejects.toThrow(
      'one-more.png: at most 10 images per message'
    )

    // Removing one from the composer makes room again.
    c.held.pop()
    await expect(c.add(image('fits.png'))).resolves.toMatchObject({
      name: 'fits.png',
    })
  })

  it('counts the bytes the composer holds', async () => {
    const c = composer()
    await c.add(image('a.png', 8 * MiB))
    await c.add(image('b.png', 8 * MiB))
    await expect(c.add(image('c.png', 5 * MiB))).rejects.toThrow(
      'c.png: total attachments exceed the 20 MiB limit'
    )
    // What is left still fits.
    await expect(
      c.add(image('d.png', MAX_TOTAL_IMAGE_BYTES - 16 * MiB))
    ).resolves.toMatchObject({ name: 'd.png' })
  })

  it('counts files added at once against each other', async () => {
    // The picker, paste and drop add every file at once, each before the
    // composer shows the ones before it.
    const c = composer()
    const files = Array.from({ length: MAX_IMAGE_COUNT + 2 }, (_, i) =>
      image(`${i}.png`)
    )
    const results = await Promise.allSettled(files.map((f) => c.add(f)))
    expect(c.held.map((a) => a.name)).toEqual(
      files.slice(0, MAX_IMAGE_COUNT).map((f) => f.name)
    )
    expect(refusals(results)).toEqual([
      '10.png: at most 10 images per message',
      '11.png: at most 10 images per message',
    ])

    const big = composer()
    const sized = await Promise.allSettled(
      ['a', 'b', 'c'].map((n) => big.add(image(`${n}.png`, 8 * MiB)))
    )
    expect(big.held.map((a) => a.name)).toEqual(['a.png', 'b.png'])
    expect(refusals(sized)).toEqual([
      'c.png: total attachments exceed the 20 MiB limit',
    ])
  })

  it('holds nothing for a file it refused or finished adding', async () => {
    const { adapter } = composer()
    await expect(
      adapter.add({ file: image('huge.png', MAX_IMAGE_BYTES + 1) })
    ).rejects.toThrow()
    // Added but never shown, as when the composer drops an add: it no
    // longer counts.
    for (let i = 0; i < MAX_IMAGE_COUNT; i++) {
      await adapter.add({ file: image(`${i}.png`) })
    }
    await expect(adapter.add({ file: image('fits.png') })).resolves.toBeTruthy()
  })

  it('reads an image the composer has not read yet when it is sent', async () => {
    const { adapter } = composer()
    const pending: PendingAttachment = {
      id: 'p1',
      type: 'image',
      name: 'cat.png',
      contentType: 'image/png',
      file: image('cat.png'),
      status: { type: 'requires-action', reason: 'composer-send' },
    }
    await expect(adapter.send(pending)).resolves.toMatchObject({
      status: { type: 'complete' },
      content: [{ type: 'image', image: 'data:image/png;base64,BwcH' }],
    })
  })
})

describe('sendAll', () => {
  it('finishes what the composer holds, as its send would', async () => {
    const c = composer()
    const pending = await c.add(image('cat.png'))
    const done = await c.adapter.send(
      await c.adapter.add({ file: image('dog.png') })
    )
    const sent = await sendAll(c.adapter, [pending, done])
    expect(sent.map((a) => [a.name, a.status.type, a.content])).toEqual([
      [
        'cat.png',
        'complete',
        [{ type: 'image', image: 'data:image/png;base64,BwcH' }],
      ],
      [
        'dog.png',
        'complete',
        [{ type: 'image', image: 'data:image/png;base64,BwcH' }],
      ],
    ])
    // An attachment already finished is passed on as it is.
    expect(sent[1]).toBe(done)
  })
})

describe('imageAttachment', () => {
  it('is an image as a sent message holds it', async () => {
    const attachment = await imageAttachment(image('cat.webp', 3, 'image/webp'))
    expect(attachment).toMatchObject({
      type: 'image',
      name: 'cat.webp',
      contentType: 'image/webp',
      status: { type: 'complete' },
      content: [{ type: 'image', image: 'data:image/webp;base64,BwcH' }],
    })
    expect(attachment.id).toBeTruthy()
  })
})

describe('attachmentRefusal', () => {
  it('passes on what the adapter said', () => {
    expect(
      attachmentRefusal({
        reason: 'adapter-error',
        message: 'huge.png: exceeds the 10 MiB per-image limit',
      })
    ).toBe('huge.png: exceeds the 10 MiB per-image limit')
  })

  it('says which types are accepted when the composer refused the type', () => {
    expect(
      attachmentRefusal({
        reason: 'not-accepted',
        message: 'File type application/pdf is not accepted.',
        contentType: 'application/pdf',
      })
    ).toBe('Unsupported type application/pdf; accepted: JPEG, PNG, GIF, WebP')
    expect(
      attachmentRefusal({ reason: 'not-accepted', message: 'not accepted' })
    ).toBe('Unsupported type unknown; accepted: JPEG, PNG, GIF, WebP')
  })
})

describe('readDataURL', () => {
  it('encodes the bytes as base64 under the file type', async () => {
    // Longer than the chunks the bytes are encoded in.
    const bytes = new Uint8Array(0x8000 * 2 + 5).map((_, i) => i % 256)
    const url = await readDataURL(
      new File([bytes], 'big.gif', { type: 'image/gif' })
    )
    let binary = ''
    for (const b of bytes) binary += String.fromCharCode(b)
    expect(url).toBe(`data:image/gif;base64,${btoa(binary)}`)
  })
})
