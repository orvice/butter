import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;

import '../../gen/agents/v1/cron.pb.dart';

/// How the user chose max output bytes: unset, explicit zero (server default),
/// or a positive limit.
enum MaxOutputChoice { unset, serverDefault, limit }

/// Editable cron job form. Starts from a deep copy of the stored message so
/// unknown fields and metadata survive saves.
class CronJobDraft {
  CronJobDraft.create() : base = null;

  CronJobDraft.edit(CronJob job) : base = job.deepCopy() {
    name = job.name;
    schedule = job.schedule;
    agentId = job.agentId;
    input = job.input;
    timezone = job.timezone;
    enabled = job.enabled;
    _loadDelivery(job.delivery);
    _loadTimeout(job.timeout);
    _loadRetry(job.retry);
    concurrencyPolicy = job.concurrencyPolicy;
    notifyOn = job.notifyOn;
    _loadMaxOutput(job.maxOutputBytes);
    metadata.addAll([
      for (final e
          in job.metadata.entries.toList()
            ..sort((a, b) => a.key.compareTo(b.key)))
        MetadataRow(key: e.key, value: e.value),
    ]);
    _initial = snapshot();
  }

  final CronJob? base;
  bool get isCreate => base == null;

  String name = '';
  String schedule = '@daily';
  String agentId = '';
  String input = '';
  String timezone = '';
  bool enabled = true;

  CronDeliveryType deliveryType = CronDeliveryType.CRON_DELIVERY_TYPE_LOG;
  String webhookUrl = '';
  String notifyGroupName = '';
  String telegramDestinationId = '';

  /// Seconds; empty means no cron-level timeout (absent on wire).
  String timeoutSeconds = '';

  String retryMaxAttempts = '0';
  String retryBackoffSeconds = '0';

  CronConcurrencyPolicy concurrencyPolicy =
      CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_SKIP;
  CronNotifyOn notifyOn = CronNotifyOn.CRON_NOTIFY_ON_FAILURE;

  MaxOutputChoice maxOutputChoice = MaxOutputChoice.unset;
  String maxOutputBytes = '';

  final List<MetadataRow> metadata = [];

  late final String _initial;

  String snapshot() => [
    name,
    schedule,
    agentId,
    input,
    timezone,
    enabled,
    deliveryType.value,
    webhookUrl,
    notifyGroupName,
    telegramDestinationId,
    timeoutSeconds,
    retryMaxAttempts,
    retryBackoffSeconds,
    concurrencyPolicy.value,
    notifyOn.value,
    maxOutputChoice.name,
    maxOutputBytes,
    for (final r in metadata) '${r.key}\u0001${r.value}',
  ].join('\u0000');

  bool get dirty => snapshot() != _initial;

  void _loadDelivery(CronDelivery d) {
    deliveryType = d.type;
    webhookUrl = d.webhookUrl;
    notifyGroupName = d.notifyGroupName;
    telegramDestinationId = d.telegramDestinationId;
  }

  void _loadTimeout(pbd.Duration? d) {
    if (d == null || !base!.hasTimeout()) {
      timeoutSeconds = '';
      return;
    }
    timeoutSeconds = '${d.seconds}';
  }

  void _loadRetry(CronRetryPolicy? r) {
    if (r == null || !base!.hasRetry()) {
      retryMaxAttempts = '0';
      retryBackoffSeconds = '0';
      return;
    }
    retryMaxAttempts = '${r.maxAttempts}';
    retryBackoffSeconds = r.hasBackoff() ? '${r.backoff.seconds}' : '0';
  }

  void _loadMaxOutput(int max) {
    if (!base!.hasMaxOutputBytes()) {
      maxOutputChoice = MaxOutputChoice.unset;
      maxOutputBytes = '';
      return;
    }
    if (max == 0) {
      maxOutputChoice = MaxOutputChoice.serverDefault;
      maxOutputBytes = '';
    } else {
      maxOutputChoice = MaxOutputChoice.limit;
      maxOutputBytes = '$max';
    }
  }

  int? _parseNonNegativeInt(String raw, {required String label}) {
    final t = raw.trim();
    if (t.isEmpty) return 0;
    final n = int.tryParse(t);
    if (n == null || n < 0) throw FormatException('Enter a valid $label.');
    return n;
  }

  String? validateName() {
    if (!isCreate) return null;
    final t = name.trim();
    if (t.isEmpty) return 'Enter a name.';
    if (t.contains(RegExp(r'\s'))) return 'A name cannot contain spaces.';
    return null;
  }

  String? validateSchedule() =>
      schedule.trim().isEmpty ? 'Enter a schedule.' : null;

  String? validateAgent() => agentId.trim().isEmpty ? 'Choose an Agent.' : null;

  String? validateDelivery() => switch (deliveryType) {
    CronDeliveryType.CRON_DELIVERY_TYPE_WEBHOOK =>
      webhookUrl.trim().isEmpty ? 'Enter a webhook URL.' : null,
    CronDeliveryType.CRON_DELIVERY_TYPE_NOTIFY_GROUP =>
      notifyGroupName.trim().isEmpty ? 'Choose a Notify Group.' : null,
    CronDeliveryType.CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION =>
      telegramDestinationId.trim().isEmpty
          ? 'Choose a Telegram Destination.'
          : null,
    CronDeliveryType.CRON_DELIVERY_TYPE_CHANNEL =>
      'Legacy CHANNEL delivery is not supported. Choose another delivery type.',
    _ => null,
  };

