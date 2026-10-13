// This is a generated file - do not edit.
//
// Generated from agents/v1/content.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use inputPartDescriptor instead')
const InputPart$json = {
  '1': 'InputPart',
  '2': [
    {'1': 'text', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'text'},
    {
      '1': 'inline_data',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.InlineData',
      '9': 0,
      '10': 'inlineData'
    },
  ],
  '8': [
    {'1': 'part'},
  ],
};

/// Descriptor for `InputPart`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List inputPartDescriptor = $convert.base64Decode(
    'CglJbnB1dFBhcnQSFAoEdGV4dBgBIAEoCUgAUgR0ZXh0EjgKC2lubGluZV9kYXRhGAIgASgLMh'
    'UuYWdlbnRzLnYxLklubGluZURhdGFIAFIKaW5saW5lRGF0YUIGCgRwYXJ0');

@$core.Deprecated('Use inlineDataDescriptor instead')
const InlineData$json = {
  '1': 'InlineData',
  '2': [
    {'1': 'mime_type', '3': 1, '4': 1, '5': 9, '10': 'mimeType'},
    {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `InlineData`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List inlineDataDescriptor = $convert.base64Decode(
    'CgpJbmxpbmVEYXRhEhsKCW1pbWVfdHlwZRgBIAEoCVIIbWltZVR5cGUSEgoEZGF0YRgCIAEoDF'
    'IEZGF0YQ==');
