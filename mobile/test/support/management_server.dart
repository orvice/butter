import 'dart:io';
import 'dart:convert';

import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:butter_mobile/gen/agents/v1/workspace.pb.dart';

import 'server.dart';

class ManagementServer extends TestServer {
  String role = 'owner';
  bool rejectWrites = false;
  Future<void> Function()? beforeMembers;
  Future<void> Function()? beforeProfile;
  final spaces = <Workspace>[
    Workspace(
      id: 'alpha',
      name: 'Alpha',
      slug: 'alpha',
      description: 'Original',
    ),
    Workspace(id: 'beta', name: 'Beta', slug: 'beta'),
  ];
  final members = <WorkspaceMember>[];
  User profile = User(
    id: 'alice',
    username: 'alice',
    displayName: 'Alice',
    avatarUrl: 'https://cdn.example/original.png',
  );
  UpdateProfileRequest? lastProfile;
  String? uploadBody;
  String? uploadType;
  int writes = 0;

  ManagementServer() {
    authorizedReply = (r, bytes) async {
      final method = r.uri.path.split('/').last;
      if (method == 'Me') {
        write(r, MeResponse(user: profile));
      } else if (method == 'ListWorkspaces') {
        write(r, ListWorkspacesResponse(workspaces: spaces));
      } else if (method == 'GetWorkspace') {
        final input = GetWorkspaceRequest.fromBuffer(bytes);
        final matching = spaces.where((w) => w.id == input.id);
        if (matching.isEmpty) {
          error(r, 404, 'not_found');
          return true;
        }
        write(r, GetWorkspaceResponse(workspace: matching.first));
      } else if (method == 'ListWorkspaceMembers') {
        final input = ListWorkspaceMembersRequest.fromBuffer(bytes);
        final snapshot = [
          WorkspaceMember(
            workspaceId: input.workspaceId,
            userId: 'alice',
            role: role,
          ),
          ...members,
        ];
        await beforeMembers?.call();
        write(r, ListWorkspaceMembersResponse(members: snapshot));
      } else if (method == 'CreateWorkspace') {
        writes++;
        if (rejectWrites) {
          error(r, 503, 'unavailable');
          return true;
        }
        final space = CreateWorkspaceRequest.fromBuffer(bytes).workspace
          ..id = 'created';
        spaces.add(space);
        write(r, CreateWorkspaceResponse(workspace: space));
      } else if ([
        'UpdateWorkspace',
        'DeleteWorkspace',
        'AddWorkspaceMember',
        'UpdateWorkspaceMember',
        'RemoveWorkspaceMember',
      ].contains(method)) {
        writes++;
        if (role != 'owner' || rejectWrites) {
          error(r, 403, 'permission_denied');
          return true;
        }
        switch (method) {
          case 'UpdateWorkspace':
            final space = UpdateWorkspaceRequest.fromBuffer(bytes).workspace;
            spaces[spaces.indexWhere((w) => w.id == space.id)] = space;
            write(r, UpdateWorkspaceResponse(workspace: space));
          case 'DeleteWorkspace':
            final input = DeleteWorkspaceRequest.fromBuffer(bytes);
            spaces.removeWhere((w) => w.id == input.id);
            write(r, DeleteWorkspaceResponse());
          case 'AddWorkspaceMember':
            final input = AddWorkspaceMemberRequest.fromBuffer(bytes);
            final member = WorkspaceMember(
              workspaceId: input.workspaceId,
              userId: input.userId,
              role: input.role,
            );
            members.add(member);
            write(r, AddWorkspaceMemberResponse(member: member));
          case 'UpdateWorkspaceMember':
            final input = UpdateWorkspaceMemberRequest.fromBuffer(bytes);
            final member = members.firstWhere((m) => m.userId == input.userId)
              ..role = input.role;
            write(r, UpdateWorkspaceMemberResponse(member: member));
          case 'RemoveWorkspaceMember':
            final input = RemoveWorkspaceMemberRequest.fromBuffer(bytes);
            members.removeWhere((m) => m.userId == input.userId);
            write(r, RemoveWorkspaceMemberResponse());
        }
      } else if (method == 'UpdateProfile') {
        writes++;
        if (rejectWrites) {
          error(r, 403, 'permission_denied');
          return true;
        }
        await beforeProfile?.call();
        lastProfile = UpdateProfileRequest.fromBuffer(bytes);
        profile.displayName = lastProfile!.displayName;
        if (lastProfile!.hasAvatarUrl()) {
          profile.avatarUrl = lastProfile!.avatarUrl;
        }
        write(r, UpdateProfileResponse(user: profile));
      } else if (method == 'ChangePassword') {
        writes++;
        final input = ChangePasswordRequest.fromBuffer(bytes);
        if (input.currentPassword != 'password') {
          error(r, 403, 'permission_denied');
          return true;
        }
        write(r, ChangePasswordResponse(user: profile));
      } else if (method == 'avatar') {
        writes++;
        uploadBody = latin1.decode(bytes);
        uploadType = r.headers.contentType.toString();
        if (rejectWrites) {
          error(r, 503, 'unavailable');
          return true;
        }
        r.response.headers.contentType = ContentType.json;
        r.response.write('{"url":"https://cdn.example/new.png"}');
      } else {
        return false;
      }
      return true;
    };
  }
}
