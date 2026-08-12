import {
  TelegramCredentialState,
  TelegramReceiveMode,
  TelegramWebhookState,
  TelegramReplyMode,
  TelegramSessionPolicy,
  TelegramTriggerMode,
} from '@/gen/agents/v1/telegram_pb'

export const RECEIVE_MODE_LABELS: Record<number, string> = {
  [TelegramReceiveMode.WEBHOOK]: 'Webhook',
  [TelegramReceiveMode.LONG_POLLING]: 'Long polling',
}

export const CREDENTIAL_STATE_LABELS: Record<number, string> = {
  [TelegramCredentialState.MISSING]: 'No token',
  [TelegramCredentialState.VALID]: 'Token valid',
  [TelegramCredentialState.INVALID]: 'Token rejected',
}

export const WEBHOOK_STATE_LABELS: Record<number, string> = {
  [TelegramWebhookState.NOT_APPLICABLE]: 'No webhook',
  [TelegramWebhookState.REGISTERED]: 'Webhook registered',
  [TelegramWebhookState.PENDING]: 'Webhook pending',
  [TelegramWebhookState.FAILED]: 'Webhook failed',
}

export const TRIGGER_MODE_OPTIONS: { value: TelegramTriggerMode; label: string }[] = [
  { value: TelegramTriggerMode.ALL, label: 'All messages' },
  { value: TelegramTriggerMode.MENTION, label: 'Mention only' },
  { value: TelegramTriggerMode.COMMAND, label: 'Command only' },
  { value: TelegramTriggerMode.MENTION_OR_COMMAND, label: 'Mention or command' },
]

export const SESSION_POLICY_OPTIONS: { value: TelegramSessionPolicy; label: string }[] = [
  { value: TelegramSessionPolicy.DESTINATION, label: 'Shared across the destination' },
  { value: TelegramSessionPolicy.USER, label: 'Separate per user' },
]

export const REPLY_MODE_OPTIONS: { value: TelegramReplyMode; label: string }[] = [
  { value: TelegramReplyMode.REPLY, label: 'Quote the incoming message' },
  { value: TelegramReplyMode.NEW_MESSAGE, label: 'Send a new message' },
]

/** Splits a comma/newline/space separated ID list typed by an operator. */
export function parseIdList(raw: string): string[] {
  return raw
    .split(/[\s,]+/)
    .map((entry) => entry.trim())
    .filter(Boolean)
}

export function formatIdList(ids: string[] | undefined): string {
  return (ids ?? []).join(', ')
}
