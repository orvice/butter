import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/telegram.pb.dart';

/// Validates a Telegram decimal ID string (chat, thread, user lists).
bool isTelegramDecimalId(String value, {bool allowNegative = true}) {
  final trimmed = value.trim();
  if (trimmed.isEmpty) return false;
  if (allowNegative && trimmed.startsWith('-')) {
    return RegExp(r'^-\d+$').hasMatch(trimmed);
  }
  return RegExp(r'^\d+$').hasMatch(trimmed);
}

List<String> parseTelegramIdList(String text) {
  if (text.trim().isEmpty) return [];
  return text
      .split(RegExp(r'[\s,]+'))
      .map((s) => s.trim())
      .where((s) => s.isNotEmpty)
      .toList();
}

String formatTelegramIdList(List<String> ids) => ids.join(', ');

class TelegramChannelDraft {
  TelegramChannelDraft.create() : base = null;

  TelegramChannelDraft.edit(TelegramChannel channel)
    : base = channel.deepCopy(),
      name = channel.name,
      receiveMode =
          channel.receiveMode ==
              TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_UNSPECIFIED
          ? TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK
          : channel.receiveMode {
    _initial = snapshot();
  }

  final TelegramChannel? base;
  bool get isCreate => base == null;

  String key = '';
  String name = '';
  TelegramReceiveMode receiveMode =
      TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK;
  String botToken = '';

  late final String _initial;

  String snapshot() => [key, name, receiveMode.value, botToken].join('\u0001');

  bool get dirty => snapshot() != _initial;

  String? validateCreate() {
    if (key.trim().isEmpty) return 'Enter a channel key.';
    if (botToken.trim().isEmpty) return 'Enter the Bot token from BotFather.';
    return null;
  }

  TelegramChannel buildCreate() => TelegramChannel()
    ..key = key.trim()
    ..name = name.trim()
    ..receiveMode = receiveMode;

  TelegramChannel buildUpdate() {
    final out = base!.deepCopy()
      ..name = name.trim()
      ..receiveMode = receiveMode;
    return out;
  }
}

class TelegramDestinationDraft {
  TelegramDestinationDraft.create({required this.channelId}) : base = null;

  TelegramDestinationDraft.edit(TelegramDestination destination)
    : base = destination.deepCopy(),
      channelId = destination.channelId,
      key = destination.key,
      name = destination.name,
      chatId = destination.chatId,
      messageThreadId = destination.messageThreadId,
      inboundEnabled = destination.inboundEnabled,
      outboundEnabled = destination.outboundEnabled {
    final config = destination.config;
    agentId = config.agentId;
    model = config.model;
    selectableAgentIds = formatTelegramIdList(config.selectableAgentIds);
    selectableModels = formatTelegramIdList(config.selectableModels);
    triggerMode =
        config.triggerMode ==
            TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_UNSPECIFIED
        ? TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_ALL
        : config.triggerMode;
    sessionPolicy =
        config.sessionPolicy ==
            TelegramSessionPolicy.TELEGRAM_SESSION_POLICY_UNSPECIFIED
        ? TelegramSessionPolicy.TELEGRAM_SESSION_POLICY_DESTINATION
        : config.sessionPolicy;
    allowedUserIds = formatTelegramIdList(config.allowedUserIds);
    controllerUserIds = formatTelegramIdList(config.controllerUserIds);
    replyMode =
        config.replyMode == TelegramReplyMode.TELEGRAM_REPLY_MODE_UNSPECIFIED
        ? TelegramReplyMode.TELEGRAM_REPLY_MODE_REPLY
        : config.replyMode;
    debugDefault = config.hasDebugDefault() ? config.debugDefault : true;
    _initial = snapshot();
  }

  final TelegramDestination? base;
  final String channelId;
  bool get isCreate => base == null;

  String key = '';
  String name = '';
  String chatId = '';
  String messageThreadId = '';
  bool inboundEnabled = true;
  bool outboundEnabled = true;

