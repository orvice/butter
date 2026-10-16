import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../core/notify_group_errors.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/telegram.pb.dart';

class NotifyGroupController extends ConfigResourceController<NotifyGroup> {
  NotifyGroupController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
  }) : super(kind: ConfigKind.notifyGroup);

  @override
  String idOf(NotifyGroup item) => item.name;

  @override
  Future<List<NotifyGroup>> readAll(ButterApi api) =>
      api.listNotifyGroups(token, workspaceId);

  @override
  Future<NotifyGroup> readOne(ButterApi api, String id) =>
      api.getNotifyGroup(token, workspaceId, id);

  @override
  Future<NotifyGroup> createOne(ButterApi api, NotifyGroup draft) =>
      api.createNotifyGroup(token, workspaceId, draft);

  @override
  Future<NotifyGroup> updateOne(ButterApi api, NotifyGroup draft) =>
      api.updateNotifyGroup(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteNotifyGroup(token, workspaceId, id);

  @override
  String writeFailureMessage(
    Object failure,
    ConfigOperation operation,
    String scope,
  ) {
    final op = switch (operation) {
      ConfigOperation.create => NotifyGroupOperation.create,
      ConfigOperation.update => NotifyGroupOperation.update,
      ConfigOperation.delete => NotifyGroupOperation.delete,
      _ => NotifyGroupOperation.load,
    };
    return notifyGroupFailureMessage(failure, op);
  }

  List<TelegramDestination> telegramDestinations = const [];
  ConfigLoad referenceState = ConfigLoad.idle;
  String? referenceError;
  int _referenceSeq = 0;

  Future<void> loadReferences() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_referenceSeq;
    referenceState = ConfigLoad.loading;
    referenceError = null;
    emit();
    try {
      telegramDestinations = List.unmodifiable(
        await api.listTelegramDestinations(token, workspaceId),
      );
      if (!live(epoch) || seq != _referenceSeq) return;
      referenceState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _referenceSeq) return;
      if (expired(failure)) return;
      referenceState = ConfigLoad.failed;
      referenceError = notifyGroupFailureMessage(
        failure,
        NotifyGroupOperation.load,
      );
    }
    emit();
  }

  @override
  void onSuspend() {
    telegramDestinations = const [];
    referenceState = ConfigLoad.idle;
    referenceError = null;
  }

  @override
  Future<void> onRecover() async => loadReferences();

  Future<bool> toggleEnabled(NotifyGroup group, bool enabled) async =>
      update(group.deepCopy()..enabled = enabled);
}
