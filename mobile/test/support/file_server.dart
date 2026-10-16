import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/features/files/file_rules.dart';
import 'package:butter_mobile/gen/agents/v1/agent_file.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

import 'mcp_server.dart';

class _Version {
  _Version(this.content, this.contentType);
  final String content;
  final String contentType;
}

class _Stored {
  _Stored(this.space);
  AgentFileSpace space;
  // path -> versions (index 0 is version 1)
  final files = <String, List<_Version>>{};
  // path -> the metadata of the latest write (not versioned, as on the server)
  final metadata = <String, Map<String, String>>{};
  final created = <String, DateTime>{};
}

/// External HTTP boundary for AgentFileService with controllable late,
/// refused and lost replies. The real semantics (normalization, versions, the
/// size limit, Workspace isolation, who may reach a Workspace) are proven
/// against the production Go handlers in files_contract_test; this fake keeps
/// to those semantics so the UI tests drive timing and failure shapes. It
/// extends the MCP fake so one server also serves Agents, MCP servers, Remote
/// Agents and Skills for the editor tests.
class FileServerFake extends McpServerFake {
  final stores = <String, Map<String, _Stored>>{'alpha': {}, 'beta': {}};
  final fileCalls = <({String method, String? workspace, List<int> body})>[];

  static const fileWriteMethods = {
    'CreateAgentFileSpace',
    'UpdateAgentFileSpace',
    'DeleteAgentFileSpace',
    'WriteAgentFile',
    'DeleteAgentFile',
  };
  int fileCount(String method) =>
      fileCalls.where((c) => c.method == method).length;
  int get fileWrites =>
      fileCalls.where((c) => fileWriteMethods.contains(c.method)).length;

  /// Held before a method answers; reads capture their answer first, so a
  /// delayed read is a stale one.
  Future<void> Function(String method)? fileBefore;

  /// Refuses a method (status, code, message), decided when it arrives.
  final fileRefusals = <String, (int, String, String)>{};

  /// Writes that are applied but whose answer is lost (503).
  final lostFileReplies = <String>{};

  int maxBytes = defaultFileMaxBytes;
  int _next = 0;

  /// Refuse everything as an older server without the service would.
  bool filesUnimplemented = false;

