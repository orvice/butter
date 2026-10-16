// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace_memory.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class WorkspaceMemoryScope extends $pb.ProtobufEnum {
  /// Treated as WORKSPACE.
  static const WorkspaceMemoryScope WORKSPACE_MEMORY_SCOPE_UNSPECIFIED =
      WorkspaceMemoryScope._(
          0, _omitEnumNames ? '' : 'WORKSPACE_MEMORY_SCOPE_UNSPECIFIED');

  /// Workspace Memory: the pool shared across the workspace.
  static const WorkspaceMemoryScope WORKSPACE_MEMORY_SCOPE_WORKSPACE =
      WorkspaceMemoryScope._(
          1, _omitEnumNames ? '' : 'WORKSPACE_MEMORY_SCOPE_WORKSPACE');

  /// Agent Memory: private to one agent; requests name it with agent_id.
  static const WorkspaceMemoryScope WORKSPACE_MEMORY_SCOPE_AGENT =
      WorkspaceMemoryScope._(
          2, _omitEnumNames ? '' : 'WORKSPACE_MEMORY_SCOPE_AGENT');

  static const $core.List<WorkspaceMemoryScope> values = <WorkspaceMemoryScope>[
    WORKSPACE_MEMORY_SCOPE_UNSPECIFIED,
    WORKSPACE_MEMORY_SCOPE_WORKSPACE,
    WORKSPACE_MEMORY_SCOPE_AGENT,
  ];

  static final $core.List<WorkspaceMemoryScope?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static WorkspaceMemoryScope? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const WorkspaceMemoryScope._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
