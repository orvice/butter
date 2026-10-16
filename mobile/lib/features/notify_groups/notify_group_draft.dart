import '../../gen/agents/v1/agent.pb.dart';

class NotifyTargetDraft {
  NotifyTargetDraft({
    this.name = '',
    this.enabled = true,
    this.type = NotifyTargetType.NOTIFY_TARGET_TYPE_TELEGRAM,
  });

  String name;
  bool enabled;
  NotifyTargetType type;
  String telegramDestinationId = '';
  String larkWebhookUrl = '';
  String larkSecret = '';
  String discordWebhookUrl = '';
  String discordUsername = '';
  String discordAvatarUrl = '';
  String discordThreadId = '';
  final Map<String, String> metadata = {};

  factory NotifyTargetDraft.fromProto(NotifyTarget t) {
    final d = NotifyTargetDraft(name: t.name, enabled: t.enabled, type: t.type);
    if (t.hasTelegram()) {
      d.telegramDestinationId = t.telegram.destinationId;
    }
    if (t.hasLark()) {
      d.larkWebhookUrl = t.lark.webhookUrl;
      d.larkSecret = t.lark.secret;
    }
    if (t.hasDiscord()) {
      d.discordWebhookUrl = t.discord.webhookUrl;
      d.discordUsername = t.discord.username;
      d.discordAvatarUrl = t.discord.avatarUrl;
      d.discordThreadId = t.discord.threadId;
    }
    d.metadata.addAll(t.metadata);
    return d;
  }

  NotifyTarget toProto() {
    final t = NotifyTarget(name: name, enabled: enabled, type: type)
      ..metadata.addAll(metadata);
    switch (type) {
      case NotifyTargetType.NOTIFY_TARGET_TYPE_TELEGRAM:
        t.telegram = TelegramNotifyTarget(
          destinationId: telegramDestinationId.trim(),
        );
      case NotifyTargetType.NOTIFY_TARGET_TYPE_LARK_WEBHOOK:
        t.lark = LarkNotifyTarget(
          webhookUrl: larkWebhookUrl.trim(),
          secret: larkSecret.trim(),
        );
      case NotifyTargetType.NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK:
        t.discord = DiscordNotifyTarget(
          webhookUrl: discordWebhookUrl.trim(),
          username: discordUsername.trim(),
          avatarUrl: discordAvatarUrl.trim(),
          threadId: discordThreadId.trim(),
        );
      default:
        break;
    }
    return t;
  }

  String? validate() {
    switch (type) {
      case NotifyTargetType.NOTIFY_TARGET_TYPE_TELEGRAM:
        if (telegramDestinationId.trim().isEmpty) {
          return 'Choose a Telegram Destination.';
        }
      case NotifyTargetType.NOTIFY_TARGET_TYPE_LARK_WEBHOOK:
        if (larkWebhookUrl.trim().isEmpty) {
          return 'Lark webhook URL is required.';
        }
      case NotifyTargetType.NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK:
        if (discordWebhookUrl.trim().isEmpty) {
          return 'Discord webhook URL is required.';
        }
      default:
        return 'Choose a target type.';
    }
    return null;
  }
}

/// Editable Notify Group form preserving unknown metadata.
class NotifyGroupDraft {
  NotifyGroupDraft.create() : base = null {
    _initial = snapshot();
  }

  NotifyGroupDraft.edit(NotifyGroup group) : base = group.deepCopy() {
    name = group.name;
    enabled = group.enabled;
    targets.addAll(group.targets.map(NotifyTargetDraft.fromProto));
    metadata.addAll(group.metadata);
    _initial = snapshot();
  }

  final NotifyGroup? base;
  bool get isCreate => base == null;

  String name = '';
  bool enabled = true;
  final List<NotifyTargetDraft> targets = [];
  final Map<String, String> metadata = {};

  late final String _initial;

  String snapshot() => [
    name,
    enabled.toString(),
    for (final t in targets)
      '${t.type.value}|${t.name}|${t.enabled}|${t.telegramDestinationId}|'
          '${t.larkWebhookUrl}|${t.discordWebhookUrl}',
    for (final e in metadata.entries) '${e.key}\u0001${e.value}',
  ].join('\u0000');

  bool get dirty => snapshot() != _initial;

  String? validate() {
    if (name.trim().isEmpty) return 'Name is required.';
    for (final t in targets) {
      final err = t.validate();
      if (err != null) return err;
    }
    return null;
  }

  NotifyGroup build() {
    final g = NotifyGroup(
      name: name.trim(),
      enabled: enabled,
      targets: targets.map((t) => t.toProto()),
    )..metadata.addAll(metadata);
    if (base != null) {
      g.workspaceId = base!.workspaceId;
    }
    return g;
  }
}

String notifyTargetTypeLabel(NotifyTargetType t) => switch (t) {
  NotifyTargetType.NOTIFY_TARGET_TYPE_TELEGRAM => 'Telegram',
  NotifyTargetType.NOTIFY_TARGET_TYPE_LARK_WEBHOOK => 'Lark Webhook',
  NotifyTargetType.NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK => 'Discord Webhook',
  _ => 'Unspecified',
};
