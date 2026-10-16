import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter/foundation.dart';

import '../../core/butter_api.dart';
import '../../core/telegram_errors.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/telegram.pb.dart';
import 'telegram_draft.dart';

enum TelegramLoad { idle, loading, ready, failed }

/// Telegram Channels, Destinations and processing records for one scope.
class TelegramController extends ChangeNotifier {
  TelegramController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.onUnauthenticated,
    required this.agents,
    required this.modelAliases,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final void Function() onUnauthenticated;
  final List<Agent> Function() agents;
  final List<String> Function() modelAliases;

  ButterApi? _apiOrNull;
  ButterApi get _api => _apiOrNull ??= _newApi();
  bool _disposed = false;
  bool get disposed => _disposed;
  int _epoch = 0;
  bool _live(int epoch) => !_disposed && epoch == _epoch;
  void _emit() {
    if (!_disposed) notifyListeners();
  }

  int _clock = 0;
  int _lastWriteTick = 0;
  bool _fresh(int startTick) => startTick > _lastWriteTick;
  int _beginRead() => ++_clock;
  void _beginWrite() {
    _clock++;
    _lastWriteTick = _clock;
  }

  void _endWrite() => _clock++;

  bool _expired(Object failure) {
    final expired =
        failure is ConnectException && failure.code == Code.unauthenticated;
    if (expired) onUnauthenticated();
    return expired;
  }

  String? writeError;
  String? writeScope;
  bool get writeBlocked => writeError == telegramUnconfirmed;

  void clearWriteError() {
    writeError = null;
    writeScope = null;
    _emit();
  }

  // ---- Channels ----

  List<TelegramChannel> channels = const [];
  TelegramLoad channelListState = TelegramLoad.idle;
  String? channelListError;
  int _channelListSeq = 0;

  TelegramChannel? channel;
  TelegramLoad channelState = TelegramLoad.idle;
  String? channelError;
  int _channelSeq = 0;

  TelegramChannelStatus? channelStatus;
  TelegramLoad statusState = TelegramLoad.idle;
  String? statusError;
  int _statusSeq = 0;

  String? platformWebhookBase;
  bool platformSettingsDenied = false;

