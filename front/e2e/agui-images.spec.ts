import { expect, test, type Locator, type Page } from '@playwright/test'
import {
  emptySnapshot,
  sendMessage as send,
  setupAGUI,
  sse,
  threadInURL,
} from './support/agui'

// Chat takes images, picked, pasted or dropped, in an open thread's
// composer and in the new-chat draft, held to the server's limits. A run's
// request carries only the trailing user message, its images as image parts
// with a data source, and a reload shows the images the user sent.

const NEW_CHAT = '/chat?agent=streamer-id'

// A 1×1 PNG.
const PNG =
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='
const MiB = 1024 * 1024

const reply = (runId: string, text: string) =>
  sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    { type: 'TEXT_MESSAGE_START', messageId: `a-${runId}`, role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: `a-${runId}`, delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: `a-${runId}` },
    { type: 'RUN_FINISHED', threadId: 't', runId },
  ])

// pausedOn is a run that replies, then pauses on an open question.
const pausedOn = (runId: string, text: string, question: string) =>
  sse([
    { type: 'RUN_STARTED', threadId: 't', runId },
    { type: 'TEXT_MESSAGE_START', messageId: `a-${runId}`, role: 'assistant' },
    { type: 'TEXT_MESSAGE_CONTENT', messageId: `a-${runId}`, delta: text },
    { type: 'TEXT_MESSAGE_END', messageId: `a-${runId}` },
    {
      type: 'RUN_FINISHED',
      threadId: 't',
      runId,
      outcome: {
        type: 'interrupt',
        interrupts: [{ id: 'int-1', reason: 'human_input', message: question }],
      },
    },
  ])

// An image as a run's request carries it.
const imagePart = (mimeType = 'image/png') => ({
  type: 'image',
  source: { type: 'data', value: PNG, mimeType },
})
const textPart = (text: string) => ({ type: 'text', text })
// A user message with images has content parts; one of text alone is text.
const userMessage = (...content: unknown[]) => ({
  id: expect.any(String),
  role: 'user',
  content,
})
const textMessage = (text: string) => ({
  id: expect.any(String),
  role: 'user',
  content: text,
})

// A file to attach: the PNG, or `size` zero bytes.
interface FileSpec {
  name: string
  mimeType?: string
  size?: number
}
const pngs = (...names: string[]): FileSpec[] => names.map((name) => ({ name }))

async function dataTransfer(page: Page, files: FileSpec[]) {
  return page.evaluateHandle(
    ({ files, png }) => {
      const dt = new DataTransfer()
      for (const f of files) {
        const bytes =
          f.size === undefined
            ? Uint8Array.from(atob(png), (c) => c.charCodeAt(0))
            : new Uint8Array(f.size)
        dt.items.add(
          new File([bytes], f.name, { type: f.mimeType ?? 'image/png' })
        )
      }
      return dt
    },
    { files, png: PNG }
  )
}

// pick attaches files through the composer's file picker.
async function pick(page: Page, files: FileSpec[]) {
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Attach images' }).click()
  await (
    await chooser
  ).setFiles(
    files.map((f) => ({
      name: f.name,
      mimeType: f.mimeType ?? 'image/png',
      buffer: Buffer.from(PNG, 'base64'),
    }))
  )
}

// paste pastes files into target, as from the clipboard.
async function paste(target: Locator, files: FileSpec[]) {
  const dt = await dataTransfer(target.page(), files)
  await target.evaluate((el, clipboardData) => {
    el.dispatchEvent(
      new ClipboardEvent('paste', {
        clipboardData,
        bubbles: true,
        cancelable: true,
      })
    )
  }, dt)
}

// drop drags files onto target and drops them.
async function drop(target: Locator, files: FileSpec[]) {
  const dt = await dataTransfer(target.page(), files)
  for (const type of ['dragenter', 'dragover', 'drop']) {
    await target.dispatchEvent(type, { dataTransfer: dt })
  }
}

const composer = (page: Page) =>
  page.getByRole('textbox', { name: /^Message/ })
const removeButton = (page: Page, name: string) =>
  page.getByRole('button', { name: `Remove ${name}` })
const sentMessages = (page: Page) =>
  page.locator('[data-message-role="user"]')
const toast = (page: Page, text: string) =>
  page.locator('[data-sonner-toast]').filter({ hasText: text })

// startThread opens a thread with a first message of text alone.
async function startThread(page: Page) {
  await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
  await send(page, 'hi')
  await expect(page.getByText('Hello.')).toBeVisible()
}

