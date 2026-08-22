import { describe, expect, it } from 'vitest'
import {
  acceptImageFiles,
  validateImageFiles,
  MAX_IMAGE_BYTES,
  MAX_IMAGE_COUNT,
  MAX_TOTAL_IMAGE_BYTES,
} from './image-attachments'

function fakeFile(name: string, type: string, sizeBytes: number = 100): File {
  const content = new Uint8Array(sizeBytes)
  return new File([content], name, { type })
}

describe('acceptImageFiles', () => {
  it('accepts JPEG, PNG, GIF, and WebP files', () => {
    const files = [
      fakeFile('a.jpg', 'image/jpeg'),
      fakeFile('b.png', 'image/png'),
      fakeFile('c.gif', 'image/gif'),
      fakeFile('d.webp', 'image/webp'),
    ]
    const result = acceptImageFiles([], files)
    expect(result.accepted).toHaveLength(4)
    expect(result.errors).toHaveLength(0)
  })

  it('rejects unsupported MIME types', () => {
    const files = [
      fakeFile('a.bmp', 'image/bmp'),
      fakeFile('b.svg', 'image/svg+xml'),
      fakeFile('c.pdf', 'application/pdf'),
    ]
    const result = acceptImageFiles([], files)
    expect(result.accepted).toHaveLength(0)
    expect(result.errors).toHaveLength(3)
    expect(result.errors[0]).toContain('unsupported type')
    expect(result.errors[0]).toContain('image/bmp')
  })

  it('rejects files over the per-image limit', () => {
    const bigFile = fakeFile('big.jpg', 'image/jpeg', MAX_IMAGE_BYTES + 1)
    const result = acceptImageFiles([], [bigFile])
    expect(result.accepted).toHaveLength(0)
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0]).toContain('per-image limit')
  })

  it('rejects files exceeding the count limit', () => {
    const existing = Array.from({ length: MAX_IMAGE_COUNT }, (_, i) =>
      fakeFile(`existing-${i}.jpg`, 'image/jpeg')
    )
    const incoming = [fakeFile('one-more.jpg', 'image/jpeg')]
    const result = acceptImageFiles(existing, incoming)
    expect(result.accepted).toHaveLength(0)
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0]).toContain(`at most ${MAX_IMAGE_COUNT}`)
  })

  it('rejects files exceeding the total size limit', () => {
    const almostFull = fakeFile(
      'big.jpg',
      'image/jpeg',
      MAX_TOTAL_IMAGE_BYTES - 10
    )
    const pushesOver = fakeFile('more.jpg', 'image/jpeg', 100)
    const result = acceptImageFiles([almostFull], [pushesOver])
    expect(result.accepted).toHaveLength(0)
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0]).toContain('total attachments exceed')
  })

  it('accepts files just under the limits', () => {
    const file = fakeFile('ok.png', 'image/png', MAX_IMAGE_BYTES)
    const result = acceptImageFiles([], [file])
    expect(result.accepted).toHaveLength(1)
    expect(result.errors).toHaveLength(0)
  })

  it('partially accepts a mixed batch', () => {
    const files = [
      fakeFile('good.jpg', 'image/jpeg'),
      fakeFile('bad.bmp', 'image/bmp'),
      fakeFile('also-good.png', 'image/png'),
    ]
    const result = acceptImageFiles([], files)
    expect(result.accepted).toHaveLength(2)
    expect(result.errors).toHaveLength(1)
  })
})

describe('validateImageFiles', () => {
  it('returns errors for invalid files', () => {
    const errors = validateImageFiles([fakeFile('bad.bmp', 'image/bmp')])
    expect(errors).toHaveLength(1)
  })

  it('returns no errors for valid files', () => {
    const errors = validateImageFiles([fakeFile('ok.jpg', 'image/jpeg')])
    expect(errors).toHaveLength(0)
  })
})
