import type { Message } from '@ag-ui/client'
import { describe, expect, it } from 'vitest'
import type { ThreadHistory } from './history'
import {
  NO_INPUT,
  imageFiles,
  lastRunOfHistory,
  lastRunOfRunError,
  lastRunOfStop,
  recordedText,
  restoreInput,
  runInput,
  runningInput,
  type Composer,
  type TurnInput,
} from './last-run'

// A 1×1 PNG, and a second image's bytes.
const PNG =
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='
const GIF = btoa('GIF89a')

const text = (t: string) => ({ type: 'text', text: t })
const image = (data = PNG, mimeType = 'image/png') => ({
  type: 'image',
  source: { type: 'data', value: data, mimeType },
})

// user is a user message as a run's request carries it.
function user(content: unknown, id = 'u'): Message {
  return { id, role: 'user', content } as Message
}

function history(partial: Partial<ThreadHistory>): ThreadHistory {
  return {
    threadId: 't',
    messages: [],
    interrupts: [],
    surfaces: [],
    ...partial,
  }
}

describe('runInput', () => {
  it('is the text a run sends', () => {
    expect(
      runInput({
        messages: [
          user('earlier', 'u1'),
          { id: 'a1', role: 'assistant', content: 'Hello.' },
          user('plan the trip', 'u2'),
        ],
      })
    ).toEqual({ text: 'plan the trip', images: [] })
  })

  it('is the text and the images of a turn with images, in order', () => {
    expect(
      runInput({
        messages: [
          user([text('What are these?'), image(), image(GIF, 'image/gif')]),
        ],
      })
    ).toEqual({
      text: 'What are these?',
      images: [
        { mimeType: 'image/png', data: PNG },
        { mimeType: 'image/gif', data: GIF },
      ],
    })
  })

  it('keeps a turn of images alone', () => {
    expect(runInput({ messages: [user([image()])] })).toEqual({
      text: '',
      images: [{ mimeType: 'image/png', data: PNG }],
    })
  })

  it('is no input for a run that answers a question', () => {
    // The answer goes as a resume entry; the readable reply stays here.
    expect(
      runInput({
        messages: [user('deploy'), user('2.4.1', 'reply')],
        resume: [
          { status: 'resolved' },
          // A placeholder for a question left open.
          { status: 'cancelled' },
        ],
      })
    ).toEqual(NO_INPUT)
  })

  it('is the message sent while questions are open, which answers the oldest', () => {
    // Only placeholders, which never leave the browser.
    expect(
      runInput({
        messages: [user('deploy'), user('eu-west', 'reply')],
        resume: [{ status: 'cancelled' }, { status: 'cancelled' }],
      })
    ).toEqual({ text: 'eu-west', images: [] })
  })

  it('is no input for a run that continues after tool results', () => {
    expect(
      runInput({
        messages: [
          user('locate me'),
          {
            id: 'a1',
            role: 'assistant',
            toolCalls: [
              {
                id: 'c1',
                type: 'function',
                function: { name: 'geolocate', arguments: '{}' },
              },
            ],
          },
          { id: 'r1', role: 'tool', toolCallId: 'c1', content: '{}' },
        ],
      })
    ).toEqual(NO_INPUT)
  })

  it('is no input without a user message', () => {
    expect(runInput({ messages: [] })).toEqual(NO_INPUT)
  })
})

describe('runningInput', () => {
  const RUN = { runId: 'run-2', invocationId: 'inv-2' }
  const before = [
    { id: 'u1', role: 'user', content: 'Book a flight' },
    { id: 'a1', role: 'assistant', content: 'Booked.' },
  ]

  it('is the turn a read during the run ends with: the one the run started from', () => {
    expect(
      runningInput(
        history({
          messages: [
            ...before,
            { id: 'u2', role: 'user', content: [text('plan it'), image()] },
          ],
          running: RUN,
        })
      )
    ).toEqual({
      text: 'plan it',
      images: [{ mimeType: 'image/png', data: PNG }],
    })
  })

  it('is no input for a run that continues after tool results', () => {
    expect(
      runningInput(
        history({
          messages: [
            ...before,
            {
              id: 'result:c1',
              role: 'tool',
              toolCallId: 'c1',
              content: '{"ok":true}',
            },
          ],
          running: RUN,
        })
      )
    ).toEqual(NO_INPUT)
  })

  it('is no input when the read names no run: its last turn is an earlier run’s', () => {
    expect(
      runningInput(
        history({
          messages: [...before, { id: 'u2', role: 'user', content: 'again' }],
        })
      )
    ).toEqual(NO_INPUT)
  })
})