  String? validateMetadata() {
    final seen = <String>{};
    for (final r in metadata) {
      if (r.isEmpty) continue;
      final k = r.key.trim();
      if (k.isEmpty) return 'Give each metadata row a key, or remove it.';
      if (!seen.add(k)) return 'The metadata key "$k" is used twice.';
    }
    return null;
  }

  String? firstError() =>
      validateName() ??
      validateSchedule() ??
      validateAgent() ??
      validateDelivery() ??
      validateMetadata();

  CronJob build() {
    final job = (base ?? CronJob()).deepCopy()
      ..name = isCreate ? name.trim() : base!.name
      ..schedule = schedule.trim()
      ..agentId = agentId.trim()
      ..input = input
      ..timezone = timezone.trim()
      ..enabled = enabled
      ..concurrencyPolicy = concurrencyPolicy
      ..notifyOn = notifyOn;

    final delivery = CronDelivery(type: deliveryType);
    switch (deliveryType) {
      case CronDeliveryType.CRON_DELIVERY_TYPE_WEBHOOK:
        delivery.webhookUrl = webhookUrl.trim();
      case CronDeliveryType.CRON_DELIVERY_TYPE_NOTIFY_GROUP:
        delivery.notifyGroupName = notifyGroupName.trim();
      case CronDeliveryType.CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION:
        delivery.telegramDestinationId = telegramDestinationId.trim();
      default:
        break;
    }
    job.delivery = delivery;

    final timeoutSec = _parseNonNegativeInt(timeoutSeconds, label: 'timeout');
    if (timeoutSec != null && timeoutSec > 0) {
      job.timeout = pbd.Duration(seconds: Int64(timeoutSec));
    } else {
      job.clearTimeout();
    }

    final attempts = _parseNonNegativeInt(
      retryMaxAttempts,
      label: 'retry count',
    )!;
    if (attempts > 0) {
      final backoffSec = _parseNonNegativeInt(
        retryBackoffSeconds,
        label: 'retry backoff',
      )!;
      job.retry = CronRetryPolicy(maxAttempts: attempts)
        ..backoff = backoffSec > 0
            ? pbd.Duration(seconds: Int64(backoffSec))
            : pbd.Duration();
    } else {
      job.clearRetry();
    }

    switch (maxOutputChoice) {
      case MaxOutputChoice.unset:
        job.clearMaxOutputBytes();
      case MaxOutputChoice.serverDefault:
        job.maxOutputBytes = 0;
      case MaxOutputChoice.limit:
        final limit = _parseNonNegativeInt(
          maxOutputBytes,
          label: 'max output bytes',
        )!;
        if (limit <= 0) {
          throw FormatException('Enter a positive max output byte limit.');
        }
        job.maxOutputBytes = limit;
    }

    job.metadata.clear();
    for (final r in metadata) {
      if (r.isEmpty) continue;
      job.metadata[r.key.trim()] = r.value;
    }
    return job;
  }
}

class MetadataRow {
  MetadataRow({this.key = '', this.value = ''});
  String key;
  String value;
  bool get isEmpty => key.trim().isEmpty && value.trim().isEmpty;
}

String cronDeliveryLabel(CronDeliveryType t) => switch (t) {
  CronDeliveryType.CRON_DELIVERY_TYPE_LOG => 'Log only',
  CronDeliveryType.CRON_DELIVERY_TYPE_WEBHOOK => 'Webhook',
  CronDeliveryType.CRON_DELIVERY_TYPE_NOTIFY_GROUP => 'Notify Group',
  CronDeliveryType.CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION =>
    'Telegram Destination',
  CronDeliveryType.CRON_DELIVERY_TYPE_CHANNEL => 'Channel (legacy)',
  _ => 'Unspecified',
};

String cronConcurrencyLabel(CronConcurrencyPolicy p) => switch (p) {
  CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_SKIP => 'Skip',
  CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_QUEUE => 'Queue',
  CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_REPLACE => 'Replace',
  CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_ALLOW => 'Allow',
  _ => 'Unspecified',
};

String cronNotifyOnLabel(CronNotifyOn n) => switch (n) {
  CronNotifyOn.CRON_NOTIFY_ON_ALWAYS => 'Always',
  CronNotifyOn.CRON_NOTIFY_ON_FAILURE => 'Failure',
  CronNotifyOn.CRON_NOTIFY_ON_SUCCESS => 'Success',
  _ => 'Unspecified',
};

String cronExecutionStatusLabel(CronExecutionStatus s) => switch (s) {
  CronExecutionStatus.CRON_EXECUTION_STATUS_SUCCESS => 'Success',
  CronExecutionStatus.CRON_EXECUTION_STATUS_ERROR => 'Error',
  CronExecutionStatus.CRON_EXECUTION_STATUS_SKIPPED => 'Skipped',
  CronExecutionStatus.CRON_EXECUTION_STATUS_CANCELLED => 'Cancelled',
  CronExecutionStatus.CRON_EXECUTION_STATUS_WAITING_INPUT =>
    'Waiting for input',
  _ => 'Unknown',
};
