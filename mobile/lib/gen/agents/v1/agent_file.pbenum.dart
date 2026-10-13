// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_file.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class AgentFileMountPermission extends $pb.ProtobufEnum {
  static const AgentFileMountPermission
      AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED = AgentFileMountPermission._(
          0, _omitEnumNames ? '' : 'AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED');
  static const AgentFileMountPermission AGENT_FILE_MOUNT_PERMISSION_READ =
      AgentFileMountPermission._(
          1, _omitEnumNames ? '' : 'AGENT_FILE_MOUNT_PERMISSION_READ');
  static const AgentFileMountPermission AGENT_FILE_MOUNT_PERMISSION_READ_WRITE =
      AgentFileMountPermission._(
          2, _omitEnumNames ? '' : 'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE');
  static const AgentFileMountPermission
      AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE =
      AgentFileMountPermission._(
          3,
          _omitEnumNames
              ? ''
              : 'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE');

  static const $core.List<AgentFileMountPermission> values =
      <AgentFileMountPermission>[
    AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED,
    AGENT_FILE_MOUNT_PERMISSION_READ,
    AGENT_FILE_MOUNT_PERMISSION_READ_WRITE,
    AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE,
  ];

  static final $core.List<AgentFileMountPermission?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static AgentFileMountPermission? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentFileMountPermission._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