test.describe('Chat images', () => {
  test('images picked, pasted and dropped in a thread go out as image parts of its message', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('r1', 'Hello.'), reply('r2', 'Three cats.')],
    })
    await startThread(page)

    await pick(page, [{ name: 'picked.png' }])
    await paste(composer(page), [{ name: 'pasted.webp', mimeType: 'image/webp' }])
    // Dropped anywhere on the conversation, not only on the composer.
    await drop(page.getByText('Hello.'), [
      { name: 'dropped.gif', mimeType: 'image/gif' },
    ])
    for (const name of ['picked.png', 'pasted.webp', 'dropped.gif']) {
      await expect(removeButton(page, name)).toBeVisible()
      await expect(page.getByRole('img', { name })).toBeVisible()
    }

    await send(page, 'What are these?')
    await expect(page.getByText('Three cats.')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
    expect(fixture.requests[1].messages).toEqual([
      userMessage(
        textPart('What are these?'),
        imagePart(),
        imagePart('image/webp'),
        imagePart('image/gif')
      ),
    ])

    // The composer is empty again; the message shows its images.
    await expect(removeButton(page, 'picked.png')).toHaveCount(0)
    const sent = sentMessages(page).last()
    await expect(sent.getByText('What are these?')).toBeVisible()
    for (const name of ['picked.png', 'pasted.webp', 'dropped.gif']) {
      await expect(sent.getByRole('img', { name })).toHaveAttribute(
        'src',
        /^data:image\/(png|webp|gif);base64,/
      )
    }
  })

  test('a removed image is not sent, and images go out without text', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('r1', 'Hello.'), reply('r2', 'A cat.')],
    })
    await startThread(page)

    await pick(page, pngs('keep.png', 'drop.png'))
    await removeButton(page, 'drop.png').click()
    await expect(page.getByRole('img', { name: 'drop.png' })).toHaveCount(0)
    await composer(page).press('Enter')

    await expect(page.getByText('A cat.')).toBeVisible()
    expect(fixture.requests[1].messages).toEqual([userMessage(imagePart())])
    // A message of images alone has no text bubble.
    const sent = sentMessages(page).last()
    await expect(sent.getByRole('img', { name: 'keep.png' })).toBeVisible()
    await expect(sent.locator('.bg-secondary')).toHaveCount(0)
  })

  test('a new chat starts with the images attached in the draft', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('r1', 'Two dogs.')],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })

    await pick(page, pngs('picked.png', 'unwanted.png'))
    await paste(composer(page), [{ name: 'pasted.jpg', mimeType: 'image/jpeg' }])
    await drop(page.getByRole('heading', { name: 'Streamer', level: 2 }), [
      { name: 'dropped.gif', mimeType: 'image/gif' },
    ])
    await removeButton(page, 'unwanted.png').click()
    await send(page, 'Who are they?')

    await expect(page.getByText('Two dogs.')).toBeVisible()
    expect(threadInURL(page)).toBe(fixture.requests[0].threadId)
    expect(fixture.requests).toHaveLength(1)
    expect(fixture.requests[0].messages).toEqual([
      userMessage(
        textPart('Who are they?'),
        imagePart(),
        imagePart('image/jpeg'),
        imagePart('image/gif')
      ),
    ])
    await expect(
      sentMessages(page).getByRole('img', { name: 'picked.png' })
    ).toBeVisible()
  })

  test('a file over a limit, or of a type the server does not take, is refused with why and not sent', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('r1', 'Hello.'), reply('r2', 'Seven cats.')],
    })
    await startThread(page)
    const target = composer(page)

    await drop(target, [{ name: 'scan.bmp', mimeType: 'image/bmp' }])
    await expect(
      toast(page, 'Unsupported type image/bmp; accepted: JPEG, PNG, GIF, WebP')
    ).toBeVisible()
    await paste(target, [{ name: 'huge.png', size: 10 * MiB + 1 }])
    await expect(
      toast(page, 'huge.png: exceeds the 10 MiB per-image limit')
    ).toBeVisible()

    // Files dropped at once count against each other: 10 at most.
    await drop(
      target,
      pngs(...Array.from({ length: 11 }, (_, i) => `photo-${i + 1}.png`))
    )
    await expect(
      toast(page, 'photo-11.png: at most 10 images per message')
    ).toBeVisible()
    await expect(page.getByRole('button', { name: /^Remove / })).toHaveCount(10)
    // Full: one more is refused as well.
    await pick(page, pngs('one-more.png'))
    await expect(
      toast(page, 'one-more.png: at most 10 images per message')
    ).toBeVisible()

    // 20 MiB in all, the images pasted at once included.
    for (const name of ['photo-10.png', 'photo-9.png', 'photo-8.png']) {
      await removeButton(page, name).click()
    }
    await paste(target, [
      { name: 'a.png', size: 8 * MiB },
      { name: 'b.png', size: 8 * MiB },
      { name: 'c.png', size: 8 * MiB },
    ])
    await expect(
      toast(page, 'c.png: total attachments exceed the 20 MiB limit')
    ).toBeVisible()
    await expect(removeButton(page, 'b.png')).toBeVisible()
    await removeButton(page, 'a.png').click()
    await removeButton(page, 'b.png').click()

    await expect(page.getByRole('button', { name: /^Remove / })).toHaveCount(7)
    for (const name of ['scan.bmp', 'huge.png', 'photo-11.png', 'c.png']) {
      await expect(removeButton(page, name)).toHaveCount(0)
    }
    await send(page, 'Count them')
    await expect(page.getByText('Seven cats.')).toBeVisible()
    const sent = fixture.requests[1].messages as Array<{ content: unknown[] }>
    expect(sent).toEqual([
      userMessage(textPart('Count them'), ...Array(7).fill(imagePart())),
    ])
  })

  test('the draft refuses a file over a limit with why', async ({ page }) => {
    const fixture = await setupAGUI(page, { runs: [] })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })

    await paste(composer(page), [{ name: 'huge.png', size: 10 * MiB + 1 }])
    await expect(
      toast(page, 'huge.png: exceeds the 10 MiB per-image limit')
    ).toBeVisible()
    await expect(removeButton(page, 'huge.png')).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Send message' })).toBeDisabled()
    expect(fixture.requests).toEqual([])
  })

  test('a later message does not send the earlier images again', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        reply('r1', 'A cat.'),
        reply('r2', 'Orange.'),
        reply('r3', 'Yes.'),
      ],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await pick(page, pngs('cat.png'))
    await send(page, 'What is this?')
    await expect(page.getByText('A cat.')).toBeVisible()
    await send(page, 'What colour is it?')
    await expect(page.getByText('Orange.')).toBeVisible()
    await pick(page, pngs('dog.png'))
    await send(page, 'Is this one a dog?')
    await expect(page.getByText('Yes.')).toBeVisible()

    expect(fixture.requests.map((r) => r.messages)).toEqual([
      [userMessage(textPart('What is this?'), imagePart())],
      [textMessage('What colour is it?')],
      [userMessage(textPart('Is this one a dog?'), imagePart())],
    ])
  })

  test('an image the user sent shows again after a reload', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [reply('r1', 'A cat.')],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await pick(page, pngs('cat.png'))
    await send(page, 'What is this?')
    await expect(page.getByText('A cat.')).toBeVisible()

    // The thread's history, as the server rebuilds it from the session.
    const threadId = threadInURL(page) ?? ''
    fixture.historyByThread = {
      [threadId]: {
        body: {
          threadId,
          messages: [
            {
              id: 'u1',
              role: 'user',
              content: [textPart('What is this?'), imagePart()],
            },
            { id: 'a1', role: 'assistant', content: 'A cat.' },
          ],
          interrupts: [],
          surfaces: [],
        },
      },
    }
    fixture.snapshots.push(emptySnapshot(threadId))
    await page.reload({ waitUntil: 'networkidle' })

    await expect(page.getByText('A cat.')).toBeVisible()
    const restored = sentMessages(page).last()
    await expect(restored.getByText('What is this?')).toBeVisible()
    const image = restored.getByRole('img', { name: 'Image 1' })
    await expect(image).toHaveAttribute('src', `data:image/png;base64,${PNG}`)
    // The image loaded.
    expect(
      await image.evaluate((img: HTMLImageElement) => img.naturalWidth)
    ).toBe(1)
  })

  test('an answer to an open question takes the images attached to it', async ({
    page,
  }) => {
    const fixture = await setupAGUI(page, {
      runs: [
        pausedOn('r1', 'One thing first.', 'Show me the receipt.'),
        reply('r2', 'Refund approved.'),
      ],
    })
    await page.goto(NEW_CHAT, { waitUntil: 'networkidle' })
    await send(page, 'refund my order')
    await expect(page.getByText('Show me the receipt.')).toBeVisible()

    await pick(page, pngs('receipt.png'))
    // An answer needs words; the images go with them.
    await composer(page).press('Enter')
    await expect(
      toast(page, 'Type your answer; the images are sent with it.')
    ).toBeVisible()
    expect(fixture.requests).toHaveLength(1)

    await send(page, 'Here it is')
    await expect(page.getByText('Refund approved.')).toBeVisible()
    expect(fixture.requests).toHaveLength(2)
    // No resume: the server answers its oldest open question (ADR-0002).
    expect(fixture.requests[1]).not.toHaveProperty('resume')
    expect(fixture.requests[1].messages).toEqual([
      userMessage(textPart('Here it is'), imagePart()),
    ])
    await expect(removeButton(page, 'receipt.png')).toHaveCount(0)
    await expect(
      sentMessages(page).last().getByRole('img', { name: 'receipt.png' })
    ).toBeVisible()
  })
})