const input: TurnInput = { text: 'plan the trip', images: [] }

describe('lastRunOfRunError', () => {
  it('is a failure with the error the run reported', () => {
    expect(lastRunOfRunError({ message: 'model exploded' }, input)).toEqual({
      status: 'failed',
      error: 'model exploded',
      input,
    })
    expect(lastRunOfRunError({}, input)).toEqual({
      status: 'failed',
      error: '',
      input,
    })
  })

  it('is a stopped run when the RUN_ERROR carries the stop code', () => {
    expect(
      lastRunOfRunError({ message: 'stopped by user', code: 'stopped' }, input)
    ).toEqual({ status: 'stopped', error: '', input })
  })

  it('is nothing for the abort that only detaches the page', () => {
    expect(
      lastRunOfRunError({ message: 'Request aborted', code: 'abort' }, input)
    ).toBeNull()
  })
})

describe('lastRunOfStop', () => {
  it('is a stopped run with the input it sent', () => {
    expect(lastRunOfStop(input)).toEqual({
      status: 'stopped',
      error: '',
      input,
    })
  })
})

describe('lastRunOfHistory', () => {
  const failed = (inputText: string) => ({
    status: 'failed',
    error: 'model exploded',
    input: inputText,
  })

  it('is nothing when the last run succeeded or one is going on', () => {
    expect(lastRunOfHistory(history({}))).toBeNull()
  })

  it('takes the turn from the history when its last user turn is the run’s', () => {
    const got = lastRunOfHistory(
      history({
        messages: [
          { id: 'u1', role: 'user', content: 'hi' },
          { id: 'a1', role: 'assistant', content: 'Hello.' },
          {
            id: 'u2',
            role: 'user',
            content: [
              text('What are these?'),
              image(),
              image(GIF, 'image/gif'),
            ],
          },
          // What the run stored before it failed.
          { id: 'a2', role: 'assistant', content: 'Looking' },
        ],
        lastRun: failed('What are these?'),
      })
    )
    expect(got).toEqual({
      status: 'failed',
      error: 'model exploded',
      input: {
        text: 'What are these?',
        images: [
          { mimeType: 'image/png', data: PNG },
          { mimeType: 'image/gif', data: GIF },
        ],
      },
    })
  })

  it('reports a stopped run as stopped', () => {
    expect(
      lastRunOfHistory(
        history({
          messages: [{ id: 'u1', role: 'user', content: 'plan the trip' }],
          lastRun: {
            status: 'cancelled',
            error: 'stopped by user',
            input: 'plan the trip',
          },
        })
      )
    ).toEqual({
      status: 'stopped',
      error: '',
      input: { text: 'plan the trip', images: [] },
    })
  })

  it('restores the record’s text alone when the run never stored its turn', () => {
    // The last user turn is an earlier one, with an image of its own.
    expect(
      lastRunOfHistory(
        history({
          messages: [
            { id: 'u1', role: 'user', content: [text('a cat?'), image()] },
            { id: 'a1', role: 'assistant', content: 'A cat.' },
          ],
          lastRun: failed('plan the trip'),
        })
      )
    ).toEqual({
      status: 'failed',
      error: 'model exploded',
      input: { text: 'plan the trip', images: [] },
    })
  })

  it('matches a long turn the record cut, and restores all of it', () => {
    const long = 'x'.repeat(5000)
    const got = lastRunOfHistory(
      history({
        messages: [{ id: 'u1', role: 'user', content: [text(long), image()] }],
        lastRun: failed(`${'x'.repeat(4096)}…`),
      })
    )
    expect(got?.input).toEqual({
      text: long,
      images: [{ mimeType: 'image/png', data: PNG }],
    })
  })

  it('matches a turn of images alone, whose record has no text', () => {
    expect(
      lastRunOfHistory(
        history({
          messages: [{ id: 'u1', role: 'user', content: [image()] }],
          lastRun: failed(''),
        })
      )?.input
    ).toEqual({ text: '', images: [{ mimeType: 'image/png', data: PNG }] })
  })

  it('leaves nothing to restore after a failed answer to a question', () => {
    // An answer is a user turn in the history, but its run records no input.
    expect(
      lastRunOfHistory(
        history({
          messages: [
            { id: 'u1', role: 'user', content: 'deploy' },
            { id: 'a1', role: 'assistant', content: 'Which version?' },
            { id: 'u2', role: 'user', content: '2.4.1' },
          ],
          lastRun: failed(''),
        })
      )?.input
    ).toEqual(NO_INPUT)
  })

  it('ignores a status it does not know', () => {
    expect(
      lastRunOfHistory(
        history({ lastRun: { status: 'succeeded', input: 'hi' } })
      )
    ).toBeNull()
  })
})

