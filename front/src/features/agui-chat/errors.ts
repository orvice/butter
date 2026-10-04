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

// The server refuses a run before its stream opens, with 403, when the
// thread cannot be used from here (docs/api.md, AG-UI Sessions): another
// user holds the threadId, it is the caller's own thread from another
// workspace, or the thread is bound to another agent.
const THREAD_REFUSALS = [
  'threadId is not available',
  'threadId belongs to another agent',
]

// threadRefusal returns the server's message when err is such a refusal,
// and null for any other failure. The AG-UI client rejects a non-2xx run
// with an Error carrying the HTTP status and the parsed JSON body.
export function threadRefusal(err: unknown): string | null {
  if (typeof err !== 'object' || err === null) return null
  const { status, payload } = err as { status?: unknown; payload?: unknown }
  if (status !== 403 || typeof payload !== 'object' || payload === null) {
    return null
  }
  const message = (payload as { error?: unknown }).error
  if (typeof message !== 'string') return null
  return THREAD_REFUSALS.some((prefix) => message.startsWith(prefix))
    ? message
    : null
}
