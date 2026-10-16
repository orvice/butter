// This is a generated file - do not edit.
//
// Generated from agents/v1/automation.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as $1;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import 'automation.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'automation.pbenum.dart';

/// Automation defines a workspace-scoped ordered workflow.
class Automation extends $pb.GeneratedMessage {
  factory Automation({
    $core.String? name,
    $core.bool? enabled,
    AutomationTrigger? trigger,
    $core.Iterable<AutomationCondition>? conditions,
    $core.Iterable<AutomationStep>? steps,
    AutomationPolicy? policy,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = Automation._();
    if (name != null) result.name = name;
    if (enabled != null) result.enabled = enabled;
    if (trigger != null) result.trigger = trigger;
    if (conditions != null) result.conditions.addAll(conditions);
    if (steps != null) result.steps.addAll(steps);
    if (policy != null) result.policy = policy;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  Automation._();

  factory Automation.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Automation()..mergeFromBuffer(data, registry);
  factory Automation.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Automation()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Automation',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: Automation.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..aOM<AutomationTrigger>(3, _omitFieldNames ? '' : 'trigger',
        subBuilder: AutomationTrigger.$_createMessage)
    ..pPM<AutomationCondition>(4, _omitFieldNames ? '' : 'conditions',
        subBuilder: AutomationCondition.$_createMessage)
    ..pPM<AutomationStep>(5, _omitFieldNames ? '' : 'steps',
        subBuilder: AutomationStep.$_createMessage)
    ..aOM<AutomationPolicy>(6, _omitFieldNames ? '' : 'policy',
        subBuilder: AutomationPolicy.$_createMessage)
    ..m<$core.String, $core.String>(20, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'Automation.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOM<$0.Timestamp>(30, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(31, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Automation clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Automation copyWith(void Function(Automation) updates) =>
      super.copyWith((message) => updates(message as Automation)) as Automation;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Automation() / Automation.new instead')
  static Automation create() => Automation._();
  static $pb.GeneratedMessage $_createMessage() => Automation._();
  @$core.override
  Automation createEmptyInstance() => Automation._();
  @$core.pragma('dart2js:noInline')
  static Automation getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Automation>(Automation.$_createMessage);
  static Automation? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty" bson:"enabled,omitempty"
  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  /// @gotags: json:"trigger,omitempty" yaml:"trigger,omitempty" bson:"trigger,omitempty"
  @$pb.TagNumber(3)
  AutomationTrigger get trigger => $_getN(2);
  @$pb.TagNumber(3)
  set trigger(AutomationTrigger value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasTrigger() => $_has(2);
  @$pb.TagNumber(3)
  void clearTrigger() => $_clearField(3);
  @$pb.TagNumber(3)
  AutomationTrigger ensureTrigger() => $_ensure(2);

  /// @gotags: json:"conditions,omitempty" yaml:"conditions,omitempty" bson:"conditions,omitempty"
  @$pb.TagNumber(4)
  $pb.PbList<AutomationCondition> get conditions => $_getList(3);

  /// @gotags: json:"steps,omitempty" yaml:"steps,omitempty" bson:"steps,omitempty"
  @$pb.TagNumber(5)
  $pb.PbList<AutomationStep> get steps => $_getList(4);

  /// @gotags: json:"policy,omitempty" yaml:"policy,omitempty" bson:"policy,omitempty"
  @$pb.TagNumber(6)
  AutomationPolicy get policy => $_getN(5);
  @$pb.TagNumber(6)
  set policy(AutomationPolicy value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPolicy() => $_has(5);
  @$pb.TagNumber(6)
  void clearPolicy() => $_clearField(6);
  @$pb.TagNumber(6)
  AutomationPolicy ensurePolicy() => $_ensure(5);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty" bson:"metadata,omitempty"
  @$pb.TagNumber(20)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(6);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(30)
  $0.Timestamp get createdAt => $_getN(7);
  @$pb.TagNumber(30)
  set createdAt($0.Timestamp value) => $_setField(30, value);
  @$pb.TagNumber(30)
  $core.bool hasCreatedAt() => $_has(7);
  @$pb.TagNumber(30)
  void clearCreatedAt() => $_clearField(30);
  @$pb.TagNumber(30)
  $0.Timestamp ensureCreatedAt() => $_ensure(7);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(31)
  $0.Timestamp get updatedAt => $_getN(8);
  @$pb.TagNumber(31)
  set updatedAt($0.Timestamp value) => $_setField(31, value);
  @$pb.TagNumber(31)
  $core.bool hasUpdatedAt() => $_has(8);
  @$pb.TagNumber(31)
  void clearUpdatedAt() => $_clearField(31);
  @$pb.TagNumber(31)
  $0.Timestamp ensureUpdatedAt() => $_ensure(8);

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(9);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(9);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class AutomationTrigger extends $pb.GeneratedMessage {
  factory AutomationTrigger({
    AutomationTriggerType? type,
    AutomationScheduleTrigger? schedule,
    AutomationWebhookTrigger? webhook,
    AutomationEventTrigger? event,
  }) {
    final result = AutomationTrigger._();
    if (type != null) result.type = type;
    if (schedule != null) result.schedule = schedule;
    if (webhook != null) result.webhook = webhook;
    if (event != null) result.event = event;
    return result;
  }

  AutomationTrigger._();

  factory AutomationTrigger.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationTrigger()..mergeFromBuffer(data, registry);
  factory AutomationTrigger.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationTrigger()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationTrigger',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationTrigger.$_createMessage)
    ..aE<AutomationTriggerType>(1, _omitFieldNames ? '' : 'type',
        enumValues: AutomationTriggerType.values)
    ..aOM<AutomationScheduleTrigger>(2, _omitFieldNames ? '' : 'schedule',
        subBuilder: AutomationScheduleTrigger.$_createMessage)
    ..aOM<AutomationWebhookTrigger>(3, _omitFieldNames ? '' : 'webhook',
        subBuilder: AutomationWebhookTrigger.$_createMessage)
    ..aOM<AutomationEventTrigger>(4, _omitFieldNames ? '' : 'event',
        subBuilder: AutomationEventTrigger.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationTrigger clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationTrigger copyWith(void Function(AutomationTrigger) updates) =>
      super.copyWith((message) => updates(message as AutomationTrigger))
          as AutomationTrigger;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AutomationTrigger() / AutomationTrigger.new instead')
  static AutomationTrigger create() => AutomationTrigger._();
  static $pb.GeneratedMessage $_createMessage() => AutomationTrigger._();
  @$core.override
  AutomationTrigger createEmptyInstance() => AutomationTrigger._();
  @$core.pragma('dart2js:noInline')
  static AutomationTrigger getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AutomationTrigger>(
          AutomationTrigger.$_createMessage);
  static AutomationTrigger? _defaultInstance;

  /// @gotags: json:"type,omitempty" yaml:"type,omitempty" bson:"type,omitempty"
  @$pb.TagNumber(1)
  AutomationTriggerType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(AutomationTriggerType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  /// @gotags: json:"schedule,omitempty" yaml:"schedule,omitempty" bson:"schedule,omitempty"
  @$pb.TagNumber(2)
  AutomationScheduleTrigger get schedule => $_getN(1);
  @$pb.TagNumber(2)
  set schedule(AutomationScheduleTrigger value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasSchedule() => $_has(1);
  @$pb.TagNumber(2)
  void clearSchedule() => $_clearField(2);
  @$pb.TagNumber(2)
  AutomationScheduleTrigger ensureSchedule() => $_ensure(1);

  /// @gotags: json:"webhook,omitempty" yaml:"webhook,omitempty" bson:"webhook,omitempty"
  @$pb.TagNumber(3)
  AutomationWebhookTrigger get webhook => $_getN(2);
  @$pb.TagNumber(3)
  set webhook(AutomationWebhookTrigger value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasWebhook() => $_has(2);
  @$pb.TagNumber(3)
  void clearWebhook() => $_clearField(3);
  @$pb.TagNumber(3)
  AutomationWebhookTrigger ensureWebhook() => $_ensure(2);

  /// Future event trigger configuration. The initial runtime only executes
  /// MANUAL and SCHEDULE triggers, but these fields keep the resource shape
  /// stable for event-triggered workflows.
  /// @gotags: json:"event,omitempty" yaml:"event,omitempty" bson:"event,omitempty"
  @$pb.TagNumber(4)
  AutomationEventTrigger get event => $_getN(3);
  @$pb.TagNumber(4)
  set event(AutomationEventTrigger value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasEvent() => $_has(3);
  @$pb.TagNumber(4)
  void clearEvent() => $_clearField(4);
  @$pb.TagNumber(4)
  AutomationEventTrigger ensureEvent() => $_ensure(3);
}

class AutomationScheduleTrigger extends $pb.GeneratedMessage {
  factory AutomationScheduleTrigger({
    $core.String? schedule,
    $core.String? timezone,
  }) {
    final result = AutomationScheduleTrigger._();
    if (schedule != null) result.schedule = schedule;
    if (timezone != null) result.timezone = timezone;
    return result;
  }

  AutomationScheduleTrigger._();

  factory AutomationScheduleTrigger.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationScheduleTrigger()..mergeFromBuffer(data, registry);
  factory AutomationScheduleTrigger.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationScheduleTrigger()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationScheduleTrigger',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationScheduleTrigger.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'schedule')
    ..aOS(2, _omitFieldNames ? '' : 'timezone')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationScheduleTrigger clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationScheduleTrigger copyWith(
          void Function(AutomationScheduleTrigger) updates) =>
      super.copyWith((message) => updates(message as AutomationScheduleTrigger))
          as AutomationScheduleTrigger;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationScheduleTrigger() / AutomationScheduleTrigger.new instead')
  static AutomationScheduleTrigger create() => AutomationScheduleTrigger._();
  static $pb.GeneratedMessage $_createMessage() =>
      AutomationScheduleTrigger._();
  @$core.override
  AutomationScheduleTrigger createEmptyInstance() =>
      AutomationScheduleTrigger._();
  @$core.pragma('dart2js:noInline')
  static AutomationScheduleTrigger getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationScheduleTrigger>(
          AutomationScheduleTrigger.$_createMessage);
  static AutomationScheduleTrigger? _defaultInstance;

  /// @gotags: json:"schedule,omitempty" yaml:"schedule,omitempty" bson:"schedule,omitempty"
  @$pb.TagNumber(1)
  $core.String get schedule => $_getSZ(0);
  @$pb.TagNumber(1)
  set schedule($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSchedule() => $_has(0);
  @$pb.TagNumber(1)
  void clearSchedule() => $_clearField(1);

  /// IANA timezone name. Defaults to UTC when empty.
  /// @gotags: json:"timezone,omitempty" yaml:"timezone,omitempty" bson:"timezone,omitempty"
  @$pb.TagNumber(2)
  $core.String get timezone => $_getSZ(1);
  @$pb.TagNumber(2)
  set timezone($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTimezone() => $_has(1);
  @$pb.TagNumber(2)
  void clearTimezone() => $_clearField(2);
}

class AutomationWebhookTrigger extends $pb.GeneratedMessage {
  factory AutomationWebhookTrigger({
    $core.String? token,
  }) {
    final result = AutomationWebhookTrigger._();
    if (token != null) result.token = token;
    return result;
  }

  AutomationWebhookTrigger._();

  factory AutomationWebhookTrigger.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationWebhookTrigger()..mergeFromBuffer(data, registry);
  factory AutomationWebhookTrigger.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationWebhookTrigger()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationWebhookTrigger',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationWebhookTrigger.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationWebhookTrigger clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationWebhookTrigger copyWith(
          void Function(AutomationWebhookTrigger) updates) =>
      super.copyWith((message) => updates(message as AutomationWebhookTrigger))
          as AutomationWebhookTrigger;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationWebhookTrigger() / AutomationWebhookTrigger.new instead')
  static AutomationWebhookTrigger create() => AutomationWebhookTrigger._();
  static $pb.GeneratedMessage $_createMessage() => AutomationWebhookTrigger._();
  @$core.override
  AutomationWebhookTrigger createEmptyInstance() =>
      AutomationWebhookTrigger._();
  @$core.pragma('dart2js:noInline')
  static AutomationWebhookTrigger getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationWebhookTrigger>(
          AutomationWebhookTrigger.$_createMessage);
  static AutomationWebhookTrigger? _defaultInstance;

  /// Stable path token used by future webhook trigger routing.
  /// @gotags: json:"token,omitempty" yaml:"token,omitempty" bson:"token,omitempty"
  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);
}

class AutomationEventTrigger extends $pb.GeneratedMessage {
  factory AutomationEventTrigger({
    $core.String? eventType,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? filters,
  }) {
    final result = AutomationEventTrigger._();
    if (eventType != null) result.eventType = eventType;
    if (filters != null) result.filters.addEntries(filters);
    return result;
  }

  AutomationEventTrigger._();

  factory AutomationEventTrigger.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationEventTrigger()..mergeFromBuffer(data, registry);
  factory AutomationEventTrigger.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationEventTrigger()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationEventTrigger',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationEventTrigger.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'eventType')
    ..m<$core.String, $core.String>(2, _omitFieldNames ? '' : 'filters',
        entryClassName: 'AutomationEventTrigger.FiltersEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationEventTrigger clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationEventTrigger copyWith(
          void Function(AutomationEventTrigger) updates) =>
      super.copyWith((message) => updates(message as AutomationEventTrigger))
          as AutomationEventTrigger;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationEventTrigger() / AutomationEventTrigger.new instead')
  static AutomationEventTrigger create() => AutomationEventTrigger._();
  static $pb.GeneratedMessage $_createMessage() => AutomationEventTrigger._();
  @$core.override
  AutomationEventTrigger createEmptyInstance() => AutomationEventTrigger._();
  @$core.pragma('dart2js:noInline')
  static AutomationEventTrigger getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationEventTrigger>(
          AutomationEventTrigger.$_createMessage);
  static AutomationEventTrigger? _defaultInstance;

  /// Event kind, e.g. "forum.post.created" or "daemon.task.failed".
  /// @gotags: json:"event_type,omitempty" yaml:"event_type,omitempty" bson:"event_type,omitempty"
  @$pb.TagNumber(1)
  $core.String get eventType => $_getSZ(0);
  @$pb.TagNumber(1)
  set eventType($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEventType() => $_has(0);
  @$pb.TagNumber(1)
  void clearEventType() => $_clearField(1);

  /// @gotags: json:"filters,omitempty" yaml:"filters,omitempty" bson:"filters,omitempty"
  @$pb.TagNumber(2)
  $pb.PbMap<$core.String, $core.String> get filters => $_getMap(1);
}

class AutomationCondition extends $pb.GeneratedMessage {
  factory AutomationCondition({
    $core.String? selector,
    AutomationConditionOperator? operator,
    $core.String? value,
  }) {
    final result = AutomationCondition._();
    if (selector != null) result.selector = selector;
    if (operator != null) result.operator = operator;
    if (value != null) result.value = value;
    return result;
  }

  AutomationCondition._();

  factory AutomationCondition.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationCondition()..mergeFromBuffer(data, registry);
  factory AutomationCondition.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationCondition()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationCondition',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationCondition.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'selector')
    ..aE<AutomationConditionOperator>(2, _omitFieldNames ? '' : 'operator',
        enumValues: AutomationConditionOperator.values)
    ..aOS(3, _omitFieldNames ? '' : 'value')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationCondition clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationCondition copyWith(void Function(AutomationCondition) updates) =>
      super.copyWith((message) => updates(message as AutomationCondition))
          as AutomationCondition;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use AutomationCondition() / AutomationCondition.new instead')
  static AutomationCondition create() => AutomationCondition._();
  static $pb.GeneratedMessage $_createMessage() => AutomationCondition._();
  @$core.override
  AutomationCondition createEmptyInstance() => AutomationCondition._();
  @$core.pragma('dart2js:noInline')
  static AutomationCondition getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationCondition>(
          AutomationCondition.$_createMessage);
  static AutomationCondition? _defaultInstance;

  /// @gotags: json:"selector,omitempty" yaml:"selector,omitempty" bson:"selector,omitempty"
  @$pb.TagNumber(1)
  $core.String get selector => $_getSZ(0);
  @$pb.TagNumber(1)
  set selector($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSelector() => $_has(0);
  @$pb.TagNumber(1)
  void clearSelector() => $_clearField(1);

  /// @gotags: json:"operator,omitempty" yaml:"operator,omitempty" bson:"operator,omitempty"
  @$pb.TagNumber(2)
  AutomationConditionOperator get operator => $_getN(1);
  @$pb.TagNumber(2)
  set operator(AutomationConditionOperator value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOperator() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperator() => $_clearField(2);

  /// @gotags: json:"value,omitempty" yaml:"value,omitempty" bson:"value,omitempty"
  @$pb.TagNumber(3)
  $core.String get value => $_getSZ(2);
  @$pb.TagNumber(3)
  set value($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasValue() => $_has(2);
  @$pb.TagNumber(3)
  void clearValue() => $_clearField(3);
}

class AutomationStep extends $pb.GeneratedMessage {
  factory AutomationStep({
    $core.String? name,
    AutomationStepType? type,
    AutomationInvokeAgentStep? invokeAgent,
    AutomationCallWebhookStep? callWebhook,
    AutomationSendNotifyGroupStep? sendNotifyGroup,
    AutomationCreateForumPostStep? createForumPost,
    AutomationPolicy? policy,
  }) {
    final result = AutomationStep._();
    if (name != null) result.name = name;
    if (type != null) result.type = type;
    if (invokeAgent != null) result.invokeAgent = invokeAgent;
    if (callWebhook != null) result.callWebhook = callWebhook;
    if (sendNotifyGroup != null) result.sendNotifyGroup = sendNotifyGroup;
    if (createForumPost != null) result.createForumPost = createForumPost;
    if (policy != null) result.policy = policy;
    return result;
  }

  AutomationStep._();

  factory AutomationStep.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationStep()..mergeFromBuffer(data, registry);
  factory AutomationStep.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationStep()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationStep',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationStep.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aE<AutomationStepType>(2, _omitFieldNames ? '' : 'type',
        enumValues: AutomationStepType.values)
    ..aOM<AutomationInvokeAgentStep>(3, _omitFieldNames ? '' : 'invokeAgent',
        subBuilder: AutomationInvokeAgentStep.$_createMessage)
    ..aOM<AutomationCallWebhookStep>(4, _omitFieldNames ? '' : 'callWebhook',
        subBuilder: AutomationCallWebhookStep.$_createMessage)
    ..aOM<AutomationSendNotifyGroupStep>(
        5, _omitFieldNames ? '' : 'sendNotifyGroup',
        subBuilder: AutomationSendNotifyGroupStep.$_createMessage)
    ..aOM<AutomationCreateForumPostStep>(
        6, _omitFieldNames ? '' : 'createForumPost',
        subBuilder: AutomationCreateForumPostStep.$_createMessage)
    ..aOM<AutomationPolicy>(10, _omitFieldNames ? '' : 'policy',
        subBuilder: AutomationPolicy.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationStep clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationStep copyWith(void Function(AutomationStep) updates) =>
      super.copyWith((message) => updates(message as AutomationStep))
          as AutomationStep;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AutomationStep() / AutomationStep.new instead')
  static AutomationStep create() => AutomationStep._();
  static $pb.GeneratedMessage $_createMessage() => AutomationStep._();
  @$core.override
  AutomationStep createEmptyInstance() => AutomationStep._();
  @$core.pragma('dart2js:noInline')
  static AutomationStep getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AutomationStep>(
          AutomationStep.$_createMessage);
  static AutomationStep? _defaultInstance;

  /// Stable name used in step-run records.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// @gotags: json:"type,omitempty" yaml:"type,omitempty" bson:"type,omitempty"
  @$pb.TagNumber(2)
  AutomationStepType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(AutomationStepType value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  /// @gotags: json:"invoke_agent,omitempty" yaml:"invoke_agent,omitempty" bson:"invoke_agent,omitempty"
  @$pb.TagNumber(3)
  AutomationInvokeAgentStep get invokeAgent => $_getN(2);
  @$pb.TagNumber(3)
  set invokeAgent(AutomationInvokeAgentStep value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasInvokeAgent() => $_has(2);
  @$pb.TagNumber(3)
  void clearInvokeAgent() => $_clearField(3);
  @$pb.TagNumber(3)
  AutomationInvokeAgentStep ensureInvokeAgent() => $_ensure(2);

  /// @gotags: json:"call_webhook,omitempty" yaml:"call_webhook,omitempty" bson:"call_webhook,omitempty"
  @$pb.TagNumber(4)
  AutomationCallWebhookStep get callWebhook => $_getN(3);
  @$pb.TagNumber(4)
  set callWebhook(AutomationCallWebhookStep value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasCallWebhook() => $_has(3);
  @$pb.TagNumber(4)
  void clearCallWebhook() => $_clearField(4);
  @$pb.TagNumber(4)
  AutomationCallWebhookStep ensureCallWebhook() => $_ensure(3);

  /// @gotags: json:"send_notify_group,omitempty" yaml:"send_notify_group,omitempty" bson:"send_notify_group,omitempty"
  @$pb.TagNumber(5)
  AutomationSendNotifyGroupStep get sendNotifyGroup => $_getN(4);
  @$pb.TagNumber(5)
  set sendNotifyGroup(AutomationSendNotifyGroupStep value) =>
      $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasSendNotifyGroup() => $_has(4);
  @$pb.TagNumber(5)
  void clearSendNotifyGroup() => $_clearField(5);
  @$pb.TagNumber(5)
  AutomationSendNotifyGroupStep ensureSendNotifyGroup() => $_ensure(4);

  /// @gotags: json:"create_forum_post,omitempty" yaml:"create_forum_post,omitempty" bson:"create_forum_post,omitempty"
  @$pb.TagNumber(6)
  AutomationCreateForumPostStep get createForumPost => $_getN(5);
  @$pb.TagNumber(6)
  set createForumPost(AutomationCreateForumPostStep value) =>
      $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCreateForumPost() => $_has(5);
  @$pb.TagNumber(6)
  void clearCreateForumPost() => $_clearField(6);
  @$pb.TagNumber(6)
  AutomationCreateForumPostStep ensureCreateForumPost() => $_ensure(5);

  /// Optional per-step overrides for retry/timeout/output policy.
  /// @gotags: json:"policy,omitempty" yaml:"policy,omitempty" bson:"policy,omitempty"
  @$pb.TagNumber(10)
  AutomationPolicy get policy => $_getN(6);
  @$pb.TagNumber(10)
  set policy(AutomationPolicy value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasPolicy() => $_has(6);
  @$pb.TagNumber(10)
  void clearPolicy() => $_clearField(10);
  @$pb.TagNumber(10)
  AutomationPolicy ensurePolicy() => $_ensure(6);
}

class AutomationInvokeAgentStep extends $pb.GeneratedMessage {
  factory AutomationInvokeAgentStep({
    $core.String? agentName,
    $core.String? input,
    $core.String? modelOverride,
    $core.String? agentId,
  }) {
    final result = AutomationInvokeAgentStep._();
    if (agentName != null) result.agentName = agentName;
    if (input != null) result.input = input;
    if (modelOverride != null) result.modelOverride = modelOverride;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  AutomationInvokeAgentStep._();

  factory AutomationInvokeAgentStep.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationInvokeAgentStep()..mergeFromBuffer(data, registry);
  factory AutomationInvokeAgentStep.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationInvokeAgentStep()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationInvokeAgentStep',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationInvokeAgentStep.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentName')
    ..aOS(2, _omitFieldNames ? '' : 'input')
    ..aOS(3, _omitFieldNames ? '' : 'modelOverride')
    ..aOS(4, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationInvokeAgentStep clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationInvokeAgentStep copyWith(
          void Function(AutomationInvokeAgentStep) updates) =>
      super.copyWith((message) => updates(message as AutomationInvokeAgentStep))
          as AutomationInvokeAgentStep;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationInvokeAgentStep() / AutomationInvokeAgentStep.new instead')
  static AutomationInvokeAgentStep create() => AutomationInvokeAgentStep._();
  static $pb.GeneratedMessage $_createMessage() =>
      AutomationInvokeAgentStep._();
  @$core.override
  AutomationInvokeAgentStep createEmptyInstance() =>
      AutomationInvokeAgentStep._();
  @$core.pragma('dart2js:noInline')
  static AutomationInvokeAgentStep getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationInvokeAgentStep>(
          AutomationInvokeAgentStep.$_createMessage);
  static AutomationInvokeAgentStep? _defaultInstance;

  /// Legacy agent name. Deprecated: use agent_id. Retained so historical
  /// records stay readable; the service layer backfills agent_id.
  /// @gotags: json:"agent_name,omitempty" yaml:"agent_name,omitempty" bson:"agent_name,omitempty"
  @$pb.TagNumber(1)
  $core.String get agentName => $_getSZ(0);
  @$pb.TagNumber(1)
  set agentName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentName() => $_clearField(1);

  /// @gotags: json:"input,omitempty" yaml:"input,omitempty" bson:"input,omitempty"
  @$pb.TagNumber(2)
  $core.String get input => $_getSZ(1);
  @$pb.TagNumber(2)
  set input($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInput() => $_has(1);
  @$pb.TagNumber(2)
  void clearInput() => $_clearField(2);

  /// @gotags: json:"model_override,omitempty" yaml:"model_override,omitempty" bson:"model_override,omitempty"
  @$pb.TagNumber(3)
  $core.String get modelOverride => $_getSZ(2);
  @$pb.TagNumber(3)
  set modelOverride($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasModelOverride() => $_has(2);
  @$pb.TagNumber(3)
  void clearModelOverride() => $_clearField(3);

  /// Immutable agent_id of the agent to invoke. Preferred over agent_name.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get agentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set agentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAgentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearAgentId() => $_clearField(4);
}

class AutomationCallWebhookStep extends $pb.GeneratedMessage {
  factory AutomationCallWebhookStep({
    $core.String? url,
    $core.String? method,
    $core.String? payloadJson,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? headers,
  }) {
    final result = AutomationCallWebhookStep._();
    if (url != null) result.url = url;
    if (method != null) result.method = method;
    if (payloadJson != null) result.payloadJson = payloadJson;
    if (headers != null) result.headers.addEntries(headers);
    return result;
  }

  AutomationCallWebhookStep._();

  factory AutomationCallWebhookStep.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationCallWebhookStep()..mergeFromBuffer(data, registry);
  factory AutomationCallWebhookStep.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationCallWebhookStep()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationCallWebhookStep',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationCallWebhookStep.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..aOS(2, _omitFieldNames ? '' : 'method')
    ..aOS(3, _omitFieldNames ? '' : 'payloadJson')
    ..m<$core.String, $core.String>(4, _omitFieldNames ? '' : 'headers',
        entryClassName: 'AutomationCallWebhookStep.HeadersEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationCallWebhookStep clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationCallWebhookStep copyWith(
          void Function(AutomationCallWebhookStep) updates) =>
      super.copyWith((message) => updates(message as AutomationCallWebhookStep))
          as AutomationCallWebhookStep;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationCallWebhookStep() / AutomationCallWebhookStep.new instead')
  static AutomationCallWebhookStep create() => AutomationCallWebhookStep._();
  static $pb.GeneratedMessage $_createMessage() =>
      AutomationCallWebhookStep._();
  @$core.override
  AutomationCallWebhookStep createEmptyInstance() =>
      AutomationCallWebhookStep._();
  @$core.pragma('dart2js:noInline')
  static AutomationCallWebhookStep getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationCallWebhookStep>(
          AutomationCallWebhookStep.$_createMessage);
  static AutomationCallWebhookStep? _defaultInstance;

  /// @gotags: json:"url,omitempty" yaml:"url,omitempty" bson:"url,omitempty"
  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => $_clearField(1);

  /// @gotags: json:"method,omitempty" yaml:"method,omitempty" bson:"method,omitempty"
  @$pb.TagNumber(2)
  $core.String get method => $_getSZ(1);
  @$pb.TagNumber(2)
  set method($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMethod() => $_has(1);
  @$pb.TagNumber(2)
  void clearMethod() => $_clearField(2);

  /// JSON payload template or static JSON body.
  /// @gotags: json:"payload_json,omitempty" yaml:"payload_json,omitempty" bson:"payload_json,omitempty"
  @$pb.TagNumber(3)
  $core.String get payloadJson => $_getSZ(2);
  @$pb.TagNumber(3)
  set payloadJson($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPayloadJson() => $_has(2);
  @$pb.TagNumber(3)
  void clearPayloadJson() => $_clearField(3);

  /// @gotags: json:"headers,omitempty" yaml:"headers,omitempty" bson:"headers,omitempty"
  @$pb.TagNumber(4)
  $pb.PbMap<$core.String, $core.String> get headers => $_getMap(3);
}

class AutomationSendNotifyGroupStep extends $pb.GeneratedMessage {
  factory AutomationSendNotifyGroupStep({
    $core.String? notifyGroupName,
    $core.String? title,
    $core.String? message,
  }) {
    final result = AutomationSendNotifyGroupStep._();
    if (notifyGroupName != null) result.notifyGroupName = notifyGroupName;
    if (title != null) result.title = title;
    if (message != null) result.message = message;
    return result;
  }

  AutomationSendNotifyGroupStep._();

  factory AutomationSendNotifyGroupStep.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationSendNotifyGroupStep()..mergeFromBuffer(data, registry);
  factory AutomationSendNotifyGroupStep.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationSendNotifyGroupStep()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationSendNotifyGroupStep',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationSendNotifyGroupStep.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'notifyGroupName')
    ..aOS(2, _omitFieldNames ? '' : 'title')
    ..aOS(3, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationSendNotifyGroupStep clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationSendNotifyGroupStep copyWith(
          void Function(AutomationSendNotifyGroupStep) updates) =>
      super.copyWith(
              (message) => updates(message as AutomationSendNotifyGroupStep))
          as AutomationSendNotifyGroupStep;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationSendNotifyGroupStep() / AutomationSendNotifyGroupStep.new instead')
  static AutomationSendNotifyGroupStep create() =>
      AutomationSendNotifyGroupStep._();
  static $pb.GeneratedMessage $_createMessage() =>
      AutomationSendNotifyGroupStep._();
  @$core.override
  AutomationSendNotifyGroupStep createEmptyInstance() =>
      AutomationSendNotifyGroupStep._();
  @$core.pragma('dart2js:noInline')
  static AutomationSendNotifyGroupStep getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationSendNotifyGroupStep>(
          AutomationSendNotifyGroupStep.$_createMessage);
  static AutomationSendNotifyGroupStep? _defaultInstance;

  /// @gotags: json:"notify_group_name,omitempty" yaml:"notify_group_name,omitempty" bson:"notify_group_name,omitempty"
  @$pb.TagNumber(1)
  $core.String get notifyGroupName => $_getSZ(0);
  @$pb.TagNumber(1)
  set notifyGroupName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNotifyGroupName() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotifyGroupName() => $_clearField(1);

  /// @gotags: json:"title,omitempty" yaml:"title,omitempty" bson:"title,omitempty"
  @$pb.TagNumber(2)
  $core.String get title => $_getSZ(1);
  @$pb.TagNumber(2)
  set title($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTitle() => $_has(1);
  @$pb.TagNumber(2)
  void clearTitle() => $_clearField(2);

  /// @gotags: json:"message,omitempty" yaml:"message,omitempty" bson:"message,omitempty"
  @$pb.TagNumber(3)
  $core.String get message => $_getSZ(2);
  @$pb.TagNumber(3)
  set message($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMessage() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessage() => $_clearField(3);
}

class AutomationCreateForumPostStep extends $pb.GeneratedMessage {
  factory AutomationCreateForumPostStep({
    $core.String? threadId,
    $core.String? body,
  }) {
    final result = AutomationCreateForumPostStep._();
    if (threadId != null) result.threadId = threadId;
    if (body != null) result.body = body;
    return result;
  }

  AutomationCreateForumPostStep._();

  factory AutomationCreateForumPostStep.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationCreateForumPostStep()..mergeFromBuffer(data, registry);
  factory AutomationCreateForumPostStep.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationCreateForumPostStep()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationCreateForumPostStep',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationCreateForumPostStep.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'threadId')
    ..aOS(2, _omitFieldNames ? '' : 'body')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationCreateForumPostStep clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationCreateForumPostStep copyWith(
          void Function(AutomationCreateForumPostStep) updates) =>
      super.copyWith(
              (message) => updates(message as AutomationCreateForumPostStep))
          as AutomationCreateForumPostStep;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationCreateForumPostStep() / AutomationCreateForumPostStep.new instead')
  static AutomationCreateForumPostStep create() =>
      AutomationCreateForumPostStep._();
  static $pb.GeneratedMessage $_createMessage() =>
      AutomationCreateForumPostStep._();
  @$core.override
  AutomationCreateForumPostStep createEmptyInstance() =>
      AutomationCreateForumPostStep._();
  @$core.pragma('dart2js:noInline')
  static AutomationCreateForumPostStep getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationCreateForumPostStep>(
          AutomationCreateForumPostStep.$_createMessage);
  static AutomationCreateForumPostStep? _defaultInstance;

  /// @gotags: json:"thread_id,omitempty" yaml:"thread_id,omitempty" bson:"thread_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get threadId => $_getSZ(0);
  @$pb.TagNumber(1)
  set threadId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasThreadId() => $_has(0);
  @$pb.TagNumber(1)
  void clearThreadId() => $_clearField(1);

  /// @gotags: json:"body,omitempty" yaml:"body,omitempty" bson:"body,omitempty"
  @$pb.TagNumber(2)
  $core.String get body => $_getSZ(1);
  @$pb.TagNumber(2)
  set body($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBody() => $_has(1);
  @$pb.TagNumber(2)
  void clearBody() => $_clearField(2);
}

class AutomationPolicy extends $pb.GeneratedMessage {
  factory AutomationPolicy({
    $1.Duration? timeout,
    AutomationRetryPolicy? retry,
    AutomationConcurrencyPolicy? concurrency,
    $core.int? maxOutputBytes,
  }) {
    final result = AutomationPolicy._();
    if (timeout != null) result.timeout = timeout;
    if (retry != null) result.retry = retry;
    if (concurrency != null) result.concurrency = concurrency;
    if (maxOutputBytes != null) result.maxOutputBytes = maxOutputBytes;
    return result;
  }

  AutomationPolicy._();

  factory AutomationPolicy.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationPolicy()..mergeFromBuffer(data, registry);
  factory AutomationPolicy.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationPolicy()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationPolicy',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationPolicy.$_createMessage)
    ..aOM<$1.Duration>(1, _omitFieldNames ? '' : 'timeout',
        subBuilder: $1.Duration.$_createMessage)
    ..aOM<AutomationRetryPolicy>(2, _omitFieldNames ? '' : 'retry',
        subBuilder: AutomationRetryPolicy.$_createMessage)
    ..aE<AutomationConcurrencyPolicy>(3, _omitFieldNames ? '' : 'concurrency',
        enumValues: AutomationConcurrencyPolicy.values)
    ..aI(4, _omitFieldNames ? '' : 'maxOutputBytes')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationPolicy clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationPolicy copyWith(void Function(AutomationPolicy) updates) =>
      super.copyWith((message) => updates(message as AutomationPolicy))
          as AutomationPolicy;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AutomationPolicy() / AutomationPolicy.new instead')
  static AutomationPolicy create() => AutomationPolicy._();
  static $pb.GeneratedMessage $_createMessage() => AutomationPolicy._();
  @$core.override
  AutomationPolicy createEmptyInstance() => AutomationPolicy._();
  @$core.pragma('dart2js:noInline')
  static AutomationPolicy getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AutomationPolicy>(
          AutomationPolicy.$_createMessage);
  static AutomationPolicy? _defaultInstance;

  /// @gotags: json:"timeout,omitempty" yaml:"timeout,omitempty" bson:"timeout,omitempty"
  @$pb.TagNumber(1)
  $1.Duration get timeout => $_getN(0);
  @$pb.TagNumber(1)
  set timeout($1.Duration value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTimeout() => $_has(0);
  @$pb.TagNumber(1)
  void clearTimeout() => $_clearField(1);
  @$pb.TagNumber(1)
  $1.Duration ensureTimeout() => $_ensure(0);

  /// @gotags: json:"retry,omitempty" yaml:"retry,omitempty" bson:"retry,omitempty"
  @$pb.TagNumber(2)
  AutomationRetryPolicy get retry => $_getN(1);
  @$pb.TagNumber(2)
  set retry(AutomationRetryPolicy value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRetry() => $_has(1);
  @$pb.TagNumber(2)
  void clearRetry() => $_clearField(2);
  @$pb.TagNumber(2)
  AutomationRetryPolicy ensureRetry() => $_ensure(1);

  /// @gotags: json:"concurrency,omitempty" yaml:"concurrency,omitempty" bson:"concurrency,omitempty"
  @$pb.TagNumber(3)
  AutomationConcurrencyPolicy get concurrency => $_getN(2);
  @$pb.TagNumber(3)
  set concurrency(AutomationConcurrencyPolicy value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasConcurrency() => $_has(2);
  @$pb.TagNumber(3)
  void clearConcurrency() => $_clearField(3);

  /// Maximum bytes persisted for run/step output previews. 0 uses server default.
  /// @gotags: json:"max_output_bytes,omitempty" yaml:"max_output_bytes,omitempty" bson:"max_output_bytes,omitempty"
  @$pb.TagNumber(4)
  $core.int get maxOutputBytes => $_getIZ(3);
  @$pb.TagNumber(4)
  set maxOutputBytes($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMaxOutputBytes() => $_has(3);
  @$pb.TagNumber(4)
  void clearMaxOutputBytes() => $_clearField(4);
}

class AutomationRetryPolicy extends $pb.GeneratedMessage {
  factory AutomationRetryPolicy({
    $core.int? maxAttempts,
    $1.Duration? backoff,
  }) {
    final result = AutomationRetryPolicy._();
    if (maxAttempts != null) result.maxAttempts = maxAttempts;
    if (backoff != null) result.backoff = backoff;
    return result;
  }

  AutomationRetryPolicy._();

  factory AutomationRetryPolicy.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationRetryPolicy()..mergeFromBuffer(data, registry);
  factory AutomationRetryPolicy.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationRetryPolicy()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationRetryPolicy',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationRetryPolicy.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'maxAttempts')
    ..aOM<$1.Duration>(2, _omitFieldNames ? '' : 'backoff',
        subBuilder: $1.Duration.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationRetryPolicy clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationRetryPolicy copyWith(
          void Function(AutomationRetryPolicy) updates) =>
      super.copyWith((message) => updates(message as AutomationRetryPolicy))
          as AutomationRetryPolicy;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AutomationRetryPolicy() / AutomationRetryPolicy.new instead')
  static AutomationRetryPolicy create() => AutomationRetryPolicy._();
  static $pb.GeneratedMessage $_createMessage() => AutomationRetryPolicy._();
  @$core.override
  AutomationRetryPolicy createEmptyInstance() => AutomationRetryPolicy._();
  @$core.pragma('dart2js:noInline')
  static AutomationRetryPolicy getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AutomationRetryPolicy>(
          AutomationRetryPolicy.$_createMessage);
  static AutomationRetryPolicy? _defaultInstance;

  /// Number of retries after the initial attempt. 0 disables retries.
  /// @gotags: json:"max_attempts,omitempty" yaml:"max_attempts,omitempty" bson:"max_attempts,omitempty"
  @$pb.TagNumber(1)
  $core.int get maxAttempts => $_getIZ(0);
  @$pb.TagNumber(1)
  set maxAttempts($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMaxAttempts() => $_has(0);
  @$pb.TagNumber(1)
  void clearMaxAttempts() => $_clearField(1);

  /// @gotags: json:"backoff,omitempty" yaml:"backoff,omitempty" bson:"backoff,omitempty"
  @$pb.TagNumber(2)
  $1.Duration get backoff => $_getN(1);
  @$pb.TagNumber(2)
  set backoff($1.Duration value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasBackoff() => $_has(1);
  @$pb.TagNumber(2)
  void clearBackoff() => $_clearField(2);
  @$pb.TagNumber(2)
  $1.Duration ensureBackoff() => $_ensure(1);
}

class AutomationRun extends $pb.GeneratedMessage {
  factory AutomationRun({
    $core.String? id,
    $core.String? automationName,
    AutomationTriggerType? triggerType,
    AutomationRunStatus? status,
    $core.String? triggerPayloadJson,
    $core.String? error,
    $0.Timestamp? startedAt,
    $0.Timestamp? finishedAt,
    $fixnum.Int64? durationMs,
    $core.String? sessionAppName,
    $core.String? sessionUserId,
    $core.String? sessionId,
    $core.String? agentName,
    $core.String? agentId,
    $core.String? workspaceId,
  }) {
    final result = AutomationRun._();
    if (id != null) result.id = id;
    if (automationName != null) result.automationName = automationName;
    if (triggerType != null) result.triggerType = triggerType;
    if (status != null) result.status = status;
    if (triggerPayloadJson != null)
      result.triggerPayloadJson = triggerPayloadJson;
    if (error != null) result.error = error;
    if (startedAt != null) result.startedAt = startedAt;
    if (finishedAt != null) result.finishedAt = finishedAt;
    if (durationMs != null) result.durationMs = durationMs;
    if (sessionAppName != null) result.sessionAppName = sessionAppName;
    if (sessionUserId != null) result.sessionUserId = sessionUserId;
    if (sessionId != null) result.sessionId = sessionId;
    if (agentName != null) result.agentName = agentName;
    if (agentId != null) result.agentId = agentId;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  AutomationRun._();

  factory AutomationRun.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationRun()..mergeFromBuffer(data, registry);
  factory AutomationRun.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationRun()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationRun',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationRun.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'automationName')
    ..aE<AutomationTriggerType>(3, _omitFieldNames ? '' : 'triggerType',
        enumValues: AutomationTriggerType.values)
    ..aE<AutomationRunStatus>(4, _omitFieldNames ? '' : 'status',
        enumValues: AutomationRunStatus.values)
    ..aOS(5, _omitFieldNames ? '' : 'triggerPayloadJson')
    ..aOS(6, _omitFieldNames ? '' : 'error')
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'finishedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aInt64(9, _omitFieldNames ? '' : 'durationMs')
    ..aOS(10, _omitFieldNames ? '' : 'sessionAppName')
    ..aOS(11, _omitFieldNames ? '' : 'sessionUserId')
    ..aOS(12, _omitFieldNames ? '' : 'sessionId')
    ..aOS(13, _omitFieldNames ? '' : 'agentName')
    ..aOS(14, _omitFieldNames ? '' : 'agentId')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationRun clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationRun copyWith(void Function(AutomationRun) updates) =>
      super.copyWith((message) => updates(message as AutomationRun))
          as AutomationRun;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AutomationRun() / AutomationRun.new instead')
  static AutomationRun create() => AutomationRun._();
  static $pb.GeneratedMessage $_createMessage() => AutomationRun._();
  @$core.override
  AutomationRun createEmptyInstance() => AutomationRun._();
  @$core.pragma('dart2js:noInline')
  static AutomationRun getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AutomationRun>(
          AutomationRun.$_createMessage);
  static AutomationRun? _defaultInstance;

  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"automation_name,omitempty" yaml:"automation_name,omitempty" bson:"automation_name,omitempty"
  @$pb.TagNumber(2)
  $core.String get automationName => $_getSZ(1);
  @$pb.TagNumber(2)
  set automationName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAutomationName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAutomationName() => $_clearField(2);

  /// @gotags: json:"trigger_type,omitempty" yaml:"trigger_type,omitempty" bson:"trigger_type,omitempty"
  @$pb.TagNumber(3)
  AutomationTriggerType get triggerType => $_getN(2);
  @$pb.TagNumber(3)
  set triggerType(AutomationTriggerType value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasTriggerType() => $_has(2);
  @$pb.TagNumber(3)
  void clearTriggerType() => $_clearField(3);

  /// @gotags: json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(4)
  AutomationRunStatus get status => $_getN(3);
  @$pb.TagNumber(4)
  set status(AutomationRunStatus value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  /// Truncated trigger payload preview.
  /// @gotags: json:"trigger_payload_json,omitempty" yaml:"trigger_payload_json,omitempty" bson:"trigger_payload_json,omitempty"
  @$pb.TagNumber(5)
  $core.String get triggerPayloadJson => $_getSZ(4);
  @$pb.TagNumber(5)
  set triggerPayloadJson($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTriggerPayloadJson() => $_has(4);
  @$pb.TagNumber(5)
  void clearTriggerPayloadJson() => $_clearField(5);

  /// @gotags: json:"error,omitempty" yaml:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(6)
  $core.String get error => $_getSZ(5);
  @$pb.TagNumber(6)
  set error($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasError() => $_has(5);
  @$pb.TagNumber(6)
  void clearError() => $_clearField(6);

  /// @gotags: json:"started_at,omitempty" yaml:"started_at,omitempty" bson:"started_at,omitempty"
  @$pb.TagNumber(7)
  $0.Timestamp get startedAt => $_getN(6);
  @$pb.TagNumber(7)
  set startedAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasStartedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearStartedAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureStartedAt() => $_ensure(6);

  /// @gotags: json:"finished_at,omitempty" yaml:"finished_at,omitempty" bson:"finished_at,omitempty"
  @$pb.TagNumber(8)
  $0.Timestamp get finishedAt => $_getN(7);
  @$pb.TagNumber(8)
  set finishedAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasFinishedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearFinishedAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureFinishedAt() => $_ensure(7);

  /// @gotags: json:"duration_ms,omitempty" yaml:"duration_ms,omitempty" bson:"duration_ms,omitempty"
  @$pb.TagNumber(9)
  $fixnum.Int64 get durationMs => $_getI64(8);
  @$pb.TagNumber(9)
  set durationMs($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasDurationMs() => $_has(8);
  @$pb.TagNumber(9)
  void clearDurationMs() => $_clearField(9);

  /// Session coordinates of a run paused on a Human Input node (WAITING_INPUT).
  /// Empty for every other status. A reply to these coordinates via ReplySession
  /// resumes the paused workflow and finalizes the run (ADR-0003, mirroring cron).
  /// @gotags: json:"session_app_name,omitempty" yaml:"session_app_name,omitempty" bson:"session_app_name,omitempty"
  @$pb.TagNumber(10)
  $core.String get sessionAppName => $_getSZ(9);
  @$pb.TagNumber(10)
  set sessionAppName($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasSessionAppName() => $_has(9);
  @$pb.TagNumber(10)
  void clearSessionAppName() => $_clearField(10);

  /// @gotags: json:"session_user_id,omitempty" yaml:"session_user_id,omitempty" bson:"session_user_id,omitempty"
  @$pb.TagNumber(11)
  $core.String get sessionUserId => $_getSZ(10);
  @$pb.TagNumber(11)
  set sessionUserId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasSessionUserId() => $_has(10);
  @$pb.TagNumber(11)
  void clearSessionUserId() => $_clearField(11);

  /// @gotags: json:"session_id,omitempty" yaml:"session_id,omitempty" bson:"session_id,omitempty"
  @$pb.TagNumber(12)
  $core.String get sessionId => $_getSZ(11);
  @$pb.TagNumber(12)
  set sessionId($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasSessionId() => $_has(11);
  @$pb.TagNumber(12)
  void clearSessionId() => $_clearField(12);

  /// Legacy name of the agent that paused. Historical records written before
  /// the Agent ID migration carry only this field.
  /// @gotags: json:"agent_name,omitempty" yaml:"agent_name,omitempty" bson:"agent_name,omitempty"
  @$pb.TagNumber(13)
  $core.String get agentName => $_getSZ(12);
  @$pb.TagNumber(13)
  set agentName($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasAgentName() => $_has(12);
  @$pb.TagNumber(13)
  void clearAgentName() => $_clearField(13);

  /// Immutable agent_id of the agent that paused; the agent a reply must
  /// address. Preferred over agent_name.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(14)
  $core.String get agentId => $_getSZ(13);
  @$pb.TagNumber(14)
  set agentId($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasAgentId() => $_has(13);
  @$pb.TagNumber(14)
  void clearAgentId() => $_clearField(14);

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(14);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(14, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(14);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class AutomationStepRun extends $pb.GeneratedMessage {
  factory AutomationStepRun({
    $core.String? id,
    $core.String? runId,
    $core.String? automationName,
    $core.String? stepName,
    AutomationStepType? stepType,
    AutomationStepRunStatus? status,
    $core.int? attemptCount,
    $core.String? inputJson,
    $core.String? outputJson,
    $core.String? error,
    $core.String? invocationId,
    $0.Timestamp? startedAt,
    $0.Timestamp? finishedAt,
    $fixnum.Int64? durationMs,
    $core.int? order,
    $core.bool? truncated,
    $core.String? workspaceId,
  }) {
    final result = AutomationStepRun._();
    if (id != null) result.id = id;
    if (runId != null) result.runId = runId;
    if (automationName != null) result.automationName = automationName;
    if (stepName != null) result.stepName = stepName;
    if (stepType != null) result.stepType = stepType;
    if (status != null) result.status = status;
    if (attemptCount != null) result.attemptCount = attemptCount;
    if (inputJson != null) result.inputJson = inputJson;
    if (outputJson != null) result.outputJson = outputJson;
    if (error != null) result.error = error;
    if (invocationId != null) result.invocationId = invocationId;
    if (startedAt != null) result.startedAt = startedAt;
    if (finishedAt != null) result.finishedAt = finishedAt;
    if (durationMs != null) result.durationMs = durationMs;
    if (order != null) result.order = order;
    if (truncated != null) result.truncated = truncated;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  AutomationStepRun._();

  factory AutomationStepRun.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationStepRun()..mergeFromBuffer(data, registry);
  factory AutomationStepRun.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AutomationStepRun()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AutomationStepRun',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AutomationStepRun.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'runId')
    ..aOS(3, _omitFieldNames ? '' : 'automationName')
    ..aOS(4, _omitFieldNames ? '' : 'stepName')
    ..aE<AutomationStepType>(5, _omitFieldNames ? '' : 'stepType',
        enumValues: AutomationStepType.values)
    ..aE<AutomationStepRunStatus>(6, _omitFieldNames ? '' : 'status',
        enumValues: AutomationStepRunStatus.values)
    ..aI(7, _omitFieldNames ? '' : 'attemptCount')
    ..aOS(8, _omitFieldNames ? '' : 'inputJson')
    ..aOS(9, _omitFieldNames ? '' : 'outputJson')
    ..aOS(10, _omitFieldNames ? '' : 'error')
    ..aOS(11, _omitFieldNames ? '' : 'invocationId')
    ..aOM<$0.Timestamp>(12, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(13, _omitFieldNames ? '' : 'finishedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aInt64(14, _omitFieldNames ? '' : 'durationMs')
    ..aI(15, _omitFieldNames ? '' : 'order')
    ..aOB(16, _omitFieldNames ? '' : 'truncated')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationStepRun clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AutomationStepRun copyWith(void Function(AutomationStepRun) updates) =>
      super.copyWith((message) => updates(message as AutomationStepRun))
          as AutomationStepRun;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AutomationStepRun() / AutomationStepRun.new instead')
  static AutomationStepRun create() => AutomationStepRun._();
  static $pb.GeneratedMessage $_createMessage() => AutomationStepRun._();
  @$core.override
  AutomationStepRun createEmptyInstance() => AutomationStepRun._();
  @$core.pragma('dart2js:noInline')
  static AutomationStepRun getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AutomationStepRun>(
          AutomationStepRun.$_createMessage);
  static AutomationStepRun? _defaultInstance;

  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"run_id,omitempty" yaml:"run_id,omitempty" bson:"run_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get runId => $_getSZ(1);
  @$pb.TagNumber(2)
  set runId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRunId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRunId() => $_clearField(2);

  /// @gotags: json:"automation_name,omitempty" yaml:"automation_name,omitempty" bson:"automation_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get automationName => $_getSZ(2);
  @$pb.TagNumber(3)
  set automationName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAutomationName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAutomationName() => $_clearField(3);

  /// @gotags: json:"step_name,omitempty" yaml:"step_name,omitempty" bson:"step_name,omitempty"
  @$pb.TagNumber(4)
  $core.String get stepName => $_getSZ(3);
  @$pb.TagNumber(4)
  set stepName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStepName() => $_has(3);
  @$pb.TagNumber(4)
  void clearStepName() => $_clearField(4);

  /// @gotags: json:"step_type,omitempty" yaml:"step_type,omitempty" bson:"step_type,omitempty"
  @$pb.TagNumber(5)
  AutomationStepType get stepType => $_getN(4);
  @$pb.TagNumber(5)
  set stepType(AutomationStepType value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStepType() => $_has(4);
  @$pb.TagNumber(5)
  void clearStepType() => $_clearField(5);

  /// @gotags: json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(6)
  AutomationStepRunStatus get status => $_getN(5);
  @$pb.TagNumber(6)
  set status(AutomationStepRunStatus value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);

  /// @gotags: json:"attempt_count,omitempty" yaml:"attempt_count,omitempty" bson:"attempt_count,omitempty"
  @$pb.TagNumber(7)
  $core.int get attemptCount => $_getIZ(6);
  @$pb.TagNumber(7)
  set attemptCount($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasAttemptCount() => $_has(6);
  @$pb.TagNumber(7)
  void clearAttemptCount() => $_clearField(7);

  /// @gotags: json:"input_json,omitempty" yaml:"input_json,omitempty" bson:"input_json,omitempty"
  @$pb.TagNumber(8)
  $core.String get inputJson => $_getSZ(7);
  @$pb.TagNumber(8)
  set inputJson($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasInputJson() => $_has(7);
  @$pb.TagNumber(8)
  void clearInputJson() => $_clearField(8);

  /// @gotags: json:"output_json,omitempty" yaml:"output_json,omitempty" bson:"output_json,omitempty"
  @$pb.TagNumber(9)
  $core.String get outputJson => $_getSZ(8);
  @$pb.TagNumber(9)
  set outputJson($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasOutputJson() => $_has(8);
  @$pb.TagNumber(9)
  void clearOutputJson() => $_clearField(9);

  /// @gotags: json:"error,omitempty" yaml:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(10)
  $core.String get error => $_getSZ(9);
  @$pb.TagNumber(10)
  set error($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasError() => $_has(9);
  @$pb.TagNumber(10)
  void clearError() => $_clearField(10);

  /// @gotags: json:"invocation_id,omitempty" yaml:"invocation_id,omitempty" bson:"invocation_id,omitempty"
  @$pb.TagNumber(11)
  $core.String get invocationId => $_getSZ(10);
  @$pb.TagNumber(11)
  set invocationId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasInvocationId() => $_has(10);
  @$pb.TagNumber(11)
  void clearInvocationId() => $_clearField(11);

  /// @gotags: json:"started_at,omitempty" yaml:"started_at,omitempty" bson:"started_at,omitempty"
  @$pb.TagNumber(12)
  $0.Timestamp get startedAt => $_getN(11);
  @$pb.TagNumber(12)
  set startedAt($0.Timestamp value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasStartedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearStartedAt() => $_clearField(12);
  @$pb.TagNumber(12)
  $0.Timestamp ensureStartedAt() => $_ensure(11);

  /// @gotags: json:"finished_at,omitempty" yaml:"finished_at,omitempty" bson:"finished_at,omitempty"
  @$pb.TagNumber(13)
  $0.Timestamp get finishedAt => $_getN(12);
  @$pb.TagNumber(13)
  set finishedAt($0.Timestamp value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasFinishedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearFinishedAt() => $_clearField(13);
  @$pb.TagNumber(13)
  $0.Timestamp ensureFinishedAt() => $_ensure(12);

  /// @gotags: json:"duration_ms,omitempty" yaml:"duration_ms,omitempty" bson:"duration_ms,omitempty"
  @$pb.TagNumber(14)
  $fixnum.Int64 get durationMs => $_getI64(13);
  @$pb.TagNumber(14)
  set durationMs($fixnum.Int64 value) => $_setInt64(13, value);
  @$pb.TagNumber(14)
  $core.bool hasDurationMs() => $_has(13);
  @$pb.TagNumber(14)
  void clearDurationMs() => $_clearField(14);

  /// @gotags: json:"order,omitempty" yaml:"order,omitempty" bson:"order,omitempty"
  @$pb.TagNumber(15)
  $core.int get order => $_getIZ(14);
  @$pb.TagNumber(15)
  set order($core.int value) => $_setSignedInt32(14, value);
  @$pb.TagNumber(15)
  $core.bool hasOrder() => $_has(14);
  @$pb.TagNumber(15)
  void clearOrder() => $_clearField(15);

  /// @gotags: json:"truncated,omitempty" yaml:"truncated,omitempty" bson:"truncated,omitempty"
  @$pb.TagNumber(16)
  $core.bool get truncated => $_getBF(15);
  @$pb.TagNumber(16)
  set truncated($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasTruncated() => $_has(15);
  @$pb.TagNumber(16)
  void clearTruncated() => $_clearField(16);

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(16);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(16, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(16);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListAutomationsRequest extends $pb.GeneratedMessage {
  factory ListAutomationsRequest() => ListAutomationsRequest._();

  ListAutomationsRequest._();

  factory ListAutomationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationsRequest()..mergeFromBuffer(data, registry);
  factory ListAutomationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAutomationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAutomationsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationsRequest copyWith(
          void Function(ListAutomationsRequest) updates) =>
      super.copyWith((message) => updates(message as ListAutomationsRequest))
          as ListAutomationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAutomationsRequest() / ListAutomationsRequest.new instead')
  static ListAutomationsRequest create() => ListAutomationsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListAutomationsRequest._();
  @$core.override
  ListAutomationsRequest createEmptyInstance() => ListAutomationsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAutomationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAutomationsRequest>(
          ListAutomationsRequest.$_createMessage);
  static ListAutomationsRequest? _defaultInstance;
}

class ListAutomationsResponse extends $pb.GeneratedMessage {
  factory ListAutomationsResponse({
    $core.Iterable<Automation>? automations,
  }) {
    final result = ListAutomationsResponse._();
    if (automations != null) result.automations.addAll(automations);
    return result;
  }

  ListAutomationsResponse._();

  factory ListAutomationsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationsResponse()..mergeFromBuffer(data, registry);
  factory ListAutomationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAutomationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAutomationsResponse.$_createMessage)
    ..pPM<Automation>(1, _omitFieldNames ? '' : 'automations',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationsResponse copyWith(
          void Function(ListAutomationsResponse) updates) =>
      super.copyWith((message) => updates(message as ListAutomationsResponse))
          as ListAutomationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAutomationsResponse() / ListAutomationsResponse.new instead')
  static ListAutomationsResponse create() => ListAutomationsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListAutomationsResponse._();
  @$core.override
  ListAutomationsResponse createEmptyInstance() => ListAutomationsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAutomationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAutomationsResponse>(
          ListAutomationsResponse.$_createMessage);
  static ListAutomationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Automation> get automations => $_getList(0);
}

class GetAutomationRequest extends $pb.GeneratedMessage {
  factory GetAutomationRequest({
    $core.String? name,
  }) {
    final result = GetAutomationRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetAutomationRequest._();

  factory GetAutomationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationRequest()..mergeFromBuffer(data, registry);
  factory GetAutomationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAutomationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAutomationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationRequest copyWith(void Function(GetAutomationRequest) updates) =>
      super.copyWith((message) => updates(message as GetAutomationRequest))
          as GetAutomationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAutomationRequest() / GetAutomationRequest.new instead')
  static GetAutomationRequest create() => GetAutomationRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAutomationRequest._();
  @$core.override
  GetAutomationRequest createEmptyInstance() => GetAutomationRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAutomationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAutomationRequest>(
          GetAutomationRequest.$_createMessage);
  static GetAutomationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetAutomationResponse extends $pb.GeneratedMessage {
  factory GetAutomationResponse({
    Automation? automation,
  }) {
    final result = GetAutomationResponse._();
    if (automation != null) result.automation = automation;
    return result;
  }

  GetAutomationResponse._();

  factory GetAutomationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationResponse()..mergeFromBuffer(data, registry);
  factory GetAutomationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAutomationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAutomationResponse.$_createMessage)
    ..aOM<Automation>(1, _omitFieldNames ? '' : 'automation',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationResponse copyWith(
          void Function(GetAutomationResponse) updates) =>
      super.copyWith((message) => updates(message as GetAutomationResponse))
          as GetAutomationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAutomationResponse() / GetAutomationResponse.new instead')
  static GetAutomationResponse create() => GetAutomationResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetAutomationResponse._();
  @$core.override
  GetAutomationResponse createEmptyInstance() => GetAutomationResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAutomationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAutomationResponse>(
          GetAutomationResponse.$_createMessage);
  static GetAutomationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Automation get automation => $_getN(0);
  @$pb.TagNumber(1)
  set automation(Automation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomation() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomation() => $_clearField(1);
  @$pb.TagNumber(1)
  Automation ensureAutomation() => $_ensure(0);
}

class CreateAutomationRequest extends $pb.GeneratedMessage {
  factory CreateAutomationRequest({
    Automation? automation,
  }) {
    final result = CreateAutomationRequest._();
    if (automation != null) result.automation = automation;
    return result;
  }

  CreateAutomationRequest._();

  factory CreateAutomationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAutomationRequest()..mergeFromBuffer(data, registry);
  factory CreateAutomationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAutomationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAutomationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAutomationRequest.$_createMessage)
    ..aOM<Automation>(1, _omitFieldNames ? '' : 'automation',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAutomationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAutomationRequest copyWith(
          void Function(CreateAutomationRequest) updates) =>
      super.copyWith((message) => updates(message as CreateAutomationRequest))
          as CreateAutomationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAutomationRequest() / CreateAutomationRequest.new instead')
  static CreateAutomationRequest create() => CreateAutomationRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateAutomationRequest._();
  @$core.override
  CreateAutomationRequest createEmptyInstance() => CreateAutomationRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateAutomationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAutomationRequest>(
          CreateAutomationRequest.$_createMessage);
  static CreateAutomationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  Automation get automation => $_getN(0);
  @$pb.TagNumber(1)
  set automation(Automation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomation() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomation() => $_clearField(1);
  @$pb.TagNumber(1)
  Automation ensureAutomation() => $_ensure(0);
}

class CreateAutomationResponse extends $pb.GeneratedMessage {
  factory CreateAutomationResponse({
    Automation? automation,
  }) {
    final result = CreateAutomationResponse._();
    if (automation != null) result.automation = automation;
    return result;
  }

  CreateAutomationResponse._();

  factory CreateAutomationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAutomationResponse()..mergeFromBuffer(data, registry);
  factory CreateAutomationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAutomationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAutomationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAutomationResponse.$_createMessage)
    ..aOM<Automation>(1, _omitFieldNames ? '' : 'automation',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAutomationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAutomationResponse copyWith(
          void Function(CreateAutomationResponse) updates) =>
      super.copyWith((message) => updates(message as CreateAutomationResponse))
          as CreateAutomationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAutomationResponse() / CreateAutomationResponse.new instead')
  static CreateAutomationResponse create() => CreateAutomationResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateAutomationResponse._();
  @$core.override
  CreateAutomationResponse createEmptyInstance() =>
      CreateAutomationResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateAutomationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAutomationResponse>(
          CreateAutomationResponse.$_createMessage);
  static CreateAutomationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Automation get automation => $_getN(0);
  @$pb.TagNumber(1)
  set automation(Automation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomation() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomation() => $_clearField(1);
  @$pb.TagNumber(1)
  Automation ensureAutomation() => $_ensure(0);
}

class UpdateAutomationRequest extends $pb.GeneratedMessage {
  factory UpdateAutomationRequest({
    Automation? automation,
  }) {
    final result = UpdateAutomationRequest._();
    if (automation != null) result.automation = automation;
    return result;
  }

  UpdateAutomationRequest._();

  factory UpdateAutomationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAutomationRequest()..mergeFromBuffer(data, registry);
  factory UpdateAutomationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAutomationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAutomationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAutomationRequest.$_createMessage)
    ..aOM<Automation>(1, _omitFieldNames ? '' : 'automation',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAutomationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAutomationRequest copyWith(
          void Function(UpdateAutomationRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateAutomationRequest))
          as UpdateAutomationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAutomationRequest() / UpdateAutomationRequest.new instead')
  static UpdateAutomationRequest create() => UpdateAutomationRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAutomationRequest._();
  @$core.override
  UpdateAutomationRequest createEmptyInstance() => UpdateAutomationRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAutomationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAutomationRequest>(
          UpdateAutomationRequest.$_createMessage);
  static UpdateAutomationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  Automation get automation => $_getN(0);
  @$pb.TagNumber(1)
  set automation(Automation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomation() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomation() => $_clearField(1);
  @$pb.TagNumber(1)
  Automation ensureAutomation() => $_ensure(0);
}

class UpdateAutomationResponse extends $pb.GeneratedMessage {
  factory UpdateAutomationResponse({
    Automation? automation,
  }) {
    final result = UpdateAutomationResponse._();
    if (automation != null) result.automation = automation;
    return result;
  }

  UpdateAutomationResponse._();

  factory UpdateAutomationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAutomationResponse()..mergeFromBuffer(data, registry);
  factory UpdateAutomationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAutomationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAutomationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAutomationResponse.$_createMessage)
    ..aOM<Automation>(1, _omitFieldNames ? '' : 'automation',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAutomationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAutomationResponse copyWith(
          void Function(UpdateAutomationResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateAutomationResponse))
          as UpdateAutomationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAutomationResponse() / UpdateAutomationResponse.new instead')
  static UpdateAutomationResponse create() => UpdateAutomationResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAutomationResponse._();
  @$core.override
  UpdateAutomationResponse createEmptyInstance() =>
      UpdateAutomationResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAutomationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAutomationResponse>(
          UpdateAutomationResponse.$_createMessage);
  static UpdateAutomationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Automation get automation => $_getN(0);
  @$pb.TagNumber(1)
  set automation(Automation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomation() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomation() => $_clearField(1);
  @$pb.TagNumber(1)
  Automation ensureAutomation() => $_ensure(0);
}

class DeleteAutomationRequest extends $pb.GeneratedMessage {
  factory DeleteAutomationRequest({
    $core.String? name,
  }) {
    final result = DeleteAutomationRequest._();
    if (name != null) result.name = name;
    return result;
  }

  DeleteAutomationRequest._();

  factory DeleteAutomationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAutomationRequest()..mergeFromBuffer(data, registry);
  factory DeleteAutomationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAutomationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAutomationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAutomationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAutomationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAutomationRequest copyWith(
          void Function(DeleteAutomationRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteAutomationRequest))
          as DeleteAutomationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAutomationRequest() / DeleteAutomationRequest.new instead')
  static DeleteAutomationRequest create() => DeleteAutomationRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAutomationRequest._();
  @$core.override
  DeleteAutomationRequest createEmptyInstance() => DeleteAutomationRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteAutomationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAutomationRequest>(
          DeleteAutomationRequest.$_createMessage);
  static DeleteAutomationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class DeleteAutomationResponse extends $pb.GeneratedMessage {
  factory DeleteAutomationResponse({
    Automation? automation,
  }) {
    final result = DeleteAutomationResponse._();
    if (automation != null) result.automation = automation;
    return result;
  }

  DeleteAutomationResponse._();

  factory DeleteAutomationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAutomationResponse()..mergeFromBuffer(data, registry);
  factory DeleteAutomationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAutomationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAutomationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAutomationResponse.$_createMessage)
    ..aOM<Automation>(1, _omitFieldNames ? '' : 'automation',
        subBuilder: Automation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAutomationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAutomationResponse copyWith(
          void Function(DeleteAutomationResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteAutomationResponse))
          as DeleteAutomationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAutomationResponse() / DeleteAutomationResponse.new instead')
  static DeleteAutomationResponse create() => DeleteAutomationResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAutomationResponse._();
  @$core.override
  DeleteAutomationResponse createEmptyInstance() =>
      DeleteAutomationResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteAutomationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAutomationResponse>(
          DeleteAutomationResponse.$_createMessage);
  static DeleteAutomationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Automation get automation => $_getN(0);
  @$pb.TagNumber(1)
  set automation(Automation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomation() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomation() => $_clearField(1);
  @$pb.TagNumber(1)
  Automation ensureAutomation() => $_ensure(0);
}

class RunAutomationNowRequest extends $pb.GeneratedMessage {
  factory RunAutomationNowRequest({
    $core.String? name,
    $core.String? triggerPayloadJson,
  }) {
    final result = RunAutomationNowRequest._();
    if (name != null) result.name = name;
    if (triggerPayloadJson != null)
      result.triggerPayloadJson = triggerPayloadJson;
    return result;
  }

  RunAutomationNowRequest._();

  factory RunAutomationNowRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunAutomationNowRequest()..mergeFromBuffer(data, registry);
  factory RunAutomationNowRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunAutomationNowRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RunAutomationNowRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RunAutomationNowRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'triggerPayloadJson')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunAutomationNowRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunAutomationNowRequest copyWith(
          void Function(RunAutomationNowRequest) updates) =>
      super.copyWith((message) => updates(message as RunAutomationNowRequest))
          as RunAutomationNowRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RunAutomationNowRequest() / RunAutomationNowRequest.new instead')
  static RunAutomationNowRequest create() => RunAutomationNowRequest._();
  static $pb.GeneratedMessage $_createMessage() => RunAutomationNowRequest._();
  @$core.override
  RunAutomationNowRequest createEmptyInstance() => RunAutomationNowRequest._();
  @$core.pragma('dart2js:noInline')
  static RunAutomationNowRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RunAutomationNowRequest>(
          RunAutomationNowRequest.$_createMessage);
  static RunAutomationNowRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get triggerPayloadJson => $_getSZ(1);
  @$pb.TagNumber(2)
  set triggerPayloadJson($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTriggerPayloadJson() => $_has(1);
  @$pb.TagNumber(2)
  void clearTriggerPayloadJson() => $_clearField(2);
}

class RunAutomationNowResponse extends $pb.GeneratedMessage {
  factory RunAutomationNowResponse({
    AutomationRun? run,
  }) {
    final result = RunAutomationNowResponse._();
    if (run != null) result.run = run;
    return result;
  }

  RunAutomationNowResponse._();

  factory RunAutomationNowResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunAutomationNowResponse()..mergeFromBuffer(data, registry);
  factory RunAutomationNowResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunAutomationNowResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RunAutomationNowResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RunAutomationNowResponse.$_createMessage)
    ..aOM<AutomationRun>(1, _omitFieldNames ? '' : 'run',
        subBuilder: AutomationRun.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunAutomationNowResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunAutomationNowResponse copyWith(
          void Function(RunAutomationNowResponse) updates) =>
      super.copyWith((message) => updates(message as RunAutomationNowResponse))
          as RunAutomationNowResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RunAutomationNowResponse() / RunAutomationNowResponse.new instead')
  static RunAutomationNowResponse create() => RunAutomationNowResponse._();
  static $pb.GeneratedMessage $_createMessage() => RunAutomationNowResponse._();
  @$core.override
  RunAutomationNowResponse createEmptyInstance() =>
      RunAutomationNowResponse._();
  @$core.pragma('dart2js:noInline')
  static RunAutomationNowResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RunAutomationNowResponse>(
          RunAutomationNowResponse.$_createMessage);
  static RunAutomationNowResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AutomationRun get run => $_getN(0);
  @$pb.TagNumber(1)
  set run(AutomationRun value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRun() => $_has(0);
  @$pb.TagNumber(1)
  void clearRun() => $_clearField(1);
  @$pb.TagNumber(1)
  AutomationRun ensureRun() => $_ensure(0);
}

class ListAutomationRunsRequest extends $pb.GeneratedMessage {
  factory ListAutomationRunsRequest({
    $core.String? automationName,
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = ListAutomationRunsRequest._();
    if (automationName != null) result.automationName = automationName;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListAutomationRunsRequest._();

  factory ListAutomationRunsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationRunsRequest()..mergeFromBuffer(data, registry);
  factory ListAutomationRunsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationRunsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAutomationRunsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAutomationRunsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'automationName')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationRunsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationRunsRequest copyWith(
          void Function(ListAutomationRunsRequest) updates) =>
      super.copyWith((message) => updates(message as ListAutomationRunsRequest))
          as ListAutomationRunsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAutomationRunsRequest() / ListAutomationRunsRequest.new instead')
  static ListAutomationRunsRequest create() => ListAutomationRunsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAutomationRunsRequest._();
  @$core.override
  ListAutomationRunsRequest createEmptyInstance() =>
      ListAutomationRunsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAutomationRunsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAutomationRunsRequest>(
          ListAutomationRunsRequest.$_createMessage);
  static ListAutomationRunsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get automationName => $_getSZ(0);
  @$pb.TagNumber(1)
  set automationName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAutomationName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAutomationName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get pageSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set pageSize($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageSize() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get pageToken => $_getSZ(2);
  @$pb.TagNumber(3)
  set pageToken($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageToken() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageToken() => $_clearField(3);
}

class ListAutomationRunsResponse extends $pb.GeneratedMessage {
  factory ListAutomationRunsResponse({
    $core.Iterable<AutomationRun>? runs,
    $core.String? nextPageToken,
  }) {
    final result = ListAutomationRunsResponse._();
    if (runs != null) result.runs.addAll(runs);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  ListAutomationRunsResponse._();

  factory ListAutomationRunsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationRunsResponse()..mergeFromBuffer(data, registry);
  factory ListAutomationRunsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationRunsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAutomationRunsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAutomationRunsResponse.$_createMessage)
    ..pPM<AutomationRun>(1, _omitFieldNames ? '' : 'runs',
        subBuilder: AutomationRun.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationRunsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationRunsResponse copyWith(
          void Function(ListAutomationRunsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListAutomationRunsResponse))
          as ListAutomationRunsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAutomationRunsResponse() / ListAutomationRunsResponse.new instead')
  static ListAutomationRunsResponse create() => ListAutomationRunsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAutomationRunsResponse._();
  @$core.override
  ListAutomationRunsResponse createEmptyInstance() =>
      ListAutomationRunsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAutomationRunsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAutomationRunsResponse>(
          ListAutomationRunsResponse.$_createMessage);
  static ListAutomationRunsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AutomationRun> get runs => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);
}

class GetAutomationRunRequest extends $pb.GeneratedMessage {
  factory GetAutomationRunRequest({
    $core.String? id,
  }) {
    final result = GetAutomationRunRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetAutomationRunRequest._();

  factory GetAutomationRunRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationRunRequest()..mergeFromBuffer(data, registry);
  factory GetAutomationRunRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationRunRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAutomationRunRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAutomationRunRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationRunRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationRunRequest copyWith(
          void Function(GetAutomationRunRequest) updates) =>
      super.copyWith((message) => updates(message as GetAutomationRunRequest))
          as GetAutomationRunRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAutomationRunRequest() / GetAutomationRunRequest.new instead')
  static GetAutomationRunRequest create() => GetAutomationRunRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAutomationRunRequest._();
  @$core.override
  GetAutomationRunRequest createEmptyInstance() => GetAutomationRunRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAutomationRunRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAutomationRunRequest>(
          GetAutomationRunRequest.$_createMessage);
  static GetAutomationRunRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetAutomationRunResponse extends $pb.GeneratedMessage {
  factory GetAutomationRunResponse({
    AutomationRun? run,
  }) {
    final result = GetAutomationRunResponse._();
    if (run != null) result.run = run;
    return result;
  }

  GetAutomationRunResponse._();

  factory GetAutomationRunResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationRunResponse()..mergeFromBuffer(data, registry);
  factory GetAutomationRunResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAutomationRunResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAutomationRunResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAutomationRunResponse.$_createMessage)
    ..aOM<AutomationRun>(1, _omitFieldNames ? '' : 'run',
        subBuilder: AutomationRun.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationRunResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAutomationRunResponse copyWith(
          void Function(GetAutomationRunResponse) updates) =>
      super.copyWith((message) => updates(message as GetAutomationRunResponse))
          as GetAutomationRunResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAutomationRunResponse() / GetAutomationRunResponse.new instead')
  static GetAutomationRunResponse create() => GetAutomationRunResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetAutomationRunResponse._();
  @$core.override
  GetAutomationRunResponse createEmptyInstance() =>
      GetAutomationRunResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAutomationRunResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAutomationRunResponse>(
          GetAutomationRunResponse.$_createMessage);
  static GetAutomationRunResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AutomationRun get run => $_getN(0);
  @$pb.TagNumber(1)
  set run(AutomationRun value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRun() => $_has(0);
  @$pb.TagNumber(1)
  void clearRun() => $_clearField(1);
  @$pb.TagNumber(1)
  AutomationRun ensureRun() => $_ensure(0);
}

class ListAutomationStepRunsRequest extends $pb.GeneratedMessage {
  factory ListAutomationStepRunsRequest({
    $core.String? runId,
  }) {
    final result = ListAutomationStepRunsRequest._();
    if (runId != null) result.runId = runId;
    return result;
  }

  ListAutomationStepRunsRequest._();

  factory ListAutomationStepRunsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationStepRunsRequest()..mergeFromBuffer(data, registry);
  factory ListAutomationStepRunsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationStepRunsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAutomationStepRunsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAutomationStepRunsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'runId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationStepRunsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationStepRunsRequest copyWith(
          void Function(ListAutomationStepRunsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListAutomationStepRunsRequest))
          as ListAutomationStepRunsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAutomationStepRunsRequest() / ListAutomationStepRunsRequest.new instead')
  static ListAutomationStepRunsRequest create() =>
      ListAutomationStepRunsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAutomationStepRunsRequest._();
  @$core.override
  ListAutomationStepRunsRequest createEmptyInstance() =>
      ListAutomationStepRunsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAutomationStepRunsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAutomationStepRunsRequest>(
          ListAutomationStepRunsRequest.$_createMessage);
  static ListAutomationStepRunsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get runId => $_getSZ(0);
  @$pb.TagNumber(1)
  set runId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRunId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRunId() => $_clearField(1);
}

class ListAutomationStepRunsResponse extends $pb.GeneratedMessage {
  factory ListAutomationStepRunsResponse({
    $core.Iterable<AutomationStepRun>? stepRuns,
  }) {
    final result = ListAutomationStepRunsResponse._();
    if (stepRuns != null) result.stepRuns.addAll(stepRuns);
    return result;
  }

  ListAutomationStepRunsResponse._();

  factory ListAutomationStepRunsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationStepRunsResponse()..mergeFromBuffer(data, registry);
  factory ListAutomationStepRunsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAutomationStepRunsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAutomationStepRunsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAutomationStepRunsResponse.$_createMessage)
    ..pPM<AutomationStepRun>(1, _omitFieldNames ? '' : 'stepRuns',
        subBuilder: AutomationStepRun.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationStepRunsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAutomationStepRunsResponse copyWith(
          void Function(ListAutomationStepRunsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListAutomationStepRunsResponse))
          as ListAutomationStepRunsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAutomationStepRunsResponse() / ListAutomationStepRunsResponse.new instead')
  static ListAutomationStepRunsResponse create() =>
      ListAutomationStepRunsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAutomationStepRunsResponse._();
  @$core.override
  ListAutomationStepRunsResponse createEmptyInstance() =>
      ListAutomationStepRunsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAutomationStepRunsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAutomationStepRunsResponse>(
          ListAutomationStepRunsResponse.$_createMessage);
  static ListAutomationStepRunsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AutomationStepRun> get stepRuns => $_getList(0);
}

class AutomationServiceApi {
  final $pb.RpcClient _client;

  AutomationServiceApi(this._client);

  $async.Future<ListAutomationsResponse> listAutomations(
          $pb.ClientContext? ctx, ListAutomationsRequest request) =>
      _client.invoke<ListAutomationsResponse>(ctx, 'AutomationService',
          'ListAutomations', request, ListAutomationsResponse());
  $async.Future<GetAutomationResponse> getAutomation(
          $pb.ClientContext? ctx, GetAutomationRequest request) =>
      _client.invoke<GetAutomationResponse>(ctx, 'AutomationService',
          'GetAutomation', request, GetAutomationResponse());
  $async.Future<CreateAutomationResponse> createAutomation(
          $pb.ClientContext? ctx, CreateAutomationRequest request) =>
      _client.invoke<CreateAutomationResponse>(ctx, 'AutomationService',
          'CreateAutomation', request, CreateAutomationResponse());
  $async.Future<UpdateAutomationResponse> updateAutomation(
          $pb.ClientContext? ctx, UpdateAutomationRequest request) =>
      _client.invoke<UpdateAutomationResponse>(ctx, 'AutomationService',
          'UpdateAutomation', request, UpdateAutomationResponse());
  $async.Future<DeleteAutomationResponse> deleteAutomation(
          $pb.ClientContext? ctx, DeleteAutomationRequest request) =>
      _client.invoke<DeleteAutomationResponse>(ctx, 'AutomationService',
          'DeleteAutomation', request, DeleteAutomationResponse());
  $async.Future<RunAutomationNowResponse> runAutomationNow(
          $pb.ClientContext? ctx, RunAutomationNowRequest request) =>
      _client.invoke<RunAutomationNowResponse>(ctx, 'AutomationService',
          'RunAutomationNow', request, RunAutomationNowResponse());
  $async.Future<ListAutomationRunsResponse> listAutomationRuns(
          $pb.ClientContext? ctx, ListAutomationRunsRequest request) =>
      _client.invoke<ListAutomationRunsResponse>(ctx, 'AutomationService',
          'ListAutomationRuns', request, ListAutomationRunsResponse());
  $async.Future<GetAutomationRunResponse> getAutomationRun(
          $pb.ClientContext? ctx, GetAutomationRunRequest request) =>
      _client.invoke<GetAutomationRunResponse>(ctx, 'AutomationService',
          'GetAutomationRun', request, GetAutomationRunResponse());
  $async.Future<ListAutomationStepRunsResponse> listAutomationStepRuns(
          $pb.ClientContext? ctx, ListAutomationStepRunsRequest request) =>
      _client.invoke<ListAutomationStepRunsResponse>(ctx, 'AutomationService',
          'ListAutomationStepRuns', request, ListAutomationStepRunsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
