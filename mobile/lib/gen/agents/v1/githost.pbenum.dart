// This is a generated file - do not edit.
//
// Generated from agents/v1/githost.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// GitHostKind selects the provider adapter used to talk to the host.
class GitHostKind extends $pb.ProtobufEnum {
  static const GitHostKind GIT_HOST_KIND_UNSPECIFIED =
      GitHostKind._(0, _omitEnumNames ? '' : 'GIT_HOST_KIND_UNSPECIFIED');

  /// GitHub.com or GitHub Enterprise (REST v3 API).
  static const GitHostKind GIT_HOST_KIND_GITHUB =
      GitHostKind._(1, _omitEnumNames ? '' : 'GIT_HOST_KIND_GITHUB');

  /// GitLab.com or self-hosted GitLab (REST v4 API).
  static const GitHostKind GIT_HOST_KIND_GITLAB =
      GitHostKind._(2, _omitEnumNames ? '' : 'GIT_HOST_KIND_GITLAB');

  static const $core.List<GitHostKind> values = <GitHostKind>[
    GIT_HOST_KIND_UNSPECIFIED,
    GIT_HOST_KIND_GITHUB,
    GIT_HOST_KIND_GITLAB,
  ];

  static final $core.List<GitHostKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static GitHostKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const GitHostKind._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