  FileServerFake() {
    final inherited = authorizedReply!;
    authorizedReply = (r, bytes) async {
      final path = r.uri.path;
      if (!path.contains('AgentFileService')) return inherited(r, bytes);
      final method = path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      fileCalls.add((method: method, workspace: ws, body: bytes));
      if (filesUnimplemented) {
        r.response.statusCode = 404;
        r.response.headers.contentType = ContentType.text;
        r.response.write('404 page not found');
        return true;
      }
      final bucket = stores[ws];
      if (bucket == null) return _fail(r, 403, 'permission_denied', 'denied');
      final refused = fileRefusals[method];
      if (refused != null && !fileWriteMethods.contains(method)) {
        // A refused read still waits its turn, like any other answer.
        await fileBefore?.call(method);
        return _fail(r, refused.$1, refused.$2, refused.$3);
      }
      if (refused != null) {
        return _fail(r, refused.$1, refused.$2, refused.$3);
      }
      if (fileWriteMethods.contains(method)) await fileBefore?.call(method);
      switch (method) {
        case 'ListAgentFileSpaces':
          final snapshot = [for (final s in bucket.values) _space(s, ws!)];
          await fileBefore?.call(method);
          write(r, ListAgentFileSpacesResponse(spaces: snapshot));
        case 'GetAgentFileSpace':
          final id = GetAgentFileSpaceRequest.fromBuffer(bytes).id;
          final found = bucket[id] == null ? null : _space(bucket[id]!, ws!);
          await fileBefore?.call(method);
          if (found == null) return _fail(r, 404, 'not_found', 'not found');
          write(r, GetAgentFileSpaceResponse(space: found));
        case 'CreateAgentFileSpace':
          final s = CreateAgentFileSpaceRequest.fromBuffer(bytes).space;
          if (s.name.isEmpty) {
            return _fail(r, 400, 'invalid_argument', 'space.name is required');
          }
          final id = s.id.isEmpty ? 'sp-${++_next}' : s.id;
          if (bucket.containsKey(id)) {
            return _fail(r, 409, 'already_exists', 'already exists');
          }
          final stored = _Stored(s.deepCopy()..id = id);
          bucket[id] = stored;
          _answer(
            method,
            r,
            CreateAgentFileSpaceResponse(space: _space(stored, ws!)),
          );
        case 'UpdateAgentFileSpace':
          final s = UpdateAgentFileSpaceRequest.fromBuffer(bytes).space;
          final stored = bucket[s.id];
          if (stored == null) return _fail(r, 404, 'not_found', 'not found');
          if (s.name.isEmpty) {
            return _fail(r, 400, 'invalid_argument', 'space.name is required');
          }
          // The whole record is replaced by what is sent.
          stored.space = s.deepCopy();
          _answer(
            method,
            r,
            UpdateAgentFileSpaceResponse(space: _space(stored, ws!)),
          );
        case 'DeleteAgentFileSpace':
          final id = DeleteAgentFileSpaceRequest.fromBuffer(bytes).id;
          if (bucket.remove(id) == null) {
            return _fail(r, 404, 'not_found', 'not found');
          }
          _answer(method, r, DeleteAgentFileSpaceResponse());
        case 'ListAgentFiles':
          final req = ListAgentFilesRequest.fromBuffer(bytes);
          final stored = bucket[req.spaceId];
          if (stored == null) return _fail(r, 404, 'not_found', 'not found');
          final prefix = normalizeFilePrefix(req.pathPrefix);
          if (prefix == null) {
            return _fail(r, 400, 'invalid_argument', 'invalid path');
          }
          final snapshot = [
            for (final p in stored.files.keys.toList()..sort())
              if (prefix.isEmpty || p == prefix || p.startsWith('$prefix/'))
                _file(stored, p, ws!),
          ];
          await fileBefore?.call(method);
          write(r, ListAgentFilesResponse(files: snapshot));
        case 'GetAgentFile':
          final req = GetAgentFileRequest.fromBuffer(bytes);
          final stored = bucket[req.spaceId];
          if (stored == null) return _fail(r, 404, 'not_found', 'not found');
          final p = normalizeFilePath(req.path);
          if (p == null) {
            return _fail(r, 400, 'invalid_argument', 'invalid path');
          }
          final versions = stored.files[p];
          final want = req.version.toInt();
          if (versions == null || want > versions.length) {
            await fileBefore?.call(method);
            return _fail(r, 404, 'not_found', 'not found');
          }
          final file = _file(stored, p, ws!, version: want == 0 ? null : want);
          final content = versions[(want == 0 ? versions.length : want) - 1];
          await fileBefore?.call(method);
          write(r, GetAgentFileResponse(file: file, content: content.content));
        case 'WriteAgentFile':
          final req = WriteAgentFileRequest.fromBuffer(bytes);
          final stored = bucket[req.spaceId];
          if (stored == null) return _fail(r, 404, 'not_found', 'not found');
          if (utf8Length(req.content) > maxBytes) {
            return _fail(
              r,
              400,
              'invalid_argument',
              'content exceeds max file size of $maxBytes bytes',
            );
          }
          final p = normalizeFilePath(req.path);
          if (p == null) {
            return _fail(r, 400, 'invalid_argument', 'invalid path');
          }
          (stored.files[p] ??= []).add(_Version(req.content, req.contentType));
          stored.metadata[p] = Map.of(req.metadata);
          stored.created.putIfAbsent(p, DateTime.now);
          _answer(
            method,
            r,
            WriteAgentFileResponse(file: _file(stored, p, ws!)),
          );
        case 'DeleteAgentFile':
          final req = DeleteAgentFileRequest.fromBuffer(bytes);
          final stored = bucket[req.spaceId];
          if (stored == null) return _fail(r, 404, 'not_found', 'not found');
          final p = normalizeFilePath(req.path);
          if (p == null || stored.files.remove(p) == null) {
            return _fail(r, 404, 'not_found', 'not found');
          }
          stored.metadata.remove(p);
          _answer(method, r, DeleteAgentFileResponse());
        case 'SearchAgentFiles':
          final req = SearchAgentFilesRequest.fromBuffer(bytes);
          final stored = bucket[req.spaceId];
          if (stored == null) return _fail(r, 404, 'not_found', 'not found');
          if (req.query.isEmpty) {
            return _fail(r, 400, 'invalid_argument', 'query is required');
          }
          final results = <AgentFileSearchResult>[];
          for (final p in stored.files.keys.toList()..sort()) {
            final text = stored.files[p]!.last.content;
            final at = text.toLowerCase().indexOf(req.query.toLowerCase());
            if (at < 0) continue;
            results.add(
              AgentFileSearchResult(
                file: _file(stored, p, ws!),
                snippets: [
                  text.substring(
                    at < 10 ? 0 : at - 10,
                    (at + req.query.length + 10).clamp(0, text.length),
                  ),
                ],
              ),
            );
          }
          await fileBefore?.call(method);
          write(r, SearchAgentFilesResponse(results: results));
        default:
          return false;
      }
      return true;
    };
  }