describe('recordedText', () => {
  it('joins text parts with spaces, as the record does', () => {
    expect(recordedText('plan the trip')).toBe('plan the trip')
    expect(recordedText([text('one'), image(), text('two')])).toBe('one two')
    expect(recordedText([image()])).toBe('')
    expect(recordedText(undefined)).toBe('')
  })

  it('cuts a text past 4096 bytes on a character boundary', () => {
    const exact = 'x'.repeat(4096)
    expect(recordedText(exact)).toBe(exact)
    expect(recordedText(`${exact}y`)).toBe(`${exact}…`)
    // "é" is two bytes: the 4096th byte falls inside the 2048th one.
    const accents = 'é'.repeat(2049)
    expect(recordedText(`x${accents}`)).toBe(`x${'é'.repeat(2047)}…`)
  })
})

// composer is a composer that records what Restore input does to it.
function composer(refuse: string[] = []) {
  const log: string[] = []
  const files: File[] = []
  const fake: Composer = {
    setText: (t) => log.push(`text:${t}`),
    clearAttachments: async () => {
      log.push('clear')
    },
    addAttachment: async (file) => {
      log.push(`add:${file.name}`)
      if (refuse.includes(file.name)) throw new Error(`${file.name}: refused`)
      files.push(file)
    },
  }
  return { fake, log, files }
}

describe('restoreInput', () => {
  it('puts the text and the images back in the composer, in order', async () => {
    const { fake, log, files } = composer()
    await restoreInput(fake, {
      text: 'What are these?',
      images: [
        { mimeType: 'image/png', data: PNG },
        { mimeType: 'image/gif', data: GIF },
      ],
    })
    expect(log).toEqual([
      'text:What are these?',
      'clear',
      'add:image-1.png',
      'add:image-2.gif',
    ])
    expect(files.map((f) => f.type)).toEqual(['image/png', 'image/gif'])
    expect(new TextDecoder().decode(await files[1].arrayBuffer())).toBe(
      'GIF89a'
    )
  })

  it('adds the other images when the composer refuses one', async () => {
    const { fake, log, files } = composer(['image-1.png'])
    await restoreInput(fake, {
      text: '',
      images: [
        { mimeType: 'image/png', data: PNG },
        { mimeType: 'image/gif', data: GIF },
      ],
    })
    expect(log).toEqual([
      'text:',
      'clear',
      'add:image-1.png',
      'add:image-2.gif',
    ])
    expect(files.map((f) => f.name)).toEqual(['image-2.gif'])
  })
})

describe('imageFiles', () => {
  it('leaves out an image whose data does not decode', () => {
    const files = imageFiles([
      { mimeType: 'image/png', data: '%%%' },
      { mimeType: 'image/webp', data: PNG },
    ])
    expect(files.map((f) => f.name)).toEqual(['image-2.webp'])
  })
})
