import { describe, expect, it } from 'vitest'
import { errorReporter } from './errors'

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
