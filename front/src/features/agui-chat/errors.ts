// errorReporter shows each failure once. The AG-UI runtime reports a failed
// run through onError (an HTTP failure even twice) and then rejects the call
// that started the run with that same error; a call that fails before any run
// starts is never reported. Giving the reporter both onError and every send's
// rejection shows each failure exactly once.
export function errorReporter(
  show: (message: string) => void
): (err: unknown) => void {
  const shown = new WeakSet<object>()
  return (err) => {
    if (typeof err === 'object' && err !== null) {
      if (shown.has(err)) return
      shown.add(err)
    }
    show(errorMessage(err))
  }
}

function errorMessage(err: unknown): string {
  if (err instanceof Error && err.message) return err.message
  if (typeof err === 'string' && err) return err
  return 'AG-UI request failed'
}
