import type { Message } from '@ag-ui/client'
import { runMessages } from './a2ui/agent'
import { STOPPED_CODE, type RunErrorEvent } from './errors'
import type {
  HistoryContentPart,
  HistoryMessage,
  ThreadHistory,
} from './history'

// A run that failed or was stopped stays visible under the conversation,
// with the turn it started from (ADR-0016 decision 8). The page learns of it
// three ways:
//   - live, from the run's RUN_ERROR: a failure, or a Stop from anywhere,
//     which carries the stop code. That is the stream of a run the page
//     started, or the log of one it found holding the thread (#407);
//   - from this page's Stop, which ends the run here, so its stream may never
//     say so;
//   - after a reload, from the thread history's lastRun (docs/api.md "Thread
//     history").
// A run that starts afterwards replaces it. Restore input puts the turn back
// in the composer, and sending it starts a new run.

// TurnInput is the turn a run started from, as the composer sends one: its
// text and its images.
export interface TurnInput {
  text: string
  images: TurnImage[]
}

// TurnImage is one image of a turn: its bytes in base64, as an image part's
// data source carries them.
export interface TurnImage {
  mimeType: string
  data: string
}

// LastRun is the thread's last run when it failed or was stopped, as its
// notice shows it.
export interface LastRun {
  status: 'failed' | 'stopped'
  // error is why a failed run ended, as the server reported it.
  error: string
  input: TurnInput
}

// NO_INPUT is the input of a run that sent no turn of its own.
export const NO_INPUT: TurnInput = { text: '', images: [] }

// ABORT_CODE is the code of the RUN_ERROR the AG-UI client reports when the
// page aborts a run's request. That only detaches the page from the run,
// which goes on (ADR-0016), so it says nothing about how the run ended.
const ABORT_CODE = 'abort'

// RECORD_LIMIT is how many bytes of a turn's text a run's Invocation record
// keeps, and so lastRun.input.
const RECORD_LIMIT = 4096

// runInput is the turn a run's request sends, which is what the server
// records as the run's input: its trailing user message (runMessages).
//   - A run that answers an Interrupt carries the answer as a resolved
//     resume entry, and one that continues after tool results carries the
//     results. Neither sends a turn of its own, and the server records no
//     input for them.
//   - A message sent while Interrupts are open carries only the client's
//     cancelled placeholders, which never leave the browser
//     (ButterAGUIAgent), so it is a turn like any other.
export function runInput(run: {
  messages: readonly Message[]
  resume?: ReadonlyArray<{ status: string }>
}): TurnInput {
  if (run.resume?.some((entry) => entry.status === 'resolved')) return NO_INPUT
  const sent = runMessages(run.messages)
  const [message] = sent
  if (sent.length !== 1 || message.role !== 'user') return NO_INPUT
  return turnInput(message.content)
}

// runningInput is the turn the run in flight started from, as a read of the
// thread during the run has it (docs/api.md "Reads during a run"): the read
// ends with that turn. A user turn is its text and its images; a run that
// continues after tool results sent no turn of its own. The read cannot tell
// an answer to a question from a message, so a run that answered one has
// that answer as its turn, where its record keeps none.
export function runningInput(history: ThreadHistory): TurnInput {
  const messages = history.running ? (history.messages ?? []) : []
  const last = messages[messages.length - 1]
  return last?.role === 'user' ? turnInput(last.content) : NO_INPUT
}

// lastRunOfRunError is the last run a RUN_ERROR reports, given the input its
// run sent: stopped when it carries the stop code, failed otherwise. The
// client's own abort ends nothing but the page's view of the run, and gives
// none.
export function lastRunOfRunError(
  event: RunErrorEvent,
  input: TurnInput
): LastRun | null {
  if (event.code === ABORT_CODE) return null
  if (event.code === STOPPED_CODE)
    return { status: 'stopped', error: '', input }
  return { status: 'failed', error: event.message ?? '', input }
}

// lastRunOfStop is the last run this page's Stop ended, given the input it
// sent.
export function lastRunOfStop(input: TurnInput): LastRun {
  return { status: 'stopped', error: '', input }
}