  Future<void> loadChannels() async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_channelListSeq, tick = _beginRead();
    channelListState = TelegramLoad.loading;
    channelListError = null;
    _emit();
    try {
      final list = await _api.listTelegramChannels(token, workspaceId);
      if (!_live(epoch) || seq != _channelListSeq || !_fresh(tick)) return;
      channels = List.unmodifiable(list);
      channelListState = TelegramLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _channelListSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      channelListState = TelegramLoad.failed;
      channelListError = telegramFailureMessage(
        failure,
        TelegramOperation.load,
      );
    }
    _emit();
  }

  Future<void> loadChannel(String id) async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_channelSeq, tick = _beginRead();
    channelState = TelegramLoad.loading;
    channelError = null;
    _emit();
    try {
      final ch = await _api.getTelegramChannel(token, workspaceId, id);
      if (!_live(epoch) || seq != _channelSeq || !_fresh(tick)) return;
      channel = ch;
      channelState = TelegramLoad.ready;
      if (writeScope == 'channel:$id' && writeBlocked) {
        clearWriteError();
      }
    } catch (failure) {
      if (!_live(epoch) || seq != _channelSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      channelState = TelegramLoad.failed;
      channelError = telegramFailureMessage(failure, TelegramOperation.load);
    }
    _emit();
  }

  Future<void> loadChannelStatus(String id) async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_statusSeq, tick = _beginRead();
    statusState = TelegramLoad.loading;
    statusError = null;
    _emit();
    try {
      final st = await _api.getTelegramChannelStatus(token, workspaceId, id);
      if (!_live(epoch) || seq != _statusSeq || !_fresh(tick)) return;
      channelStatus = st;
      statusState = TelegramLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _statusSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      statusState = TelegramLoad.failed;
      statusError = telegramFailureMessage(failure, TelegramOperation.status);
    }
    _emit();
  }

  Future<void> loadPlatformSettings() async {
    if (disposed) return;
    try {
      final settings = await _api.getTelegramSettings(token);
      if (disposed) return;
      platformSettingsDenied = settings == null;
      platformWebhookBase = settings?.webhookBaseUrl;
      _emit();
    } catch (_) {
      // Non-fatal; channel status blockers remain authoritative.
    }
  }

  Future<bool> createChannel(
    TelegramChannelDraft draft,
    String botToken,
  ) async {
    if (disposed || writeBlocked) return false;
    _beginWrite();
    writeScope = 'channels';
    writeError = null;
    _emit();
    try {
      final created = await _api.createTelegramChannel(
        token,
        workspaceId,
        channel: draft.buildCreate(),
        botToken: botToken,
      );
      if (disposed) return false;
      _endWrite();
      channels = List.unmodifiable([...channels, created]);
      channel = created;
      channelState = TelegramLoad.ready;
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.createChannel,
      );
      _emit();
      return false;
    }
  }

  Future<bool> updateChannel(TelegramChannelDraft draft) async {
    if (disposed || writeBlocked || draft.base == null) return false;
    final id = draft.base!.id;
    _beginWrite();
    writeScope = 'channel:$id';
    writeError = null;
    _emit();
    try {
      final updated = await _api.updateTelegramChannel(
        token,
        workspaceId,
        draft.buildUpdate(),
      );
      if (disposed) return false;
      _endWrite();
      channel = updated;
      channels = List.unmodifiable([
        for (final c in channels) c.id == id ? updated : c,
      ]);
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.updateChannel,
      );
      _emit();
      return false;
    }
  }

  Future<bool> rotateChannelToken(String channelId, String botToken) async {
    if (disposed || writeBlocked || botToken.trim().isEmpty) return false;
    _beginWrite();
    writeScope = 'channel:$channelId';
    writeError = null;
    _emit();
    try {
      final updated = await _api.rotateTelegramChannelCredential(
        token,
        workspaceId,
        channelId,
        botToken.trim(),
      );
      if (disposed) return false;
      _endWrite();
      channel = updated;
      channels = List.unmodifiable([
        for (final c in channels) c.id == channelId ? updated : c,
      ]);
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.rotateToken,
      );
      _emit();
      return false;
    }
  }

  Future<({bool ok, List<String> warnings})> setChannelEnabled({
    required String channelId,
    required Int64 revision,
    required bool inbound,
    required bool outbound,
  }) async {
    if (disposed || writeBlocked) {
      return (ok: false, warnings: const <String>[]);
    }
    _beginWrite();
    writeScope = 'channel:$channelId';
    writeError = null;
    _emit();
    try {
      final result = await _api.setTelegramChannelEnabled(
        token,
        workspaceId,
        channelId: channelId,
        revision: revision,
        inboundEnabled: inbound,
        outboundEnabled: outbound,
      );
      if (disposed) return (ok: false, warnings: const <String>[]);
      _endWrite();
      channel = result.channel;
      channels = List.unmodifiable([
        for (final c in channels) c.id == channelId ? result.channel : c,
      ]);
      clearWriteError();
      _emit();
      return (ok: true, warnings: List<String>.from(result.warnings));
    } catch (failure) {
      if (disposed) return (ok: false, warnings: const <String>[]);
      _endWrite();
      if (_expired(failure)) return (ok: false, warnings: const <String>[]);
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.setEnabled,
      );
      _emit();
      return (ok: false, warnings: const <String>[]);
    }
  }

  Future<bool> deleteChannel(String id) async {
    if (disposed || writeBlocked) return false;
    _beginWrite();
    writeScope = 'channels';
    writeError = null;
    _emit();
    try {
      await _api.deleteTelegramChannel(token, workspaceId, id);
      if (disposed) return false;
      _endWrite();
      channels = List.unmodifiable([
        for (final c in channels)
          if (c.id != id) c,
      ]);
      if (channel?.id == id) {
        channel = null;
        channelState = TelegramLoad.idle;
      }
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.deleteChannel,
      );
      _emit();
      return false;
    }
  }

  // ---- Destinations ----

  List<TelegramDestination> destinations = const [];
  TelegramLoad destinationListState = TelegramLoad.idle;
  String? destinationListError;
  int _destListSeq = 0;
  String? _destFilterChannel;

  TelegramDestination? destination;
  TelegramLoad destinationState = TelegramLoad.idle;
  String? destinationError;
  int _destSeq = 0;

  Future<void> loadDestinations({String channelId = ''}) async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_destListSeq, tick = _beginRead();
    _destFilterChannel = channelId;
    destinationListState = TelegramLoad.loading;
    destinationListError = null;
    _emit();
    try {
      final list = await _api.listTelegramDestinations(
        token,
        workspaceId,
        channelId: channelId,
      );
      if (!_live(epoch) || seq != _destListSeq || !_fresh(tick)) return;
      destinations = List.unmodifiable(list);
      destinationListState = TelegramLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _destListSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      destinationListState = TelegramLoad.failed;
      destinationListError = telegramFailureMessage(
        failure,
        TelegramOperation.load,
      );
    }
    _emit();
  }

  Future<void> loadDestination(String id) async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_destSeq, tick = _beginRead();
    destinationState = TelegramLoad.loading;
    destinationError = null;
    _emit();
    try {
      final dest = await _api.getTelegramDestination(token, workspaceId, id);
      if (!_live(epoch) || seq != _destSeq || !_fresh(tick)) return;
      destination = dest;
      destinationState = TelegramLoad.ready;
      if (writeScope == 'destination:$id' && writeBlocked) {
        clearWriteError();
      }
    } catch (failure) {
      if (!_live(epoch) || seq != _destSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      destinationState = TelegramLoad.failed;
      destinationError = telegramFailureMessage(
        failure,
        TelegramOperation.load,
      );
    }
    _emit();
  }

  Future<bool> createDestination(
    TelegramDestinationDraft draft, {
    required bool ownsModel,
  }) async {
    if (disposed || writeBlocked) return false;
    _beginWrite();
    writeScope = 'destinations';
    writeError = null;
    _emit();
    try {
      final created = await _api.createTelegramDestination(
        token,
        workspaceId,
        draft.buildCreate(ownsModel: ownsModel),
      );
      if (disposed) return false;
      _endWrite();
      destinations = List.unmodifiable([...destinations, created]);
      destination = created;
      destinationState = TelegramLoad.ready;
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.createDestination,
      );
      _emit();
      return false;
    }
  }

  Future<bool> updateDestination(
    TelegramDestinationDraft draft, {
    required bool ownsModel,
  }) async {
    if (disposed || writeBlocked || draft.base == null) return false;
    final id = draft.base!.id;
    _beginWrite();
    writeScope = 'destination:$id';
    writeError = null;
    _emit();
    try {
      final updated = await _api.updateTelegramDestination(
        token,
        workspaceId,
        draft.buildUpdate(ownsModel: ownsModel),
      );
      if (disposed) return false;
      _endWrite();
      destination = updated;
      destinations = List.unmodifiable([
        for (final d in destinations) d.id == id ? updated : d,
      ]);
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.updateDestination,
      );
      _emit();
      return false;
    }
  }

  Future<bool> deleteDestination(String id) async {
    if (disposed || writeBlocked) return false;
    _beginWrite();
    writeScope = 'destinations';
    writeError = null;
    _emit();
    try {
      await _api.deleteTelegramDestination(token, workspaceId, id);
      if (disposed) return false;
      _endWrite();
      destinations = List.unmodifiable([
        for (final d in destinations)
          if (d.id != id) d,
      ]);
      if (destination?.id == id) {
        destination = null;
        destinationState = TelegramLoad.idle;
      }
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(
        failure,
        TelegramOperation.deleteDestination,
      );
      _emit();
      return false;
    }
  }

  Future<({bool ok, List<String> messageIds})> sendTestMessage(
    String destinationId, {
    String text = '',
  }) async {
    if (disposed || writeBlocked) {
      return (ok: false, messageIds: const <String>[]);
    }
    _beginWrite();
    writeScope = 'destination:$destinationId';
    writeError = null;
    _emit();
    try {
      final reply = await _api.sendTelegramTestMessage(
        token,
        workspaceId,
        destinationId,
        text: text,
      );
      if (disposed) return (ok: false, messageIds: const <String>[]);
      _endWrite();
      if (reply.hasDestination()) {
        destination = reply.destination;
        destinations = List.unmodifiable([
          for (final d in destinations)
            d.id == destinationId ? reply.destination : d,
        ]);
      }
      clearWriteError();
      _emit();
      return (ok: true, messageIds: List<String>.from(reply.messageIds));
    } catch (failure) {
      if (disposed) return (ok: false, messageIds: const <String>[]);
      _endWrite();
      if (_expired(failure)) return (ok: false, messageIds: const <String>[]);
      writeError = telegramFailureMessage(failure, TelegramOperation.testSend);
      _emit();
      return (ok: false, messageIds: const <String>[]);
    }
  }

  // ---- Processing ----

  List<TelegramProcessingRecord> records = const [];
  TelegramLoad processingState = TelegramLoad.idle;
  String? processingError;
  int _processingSeq = 0;

  TelegramProcessingRecord? record;
  TelegramLoad recordState = TelegramLoad.idle;
  String? recordError;
  int _recordSeq = 0;

  Future<void> loadProcessingRecords({
    String channelId = '',
    String destinationId = '',
  }) async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_processingSeq, tick = _beginRead();
    processingState = TelegramLoad.loading;
    processingError = null;
    _emit();
    try {
      final list = await _api.listTelegramProcessingRecords(
        token,
        workspaceId,
        channelId: channelId,
        destinationId: destinationId,
      );
      if (!_live(epoch) || seq != _processingSeq || !_fresh(tick)) return;
      records = List.unmodifiable(list);
      processingState = TelegramLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _processingSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      processingState = TelegramLoad.failed;
      processingError = telegramFailureMessage(
        failure,
        TelegramOperation.processing,
      );
    }
    _emit();
  }

  Future<void> loadProcessingRecord(String id) async {
    if (disposed) return;
    final epoch = _epoch, seq = ++_recordSeq, tick = _beginRead();
    recordState = TelegramLoad.loading;
    recordError = null;
    _emit();
    try {
      final rec = await _api.getTelegramProcessingRecord(
        token,
        workspaceId,
        id,
      );
      if (!_live(epoch) || seq != _recordSeq || !_fresh(tick)) return;
      record = rec;
      recordState = TelegramLoad.ready;
      if (writeScope == 'resend:$id' && writeBlocked) {
        clearWriteError();
      }
    } catch (failure) {
      if (!_live(epoch) || seq != _recordSeq) return;
      if (_expired(failure)) return;
      if (!_fresh(tick)) return;
      recordState = TelegramLoad.failed;
      recordError = telegramFailureMessage(
        failure,
        TelegramOperation.processing,
      );
    }
    _emit();
  }

  Future<bool> resendReply(String id) async {
    if (disposed || writeBlocked) return false;
    _beginWrite();
    writeScope = 'resend:$id';
    writeError = null;
    _emit();
    try {
      final reply = await _api.resendTelegramReply(token, workspaceId, id);
      if (disposed) return false;
      _endWrite();
      if (reply.hasRecord()) {
        record = reply.record;
        records = List.unmodifiable([
          for (final r in records) r.id == id ? reply.record : r,
        ]);
      }
      clearWriteError();
      _emit();
      return true;
    } catch (failure) {
      if (disposed) return false;
      _endWrite();
      if (_expired(failure)) return false;
      writeError = telegramFailureMessage(failure, TelegramOperation.resend);
      _emit();
      return false;
    }
  }

  void suspend() {
    _epoch++;
    channels = const [];
    channel = null;
    channelStatus = null;
    destinations = const [];
    destination = null;
    records = const [];
    record = null;
    channelListState = TelegramLoad.idle;
    channelState = TelegramLoad.idle;
    destinationListState = TelegramLoad.idle;
    destinationState = TelegramLoad.idle;
    processingState = TelegramLoad.idle;
    recordState = TelegramLoad.idle;
    _emit();
  }

  Future<void> recover() async {
    await loadChannels();
    if (channel != null) {
      await loadChannel(channel!.id);
      await loadChannelStatus(channel!.id);
    }
    await loadDestinations(channelId: _destFilterChannel ?? '');
    if (destination != null) {
      await loadDestination(destination!.id);
    }
    if (records.isNotEmpty || processingState == TelegramLoad.ready) {
      await loadProcessingRecords();
    }
    if (record != null) {
      await loadProcessingRecord(record!.id);
    }
    await loadPlatformSettings();
  }

  @override
  void dispose() {
    _disposed = true;
    _epoch++;
    super.dispose();
  }
}
