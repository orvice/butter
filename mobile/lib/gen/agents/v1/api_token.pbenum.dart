// This is a generated file - do not edit.
//
// Generated from agents/v1/api_token.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class APITokenKind extends $pb.ProtobufEnum {
  static const APITokenKind API_TOKEN_KIND_UNSPECIFIED =
      APITokenKind._(0, _omitEnumNames ? '' : 'API_TOKEN_KIND_UNSPECIFIED');
  static const APITokenKind API_TOKEN_KIND_USER =
      APITokenKind._(1, _omitEnumNames ? '' : 'API_TOKEN_KIND_USER');
  static const APITokenKind API_TOKEN_KIND_DAEMON =
      APITokenKind._(2, _omitEnumNames ? '' : 'API_TOKEN_KIND_DAEMON');

  static const $core.List<APITokenKind> values = <APITokenKind>[
    API_TOKEN_KIND_UNSPECIFIED,
    API_TOKEN_KIND_USER,
    API_TOKEN_KIND_DAEMON,
  ];

  static final $core.List<APITokenKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static APITokenKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const APITokenKind._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
