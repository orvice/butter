import '../../gen/agents/v1/telegram.pb.dart';

String receiveModeLabel(TelegramReceiveMode mode) => switch (mode) {
  TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK => 'Webhook',
  TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_LONG_POLLING => 'Long polling',
  _ => 'Unset',
};

String credentialStateLabel(TelegramCredentialState state) => switch (state) {
  TelegramCredentialState.TELEGRAM_CREDENTIAL_STATE_VALID => 'Token valid',
  TelegramCredentialState.TELEGRAM_CREDENTIAL_STATE_INVALID => 'Token invalid',
  TelegramCredentialState.TELEGRAM_CREDENTIAL_STATE_MISSING => 'No token',
  _ => 'Unknown',
};

String webhookStateLabel(TelegramWebhookState state) => switch (state) {
  TelegramWebhookState.TELEGRAM_WEBHOOK_STATE_REGISTERED =>
    'Webhook registered',
  TelegramWebhookState.TELEGRAM_WEBHOOK_STATE_PENDING => 'Webhook pending',
  TelegramWebhookState.TELEGRAM_WEBHOOK_STATE_FAILED => 'Webhook failed',
  TelegramWebhookState.TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE =>
    'Not webhook mode',
  _ => 'Unknown',
};

String processingStatusLabel(
  TelegramProcessingStatus status,
) => switch (status) {
  TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_SUCCEEDED => 'Succeeded',
  TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_FAILED => 'Failed',
  TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN =>
    'Failed (uncertain)',
  TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER =>
    'Ready to deliver',
  TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_PROCESSING =>
    'Processing',
  TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_RECEIVED => 'Received',
  _ => 'Unknown',
};

String triggerModeLabel(TelegramTriggerMode mode) => switch (mode) {
  TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_ALL => 'All messages',
  TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_MENTION => 'Bot mention',
  TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_COMMAND => 'Commands',
  TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND =>
    'Mention or command',
  _ => 'All messages',
};

String sessionPolicyLabel(TelegramSessionPolicy policy) => switch (policy) {
  TelegramSessionPolicy.TELEGRAM_SESSION_POLICY_USER => 'Per user',
  TelegramSessionPolicy.TELEGRAM_SESSION_POLICY_DESTINATION =>
    'Shared at destination',
  _ => 'Shared at destination',
};

String replyModeLabel(TelegramReplyMode mode) => switch (mode) {
  TelegramReplyMode.TELEGRAM_REPLY_MODE_NEW_MESSAGE => 'New message',
  TelegramReplyMode.TELEGRAM_REPLY_MODE_REPLY => 'Quote inbound',
  _ => 'Quote inbound',
};

String formatAddress(TelegramDestination d) {
  final thread = d.messageThreadId.trim();
  return thread.isEmpty
      ? 'chat ${d.chatId} · no topic'
      : 'chat ${d.chatId} · topic $thread';
}