  /// Seeds a space (its ID is [id]) on [ws] and returns it.
  AgentFileSpace seedSpace(
    String ws,
    String id,
    String name, {
    String description = '',
    Map<String, String> metadata = const {},
    List<int> unknown = const [],
  }) {
    final space = AgentFileSpace(
      id: id,
      name: name,
      description: description,
      metadata: metadata.entries,
    );
    for (final tag in unknown) {
      space.unknownFields.mergeVarintField(tag, Int64(7));
    }
    stores[ws]![id] = _Stored(space);
    return space;
  }

  /// Seeds a file with one version per entry of [versions] (oldest first).
  void seedFile(
    String ws,
    String space,
    String path,
    List<String> versions, {
    String contentType = defaultFileContentType,
    Map<String, String> metadata = const {},
  }) {
    final stored = stores[ws]![space]!;
    stored.files[path] = [for (final v in versions) _Version(v, contentType)];
    stored.metadata[path] = Map.of(metadata);
    stored.created[path] = DateTime.now();
  }

  /// The latest content of a seeded or written file, or null.
  String? contentOf(String ws, String space, String path) =>
      stores[ws]?[space]?.files[path]?.last.content;

  int versionsOf(String ws, String space, String path) =>
      stores[ws]?[space]?.files[path]?.length ?? 0;

  Map<String, String>? metadataOf(String ws, String space, String path) =>
      stores[ws]?[space]?.metadata[path];

  AgentFileSpace spaceOf(String ws, String space) =>
      stores[ws]![space]!.space.deepCopy();

  AgentFileSpace _space(_Stored s, String ws) =>
      s.space.deepCopy()..workspaceId = ws;

  AgentFile _file(_Stored s, String path, String ws, {int? version}) {
    final versions = s.files[path]!;
    final v = version ?? versions.length;
    final content = versions[v - 1];
    return AgentFile(
      id: 'f-${path.hashCode}',
      spaceId: s.space.id,
      path: path,
      contentType: content.contentType,
      sizeBytes: Int64(utf8Length(content.content)),
      version: Int64(v),
      metadata: v == versions.length
          ? (s.metadata[path] ?? const {}).entries
          : <String, String>{}.entries,
      createdAt: Timestamp.fromDateTime(s.created[path] ?? DateTime.now()),
      updatedAt: Timestamp.fromDateTime(DateTime.now()),
      workspaceId: ws,
    );
  }

  void _answer(String method, HttpRequest r, dynamic message) {
    if (lostFileReplies.contains(method)) {
      _fail(r, 503, 'unavailable', 'connection lost');
    } else {
      write(r, message);
    }
  }

  bool _fail(HttpRequest r, int status, String code, String message) {
    return rpcFail(r, status, code, message);
  }

  /// Connect-style JSON error for fakes extended by other test servers.
  bool rpcFail(HttpRequest r, int status, String code, String message) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': message}));
    return true;
  }
}
