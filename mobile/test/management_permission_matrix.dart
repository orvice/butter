/// Single source for workspace-role expectations exercised by
/// [management_permission_matrix_test.dart].
///
/// Server rules (verified in contract tests unless noted):
/// - Workspace metadata/members: owner or global admin write; any member read;
///   workspace **admin** membership alone does **not** grant workspace write.
/// - Many resource RPCs (Agents, Model providers, MCP, files, Cron, …): any
///   workspace member may write; the App does not add owner-only gates there.
/// - Owner, workspace admin, or global admin: Memory connection, Git binding,
///   Telegram/Linear, personal API token create/revoke, memory delete.
library;

enum WorkspaceRole { member, workspaceAdmin, owner }

enum GlobalRole { none, globalAdmin }

class PermissionExpectation {
  const PermissionExpectation({
    required this.label,
    required this.canManageWorkspace,
    required this.canManageWorkspaceResources,
  });

  final String label;
  final bool canManageWorkspace;
  final bool canManageWorkspaceResources;
}

/// Rows keyed by (workspace role, global role).
const managementPermissionMatrix =
    <(WorkspaceRole, GlobalRole), PermissionExpectation>{
      (WorkspaceRole.member, GlobalRole.none): PermissionExpectation(
        label: 'member',
        canManageWorkspace: false,
        canManageWorkspaceResources: false,
      ),
      (WorkspaceRole.workspaceAdmin, GlobalRole.none): PermissionExpectation(
        label: 'workspace admin',
        canManageWorkspace: false,
        canManageWorkspaceResources: true,
      ),
      (WorkspaceRole.owner, GlobalRole.none): PermissionExpectation(
        label: 'owner',
        canManageWorkspace: true,
        canManageWorkspaceResources: true,
      ),
      (WorkspaceRole.member, GlobalRole.globalAdmin): PermissionExpectation(
        label: 'global admin (member in workspace)',
        canManageWorkspace: true,
        canManageWorkspaceResources: true,
      ),
    };

/// UI affordances on [WorkspaceDetailPage] and hub pages that must follow the
/// same gates (widget tests).
class UiWriteControl {
  const UiWriteControl({
    required this.pageKey,
    required this.writeControlFinderLabel,
    required this.requiresResourcesGate,
    required this.requiresWorkspaceGate,
  });

  final String pageKey;
  final String writeControlFinderLabel;
  final bool requiresResourcesGate;
  final bool requiresWorkspaceGate;
}

const uiWriteControls = [
  UiWriteControl(
    pageKey: 'workspace-detail',
    writeControlFinderLabel: 'Edit Workspace',
    requiresResourcesGate: false,
    requiresWorkspaceGate: true,
  ),
  UiWriteControl(
    pageKey: 'api-tokens',
    writeControlFinderLabel: 'Create token',
    requiresResourcesGate: true,
    requiresWorkspaceGate: false,
  ),
];

/// Web global-admin routes that must never appear in the App shell (#469).
/// Widget coverage: `management_widget_test` (no Create user on profile).
const excludedGlobalAdminNavLabels = [
  'Git hosts',
  'Global MCP',
  'Platform users',
  'Create user',
  'Telegram endpoints',
  'Linear settings',
  'Admin',
];
