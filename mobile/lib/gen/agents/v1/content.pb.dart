// This is a generated file - do not edit.
//
// Generated from agents/v1/content.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

enum InputPart_Part { text, inlineData, notSet }

/// InputPart is one piece of multimodal user input. A request's `parts`
/// list is ordered; text and images may interleave. The oneof is
/// intentionally extensible: a file-reference variant can be added later
/// without breaking existing clients.
class InputPart extends $pb.GeneratedMessage {
  factory InputPart({
    $core.String? text,
    InlineData? inlineData,
  }) {
    final result = InputPart._();
    if (text != null) result.text = text;
    if (inlineData != null) result.inlineData = inlineData;
    return result;
  }

  InputPart._();

  factory InputPart.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InputPart()..mergeFromBuffer(data, registry);
  factory InputPart.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InputPart()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, InputPart_Part> _InputPart_PartByTag = {
    1: InputPart_Part.text,
    2: InputPart_Part.inlineData,
    0: InputPart_Part.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InputPart',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: InputPart.$_createMessage)
    ..oo(0, [1, 2])
    ..aOS(1, _omitFieldNames ? '' : 'text')
    ..aOM<InlineData>(2, _omitFieldNames ? '' : 'inlineData',
        subBuilder: InlineData.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InputPart clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InputPart copyWith(void Function(InputPart) updates) =>
      super.copyWith((message) => updates(message as InputPart)) as InputPart;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use InputPart() / InputPart.new instead')
  static InputPart create() => InputPart._();
  static $pb.GeneratedMessage $_createMessage() => InputPart._();
  @$core.override
  InputPart createEmptyInstance() => InputPart._();
  @$core.pragma('dart2js:noInline')
  static InputPart getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InputPart>(InputPart.$_createMessage);
  static InputPart? _defaultInstance;

  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  InputPart_Part whichPart() => _InputPart_PartByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  void clearPart() => $_clearField($_whichOneof(0));

  /// Plain text content.
  @$pb.TagNumber(1)
  $core.String get text => $_getSZ(0);
  @$pb.TagNumber(1)
  set text($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasText() => $_has(0);
  @$pb.TagNumber(1)
  void clearText() => $_clearField(1);

  /// Inline binary content (images).
  @$pb.TagNumber(2)
  InlineData get inlineData => $_getN(1);
  @$pb.TagNumber(2)
  set inlineData(InlineData value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasInlineData() => $_has(1);
  @$pb.TagNumber(2)
  void clearInlineData() => $_clearField(2);
  @$pb.TagNumber(2)
  InlineData ensureInlineData() => $_ensure(1);
}

/// InlineData carries raw bytes plus their MIME type. The accepted MIME
/// whitelist and size limits are enforced by the application layer, not
/// by schema constraints.
class InlineData extends $pb.GeneratedMessage {
  factory InlineData({
    $core.String? mimeType,
    $core.List<$core.int>? data,
  }) {
    final result = InlineData._();
    if (mimeType != null) result.mimeType = mimeType;
    if (data != null) result.data = data;
    return result;
  }

  InlineData._();

  factory InlineData.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InlineData()..mergeFromBuffer(data, registry);
  factory InlineData.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InlineData()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InlineData',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: InlineData.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'mimeType')
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InlineData clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InlineData copyWith(void Function(InlineData) updates) =>
      super.copyWith((message) => updates(message as InlineData)) as InlineData;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use InlineData() / InlineData.new instead')
  static InlineData create() => InlineData._();
  static $pb.GeneratedMessage $_createMessage() => InlineData._();
  @$core.override
  InlineData createEmptyInstance() => InlineData._();
  @$core.pragma('dart2js:noInline')
  static InlineData getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InlineData>(InlineData.$_createMessage);
  static InlineData? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get mimeType => $_getSZ(0);
  @$pb.TagNumber(1)
  set mimeType($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMimeType() => $_has(0);
  @$pb.TagNumber(1)
  void clearMimeType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get data => $_getN(1);
  @$pb.TagNumber(2)
  set data($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(2)
  void clearData() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
