import { useAui, useAuiState } from '@assistant-ui/react'
import { RunNotice } from '@/components/chat/run-notice'
import { restoreInput, type LastRun } from './last-run'

// LastRunNotice shows the thread's last run under the conversation when it
// failed or was stopped, while no run is going on. Restore input puts the
// turn the run started from back in the thread's composer. A run that sent
// no turn of its own, such as an answer to a question, leaves nothing to put
// back.
export function LastRunNotice({ lastRun }: { lastRun: LastRun | null }) {
  const aui = useAui()
  const running = useAuiState((s) => s.thread.isRunning)
  if (!lastRun || running) return null
  const { input } = lastRun
  const restorable = input.text.trim() !== '' || input.images.length > 0
  return (
    <RunNotice
      notice={{
        status: lastRun.status,
        error: lastRun.error,
        input: input.text,
        images: input.images.length,
      }}
      onRestore={
        restorable ? () => void restoreInput(aui.composer, input) : undefined
      }
    />
  )
}
