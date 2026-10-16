import 'dart:async';

import 'package:connectrpc/connect.dart';

import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../gen/agents/v1/agent_file.pb.dart';
import 'file_rules.dart';

/// One version of a file as read from the server.
class FileDocument {
  const FileDocument({
    required this.file,
    required this.content,
    required this.latest,
  });
  final AgentFile file;
  final String content;

  /// The newest version the server held when the reader looked.
  final int latest;

  int get version => file.version.toInt();
  bool get isLatest => version >= latest;
}

/// The answer to one file read. A read never throws and never changes the
/// controller; a [stale] answer belongs to a context or a moment that has
/// passed and must be ignored.
class FileRead {
  const FileRead.ok(FileDocument this.document)
    : error = null,
      notFound = false,
      stale = false;
  const FileRead.failed(String this.error, {this.notFound = false})
    : document = null,
      stale = false;
  const FileRead.stale()
    : document = null,
      error = null,
      notFound = false,
      stale = true;
  final FileDocument? document;
  final String? error;
  final bool notFound;
  final bool stale;
}

enum FileConflictKind {
  /// A new file was started where one already exists.
  exists,

  /// The file moved on since the draft's version.
  changed,

  /// The file was deleted since the draft's version.
  deleted,
}

/// The server has no compare-and-swap: every write makes a new version. The
/// App therefore reads the file just before it writes and refuses a save that
/// would silently replace something the person never saw. It narrows the
/// window, it does not close it, and says so in its wording.
class FileConflict extends ConfigUnsupported {
  FileConflict(this.kind, this.path, {this.currentVersion})
    : super(switch (kind) {
        FileConflictKind.exists =>
          'A file already exists at $path, so nothing was saved. Open it, or choose another path.',
        FileConflictKind.changed =>
          '$path changed since you opened it, so nothing was saved.',
        FileConflictKind.deleted =>
          '$path was deleted since you opened it, so nothing was saved.',
      });
  final FileConflictKind kind;
  final String path;
  final int? currentVersion;
}

/// How an unconfirmed write turned out when the file was read again.
enum FileVerification { applied, notApplied, unknown }

/// The Agents that mount a space, as far as they could be read.
class SpaceReferences {
  const SpaceReferences(this.agents, {required this.complete});
  final List<String> agents;
  final bool complete;
}

