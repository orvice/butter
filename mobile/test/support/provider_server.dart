import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/workspace.pb.dart';

import 'server.dart';

/// Controllable external HTTP boundary for ModelProviderService. Like the real
/// service it keeps providers per Workspace header, replaces the whole object
/// on update and stamps `workspace_id`; the real semantics themselves are
/// proven against the production Go handlers in the contract tests.
class ProviderServer extends TestServer {
  final providers = <String, Map<String, ModelProvider>>{
    'alpha': {},
    'beta': {},
  };
  final calls = <({String method, String? workspace, List<int> body})>[];
  int get writes => calls.where((c) => _writes.contains(c.method)).length;
  static const _writes = {
    'CreateModelProvider',
    'UpdateModelProvider',
    'DeleteModelProvider',
  };

  final spaces = [
    Workspace(id: 'alpha', name: 'Alpha'),
    Workspace(id: 'beta', name: 'Beta'),
  ];

  /// Workspace ids whose GetWorkspace is answered 404 (management reads).
  final missingWorkspaces = <String>{};

  Future<void> Function()? beforeList;
  Future<void> Function()? beforeGet;
  Future<void> Function()? beforeWrite;

  /// A write answered with this (status, code, message) instead of being applied.
  (int, String, String)? refusal;

  /// A write that is applied but whose answer is lost (503).
  bool loseWriteReply = false;

  ProviderServer() {
    authorizedReply = (r, bytes) async {
      final method = r.uri.path.split('/').last;
      if (method == 'ListWorkspaces') {
        write(r, ListWorkspacesResponse(workspaces: spaces));
        return true;
      }
      if (method == 'GetWorkspace') {
        final id = GetWorkspaceRequest.fromBuffer(bytes).id;
        final matching = spaces.where((w) => w.id == id);
        if (missingWorkspaces.contains(id) || matching.isEmpty) {
          _fail(r, 404, 'not_found', 'workspace not found');
        } else {
          write(r, GetWorkspaceResponse(workspace: matching.first));
        }
        return true;
      }
      if (method == 'ListWorkspaceMembers') {
        final id = ListWorkspaceMembersRequest.fromBuffer(bytes).workspaceId;
        write(
          r,
          ListWorkspaceMembersResponse(
            members: [
              WorkspaceMember(workspaceId: id, userId: 'alice', role: 'member'),
            ],
          ),
        );
        return true;
      }
      if (!r.uri.path.contains('ModelProviderService')) return false;
      final workspace = r.headers.value('x-workspace-id');
      calls.add((method: method, workspace: workspace, body: bytes));
      final bucket = providers[workspace];
      if (bucket == null) {
        _fail(r, 403, 'permission_denied', 'workspace access denied');
        return true;
      }
      if (_writes.contains(method)) {
        await beforeWrite?.call();
        final refused = refusal;
        if (refused != null) {
          _fail(r, refused.$1, refused.$2, refused.$3);
          return true;
        }
      }
      switch (method) {
        case 'ListModelProviders':
          await beforeList?.call();
          write(
            r,
            ListModelProvidersResponse(modelProviders: bucket.values.toList()),
          );
        case 'GetModelProvider':
          await beforeGet?.call();
          final name = GetModelProviderRequest.fromBuffer(bytes).name;
          final found = bucket[name];
          if (found == null) {
            _fail(r, 404, 'not_found', 'model provider not found');
          } else {
            write(r, GetModelProviderResponse(modelProvider: found));
          }
        case 'CreateModelProvider':
          final incoming = CreateModelProviderRequest.fromBuffer(
            bytes,
          ).modelProvider;
          if (bucket.containsKey(incoming.name)) {
            _fail(r, 409, 'already_exists', 'model provider already exists');
          } else {
            final stored = incoming.deepCopy()..workspaceId = workspace!;
            bucket[stored.name] = stored;
            _answer(r, CreateModelProviderResponse(modelProvider: stored));
          }
        case 'UpdateModelProvider':
          final incoming = UpdateModelProviderRequest.fromBuffer(
            bytes,
          ).modelProvider;
          if (!bucket.containsKey(incoming.name)) {
            _fail(r, 404, 'not_found', 'model provider not found');
          } else {
            final stored = incoming.deepCopy()..workspaceId = workspace!;
            bucket[stored.name] = stored;
            _answer(r, UpdateModelProviderResponse(modelProvider: stored));
          }
        case 'DeleteModelProvider':
          final name = DeleteModelProviderRequest.fromBuffer(bytes).name;
          if (bucket.remove(name) == null) {
            _fail(r, 404, 'not_found', 'model provider not found');
          } else {
            _answer(r, DeleteModelProviderResponse());
          }
        default:
          return false;
      }
      return true;
    };
  }

  void _answer(HttpRequest r, dynamic message) {
    if (loseWriteReply) {
      _fail(r, 503, 'unavailable', 'connection lost');
    } else {
      write(r, message);
    }
  }

  void _fail(HttpRequest r, int status, String code, String message) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': message}));
  }

  /// The decoded body of the most recent call to [method].
  List<int> lastBody(String method) =>
      calls.lastWhere((c) => c.method == method).body;
}
