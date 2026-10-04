import { describe, expect, it, vi } from 'vitest'
import { stopper, type LocalRun } from './stop'

// localRun is a run as the page follows it. log records the Stop requests
// sent and the local cancels, in order.
function localRun() {
  const log: string[] = []
  let going = true
  const run: LocalRun = {
    canCancel: () => going,
    cancel: () => {
      log.push('cancel')
      going = false
    },
  }
  // end is the run ending on its own here, as RUN_ERROR "stopped" ends it.
  return { run, log, end: () => (going = false) }
}

// answers hands out one pending answer per Stop request, for the test to
// settle.
function answers(log: string[]) {
  const pending: Array<{ resolve: () => void; reject: (e: unknown) => void }> =
    []
  const stop = () => {
    log.push('stop')
    return new Promise<void>((resolve, reject) =>
      pending.push({ resolve, reject })
    )
  }
  return { stop, pending }
}

describe('stopper', () => {
  it('stops the run on the server, then cancels it here', async () => {
    const { run, log } = localRun()
    const { stop, pending } = answers(log)
    const done = stopper(stop, run)()
    await Promise.resolve()
    // Nothing is cancelled here before the server took the Stop.
    expect(log).toEqual(['stop'])
    pending[0].resolve()
    await done
    expect(log).toEqual(['stop', 'cancel'])
  })

  it('cancels here when the server found no run to stop', async () => {
    const { run, log } = localRun()
    await stopper(async () => {
      log.push('stop')
      return null
    }, run)()
    expect(log).toEqual(['stop', 'cancel'])
  })

  it('leaves alone a run that ended before the Stop was answered', async () => {
    const { run, log, end } = localRun()
    const { stop, pending } = answers(log)
    const done = stopper(stop, run)()
    // The run's RUN_ERROR "stopped" arrives ahead of the Stop's answer.
    end()
    pending[0].resolve()
    await done
    expect(log).toEqual(['stop'])
  })

  it('leaves the run going when the server refuses the Stop', async () => {
    const { run, log } = localRun()
    const { stop, pending } = answers(log)
    const refused = new Error('stop unavailable, retry later')
    const done = stopper(stop, run)()
    pending[0].reject(refused)
    await expect(done).rejects.toBe(refused)
    expect(log).toEqual(['stop'])
    expect(run.canCancel()).toBe(true)
  })

  it('sends one request for every click while a Stop is under way', async () => {
    const { run, log } = localRun()
    const { stop, pending } = answers(log)
    const click = stopper(stop, run)
    const first = click()
    const second = click()
    expect(second).toBe(first)
    pending[0].resolve()
    await Promise.all([first, second])
    expect(log).toEqual(['stop', 'cancel'])
  })

  it('waits for the end of a run the page waits out, and cancels nothing here', async () => {
    const { run, log } = localRun()
    const { stop, pending } = answers(log)
    let ended!: () => void
    const waitedOut: LocalRun = {
      ...run,
      untilEnded: () => {
        log.push('wait')
        return new Promise<void>((resolve) => (ended = resolve))
      },
    }
    const click = stopper(stop, waitedOut)
    let settled = false
    const done = click().then(() => (settled = true))
    pending[0].resolve()
    await vi.waitFor(() => expect(log).toEqual(['stop', 'wait']))
    // The Stop lasts until the page shows the run ended; clicks join it.
    expect(click()).toBe(click())
    await Promise.resolve()
    expect(settled).toBe(false)
    ended()
    await done
    expect(log).toEqual(['stop', 'wait'])
    expect(run.canCancel()).toBe(true)
  })

  it('cancels the run here when there is no end to wait for', async () => {
    const { run, log } = localRun()
    const streamed: LocalRun = { ...run, untilEnded: () => undefined }
    await stopper(async () => {
      log.push('stop')
    }, streamed)()
    expect(log).toEqual(['stop', 'cancel'])
  })

  it('leaves a run the page waits out alone when the server refuses the Stop', async () => {
    const { run, log } = localRun()
    const { stop, pending } = answers(log)
    const waitedOut: LocalRun = {
      ...run,
      untilEnded: () => {
        log.push('wait')
        return Promise.resolve()
      },
    }
    const refused = new Error('stop unavailable, retry later')
    const done = stopper(stop, waitedOut)()
    pending[0].reject(refused)
    await expect(done).rejects.toBe(refused)
    expect(log).toEqual(['stop'])
  })

  it('asks again on the next click once a Stop was refused', async () => {
    const { run, log } = localRun()
    const { stop, pending } = answers(log)
    const click = stopper(stop, run)
    const refused = click()
    pending[0].reject(new Error('stop unavailable, retry later'))
    await expect(refused).rejects.toThrow('stop unavailable')
    const retried = click()
    expect(retried).not.toBe(refused)
    pending[1].resolve()
    await retried
    expect(log).toEqual(['stop', 'stop', 'cancel'])
  })
})