/// File spaces and the files in them for one server, user and Workspace.
///
/// Spaces are listed, read, created, updated and deleted by the base class.
/// The open space's file list lives here; file reads return values and do not
/// touch the controller. Every file write is one explicit call: it is refused
/// while another write is in flight or unconfirmed, it is never repeated, and
/// an ambiguous outcome is settled only by a later read of the same file.
class FileSpaceController extends ConfigResourceController<AgentFileSpace> {
  FileSpaceController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
  }) : super(kind: ConfigKind.fileSpace);

  static const fileScope = 'file';

  /// Whether a file write is waiting to be settled by a read of that file.
  bool get fileWriteUnconfirmed => unconfirmedIds(fileScope).isNotEmpty;

  @override
  String idOf(AgentFileSpace item) => item.id;

  @override
  Future<List<AgentFileSpace>> readAll(ButterApi api) =>
      api.fileSpaces(token, workspaceId);

  @override
  Future<AgentFileSpace> readOne(ButterApi api, String id) =>
      api.fileSpace(token, workspaceId, id);

  @override
  Future<AgentFileSpace> createOne(ButterApi api, AgentFileSpace draft) =>
      api.createFileSpace(token, workspaceId, draft);

  @override
  Future<AgentFileSpace> updateOne(ButterApi api, AgentFileSpace draft) =>
      api.updateFileSpace(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteFileSpace(token, workspaceId, id);

  @override
  String writeFailureMessage(
    Object failure,
    ConfigOperation operation,
    String scope,
  ) => configFailureMessage(
    failure,
    scope == fileScope ? ConfigKind.file : kind,
    operation,
  );

  // ---- The open space's file list ----

  String? spaceId;
  List<AgentFile> files = const [];
  ConfigLoad filesState = ConfigLoad.idle;
  String? filesError;
  String filesPrefix = '';
  int _filesSeq = 0;
  bool _reloadFiles = false;

  /// Starts reading [id]'s files; anything held for another space is dropped.
  Future<void> openSpace(String id, {String prefix = ''}) async {
    if (spaceId != id) {
      spaceId = id;
      files = const [];
      filesState = ConfigLoad.idle;
    }
    await loadFiles(prefix: prefix);
  }

  void closeSpace({bool notify = true}) {
    spaceId = null;
    files = const [];
    filesState = ConfigLoad.idle;
    filesError = null;
    filesPrefix = '';
    _filesSeq++;
    if (notify) emit();
  }

  /// Reads the open space's files under [prefix] (empty: every file). A read
  /// that started before a write began or ended is dropped: it may carry the
  /// state from before it.
  Future<void> loadFiles({String? prefix}) async {
    final id = spaceId;
    if (disposed || id == null) return;
    final wanted = prefix ?? filesPrefix;
    final normalized = normalizeFilePrefix(wanted);
    if (normalized == null) {
      filesError = 'That is not a valid path prefix.';
      filesState = ConfigLoad.failed;
      emit();
      return;
    }
    final epoch = this.epoch, seq = ++_filesSeq, started = beginRead();
    final before = filesState == ConfigLoad.loading
        ? ConfigLoad.idle
        : filesState;
    filesPrefix = wanted;
    filesState = ConfigLoad.loading;
    filesError = null;
    _reloadFiles = true;
    emit();
    try {
      final list = await api.agentFiles(
        token,
        workspaceId,
        id,
        pathPrefix: normalized,
      );
      if (!live(epoch) || seq != _filesSeq || spaceId != id) return;
      if (!readIsFresh(started)) {
        filesState = before;
        emit();
        return;
      }
      files = List.unmodifiable(list);
      filesState = ConfigLoad.ready;
      if (normalized.isEmpty) {
        for (final w in unconfirmedIds(fileScope)) {
          if (w != null && w.startsWith('$id:')) {
            reconcile(started, id: w, scope: fileScope);
          }
        }
      }
    } catch (failure) {
      if (!live(epoch) || seq != _filesSeq || spaceId != id) return;
      if (!readIsFresh(started)) {
        filesState = before;
        emit();
        return;
      }
      if (expired(failure)) return;
      files = const [];
      filesState = ConfigLoad.failed;
      filesError = configFailureMessage(
        failure,
        ConfigKind.file,
        ConfigOperation.load,
      );
    }
    emit();
  }

  // ---- Reads that return values ----

  /// Reads [path] (the latest version, or [version] when above zero).
  /// [latest] is the newest version known to the caller, needed to know how
  /// an older version relates to it.
  Future<FileRead> readFile(
    String space,
    String path, {
    int version = 0,
    int? latest,
  }) async {
    if (disposed) return const FileRead.stale();
    final epoch = this.epoch, started = beginRead();
    try {
      final reply = await api.readAgentFile(
        token,
        workspaceId,
        space,
        path,
        version: version,
      );
      if (!live(epoch) || !readIsFresh(started)) return const FileRead.stale();
      final read = reply.file.version.toInt();
      // Only the latest version says what the file is now.
      if (version == 0) {
        reconcile(started, id: '$space:$path', scope: fileScope);
      }
      return FileRead.ok(
        FileDocument(
          file: reply.file,
          content: reply.content,
          latest: version > 0 ? (latest ?? read) : read,
        ),
      );
    } catch (failure) {
      if (!live(epoch) || !readIsFresh(started)) return const FileRead.stale();
      if (expired(failure)) return const FileRead.stale();
      final missing =
          failure is ConnectException && failure.code == Code.notFound;
      if (missing && version == 0) {
        reconcile(started, id: '$space:$path', scope: fileScope);
      }
      return FileRead.failed(
        configFailureMessage(failure, ConfigKind.file, ConfigOperation.load),
        notFound: missing,
      );
    }
  }

  /// Searches the open space for [query] (a substring).
  Future<({List<AgentFileSearchResult>? results, String? error, bool stale})>
  search(String space, String query) async {
    if (disposed) return (results: null, error: null, stale: true);
    final text = query.trim();
    if (text.isEmpty) {
      return (results: null, error: 'Enter text to search for.', stale: false);
    }
    final epoch = this.epoch, started = beginRead();
    try {
      final found = await api.searchAgentFiles(token, workspaceId, space, text);
      if (!live(epoch) || !readIsFresh(started)) {
        return (results: null, error: null, stale: true);
      }
      return (results: found, error: null, stale: false);
    } catch (failure) {
      if (!live(epoch) || !readIsFresh(started) || expired(failure)) {
        return (results: null, error: null, stale: true);
      }
      return (
        results: null,
        error: configFailureMessage(
          failure,
          ConfigKind.file,
          ConfigOperation.load,
        ),
        stale: false,
      );
    }
  }

  // ---- File writes ----

  /// The conflict the last refused save found, for the screen to offer a
  /// choice; cleared by [clearConflict] and by the next write.
  FileConflict? conflict;

  void clearConflict() {
    if (conflict == null) return;
    conflict = null;
    emit();
  }

  /// Saves [content] at [path] as a new version.
  ///
  /// [baseVersion] is the version the draft started from (null: a new file).
  /// Unless [overwrite], the file is read just before the write and a save
  /// over a different version is refused (see [FileConflict]). [contentType]
  /// and [metadata] are carried over from the file being edited because a
  /// write replaces them. Returns the stored file, or null.
  Future<AgentFile?> saveFile({
    required String space,
    required String path,
    required String content,
    required int? baseVersion,
    String contentType = defaultFileContentType,
    Map<String, String> metadata = const {},
    bool overwrite = false,
  }) async {
    final normalized = normalizeFilePath(path);
    if (normalized == null) return null;
    conflict = null;
    AgentFile? saved;
    final ok = await runWrite(id: '$space:$normalized', scope: fileScope, (
      epoch,
    ) async {
      if (!overwrite) {
        final current = await _currentVersion(space, normalized, epoch);
        if (!live(epoch)) return;
        final problem = switch ((baseVersion, current)) {
          (null, final int _) => FileConflictKind.exists,
          (final int _, null) => FileConflictKind.deleted,
          (final int b, final int c) when b != c => FileConflictKind.changed,
          _ => null,
        };
        if (problem != null) {
          conflict = FileConflict(problem, normalized, currentVersion: current);
          throw conflict!;
        }
      }
      // The check above can take a while: a dead context sends nothing.
      if (!live(epoch)) return;
      final file = await api.writeAgentFile(
        token,
        workspaceId,
        space,
        normalized,
        content,
        contentType: contentType,
        metadata: metadata,
      );
      if (!live(epoch)) return;
      saved = file;
      if (spaceId == space) {
        files = List.unmodifiable(
          [
            for (final f in files)
              if (f.path != file.path) f,
            file,
          ]..sort((a, b) => a.path.compareTo(b.path)),
        );
      }
    }, ConfigOperation.update);
    return ok ? saved : null;
  }

  /// The latest version of the file, null when there is none. A failure to
  /// look is a definite refusal: nothing has been written yet.
  Future<int?> _currentVersion(String space, String path, int epoch) async {
    try {
      final reply = await api.readAgentFile(token, workspaceId, space, path);
      return reply.file.version.toInt();
    } on ConnectException catch (failure) {
      if (failure.code == Code.notFound) return null;
      if (failure.code == Code.unauthenticated) rethrow;
      throw const ConfigUnsupported(
        'Could not check the file\'s current version, so nothing was saved. Try again.',
      );
    }
  }

  Future<bool> deleteFile(String space, String path) =>
      runWrite(id: '$space:$path', scope: fileScope, (epoch) async {
        await api.deleteAgentFile(token, workspaceId, space, path);
        if (!live(epoch)) return;
        if (spaceId == space) {
          files = List.unmodifiable(files.where((f) => f.path != path));
        }
      }, ConfigOperation.delete);

  /// Reads [path] again after an unconfirmed write and says whether it took
  /// effect. [expectedContent] is what a save sent; null checks a delete.
  Future<FileVerification> verifyWrite(
    String space,
    String path, {
    String? expectedContent,
  }) async {
    final read = await readFile(space, path);
    if (read.stale) return FileVerification.unknown;
    if (expectedContent == null) {
      return read.notFound
          ? FileVerification.applied
          : read.document != null
          ? FileVerification.notApplied
          : FileVerification.unknown;
    }
    final doc = read.document;
    if (doc == null) {
      return read.notFound
          ? FileVerification.notApplied
          : FileVerification.unknown;
    }
    return doc.content == expectedContent
        ? FileVerification.applied
        : FileVerification.notApplied;
  }

  // ---- Deleting a space ----

  /// Which Agents of the Workspace mount [space], read page by page. A space
  /// delete has no reference guard on the server: this is only a warning, and
  /// says so when it could not see every Agent.
  Future<SpaceReferences?> agentsMounting(String space) async {
    if (disposed) return null;
    final epoch = this.epoch;
    final names = <String>[];
    final seen = <String>{};
    var page = '';
    try {
      do {
        final reply = await api.agents(token, workspaceId, pageToken: page);
        if (!live(epoch)) return null;
        for (final agent in reply.agents) {
          if (agent.config.fileMounts.any((m) => m.spaceId == space)) {
            names.add(agent.name.isEmpty ? agent.agentId : agent.name);
          }
        }
        page = reply.nextPageToken;
        if (page.isNotEmpty && !seen.add(page)) {
          return SpaceReferences(names, complete: false);
        }
      } while (page.isNotEmpty);
      return SpaceReferences(names, complete: true);
    } catch (failure) {
      if (!live(epoch) || expired(failure)) return null;
      return SpaceReferences(names, complete: false);
    }
  }

  @override
  void objectRemoved(String id) {
    if (spaceId == id) {
      files = const [];
      filesState = ConfigLoad.idle;
      _filesSeq++;
    }
  }

  @override
  void onSuspend() {
    // File names and contents do not wait in memory in the background.
    files = const [];
    filesState = ConfigLoad.idle;
    _filesSeq++;
  }

  @override
  Future<void> onRecover() async {
    if (spaceId != null && _reloadFiles) await loadFiles();
  }

  @override
  void onDispose() {
    files = const [];
    spaceId = null;
  }
}