// lastRunOfHistory is the last run the thread's history reports, if it
// failed or was stopped. Its record keeps only the text of the turn, cut to
// RECORD_LIMIT bytes. Once the run stored its turn, that is the history's
// last user turn, whose text as a record keeps it (recordedText) is the
// record's; the turn then comes from the history, with its whole text and
// its images. Otherwise the record's text is all there is to restore.
export function lastRunOfHistory(history: ThreadHistory): LastRun | null {
  const last = history.lastRun
  const status =
    last?.status === 'failed'
      ? 'failed'
      : last?.status === 'cancelled'
        ? 'stopped'
        : null
  if (!last || !status) return null
  const recorded = last.input ?? ''
  const turn = lastUserTurn(history.messages ?? [])
  const input =
    turn && recordedText(turn.content) === recorded
      ? turnInput(turn.content)
      : { text: recorded, images: [] }
  return {
    status,
    error: status === 'failed' ? (last.error ?? '') : '',
    input,
  }
}

// recordedText is a turn's text as a run's record keeps it: its text parts
// joined by spaces, and a text over RECORD_LIMIT bytes cut on a character
// boundary, ending with "…".
export function recordedText(content: HistoryMessage['content']): string {
  const text =
    typeof content === 'string' ? content : textParts(content ?? []).join(' ')
  const bytes = new TextEncoder().encode(text)
  if (bytes.length <= RECORD_LIMIT) return text
  let cut = RECORD_LIMIT
  while (cut > 0 && (bytes[cut] & 0xc0) === 0x80) cut--
  return `${new TextDecoder().decode(bytes.subarray(0, cut))}…`
}

// Composer is what Restore input fills: the thread's composer.
export interface Composer {
  setText(text: string): void
  clearAttachments(): Promise<void>
  addAttachment(file: File): Promise<void>
}

// restoreInput puts a turn back in the composer, in place of what it holds:
// its text, and its images as files, in order. The composer takes them as it
// takes picked ones (ImageAttachmentAdapter), limits included. An image it
// refuses is reported as any refused file is (composer.attachmentAddError),
// and the rest are still added.
export async function restoreInput(
  composer: Composer,
  input: TurnInput
): Promise<void> {
  composer.setText(input.text)
  await composer.clearAttachments()
  for (const file of imageFiles(input.images)) {
    try {
      await composer.addAttachment(file)
    } catch {
      // The composer reported it.
    }
  }
}

// imageFiles turns a turn's images back into files, named by their place in
// the turn. An image whose data does not decode is left out.
export function imageFiles(images: readonly TurnImage[]): File[] {
  return images.flatMap((image, i) => {
    let binary: string
    try {
      binary = atob(image.data)
    } catch {
      return []
    }
    const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0))
    const ext = image.mimeType.split('/')[1] || 'bin'
    return [
      new File([bytes], `image-${i + 1}.${ext}`, { type: image.mimeType }),
    ]
  })
}

// turnInput reads a user turn's content, as a run's request and the thread
// history carry it: its text alone, or AG-UI content parts, its text and
// its images inline as data sources. Text parts read as one paragraph each,
// as the history shows them.
function turnInput(content: unknown): TurnInput {
  if (typeof content === 'string') return { text: content, images: [] }
  if (!Array.isArray(content)) return NO_INPUT
  const parts = content as HistoryContentPart[]
  const images: TurnImage[] = []
  for (const { type, source } of parts) {
    if (type !== 'image' || source?.type !== 'data') continue
    if (source.value && source.mimeType) {
      images.push({ mimeType: source.mimeType, data: source.value })
    }
  }
  return { text: textParts(parts).join('\n\n'), images }
}

function textParts(parts: readonly HistoryContentPart[]): string[] {
  return parts.flatMap((part) =>
    part.type === 'text' && part.text ? [part.text] : []
  )
}

// lastUserTurn is the history's last user message: the turn its last run
// started from, if that run stored one.
function lastUserTurn(
  messages: readonly HistoryMessage[]
): HistoryMessage | undefined {
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i].role === 'user') return messages[i]
  }
  return undefined
}
