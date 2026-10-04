import { describe, expect, it } from 'vitest'
import { errorReporter, threadRefusal } from './errors'

// httpError is how the AG-UI client rejects a run the server answered with
// a non-2xx status: the status and the parsed JSON body ride on the Error.
function httpError(status: number, payload: unknown) {
  return Object.assign(
    new Error(`HTTP ${status}: ${JSON.stringify(payload)}`),
    { status, payload }
  )
}

describe('threadRefusal', () => {
  it('names a thread the server will not run from here', () => {
    const unavailable = 'threadId is not available; start a new thread'
    const otherAgent = 'threadId belongs to another agent; start a new thread'
    expect(threadRefusal(httpError(403, { error: unavailable }))).toBe(
      unavailable
    )
    expect(threadRefusal(httpError(403, { error: otherAgent }))).toBe(
      otherAgent
    )
  })

  it('leaves every other failure alone', () => {
    // Another refusal on 403, and the same words on another status.
    expect(
      threadRefusal(httpError(403, { error: 'agent not found: x' }))
    ).toBeNull()
    expect(
      threadRefusal(
        httpError(409, {
          error: 'threadId is not available; start a new thread',
        })
      )
    ).toBeNull()
    expect(
      threadRefusal(httpError(403, 'threadId is not available'))
    ).toBeNull()
    expect(threadRefusal(new Error('threadId is not available'))).toBeNull()
    expect(threadRefusal(undefined)).toBeNull()
  })
})

function reporter() {
  const shown: string[] = []
  return { report: errorReporter((m) => shown.push(m)), shown }
}

describe('errorReporter', () => {
  it('shows an error reported more than once only once', () => {
    // The runtime reports a failed run, then rejects the send with it.
    const { report, shown } = reporter()
    const err = new Error('HTTP 409: busy')
    report(err)
    report(err)
    report(err)
    expect(shown).toEqual(['HTTP 409: busy'])
  })

  it('shows every distinct failure, even with the same message', () => {
    const { report, shown } = reporter()
    report(new Error('HTTP 409: busy'))
    report(new Error('HTTP 409: busy'))
    expect(shown).toEqual(['HTTP 409: busy', 'HTTP 409: busy'])
  })

  it('names a failure that carries no message', () => {
    const { report, shown } = reporter()
    report(new Error(''))
    report('the stream broke')
    report(undefined)
    expect(shown).toEqual([
      'AG-UI request failed',
      'the stream broke',
      'AG-UI request failed',
    ])
  })
})