  String agentId = '';
  String model = '';
  String selectableAgentIds = '';
  String selectableModels = '';
  TelegramTriggerMode triggerMode =
      TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_ALL;
  TelegramSessionPolicy sessionPolicy =
      TelegramSessionPolicy.TELEGRAM_SESSION_POLICY_DESTINATION;
  String allowedUserIds = '';
  String controllerUserIds = '';
  TelegramReplyMode replyMode = TelegramReplyMode.TELEGRAM_REPLY_MODE_REPLY;
  bool debugDefault = true;

  late final String _initial;

  String snapshot() => [
    key,
    name,
    chatId,
    messageThreadId,
    inboundEnabled,
    outboundEnabled,
    agentId,
    model,
    selectableAgentIds,
    selectableModels,
    triggerMode.value,
    sessionPolicy.value,
    allowedUserIds,
    controllerUserIds,
    replyMode.value,
    debugDefault,
  ].join('\u0001');

  bool get dirty => snapshot() != _initial;

  bool agentOwnsModel(Agent? agent) =>
      agent != null &&
      (agent.type == AgentType.AGENT_TYPE_PI ||
          agent.type == AgentType.AGENT_TYPE_CURSOR);

  String? validate({required bool ownsModel}) {
    if (isCreate) {
      if (key.trim().isEmpty) return 'Enter a destination key.';
      if (!isTelegramDecimalId(chatId, allowNegative: true)) {
        return 'Chat ID must be a decimal string (for example -1001234567890).';
      }
      final thread = messageThreadId.trim();
      if (thread.isNotEmpty &&
          !isTelegramDecimalId(thread, allowNegative: false)) {
        return 'Topic ID must be a positive decimal string, or leave empty for no topic.';
      }
    }
    if (inboundEnabled && agentId.trim().isEmpty) {
      return 'Choose a default Agent for inbound messages.';
    }
    for (final list in [
      allowedUserIds,
      controllerUserIds,
      if (!ownsModel) selectableModels,
    ]) {
      for (final id in parseTelegramIdList(list)) {
        if (!isTelegramDecimalId(id, allowNegative: false)) {
          return 'User and model lists use decimal IDs only.';
        }
      }
    }
    if (!ownsModel) {
      for (final id in parseTelegramIdList(selectableAgentIds)) {
        if (id.trim().isEmpty) continue;
      }
    }
    return null;
  }

  TelegramDestinationConfig _buildConfig({required bool ownsModel}) {
    final config = TelegramDestinationConfig()
      ..agentId = agentId.trim()
      ..triggerMode = triggerMode
      ..sessionPolicy = sessionPolicy
      ..allowedUserIds.addAll(parseTelegramIdList(allowedUserIds))
      ..controllerUserIds.addAll(parseTelegramIdList(controllerUserIds))
      ..replyMode = replyMode
      ..debugDefault = debugDefault;
    if (!ownsModel) {
      config.model = model.trim();
      config.selectableModels.addAll(parseTelegramIdList(selectableModels));
      config.selectableAgentIds.addAll(parseTelegramIdList(selectableAgentIds));
    }
    return config;
  }

  TelegramDestination buildCreate({required bool ownsModel}) =>
      TelegramDestination()
        ..key = key.trim()
        ..name = name.trim()
        ..channelId = channelId
        ..chatId = chatId.trim()
        ..messageThreadId = messageThreadId.trim()
        ..inboundEnabled = inboundEnabled
        ..outboundEnabled = outboundEnabled
        ..config = _buildConfig(ownsModel: ownsModel);

  TelegramDestination buildUpdate({required bool ownsModel}) {
    final out = base!.deepCopy()
      ..name = name.trim()
      ..inboundEnabled = inboundEnabled
      ..outboundEnabled = outboundEnabled
      ..config = _buildConfig(ownsModel: ownsModel);
    return out;
  }
}
