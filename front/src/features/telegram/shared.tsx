import {
  TelegramCredentialState,
  type TelegramDestination,
} from '@/gen/agents/v1/telegram_pb'
import { Badge } from '@/components/ui/badge'
import { CREDENTIAL_STATE_LABELS } from './labels'

/**
 * Renders the exact Telegram address. An absent thread ID is shown as
 * "no topic" rather than omitted, because it is a distinct address from any
 * Forum Topic in the same chat, not a wildcard over the group.
 */
export function AddressLabel({ destination }: { destination: TelegramDestination }) {
  return (
    <span className='font-mono text-xs'>
      chat {destination.chatId}
      {destination.messageThreadId ? (
        <> · topic {destination.messageThreadId}</>
      ) : (
        <span className='text-muted-foreground'> · no topic</span>
      )}
    </span>
  )
}

export function CredentialStateBadge({ state }: { state: TelegramCredentialState }) {
  const label = CREDENTIAL_STATE_LABELS[state] ?? 'Unknown'
  const variant =
    state === TelegramCredentialState.VALID
      ? 'bg-success-muted text-success-foreground'
      : state === TelegramCredentialState.INVALID
        ? 'bg-destructive/10 text-destructive'
        : 'bg-muted text-muted-foreground'
  return <Badge className={variant}>{label}</Badge>
}
