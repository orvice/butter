// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_service.proto.

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
    as $7;
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart' as $6;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $1;

import 'agent.pb.dart' as $0;
import 'agent_operation.pb.dart' as $3;
import 'agent_service.pbenum.dart';
import 'agentchannel.pb.dart' as $5;
import 'content.pb.dart' as $2;
import 'dashboard.pb.dart' as $8;
import 'repobinding.pb.dart' as $4;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'agent_service.pbenum.dart';

class ListAgentsRequest extends $pb.GeneratedMessage {
  factory ListAgentsRequest({
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = ListAgentsRequest._();
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListAgentsRequest._();

  factory ListAgentsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentsRequest()..mergeFromBuffer(data, registry);
  factory ListAgentsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentsRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'pageSize')
    ..aOS(2, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentsRequest copyWith(void Function(ListAgentsRequest) updates) =>
      super.copyWith((message) => updates(message as ListAgentsRequest))
          as ListAgentsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListAgentsRequest() / ListAgentsRequest.new instead')
  static ListAgentsRequest create() => ListAgentsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListAgentsRequest._();
  @$core.override
  ListAgentsRequest createEmptyInstance() => ListAgentsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAgentsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListAgentsRequest>(
          ListAgentsRequest.$_createMessage);
  static ListAgentsRequest? _defaultInstance;

  /// Maximum records per page. Defaults to 50.
  @$pb.TagNumber(1)
  $core.int get pageSize => $_getIZ(0);
  @$pb.TagNumber(1)
  set pageSize($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPageSize() => $_has(0);
  @$pb.TagNumber(1)
  void clearPageSize() => $_clearField(1);

  /// Pagination token from a previous response.
  @$pb.TagNumber(2)
  $core.String get pageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set pageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageToken() => $_clearField(2);
}

class ListAgentsResponse extends $pb.GeneratedMessage {
  factory ListAgentsResponse({
    $core.Iterable<$0.Agent>? agents,
    $core.String? nextPageToken,
    $core.int? total,
  }) {
    final result = ListAgentsResponse._();
    if (agents != null) result.agents.addAll(agents);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    if (total != null) result.total = total;
    return result;
  }

  ListAgentsResponse._();

  factory ListAgentsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentsResponse()..mergeFromBuffer(data, registry);
  factory ListAgentsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentsResponse.$_createMessage)
    ..pPM<$0.Agent>(1, _omitFieldNames ? '' : 'agents',
        subBuilder: $0.Agent.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..aI(3, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentsResponse copyWith(void Function(ListAgentsResponse) updates) =>
      super.copyWith((message) => updates(message as ListAgentsResponse))
          as ListAgentsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListAgentsResponse() / ListAgentsResponse.new instead')
  static ListAgentsResponse create() => ListAgentsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListAgentsResponse._();
  @$core.override
  ListAgentsResponse createEmptyInstance() => ListAgentsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAgentsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentsResponse>(
          ListAgentsResponse.$_createMessage);
  static ListAgentsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$0.Agent> get agents => $_getList(0);

  /// Token for the next page; empty if there are no more results.
  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);

  /// Total agents across all pages.
  @$pb.TagNumber(3)
  $core.int get total => $_getIZ(2);
  @$pb.TagNumber(3)
  set total($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTotal() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotal() => $_clearField(3);
}

class ReloadAgentsRequest extends $pb.GeneratedMessage {
  factory ReloadAgentsRequest() => ReloadAgentsRequest._();

  ReloadAgentsRequest._();

  factory ReloadAgentsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReloadAgentsRequest()..mergeFromBuffer(data, registry);
  factory ReloadAgentsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReloadAgentsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReloadAgentsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ReloadAgentsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReloadAgentsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReloadAgentsRequest copyWith(void Function(ReloadAgentsRequest) updates) =>
      super.copyWith((message) => updates(message as ReloadAgentsRequest))
          as ReloadAgentsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ReloadAgentsRequest() / ReloadAgentsRequest.new instead')
  static ReloadAgentsRequest create() => ReloadAgentsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ReloadAgentsRequest._();
  @$core.override
  ReloadAgentsRequest createEmptyInstance() => ReloadAgentsRequest._();
  @$core.pragma('dart2js:noInline')
  static ReloadAgentsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReloadAgentsRequest>(
          ReloadAgentsRequest.$_createMessage);
  static ReloadAgentsRequest? _defaultInstance;
}

class ReloadAgentsResponse extends $pb.GeneratedMessage {
  factory ReloadAgentsResponse({
    $1.Timestamp? reloadedAt,
  }) {
    final result = ReloadAgentsResponse._();
    if (reloadedAt != null) result.reloadedAt = reloadedAt;
    return result;
  }

  ReloadAgentsResponse._();

  factory ReloadAgentsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReloadAgentsResponse()..mergeFromBuffer(data, registry);
  factory ReloadAgentsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReloadAgentsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReloadAgentsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ReloadAgentsResponse.$_createMessage)
    ..aOM<$1.Timestamp>(1, _omitFieldNames ? '' : 'reloadedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReloadAgentsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReloadAgentsResponse copyWith(void Function(ReloadAgentsResponse) updates) =>
      super.copyWith((message) => updates(message as ReloadAgentsResponse))
          as ReloadAgentsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReloadAgentsResponse() / ReloadAgentsResponse.new instead')
  static ReloadAgentsResponse create() => ReloadAgentsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ReloadAgentsResponse._();
  @$core.override
  ReloadAgentsResponse createEmptyInstance() => ReloadAgentsResponse._();
  @$core.pragma('dart2js:noInline')
  static ReloadAgentsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReloadAgentsResponse>(
          ReloadAgentsResponse.$_createMessage);
  static ReloadAgentsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $1.Timestamp get reloadedAt => $_getN(0);
  @$pb.TagNumber(1)
  set reloadedAt($1.Timestamp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasReloadedAt() => $_has(0);
  @$pb.TagNumber(1)
  void clearReloadedAt() => $_clearField(1);
  @$pb.TagNumber(1)
  $1.Timestamp ensureReloadedAt() => $_ensure(0);
}

class AgentRuntimeStatus extends $pb.GeneratedMessage {
  factory AgentRuntimeStatus({
    $core.String? name,
    AgentRuntimeState? state,
    $1.Timestamp? lastRunAt,
    $core.String? lastInvocationId,
    $core.int? inFlight,
    $core.String? agentId,
  }) {
    final result = AgentRuntimeStatus._();
    if (name != null) result.name = name;
    if (state != null) result.state = state;
    if (lastRunAt != null) result.lastRunAt = lastRunAt;
    if (lastInvocationId != null) result.lastInvocationId = lastInvocationId;
    if (inFlight != null) result.inFlight = inFlight;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  AgentRuntimeStatus._();

  factory AgentRuntimeStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentRuntimeStatus()..mergeFromBuffer(data, registry);
  factory AgentRuntimeStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentRuntimeStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentRuntimeStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentRuntimeStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aE<AgentRuntimeState>(2, _omitFieldNames ? '' : 'state',
        enumValues: AgentRuntimeState.values)
    ..aOM<$1.Timestamp>(3, _omitFieldNames ? '' : 'lastRunAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOS(4, _omitFieldNames ? '' : 'lastInvocationId')
    ..aI(5, _omitFieldNames ? '' : 'inFlight')
    ..aOS(6, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentRuntimeStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentRuntimeStatus copyWith(void Function(AgentRuntimeStatus) updates) =>
      super.copyWith((message) => updates(message as AgentRuntimeStatus))
          as AgentRuntimeStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentRuntimeStatus() / AgentRuntimeStatus.new instead')
  static AgentRuntimeStatus create() => AgentRuntimeStatus._();
  static $pb.GeneratedMessage $_createMessage() => AgentRuntimeStatus._();
  @$core.override
  AgentRuntimeStatus createEmptyInstance() => AgentRuntimeStatus._();
  @$core.pragma('dart2js:noInline')
  static AgentRuntimeStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentRuntimeStatus>(
          AgentRuntimeStatus.$_createMessage);
  static AgentRuntimeStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  AgentRuntimeState get state => $_getN(1);
  @$pb.TagNumber(2)
  set state(AgentRuntimeState value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => $_clearField(2);

  @$pb.TagNumber(3)
  $1.Timestamp get lastRunAt => $_getN(2);
  @$pb.TagNumber(3)
  set lastRunAt($1.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasLastRunAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearLastRunAt() => $_clearField(3);
  @$pb.TagNumber(3)
  $1.Timestamp ensureLastRunAt() => $_ensure(2);

  /// ID of the latest invocation, if any.
  @$pb.TagNumber(4)
  $core.String get lastInvocationId => $_getSZ(3);
  @$pb.TagNumber(4)
  set lastInvocationId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLastInvocationId() => $_has(3);
  @$pb.TagNumber(4)
  void clearLastInvocationId() => $_clearField(4);

  /// Number of in-flight (RUNNING) invocations for this agent.
  @$pb.TagNumber(5)
  $core.int get inFlight => $_getIZ(4);
  @$pb.TagNumber(5)
  set inFlight($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInFlight() => $_has(4);
  @$pb.TagNumber(5)
  void clearInFlight() => $_clearField(5);

  /// Immutable agent_id of the agent, when assigned.
  @$pb.TagNumber(6)
  $core.String get agentId => $_getSZ(5);
  @$pb.TagNumber(6)
  set agentId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAgentId() => $_has(5);
  @$pb.TagNumber(6)
  void clearAgentId() => $_clearField(6);
}

class GetAgentRuntimeStatusRequest extends $pb.GeneratedMessage {
  factory GetAgentRuntimeStatusRequest({
    @$core.Deprecated('This field is deprecated.') $core.String? name,
    $core.String? agentId,
  }) {
    final result = GetAgentRuntimeStatusRequest._();
    if (name != null) result.name = name;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  GetAgentRuntimeStatusRequest._();

  factory GetAgentRuntimeStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentRuntimeStatusRequest()..mergeFromBuffer(data, registry);
  factory GetAgentRuntimeStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentRuntimeStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentRuntimeStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentRuntimeStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentRuntimeStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentRuntimeStatusRequest copyWith(
          void Function(GetAgentRuntimeStatusRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetAgentRuntimeStatusRequest))
          as GetAgentRuntimeStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentRuntimeStatusRequest() / GetAgentRuntimeStatusRequest.new instead')
  static GetAgentRuntimeStatusRequest create() =>
      GetAgentRuntimeStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAgentRuntimeStatusRequest._();
  @$core.override
  GetAgentRuntimeStatusRequest createEmptyInstance() =>
      GetAgentRuntimeStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAgentRuntimeStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentRuntimeStatusRequest>(
          GetAgentRuntimeStatusRequest.$_createMessage);
  static GetAgentRuntimeStatusRequest? _defaultInstance;

  /// Legacy name lookup. Deprecated (issue #241): names never select an
  /// Agent; requests without agent_id are rejected.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Immutable agent_id of the agent. Required.
  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);
}

class GetAgentRuntimeStatusResponse extends $pb.GeneratedMessage {
  factory GetAgentRuntimeStatusResponse({
    AgentRuntimeStatus? status,
  }) {
    final result = GetAgentRuntimeStatusResponse._();
    if (status != null) result.status = status;
    return result;
  }

  GetAgentRuntimeStatusResponse._();

  factory GetAgentRuntimeStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentRuntimeStatusResponse()..mergeFromBuffer(data, registry);
  factory GetAgentRuntimeStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentRuntimeStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentRuntimeStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentRuntimeStatusResponse.$_createMessage)
    ..aOM<AgentRuntimeStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: AgentRuntimeStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentRuntimeStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentRuntimeStatusResponse copyWith(
          void Function(GetAgentRuntimeStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetAgentRuntimeStatusResponse))
          as GetAgentRuntimeStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentRuntimeStatusResponse() / GetAgentRuntimeStatusResponse.new instead')
  static GetAgentRuntimeStatusResponse create() =>
      GetAgentRuntimeStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAgentRuntimeStatusResponse._();
  @$core.override
  GetAgentRuntimeStatusResponse createEmptyInstance() =>
      GetAgentRuntimeStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAgentRuntimeStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentRuntimeStatusResponse>(
          GetAgentRuntimeStatusResponse.$_createMessage);
  static GetAgentRuntimeStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AgentRuntimeStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(AgentRuntimeStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentRuntimeStatus ensureStatus() => $_ensure(0);
}

class ListAgentRuntimeStatusesRequest extends $pb.GeneratedMessage {
  factory ListAgentRuntimeStatusesRequest({
    @$core.Deprecated('This field is deprecated.')
    $core.Iterable<$core.String>? names,
    $core.Iterable<$core.String>? agentIds,
  }) {
    final result = ListAgentRuntimeStatusesRequest._();
    if (names != null) result.names.addAll(names);
    if (agentIds != null) result.agentIds.addAll(agentIds);
    return result;
  }

  ListAgentRuntimeStatusesRequest._();

  factory ListAgentRuntimeStatusesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentRuntimeStatusesRequest()..mergeFromBuffer(data, registry);
  factory ListAgentRuntimeStatusesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentRuntimeStatusesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentRuntimeStatusesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentRuntimeStatusesRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'names')
    ..pPS(2, _omitFieldNames ? '' : 'agentIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentRuntimeStatusesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentRuntimeStatusesRequest copyWith(
          void Function(ListAgentRuntimeStatusesRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentRuntimeStatusesRequest))
          as ListAgentRuntimeStatusesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentRuntimeStatusesRequest() / ListAgentRuntimeStatusesRequest.new instead')
  static ListAgentRuntimeStatusesRequest create() =>
      ListAgentRuntimeStatusesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentRuntimeStatusesRequest._();
  @$core.override
  ListAgentRuntimeStatusesRequest createEmptyInstance() =>
      ListAgentRuntimeStatusesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAgentRuntimeStatusesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentRuntimeStatusesRequest>(
          ListAgentRuntimeStatusesRequest.$_createMessage);
  static ListAgentRuntimeStatusesRequest? _defaultInstance;

  /// Legacy name filter. Deprecated (issue #241): names never select Agents;
  /// requests carrying names are rejected.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get names => $_getList(0);

  /// Filter by immutable agent_ids. Empty returns statuses for all agents.
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get agentIds => $_getList(1);
}

class ListAgentRuntimeStatusesResponse extends $pb.GeneratedMessage {
  factory ListAgentRuntimeStatusesResponse({
    $core.Iterable<AgentRuntimeStatus>? statuses,
  }) {
    final result = ListAgentRuntimeStatusesResponse._();
    if (statuses != null) result.statuses.addAll(statuses);
    return result;
  }

  ListAgentRuntimeStatusesResponse._();

  factory ListAgentRuntimeStatusesResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentRuntimeStatusesResponse()..mergeFromBuffer(data, registry);
  factory ListAgentRuntimeStatusesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentRuntimeStatusesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentRuntimeStatusesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentRuntimeStatusesResponse.$_createMessage)
    ..pPM<AgentRuntimeStatus>(1, _omitFieldNames ? '' : 'statuses',
        subBuilder: AgentRuntimeStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentRuntimeStatusesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentRuntimeStatusesResponse copyWith(
          void Function(ListAgentRuntimeStatusesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentRuntimeStatusesResponse))
          as ListAgentRuntimeStatusesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentRuntimeStatusesResponse() / ListAgentRuntimeStatusesResponse.new instead')
  static ListAgentRuntimeStatusesResponse create() =>
      ListAgentRuntimeStatusesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentRuntimeStatusesResponse._();
  @$core.override
  ListAgentRuntimeStatusesResponse createEmptyInstance() =>
      ListAgentRuntimeStatusesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAgentRuntimeStatusesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentRuntimeStatusesResponse>(
          ListAgentRuntimeStatusesResponse.$_createMessage);
  static ListAgentRuntimeStatusesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AgentRuntimeStatus> get statuses => $_getList(0);
}

class CancelAgentInvocationRequest extends $pb.GeneratedMessage {
  factory CancelAgentInvocationRequest({
    $core.String? invocationId,
  }) {
    final result = CancelAgentInvocationRequest._();
    if (invocationId != null) result.invocationId = invocationId;
    return result;
  }

  CancelAgentInvocationRequest._();

  factory CancelAgentInvocationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelAgentInvocationRequest()..mergeFromBuffer(data, registry);
  factory CancelAgentInvocationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelAgentInvocationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelAgentInvocationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CancelAgentInvocationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invocationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelAgentInvocationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelAgentInvocationRequest copyWith(
          void Function(CancelAgentInvocationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CancelAgentInvocationRequest))
          as CancelAgentInvocationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelAgentInvocationRequest() / CancelAgentInvocationRequest.new instead')
  static CancelAgentInvocationRequest create() =>
      CancelAgentInvocationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CancelAgentInvocationRequest._();
  @$core.override
  CancelAgentInvocationRequest createEmptyInstance() =>
      CancelAgentInvocationRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelAgentInvocationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelAgentInvocationRequest>(
          CancelAgentInvocationRequest.$_createMessage);
  static CancelAgentInvocationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invocationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invocationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocationId() => $_clearField(1);
}

class CancelAgentInvocationResponse extends $pb.GeneratedMessage {
  factory CancelAgentInvocationResponse({
    $core.bool? cancelled,
  }) {
    final result = CancelAgentInvocationResponse._();
    if (cancelled != null) result.cancelled = cancelled;
    return result;
  }

  CancelAgentInvocationResponse._();

  factory CancelAgentInvocationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelAgentInvocationResponse()..mergeFromBuffer(data, registry);
  factory CancelAgentInvocationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelAgentInvocationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelAgentInvocationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CancelAgentInvocationResponse.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'cancelled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelAgentInvocationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelAgentInvocationResponse copyWith(
          void Function(CancelAgentInvocationResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CancelAgentInvocationResponse))
          as CancelAgentInvocationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelAgentInvocationResponse() / CancelAgentInvocationResponse.new instead')
  static CancelAgentInvocationResponse create() =>
      CancelAgentInvocationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CancelAgentInvocationResponse._();
  @$core.override
  CancelAgentInvocationResponse createEmptyInstance() =>
      CancelAgentInvocationResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelAgentInvocationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelAgentInvocationResponse>(
          CancelAgentInvocationResponse.$_createMessage);
  static CancelAgentInvocationResponse? _defaultInstance;

  /// True if the invocation was found and the cancel signal was delivered.
  @$pb.TagNumber(1)
  $core.bool get cancelled => $_getBF(0);
  @$pb.TagNumber(1)
  set cancelled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCancelled() => $_has(0);
  @$pb.TagNumber(1)
  void clearCancelled() => $_clearField(1);
}

class StreamAgentRequest extends $pb.GeneratedMessage {
  factory StreamAgentRequest({
    @$core.Deprecated('This field is deprecated.') $core.String? agentName,
    $core.String? message,
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $core.String? modelOverride,
    $core.Iterable<$2.InputPart>? parts,
    $core.String? agentId,
  }) {
    final result = StreamAgentRequest._();
    if (agentName != null) result.agentName = agentName;
    if (message != null) result.message = message;
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (modelOverride != null) result.modelOverride = modelOverride;
    if (parts != null) result.parts.addAll(parts);
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  StreamAgentRequest._();

  factory StreamAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentRequest()..mergeFromBuffer(data, registry);
  factory StreamAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StreamAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StreamAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentName')
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..aOS(3, _omitFieldNames ? '' : 'appName')
    ..aOS(4, _omitFieldNames ? '' : 'userId')
    ..aOS(5, _omitFieldNames ? '' : 'sessionId')
    ..aOS(6, _omitFieldNames ? '' : 'modelOverride')
    ..pPM<$2.InputPart>(7, _omitFieldNames ? '' : 'parts',
        subBuilder: $2.InputPart.$_createMessage)
    ..aOS(8, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentRequest copyWith(void Function(StreamAgentRequest) updates) =>
      super.copyWith((message) => updates(message as StreamAgentRequest))
          as StreamAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StreamAgentRequest() / StreamAgentRequest.new instead')
  static StreamAgentRequest create() => StreamAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => StreamAgentRequest._();
  @$core.override
  StreamAgentRequest createEmptyInstance() => StreamAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static StreamAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StreamAgentRequest>(
          StreamAgentRequest.$_createMessage);
  static StreamAgentRequest? _defaultInstance;

  /// Legacy agent name. Deprecated (issue #241): never resolved as an input;
  /// agent_id is the sole agent reference.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get agentName => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set agentName($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasAgentName() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearAgentName() => $_clearField(1);

  /// User prompt to send to the agent. Required, non-empty.
  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => $_clearField(2);

  /// ADK app name. Defaults to "api" when empty.
  @$pb.TagNumber(3)
  $core.String get appName => $_getSZ(2);
  @$pb.TagNumber(3)
  set appName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAppName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAppName() => $_clearField(3);

  /// ADK user id. Defaults to "api" when empty.
  @$pb.TagNumber(4)
  $core.String get userId => $_getSZ(3);
  @$pb.TagNumber(4)
  set userId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasUserId() => $_has(3);
  @$pb.TagNumber(4)
  void clearUserId() => $_clearField(4);

  /// Existing session id, or empty to start an ephemeral session.
  @$pb.TagNumber(5)
  $core.String get sessionId => $_getSZ(4);
  @$pb.TagNumber(5)
  set sessionId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSessionId() => $_has(4);
  @$pb.TagNumber(5)
  void clearSessionId() => $_clearField(5);

  /// Optional model override; if set the agent runs with this model
  /// instead of the configured default.
  @$pb.TagNumber(6)
  $core.String get modelOverride => $_getSZ(5);
  @$pb.TagNumber(6)
  set modelOverride($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasModelOverride() => $_has(5);
  @$pb.TagNumber(6)
  void clearModelOverride() => $_clearField(6);

  /// Multimodal input parts (text and inline images). When non-empty,
  /// `parts` is used as the user input and `message` is ignored. When
  /// empty, `message` is used as before. Validation (MIME whitelist,
  /// size and count limits) happens in the application layer.
  @$pb.TagNumber(7)
  $pb.PbList<$2.InputPart> get parts => $_getList(6);

  /// Immutable agent_id of the agent to invoke. Preferred over agent_name.
  @$pb.TagNumber(8)
  $core.String get agentId => $_getSZ(7);
  @$pb.TagNumber(8)
  set agentId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAgentId() => $_has(7);
  @$pb.TagNumber(8)
  void clearAgentId() => $_clearField(8);
}

enum StreamAgentResponse_Event { started, runEvent, textDelta, final_4, notSet }

/// StreamAgentResponse is sent for every observable transition during
/// the agent invocation. The oneof guarantees clients see exactly one
/// variant per message — no ad-hoc JSON payload disambiguation.
class StreamAgentResponse extends $pb.GeneratedMessage {
  factory StreamAgentResponse({
    StreamAgentStarted? started,
    StreamAgentRunEvent? runEvent,
    StreamAgentTextDelta? textDelta,
    StreamAgentFinal? final_4,
  }) {
    final result = StreamAgentResponse._();
    if (started != null) result.started = started;
    if (runEvent != null) result.runEvent = runEvent;
    if (textDelta != null) result.textDelta = textDelta;
    if (final_4 != null) result.final_4 = final_4;
    return result;
  }

  StreamAgentResponse._();

  factory StreamAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentResponse()..mergeFromBuffer(data, registry);
  factory StreamAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentResponse()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, StreamAgentResponse_Event>
      _StreamAgentResponse_EventByTag = {
    1: StreamAgentResponse_Event.started,
    2: StreamAgentResponse_Event.runEvent,
    3: StreamAgentResponse_Event.textDelta,
    4: StreamAgentResponse_Event.final_4,
    0: StreamAgentResponse_Event.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StreamAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StreamAgentResponse.$_createMessage)
    ..oo(0, [1, 2, 3, 4])
    ..aOM<StreamAgentStarted>(1, _omitFieldNames ? '' : 'started',
        subBuilder: StreamAgentStarted.$_createMessage)
    ..aOM<StreamAgentRunEvent>(2, _omitFieldNames ? '' : 'runEvent',
        subBuilder: StreamAgentRunEvent.$_createMessage)
    ..aOM<StreamAgentTextDelta>(3, _omitFieldNames ? '' : 'textDelta',
        subBuilder: StreamAgentTextDelta.$_createMessage)
    ..aOM<StreamAgentFinal>(4, _omitFieldNames ? '' : 'final',
        subBuilder: StreamAgentFinal.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentResponse copyWith(void Function(StreamAgentResponse) updates) =>
      super.copyWith((message) => updates(message as StreamAgentResponse))
          as StreamAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use StreamAgentResponse() / StreamAgentResponse.new instead')
  static StreamAgentResponse create() => StreamAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => StreamAgentResponse._();
  @$core.override
  StreamAgentResponse createEmptyInstance() => StreamAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static StreamAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StreamAgentResponse>(
          StreamAgentResponse.$_createMessage);
  static StreamAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  @$pb.TagNumber(3)
  @$pb.TagNumber(4)
  StreamAgentResponse_Event whichEvent() =>
      _StreamAgentResponse_EventByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  @$pb.TagNumber(3)
  @$pb.TagNumber(4)
  void clearEvent() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  StreamAgentStarted get started => $_getN(0);
  @$pb.TagNumber(1)
  set started(StreamAgentStarted value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStarted() => $_has(0);
  @$pb.TagNumber(1)
  void clearStarted() => $_clearField(1);
  @$pb.TagNumber(1)
  StreamAgentStarted ensureStarted() => $_ensure(0);

  @$pb.TagNumber(2)
  StreamAgentRunEvent get runEvent => $_getN(1);
  @$pb.TagNumber(2)
  set runEvent(StreamAgentRunEvent value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRunEvent() => $_has(1);
  @$pb.TagNumber(2)
  void clearRunEvent() => $_clearField(2);
  @$pb.TagNumber(2)
  StreamAgentRunEvent ensureRunEvent() => $_ensure(1);

  @$pb.TagNumber(3)
  StreamAgentTextDelta get textDelta => $_getN(2);
  @$pb.TagNumber(3)
  set textDelta(StreamAgentTextDelta value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasTextDelta() => $_has(2);
  @$pb.TagNumber(3)
  void clearTextDelta() => $_clearField(3);
  @$pb.TagNumber(3)
  StreamAgentTextDelta ensureTextDelta() => $_ensure(2);

  @$pb.TagNumber(4)
  StreamAgentFinal get final_4 => $_getN(3);
  @$pb.TagNumber(4)
  set final_4(StreamAgentFinal value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasFinal_4() => $_has(3);
  @$pb.TagNumber(4)
  void clearFinal_4() => $_clearField(4);
  @$pb.TagNumber(4)
  StreamAgentFinal ensureFinal_4() => $_ensure(3);
}

/// StreamAgentStarted is the first event in the stream; it carries the
/// server-assigned invocation_id so callers can correlate with
/// CancelAgentInvocation.
class StreamAgentStarted extends $pb.GeneratedMessage {
  factory StreamAgentStarted({
    $core.String? invocationId,
    $core.String? sessionId,
    $core.String? agentName,
    $core.String? agentId,
  }) {
    final result = StreamAgentStarted._();
    if (invocationId != null) result.invocationId = invocationId;
    if (sessionId != null) result.sessionId = sessionId;
    if (agentName != null) result.agentName = agentName;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  StreamAgentStarted._();

  factory StreamAgentStarted.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentStarted()..mergeFromBuffer(data, registry);
  factory StreamAgentStarted.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentStarted()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StreamAgentStarted',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StreamAgentStarted.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invocationId')
    ..aOS(2, _omitFieldNames ? '' : 'sessionId')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aOS(4, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentStarted clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentStarted copyWith(void Function(StreamAgentStarted) updates) =>
      super.copyWith((message) => updates(message as StreamAgentStarted))
          as StreamAgentStarted;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StreamAgentStarted() / StreamAgentStarted.new instead')
  static StreamAgentStarted create() => StreamAgentStarted._();
  static $pb.GeneratedMessage $_createMessage() => StreamAgentStarted._();
  @$core.override
  StreamAgentStarted createEmptyInstance() => StreamAgentStarted._();
  @$core.pragma('dart2js:noInline')
  static StreamAgentStarted getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StreamAgentStarted>(
          StreamAgentStarted.$_createMessage);
  static StreamAgentStarted? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invocationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invocationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocationId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get sessionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set sessionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSessionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSessionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  /// Immutable agent_id of the resolved agent, when assigned.
  @$pb.TagNumber(4)
  $core.String get agentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set agentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAgentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearAgentId() => $_clearField(4);
}

/// StreamAgentRunEvent mirrors a single ADK session.Event. content_json
/// is the genai.Content payload serialized as JSON; the dashboard parses
/// it client-side to render tool calls / function responses / etc.
class StreamAgentRunEvent extends $pb.GeneratedMessage {
  factory StreamAgentRunEvent({
    $core.String? invocationId,
    $core.String? sessionId,
    $core.String? agentName,
    $core.String? eventId,
    $core.String? author,
    $core.String? branch,
    $core.bool? partial,
    $core.bool? finalResponse,
    $core.String? contentJson,
    $1.Timestamp? timestamp,
    $core.String? agentId,
  }) {
    final result = StreamAgentRunEvent._();
    if (invocationId != null) result.invocationId = invocationId;
    if (sessionId != null) result.sessionId = sessionId;
    if (agentName != null) result.agentName = agentName;
    if (eventId != null) result.eventId = eventId;
    if (author != null) result.author = author;
    if (branch != null) result.branch = branch;
    if (partial != null) result.partial = partial;
    if (finalResponse != null) result.finalResponse = finalResponse;
    if (contentJson != null) result.contentJson = contentJson;
    if (timestamp != null) result.timestamp = timestamp;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  StreamAgentRunEvent._();

  factory StreamAgentRunEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentRunEvent()..mergeFromBuffer(data, registry);
  factory StreamAgentRunEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentRunEvent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StreamAgentRunEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StreamAgentRunEvent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invocationId')
    ..aOS(2, _omitFieldNames ? '' : 'sessionId')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aOS(4, _omitFieldNames ? '' : 'eventId')
    ..aOS(5, _omitFieldNames ? '' : 'author')
    ..aOS(6, _omitFieldNames ? '' : 'branch')
    ..aOB(7, _omitFieldNames ? '' : 'partial')
    ..aOB(8, _omitFieldNames ? '' : 'finalResponse')
    ..aOS(9, _omitFieldNames ? '' : 'contentJson')
    ..aOM<$1.Timestamp>(10, _omitFieldNames ? '' : 'timestamp',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOS(11, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentRunEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentRunEvent copyWith(void Function(StreamAgentRunEvent) updates) =>
      super.copyWith((message) => updates(message as StreamAgentRunEvent))
          as StreamAgentRunEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use StreamAgentRunEvent() / StreamAgentRunEvent.new instead')
  static StreamAgentRunEvent create() => StreamAgentRunEvent._();
  static $pb.GeneratedMessage $_createMessage() => StreamAgentRunEvent._();
  @$core.override
  StreamAgentRunEvent createEmptyInstance() => StreamAgentRunEvent._();
  @$core.pragma('dart2js:noInline')
  static StreamAgentRunEvent getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StreamAgentRunEvent>(
          StreamAgentRunEvent.$_createMessage);
  static StreamAgentRunEvent? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invocationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invocationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocationId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get sessionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set sessionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSessionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSessionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get eventId => $_getSZ(3);
  @$pb.TagNumber(4)
  set eventId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEventId() => $_has(3);
  @$pb.TagNumber(4)
  void clearEventId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get author => $_getSZ(4);
  @$pb.TagNumber(5)
  set author($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAuthor() => $_has(4);
  @$pb.TagNumber(5)
  void clearAuthor() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get branch => $_getSZ(5);
  @$pb.TagNumber(6)
  set branch($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasBranch() => $_has(5);
  @$pb.TagNumber(6)
  void clearBranch() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get partial => $_getBF(6);
  @$pb.TagNumber(7)
  set partial($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPartial() => $_has(6);
  @$pb.TagNumber(7)
  void clearPartial() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get finalResponse => $_getBF(7);
  @$pb.TagNumber(8)
  set finalResponse($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasFinalResponse() => $_has(7);
  @$pb.TagNumber(8)
  void clearFinalResponse() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get contentJson => $_getSZ(8);
  @$pb.TagNumber(9)
  set contentJson($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasContentJson() => $_has(8);
  @$pb.TagNumber(9)
  void clearContentJson() => $_clearField(9);

  @$pb.TagNumber(10)
  $1.Timestamp get timestamp => $_getN(9);
  @$pb.TagNumber(10)
  set timestamp($1.Timestamp value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasTimestamp() => $_has(9);
  @$pb.TagNumber(10)
  void clearTimestamp() => $_clearField(10);
  @$pb.TagNumber(10)
  $1.Timestamp ensureTimestamp() => $_ensure(9);

  /// Immutable agent_id of the resolved agent, when assigned.
  @$pb.TagNumber(11)
  $core.String get agentId => $_getSZ(10);
  @$pb.TagNumber(11)
  set agentId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasAgentId() => $_has(10);
  @$pb.TagNumber(11)
  void clearAgentId() => $_clearField(11);
}

/// StreamAgentTextDelta is emitted for partial assistant-text chunks so
/// the UI can render the response as it streams in.
class StreamAgentTextDelta extends $pb.GeneratedMessage {
  factory StreamAgentTextDelta({
    $core.String? invocationId,
    $core.String? sessionId,
    $core.String? agentName,
    $core.String? text,
    $core.String? agentId,
  }) {
    final result = StreamAgentTextDelta._();
    if (invocationId != null) result.invocationId = invocationId;
    if (sessionId != null) result.sessionId = sessionId;
    if (agentName != null) result.agentName = agentName;
    if (text != null) result.text = text;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  StreamAgentTextDelta._();

  factory StreamAgentTextDelta.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentTextDelta()..mergeFromBuffer(data, registry);
  factory StreamAgentTextDelta.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentTextDelta()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StreamAgentTextDelta',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StreamAgentTextDelta.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invocationId')
    ..aOS(2, _omitFieldNames ? '' : 'sessionId')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aOS(4, _omitFieldNames ? '' : 'text')
    ..aOS(5, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentTextDelta clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentTextDelta copyWith(void Function(StreamAgentTextDelta) updates) =>
      super.copyWith((message) => updates(message as StreamAgentTextDelta))
          as StreamAgentTextDelta;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StreamAgentTextDelta() / StreamAgentTextDelta.new instead')
  static StreamAgentTextDelta create() => StreamAgentTextDelta._();
  static $pb.GeneratedMessage $_createMessage() => StreamAgentTextDelta._();
  @$core.override
  StreamAgentTextDelta createEmptyInstance() => StreamAgentTextDelta._();
  @$core.pragma('dart2js:noInline')
  static StreamAgentTextDelta getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StreamAgentTextDelta>(
          StreamAgentTextDelta.$_createMessage);
  static StreamAgentTextDelta? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invocationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invocationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocationId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get sessionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set sessionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSessionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSessionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get text => $_getSZ(3);
  @$pb.TagNumber(4)
  set text($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasText() => $_has(3);
  @$pb.TagNumber(4)
  void clearText() => $_clearField(4);

  /// Immutable agent_id of the resolved agent, when assigned.
  @$pb.TagNumber(5)
  $core.String get agentId => $_getSZ(4);
  @$pb.TagNumber(5)
  set agentId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAgentId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAgentId() => $_clearField(5);
}

/// StreamAgentFinal is sent once after the runner returns the final
/// response text. The stream is closed by the server after this.
class StreamAgentFinal extends $pb.GeneratedMessage {
  factory StreamAgentFinal({
    $core.String? invocationId,
    $core.String? sessionId,
    $core.String? agentName,
    $core.String? response,
    $core.String? agentId,
  }) {
    final result = StreamAgentFinal._();
    if (invocationId != null) result.invocationId = invocationId;
    if (sessionId != null) result.sessionId = sessionId;
    if (agentName != null) result.agentName = agentName;
    if (response != null) result.response = response;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  StreamAgentFinal._();

  factory StreamAgentFinal.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentFinal()..mergeFromBuffer(data, registry);
  factory StreamAgentFinal.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StreamAgentFinal()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StreamAgentFinal',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StreamAgentFinal.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invocationId')
    ..aOS(2, _omitFieldNames ? '' : 'sessionId')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aOS(4, _omitFieldNames ? '' : 'response')
    ..aOS(5, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentFinal clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StreamAgentFinal copyWith(void Function(StreamAgentFinal) updates) =>
      super.copyWith((message) => updates(message as StreamAgentFinal))
          as StreamAgentFinal;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StreamAgentFinal() / StreamAgentFinal.new instead')
  static StreamAgentFinal create() => StreamAgentFinal._();
  static $pb.GeneratedMessage $_createMessage() => StreamAgentFinal._();
  @$core.override
  StreamAgentFinal createEmptyInstance() => StreamAgentFinal._();
  @$core.pragma('dart2js:noInline')
  static StreamAgentFinal getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StreamAgentFinal>(
          StreamAgentFinal.$_createMessage);
  static StreamAgentFinal? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get invocationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invocationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocationId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get sessionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set sessionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSessionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSessionId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get response => $_getSZ(3);
  @$pb.TagNumber(4)
  set response($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasResponse() => $_has(3);
  @$pb.TagNumber(4)
  void clearResponse() => $_clearField(4);

  /// Immutable agent_id of the resolved agent, when assigned.
  @$pb.TagNumber(5)
  $core.String get agentId => $_getSZ(4);
  @$pb.TagNumber(5)
  set agentId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAgentId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAgentId() => $_clearField(5);
}

class GetAgentRequest extends $pb.GeneratedMessage {
  factory GetAgentRequest({
    @$core.Deprecated('This field is deprecated.') $core.String? name,
    $core.String? agentId,
  }) {
    final result = GetAgentRequest._();
    if (name != null) result.name = name;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  GetAgentRequest._();

  factory GetAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentRequest()..mergeFromBuffer(data, registry);
  factory GetAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentRequest copyWith(void Function(GetAgentRequest) updates) =>
      super.copyWith((message) => updates(message as GetAgentRequest))
          as GetAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetAgentRequest() / GetAgentRequest.new instead')
  static GetAgentRequest create() => GetAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAgentRequest._();
  @$core.override
  GetAgentRequest createEmptyInstance() => GetAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAgentRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAgentRequest>(
          GetAgentRequest.$_createMessage);
  static GetAgentRequest? _defaultInstance;

  /// Legacy name lookup key. Deprecated (issue #241): names never select an
  /// Agent; requests without agent_id are rejected.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Immutable agent_id of the Agent. Required.
  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);
}

class GetAgentResponse extends $pb.GeneratedMessage {
  factory GetAgentResponse({
    $0.Agent? agent,
  }) {
    final result = GetAgentResponse._();
    if (agent != null) result.agent = agent;
    return result;
  }

  GetAgentResponse._();

  factory GetAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentResponse()..mergeFromBuffer(data, registry);
  factory GetAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentResponse copyWith(void Function(GetAgentResponse) updates) =>
      super.copyWith((message) => updates(message as GetAgentResponse))
          as GetAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetAgentResponse() / GetAgentResponse.new instead')
  static GetAgentResponse create() => GetAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetAgentResponse._();
  @$core.override
  GetAgentResponse createEmptyInstance() => GetAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAgentResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAgentResponse>(
          GetAgentResponse.$_createMessage);
  static GetAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);
}

class CreateAgentRequest extends $pb.GeneratedMessage {
  factory CreateAgentRequest({
    $0.Agent? agent,
    AgentContentInput? initialContent,
    $core.String? operationId,
  }) {
    final result = CreateAgentRequest._();
    if (agent != null) result.agent = agent;
    if (initialContent != null) result.initialContent = initialContent;
    if (operationId != null) result.operationId = operationId;
    return result;
  }

  CreateAgentRequest._();

  factory CreateAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentRequest()..mergeFromBuffer(data, registry);
  factory CreateAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAgentRequest.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..aOM<AgentContentInput>(2, _omitFieldNames ? '' : 'initialContent',
        subBuilder: AgentContentInput.$_createMessage)
    ..aOS(3, _omitFieldNames ? '' : 'operationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentRequest copyWith(void Function(CreateAgentRequest) updates) =>
      super.copyWith((message) => updates(message as CreateAgentRequest))
          as CreateAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateAgentRequest() / CreateAgentRequest.new instead')
  static CreateAgentRequest create() => CreateAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateAgentRequest._();
  @$core.override
  CreateAgentRequest createEmptyInstance() => CreateAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAgentRequest>(
          CreateAgentRequest.$_createMessage);
  static CreateAgentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);

  /// Initial Agent Content for a Git-bound workspace. When the workspace has a
  /// repository binding and this is set, the agent is created through the create
  /// Saga: the DB row starts in PROVISIONING, the content is committed and
  /// published, and the agent activates only after publication succeeds. For an
  /// LLM agent, prompt must be non-empty. Ignored for unbound workspaces (which
  /// keep content in the DB and activate immediately).
  @$pb.TagNumber(2)
  AgentContentInput get initialContent => $_getN(1);
  @$pb.TagNumber(2)
  set initialContent(AgentContentInput value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasInitialContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearInitialContent() => $_clearField(2);
  @$pb.TagNumber(2)
  AgentContentInput ensureInitialContent() => $_ensure(1);

  /// Optional caller-supplied operation ID for idempotent retries of the create
  /// Saga. When empty the server generates one.
  @$pb.TagNumber(3)
  $core.String get operationId => $_getSZ(2);
  @$pb.TagNumber(3)
  set operationId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasOperationId() => $_has(2);
  @$pb.TagNumber(3)
  void clearOperationId() => $_clearField(3);
}

class CreateAgentResponse extends $pb.GeneratedMessage {
  factory CreateAgentResponse({
    $0.Agent? agent,
    $3.AgentOperation? operation,
  }) {
    final result = CreateAgentResponse._();
    if (agent != null) result.agent = agent;
    if (operation != null) result.operation = operation;
    return result;
  }

  CreateAgentResponse._();

  factory CreateAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentResponse()..mergeFromBuffer(data, registry);
  factory CreateAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAgentResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..aOM<$3.AgentOperation>(2, _omitFieldNames ? '' : 'operation',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentResponse copyWith(void Function(CreateAgentResponse) updates) =>
      super.copyWith((message) => updates(message as CreateAgentResponse))
          as CreateAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CreateAgentResponse() / CreateAgentResponse.new instead')
  static CreateAgentResponse create() => CreateAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateAgentResponse._();
  @$core.override
  CreateAgentResponse createEmptyInstance() => CreateAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAgentResponse>(
          CreateAgentResponse.$_createMessage);
  static CreateAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);

  /// The lifecycle operation record, set when the create ran as a Saga (bound
  /// workspace). Nil for the immediate single-step path (unbound workspace).
  @$pb.TagNumber(2)
  $3.AgentOperation get operation => $_getN(1);
  @$pb.TagNumber(2)
  set operation($3.AgentOperation value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOperation() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperation() => $_clearField(2);
  @$pb.TagNumber(2)
  $3.AgentOperation ensureOperation() => $_ensure(1);
}

/// AgentContentInput carries the reserved Markdown files for an agent, mapped to
/// agents/{agent-id}/{description.md, prompt.md, global-prompt.md}.
class AgentContentInput extends $pb.GeneratedMessage {
  factory AgentContentInput({
    $core.String? description,
    $core.String? prompt,
    $core.String? globalPrompt,
  }) {
    final result = AgentContentInput._();
    if (description != null) result.description = description;
    if (prompt != null) result.prompt = prompt;
    if (globalPrompt != null) result.globalPrompt = globalPrompt;
    return result;
  }

  AgentContentInput._();

  factory AgentContentInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentContentInput()..mergeFromBuffer(data, registry);
  factory AgentContentInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentContentInput()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentContentInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentContentInput.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'description')
    ..aOS(2, _omitFieldNames ? '' : 'prompt')
    ..aOS(3, _omitFieldNames ? '' : 'globalPrompt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentContentInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentContentInput copyWith(void Function(AgentContentInput) updates) =>
      super.copyWith((message) => updates(message as AgentContentInput))
          as AgentContentInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentContentInput() / AgentContentInput.new instead')
  static AgentContentInput create() => AgentContentInput._();
  static $pb.GeneratedMessage $_createMessage() => AgentContentInput._();
  @$core.override
  AgentContentInput createEmptyInstance() => AgentContentInput._();
  @$core.pragma('dart2js:noInline')
  static AgentContentInput getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentContentInput>(
          AgentContentInput.$_createMessage);
  static AgentContentInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get description => $_getSZ(0);
  @$pb.TagNumber(1)
  set description($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDescription() => $_has(0);
  @$pb.TagNumber(1)
  void clearDescription() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get prompt => $_getSZ(1);
  @$pb.TagNumber(2)
  set prompt($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPrompt() => $_has(1);
  @$pb.TagNumber(2)
  void clearPrompt() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get globalPrompt => $_getSZ(2);
  @$pb.TagNumber(3)
  set globalPrompt($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasGlobalPrompt() => $_has(2);
  @$pb.TagNumber(3)
  void clearGlobalPrompt() => $_clearField(3);
}

/// UpdateAgentConfigurationRequest is the composite-save command. The agent
/// patch and the content changes retain separate ownership (DB vs Git); the
/// operation coordinates them but does not claim a distributed transaction.
class UpdateAgentConfigurationRequest extends $pb.GeneratedMessage {
  factory UpdateAgentConfigurationRequest({
    $0.Agent? agentPatch,
    $core.Iterable<$4.ContentFileAction>? contentChanges,
    $fixnum.Int64? expectedAgentVersion,
    $core.String? baseCommitSha,
    $core.String? operationId,
  }) {
    final result = UpdateAgentConfigurationRequest._();
    if (agentPatch != null) result.agentPatch = agentPatch;
    if (contentChanges != null) result.contentChanges.addAll(contentChanges);
    if (expectedAgentVersion != null)
      result.expectedAgentVersion = expectedAgentVersion;
    if (baseCommitSha != null) result.baseCommitSha = baseCommitSha;
    if (operationId != null) result.operationId = operationId;
    return result;
  }

  UpdateAgentConfigurationRequest._();

  factory UpdateAgentConfigurationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentConfigurationRequest()..mergeFromBuffer(data, registry);
  factory UpdateAgentConfigurationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentConfigurationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAgentConfigurationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAgentConfigurationRequest.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agentPatch',
        subBuilder: $0.Agent.$_createMessage)
    ..pPM<$4.ContentFileAction>(2, _omitFieldNames ? '' : 'contentChanges',
        subBuilder: $4.ContentFileAction.$_createMessage)
    ..aInt64(3, _omitFieldNames ? '' : 'expectedAgentVersion')
    ..aOS(4, _omitFieldNames ? '' : 'baseCommitSha')
    ..aOS(5, _omitFieldNames ? '' : 'operationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentConfigurationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentConfigurationRequest copyWith(
          void Function(UpdateAgentConfigurationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateAgentConfigurationRequest))
          as UpdateAgentConfigurationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAgentConfigurationRequest() / UpdateAgentConfigurationRequest.new instead')
  static UpdateAgentConfigurationRequest create() =>
      UpdateAgentConfigurationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateAgentConfigurationRequest._();
  @$core.override
  UpdateAgentConfigurationRequest createEmptyInstance() =>
      UpdateAgentConfigurationRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAgentConfigurationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAgentConfigurationRequest>(
          UpdateAgentConfigurationRequest.$_createMessage);
  static UpdateAgentConfigurationRequest? _defaultInstance;

  /// The full operational config to write (DB-owned). Identity fields
  /// (agent_id, lifecycle_status, legacy_name) are server-enforced.
  @$pb.TagNumber(1)
  $0.Agent get agentPatch => $_getN(0);
  @$pb.TagNumber(1)
  set agentPatch($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentPatch() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentPatch() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgentPatch() => $_ensure(0);

  /// Git content changes to apply as a single commit before the DB patch.
  @$pb.TagNumber(2)
  $pb.PbList<$4.ContentFileAction> get contentChanges => $_getList(1);

  /// The agent version the caller observed. A mismatch aborts with Aborted.
  @$pb.TagNumber(3)
  $fixnum.Int64 get expectedAgentVersion => $_getI64(2);
  @$pb.TagNumber(3)
  set expectedAgentVersion($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpectedAgentVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpectedAgentVersion() => $_clearField(3);

  /// The cached base revision SHA (informational; last-write-wins applies).
  @$pb.TagNumber(4)
  $core.String get baseCommitSha => $_getSZ(3);
  @$pb.TagNumber(4)
  set baseCommitSha($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBaseCommitSha() => $_has(3);
  @$pb.TagNumber(4)
  void clearBaseCommitSha() => $_clearField(4);

  /// Optional caller-supplied operation ID for idempotent retries.
  @$pb.TagNumber(5)
  $core.String get operationId => $_getSZ(4);
  @$pb.TagNumber(5)
  set operationId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOperationId() => $_has(4);
  @$pb.TagNumber(5)
  void clearOperationId() => $_clearField(5);
}

class UpdateAgentConfigurationResponse extends $pb.GeneratedMessage {
  factory UpdateAgentConfigurationResponse({
    $0.Agent? agent,
    $3.AgentOperation? operation,
    $core.Iterable<$core.String>? validationErrors,
  }) {
    final result = UpdateAgentConfigurationResponse._();
    if (agent != null) result.agent = agent;
    if (operation != null) result.operation = operation;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    return result;
  }

  UpdateAgentConfigurationResponse._();

  factory UpdateAgentConfigurationResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentConfigurationResponse()..mergeFromBuffer(data, registry);
  factory UpdateAgentConfigurationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentConfigurationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAgentConfigurationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAgentConfigurationResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..aOM<$3.AgentOperation>(2, _omitFieldNames ? '' : 'operation',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..pPS(3, _omitFieldNames ? '' : 'validationErrors')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentConfigurationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentConfigurationResponse copyWith(
          void Function(UpdateAgentConfigurationResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateAgentConfigurationResponse))
          as UpdateAgentConfigurationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAgentConfigurationResponse() / UpdateAgentConfigurationResponse.new instead')
  static UpdateAgentConfigurationResponse create() =>
      UpdateAgentConfigurationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateAgentConfigurationResponse._();
  @$core.override
  UpdateAgentConfigurationResponse createEmptyInstance() =>
      UpdateAgentConfigurationResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAgentConfigurationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAgentConfigurationResponse>(
          UpdateAgentConfigurationResponse.$_createMessage);
  static UpdateAgentConfigurationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);

  @$pb.TagNumber(2)
  $3.AgentOperation get operation => $_getN(1);
  @$pb.TagNumber(2)
  set operation($3.AgentOperation value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOperation() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperation() => $_clearField(2);
  @$pb.TagNumber(2)
  $3.AgentOperation ensureOperation() => $_ensure(1);

  /// Per-agent content validation errors that blocked the commit (empty on success).
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get validationErrors => $_getList(2);
}

class RestoreAgentRequest extends $pb.GeneratedMessage {
  factory RestoreAgentRequest({
    $core.String? agentId,
    $core.String? operationId,
  }) {
    final result = RestoreAgentRequest._();
    if (agentId != null) result.agentId = agentId;
    if (operationId != null) result.operationId = operationId;
    return result;
  }

  RestoreAgentRequest._();

  factory RestoreAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreAgentRequest()..mergeFromBuffer(data, registry);
  factory RestoreAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RestoreAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentId')
    ..aOS(2, _omitFieldNames ? '' : 'operationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreAgentRequest copyWith(void Function(RestoreAgentRequest) updates) =>
      super.copyWith((message) => updates(message as RestoreAgentRequest))
          as RestoreAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use RestoreAgentRequest() / RestoreAgentRequest.new instead')
  static RestoreAgentRequest create() => RestoreAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => RestoreAgentRequest._();
  @$core.override
  RestoreAgentRequest createEmptyInstance() => RestoreAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static RestoreAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreAgentRequest>(
          RestoreAgentRequest.$_createMessage);
  static RestoreAgentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get agentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set agentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentId() => $_clearField(1);

  /// Optional caller-supplied operation ID for idempotent retries.
  @$pb.TagNumber(2)
  $core.String get operationId => $_getSZ(1);
  @$pb.TagNumber(2)
  set operationId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasOperationId() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperationId() => $_clearField(2);
}

class RestoreAgentResponse extends $pb.GeneratedMessage {
  factory RestoreAgentResponse({
    $0.Agent? agent,
    $3.AgentOperation? operation,
    $core.Iterable<$core.String>? validationErrors,
  }) {
    final result = RestoreAgentResponse._();
    if (agent != null) result.agent = agent;
    if (operation != null) result.operation = operation;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    return result;
  }

  RestoreAgentResponse._();

  factory RestoreAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreAgentResponse()..mergeFromBuffer(data, registry);
  factory RestoreAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestoreAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestoreAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RestoreAgentResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..aOM<$3.AgentOperation>(2, _omitFieldNames ? '' : 'operation',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..pPS(3, _omitFieldNames ? '' : 'validationErrors')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestoreAgentResponse copyWith(void Function(RestoreAgentResponse) updates) =>
      super.copyWith((message) => updates(message as RestoreAgentResponse))
          as RestoreAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestoreAgentResponse() / RestoreAgentResponse.new instead')
  static RestoreAgentResponse create() => RestoreAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => RestoreAgentResponse._();
  @$core.override
  RestoreAgentResponse createEmptyInstance() => RestoreAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static RestoreAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestoreAgentResponse>(
          RestoreAgentResponse.$_createMessage);
  static RestoreAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);

  @$pb.TagNumber(2)
  $3.AgentOperation get operation => $_getN(1);
  @$pb.TagNumber(2)
  set operation($3.AgentOperation value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOperation() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperation() => $_clearField(2);
  @$pb.TagNumber(2)
  $3.AgentOperation ensureOperation() => $_ensure(1);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get validationErrors => $_getList(2);
}

class GetAgentOperationRequest extends $pb.GeneratedMessage {
  factory GetAgentOperationRequest({
    $core.String? operationId,
  }) {
    final result = GetAgentOperationRequest._();
    if (operationId != null) result.operationId = operationId;
    return result;
  }

  GetAgentOperationRequest._();

  factory GetAgentOperationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentOperationRequest()..mergeFromBuffer(data, registry);
  factory GetAgentOperationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentOperationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentOperationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentOperationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'operationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentOperationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentOperationRequest copyWith(
          void Function(GetAgentOperationRequest) updates) =>
      super.copyWith((message) => updates(message as GetAgentOperationRequest))
          as GetAgentOperationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentOperationRequest() / GetAgentOperationRequest.new instead')
  static GetAgentOperationRequest create() => GetAgentOperationRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAgentOperationRequest._();
  @$core.override
  GetAgentOperationRequest createEmptyInstance() =>
      GetAgentOperationRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAgentOperationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentOperationRequest>(
          GetAgentOperationRequest.$_createMessage);
  static GetAgentOperationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get operationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set operationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOperationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperationId() => $_clearField(1);
}

class GetAgentOperationResponse extends $pb.GeneratedMessage {
  factory GetAgentOperationResponse({
    $3.AgentOperation? operation,
  }) {
    final result = GetAgentOperationResponse._();
    if (operation != null) result.operation = operation;
    return result;
  }

  GetAgentOperationResponse._();

  factory GetAgentOperationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentOperationResponse()..mergeFromBuffer(data, registry);
  factory GetAgentOperationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentOperationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentOperationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentOperationResponse.$_createMessage)
    ..aOM<$3.AgentOperation>(1, _omitFieldNames ? '' : 'operation',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentOperationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentOperationResponse copyWith(
          void Function(GetAgentOperationResponse) updates) =>
      super.copyWith((message) => updates(message as GetAgentOperationResponse))
          as GetAgentOperationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentOperationResponse() / GetAgentOperationResponse.new instead')
  static GetAgentOperationResponse create() => GetAgentOperationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAgentOperationResponse._();
  @$core.override
  GetAgentOperationResponse createEmptyInstance() =>
      GetAgentOperationResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAgentOperationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentOperationResponse>(
          GetAgentOperationResponse.$_createMessage);
  static GetAgentOperationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $3.AgentOperation get operation => $_getN(0);
  @$pb.TagNumber(1)
  set operation($3.AgentOperation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOperation() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperation() => $_clearField(1);
  @$pb.TagNumber(1)
  $3.AgentOperation ensureOperation() => $_ensure(0);
}

class ListAgentOperationsRequest extends $pb.GeneratedMessage {
  factory ListAgentOperationsRequest({
    $3.AgentOperationStatus? status,
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = ListAgentOperationsRequest._();
    if (status != null) result.status = status;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListAgentOperationsRequest._();

  factory ListAgentOperationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentOperationsRequest()..mergeFromBuffer(data, registry);
  factory ListAgentOperationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentOperationsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentOperationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentOperationsRequest.$_createMessage)
    ..aE<$3.AgentOperationStatus>(1, _omitFieldNames ? '' : 'status',
        enumValues: $3.AgentOperationStatus.values)
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentOperationsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentOperationsRequest copyWith(
          void Function(ListAgentOperationsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentOperationsRequest))
          as ListAgentOperationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentOperationsRequest() / ListAgentOperationsRequest.new instead')
  static ListAgentOperationsRequest create() => ListAgentOperationsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentOperationsRequest._();
  @$core.override
  ListAgentOperationsRequest createEmptyInstance() =>
      ListAgentOperationsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAgentOperationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentOperationsRequest>(
          ListAgentOperationsRequest.$_createMessage);
  static ListAgentOperationsRequest? _defaultInstance;

  /// Optional status filter; UNSPECIFIED returns all.
  @$pb.TagNumber(1)
  $3.AgentOperationStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status($3.AgentOperationStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);

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

class ListAgentOperationsResponse extends $pb.GeneratedMessage {
  factory ListAgentOperationsResponse({
    $core.Iterable<$3.AgentOperation>? operations,
    $core.String? nextPageToken,
  }) {
    final result = ListAgentOperationsResponse._();
    if (operations != null) result.operations.addAll(operations);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  ListAgentOperationsResponse._();

  factory ListAgentOperationsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentOperationsResponse()..mergeFromBuffer(data, registry);
  factory ListAgentOperationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentOperationsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentOperationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentOperationsResponse.$_createMessage)
    ..pPM<$3.AgentOperation>(1, _omitFieldNames ? '' : 'operations',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentOperationsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentOperationsResponse copyWith(
          void Function(ListAgentOperationsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentOperationsResponse))
          as ListAgentOperationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentOperationsResponse() / ListAgentOperationsResponse.new instead')
  static ListAgentOperationsResponse create() =>
      ListAgentOperationsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentOperationsResponse._();
  @$core.override
  ListAgentOperationsResponse createEmptyInstance() =>
      ListAgentOperationsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAgentOperationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentOperationsResponse>(
          ListAgentOperationsResponse.$_createMessage);
  static ListAgentOperationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$3.AgentOperation> get operations => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);
}

class RetryAgentOperationRequest extends $pb.GeneratedMessage {
  factory RetryAgentOperationRequest({
    $core.String? operationId,
  }) {
    final result = RetryAgentOperationRequest._();
    if (operationId != null) result.operationId = operationId;
    return result;
  }

  RetryAgentOperationRequest._();

  factory RetryAgentOperationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RetryAgentOperationRequest()..mergeFromBuffer(data, registry);
  factory RetryAgentOperationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RetryAgentOperationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RetryAgentOperationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RetryAgentOperationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'operationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RetryAgentOperationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RetryAgentOperationRequest copyWith(
          void Function(RetryAgentOperationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RetryAgentOperationRequest))
          as RetryAgentOperationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RetryAgentOperationRequest() / RetryAgentOperationRequest.new instead')
  static RetryAgentOperationRequest create() => RetryAgentOperationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RetryAgentOperationRequest._();
  @$core.override
  RetryAgentOperationRequest createEmptyInstance() =>
      RetryAgentOperationRequest._();
  @$core.pragma('dart2js:noInline')
  static RetryAgentOperationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RetryAgentOperationRequest>(
          RetryAgentOperationRequest.$_createMessage);
  static RetryAgentOperationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get operationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set operationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOperationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperationId() => $_clearField(1);
}

class RetryAgentOperationResponse extends $pb.GeneratedMessage {
  factory RetryAgentOperationResponse({
    $0.Agent? agent,
    $3.AgentOperation? operation,
    $core.Iterable<$core.String>? validationErrors,
  }) {
    final result = RetryAgentOperationResponse._();
    if (agent != null) result.agent = agent;
    if (operation != null) result.operation = operation;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    return result;
  }

  RetryAgentOperationResponse._();

  factory RetryAgentOperationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RetryAgentOperationResponse()..mergeFromBuffer(data, registry);
  factory RetryAgentOperationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RetryAgentOperationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RetryAgentOperationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RetryAgentOperationResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..aOM<$3.AgentOperation>(2, _omitFieldNames ? '' : 'operation',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..pPS(3, _omitFieldNames ? '' : 'validationErrors')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RetryAgentOperationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RetryAgentOperationResponse copyWith(
          void Function(RetryAgentOperationResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RetryAgentOperationResponse))
          as RetryAgentOperationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RetryAgentOperationResponse() / RetryAgentOperationResponse.new instead')
  static RetryAgentOperationResponse create() =>
      RetryAgentOperationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RetryAgentOperationResponse._();
  @$core.override
  RetryAgentOperationResponse createEmptyInstance() =>
      RetryAgentOperationResponse._();
  @$core.pragma('dart2js:noInline')
  static RetryAgentOperationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RetryAgentOperationResponse>(
          RetryAgentOperationResponse.$_createMessage);
  static RetryAgentOperationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);

  @$pb.TagNumber(2)
  $3.AgentOperation get operation => $_getN(1);
  @$pb.TagNumber(2)
  set operation($3.AgentOperation value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOperation() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperation() => $_clearField(2);
  @$pb.TagNumber(2)
  $3.AgentOperation ensureOperation() => $_ensure(1);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get validationErrors => $_getList(2);
}

class UpdateAgentRequest extends $pb.GeneratedMessage {
  factory UpdateAgentRequest({
    $0.Agent? agent,
  }) {
    final result = UpdateAgentRequest._();
    if (agent != null) result.agent = agent;
    return result;
  }

  UpdateAgentRequest._();

  factory UpdateAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentRequest()..mergeFromBuffer(data, registry);
  factory UpdateAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAgentRequest.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentRequest copyWith(void Function(UpdateAgentRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateAgentRequest))
          as UpdateAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UpdateAgentRequest() / UpdateAgentRequest.new instead')
  static UpdateAgentRequest create() => UpdateAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAgentRequest._();
  @$core.override
  UpdateAgentRequest createEmptyInstance() => UpdateAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAgentRequest>(
          UpdateAgentRequest.$_createMessage);
  static UpdateAgentRequest? _defaultInstance;

  /// The agent to update, located by its immutable agent.agent_id (required).
  /// agent.name is server-controlled and never selects the record.
  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);
}

class UpdateAgentResponse extends $pb.GeneratedMessage {
  factory UpdateAgentResponse({
    $0.Agent? agent,
  }) {
    final result = UpdateAgentResponse._();
    if (agent != null) result.agent = agent;
    return result;
  }

  UpdateAgentResponse._();

  factory UpdateAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentResponse()..mergeFromBuffer(data, registry);
  factory UpdateAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAgentResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentResponse copyWith(void Function(UpdateAgentResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateAgentResponse))
          as UpdateAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use UpdateAgentResponse() / UpdateAgentResponse.new instead')
  static UpdateAgentResponse create() => UpdateAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateAgentResponse._();
  @$core.override
  UpdateAgentResponse createEmptyInstance() => UpdateAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAgentResponse>(
          UpdateAgentResponse.$_createMessage);
  static UpdateAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);
}

class DeleteAgentRequest extends $pb.GeneratedMessage {
  factory DeleteAgentRequest({
    @$core.Deprecated('This field is deprecated.') $core.String? name,
    $core.String? agentId,
    $core.String? operationId,
  }) {
    final result = DeleteAgentRequest._();
    if (name != null) result.name = name;
    if (agentId != null) result.agentId = agentId;
    if (operationId != null) result.operationId = operationId;
    return result;
  }

  DeleteAgentRequest._();

  factory DeleteAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentRequest()..mergeFromBuffer(data, registry);
  factory DeleteAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..aOS(3, _omitFieldNames ? '' : 'operationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentRequest copyWith(void Function(DeleteAgentRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteAgentRequest))
          as DeleteAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeleteAgentRequest() / DeleteAgentRequest.new instead')
  static DeleteAgentRequest create() => DeleteAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAgentRequest._();
  @$core.override
  DeleteAgentRequest createEmptyInstance() => DeleteAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAgentRequest>(
          DeleteAgentRequest.$_createMessage);
  static DeleteAgentRequest? _defaultInstance;

  /// Legacy name lookup key. Deprecated (issue #241): names never select an
  /// Agent; requests without agent_id are rejected.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Immutable agent_id of the Agent. Required.
  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);

  /// Optional caller-supplied operation ID for idempotent retries.
  @$pb.TagNumber(3)
  $core.String get operationId => $_getSZ(2);
  @$pb.TagNumber(3)
  set operationId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasOperationId() => $_has(2);
  @$pb.TagNumber(3)
  void clearOperationId() => $_clearField(3);
}

class DeleteAgentResponse extends $pb.GeneratedMessage {
  factory DeleteAgentResponse({
    $3.AgentOperation? operation,
  }) {
    final result = DeleteAgentResponse._();
    if (operation != null) result.operation = operation;
    return result;
  }

  DeleteAgentResponse._();

  factory DeleteAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentResponse()..mergeFromBuffer(data, registry);
  factory DeleteAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAgentResponse.$_createMessage)
    ..aOM<$3.AgentOperation>(1, _omitFieldNames ? '' : 'operation',
        subBuilder: $3.AgentOperation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentResponse copyWith(void Function(DeleteAgentResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteAgentResponse))
          as DeleteAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DeleteAgentResponse() / DeleteAgentResponse.new instead')
  static DeleteAgentResponse create() => DeleteAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAgentResponse._();
  @$core.override
  DeleteAgentResponse createEmptyInstance() => DeleteAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAgentResponse>(
          DeleteAgentResponse.$_createMessage);
  static DeleteAgentResponse? _defaultInstance;

  /// The durable tombstone operation. Nil only when lifecycle Saga support is
  /// unavailable and the legacy single-step path is used.
  @$pb.TagNumber(1)
  $3.AgentOperation get operation => $_getN(0);
  @$pb.TagNumber(1)
  set operation($3.AgentOperation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasOperation() => $_has(0);
  @$pb.TagNumber(1)
  void clearOperation() => $_clearField(1);
  @$pb.TagNumber(1)
  $3.AgentOperation ensureOperation() => $_ensure(0);
}

@$core.Deprecated('This message is deprecated')
class AssignAgentIDRequest extends $pb.GeneratedMessage {
  factory AssignAgentIDRequest({
    $core.String? name,
    $core.String? agentId,
  }) {
    final result = AssignAgentIDRequest._();
    if (name != null) result.name = name;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  AssignAgentIDRequest._();

  factory AssignAgentIDRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignAgentIDRequest()..mergeFromBuffer(data, registry);
  factory AssignAgentIDRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignAgentIDRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignAgentIDRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AssignAgentIDRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignAgentIDRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignAgentIDRequest copyWith(void Function(AssignAgentIDRequest) updates) =>
      super.copyWith((message) => updates(message as AssignAgentIDRequest))
          as AssignAgentIDRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AssignAgentIDRequest() / AssignAgentIDRequest.new instead')
  static AssignAgentIDRequest create() => AssignAgentIDRequest._();
  static $pb.GeneratedMessage $_createMessage() => AssignAgentIDRequest._();
  @$core.override
  AssignAgentIDRequest createEmptyInstance() => AssignAgentIDRequest._();
  @$core.pragma('dart2js:noInline')
  static AssignAgentIDRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignAgentIDRequest>(
          AssignAgentIDRequest.$_createMessage);
  static AssignAgentIDRequest? _defaultInstance;

  /// Name of the existing agent (the former lookup key).
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// The slug to assign as the Agent ID.
  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);
}

class AssignAgentIDResponse extends $pb.GeneratedMessage {
  factory AssignAgentIDResponse({
    $0.Agent? agent,
  }) {
    final result = AssignAgentIDResponse._();
    if (agent != null) result.agent = agent;
    return result;
  }

  AssignAgentIDResponse._();

  factory AssignAgentIDResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignAgentIDResponse()..mergeFromBuffer(data, registry);
  factory AssignAgentIDResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AssignAgentIDResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AssignAgentIDResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AssignAgentIDResponse.$_createMessage)
    ..aOM<$0.Agent>(1, _omitFieldNames ? '' : 'agent',
        subBuilder: $0.Agent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignAgentIDResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AssignAgentIDResponse copyWith(
          void Function(AssignAgentIDResponse) updates) =>
      super.copyWith((message) => updates(message as AssignAgentIDResponse))
          as AssignAgentIDResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AssignAgentIDResponse() / AssignAgentIDResponse.new instead')
  static AssignAgentIDResponse create() => AssignAgentIDResponse._();
  static $pb.GeneratedMessage $_createMessage() => AssignAgentIDResponse._();
  @$core.override
  AssignAgentIDResponse createEmptyInstance() => AssignAgentIDResponse._();
  @$core.pragma('dart2js:noInline')
  static AssignAgentIDResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AssignAgentIDResponse>(
          AssignAgentIDResponse.$_createMessage);
  static AssignAgentIDResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Agent get agent => $_getN(0);
  @$pb.TagNumber(1)
  set agent($0.Agent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Agent ensureAgent() => $_ensure(0);
}

class GetMigrationReadinessRequest extends $pb.GeneratedMessage {
  factory GetMigrationReadinessRequest() => GetMigrationReadinessRequest._();

  GetMigrationReadinessRequest._();

  factory GetMigrationReadinessRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMigrationReadinessRequest()..mergeFromBuffer(data, registry);
  factory GetMigrationReadinessRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMigrationReadinessRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMigrationReadinessRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMigrationReadinessRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMigrationReadinessRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMigrationReadinessRequest copyWith(
          void Function(GetMigrationReadinessRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetMigrationReadinessRequest))
          as GetMigrationReadinessRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMigrationReadinessRequest() / GetMigrationReadinessRequest.new instead')
  static GetMigrationReadinessRequest create() =>
      GetMigrationReadinessRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetMigrationReadinessRequest._();
  @$core.override
  GetMigrationReadinessRequest createEmptyInstance() =>
      GetMigrationReadinessRequest._();
  @$core.pragma('dart2js:noInline')
  static GetMigrationReadinessRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMigrationReadinessRequest>(
          GetMigrationReadinessRequest.$_createMessage);
  static GetMigrationReadinessRequest? _defaultInstance;
}

class GetMigrationReadinessResponse extends $pb.GeneratedMessage {
  factory GetMigrationReadinessResponse({
    $core.Iterable<AgentMigrationStatus>? statuses,
  }) {
    final result = GetMigrationReadinessResponse._();
    if (statuses != null) result.statuses.addAll(statuses);
    return result;
  }

  GetMigrationReadinessResponse._();

  factory GetMigrationReadinessResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMigrationReadinessResponse()..mergeFromBuffer(data, registry);
  factory GetMigrationReadinessResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMigrationReadinessResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMigrationReadinessResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMigrationReadinessResponse.$_createMessage)
    ..pPM<AgentMigrationStatus>(1, _omitFieldNames ? '' : 'statuses',
        subBuilder: AgentMigrationStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMigrationReadinessResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMigrationReadinessResponse copyWith(
          void Function(GetMigrationReadinessResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetMigrationReadinessResponse))
          as GetMigrationReadinessResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMigrationReadinessResponse() / GetMigrationReadinessResponse.new instead')
  static GetMigrationReadinessResponse create() =>
      GetMigrationReadinessResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetMigrationReadinessResponse._();
  @$core.override
  GetMigrationReadinessResponse createEmptyInstance() =>
      GetMigrationReadinessResponse._();
  @$core.pragma('dart2js:noInline')
  static GetMigrationReadinessResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMigrationReadinessResponse>(
          GetMigrationReadinessResponse.$_createMessage);
  static GetMigrationReadinessResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AgentMigrationStatus> get statuses => $_getList(0);
}

class AgentMigrationStatus extends $pb.GeneratedMessage {
  factory AgentMigrationStatus({
    $core.String? name,
    $core.String? agentId,
    MigrationReadiness? readiness,
    $core.String? detail,
  }) {
    final result = AgentMigrationStatus._();
    if (name != null) result.name = name;
    if (agentId != null) result.agentId = agentId;
    if (readiness != null) result.readiness = readiness;
    if (detail != null) result.detail = detail;
    return result;
  }

  AgentMigrationStatus._();

  factory AgentMigrationStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentMigrationStatus()..mergeFromBuffer(data, registry);
  factory AgentMigrationStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentMigrationStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentMigrationStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentMigrationStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..aE<MigrationReadiness>(3, _omitFieldNames ? '' : 'readiness',
        enumValues: MigrationReadiness.values)
    ..aOS(4, _omitFieldNames ? '' : 'detail')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentMigrationStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentMigrationStatus copyWith(void Function(AgentMigrationStatus) updates) =>
      super.copyWith((message) => updates(message as AgentMigrationStatus))
          as AgentMigrationStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AgentMigrationStatus() / AgentMigrationStatus.new instead')
  static AgentMigrationStatus create() => AgentMigrationStatus._();
  static $pb.GeneratedMessage $_createMessage() => AgentMigrationStatus._();
  @$core.override
  AgentMigrationStatus createEmptyInstance() => AgentMigrationStatus._();
  @$core.pragma('dart2js:noInline')
  static AgentMigrationStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentMigrationStatus>(
          AgentMigrationStatus.$_createMessage);
  static AgentMigrationStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);

  @$pb.TagNumber(3)
  MigrationReadiness get readiness => $_getN(2);
  @$pb.TagNumber(3)
  set readiness(MigrationReadiness value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasReadiness() => $_has(2);
  @$pb.TagNumber(3)
  void clearReadiness() => $_clearField(3);

  /// Human-readable detail when readiness is not READY.
  @$pb.TagNumber(4)
  $core.String get detail => $_getSZ(3);
  @$pb.TagNumber(4)
  set detail($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDetail() => $_has(3);
  @$pb.TagNumber(4)
  void clearDetail() => $_clearField(4);
}

class MigrateAgentsV2Request extends $pb.GeneratedMessage {
  factory MigrateAgentsV2Request({
    MigrateMode? mode,
  }) {
    final result = MigrateAgentsV2Request._();
    if (mode != null) result.mode = mode;
    return result;
  }

  MigrateAgentsV2Request._();

  factory MigrateAgentsV2Request.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MigrateAgentsV2Request()..mergeFromBuffer(data, registry);
  factory MigrateAgentsV2Request.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MigrateAgentsV2Request()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MigrateAgentsV2Request',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MigrateAgentsV2Request.$_createMessage)
    ..aE<MigrateMode>(1, _omitFieldNames ? '' : 'mode',
        enumValues: MigrateMode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MigrateAgentsV2Request clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MigrateAgentsV2Request copyWith(
          void Function(MigrateAgentsV2Request) updates) =>
      super.copyWith((message) => updates(message as MigrateAgentsV2Request))
          as MigrateAgentsV2Request;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MigrateAgentsV2Request() / MigrateAgentsV2Request.new instead')
  static MigrateAgentsV2Request create() => MigrateAgentsV2Request._();
  static $pb.GeneratedMessage $_createMessage() => MigrateAgentsV2Request._();
  @$core.override
  MigrateAgentsV2Request createEmptyInstance() => MigrateAgentsV2Request._();
  @$core.pragma('dart2js:noInline')
  static MigrateAgentsV2Request getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MigrateAgentsV2Request>(
          MigrateAgentsV2Request.$_createMessage);
  static MigrateAgentsV2Request? _defaultInstance;

  @$pb.TagNumber(1)
  MigrateMode get mode => $_getN(0);
  @$pb.TagNumber(1)
  set mode(MigrateMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearMode() => $_clearField(1);
}

/// MigrateAgentResult describes what happened (or would happen) for one agent.
class MigrateAgentResult extends $pb.GeneratedMessage {
  factory MigrateAgentResult({
    $core.String? name,
    $core.String? agentId,
    $core.String? action,
    $core.String? detail,
  }) {
    final result = MigrateAgentResult._();
    if (name != null) result.name = name;
    if (agentId != null) result.agentId = agentId;
    if (action != null) result.action = action;
    if (detail != null) result.detail = detail;
    return result;
  }

  MigrateAgentResult._();

  factory MigrateAgentResult.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MigrateAgentResult()..mergeFromBuffer(data, registry);
  factory MigrateAgentResult.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MigrateAgentResult()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MigrateAgentResult',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MigrateAgentResult.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..aOS(3, _omitFieldNames ? '' : 'action')
    ..aOS(4, _omitFieldNames ? '' : 'detail')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MigrateAgentResult clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MigrateAgentResult copyWith(void Function(MigrateAgentResult) updates) =>
      super.copyWith((message) => updates(message as MigrateAgentResult))
          as MigrateAgentResult;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MigrateAgentResult() / MigrateAgentResult.new instead')
  static MigrateAgentResult create() => MigrateAgentResult._();
  static $pb.GeneratedMessage $_createMessage() => MigrateAgentResult._();
  @$core.override
  MigrateAgentResult createEmptyInstance() => MigrateAgentResult._();
  @$core.pragma('dart2js:noInline')
  static MigrateAgentResult getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MigrateAgentResult>(
          MigrateAgentResult.$_createMessage);
  static MigrateAgentResult? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);

  /// Action taken: "expanded", "skipped", "already_independent",
  /// "missing_id", "error", "ok", or "migration_required".
  @$pb.TagNumber(3)
  $core.String get action => $_getSZ(2);
  @$pb.TagNumber(3)
  set action($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAction() => $_has(2);
  @$pb.TagNumber(3)
  void clearAction() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get detail => $_getSZ(3);
  @$pb.TagNumber(4)
  set detail($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDetail() => $_has(3);
  @$pb.TagNumber(4)
  void clearDetail() => $_clearField(4);
}

class MigrateAgentsV2Response extends $pb.GeneratedMessage {
  factory MigrateAgentsV2Response({
    MigrateMode? mode,
    $core.Iterable<MigrateAgentResult>? results,
    $core.int? total,
    $core.int? migrated,
    $core.int? skipped,
    $core.int? errors,
  }) {
    final result = MigrateAgentsV2Response._();
    if (mode != null) result.mode = mode;
    if (results != null) result.results.addAll(results);
    if (total != null) result.total = total;
    if (migrated != null) result.migrated = migrated;
    if (skipped != null) result.skipped = skipped;
    if (errors != null) result.errors = errors;
    return result;
  }

  MigrateAgentsV2Response._();

  factory MigrateAgentsV2Response.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MigrateAgentsV2Response()..mergeFromBuffer(data, registry);
  factory MigrateAgentsV2Response.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MigrateAgentsV2Response()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MigrateAgentsV2Response',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MigrateAgentsV2Response.$_createMessage)
    ..aE<MigrateMode>(1, _omitFieldNames ? '' : 'mode',
        enumValues: MigrateMode.values)
    ..pPM<MigrateAgentResult>(2, _omitFieldNames ? '' : 'results',
        subBuilder: MigrateAgentResult.$_createMessage)
    ..aI(3, _omitFieldNames ? '' : 'total')
    ..aI(4, _omitFieldNames ? '' : 'migrated')
    ..aI(5, _omitFieldNames ? '' : 'skipped')
    ..aI(6, _omitFieldNames ? '' : 'errors')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MigrateAgentsV2Response clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MigrateAgentsV2Response copyWith(
          void Function(MigrateAgentsV2Response) updates) =>
      super.copyWith((message) => updates(message as MigrateAgentsV2Response))
          as MigrateAgentsV2Response;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MigrateAgentsV2Response() / MigrateAgentsV2Response.new instead')
  static MigrateAgentsV2Response create() => MigrateAgentsV2Response._();
  static $pb.GeneratedMessage $_createMessage() => MigrateAgentsV2Response._();
  @$core.override
  MigrateAgentsV2Response createEmptyInstance() => MigrateAgentsV2Response._();
  @$core.pragma('dart2js:noInline')
  static MigrateAgentsV2Response getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MigrateAgentsV2Response>(
          MigrateAgentsV2Response.$_createMessage);
  static MigrateAgentsV2Response? _defaultInstance;

  @$pb.TagNumber(1)
  MigrateMode get mode => $_getN(0);
  @$pb.TagNumber(1)
  set mode(MigrateMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearMode() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<MigrateAgentResult> get results => $_getList(1);

  /// Summary counts.
  @$pb.TagNumber(3)
  $core.int get total => $_getIZ(2);
  @$pb.TagNumber(3)
  set total($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTotal() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotal() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get migrated => $_getIZ(3);
  @$pb.TagNumber(4)
  set migrated($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMigrated() => $_has(3);
  @$pb.TagNumber(4)
  void clearMigrated() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get skipped => $_getIZ(4);
  @$pb.TagNumber(5)
  set skipped($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSkipped() => $_has(4);
  @$pb.TagNumber(5)
  void clearSkipped() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get errors => $_getIZ(5);
  @$pb.TagNumber(6)
  set errors($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasErrors() => $_has(5);
  @$pb.TagNumber(6)
  void clearErrors() => $_clearField(6);
}

class VerifyAgentIDCutoverRequest extends $pb.GeneratedMessage {
  factory VerifyAgentIDCutoverRequest() => VerifyAgentIDCutoverRequest._();

  VerifyAgentIDCutoverRequest._();

  factory VerifyAgentIDCutoverRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VerifyAgentIDCutoverRequest()..mergeFromBuffer(data, registry);
  factory VerifyAgentIDCutoverRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VerifyAgentIDCutoverRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VerifyAgentIDCutoverRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: VerifyAgentIDCutoverRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyAgentIDCutoverRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyAgentIDCutoverRequest copyWith(
          void Function(VerifyAgentIDCutoverRequest) updates) =>
      super.copyWith(
              (message) => updates(message as VerifyAgentIDCutoverRequest))
          as VerifyAgentIDCutoverRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use VerifyAgentIDCutoverRequest() / VerifyAgentIDCutoverRequest.new instead')
  static VerifyAgentIDCutoverRequest create() =>
      VerifyAgentIDCutoverRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      VerifyAgentIDCutoverRequest._();
  @$core.override
  VerifyAgentIDCutoverRequest createEmptyInstance() =>
      VerifyAgentIDCutoverRequest._();
  @$core.pragma('dart2js:noInline')
  static VerifyAgentIDCutoverRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VerifyAgentIDCutoverRequest>(
          VerifyAgentIDCutoverRequest.$_createMessage);
  static VerifyAgentIDCutoverRequest? _defaultInstance;
}

/// AgentIDCutoverFinding is one actionable diagnostic from the final cutover
/// verifier: exactly which entity in which workspace still violates the
/// agent_id contract, and why.
class AgentIDCutoverFinding extends $pb.GeneratedMessage {
  factory AgentIDCutoverFinding({
    $core.String? workspaceId,
    $core.String? check_2,
    $core.String? entityKind,
    $core.String? entity,
    $core.String? detail,
  }) {
    final result = AgentIDCutoverFinding._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (check_2 != null) result.check_2 = check_2;
    if (entityKind != null) result.entityKind = entityKind;
    if (entity != null) result.entity = entity;
    if (detail != null) result.detail = detail;
    return result;
  }

  AgentIDCutoverFinding._();

  factory AgentIDCutoverFinding.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentIDCutoverFinding()..mergeFromBuffer(data, registry);
  factory AgentIDCutoverFinding.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentIDCutoverFinding()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentIDCutoverFinding',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentIDCutoverFinding.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'check')
    ..aOS(3, _omitFieldNames ? '' : 'entityKind')
    ..aOS(4, _omitFieldNames ? '' : 'entity')
    ..aOS(5, _omitFieldNames ? '' : 'detail')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentIDCutoverFinding clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentIDCutoverFinding copyWith(
          void Function(AgentIDCutoverFinding) updates) =>
      super.copyWith((message) => updates(message as AgentIDCutoverFinding))
          as AgentIDCutoverFinding;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AgentIDCutoverFinding() / AgentIDCutoverFinding.new instead')
  static AgentIDCutoverFinding create() => AgentIDCutoverFinding._();
  static $pb.GeneratedMessage $_createMessage() => AgentIDCutoverFinding._();
  @$core.override
  AgentIDCutoverFinding createEmptyInstance() => AgentIDCutoverFinding._();
  @$core.pragma('dart2js:noInline')
  static AgentIDCutoverFinding getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentIDCutoverFinding>(
          AgentIDCutoverFinding.$_createMessage);
  static AgentIDCutoverFinding? _defaultInstance;

  /// Workspace the offending entity belongs to.
  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  /// Machine-readable check name: "agent_id_missing", "agent_id_invalid",
  /// "agent_id_duplicate", "embedded_sub_agents", "workflow_legacy_agent_ref",
  /// "child_agent_id_unresolved", "workflow_agent_id_unresolved",
  /// "lifecycle_migration_required", "channel_agent_id_missing",
  /// "cron_agent_id_missing", "automation_agent_id_missing",
  /// "forum_agent_id_missing", or "runtime_name_conflict".
  @$pb.TagNumber(2)
  $core.String get check_2 => $_getSZ(1);
  @$pb.TagNumber(2)
  set check_2($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCheck_2() => $_has(1);
  @$pb.TagNumber(2)
  void clearCheck_2() => $_clearField(2);

  /// Kind of the offending entity: "agent", "channel", "cron_job",
  /// "automation", or "forum_thread".
  @$pb.TagNumber(3)
  $core.String get entityKind => $_getSZ(2);
  @$pb.TagNumber(3)
  set entityKind($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEntityKind() => $_has(2);
  @$pb.TagNumber(3)
  void clearEntityKind() => $_clearField(3);

  /// Identifier of the offending entity (agent_id or name for agents, record
  /// ID for consumer records).
  @$pb.TagNumber(4)
  $core.String get entity => $_getSZ(3);
  @$pb.TagNumber(4)
  set entity($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEntity() => $_has(3);
  @$pb.TagNumber(4)
  void clearEntity() => $_clearField(4);

  /// Human-readable diagnostic with enough context to fix the record.
  @$pb.TagNumber(5)
  $core.String get detail => $_getSZ(4);
  @$pb.TagNumber(5)
  set detail($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDetail() => $_has(4);
  @$pb.TagNumber(5)
  void clearDetail() => $_clearField(5);
}

class VerifyAgentIDCutoverResponse extends $pb.GeneratedMessage {
  factory VerifyAgentIDCutoverResponse({
    $core.bool? passed,
    $core.Iterable<AgentIDCutoverFinding>? findings,
  }) {
    final result = VerifyAgentIDCutoverResponse._();
    if (passed != null) result.passed = passed;
    if (findings != null) result.findings.addAll(findings);
    return result;
  }

  VerifyAgentIDCutoverResponse._();

  factory VerifyAgentIDCutoverResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VerifyAgentIDCutoverResponse()..mergeFromBuffer(data, registry);
  factory VerifyAgentIDCutoverResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VerifyAgentIDCutoverResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VerifyAgentIDCutoverResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: VerifyAgentIDCutoverResponse.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'passed')
    ..pPM<AgentIDCutoverFinding>(2, _omitFieldNames ? '' : 'findings',
        subBuilder: AgentIDCutoverFinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyAgentIDCutoverResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyAgentIDCutoverResponse copyWith(
          void Function(VerifyAgentIDCutoverResponse) updates) =>
      super.copyWith(
              (message) => updates(message as VerifyAgentIDCutoverResponse))
          as VerifyAgentIDCutoverResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use VerifyAgentIDCutoverResponse() / VerifyAgentIDCutoverResponse.new instead')
  static VerifyAgentIDCutoverResponse create() =>
      VerifyAgentIDCutoverResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      VerifyAgentIDCutoverResponse._();
  @$core.override
  VerifyAgentIDCutoverResponse createEmptyInstance() =>
      VerifyAgentIDCutoverResponse._();
  @$core.pragma('dart2js:noInline')
  static VerifyAgentIDCutoverResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VerifyAgentIDCutoverResponse>(
          VerifyAgentIDCutoverResponse.$_createMessage);
  static VerifyAgentIDCutoverResponse? _defaultInstance;

  /// True when every check passed and the legacy compatibility paths are safe
  /// to retire in this deployment.
  @$pb.TagNumber(1)
  $core.bool get passed => $_getBF(0);
  @$pb.TagNumber(1)
  set passed($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPassed() => $_has(0);
  @$pb.TagNumber(1)
  void clearPassed() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<AgentIDCutoverFinding> get findings => $_getList(1);
}

class InvokeAgentRequest extends $pb.GeneratedMessage {
  factory InvokeAgentRequest({
    @$core.Deprecated('This field is deprecated.') $core.String? agentName,
    $core.String? input,
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $core.String? modelOverride,
    $core.String? agentId,
  }) {
    final result = InvokeAgentRequest._();
    if (agentName != null) result.agentName = agentName;
    if (input != null) result.input = input;
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (modelOverride != null) result.modelOverride = modelOverride;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  InvokeAgentRequest._();

  factory InvokeAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InvokeAgentRequest()..mergeFromBuffer(data, registry);
  factory InvokeAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InvokeAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InvokeAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: InvokeAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentName')
    ..aOS(2, _omitFieldNames ? '' : 'input')
    ..aOS(3, _omitFieldNames ? '' : 'appName')
    ..aOS(4, _omitFieldNames ? '' : 'userId')
    ..aOS(5, _omitFieldNames ? '' : 'sessionId')
    ..aOS(6, _omitFieldNames ? '' : 'modelOverride')
    ..aOS(7, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvokeAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvokeAgentRequest copyWith(void Function(InvokeAgentRequest) updates) =>
      super.copyWith((message) => updates(message as InvokeAgentRequest))
          as InvokeAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use InvokeAgentRequest() / InvokeAgentRequest.new instead')
  static InvokeAgentRequest create() => InvokeAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => InvokeAgentRequest._();
  @$core.override
  InvokeAgentRequest createEmptyInstance() => InvokeAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static InvokeAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InvokeAgentRequest>(
          InvokeAgentRequest.$_createMessage);
  static InvokeAgentRequest? _defaultInstance;

  /// Legacy agent name. Deprecated (issue #241): never resolved as an input;
  /// agent_id is the sole agent reference.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get agentName => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set agentName($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasAgentName() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearAgentName() => $_clearField(1);

  /// Input message text (required).
  @$pb.TagNumber(2)
  $core.String get input => $_getSZ(1);
  @$pb.TagNumber(2)
  set input($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInput() => $_has(1);
  @$pb.TagNumber(2)
  void clearInput() => $_clearField(2);

  /// App / channel name for the invocation context. Defaults to "api" if empty.
  @$pb.TagNumber(3)
  $core.String get appName => $_getSZ(2);
  @$pb.TagNumber(3)
  set appName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAppName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAppName() => $_clearField(3);

  /// User id for the invocation context. Defaults to "api" if empty.
  @$pb.TagNumber(4)
  $core.String get userId => $_getSZ(3);
  @$pb.TagNumber(4)
  set userId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasUserId() => $_has(3);
  @$pb.TagNumber(4)
  void clearUserId() => $_clearField(4);

  /// Existing session id. Empty triggers a new ephemeral session id.
  @$pb.TagNumber(5)
  $core.String get sessionId => $_getSZ(4);
  @$pb.TagNumber(5)
  set sessionId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSessionId() => $_has(4);
  @$pb.TagNumber(5)
  void clearSessionId() => $_clearField(5);

  /// Optional model override alias or full name.
  @$pb.TagNumber(6)
  $core.String get modelOverride => $_getSZ(5);
  @$pb.TagNumber(6)
  set modelOverride($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasModelOverride() => $_has(5);
  @$pb.TagNumber(6)
  void clearModelOverride() => $_clearField(6);

  /// Immutable agent_id of the agent to invoke. Preferred over agent_name.
  @$pb.TagNumber(7)
  $core.String get agentId => $_getSZ(6);
  @$pb.TagNumber(7)
  set agentId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasAgentId() => $_has(6);
  @$pb.TagNumber(7)
  void clearAgentId() => $_clearField(7);
}

class InvokeAgentResponse extends $pb.GeneratedMessage {
  factory InvokeAgentResponse({
    $core.String? sessionId,
    $core.String? response,
  }) {
    final result = InvokeAgentResponse._();
    if (sessionId != null) result.sessionId = sessionId;
    if (response != null) result.response = response;
    return result;
  }

  InvokeAgentResponse._();

  factory InvokeAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InvokeAgentResponse()..mergeFromBuffer(data, registry);
  factory InvokeAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InvokeAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InvokeAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: InvokeAgentResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..aOS(2, _omitFieldNames ? '' : 'response')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvokeAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvokeAgentResponse copyWith(void Function(InvokeAgentResponse) updates) =>
      super.copyWith((message) => updates(message as InvokeAgentResponse))
          as InvokeAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use InvokeAgentResponse() / InvokeAgentResponse.new instead')
  static InvokeAgentResponse create() => InvokeAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => InvokeAgentResponse._();
  @$core.override
  InvokeAgentResponse createEmptyInstance() => InvokeAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static InvokeAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InvokeAgentResponse>(
          InvokeAgentResponse.$_createMessage);
  static InvokeAgentResponse? _defaultInstance;

  /// Session id used (echoed back, may have been auto-generated).
  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);

  /// Final agent response text.
  @$pb.TagNumber(2)
  $core.String get response => $_getSZ(1);
  @$pb.TagNumber(2)
  set response($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasResponse() => $_has(1);
  @$pb.TagNumber(2)
  void clearResponse() => $_clearField(2);
}

/// Invocation is a persisted record of a single Runner.Run call.
class Invocation extends $pb.GeneratedMessage {
  factory Invocation({
    $core.String? id,
    $core.String? agentName,
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    InvocationStatus? status,
    $core.String? input,
    $core.String? output,
    $core.String? error,
    $1.Timestamp? startedAt,
    $1.Timestamp? finishedAt,
    $fixnum.Int64? latencyMs,
    $core.String? modelOverride,
    $core.String? source,
    $core.String? agentId,
    $core.String? agentDisplayName,
    $core.String? requestId,
    $core.String? workspaceId,
  }) {
    final result = Invocation._();
    if (id != null) result.id = id;
    if (agentName != null) result.agentName = agentName;
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (status != null) result.status = status;
    if (input != null) result.input = input;
    if (output != null) result.output = output;
    if (error != null) result.error = error;
    if (startedAt != null) result.startedAt = startedAt;
    if (finishedAt != null) result.finishedAt = finishedAt;
    if (latencyMs != null) result.latencyMs = latencyMs;
    if (modelOverride != null) result.modelOverride = modelOverride;
    if (source != null) result.source = source;
    if (agentId != null) result.agentId = agentId;
    if (agentDisplayName != null) result.agentDisplayName = agentDisplayName;
    if (requestId != null) result.requestId = requestId;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  Invocation._();

  factory Invocation.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Invocation()..mergeFromBuffer(data, registry);
  factory Invocation.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Invocation()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Invocation',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: Invocation.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'agentName')
    ..aOS(3, _omitFieldNames ? '' : 'appName')
    ..aOS(4, _omitFieldNames ? '' : 'userId')
    ..aOS(5, _omitFieldNames ? '' : 'sessionId')
    ..aE<InvocationStatus>(6, _omitFieldNames ? '' : 'status',
        enumValues: InvocationStatus.values)
    ..aOS(7, _omitFieldNames ? '' : 'input')
    ..aOS(8, _omitFieldNames ? '' : 'output')
    ..aOS(9, _omitFieldNames ? '' : 'error')
    ..aOM<$1.Timestamp>(10, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOM<$1.Timestamp>(11, _omitFieldNames ? '' : 'finishedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aInt64(12, _omitFieldNames ? '' : 'latencyMs')
    ..aOS(13, _omitFieldNames ? '' : 'modelOverride')
    ..aOS(14, _omitFieldNames ? '' : 'source')
    ..aOS(15, _omitFieldNames ? '' : 'agentId')
    ..aOS(16, _omitFieldNames ? '' : 'agentDisplayName')
    ..aOS(17, _omitFieldNames ? '' : 'requestId')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Invocation clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Invocation copyWith(void Function(Invocation) updates) =>
      super.copyWith((message) => updates(message as Invocation)) as Invocation;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Invocation() / Invocation.new instead')
  static Invocation create() => Invocation._();
  static $pb.GeneratedMessage $_createMessage() => Invocation._();
  @$core.override
  Invocation createEmptyInstance() => Invocation._();
  @$core.pragma('dart2js:noInline')
  static Invocation getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Invocation>(Invocation.$_createMessage);
  static Invocation? _defaultInstance;

  /// @gotags: bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: bson:"agent_name,omitempty"
  @$pb.TagNumber(2)
  $core.String get agentName => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentName() => $_clearField(2);

  /// @gotags: bson:"app_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get appName => $_getSZ(2);
  @$pb.TagNumber(3)
  set appName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAppName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAppName() => $_clearField(3);

  /// @gotags: bson:"user_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get userId => $_getSZ(3);
  @$pb.TagNumber(4)
  set userId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasUserId() => $_has(3);
  @$pb.TagNumber(4)
  void clearUserId() => $_clearField(4);

  /// @gotags: bson:"session_id,omitempty"
  @$pb.TagNumber(5)
  $core.String get sessionId => $_getSZ(4);
  @$pb.TagNumber(5)
  set sessionId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSessionId() => $_has(4);
  @$pb.TagNumber(5)
  void clearSessionId() => $_clearField(5);

  /// @gotags: bson:"status,omitempty"
  @$pb.TagNumber(6)
  InvocationStatus get status => $_getN(5);
  @$pb.TagNumber(6)
  set status(InvocationStatus value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);

  /// @gotags: bson:"input,omitempty"
  @$pb.TagNumber(7)
  $core.String get input => $_getSZ(6);
  @$pb.TagNumber(7)
  set input($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInput() => $_has(6);
  @$pb.TagNumber(7)
  void clearInput() => $_clearField(7);

  /// @gotags: bson:"output,omitempty"
  @$pb.TagNumber(8)
  $core.String get output => $_getSZ(7);
  @$pb.TagNumber(8)
  set output($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasOutput() => $_has(7);
  @$pb.TagNumber(8)
  void clearOutput() => $_clearField(8);

  /// @gotags: bson:"error,omitempty"
  @$pb.TagNumber(9)
  $core.String get error => $_getSZ(8);
  @$pb.TagNumber(9)
  set error($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasError() => $_has(8);
  @$pb.TagNumber(9)
  void clearError() => $_clearField(9);

  /// @gotags: bson:"started_at,omitempty"
  @$pb.TagNumber(10)
  $1.Timestamp get startedAt => $_getN(9);
  @$pb.TagNumber(10)
  set startedAt($1.Timestamp value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasStartedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearStartedAt() => $_clearField(10);
  @$pb.TagNumber(10)
  $1.Timestamp ensureStartedAt() => $_ensure(9);

  /// @gotags: bson:"finished_at,omitempty"
  @$pb.TagNumber(11)
  $1.Timestamp get finishedAt => $_getN(10);
  @$pb.TagNumber(11)
  set finishedAt($1.Timestamp value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasFinishedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearFinishedAt() => $_clearField(11);
  @$pb.TagNumber(11)
  $1.Timestamp ensureFinishedAt() => $_ensure(10);

  /// @gotags: bson:"latency_ms,omitempty"
  @$pb.TagNumber(12)
  $fixnum.Int64 get latencyMs => $_getI64(11);
  @$pb.TagNumber(12)
  set latencyMs($fixnum.Int64 value) => $_setInt64(11, value);
  @$pb.TagNumber(12)
  $core.bool hasLatencyMs() => $_has(11);
  @$pb.TagNumber(12)
  void clearLatencyMs() => $_clearField(12);

  /// @gotags: bson:"model_override,omitempty"
  @$pb.TagNumber(13)
  $core.String get modelOverride => $_getSZ(12);
  @$pb.TagNumber(13)
  set modelOverride($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasModelOverride() => $_has(12);
  @$pb.TagNumber(13)
  void clearModelOverride() => $_clearField(13);

  /// @gotags: bson:"source,omitempty"
  @$pb.TagNumber(14)
  $core.String get source => $_getSZ(13);
  @$pb.TagNumber(14)
  set source($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasSource() => $_has(13);
  @$pb.TagNumber(14)
  void clearSource() => $_clearField(14);

  /// Immutable agent_id of the invoked agent. Empty on historical records
  /// written before the Agent ID migration.
  /// @gotags: bson:"agent_id,omitempty"
  @$pb.TagNumber(15)
  $core.String get agentId => $_getSZ(14);
  @$pb.TagNumber(15)
  set agentId($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasAgentId() => $_has(14);
  @$pb.TagNumber(15)
  void clearAgentId() => $_clearField(15);

  /// Display-name snapshot taken at invocation time so historical rows stay
  /// readable after renames.
  /// @gotags: bson:"agent_display_name,omitempty"
  @$pb.TagNumber(16)
  $core.String get agentDisplayName => $_getSZ(15);
  @$pb.TagNumber(16)
  set agentDisplayName($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasAgentDisplayName() => $_has(15);
  @$pb.TagNumber(16)
  void clearAgentDisplayName() => $_clearField(16);

  /// Idempotency key of an owned record: a detached AG-UI run's thread-scoped
  /// runId. Records of the retired async chat carry its client-generated key.
  /// @gotags: bson:"request_id,omitempty"
  @$pb.TagNumber(17)
  $core.String get requestId => $_getSZ(16);
  @$pb.TagNumber(17)
  set requestId($core.String value) => $_setString(16, value);
  @$pb.TagNumber(17)
  $core.bool hasRequestId() => $_has(16);
  @$pb.TagNumber(17)
  void clearRequestId() => $_clearField(17);

  /// @gotags: bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(17);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(17, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(17);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListAgentInvocationsRequest extends $pb.GeneratedMessage {
  factory ListAgentInvocationsRequest({
    $core.String? agentName,
    $core.String? sessionId,
    $core.int? pageSize,
    $core.String? pageToken,
    $core.String? agentId,
  }) {
    final result = ListAgentInvocationsRequest._();
    if (agentName != null) result.agentName = agentName;
    if (sessionId != null) result.sessionId = sessionId;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  ListAgentInvocationsRequest._();

  factory ListAgentInvocationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentInvocationsRequest()..mergeFromBuffer(data, registry);
  factory ListAgentInvocationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentInvocationsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentInvocationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentInvocationsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentName')
    ..aOS(2, _omitFieldNames ? '' : 'sessionId')
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..aOS(4, _omitFieldNames ? '' : 'pageToken')
    ..aOS(5, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentInvocationsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentInvocationsRequest copyWith(
          void Function(ListAgentInvocationsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentInvocationsRequest))
          as ListAgentInvocationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentInvocationsRequest() / ListAgentInvocationsRequest.new instead')
  static ListAgentInvocationsRequest create() =>
      ListAgentInvocationsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentInvocationsRequest._();
  @$core.override
  ListAgentInvocationsRequest createEmptyInstance() =>
      ListAgentInvocationsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAgentInvocationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentInvocationsRequest>(
          ListAgentInvocationsRequest.$_createMessage);
  static ListAgentInvocationsRequest? _defaultInstance;

  /// Legacy: filter by agent name. Deprecated: use agent_id.
  @$pb.TagNumber(1)
  $core.String get agentName => $_getSZ(0);
  @$pb.TagNumber(1)
  set agentName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentName() => $_clearField(1);

  /// Optional: filter by session id.
  @$pb.TagNumber(2)
  $core.String get sessionId => $_getSZ(1);
  @$pb.TagNumber(2)
  set sessionId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSessionId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSessionId() => $_clearField(2);

  /// Maximum records per page. Defaults to 20.
  @$pb.TagNumber(3)
  $core.int get pageSize => $_getIZ(2);
  @$pb.TagNumber(3)
  set pageSize($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageSize() => $_clearField(3);

  /// Pagination token from a previous response.
  @$pb.TagNumber(4)
  $core.String get pageToken => $_getSZ(3);
  @$pb.TagNumber(4)
  set pageToken($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPageToken() => $_has(3);
  @$pb.TagNumber(4)
  void clearPageToken() => $_clearField(4);

  /// Optional: filter by immutable agent_id. Preferred over agent_name.
  @$pb.TagNumber(5)
  $core.String get agentId => $_getSZ(4);
  @$pb.TagNumber(5)
  set agentId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAgentId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAgentId() => $_clearField(5);
}

class ListAgentInvocationsResponse extends $pb.GeneratedMessage {
  factory ListAgentInvocationsResponse({
    $core.Iterable<Invocation>? invocations,
    $core.String? nextPageToken,
    $core.int? total,
  }) {
    final result = ListAgentInvocationsResponse._();
    if (invocations != null) result.invocations.addAll(invocations);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    if (total != null) result.total = total;
    return result;
  }

  ListAgentInvocationsResponse._();

  factory ListAgentInvocationsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentInvocationsResponse()..mergeFromBuffer(data, registry);
  factory ListAgentInvocationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentInvocationsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentInvocationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentInvocationsResponse.$_createMessage)
    ..pPM<Invocation>(1, _omitFieldNames ? '' : 'invocations',
        subBuilder: Invocation.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..aI(3, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentInvocationsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentInvocationsResponse copyWith(
          void Function(ListAgentInvocationsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentInvocationsResponse))
          as ListAgentInvocationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentInvocationsResponse() / ListAgentInvocationsResponse.new instead')
  static ListAgentInvocationsResponse create() =>
      ListAgentInvocationsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentInvocationsResponse._();
  @$core.override
  ListAgentInvocationsResponse createEmptyInstance() =>
      ListAgentInvocationsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAgentInvocationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentInvocationsResponse>(
          ListAgentInvocationsResponse.$_createMessage);
  static ListAgentInvocationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Invocation> get invocations => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get total => $_getIZ(2);
  @$pb.TagNumber(3)
  set total($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTotal() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotal() => $_clearField(3);
}

class GetAgentInvocationRequest extends $pb.GeneratedMessage {
  factory GetAgentInvocationRequest({
    $core.String? invocationId,
  }) {
    final result = GetAgentInvocationRequest._();
    if (invocationId != null) result.invocationId = invocationId;
    return result;
  }

  GetAgentInvocationRequest._();

  factory GetAgentInvocationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentInvocationRequest()..mergeFromBuffer(data, registry);
  factory GetAgentInvocationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentInvocationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentInvocationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentInvocationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'invocationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentInvocationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentInvocationRequest copyWith(
          void Function(GetAgentInvocationRequest) updates) =>
      super.copyWith((message) => updates(message as GetAgentInvocationRequest))
          as GetAgentInvocationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentInvocationRequest() / GetAgentInvocationRequest.new instead')
  static GetAgentInvocationRequest create() => GetAgentInvocationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAgentInvocationRequest._();
  @$core.override
  GetAgentInvocationRequest createEmptyInstance() =>
      GetAgentInvocationRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAgentInvocationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentInvocationRequest>(
          GetAgentInvocationRequest.$_createMessage);
  static GetAgentInvocationRequest? _defaultInstance;

  /// Required.
  @$pb.TagNumber(1)
  $core.String get invocationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set invocationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocationId() => $_clearField(1);
}

class GetAgentInvocationResponse extends $pb.GeneratedMessage {
  factory GetAgentInvocationResponse({
    Invocation? invocation,
  }) {
    final result = GetAgentInvocationResponse._();
    if (invocation != null) result.invocation = invocation;
    return result;
  }

  GetAgentInvocationResponse._();

  factory GetAgentInvocationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentInvocationResponse()..mergeFromBuffer(data, registry);
  factory GetAgentInvocationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentInvocationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentInvocationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentInvocationResponse.$_createMessage)
    ..aOM<Invocation>(1, _omitFieldNames ? '' : 'invocation',
        subBuilder: Invocation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentInvocationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentInvocationResponse copyWith(
          void Function(GetAgentInvocationResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetAgentInvocationResponse))
          as GetAgentInvocationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentInvocationResponse() / GetAgentInvocationResponse.new instead')
  static GetAgentInvocationResponse create() => GetAgentInvocationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAgentInvocationResponse._();
  @$core.override
  GetAgentInvocationResponse createEmptyInstance() =>
      GetAgentInvocationResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAgentInvocationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentInvocationResponse>(
          GetAgentInvocationResponse.$_createMessage);
  static GetAgentInvocationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Invocation get invocation => $_getN(0);
  @$pb.TagNumber(1)
  set invocation(Invocation value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasInvocation() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvocation() => $_clearField(1);
  @$pb.TagNumber(1)
  Invocation ensureInvocation() => $_ensure(0);
}

class ListMCPServersRequest extends $pb.GeneratedMessage {
  factory ListMCPServersRequest() => ListMCPServersRequest._();

  ListMCPServersRequest._();

  factory ListMCPServersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPServersRequest()..mergeFromBuffer(data, registry);
  factory ListMCPServersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPServersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMCPServersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListMCPServersRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPServersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPServersRequest copyWith(
          void Function(ListMCPServersRequest) updates) =>
      super.copyWith((message) => updates(message as ListMCPServersRequest))
          as ListMCPServersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListMCPServersRequest() / ListMCPServersRequest.new instead')
  static ListMCPServersRequest create() => ListMCPServersRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListMCPServersRequest._();
  @$core.override
  ListMCPServersRequest createEmptyInstance() => ListMCPServersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListMCPServersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMCPServersRequest>(
          ListMCPServersRequest.$_createMessage);
  static ListMCPServersRequest? _defaultInstance;
}

class ListMCPServersResponse extends $pb.GeneratedMessage {
  factory ListMCPServersResponse({
    $core.Iterable<$0.MCPServer>? mcpServers,
  }) {
    final result = ListMCPServersResponse._();
    if (mcpServers != null) result.mcpServers.addAll(mcpServers);
    return result;
  }

  ListMCPServersResponse._();

  factory ListMCPServersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPServersResponse()..mergeFromBuffer(data, registry);
  factory ListMCPServersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPServersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMCPServersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListMCPServersResponse.$_createMessage)
    ..pPM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServers',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPServersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPServersResponse copyWith(
          void Function(ListMCPServersResponse) updates) =>
      super.copyWith((message) => updates(message as ListMCPServersResponse))
          as ListMCPServersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListMCPServersResponse() / ListMCPServersResponse.new instead')
  static ListMCPServersResponse create() => ListMCPServersResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListMCPServersResponse._();
  @$core.override
  ListMCPServersResponse createEmptyInstance() => ListMCPServersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListMCPServersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMCPServersResponse>(
          ListMCPServersResponse.$_createMessage);
  static ListMCPServersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$0.MCPServer> get mcpServers => $_getList(0);
}

class GetMCPServerRequest extends $pb.GeneratedMessage {
  factory GetMCPServerRequest({
    $core.String? id,
  }) {
    final result = GetMCPServerRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetMCPServerRequest._();

  factory GetMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerRequest()..mergeFromBuffer(data, registry);
  factory GetMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMCPServerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerRequest copyWith(void Function(GetMCPServerRequest) updates) =>
      super.copyWith((message) => updates(message as GetMCPServerRequest))
          as GetMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetMCPServerRequest() / GetMCPServerRequest.new instead')
  static GetMCPServerRequest create() => GetMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetMCPServerRequest._();
  @$core.override
  GetMCPServerRequest createEmptyInstance() => GetMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static GetMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMCPServerRequest>(
          GetMCPServerRequest.$_createMessage);
  static GetMCPServerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetMCPServerResponse extends $pb.GeneratedMessage {
  factory GetMCPServerResponse({
    $0.MCPServer? mcpServer,
  }) {
    final result = GetMCPServerResponse._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  GetMCPServerResponse._();

  factory GetMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerResponse()..mergeFromBuffer(data, registry);
  factory GetMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMCPServerResponse.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerResponse copyWith(void Function(GetMCPServerResponse) updates) =>
      super.copyWith((message) => updates(message as GetMCPServerResponse))
          as GetMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMCPServerResponse() / GetMCPServerResponse.new instead')
  static GetMCPServerResponse create() => GetMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetMCPServerResponse._();
  @$core.override
  GetMCPServerResponse createEmptyInstance() => GetMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static GetMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMCPServerResponse>(
          GetMCPServerResponse.$_createMessage);
  static GetMCPServerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class CreateMCPServerRequest extends $pb.GeneratedMessage {
  factory CreateMCPServerRequest({
    $0.MCPServer? mcpServer,
  }) {
    final result = CreateMCPServerRequest._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  CreateMCPServerRequest._();

  factory CreateMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateMCPServerRequest()..mergeFromBuffer(data, registry);
  factory CreateMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateMCPServerRequest.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMCPServerRequest copyWith(
          void Function(CreateMCPServerRequest) updates) =>
      super.copyWith((message) => updates(message as CreateMCPServerRequest))
          as CreateMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateMCPServerRequest() / CreateMCPServerRequest.new instead')
  static CreateMCPServerRequest create() => CreateMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateMCPServerRequest._();
  @$core.override
  CreateMCPServerRequest createEmptyInstance() => CreateMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateMCPServerRequest>(
          CreateMCPServerRequest.$_createMessage);
  static CreateMCPServerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class CreateMCPServerResponse extends $pb.GeneratedMessage {
  factory CreateMCPServerResponse({
    $0.MCPServer? mcpServer,
  }) {
    final result = CreateMCPServerResponse._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  CreateMCPServerResponse._();

  factory CreateMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateMCPServerResponse()..mergeFromBuffer(data, registry);
  factory CreateMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateMCPServerResponse.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateMCPServerResponse copyWith(
          void Function(CreateMCPServerResponse) updates) =>
      super.copyWith((message) => updates(message as CreateMCPServerResponse))
          as CreateMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateMCPServerResponse() / CreateMCPServerResponse.new instead')
  static CreateMCPServerResponse create() => CreateMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateMCPServerResponse._();
  @$core.override
  CreateMCPServerResponse createEmptyInstance() => CreateMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateMCPServerResponse>(
          CreateMCPServerResponse.$_createMessage);
  static CreateMCPServerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class UpdateMCPServerRequest extends $pb.GeneratedMessage {
  factory UpdateMCPServerRequest({
    $0.MCPServer? mcpServer,
    $core.Iterable<$core.String>? keepHeaderKeys,
    $core.bool? keepOauthClientSecret,
  }) {
    final result = UpdateMCPServerRequest._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    if (keepHeaderKeys != null) result.keepHeaderKeys.addAll(keepHeaderKeys);
    if (keepOauthClientSecret != null)
      result.keepOauthClientSecret = keepOauthClientSecret;
    return result;
  }

  UpdateMCPServerRequest._();

  factory UpdateMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateMCPServerRequest()..mergeFromBuffer(data, registry);
  factory UpdateMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateMCPServerRequest.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..pPS(2, _omitFieldNames ? '' : 'keepHeaderKeys')
    ..aOB(3, _omitFieldNames ? '' : 'keepOauthClientSecret')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMCPServerRequest copyWith(
          void Function(UpdateMCPServerRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateMCPServerRequest))
          as UpdateMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateMCPServerRequest() / UpdateMCPServerRequest.new instead')
  static UpdateMCPServerRequest create() => UpdateMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateMCPServerRequest._();
  @$core.override
  UpdateMCPServerRequest createEmptyInstance() => UpdateMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateMCPServerRequest>(
          UpdateMCPServerRequest.$_createMessage);
  static UpdateMCPServerRequest? _defaultInstance;

  /// The whole server, as before: an update stores exactly what it is sent.
  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);

  /// Keeps stored credentials the caller cannot read. Reads of a server
  /// installed from a global preset hide header values and the OAuth client
  /// secret, so a whole-object update built from such a read would overwrite
  /// them with the masks. Naming a header here keeps its stored value; the
  /// header must not also appear (with a value) in mcp_server.headers, and
  /// must exist on the stored server. A header left out of both is removed,
  /// as before. Empty keeps the previous behavior.
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get keepHeaderKeys => $_getList(1);

  /// Keeps the stored OAuth client secret. Requires oauth2 auth with the same
  /// client_id as the stored server, because a secret belongs to its client,
  /// and an empty client_secret in mcp_server. When either keep field is used
  /// the response hides credentials the way reads of an installed preset do.
  @$pb.TagNumber(3)
  $core.bool get keepOauthClientSecret => $_getBF(2);
  @$pb.TagNumber(3)
  set keepOauthClientSecret($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasKeepOauthClientSecret() => $_has(2);
  @$pb.TagNumber(3)
  void clearKeepOauthClientSecret() => $_clearField(3);
}

class UpdateMCPServerResponse extends $pb.GeneratedMessage {
  factory UpdateMCPServerResponse({
    $0.MCPServer? mcpServer,
  }) {
    final result = UpdateMCPServerResponse._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  UpdateMCPServerResponse._();

  factory UpdateMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateMCPServerResponse()..mergeFromBuffer(data, registry);
  factory UpdateMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateMCPServerResponse.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateMCPServerResponse copyWith(
          void Function(UpdateMCPServerResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateMCPServerResponse))
          as UpdateMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateMCPServerResponse() / UpdateMCPServerResponse.new instead')
  static UpdateMCPServerResponse create() => UpdateMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateMCPServerResponse._();
  @$core.override
  UpdateMCPServerResponse createEmptyInstance() => UpdateMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateMCPServerResponse>(
          UpdateMCPServerResponse.$_createMessage);
  static UpdateMCPServerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class DeleteMCPServerRequest extends $pb.GeneratedMessage {
  factory DeleteMCPServerRequest({
    $core.String? id,
  }) {
    final result = DeleteMCPServerRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteMCPServerRequest._();

  factory DeleteMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteMCPServerRequest()..mergeFromBuffer(data, registry);
  factory DeleteMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteMCPServerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMCPServerRequest copyWith(
          void Function(DeleteMCPServerRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteMCPServerRequest))
          as DeleteMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteMCPServerRequest() / DeleteMCPServerRequest.new instead')
  static DeleteMCPServerRequest create() => DeleteMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteMCPServerRequest._();
  @$core.override
  DeleteMCPServerRequest createEmptyInstance() => DeleteMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteMCPServerRequest>(
          DeleteMCPServerRequest.$_createMessage);
  static DeleteMCPServerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteMCPServerResponse extends $pb.GeneratedMessage {
  factory DeleteMCPServerResponse() => DeleteMCPServerResponse._();

  DeleteMCPServerResponse._();

  factory DeleteMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteMCPServerResponse()..mergeFromBuffer(data, registry);
  factory DeleteMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteMCPServerResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteMCPServerResponse copyWith(
          void Function(DeleteMCPServerResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteMCPServerResponse))
          as DeleteMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteMCPServerResponse() / DeleteMCPServerResponse.new instead')
  static DeleteMCPServerResponse create() => DeleteMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteMCPServerResponse._();
  @$core.override
  DeleteMCPServerResponse createEmptyInstance() => DeleteMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteMCPServerResponse>(
          DeleteMCPServerResponse.$_createMessage);
  static DeleteMCPServerResponse? _defaultInstance;
}

class StartMCPServerOAuthRequest extends $pb.GeneratedMessage {
  factory StartMCPServerOAuthRequest({
    $core.String? serverId,
    $core.String? returnUrl,
  }) {
    final result = StartMCPServerOAuthRequest._();
    if (serverId != null) result.serverId = serverId;
    if (returnUrl != null) result.returnUrl = returnUrl;
    return result;
  }

  StartMCPServerOAuthRequest._();

  factory StartMCPServerOAuthRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartMCPServerOAuthRequest()..mergeFromBuffer(data, registry);
  factory StartMCPServerOAuthRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartMCPServerOAuthRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartMCPServerOAuthRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StartMCPServerOAuthRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'serverId')
    ..aOS(2, _omitFieldNames ? '' : 'returnUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartMCPServerOAuthRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartMCPServerOAuthRequest copyWith(
          void Function(StartMCPServerOAuthRequest) updates) =>
      super.copyWith(
              (message) => updates(message as StartMCPServerOAuthRequest))
          as StartMCPServerOAuthRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StartMCPServerOAuthRequest() / StartMCPServerOAuthRequest.new instead')
  static StartMCPServerOAuthRequest create() => StartMCPServerOAuthRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      StartMCPServerOAuthRequest._();
  @$core.override
  StartMCPServerOAuthRequest createEmptyInstance() =>
      StartMCPServerOAuthRequest._();
  @$core.pragma('dart2js:noInline')
  static StartMCPServerOAuthRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartMCPServerOAuthRequest>(
          StartMCPServerOAuthRequest.$_createMessage);
  static StartMCPServerOAuthRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get serverId => $_getSZ(0);
  @$pb.TagNumber(1)
  set serverId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasServerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearServerId() => $_clearField(1);

  /// Optional dashboard return URL. The server callback will redirect here
  /// after a browser-based flow when provided.
  @$pb.TagNumber(2)
  $core.String get returnUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set returnUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReturnUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearReturnUrl() => $_clearField(2);
}

class StartMCPServerOAuthResponse extends $pb.GeneratedMessage {
  factory StartMCPServerOAuthResponse({
    $core.String? authorizationUrl,
    $core.String? flowId,
  }) {
    final result = StartMCPServerOAuthResponse._();
    if (authorizationUrl != null) result.authorizationUrl = authorizationUrl;
    if (flowId != null) result.flowId = flowId;
    return result;
  }

  StartMCPServerOAuthResponse._();

  factory StartMCPServerOAuthResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartMCPServerOAuthResponse()..mergeFromBuffer(data, registry);
  factory StartMCPServerOAuthResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StartMCPServerOAuthResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartMCPServerOAuthResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: StartMCPServerOAuthResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'authorizationUrl')
    ..aOS(2, _omitFieldNames ? '' : 'flowId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartMCPServerOAuthResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartMCPServerOAuthResponse copyWith(
          void Function(StartMCPServerOAuthResponse) updates) =>
      super.copyWith(
              (message) => updates(message as StartMCPServerOAuthResponse))
          as StartMCPServerOAuthResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use StartMCPServerOAuthResponse() / StartMCPServerOAuthResponse.new instead')
  static StartMCPServerOAuthResponse create() =>
      StartMCPServerOAuthResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      StartMCPServerOAuthResponse._();
  @$core.override
  StartMCPServerOAuthResponse createEmptyInstance() =>
      StartMCPServerOAuthResponse._();
  @$core.pragma('dart2js:noInline')
  static StartMCPServerOAuthResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartMCPServerOAuthResponse>(
          StartMCPServerOAuthResponse.$_createMessage);
  static StartMCPServerOAuthResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get authorizationUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set authorizationUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAuthorizationUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearAuthorizationUrl() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get flowId => $_getSZ(1);
  @$pb.TagNumber(2)
  set flowId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFlowId() => $_has(1);
  @$pb.TagNumber(2)
  void clearFlowId() => $_clearField(2);
}

class CompleteMCPServerOAuthRequest extends $pb.GeneratedMessage {
  factory CompleteMCPServerOAuthRequest({
    $core.String? flowId,
    $core.String? code,
    $core.String? state,
  }) {
    final result = CompleteMCPServerOAuthRequest._();
    if (flowId != null) result.flowId = flowId;
    if (code != null) result.code = code;
    if (state != null) result.state = state;
    return result;
  }

  CompleteMCPServerOAuthRequest._();

  factory CompleteMCPServerOAuthRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteMCPServerOAuthRequest()..mergeFromBuffer(data, registry);
  factory CompleteMCPServerOAuthRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteMCPServerOAuthRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteMCPServerOAuthRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CompleteMCPServerOAuthRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'flowId')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'state')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteMCPServerOAuthRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteMCPServerOAuthRequest copyWith(
          void Function(CompleteMCPServerOAuthRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CompleteMCPServerOAuthRequest))
          as CompleteMCPServerOAuthRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteMCPServerOAuthRequest() / CompleteMCPServerOAuthRequest.new instead')
  static CompleteMCPServerOAuthRequest create() =>
      CompleteMCPServerOAuthRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CompleteMCPServerOAuthRequest._();
  @$core.override
  CompleteMCPServerOAuthRequest createEmptyInstance() =>
      CompleteMCPServerOAuthRequest._();
  @$core.pragma('dart2js:noInline')
  static CompleteMCPServerOAuthRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteMCPServerOAuthRequest>(
          CompleteMCPServerOAuthRequest.$_createMessage);
  static CompleteMCPServerOAuthRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get flowId => $_getSZ(0);
  @$pb.TagNumber(1)
  set flowId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFlowId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFlowId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get state => $_getSZ(2);
  @$pb.TagNumber(3)
  set state($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => $_clearField(3);
}

class CompleteMCPServerOAuthResponse extends $pb.GeneratedMessage {
  factory CompleteMCPServerOAuthResponse({
    MCPOAuthConnectionStatus? status,
  }) {
    final result = CompleteMCPServerOAuthResponse._();
    if (status != null) result.status = status;
    return result;
  }

  CompleteMCPServerOAuthResponse._();

  factory CompleteMCPServerOAuthResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteMCPServerOAuthResponse()..mergeFromBuffer(data, registry);
  factory CompleteMCPServerOAuthResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteMCPServerOAuthResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteMCPServerOAuthResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CompleteMCPServerOAuthResponse.$_createMessage)
    ..aOM<MCPOAuthConnectionStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: MCPOAuthConnectionStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteMCPServerOAuthResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteMCPServerOAuthResponse copyWith(
          void Function(CompleteMCPServerOAuthResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CompleteMCPServerOAuthResponse))
          as CompleteMCPServerOAuthResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteMCPServerOAuthResponse() / CompleteMCPServerOAuthResponse.new instead')
  static CompleteMCPServerOAuthResponse create() =>
      CompleteMCPServerOAuthResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CompleteMCPServerOAuthResponse._();
  @$core.override
  CompleteMCPServerOAuthResponse createEmptyInstance() =>
      CompleteMCPServerOAuthResponse._();
  @$core.pragma('dart2js:noInline')
  static CompleteMCPServerOAuthResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteMCPServerOAuthResponse>(
          CompleteMCPServerOAuthResponse.$_createMessage);
  static CompleteMCPServerOAuthResponse? _defaultInstance;

  @$pb.TagNumber(1)
  MCPOAuthConnectionStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(MCPOAuthConnectionStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  MCPOAuthConnectionStatus ensureStatus() => $_ensure(0);
}

class GetMCPServerOAuthStatusRequest extends $pb.GeneratedMessage {
  factory GetMCPServerOAuthStatusRequest({
    $core.String? serverId,
  }) {
    final result = GetMCPServerOAuthStatusRequest._();
    if (serverId != null) result.serverId = serverId;
    return result;
  }

  GetMCPServerOAuthStatusRequest._();

  factory GetMCPServerOAuthStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerOAuthStatusRequest()..mergeFromBuffer(data, registry);
  factory GetMCPServerOAuthStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerOAuthStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMCPServerOAuthStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMCPServerOAuthStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'serverId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerOAuthStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerOAuthStatusRequest copyWith(
          void Function(GetMCPServerOAuthStatusRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetMCPServerOAuthStatusRequest))
          as GetMCPServerOAuthStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMCPServerOAuthStatusRequest() / GetMCPServerOAuthStatusRequest.new instead')
  static GetMCPServerOAuthStatusRequest create() =>
      GetMCPServerOAuthStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetMCPServerOAuthStatusRequest._();
  @$core.override
  GetMCPServerOAuthStatusRequest createEmptyInstance() =>
      GetMCPServerOAuthStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetMCPServerOAuthStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMCPServerOAuthStatusRequest>(
          GetMCPServerOAuthStatusRequest.$_createMessage);
  static GetMCPServerOAuthStatusRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get serverId => $_getSZ(0);
  @$pb.TagNumber(1)
  set serverId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasServerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearServerId() => $_clearField(1);
}

class GetMCPServerOAuthStatusResponse extends $pb.GeneratedMessage {
  factory GetMCPServerOAuthStatusResponse({
    MCPOAuthConnectionStatus? status,
  }) {
    final result = GetMCPServerOAuthStatusResponse._();
    if (status != null) result.status = status;
    return result;
  }

  GetMCPServerOAuthStatusResponse._();

  factory GetMCPServerOAuthStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerOAuthStatusResponse()..mergeFromBuffer(data, registry);
  factory GetMCPServerOAuthStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerOAuthStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMCPServerOAuthStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMCPServerOAuthStatusResponse.$_createMessage)
    ..aOM<MCPOAuthConnectionStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: MCPOAuthConnectionStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerOAuthStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerOAuthStatusResponse copyWith(
          void Function(GetMCPServerOAuthStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetMCPServerOAuthStatusResponse))
          as GetMCPServerOAuthStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMCPServerOAuthStatusResponse() / GetMCPServerOAuthStatusResponse.new instead')
  static GetMCPServerOAuthStatusResponse create() =>
      GetMCPServerOAuthStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetMCPServerOAuthStatusResponse._();
  @$core.override
  GetMCPServerOAuthStatusResponse createEmptyInstance() =>
      GetMCPServerOAuthStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetMCPServerOAuthStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMCPServerOAuthStatusResponse>(
          GetMCPServerOAuthStatusResponse.$_createMessage);
  static GetMCPServerOAuthStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  MCPOAuthConnectionStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(MCPOAuthConnectionStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  MCPOAuthConnectionStatus ensureStatus() => $_ensure(0);
}

class DisconnectMCPServerOAuthRequest extends $pb.GeneratedMessage {
  factory DisconnectMCPServerOAuthRequest({
    $core.String? serverId,
  }) {
    final result = DisconnectMCPServerOAuthRequest._();
    if (serverId != null) result.serverId = serverId;
    return result;
  }

  DisconnectMCPServerOAuthRequest._();

  factory DisconnectMCPServerOAuthRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisconnectMCPServerOAuthRequest()..mergeFromBuffer(data, registry);
  factory DisconnectMCPServerOAuthRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisconnectMCPServerOAuthRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisconnectMCPServerOAuthRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DisconnectMCPServerOAuthRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'serverId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisconnectMCPServerOAuthRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisconnectMCPServerOAuthRequest copyWith(
          void Function(DisconnectMCPServerOAuthRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DisconnectMCPServerOAuthRequest))
          as DisconnectMCPServerOAuthRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DisconnectMCPServerOAuthRequest() / DisconnectMCPServerOAuthRequest.new instead')
  static DisconnectMCPServerOAuthRequest create() =>
      DisconnectMCPServerOAuthRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DisconnectMCPServerOAuthRequest._();
  @$core.override
  DisconnectMCPServerOAuthRequest createEmptyInstance() =>
      DisconnectMCPServerOAuthRequest._();
  @$core.pragma('dart2js:noInline')
  static DisconnectMCPServerOAuthRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisconnectMCPServerOAuthRequest>(
          DisconnectMCPServerOAuthRequest.$_createMessage);
  static DisconnectMCPServerOAuthRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get serverId => $_getSZ(0);
  @$pb.TagNumber(1)
  set serverId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasServerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearServerId() => $_clearField(1);
}

class DisconnectMCPServerOAuthResponse extends $pb.GeneratedMessage {
  factory DisconnectMCPServerOAuthResponse({
    MCPOAuthConnectionStatus? status,
  }) {
    final result = DisconnectMCPServerOAuthResponse._();
    if (status != null) result.status = status;
    return result;
  }

  DisconnectMCPServerOAuthResponse._();

  factory DisconnectMCPServerOAuthResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisconnectMCPServerOAuthResponse()..mergeFromBuffer(data, registry);
  factory DisconnectMCPServerOAuthResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DisconnectMCPServerOAuthResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisconnectMCPServerOAuthResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DisconnectMCPServerOAuthResponse.$_createMessage)
    ..aOM<MCPOAuthConnectionStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: MCPOAuthConnectionStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisconnectMCPServerOAuthResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisconnectMCPServerOAuthResponse copyWith(
          void Function(DisconnectMCPServerOAuthResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DisconnectMCPServerOAuthResponse))
          as DisconnectMCPServerOAuthResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DisconnectMCPServerOAuthResponse() / DisconnectMCPServerOAuthResponse.new instead')
  static DisconnectMCPServerOAuthResponse create() =>
      DisconnectMCPServerOAuthResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DisconnectMCPServerOAuthResponse._();
  @$core.override
  DisconnectMCPServerOAuthResponse createEmptyInstance() =>
      DisconnectMCPServerOAuthResponse._();
  @$core.pragma('dart2js:noInline')
  static DisconnectMCPServerOAuthResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisconnectMCPServerOAuthResponse>(
          DisconnectMCPServerOAuthResponse.$_createMessage);
  static DisconnectMCPServerOAuthResponse? _defaultInstance;

  @$pb.TagNumber(1)
  MCPOAuthConnectionStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(MCPOAuthConnectionStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  MCPOAuthConnectionStatus ensureStatus() => $_ensure(0);
}

class MCPOAuthConnectionStatus extends $pb.GeneratedMessage {
  factory MCPOAuthConnectionStatus({
    $core.String? serverId,
    MCPOAuthConnectionState? state,
    $core.String? detail,
    $core.Iterable<$core.String>? scopes,
    $1.Timestamp? connectedAt,
    $1.Timestamp? expiresAt,
    $1.Timestamp? checkedAt,
  }) {
    final result = MCPOAuthConnectionStatus._();
    if (serverId != null) result.serverId = serverId;
    if (state != null) result.state = state;
    if (detail != null) result.detail = detail;
    if (scopes != null) result.scopes.addAll(scopes);
    if (connectedAt != null) result.connectedAt = connectedAt;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (checkedAt != null) result.checkedAt = checkedAt;
    return result;
  }

  MCPOAuthConnectionStatus._();

  factory MCPOAuthConnectionStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPOAuthConnectionStatus()..mergeFromBuffer(data, registry);
  factory MCPOAuthConnectionStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPOAuthConnectionStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MCPOAuthConnectionStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MCPOAuthConnectionStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'serverId')
    ..aE<MCPOAuthConnectionState>(2, _omitFieldNames ? '' : 'state',
        enumValues: MCPOAuthConnectionState.values)
    ..aOS(3, _omitFieldNames ? '' : 'detail')
    ..pPS(4, _omitFieldNames ? '' : 'scopes')
    ..aOM<$1.Timestamp>(5, _omitFieldNames ? '' : 'connectedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOM<$1.Timestamp>(6, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOM<$1.Timestamp>(7, _omitFieldNames ? '' : 'checkedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPOAuthConnectionStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPOAuthConnectionStatus copyWith(
          void Function(MCPOAuthConnectionStatus) updates) =>
      super.copyWith((message) => updates(message as MCPOAuthConnectionStatus))
          as MCPOAuthConnectionStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MCPOAuthConnectionStatus() / MCPOAuthConnectionStatus.new instead')
  static MCPOAuthConnectionStatus create() => MCPOAuthConnectionStatus._();
  static $pb.GeneratedMessage $_createMessage() => MCPOAuthConnectionStatus._();
  @$core.override
  MCPOAuthConnectionStatus createEmptyInstance() =>
      MCPOAuthConnectionStatus._();
  @$core.pragma('dart2js:noInline')
  static MCPOAuthConnectionStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MCPOAuthConnectionStatus>(
          MCPOAuthConnectionStatus.$_createMessage);
  static MCPOAuthConnectionStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get serverId => $_getSZ(0);
  @$pb.TagNumber(1)
  set serverId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasServerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearServerId() => $_clearField(1);

  @$pb.TagNumber(2)
  MCPOAuthConnectionState get state => $_getN(1);
  @$pb.TagNumber(2)
  set state(MCPOAuthConnectionState value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get detail => $_getSZ(2);
  @$pb.TagNumber(3)
  set detail($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDetail() => $_has(2);
  @$pb.TagNumber(3)
  void clearDetail() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get scopes => $_getList(3);

  @$pb.TagNumber(5)
  $1.Timestamp get connectedAt => $_getN(4);
  @$pb.TagNumber(5)
  set connectedAt($1.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasConnectedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearConnectedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Timestamp ensureConnectedAt() => $_ensure(4);

  @$pb.TagNumber(6)
  $1.Timestamp get expiresAt => $_getN(5);
  @$pb.TagNumber(6)
  set expiresAt($1.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasExpiresAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearExpiresAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Timestamp ensureExpiresAt() => $_ensure(5);

  @$pb.TagNumber(7)
  $1.Timestamp get checkedAt => $_getN(6);
  @$pb.TagNumber(7)
  set checkedAt($1.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasCheckedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearCheckedAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Timestamp ensureCheckedAt() => $_ensure(6);
}

class ListGlobalMCPServersRequest extends $pb.GeneratedMessage {
  factory ListGlobalMCPServersRequest() => ListGlobalMCPServersRequest._();

  ListGlobalMCPServersRequest._();

  factory ListGlobalMCPServersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGlobalMCPServersRequest()..mergeFromBuffer(data, registry);
  factory ListGlobalMCPServersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGlobalMCPServersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListGlobalMCPServersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListGlobalMCPServersRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGlobalMCPServersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGlobalMCPServersRequest copyWith(
          void Function(ListGlobalMCPServersRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListGlobalMCPServersRequest))
          as ListGlobalMCPServersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListGlobalMCPServersRequest() / ListGlobalMCPServersRequest.new instead')
  static ListGlobalMCPServersRequest create() =>
      ListGlobalMCPServersRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListGlobalMCPServersRequest._();
  @$core.override
  ListGlobalMCPServersRequest createEmptyInstance() =>
      ListGlobalMCPServersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListGlobalMCPServersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListGlobalMCPServersRequest>(
          ListGlobalMCPServersRequest.$_createMessage);
  static ListGlobalMCPServersRequest? _defaultInstance;
}

class ListGlobalMCPServersResponse extends $pb.GeneratedMessage {
  factory ListGlobalMCPServersResponse({
    $core.Iterable<$0.MCPServer>? mcpServers,
  }) {
    final result = ListGlobalMCPServersResponse._();
    if (mcpServers != null) result.mcpServers.addAll(mcpServers);
    return result;
  }

  ListGlobalMCPServersResponse._();

  factory ListGlobalMCPServersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGlobalMCPServersResponse()..mergeFromBuffer(data, registry);
  factory ListGlobalMCPServersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGlobalMCPServersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListGlobalMCPServersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListGlobalMCPServersResponse.$_createMessage)
    ..pPM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServers',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGlobalMCPServersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGlobalMCPServersResponse copyWith(
          void Function(ListGlobalMCPServersResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListGlobalMCPServersResponse))
          as ListGlobalMCPServersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListGlobalMCPServersResponse() / ListGlobalMCPServersResponse.new instead')
  static ListGlobalMCPServersResponse create() =>
      ListGlobalMCPServersResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListGlobalMCPServersResponse._();
  @$core.override
  ListGlobalMCPServersResponse createEmptyInstance() =>
      ListGlobalMCPServersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListGlobalMCPServersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListGlobalMCPServersResponse>(
          ListGlobalMCPServersResponse.$_createMessage);
  static ListGlobalMCPServersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$0.MCPServer> get mcpServers => $_getList(0);
}

class CreateGlobalMCPServerRequest extends $pb.GeneratedMessage {
  factory CreateGlobalMCPServerRequest({
    $0.MCPServer? mcpServer,
  }) {
    final result = CreateGlobalMCPServerRequest._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  CreateGlobalMCPServerRequest._();

  factory CreateGlobalMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGlobalMCPServerRequest()..mergeFromBuffer(data, registry);
  factory CreateGlobalMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGlobalMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateGlobalMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateGlobalMCPServerRequest.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGlobalMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGlobalMCPServerRequest copyWith(
          void Function(CreateGlobalMCPServerRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateGlobalMCPServerRequest))
          as CreateGlobalMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateGlobalMCPServerRequest() / CreateGlobalMCPServerRequest.new instead')
  static CreateGlobalMCPServerRequest create() =>
      CreateGlobalMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateGlobalMCPServerRequest._();
  @$core.override
  CreateGlobalMCPServerRequest createEmptyInstance() =>
      CreateGlobalMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateGlobalMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateGlobalMCPServerRequest>(
          CreateGlobalMCPServerRequest.$_createMessage);
  static CreateGlobalMCPServerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class CreateGlobalMCPServerResponse extends $pb.GeneratedMessage {
  factory CreateGlobalMCPServerResponse({
    $0.MCPServer? mcpServer,
  }) {
    final result = CreateGlobalMCPServerResponse._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  CreateGlobalMCPServerResponse._();

  factory CreateGlobalMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGlobalMCPServerResponse()..mergeFromBuffer(data, registry);
  factory CreateGlobalMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGlobalMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateGlobalMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateGlobalMCPServerResponse.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGlobalMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGlobalMCPServerResponse copyWith(
          void Function(CreateGlobalMCPServerResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateGlobalMCPServerResponse))
          as CreateGlobalMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateGlobalMCPServerResponse() / CreateGlobalMCPServerResponse.new instead')
  static CreateGlobalMCPServerResponse create() =>
      CreateGlobalMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateGlobalMCPServerResponse._();
  @$core.override
  CreateGlobalMCPServerResponse createEmptyInstance() =>
      CreateGlobalMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateGlobalMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateGlobalMCPServerResponse>(
          CreateGlobalMCPServerResponse.$_createMessage);
  static CreateGlobalMCPServerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class UpdateGlobalMCPServerRequest extends $pb.GeneratedMessage {
  factory UpdateGlobalMCPServerRequest({
    $0.MCPServer? mcpServer,
  }) {
    final result = UpdateGlobalMCPServerRequest._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  UpdateGlobalMCPServerRequest._();

  factory UpdateGlobalMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGlobalMCPServerRequest()..mergeFromBuffer(data, registry);
  factory UpdateGlobalMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGlobalMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateGlobalMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateGlobalMCPServerRequest.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGlobalMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGlobalMCPServerRequest copyWith(
          void Function(UpdateGlobalMCPServerRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateGlobalMCPServerRequest))
          as UpdateGlobalMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateGlobalMCPServerRequest() / UpdateGlobalMCPServerRequest.new instead')
  static UpdateGlobalMCPServerRequest create() =>
      UpdateGlobalMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateGlobalMCPServerRequest._();
  @$core.override
  UpdateGlobalMCPServerRequest createEmptyInstance() =>
      UpdateGlobalMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateGlobalMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateGlobalMCPServerRequest>(
          UpdateGlobalMCPServerRequest.$_createMessage);
  static UpdateGlobalMCPServerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class UpdateGlobalMCPServerResponse extends $pb.GeneratedMessage {
  factory UpdateGlobalMCPServerResponse({
    $0.MCPServer? mcpServer,
  }) {
    final result = UpdateGlobalMCPServerResponse._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  UpdateGlobalMCPServerResponse._();

  factory UpdateGlobalMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGlobalMCPServerResponse()..mergeFromBuffer(data, registry);
  factory UpdateGlobalMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGlobalMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateGlobalMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateGlobalMCPServerResponse.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGlobalMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGlobalMCPServerResponse copyWith(
          void Function(UpdateGlobalMCPServerResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateGlobalMCPServerResponse))
          as UpdateGlobalMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateGlobalMCPServerResponse() / UpdateGlobalMCPServerResponse.new instead')
  static UpdateGlobalMCPServerResponse create() =>
      UpdateGlobalMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateGlobalMCPServerResponse._();
  @$core.override
  UpdateGlobalMCPServerResponse createEmptyInstance() =>
      UpdateGlobalMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateGlobalMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateGlobalMCPServerResponse>(
          UpdateGlobalMCPServerResponse.$_createMessage);
  static UpdateGlobalMCPServerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class DeleteGlobalMCPServerRequest extends $pb.GeneratedMessage {
  factory DeleteGlobalMCPServerRequest({
    $core.String? id,
  }) {
    final result = DeleteGlobalMCPServerRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteGlobalMCPServerRequest._();

  factory DeleteGlobalMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGlobalMCPServerRequest()..mergeFromBuffer(data, registry);
  factory DeleteGlobalMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGlobalMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteGlobalMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteGlobalMCPServerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGlobalMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGlobalMCPServerRequest copyWith(
          void Function(DeleteGlobalMCPServerRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteGlobalMCPServerRequest))
          as DeleteGlobalMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteGlobalMCPServerRequest() / DeleteGlobalMCPServerRequest.new instead')
  static DeleteGlobalMCPServerRequest create() =>
      DeleteGlobalMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteGlobalMCPServerRequest._();
  @$core.override
  DeleteGlobalMCPServerRequest createEmptyInstance() =>
      DeleteGlobalMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteGlobalMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteGlobalMCPServerRequest>(
          DeleteGlobalMCPServerRequest.$_createMessage);
  static DeleteGlobalMCPServerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteGlobalMCPServerResponse extends $pb.GeneratedMessage {
  factory DeleteGlobalMCPServerResponse() => DeleteGlobalMCPServerResponse._();

  DeleteGlobalMCPServerResponse._();

  factory DeleteGlobalMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGlobalMCPServerResponse()..mergeFromBuffer(data, registry);
  factory DeleteGlobalMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGlobalMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteGlobalMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteGlobalMCPServerResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGlobalMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGlobalMCPServerResponse copyWith(
          void Function(DeleteGlobalMCPServerResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteGlobalMCPServerResponse))
          as DeleteGlobalMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteGlobalMCPServerResponse() / DeleteGlobalMCPServerResponse.new instead')
  static DeleteGlobalMCPServerResponse create() =>
      DeleteGlobalMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteGlobalMCPServerResponse._();
  @$core.override
  DeleteGlobalMCPServerResponse createEmptyInstance() =>
      DeleteGlobalMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteGlobalMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteGlobalMCPServerResponse>(
          DeleteGlobalMCPServerResponse.$_createMessage);
  static DeleteGlobalMCPServerResponse? _defaultInstance;
}

class InstallGlobalMCPServerRequest extends $pb.GeneratedMessage {
  factory InstallGlobalMCPServerRequest({
    $core.String? id,
    $core.String? workspaceId,
  }) {
    final result = InstallGlobalMCPServerRequest._();
    if (id != null) result.id = id;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  InstallGlobalMCPServerRequest._();

  factory InstallGlobalMCPServerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InstallGlobalMCPServerRequest()..mergeFromBuffer(data, registry);
  factory InstallGlobalMCPServerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InstallGlobalMCPServerRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InstallGlobalMCPServerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: InstallGlobalMCPServerRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InstallGlobalMCPServerRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InstallGlobalMCPServerRequest copyWith(
          void Function(InstallGlobalMCPServerRequest) updates) =>
      super.copyWith(
              (message) => updates(message as InstallGlobalMCPServerRequest))
          as InstallGlobalMCPServerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use InstallGlobalMCPServerRequest() / InstallGlobalMCPServerRequest.new instead')
  static InstallGlobalMCPServerRequest create() =>
      InstallGlobalMCPServerRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      InstallGlobalMCPServerRequest._();
  @$core.override
  InstallGlobalMCPServerRequest createEmptyInstance() =>
      InstallGlobalMCPServerRequest._();
  @$core.pragma('dart2js:noInline')
  static InstallGlobalMCPServerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InstallGlobalMCPServerRequest>(
          InstallGlobalMCPServerRequest.$_createMessage);
  static InstallGlobalMCPServerRequest? _defaultInstance;

  /// ID of the preset to install.
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Optional target workspace. When empty, the caller's active workspace
  /// (X-Workspace-ID header) is used. Admins may pass a different workspace;
  /// the server audit-logs cross-workspace installs.
  @$pb.TagNumber(2)
  $core.String get workspaceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set workspaceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkspaceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkspaceId() => $_clearField(2);
}

class InstallGlobalMCPServerResponse extends $pb.GeneratedMessage {
  factory InstallGlobalMCPServerResponse({
    $0.MCPServer? mcpServer,
  }) {
    final result = InstallGlobalMCPServerResponse._();
    if (mcpServer != null) result.mcpServer = mcpServer;
    return result;
  }

  InstallGlobalMCPServerResponse._();

  factory InstallGlobalMCPServerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InstallGlobalMCPServerResponse()..mergeFromBuffer(data, registry);
  factory InstallGlobalMCPServerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      InstallGlobalMCPServerResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InstallGlobalMCPServerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: InstallGlobalMCPServerResponse.$_createMessage)
    ..aOM<$0.MCPServer>(1, _omitFieldNames ? '' : 'mcpServer',
        subBuilder: $0.MCPServer.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InstallGlobalMCPServerResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InstallGlobalMCPServerResponse copyWith(
          void Function(InstallGlobalMCPServerResponse) updates) =>
      super.copyWith(
              (message) => updates(message as InstallGlobalMCPServerResponse))
          as InstallGlobalMCPServerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use InstallGlobalMCPServerResponse() / InstallGlobalMCPServerResponse.new instead')
  static InstallGlobalMCPServerResponse create() =>
      InstallGlobalMCPServerResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      InstallGlobalMCPServerResponse._();
  @$core.override
  InstallGlobalMCPServerResponse createEmptyInstance() =>
      InstallGlobalMCPServerResponse._();
  @$core.pragma('dart2js:noInline')
  static InstallGlobalMCPServerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InstallGlobalMCPServerResponse>(
          InstallGlobalMCPServerResponse.$_createMessage);
  static InstallGlobalMCPServerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.MCPServer get mcpServer => $_getN(0);
  @$pb.TagNumber(1)
  set mcpServer($0.MCPServer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMcpServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearMcpServer() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.MCPServer ensureMcpServer() => $_ensure(0);
}

class ListModelProvidersRequest extends $pb.GeneratedMessage {
  factory ListModelProvidersRequest() => ListModelProvidersRequest._();

  ListModelProvidersRequest._();

  factory ListModelProvidersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListModelProvidersRequest()..mergeFromBuffer(data, registry);
  factory ListModelProvidersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListModelProvidersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListModelProvidersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListModelProvidersRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListModelProvidersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListModelProvidersRequest copyWith(
          void Function(ListModelProvidersRequest) updates) =>
      super.copyWith((message) => updates(message as ListModelProvidersRequest))
          as ListModelProvidersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListModelProvidersRequest() / ListModelProvidersRequest.new instead')
  static ListModelProvidersRequest create() => ListModelProvidersRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListModelProvidersRequest._();
  @$core.override
  ListModelProvidersRequest createEmptyInstance() =>
      ListModelProvidersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListModelProvidersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListModelProvidersRequest>(
          ListModelProvidersRequest.$_createMessage);
  static ListModelProvidersRequest? _defaultInstance;
}

class ListModelProvidersResponse extends $pb.GeneratedMessage {
  factory ListModelProvidersResponse({
    $core.Iterable<$0.ModelProvider>? modelProviders,
  }) {
    final result = ListModelProvidersResponse._();
    if (modelProviders != null) result.modelProviders.addAll(modelProviders);
    return result;
  }

  ListModelProvidersResponse._();

  factory ListModelProvidersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListModelProvidersResponse()..mergeFromBuffer(data, registry);
  factory ListModelProvidersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListModelProvidersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListModelProvidersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListModelProvidersResponse.$_createMessage)
    ..pPM<$0.ModelProvider>(1, _omitFieldNames ? '' : 'modelProviders',
        subBuilder: $0.ModelProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListModelProvidersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListModelProvidersResponse copyWith(
          void Function(ListModelProvidersResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListModelProvidersResponse))
          as ListModelProvidersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListModelProvidersResponse() / ListModelProvidersResponse.new instead')
  static ListModelProvidersResponse create() => ListModelProvidersResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListModelProvidersResponse._();
  @$core.override
  ListModelProvidersResponse createEmptyInstance() =>
      ListModelProvidersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListModelProvidersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListModelProvidersResponse>(
          ListModelProvidersResponse.$_createMessage);
  static ListModelProvidersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$0.ModelProvider> get modelProviders => $_getList(0);
}

class GetModelProviderRequest extends $pb.GeneratedMessage {
  factory GetModelProviderRequest({
    $core.String? name,
  }) {
    final result = GetModelProviderRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetModelProviderRequest._();

  factory GetModelProviderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetModelProviderRequest()..mergeFromBuffer(data, registry);
  factory GetModelProviderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetModelProviderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetModelProviderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetModelProviderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetModelProviderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetModelProviderRequest copyWith(
          void Function(GetModelProviderRequest) updates) =>
      super.copyWith((message) => updates(message as GetModelProviderRequest))
          as GetModelProviderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetModelProviderRequest() / GetModelProviderRequest.new instead')
  static GetModelProviderRequest create() => GetModelProviderRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetModelProviderRequest._();
  @$core.override
  GetModelProviderRequest createEmptyInstance() => GetModelProviderRequest._();
  @$core.pragma('dart2js:noInline')
  static GetModelProviderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetModelProviderRequest>(
          GetModelProviderRequest.$_createMessage);
  static GetModelProviderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetModelProviderResponse extends $pb.GeneratedMessage {
  factory GetModelProviderResponse({
    $0.ModelProvider? modelProvider,
  }) {
    final result = GetModelProviderResponse._();
    if (modelProvider != null) result.modelProvider = modelProvider;
    return result;
  }

  GetModelProviderResponse._();

  factory GetModelProviderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetModelProviderResponse()..mergeFromBuffer(data, registry);
  factory GetModelProviderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetModelProviderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetModelProviderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetModelProviderResponse.$_createMessage)
    ..aOM<$0.ModelProvider>(1, _omitFieldNames ? '' : 'modelProvider',
        subBuilder: $0.ModelProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetModelProviderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetModelProviderResponse copyWith(
          void Function(GetModelProviderResponse) updates) =>
      super.copyWith((message) => updates(message as GetModelProviderResponse))
          as GetModelProviderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetModelProviderResponse() / GetModelProviderResponse.new instead')
  static GetModelProviderResponse create() => GetModelProviderResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetModelProviderResponse._();
  @$core.override
  GetModelProviderResponse createEmptyInstance() =>
      GetModelProviderResponse._();
  @$core.pragma('dart2js:noInline')
  static GetModelProviderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetModelProviderResponse>(
          GetModelProviderResponse.$_createMessage);
  static GetModelProviderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.ModelProvider get modelProvider => $_getN(0);
  @$pb.TagNumber(1)
  set modelProvider($0.ModelProvider value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasModelProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearModelProvider() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.ModelProvider ensureModelProvider() => $_ensure(0);
}

class CreateModelProviderRequest extends $pb.GeneratedMessage {
  factory CreateModelProviderRequest({
    $0.ModelProvider? modelProvider,
  }) {
    final result = CreateModelProviderRequest._();
    if (modelProvider != null) result.modelProvider = modelProvider;
    return result;
  }

  CreateModelProviderRequest._();

  factory CreateModelProviderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateModelProviderRequest()..mergeFromBuffer(data, registry);
  factory CreateModelProviderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateModelProviderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateModelProviderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateModelProviderRequest.$_createMessage)
    ..aOM<$0.ModelProvider>(1, _omitFieldNames ? '' : 'modelProvider',
        subBuilder: $0.ModelProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateModelProviderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateModelProviderRequest copyWith(
          void Function(CreateModelProviderRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateModelProviderRequest))
          as CreateModelProviderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateModelProviderRequest() / CreateModelProviderRequest.new instead')
  static CreateModelProviderRequest create() => CreateModelProviderRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateModelProviderRequest._();
  @$core.override
  CreateModelProviderRequest createEmptyInstance() =>
      CreateModelProviderRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateModelProviderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateModelProviderRequest>(
          CreateModelProviderRequest.$_createMessage);
  static CreateModelProviderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.ModelProvider get modelProvider => $_getN(0);
  @$pb.TagNumber(1)
  set modelProvider($0.ModelProvider value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasModelProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearModelProvider() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.ModelProvider ensureModelProvider() => $_ensure(0);
}

class CreateModelProviderResponse extends $pb.GeneratedMessage {
  factory CreateModelProviderResponse({
    $0.ModelProvider? modelProvider,
  }) {
    final result = CreateModelProviderResponse._();
    if (modelProvider != null) result.modelProvider = modelProvider;
    return result;
  }

  CreateModelProviderResponse._();

  factory CreateModelProviderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateModelProviderResponse()..mergeFromBuffer(data, registry);
  factory CreateModelProviderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateModelProviderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateModelProviderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateModelProviderResponse.$_createMessage)
    ..aOM<$0.ModelProvider>(1, _omitFieldNames ? '' : 'modelProvider',
        subBuilder: $0.ModelProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateModelProviderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateModelProviderResponse copyWith(
          void Function(CreateModelProviderResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateModelProviderResponse))
          as CreateModelProviderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateModelProviderResponse() / CreateModelProviderResponse.new instead')
  static CreateModelProviderResponse create() =>
      CreateModelProviderResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateModelProviderResponse._();
  @$core.override
  CreateModelProviderResponse createEmptyInstance() =>
      CreateModelProviderResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateModelProviderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateModelProviderResponse>(
          CreateModelProviderResponse.$_createMessage);
  static CreateModelProviderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.ModelProvider get modelProvider => $_getN(0);
  @$pb.TagNumber(1)
  set modelProvider($0.ModelProvider value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasModelProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearModelProvider() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.ModelProvider ensureModelProvider() => $_ensure(0);
}

class UpdateModelProviderRequest extends $pb.GeneratedMessage {
  factory UpdateModelProviderRequest({
    $0.ModelProvider? modelProvider,
  }) {
    final result = UpdateModelProviderRequest._();
    if (modelProvider != null) result.modelProvider = modelProvider;
    return result;
  }

  UpdateModelProviderRequest._();

  factory UpdateModelProviderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateModelProviderRequest()..mergeFromBuffer(data, registry);
  factory UpdateModelProviderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateModelProviderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateModelProviderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateModelProviderRequest.$_createMessage)
    ..aOM<$0.ModelProvider>(1, _omitFieldNames ? '' : 'modelProvider',
        subBuilder: $0.ModelProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateModelProviderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateModelProviderRequest copyWith(
          void Function(UpdateModelProviderRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateModelProviderRequest))
          as UpdateModelProviderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateModelProviderRequest() / UpdateModelProviderRequest.new instead')
  static UpdateModelProviderRequest create() => UpdateModelProviderRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateModelProviderRequest._();
  @$core.override
  UpdateModelProviderRequest createEmptyInstance() =>
      UpdateModelProviderRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateModelProviderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateModelProviderRequest>(
          UpdateModelProviderRequest.$_createMessage);
  static UpdateModelProviderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.ModelProvider get modelProvider => $_getN(0);
  @$pb.TagNumber(1)
  set modelProvider($0.ModelProvider value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasModelProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearModelProvider() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.ModelProvider ensureModelProvider() => $_ensure(0);
}

class UpdateModelProviderResponse extends $pb.GeneratedMessage {
  factory UpdateModelProviderResponse({
    $0.ModelProvider? modelProvider,
  }) {
    final result = UpdateModelProviderResponse._();
    if (modelProvider != null) result.modelProvider = modelProvider;
    return result;
  }

  UpdateModelProviderResponse._();

  factory UpdateModelProviderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateModelProviderResponse()..mergeFromBuffer(data, registry);
  factory UpdateModelProviderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateModelProviderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateModelProviderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateModelProviderResponse.$_createMessage)
    ..aOM<$0.ModelProvider>(1, _omitFieldNames ? '' : 'modelProvider',
        subBuilder: $0.ModelProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateModelProviderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateModelProviderResponse copyWith(
          void Function(UpdateModelProviderResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateModelProviderResponse))
          as UpdateModelProviderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateModelProviderResponse() / UpdateModelProviderResponse.new instead')
  static UpdateModelProviderResponse create() =>
      UpdateModelProviderResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateModelProviderResponse._();
  @$core.override
  UpdateModelProviderResponse createEmptyInstance() =>
      UpdateModelProviderResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateModelProviderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateModelProviderResponse>(
          UpdateModelProviderResponse.$_createMessage);
  static UpdateModelProviderResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.ModelProvider get modelProvider => $_getN(0);
  @$pb.TagNumber(1)
  set modelProvider($0.ModelProvider value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasModelProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearModelProvider() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.ModelProvider ensureModelProvider() => $_ensure(0);
}

class DeleteModelProviderRequest extends $pb.GeneratedMessage {
  factory DeleteModelProviderRequest({
    $core.String? name,
  }) {
    final result = DeleteModelProviderRequest._();
    if (name != null) result.name = name;
    return result;
  }

  DeleteModelProviderRequest._();

  factory DeleteModelProviderRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteModelProviderRequest()..mergeFromBuffer(data, registry);
  factory DeleteModelProviderRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteModelProviderRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteModelProviderRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteModelProviderRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteModelProviderRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteModelProviderRequest copyWith(
          void Function(DeleteModelProviderRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteModelProviderRequest))
          as DeleteModelProviderRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteModelProviderRequest() / DeleteModelProviderRequest.new instead')
  static DeleteModelProviderRequest create() => DeleteModelProviderRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteModelProviderRequest._();
  @$core.override
  DeleteModelProviderRequest createEmptyInstance() =>
      DeleteModelProviderRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteModelProviderRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteModelProviderRequest>(
          DeleteModelProviderRequest.$_createMessage);
  static DeleteModelProviderRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class DeleteModelProviderResponse extends $pb.GeneratedMessage {
  factory DeleteModelProviderResponse() => DeleteModelProviderResponse._();

  DeleteModelProviderResponse._();

  factory DeleteModelProviderResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteModelProviderResponse()..mergeFromBuffer(data, registry);
  factory DeleteModelProviderResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteModelProviderResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteModelProviderResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteModelProviderResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteModelProviderResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteModelProviderResponse copyWith(
          void Function(DeleteModelProviderResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteModelProviderResponse))
          as DeleteModelProviderResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteModelProviderResponse() / DeleteModelProviderResponse.new instead')
  static DeleteModelProviderResponse create() =>
      DeleteModelProviderResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteModelProviderResponse._();
  @$core.override
  DeleteModelProviderResponse createEmptyInstance() =>
      DeleteModelProviderResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteModelProviderResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteModelProviderResponse>(
          DeleteModelProviderResponse.$_createMessage);
  static DeleteModelProviderResponse? _defaultInstance;
}

class ListNotifyGroupsRequest extends $pb.GeneratedMessage {
  factory ListNotifyGroupsRequest() => ListNotifyGroupsRequest._();

  ListNotifyGroupsRequest._();

  factory ListNotifyGroupsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotifyGroupsRequest()..mergeFromBuffer(data, registry);
  factory ListNotifyGroupsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotifyGroupsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListNotifyGroupsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListNotifyGroupsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotifyGroupsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotifyGroupsRequest copyWith(
          void Function(ListNotifyGroupsRequest) updates) =>
      super.copyWith((message) => updates(message as ListNotifyGroupsRequest))
          as ListNotifyGroupsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListNotifyGroupsRequest() / ListNotifyGroupsRequest.new instead')
  static ListNotifyGroupsRequest create() => ListNotifyGroupsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListNotifyGroupsRequest._();
  @$core.override
  ListNotifyGroupsRequest createEmptyInstance() => ListNotifyGroupsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListNotifyGroupsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListNotifyGroupsRequest>(
          ListNotifyGroupsRequest.$_createMessage);
  static ListNotifyGroupsRequest? _defaultInstance;
}

class ListNotifyGroupsResponse extends $pb.GeneratedMessage {
  factory ListNotifyGroupsResponse({
    $core.Iterable<$0.NotifyGroup>? notifyGroups,
  }) {
    final result = ListNotifyGroupsResponse._();
    if (notifyGroups != null) result.notifyGroups.addAll(notifyGroups);
    return result;
  }

  ListNotifyGroupsResponse._();

  factory ListNotifyGroupsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotifyGroupsResponse()..mergeFromBuffer(data, registry);
  factory ListNotifyGroupsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListNotifyGroupsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListNotifyGroupsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListNotifyGroupsResponse.$_createMessage)
    ..pPM<$0.NotifyGroup>(1, _omitFieldNames ? '' : 'notifyGroups',
        subBuilder: $0.NotifyGroup.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotifyGroupsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListNotifyGroupsResponse copyWith(
          void Function(ListNotifyGroupsResponse) updates) =>
      super.copyWith((message) => updates(message as ListNotifyGroupsResponse))
          as ListNotifyGroupsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListNotifyGroupsResponse() / ListNotifyGroupsResponse.new instead')
  static ListNotifyGroupsResponse create() => ListNotifyGroupsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListNotifyGroupsResponse._();
  @$core.override
  ListNotifyGroupsResponse createEmptyInstance() =>
      ListNotifyGroupsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListNotifyGroupsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListNotifyGroupsResponse>(
          ListNotifyGroupsResponse.$_createMessage);
  static ListNotifyGroupsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$0.NotifyGroup> get notifyGroups => $_getList(0);
}

class GetNotifyGroupRequest extends $pb.GeneratedMessage {
  factory GetNotifyGroupRequest({
    $core.String? name,
  }) {
    final result = GetNotifyGroupRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetNotifyGroupRequest._();

  factory GetNotifyGroupRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetNotifyGroupRequest()..mergeFromBuffer(data, registry);
  factory GetNotifyGroupRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetNotifyGroupRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetNotifyGroupRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetNotifyGroupRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetNotifyGroupRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetNotifyGroupRequest copyWith(
          void Function(GetNotifyGroupRequest) updates) =>
      super.copyWith((message) => updates(message as GetNotifyGroupRequest))
          as GetNotifyGroupRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetNotifyGroupRequest() / GetNotifyGroupRequest.new instead')
  static GetNotifyGroupRequest create() => GetNotifyGroupRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetNotifyGroupRequest._();
  @$core.override
  GetNotifyGroupRequest createEmptyInstance() => GetNotifyGroupRequest._();
  @$core.pragma('dart2js:noInline')
  static GetNotifyGroupRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetNotifyGroupRequest>(
          GetNotifyGroupRequest.$_createMessage);
  static GetNotifyGroupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetNotifyGroupResponse extends $pb.GeneratedMessage {
  factory GetNotifyGroupResponse({
    $0.NotifyGroup? notifyGroup,
  }) {
    final result = GetNotifyGroupResponse._();
    if (notifyGroup != null) result.notifyGroup = notifyGroup;
    return result;
  }

  GetNotifyGroupResponse._();

  factory GetNotifyGroupResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetNotifyGroupResponse()..mergeFromBuffer(data, registry);
  factory GetNotifyGroupResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetNotifyGroupResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetNotifyGroupResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetNotifyGroupResponse.$_createMessage)
    ..aOM<$0.NotifyGroup>(1, _omitFieldNames ? '' : 'notifyGroup',
        subBuilder: $0.NotifyGroup.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetNotifyGroupResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetNotifyGroupResponse copyWith(
          void Function(GetNotifyGroupResponse) updates) =>
      super.copyWith((message) => updates(message as GetNotifyGroupResponse))
          as GetNotifyGroupResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetNotifyGroupResponse() / GetNotifyGroupResponse.new instead')
  static GetNotifyGroupResponse create() => GetNotifyGroupResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetNotifyGroupResponse._();
  @$core.override
  GetNotifyGroupResponse createEmptyInstance() => GetNotifyGroupResponse._();
  @$core.pragma('dart2js:noInline')
  static GetNotifyGroupResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetNotifyGroupResponse>(
          GetNotifyGroupResponse.$_createMessage);
  static GetNotifyGroupResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.NotifyGroup get notifyGroup => $_getN(0);
  @$pb.TagNumber(1)
  set notifyGroup($0.NotifyGroup value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasNotifyGroup() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotifyGroup() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.NotifyGroup ensureNotifyGroup() => $_ensure(0);
}

class CreateNotifyGroupRequest extends $pb.GeneratedMessage {
  factory CreateNotifyGroupRequest({
    $0.NotifyGroup? notifyGroup,
  }) {
    final result = CreateNotifyGroupRequest._();
    if (notifyGroup != null) result.notifyGroup = notifyGroup;
    return result;
  }

  CreateNotifyGroupRequest._();

  factory CreateNotifyGroupRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateNotifyGroupRequest()..mergeFromBuffer(data, registry);
  factory CreateNotifyGroupRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateNotifyGroupRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateNotifyGroupRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateNotifyGroupRequest.$_createMessage)
    ..aOM<$0.NotifyGroup>(1, _omitFieldNames ? '' : 'notifyGroup',
        subBuilder: $0.NotifyGroup.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateNotifyGroupRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateNotifyGroupRequest copyWith(
          void Function(CreateNotifyGroupRequest) updates) =>
      super.copyWith((message) => updates(message as CreateNotifyGroupRequest))
          as CreateNotifyGroupRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateNotifyGroupRequest() / CreateNotifyGroupRequest.new instead')
  static CreateNotifyGroupRequest create() => CreateNotifyGroupRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateNotifyGroupRequest._();
  @$core.override
  CreateNotifyGroupRequest createEmptyInstance() =>
      CreateNotifyGroupRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateNotifyGroupRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateNotifyGroupRequest>(
          CreateNotifyGroupRequest.$_createMessage);
  static CreateNotifyGroupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.NotifyGroup get notifyGroup => $_getN(0);
  @$pb.TagNumber(1)
  set notifyGroup($0.NotifyGroup value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasNotifyGroup() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotifyGroup() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.NotifyGroup ensureNotifyGroup() => $_ensure(0);
}

class CreateNotifyGroupResponse extends $pb.GeneratedMessage {
  factory CreateNotifyGroupResponse({
    $0.NotifyGroup? notifyGroup,
  }) {
    final result = CreateNotifyGroupResponse._();
    if (notifyGroup != null) result.notifyGroup = notifyGroup;
    return result;
  }

  CreateNotifyGroupResponse._();

  factory CreateNotifyGroupResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateNotifyGroupResponse()..mergeFromBuffer(data, registry);
  factory CreateNotifyGroupResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateNotifyGroupResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateNotifyGroupResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateNotifyGroupResponse.$_createMessage)
    ..aOM<$0.NotifyGroup>(1, _omitFieldNames ? '' : 'notifyGroup',
        subBuilder: $0.NotifyGroup.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateNotifyGroupResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateNotifyGroupResponse copyWith(
          void Function(CreateNotifyGroupResponse) updates) =>
      super.copyWith((message) => updates(message as CreateNotifyGroupResponse))
          as CreateNotifyGroupResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateNotifyGroupResponse() / CreateNotifyGroupResponse.new instead')
  static CreateNotifyGroupResponse create() => CreateNotifyGroupResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateNotifyGroupResponse._();
  @$core.override
  CreateNotifyGroupResponse createEmptyInstance() =>
      CreateNotifyGroupResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateNotifyGroupResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateNotifyGroupResponse>(
          CreateNotifyGroupResponse.$_createMessage);
  static CreateNotifyGroupResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.NotifyGroup get notifyGroup => $_getN(0);
  @$pb.TagNumber(1)
  set notifyGroup($0.NotifyGroup value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasNotifyGroup() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotifyGroup() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.NotifyGroup ensureNotifyGroup() => $_ensure(0);
}

class UpdateNotifyGroupRequest extends $pb.GeneratedMessage {
  factory UpdateNotifyGroupRequest({
    $0.NotifyGroup? notifyGroup,
  }) {
    final result = UpdateNotifyGroupRequest._();
    if (notifyGroup != null) result.notifyGroup = notifyGroup;
    return result;
  }

  UpdateNotifyGroupRequest._();

  factory UpdateNotifyGroupRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateNotifyGroupRequest()..mergeFromBuffer(data, registry);
  factory UpdateNotifyGroupRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateNotifyGroupRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateNotifyGroupRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateNotifyGroupRequest.$_createMessage)
    ..aOM<$0.NotifyGroup>(1, _omitFieldNames ? '' : 'notifyGroup',
        subBuilder: $0.NotifyGroup.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateNotifyGroupRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateNotifyGroupRequest copyWith(
          void Function(UpdateNotifyGroupRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateNotifyGroupRequest))
          as UpdateNotifyGroupRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateNotifyGroupRequest() / UpdateNotifyGroupRequest.new instead')
  static UpdateNotifyGroupRequest create() => UpdateNotifyGroupRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateNotifyGroupRequest._();
  @$core.override
  UpdateNotifyGroupRequest createEmptyInstance() =>
      UpdateNotifyGroupRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateNotifyGroupRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateNotifyGroupRequest>(
          UpdateNotifyGroupRequest.$_createMessage);
  static UpdateNotifyGroupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.NotifyGroup get notifyGroup => $_getN(0);
  @$pb.TagNumber(1)
  set notifyGroup($0.NotifyGroup value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasNotifyGroup() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotifyGroup() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.NotifyGroup ensureNotifyGroup() => $_ensure(0);
}

class UpdateNotifyGroupResponse extends $pb.GeneratedMessage {
  factory UpdateNotifyGroupResponse({
    $0.NotifyGroup? notifyGroup,
  }) {
    final result = UpdateNotifyGroupResponse._();
    if (notifyGroup != null) result.notifyGroup = notifyGroup;
    return result;
  }

  UpdateNotifyGroupResponse._();

  factory UpdateNotifyGroupResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateNotifyGroupResponse()..mergeFromBuffer(data, registry);
  factory UpdateNotifyGroupResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateNotifyGroupResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateNotifyGroupResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateNotifyGroupResponse.$_createMessage)
    ..aOM<$0.NotifyGroup>(1, _omitFieldNames ? '' : 'notifyGroup',
        subBuilder: $0.NotifyGroup.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateNotifyGroupResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateNotifyGroupResponse copyWith(
          void Function(UpdateNotifyGroupResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateNotifyGroupResponse))
          as UpdateNotifyGroupResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateNotifyGroupResponse() / UpdateNotifyGroupResponse.new instead')
  static UpdateNotifyGroupResponse create() => UpdateNotifyGroupResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateNotifyGroupResponse._();
  @$core.override
  UpdateNotifyGroupResponse createEmptyInstance() =>
      UpdateNotifyGroupResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateNotifyGroupResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateNotifyGroupResponse>(
          UpdateNotifyGroupResponse.$_createMessage);
  static UpdateNotifyGroupResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.NotifyGroup get notifyGroup => $_getN(0);
  @$pb.TagNumber(1)
  set notifyGroup($0.NotifyGroup value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasNotifyGroup() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotifyGroup() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.NotifyGroup ensureNotifyGroup() => $_ensure(0);
}

class DeleteNotifyGroupRequest extends $pb.GeneratedMessage {
  factory DeleteNotifyGroupRequest({
    $core.String? name,
  }) {
    final result = DeleteNotifyGroupRequest._();
    if (name != null) result.name = name;
    return result;
  }

  DeleteNotifyGroupRequest._();

  factory DeleteNotifyGroupRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteNotifyGroupRequest()..mergeFromBuffer(data, registry);
  factory DeleteNotifyGroupRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteNotifyGroupRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteNotifyGroupRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteNotifyGroupRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteNotifyGroupRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteNotifyGroupRequest copyWith(
          void Function(DeleteNotifyGroupRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteNotifyGroupRequest))
          as DeleteNotifyGroupRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteNotifyGroupRequest() / DeleteNotifyGroupRequest.new instead')
  static DeleteNotifyGroupRequest create() => DeleteNotifyGroupRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteNotifyGroupRequest._();
  @$core.override
  DeleteNotifyGroupRequest createEmptyInstance() =>
      DeleteNotifyGroupRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteNotifyGroupRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteNotifyGroupRequest>(
          DeleteNotifyGroupRequest.$_createMessage);
  static DeleteNotifyGroupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class DeleteNotifyGroupResponse extends $pb.GeneratedMessage {
  factory DeleteNotifyGroupResponse() => DeleteNotifyGroupResponse._();

  DeleteNotifyGroupResponse._();

  factory DeleteNotifyGroupResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteNotifyGroupResponse()..mergeFromBuffer(data, registry);
  factory DeleteNotifyGroupResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteNotifyGroupResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteNotifyGroupResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteNotifyGroupResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteNotifyGroupResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteNotifyGroupResponse copyWith(
          void Function(DeleteNotifyGroupResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteNotifyGroupResponse))
          as DeleteNotifyGroupResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteNotifyGroupResponse() / DeleteNotifyGroupResponse.new instead')
  static DeleteNotifyGroupResponse create() => DeleteNotifyGroupResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteNotifyGroupResponse._();
  @$core.override
  DeleteNotifyGroupResponse createEmptyInstance() =>
      DeleteNotifyGroupResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteNotifyGroupResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteNotifyGroupResponse>(
          DeleteNotifyGroupResponse.$_createMessage);
  static DeleteNotifyGroupResponse? _defaultInstance;
}

class ListRemoteAgentsRequest extends $pb.GeneratedMessage {
  factory ListRemoteAgentsRequest() => ListRemoteAgentsRequest._();

  ListRemoteAgentsRequest._();

  factory ListRemoteAgentsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRemoteAgentsRequest()..mergeFromBuffer(data, registry);
  factory ListRemoteAgentsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRemoteAgentsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRemoteAgentsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListRemoteAgentsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRemoteAgentsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRemoteAgentsRequest copyWith(
          void Function(ListRemoteAgentsRequest) updates) =>
      super.copyWith((message) => updates(message as ListRemoteAgentsRequest))
          as ListRemoteAgentsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListRemoteAgentsRequest() / ListRemoteAgentsRequest.new instead')
  static ListRemoteAgentsRequest create() => ListRemoteAgentsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListRemoteAgentsRequest._();
  @$core.override
  ListRemoteAgentsRequest createEmptyInstance() => ListRemoteAgentsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListRemoteAgentsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRemoteAgentsRequest>(
          ListRemoteAgentsRequest.$_createMessage);
  static ListRemoteAgentsRequest? _defaultInstance;
}

class ListRemoteAgentsResponse extends $pb.GeneratedMessage {
  factory ListRemoteAgentsResponse({
    $core.Iterable<$0.RemoteAgent>? remoteAgents,
  }) {
    final result = ListRemoteAgentsResponse._();
    if (remoteAgents != null) result.remoteAgents.addAll(remoteAgents);
    return result;
  }

  ListRemoteAgentsResponse._();

  factory ListRemoteAgentsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRemoteAgentsResponse()..mergeFromBuffer(data, registry);
  factory ListRemoteAgentsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRemoteAgentsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRemoteAgentsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListRemoteAgentsResponse.$_createMessage)
    ..pPM<$0.RemoteAgent>(1, _omitFieldNames ? '' : 'remoteAgents',
        subBuilder: $0.RemoteAgent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRemoteAgentsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRemoteAgentsResponse copyWith(
          void Function(ListRemoteAgentsResponse) updates) =>
      super.copyWith((message) => updates(message as ListRemoteAgentsResponse))
          as ListRemoteAgentsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListRemoteAgentsResponse() / ListRemoteAgentsResponse.new instead')
  static ListRemoteAgentsResponse create() => ListRemoteAgentsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListRemoteAgentsResponse._();
  @$core.override
  ListRemoteAgentsResponse createEmptyInstance() =>
      ListRemoteAgentsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListRemoteAgentsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRemoteAgentsResponse>(
          ListRemoteAgentsResponse.$_createMessage);
  static ListRemoteAgentsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$0.RemoteAgent> get remoteAgents => $_getList(0);
}

class GetRemoteAgentRequest extends $pb.GeneratedMessage {
  factory GetRemoteAgentRequest({
    $core.String? id,
  }) {
    final result = GetRemoteAgentRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetRemoteAgentRequest._();

  factory GetRemoteAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentRequest()..mergeFromBuffer(data, registry);
  factory GetRemoteAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRemoteAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRemoteAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentRequest copyWith(
          void Function(GetRemoteAgentRequest) updates) =>
      super.copyWith((message) => updates(message as GetRemoteAgentRequest))
          as GetRemoteAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRemoteAgentRequest() / GetRemoteAgentRequest.new instead')
  static GetRemoteAgentRequest create() => GetRemoteAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetRemoteAgentRequest._();
  @$core.override
  GetRemoteAgentRequest createEmptyInstance() => GetRemoteAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static GetRemoteAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRemoteAgentRequest>(
          GetRemoteAgentRequest.$_createMessage);
  static GetRemoteAgentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetRemoteAgentResponse extends $pb.GeneratedMessage {
  factory GetRemoteAgentResponse({
    $0.RemoteAgent? remoteAgent,
  }) {
    final result = GetRemoteAgentResponse._();
    if (remoteAgent != null) result.remoteAgent = remoteAgent;
    return result;
  }

  GetRemoteAgentResponse._();

  factory GetRemoteAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentResponse()..mergeFromBuffer(data, registry);
  factory GetRemoteAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRemoteAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRemoteAgentResponse.$_createMessage)
    ..aOM<$0.RemoteAgent>(1, _omitFieldNames ? '' : 'remoteAgent',
        subBuilder: $0.RemoteAgent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentResponse copyWith(
          void Function(GetRemoteAgentResponse) updates) =>
      super.copyWith((message) => updates(message as GetRemoteAgentResponse))
          as GetRemoteAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRemoteAgentResponse() / GetRemoteAgentResponse.new instead')
  static GetRemoteAgentResponse create() => GetRemoteAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetRemoteAgentResponse._();
  @$core.override
  GetRemoteAgentResponse createEmptyInstance() => GetRemoteAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static GetRemoteAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRemoteAgentResponse>(
          GetRemoteAgentResponse.$_createMessage);
  static GetRemoteAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.RemoteAgent get remoteAgent => $_getN(0);
  @$pb.TagNumber(1)
  set remoteAgent($0.RemoteAgent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRemoteAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemoteAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.RemoteAgent ensureRemoteAgent() => $_ensure(0);
}

class CreateRemoteAgentRequest extends $pb.GeneratedMessage {
  factory CreateRemoteAgentRequest({
    $0.RemoteAgent? remoteAgent,
  }) {
    final result = CreateRemoteAgentRequest._();
    if (remoteAgent != null) result.remoteAgent = remoteAgent;
    return result;
  }

  CreateRemoteAgentRequest._();

  factory CreateRemoteAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRemoteAgentRequest()..mergeFromBuffer(data, registry);
  factory CreateRemoteAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRemoteAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateRemoteAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateRemoteAgentRequest.$_createMessage)
    ..aOM<$0.RemoteAgent>(1, _omitFieldNames ? '' : 'remoteAgent',
        subBuilder: $0.RemoteAgent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRemoteAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRemoteAgentRequest copyWith(
          void Function(CreateRemoteAgentRequest) updates) =>
      super.copyWith((message) => updates(message as CreateRemoteAgentRequest))
          as CreateRemoteAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateRemoteAgentRequest() / CreateRemoteAgentRequest.new instead')
  static CreateRemoteAgentRequest create() => CreateRemoteAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateRemoteAgentRequest._();
  @$core.override
  CreateRemoteAgentRequest createEmptyInstance() =>
      CreateRemoteAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateRemoteAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateRemoteAgentRequest>(
          CreateRemoteAgentRequest.$_createMessage);
  static CreateRemoteAgentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.RemoteAgent get remoteAgent => $_getN(0);
  @$pb.TagNumber(1)
  set remoteAgent($0.RemoteAgent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRemoteAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemoteAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.RemoteAgent ensureRemoteAgent() => $_ensure(0);
}

class CreateRemoteAgentResponse extends $pb.GeneratedMessage {
  factory CreateRemoteAgentResponse({
    $0.RemoteAgent? remoteAgent,
  }) {
    final result = CreateRemoteAgentResponse._();
    if (remoteAgent != null) result.remoteAgent = remoteAgent;
    return result;
  }

  CreateRemoteAgentResponse._();

  factory CreateRemoteAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRemoteAgentResponse()..mergeFromBuffer(data, registry);
  factory CreateRemoteAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateRemoteAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateRemoteAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateRemoteAgentResponse.$_createMessage)
    ..aOM<$0.RemoteAgent>(1, _omitFieldNames ? '' : 'remoteAgent',
        subBuilder: $0.RemoteAgent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRemoteAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateRemoteAgentResponse copyWith(
          void Function(CreateRemoteAgentResponse) updates) =>
      super.copyWith((message) => updates(message as CreateRemoteAgentResponse))
          as CreateRemoteAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateRemoteAgentResponse() / CreateRemoteAgentResponse.new instead')
  static CreateRemoteAgentResponse create() => CreateRemoteAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateRemoteAgentResponse._();
  @$core.override
  CreateRemoteAgentResponse createEmptyInstance() =>
      CreateRemoteAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateRemoteAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateRemoteAgentResponse>(
          CreateRemoteAgentResponse.$_createMessage);
  static CreateRemoteAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.RemoteAgent get remoteAgent => $_getN(0);
  @$pb.TagNumber(1)
  set remoteAgent($0.RemoteAgent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRemoteAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemoteAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.RemoteAgent ensureRemoteAgent() => $_ensure(0);
}

class UpdateRemoteAgentRequest extends $pb.GeneratedMessage {
  factory UpdateRemoteAgentRequest({
    $0.RemoteAgent? remoteAgent,
  }) {
    final result = UpdateRemoteAgentRequest._();
    if (remoteAgent != null) result.remoteAgent = remoteAgent;
    return result;
  }

  UpdateRemoteAgentRequest._();

  factory UpdateRemoteAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRemoteAgentRequest()..mergeFromBuffer(data, registry);
  factory UpdateRemoteAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRemoteAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateRemoteAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateRemoteAgentRequest.$_createMessage)
    ..aOM<$0.RemoteAgent>(1, _omitFieldNames ? '' : 'remoteAgent',
        subBuilder: $0.RemoteAgent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRemoteAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRemoteAgentRequest copyWith(
          void Function(UpdateRemoteAgentRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateRemoteAgentRequest))
          as UpdateRemoteAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateRemoteAgentRequest() / UpdateRemoteAgentRequest.new instead')
  static UpdateRemoteAgentRequest create() => UpdateRemoteAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateRemoteAgentRequest._();
  @$core.override
  UpdateRemoteAgentRequest createEmptyInstance() =>
      UpdateRemoteAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateRemoteAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateRemoteAgentRequest>(
          UpdateRemoteAgentRequest.$_createMessage);
  static UpdateRemoteAgentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $0.RemoteAgent get remoteAgent => $_getN(0);
  @$pb.TagNumber(1)
  set remoteAgent($0.RemoteAgent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRemoteAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemoteAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.RemoteAgent ensureRemoteAgent() => $_ensure(0);
}

class UpdateRemoteAgentResponse extends $pb.GeneratedMessage {
  factory UpdateRemoteAgentResponse({
    $0.RemoteAgent? remoteAgent,
  }) {
    final result = UpdateRemoteAgentResponse._();
    if (remoteAgent != null) result.remoteAgent = remoteAgent;
    return result;
  }

  UpdateRemoteAgentResponse._();

  factory UpdateRemoteAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRemoteAgentResponse()..mergeFromBuffer(data, registry);
  factory UpdateRemoteAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateRemoteAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateRemoteAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateRemoteAgentResponse.$_createMessage)
    ..aOM<$0.RemoteAgent>(1, _omitFieldNames ? '' : 'remoteAgent',
        subBuilder: $0.RemoteAgent.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRemoteAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateRemoteAgentResponse copyWith(
          void Function(UpdateRemoteAgentResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateRemoteAgentResponse))
          as UpdateRemoteAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateRemoteAgentResponse() / UpdateRemoteAgentResponse.new instead')
  static UpdateRemoteAgentResponse create() => UpdateRemoteAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateRemoteAgentResponse._();
  @$core.override
  UpdateRemoteAgentResponse createEmptyInstance() =>
      UpdateRemoteAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateRemoteAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateRemoteAgentResponse>(
          UpdateRemoteAgentResponse.$_createMessage);
  static UpdateRemoteAgentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $0.RemoteAgent get remoteAgent => $_getN(0);
  @$pb.TagNumber(1)
  set remoteAgent($0.RemoteAgent value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRemoteAgent() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemoteAgent() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.RemoteAgent ensureRemoteAgent() => $_ensure(0);
}

class DeleteRemoteAgentRequest extends $pb.GeneratedMessage {
  factory DeleteRemoteAgentRequest({
    $core.String? id,
  }) {
    final result = DeleteRemoteAgentRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteRemoteAgentRequest._();

  factory DeleteRemoteAgentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRemoteAgentRequest()..mergeFromBuffer(data, registry);
  factory DeleteRemoteAgentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRemoteAgentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteRemoteAgentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteRemoteAgentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRemoteAgentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRemoteAgentRequest copyWith(
          void Function(DeleteRemoteAgentRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteRemoteAgentRequest))
          as DeleteRemoteAgentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteRemoteAgentRequest() / DeleteRemoteAgentRequest.new instead')
  static DeleteRemoteAgentRequest create() => DeleteRemoteAgentRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteRemoteAgentRequest._();
  @$core.override
  DeleteRemoteAgentRequest createEmptyInstance() =>
      DeleteRemoteAgentRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteRemoteAgentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteRemoteAgentRequest>(
          DeleteRemoteAgentRequest.$_createMessage);
  static DeleteRemoteAgentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteRemoteAgentResponse extends $pb.GeneratedMessage {
  factory DeleteRemoteAgentResponse() => DeleteRemoteAgentResponse._();

  DeleteRemoteAgentResponse._();

  factory DeleteRemoteAgentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRemoteAgentResponse()..mergeFromBuffer(data, registry);
  factory DeleteRemoteAgentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteRemoteAgentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteRemoteAgentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteRemoteAgentResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRemoteAgentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteRemoteAgentResponse copyWith(
          void Function(DeleteRemoteAgentResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteRemoteAgentResponse))
          as DeleteRemoteAgentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteRemoteAgentResponse() / DeleteRemoteAgentResponse.new instead')
  static DeleteRemoteAgentResponse create() => DeleteRemoteAgentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteRemoteAgentResponse._();
  @$core.override
  DeleteRemoteAgentResponse createEmptyInstance() =>
      DeleteRemoteAgentResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteRemoteAgentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteRemoteAgentResponse>(
          DeleteRemoteAgentResponse.$_createMessage);
  static DeleteRemoteAgentResponse? _defaultInstance;
}

class ListChannelsRequest extends $pb.GeneratedMessage {
  factory ListChannelsRequest() => ListChannelsRequest._();

  ListChannelsRequest._();

  factory ListChannelsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListChannelsRequest()..mergeFromBuffer(data, registry);
  factory ListChannelsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListChannelsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListChannelsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListChannelsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListChannelsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListChannelsRequest copyWith(void Function(ListChannelsRequest) updates) =>
      super.copyWith((message) => updates(message as ListChannelsRequest))
          as ListChannelsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListChannelsRequest() / ListChannelsRequest.new instead')
  static ListChannelsRequest create() => ListChannelsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListChannelsRequest._();
  @$core.override
  ListChannelsRequest createEmptyInstance() => ListChannelsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListChannelsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListChannelsRequest>(
          ListChannelsRequest.$_createMessage);
  static ListChannelsRequest? _defaultInstance;
}

class ListChannelsResponse extends $pb.GeneratedMessage {
  factory ListChannelsResponse({
    $core.Iterable<$5.AgentChannel>? channels,
  }) {
    final result = ListChannelsResponse._();
    if (channels != null) result.channels.addAll(channels);
    return result;
  }

  ListChannelsResponse._();

  factory ListChannelsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListChannelsResponse()..mergeFromBuffer(data, registry);
  factory ListChannelsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListChannelsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListChannelsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListChannelsResponse.$_createMessage)
    ..pPM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channels',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListChannelsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListChannelsResponse copyWith(void Function(ListChannelsResponse) updates) =>
      super.copyWith((message) => updates(message as ListChannelsResponse))
          as ListChannelsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListChannelsResponse() / ListChannelsResponse.new instead')
  static ListChannelsResponse create() => ListChannelsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListChannelsResponse._();
  @$core.override
  ListChannelsResponse createEmptyInstance() => ListChannelsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListChannelsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListChannelsResponse>(
          ListChannelsResponse.$_createMessage);
  static ListChannelsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$5.AgentChannel> get channels => $_getList(0);
}

class GetChannelRequest extends $pb.GeneratedMessage {
  factory GetChannelRequest({
    $core.String? name,
  }) {
    final result = GetChannelRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetChannelRequest._();

  factory GetChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelRequest()..mergeFromBuffer(data, registry);
  factory GetChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelRequest copyWith(void Function(GetChannelRequest) updates) =>
      super.copyWith((message) => updates(message as GetChannelRequest))
          as GetChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetChannelRequest() / GetChannelRequest.new instead')
  static GetChannelRequest create() => GetChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetChannelRequest._();
  @$core.override
  GetChannelRequest createEmptyInstance() => GetChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static GetChannelRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetChannelRequest>(
          GetChannelRequest.$_createMessage);
  static GetChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetChannelResponse extends $pb.GeneratedMessage {
  factory GetChannelResponse({
    $5.AgentChannel? channel,
  }) {
    final result = GetChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  GetChannelResponse._();

  factory GetChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelResponse()..mergeFromBuffer(data, registry);
  factory GetChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetChannelResponse.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelResponse copyWith(void Function(GetChannelResponse) updates) =>
      super.copyWith((message) => updates(message as GetChannelResponse))
          as GetChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetChannelResponse() / GetChannelResponse.new instead')
  static GetChannelResponse create() => GetChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetChannelResponse._();
  @$core.override
  GetChannelResponse createEmptyInstance() => GetChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static GetChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetChannelResponse>(
          GetChannelResponse.$_createMessage);
  static GetChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class CreateChannelRequest extends $pb.GeneratedMessage {
  factory CreateChannelRequest({
    $5.AgentChannel? channel,
  }) {
    final result = CreateChannelRequest._();
    if (channel != null) result.channel = channel;
    return result;
  }

  CreateChannelRequest._();

  factory CreateChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateChannelRequest()..mergeFromBuffer(data, registry);
  factory CreateChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateChannelRequest.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateChannelRequest copyWith(void Function(CreateChannelRequest) updates) =>
      super.copyWith((message) => updates(message as CreateChannelRequest))
          as CreateChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateChannelRequest() / CreateChannelRequest.new instead')
  static CreateChannelRequest create() => CreateChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateChannelRequest._();
  @$core.override
  CreateChannelRequest createEmptyInstance() => CreateChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateChannelRequest>(
          CreateChannelRequest.$_createMessage);
  static CreateChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class CreateChannelResponse extends $pb.GeneratedMessage {
  factory CreateChannelResponse({
    $5.AgentChannel? channel,
  }) {
    final result = CreateChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  CreateChannelResponse._();

  factory CreateChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateChannelResponse()..mergeFromBuffer(data, registry);
  factory CreateChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateChannelResponse.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateChannelResponse copyWith(
          void Function(CreateChannelResponse) updates) =>
      super.copyWith((message) => updates(message as CreateChannelResponse))
          as CreateChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateChannelResponse() / CreateChannelResponse.new instead')
  static CreateChannelResponse create() => CreateChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateChannelResponse._();
  @$core.override
  CreateChannelResponse createEmptyInstance() => CreateChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateChannelResponse>(
          CreateChannelResponse.$_createMessage);
  static CreateChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class UpdateChannelRequest extends $pb.GeneratedMessage {
  factory UpdateChannelRequest({
    $5.AgentChannel? channel,
  }) {
    final result = UpdateChannelRequest._();
    if (channel != null) result.channel = channel;
    return result;
  }

  UpdateChannelRequest._();

  factory UpdateChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateChannelRequest()..mergeFromBuffer(data, registry);
  factory UpdateChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateChannelRequest.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateChannelRequest copyWith(void Function(UpdateChannelRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateChannelRequest))
          as UpdateChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateChannelRequest() / UpdateChannelRequest.new instead')
  static UpdateChannelRequest create() => UpdateChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateChannelRequest._();
  @$core.override
  UpdateChannelRequest createEmptyInstance() => UpdateChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateChannelRequest>(
          UpdateChannelRequest.$_createMessage);
  static UpdateChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class UpdateChannelResponse extends $pb.GeneratedMessage {
  factory UpdateChannelResponse({
    $5.AgentChannel? channel,
  }) {
    final result = UpdateChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  UpdateChannelResponse._();

  factory UpdateChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateChannelResponse()..mergeFromBuffer(data, registry);
  factory UpdateChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateChannelResponse.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateChannelResponse copyWith(
          void Function(UpdateChannelResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateChannelResponse))
          as UpdateChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateChannelResponse() / UpdateChannelResponse.new instead')
  static UpdateChannelResponse create() => UpdateChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateChannelResponse._();
  @$core.override
  UpdateChannelResponse createEmptyInstance() => UpdateChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateChannelResponse>(
          UpdateChannelResponse.$_createMessage);
  static UpdateChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class DeleteChannelRequest extends $pb.GeneratedMessage {
  factory DeleteChannelRequest({
    $core.String? name,
  }) {
    final result = DeleteChannelRequest._();
    if (name != null) result.name = name;
    return result;
  }

  DeleteChannelRequest._();

  factory DeleteChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteChannelRequest()..mergeFromBuffer(data, registry);
  factory DeleteChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteChannelRequest copyWith(void Function(DeleteChannelRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteChannelRequest))
          as DeleteChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteChannelRequest() / DeleteChannelRequest.new instead')
  static DeleteChannelRequest create() => DeleteChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteChannelRequest._();
  @$core.override
  DeleteChannelRequest createEmptyInstance() => DeleteChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteChannelRequest>(
          DeleteChannelRequest.$_createMessage);
  static DeleteChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class DeleteChannelResponse extends $pb.GeneratedMessage {
  factory DeleteChannelResponse() => DeleteChannelResponse._();

  DeleteChannelResponse._();

  factory DeleteChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteChannelResponse()..mergeFromBuffer(data, registry);
  factory DeleteChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteChannelResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteChannelResponse copyWith(
          void Function(DeleteChannelResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteChannelResponse))
          as DeleteChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteChannelResponse() / DeleteChannelResponse.new instead')
  static DeleteChannelResponse create() => DeleteChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteChannelResponse._();
  @$core.override
  DeleteChannelResponse createEmptyInstance() => DeleteChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteChannelResponse>(
          DeleteChannelResponse.$_createMessage);
  static DeleteChannelResponse? _defaultInstance;
}

class RestartChannelRequest extends $pb.GeneratedMessage {
  factory RestartChannelRequest({
    $core.String? name,
  }) {
    final result = RestartChannelRequest._();
    if (name != null) result.name = name;
    return result;
  }

  RestartChannelRequest._();

  factory RestartChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestartChannelRequest()..mergeFromBuffer(data, registry);
  factory RestartChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestartChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestartChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RestartChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestartChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestartChannelRequest copyWith(
          void Function(RestartChannelRequest) updates) =>
      super.copyWith((message) => updates(message as RestartChannelRequest))
          as RestartChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestartChannelRequest() / RestartChannelRequest.new instead')
  static RestartChannelRequest create() => RestartChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => RestartChannelRequest._();
  @$core.override
  RestartChannelRequest createEmptyInstance() => RestartChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static RestartChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestartChannelRequest>(
          RestartChannelRequest.$_createMessage);
  static RestartChannelRequest? _defaultInstance;

  /// Optional channel name; if empty, restarts the whole channel manager.
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class RestartChannelResponse extends $pb.GeneratedMessage {
  factory RestartChannelResponse({
    $5.AgentChannel? channel,
  }) {
    final result = RestartChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  RestartChannelResponse._();

  factory RestartChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestartChannelResponse()..mergeFromBuffer(data, registry);
  factory RestartChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RestartChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RestartChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RestartChannelResponse.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestartChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RestartChannelResponse copyWith(
          void Function(RestartChannelResponse) updates) =>
      super.copyWith((message) => updates(message as RestartChannelResponse))
          as RestartChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RestartChannelResponse() / RestartChannelResponse.new instead')
  static RestartChannelResponse create() => RestartChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => RestartChannelResponse._();
  @$core.override
  RestartChannelResponse createEmptyInstance() => RestartChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static RestartChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RestartChannelResponse>(
          RestartChannelResponse.$_createMessage);
  static RestartChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class PauseChannelRequest extends $pb.GeneratedMessage {
  factory PauseChannelRequest({
    $core.String? name,
  }) {
    final result = PauseChannelRequest._();
    if (name != null) result.name = name;
    return result;
  }

  PauseChannelRequest._();

  factory PauseChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PauseChannelRequest()..mergeFromBuffer(data, registry);
  factory PauseChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PauseChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PauseChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PauseChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PauseChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PauseChannelRequest copyWith(void Function(PauseChannelRequest) updates) =>
      super.copyWith((message) => updates(message as PauseChannelRequest))
          as PauseChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use PauseChannelRequest() / PauseChannelRequest.new instead')
  static PauseChannelRequest create() => PauseChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => PauseChannelRequest._();
  @$core.override
  PauseChannelRequest createEmptyInstance() => PauseChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static PauseChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PauseChannelRequest>(
          PauseChannelRequest.$_createMessage);
  static PauseChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class PauseChannelResponse extends $pb.GeneratedMessage {
  factory PauseChannelResponse({
    $5.AgentChannel? channel,
  }) {
    final result = PauseChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  PauseChannelResponse._();

  factory PauseChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PauseChannelResponse()..mergeFromBuffer(data, registry);
  factory PauseChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PauseChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PauseChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PauseChannelResponse.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PauseChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PauseChannelResponse copyWith(void Function(PauseChannelResponse) updates) =>
      super.copyWith((message) => updates(message as PauseChannelResponse))
          as PauseChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PauseChannelResponse() / PauseChannelResponse.new instead')
  static PauseChannelResponse create() => PauseChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => PauseChannelResponse._();
  @$core.override
  PauseChannelResponse createEmptyInstance() => PauseChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static PauseChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PauseChannelResponse>(
          PauseChannelResponse.$_createMessage);
  static PauseChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class ResumeChannelRequest extends $pb.GeneratedMessage {
  factory ResumeChannelRequest({
    $core.String? name,
  }) {
    final result = ResumeChannelRequest._();
    if (name != null) result.name = name;
    return result;
  }

  ResumeChannelRequest._();

  factory ResumeChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResumeChannelRequest()..mergeFromBuffer(data, registry);
  factory ResumeChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResumeChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResumeChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResumeChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResumeChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResumeChannelRequest copyWith(void Function(ResumeChannelRequest) updates) =>
      super.copyWith((message) => updates(message as ResumeChannelRequest))
          as ResumeChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResumeChannelRequest() / ResumeChannelRequest.new instead')
  static ResumeChannelRequest create() => ResumeChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() => ResumeChannelRequest._();
  @$core.override
  ResumeChannelRequest createEmptyInstance() => ResumeChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static ResumeChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResumeChannelRequest>(
          ResumeChannelRequest.$_createMessage);
  static ResumeChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class ResumeChannelResponse extends $pb.GeneratedMessage {
  factory ResumeChannelResponse({
    $5.AgentChannel? channel,
  }) {
    final result = ResumeChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  ResumeChannelResponse._();

  factory ResumeChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResumeChannelResponse()..mergeFromBuffer(data, registry);
  factory ResumeChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResumeChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResumeChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResumeChannelResponse.$_createMessage)
    ..aOM<$5.AgentChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: $5.AgentChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResumeChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResumeChannelResponse copyWith(
          void Function(ResumeChannelResponse) updates) =>
      super.copyWith((message) => updates(message as ResumeChannelResponse))
          as ResumeChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResumeChannelResponse() / ResumeChannelResponse.new instead')
  static ResumeChannelResponse create() => ResumeChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() => ResumeChannelResponse._();
  @$core.override
  ResumeChannelResponse createEmptyInstance() => ResumeChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static ResumeChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResumeChannelResponse>(
          ResumeChannelResponse.$_createMessage);
  static ResumeChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $5.AgentChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($5.AgentChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  $5.AgentChannel ensureChannel() => $_ensure(0);
}

class CreateSessionRequest extends $pb.GeneratedMessage {
  factory CreateSessionRequest({
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $6.Struct? state,
  }) {
    final result = CreateSessionRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (state != null) result.state = state;
    return result;
  }

  CreateSessionRequest._();

  factory CreateSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionRequest()..mergeFromBuffer(data, registry);
  factory CreateSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'sessionId')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'state',
        subBuilder: $6.Struct.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionRequest copyWith(void Function(CreateSessionRequest) updates) =>
      super.copyWith((message) => updates(message as CreateSessionRequest))
          as CreateSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSessionRequest() / CreateSessionRequest.new instead')
  static CreateSessionRequest create() => CreateSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateSessionRequest._();
  @$core.override
  CreateSessionRequest createEmptyInstance() => CreateSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSessionRequest>(
          CreateSessionRequest.$_createMessage);
  static CreateSessionRequest? _defaultInstance;

  /// App name (channel name) for the session.
  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  /// User ID owning the session.
  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  /// Optional session ID; auto-generated if empty.
  @$pb.TagNumber(3)
  $core.String get sessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set sessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSessionId() => $_clearField(3);

  /// Optional initial state.
  @$pb.TagNumber(4)
  $6.Struct get state => $_getN(3);
  @$pb.TagNumber(4)
  set state($6.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasState() => $_has(3);
  @$pb.TagNumber(4)
  void clearState() => $_clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureState() => $_ensure(3);
}

class CreateSessionResponse extends $pb.GeneratedMessage {
  factory CreateSessionResponse({
    SessionInfo? session,
  }) {
    final result = CreateSessionResponse._();
    if (session != null) result.session = session;
    return result;
  }

  CreateSessionResponse._();

  factory CreateSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionResponse()..mergeFromBuffer(data, registry);
  factory CreateSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateSessionResponse.$_createMessage)
    ..aOM<SessionInfo>(1, _omitFieldNames ? '' : 'session',
        subBuilder: SessionInfo.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionResponse copyWith(
          void Function(CreateSessionResponse) updates) =>
      super.copyWith((message) => updates(message as CreateSessionResponse))
          as CreateSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSessionResponse() / CreateSessionResponse.new instead')
  static CreateSessionResponse create() => CreateSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateSessionResponse._();
  @$core.override
  CreateSessionResponse createEmptyInstance() => CreateSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSessionResponse>(
          CreateSessionResponse.$_createMessage);
  static CreateSessionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionInfo get session => $_getN(0);
  @$pb.TagNumber(1)
  set session(SessionInfo value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearSession() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionInfo ensureSession() => $_ensure(0);
}

class GetSessionRequest extends $pb.GeneratedMessage {
  factory GetSessionRequest({
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $core.int? numRecentEvents,
  }) {
    final result = GetSessionRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (numRecentEvents != null) result.numRecentEvents = numRecentEvents;
    return result;
  }

  GetSessionRequest._();

  factory GetSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSessionRequest()..mergeFromBuffer(data, registry);
  factory GetSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'sessionId')
    ..aI(4, _omitFieldNames ? '' : 'numRecentEvents')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSessionRequest copyWith(void Function(GetSessionRequest) updates) =>
      super.copyWith((message) => updates(message as GetSessionRequest))
          as GetSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetSessionRequest() / GetSessionRequest.new instead')
  static GetSessionRequest create() => GetSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetSessionRequest._();
  @$core.override
  GetSessionRequest createEmptyInstance() => GetSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static GetSessionRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetSessionRequest>(
          GetSessionRequest.$_createMessage);
  static GetSessionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get sessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set sessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSessionId() => $_clearField(3);

  /// Max number of recent events to include. 0 = all.
  @$pb.TagNumber(4)
  $core.int get numRecentEvents => $_getIZ(3);
  @$pb.TagNumber(4)
  set numRecentEvents($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasNumRecentEvents() => $_has(3);
  @$pb.TagNumber(4)
  void clearNumRecentEvents() => $_clearField(4);
}

class GetSessionResponse extends $pb.GeneratedMessage {
  factory GetSessionResponse({
    SessionDetail? sessionDetail,
  }) {
    final result = GetSessionResponse._();
    if (sessionDetail != null) result.sessionDetail = sessionDetail;
    return result;
  }

  GetSessionResponse._();

  factory GetSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSessionResponse()..mergeFromBuffer(data, registry);
  factory GetSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetSessionResponse.$_createMessage)
    ..aOM<SessionDetail>(1, _omitFieldNames ? '' : 'sessionDetail',
        subBuilder: SessionDetail.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSessionResponse copyWith(void Function(GetSessionResponse) updates) =>
      super.copyWith((message) => updates(message as GetSessionResponse))
          as GetSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetSessionResponse() / GetSessionResponse.new instead')
  static GetSessionResponse create() => GetSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetSessionResponse._();
  @$core.override
  GetSessionResponse createEmptyInstance() => GetSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static GetSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSessionResponse>(
          GetSessionResponse.$_createMessage);
  static GetSessionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionDetail get sessionDetail => $_getN(0);
  @$pb.TagNumber(1)
  set sessionDetail(SessionDetail value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionDetail() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionDetail() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionDetail ensureSessionDetail() => $_ensure(0);
}

class ListSessionsRequest extends $pb.GeneratedMessage {
  factory ListSessionsRequest({
    $core.String? appName,
    $core.String? userId,
    $1.Timestamp? startTime,
    $1.Timestamp? endTime,
    $core.int? pageSize,
    $core.String? pageToken,
    $core.bool? workspaceScoped,
  }) {
    final result = ListSessionsRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (startTime != null) result.startTime = startTime;
    if (endTime != null) result.endTime = endTime;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    if (workspaceScoped != null) result.workspaceScoped = workspaceScoped;
    return result;
  }

  ListSessionsRequest._();

  factory ListSessionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSessionsRequest()..mergeFromBuffer(data, registry);
  factory ListSessionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSessionsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSessionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListSessionsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOM<$1.Timestamp>(3, _omitFieldNames ? '' : 'startTime',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOM<$1.Timestamp>(4, _omitFieldNames ? '' : 'endTime',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aI(5, _omitFieldNames ? '' : 'pageSize')
    ..aOS(6, _omitFieldNames ? '' : 'pageToken')
    ..aOB(7, _omitFieldNames ? '' : 'workspaceScoped')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSessionsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSessionsRequest copyWith(void Function(ListSessionsRequest) updates) =>
      super.copyWith((message) => updates(message as ListSessionsRequest))
          as ListSessionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListSessionsRequest() / ListSessionsRequest.new instead')
  static ListSessionsRequest create() => ListSessionsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListSessionsRequest._();
  @$core.override
  ListSessionsRequest createEmptyInstance() => ListSessionsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListSessionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSessionsRequest>(
          ListSessionsRequest.$_createMessage);
  static ListSessionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  /// Optional: only return sessions whose last_update_time is at or after this.
  @$pb.TagNumber(3)
  $1.Timestamp get startTime => $_getN(2);
  @$pb.TagNumber(3)
  set startTime($1.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasStartTime() => $_has(2);
  @$pb.TagNumber(3)
  void clearStartTime() => $_clearField(3);
  @$pb.TagNumber(3)
  $1.Timestamp ensureStartTime() => $_ensure(2);

  /// Optional: only return sessions whose last_update_time is at or before this.
  @$pb.TagNumber(4)
  $1.Timestamp get endTime => $_getN(3);
  @$pb.TagNumber(4)
  set endTime($1.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasEndTime() => $_has(3);
  @$pb.TagNumber(4)
  void clearEndTime() => $_clearField(4);
  @$pb.TagNumber(4)
  $1.Timestamp ensureEndTime() => $_ensure(3);

  /// Maximum records per page. Defaults to 20.
  @$pb.TagNumber(5)
  $core.int get pageSize => $_getIZ(4);
  @$pb.TagNumber(5)
  set pageSize($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPageSize() => $_has(4);
  @$pb.TagNumber(5)
  void clearPageSize() => $_clearField(5);

  /// Pagination token from a previous response.
  @$pb.TagNumber(6)
  $core.String get pageToken => $_getSZ(5);
  @$pb.TagNumber(6)
  set pageToken($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPageToken() => $_has(5);
  @$pb.TagNumber(6)
  void clearPageToken() => $_clearField(6);

  /// When true, only return sessions that belong to the active workspace
  /// (from X-Workspace-ID header). Legacy sessions without workspace_id are
  /// excluded. When false (default), behaves as before for backwards compat.
  @$pb.TagNumber(7)
  $core.bool get workspaceScoped => $_getBF(6);
  @$pb.TagNumber(7)
  set workspaceScoped($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasWorkspaceScoped() => $_has(6);
  @$pb.TagNumber(7)
  void clearWorkspaceScoped() => $_clearField(7);
}

class ListSessionsResponse extends $pb.GeneratedMessage {
  factory ListSessionsResponse({
    $core.Iterable<SessionInfo>? sessions,
    $core.String? nextPageToken,
    $core.int? total,
  }) {
    final result = ListSessionsResponse._();
    if (sessions != null) result.sessions.addAll(sessions);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    if (total != null) result.total = total;
    return result;
  }

  ListSessionsResponse._();

  factory ListSessionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSessionsResponse()..mergeFromBuffer(data, registry);
  factory ListSessionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSessionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSessionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListSessionsResponse.$_createMessage)
    ..pPM<SessionInfo>(1, _omitFieldNames ? '' : 'sessions',
        subBuilder: SessionInfo.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..aI(3, _omitFieldNames ? '' : 'total')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSessionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSessionsResponse copyWith(void Function(ListSessionsResponse) updates) =>
      super.copyWith((message) => updates(message as ListSessionsResponse))
          as ListSessionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSessionsResponse() / ListSessionsResponse.new instead')
  static ListSessionsResponse create() => ListSessionsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListSessionsResponse._();
  @$core.override
  ListSessionsResponse createEmptyInstance() => ListSessionsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListSessionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSessionsResponse>(
          ListSessionsResponse.$_createMessage);
  static ListSessionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<SessionInfo> get sessions => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get total => $_getIZ(2);
  @$pb.TagNumber(3)
  set total($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTotal() => $_has(2);
  @$pb.TagNumber(3)
  void clearTotal() => $_clearField(3);
}

class DeleteSessionRequest extends $pb.GeneratedMessage {
  factory DeleteSessionRequest({
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
  }) {
    final result = DeleteSessionRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    return result;
  }

  DeleteSessionRequest._();

  factory DeleteSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSessionRequest()..mergeFromBuffer(data, registry);
  factory DeleteSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'sessionId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSessionRequest copyWith(void Function(DeleteSessionRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteSessionRequest))
          as DeleteSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteSessionRequest() / DeleteSessionRequest.new instead')
  static DeleteSessionRequest create() => DeleteSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteSessionRequest._();
  @$core.override
  DeleteSessionRequest createEmptyInstance() => DeleteSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteSessionRequest>(
          DeleteSessionRequest.$_createMessage);
  static DeleteSessionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get sessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set sessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSessionId() => $_clearField(3);
}

class DeleteSessionResponse extends $pb.GeneratedMessage {
  factory DeleteSessionResponse() => DeleteSessionResponse._();

  DeleteSessionResponse._();

  factory DeleteSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSessionResponse()..mergeFromBuffer(data, registry);
  factory DeleteSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteSessionResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSessionResponse copyWith(
          void Function(DeleteSessionResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteSessionResponse))
          as DeleteSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteSessionResponse() / DeleteSessionResponse.new instead')
  static DeleteSessionResponse create() => DeleteSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteSessionResponse._();
  @$core.override
  DeleteSessionResponse createEmptyInstance() => DeleteSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteSessionResponse>(
          DeleteSessionResponse.$_createMessage);
  static DeleteSessionResponse? _defaultInstance;
}

class ReplySessionRequest extends $pb.GeneratedMessage {
  factory ReplySessionRequest({
    @$core.Deprecated('This field is deprecated.') $core.String? agentName,
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $core.String? message,
    $core.String? modelOverride,
    $core.Iterable<$2.InputPart>? parts,
    $core.String? agentId,
  }) {
    final result = ReplySessionRequest._();
    if (agentName != null) result.agentName = agentName;
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (message != null) result.message = message;
    if (modelOverride != null) result.modelOverride = modelOverride;
    if (parts != null) result.parts.addAll(parts);
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  ReplySessionRequest._();

  factory ReplySessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReplySessionRequest()..mergeFromBuffer(data, registry);
  factory ReplySessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReplySessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReplySessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ReplySessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentName')
    ..aOS(2, _omitFieldNames ? '' : 'appName')
    ..aOS(3, _omitFieldNames ? '' : 'userId')
    ..aOS(4, _omitFieldNames ? '' : 'sessionId')
    ..aOS(5, _omitFieldNames ? '' : 'message')
    ..aOS(6, _omitFieldNames ? '' : 'modelOverride')
    ..pPM<$2.InputPart>(7, _omitFieldNames ? '' : 'parts',
        subBuilder: $2.InputPart.$_createMessage)
    ..aOS(8, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReplySessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReplySessionRequest copyWith(void Function(ReplySessionRequest) updates) =>
      super.copyWith((message) => updates(message as ReplySessionRequest))
          as ReplySessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ReplySessionRequest() / ReplySessionRequest.new instead')
  static ReplySessionRequest create() => ReplySessionRequest._();
  static $pb.GeneratedMessage $_createMessage() => ReplySessionRequest._();
  @$core.override
  ReplySessionRequest createEmptyInstance() => ReplySessionRequest._();
  @$core.pragma('dart2js:noInline')
  static ReplySessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReplySessionRequest>(
          ReplySessionRequest.$_createMessage);
  static ReplySessionRequest? _defaultInstance;

  /// Legacy agent name. Deprecated (issue #241): never resolved as an input;
  /// agent_id is the sole agent reference.
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get agentName => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set agentName($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasAgentName() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearAgentName() => $_clearField(1);

  /// App name (channel name) for session lookup.
  @$pb.TagNumber(2)
  $core.String get appName => $_getSZ(1);
  @$pb.TagNumber(2)
  set appName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAppName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAppName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get userId => $_getSZ(2);
  @$pb.TagNumber(3)
  set userId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUserId() => $_has(2);
  @$pb.TagNumber(3)
  void clearUserId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get sessionId => $_getSZ(3);
  @$pb.TagNumber(4)
  set sessionId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSessionId() => $_has(3);
  @$pb.TagNumber(4)
  void clearSessionId() => $_clearField(4);

  /// The user message text to send.
  @$pb.TagNumber(5)
  $core.String get message => $_getSZ(4);
  @$pb.TagNumber(5)
  set message($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasMessage() => $_has(4);
  @$pb.TagNumber(5)
  void clearMessage() => $_clearField(5);

  /// Optional model override (alias or full name).
  @$pb.TagNumber(6)
  $core.String get modelOverride => $_getSZ(5);
  @$pb.TagNumber(6)
  set modelOverride($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasModelOverride() => $_has(5);
  @$pb.TagNumber(6)
  void clearModelOverride() => $_clearField(6);

  /// Multimodal input parts (text and inline images). When non-empty,
  /// `parts` is used as the user input and `message` is ignored. When
  /// empty, `message` is used as before. Validation (MIME whitelist,
  /// size and count limits) happens in the application layer.
  @$pb.TagNumber(7)
  $pb.PbList<$2.InputPart> get parts => $_getList(6);

  /// Immutable agent_id of the agent to invoke. Preferred over agent_name.
  @$pb.TagNumber(8)
  $core.String get agentId => $_getSZ(7);
  @$pb.TagNumber(8)
  set agentId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAgentId() => $_has(7);
  @$pb.TagNumber(8)
  void clearAgentId() => $_clearField(8);
}

class ReplySessionResponse extends $pb.GeneratedMessage {
  factory ReplySessionResponse({
    $core.String? response,
  }) {
    final result = ReplySessionResponse._();
    if (response != null) result.response = response;
    return result;
  }

  ReplySessionResponse._();

  factory ReplySessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReplySessionResponse()..mergeFromBuffer(data, registry);
  factory ReplySessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReplySessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReplySessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ReplySessionResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'response')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReplySessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReplySessionResponse copyWith(void Function(ReplySessionResponse) updates) =>
      super.copyWith((message) => updates(message as ReplySessionResponse))
          as ReplySessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReplySessionResponse() / ReplySessionResponse.new instead')
  static ReplySessionResponse create() => ReplySessionResponse._();
  static $pb.GeneratedMessage $_createMessage() => ReplySessionResponse._();
  @$core.override
  ReplySessionResponse createEmptyInstance() => ReplySessionResponse._();
  @$core.pragma('dart2js:noInline')
  static ReplySessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReplySessionResponse>(
          ReplySessionResponse.$_createMessage);
  static ReplySessionResponse? _defaultInstance;

  /// The agent's response text.
  @$pb.TagNumber(1)
  $core.String get response => $_getSZ(0);
  @$pb.TagNumber(1)
  set response($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasResponse() => $_has(0);
  @$pb.TagNumber(1)
  void clearResponse() => $_clearField(1);
}

class UpdateSessionTitleRequest extends $pb.GeneratedMessage {
  factory UpdateSessionTitleRequest({
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $core.String? title,
  }) {
    final result = UpdateSessionTitleRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (title != null) result.title = title;
    return result;
  }

  UpdateSessionTitleRequest._();

  factory UpdateSessionTitleRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSessionTitleRequest()..mergeFromBuffer(data, registry);
  factory UpdateSessionTitleRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSessionTitleRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateSessionTitleRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateSessionTitleRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'sessionId')
    ..aOS(4, _omitFieldNames ? '' : 'title')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSessionTitleRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSessionTitleRequest copyWith(
          void Function(UpdateSessionTitleRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateSessionTitleRequest))
          as UpdateSessionTitleRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateSessionTitleRequest() / UpdateSessionTitleRequest.new instead')
  static UpdateSessionTitleRequest create() => UpdateSessionTitleRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateSessionTitleRequest._();
  @$core.override
  UpdateSessionTitleRequest createEmptyInstance() =>
      UpdateSessionTitleRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateSessionTitleRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateSessionTitleRequest>(
          UpdateSessionTitleRequest.$_createMessage);
  static UpdateSessionTitleRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get sessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set sessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSessionId() => $_clearField(3);

  /// New title for the session. Trimmed and normalized to one line;
  /// must be non-blank and at most 100 Unicode code points.
  @$pb.TagNumber(4)
  $core.String get title => $_getSZ(3);
  @$pb.TagNumber(4)
  set title($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTitle() => $_has(3);
  @$pb.TagNumber(4)
  void clearTitle() => $_clearField(4);
}

class UpdateSessionTitleResponse extends $pb.GeneratedMessage {
  factory UpdateSessionTitleResponse({
    SessionInfo? session,
  }) {
    final result = UpdateSessionTitleResponse._();
    if (session != null) result.session = session;
    return result;
  }

  UpdateSessionTitleResponse._();

  factory UpdateSessionTitleResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSessionTitleResponse()..mergeFromBuffer(data, registry);
  factory UpdateSessionTitleResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSessionTitleResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateSessionTitleResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateSessionTitleResponse.$_createMessage)
    ..aOM<SessionInfo>(1, _omitFieldNames ? '' : 'session',
        subBuilder: SessionInfo.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSessionTitleResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSessionTitleResponse copyWith(
          void Function(UpdateSessionTitleResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateSessionTitleResponse))
          as UpdateSessionTitleResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateSessionTitleResponse() / UpdateSessionTitleResponse.new instead')
  static UpdateSessionTitleResponse create() => UpdateSessionTitleResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateSessionTitleResponse._();
  @$core.override
  UpdateSessionTitleResponse createEmptyInstance() =>
      UpdateSessionTitleResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateSessionTitleResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateSessionTitleResponse>(
          UpdateSessionTitleResponse.$_createMessage);
  static UpdateSessionTitleResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionInfo get session => $_getN(0);
  @$pb.TagNumber(1)
  set session(SessionInfo value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearSession() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionInfo ensureSession() => $_ensure(0);
}

/// SessionInfo is a lightweight session summary (no events).
class SessionInfo extends $pb.GeneratedMessage {
  factory SessionInfo({
    $core.String? sessionId,
    $core.String? appName,
    $core.String? userId,
    $6.Struct? state,
    $1.Timestamp? lastUpdateTime,
    $core.int? turnCount,
    $core.String? title,
    $core.String? latestInvocationId,
    InvocationStatus? latestInvocationStatus,
    $core.bool? hasPendingInterrupt,
    $core.bool? unread,
    $1.Timestamp? lastReadAt,
    $core.String? workspaceId,
  }) {
    final result = SessionInfo._();
    if (sessionId != null) result.sessionId = sessionId;
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (state != null) result.state = state;
    if (lastUpdateTime != null) result.lastUpdateTime = lastUpdateTime;
    if (turnCount != null) result.turnCount = turnCount;
    if (title != null) result.title = title;
    if (latestInvocationId != null)
      result.latestInvocationId = latestInvocationId;
    if (latestInvocationStatus != null)
      result.latestInvocationStatus = latestInvocationStatus;
    if (hasPendingInterrupt != null)
      result.hasPendingInterrupt = hasPendingInterrupt;
    if (unread != null) result.unread = unread;
    if (lastReadAt != null) result.lastReadAt = lastReadAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  SessionInfo._();

  factory SessionInfo.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionInfo()..mergeFromBuffer(data, registry);
  factory SessionInfo.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionInfo()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SessionInfo',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SessionInfo.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..aOS(2, _omitFieldNames ? '' : 'appName')
    ..aOS(3, _omitFieldNames ? '' : 'userId')
    ..aOM<$6.Struct>(4, _omitFieldNames ? '' : 'state',
        subBuilder: $6.Struct.$_createMessage)
    ..aOM<$1.Timestamp>(5, _omitFieldNames ? '' : 'lastUpdateTime',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aI(6, _omitFieldNames ? '' : 'turnCount')
    ..aOS(7, _omitFieldNames ? '' : 'title')
    ..aOS(8, _omitFieldNames ? '' : 'latestInvocationId')
    ..aE<InvocationStatus>(9, _omitFieldNames ? '' : 'latestInvocationStatus',
        enumValues: InvocationStatus.values)
    ..aOB(10, _omitFieldNames ? '' : 'hasPendingInterrupt')
    ..aOB(11, _omitFieldNames ? '' : 'unread')
    ..aOM<$1.Timestamp>(12, _omitFieldNames ? '' : 'lastReadAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionInfo clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionInfo copyWith(void Function(SessionInfo) updates) =>
      super.copyWith((message) => updates(message as SessionInfo))
          as SessionInfo;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SessionInfo() / SessionInfo.new instead')
  static SessionInfo create() => SessionInfo._();
  static $pb.GeneratedMessage $_createMessage() => SessionInfo._();
  @$core.override
  SessionInfo createEmptyInstance() => SessionInfo._();
  @$core.pragma('dart2js:noInline')
  static SessionInfo getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SessionInfo>(
          SessionInfo.$_createMessage);
  static SessionInfo? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get appName => $_getSZ(1);
  @$pb.TagNumber(2)
  set appName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAppName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAppName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get userId => $_getSZ(2);
  @$pb.TagNumber(3)
  set userId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUserId() => $_has(2);
  @$pb.TagNumber(3)
  void clearUserId() => $_clearField(3);

  @$pb.TagNumber(4)
  $6.Struct get state => $_getN(3);
  @$pb.TagNumber(4)
  set state($6.Struct value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasState() => $_has(3);
  @$pb.TagNumber(4)
  void clearState() => $_clearField(4);
  @$pb.TagNumber(4)
  $6.Struct ensureState() => $_ensure(3);

  @$pb.TagNumber(5)
  $1.Timestamp get lastUpdateTime => $_getN(4);
  @$pb.TagNumber(5)
  set lastUpdateTime($1.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasLastUpdateTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearLastUpdateTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Timestamp ensureLastUpdateTime() => $_ensure(4);

  /// Number of recorded events (rough "turn" count).
  @$pb.TagNumber(6)
  $core.int get turnCount => $_getIZ(5);
  @$pb.TagNumber(6)
  set turnCount($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTurnCount() => $_has(5);
  @$pb.TagNumber(6)
  void clearTurnCount() => $_clearField(6);

  /// First-class session title. Effective-title precedence: this field,
  /// then legacy state["title"], then agent name, then shortened session ID.
  @$pb.TagNumber(7)
  $core.String get title => $_getSZ(6);
  @$pb.TagNumber(7)
  set title($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasTitle() => $_has(6);
  @$pb.TagNumber(7)
  void clearTitle() => $_clearField(7);

  /// Active or latest dashboard Invocation ID for this session. Empty when
  /// the session has no recorded invocations.
  @$pb.TagNumber(8)
  $core.String get latestInvocationId => $_getSZ(7);
  @$pb.TagNumber(8)
  set latestInvocationId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasLatestInvocationId() => $_has(7);
  @$pb.TagNumber(8)
  void clearLatestInvocationId() => $_clearField(8);

  /// Status of the invocation referenced by latest_invocation_id.
  @$pb.TagNumber(9)
  InvocationStatus get latestInvocationStatus => $_getN(8);
  @$pb.TagNumber(9)
  set latestInvocationStatus(InvocationStatus value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasLatestInvocationStatus() => $_has(8);
  @$pb.TagNumber(9)
  void clearLatestInvocationStatus() => $_clearField(9);

  /// True when the session has unanswered Workflow Agent Human Input
  /// Interrupts, derived through the existing interrupt seam (ADR-0002).
  @$pb.TagNumber(10)
  $core.bool get hasPendingInterrupt => $_getBF(9);
  @$pb.TagNumber(10)
  set hasPendingInterrupt($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasHasPendingInterrupt() => $_has(9);
  @$pb.TagNumber(10)
  void clearHasPendingInterrupt() => $_clearField(10);

  /// True when the session contains content the authenticated user has not
  /// yet viewed (last_update_time > last_read_at, or no read record).
  @$pb.TagNumber(11)
  $core.bool get unread => $_getBF(10);
  @$pb.TagNumber(11)
  set unread($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasUnread() => $_has(10);
  @$pb.TagNumber(11)
  void clearUnread() => $_clearField(11);

  /// Server-backed timestamp recording when the authenticated user last
  /// viewed this session's content. Zero when the session has never been
  /// read.
  @$pb.TagNumber(12)
  $1.Timestamp get lastReadAt => $_getN(11);
  @$pb.TagNumber(12)
  set lastReadAt($1.Timestamp value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasLastReadAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearLastReadAt() => $_clearField(12);
  @$pb.TagNumber(12)
  $1.Timestamp ensureLastReadAt() => $_ensure(11);

  /// Workspace that owns this session. Derived from the authenticated
  /// request context at creation time; clients cannot override it.
  /// Empty on legacy sessions created before workspace scoping.
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(12);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(12, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(12);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// SessionDetail includes the session info plus recent events.
class SessionDetail extends $pb.GeneratedMessage {
  factory SessionDetail({
    SessionInfo? session,
    $core.Iterable<SessionEvent>? events,
    $7.Duration? duration,
  }) {
    final result = SessionDetail._();
    if (session != null) result.session = session;
    if (events != null) result.events.addAll(events);
    if (duration != null) result.duration = duration;
    return result;
  }

  SessionDetail._();

  factory SessionDetail.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionDetail()..mergeFromBuffer(data, registry);
  factory SessionDetail.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionDetail()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SessionDetail',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SessionDetail.$_createMessage)
    ..aOM<SessionInfo>(1, _omitFieldNames ? '' : 'session',
        subBuilder: SessionInfo.$_createMessage)
    ..pPM<SessionEvent>(2, _omitFieldNames ? '' : 'events',
        subBuilder: SessionEvent.$_createMessage)
    ..aOM<$7.Duration>(3, _omitFieldNames ? '' : 'duration',
        subBuilder: $7.Duration.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionDetail clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionDetail copyWith(void Function(SessionDetail) updates) =>
      super.copyWith((message) => updates(message as SessionDetail))
          as SessionDetail;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SessionDetail() / SessionDetail.new instead')
  static SessionDetail create() => SessionDetail._();
  static $pb.GeneratedMessage $_createMessage() => SessionDetail._();
  @$core.override
  SessionDetail createEmptyInstance() => SessionDetail._();
  @$core.pragma('dart2js:noInline')
  static SessionDetail getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SessionDetail>(
          SessionDetail.$_createMessage);
  static SessionDetail? _defaultInstance;

  @$pb.TagNumber(1)
  SessionInfo get session => $_getN(0);
  @$pb.TagNumber(1)
  set session(SessionInfo value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearSession() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionInfo ensureSession() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<SessionEvent> get events => $_getList(1);

  /// Wallclock duration between the first and last recorded events.
  @$pb.TagNumber(3)
  $7.Duration get duration => $_getN(2);
  @$pb.TagNumber(3)
  set duration($7.Duration value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDuration() => $_has(2);
  @$pb.TagNumber(3)
  void clearDuration() => $_clearField(3);
  @$pb.TagNumber(3)
  $7.Duration ensureDuration() => $_ensure(2);
}

class GenerateSessionTitleRequest extends $pb.GeneratedMessage {
  factory GenerateSessionTitleRequest({
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
  }) {
    final result = GenerateSessionTitleRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    return result;
  }

  GenerateSessionTitleRequest._();

  factory GenerateSessionTitleRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GenerateSessionTitleRequest()..mergeFromBuffer(data, registry);
  factory GenerateSessionTitleRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GenerateSessionTitleRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GenerateSessionTitleRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GenerateSessionTitleRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'sessionId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateSessionTitleRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateSessionTitleRequest copyWith(
          void Function(GenerateSessionTitleRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GenerateSessionTitleRequest))
          as GenerateSessionTitleRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GenerateSessionTitleRequest() / GenerateSessionTitleRequest.new instead')
  static GenerateSessionTitleRequest create() =>
      GenerateSessionTitleRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GenerateSessionTitleRequest._();
  @$core.override
  GenerateSessionTitleRequest createEmptyInstance() =>
      GenerateSessionTitleRequest._();
  @$core.pragma('dart2js:noInline')
  static GenerateSessionTitleRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GenerateSessionTitleRequest>(
          GenerateSessionTitleRequest.$_createMessage);
  static GenerateSessionTitleRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get sessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set sessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSessionId() => $_clearField(3);
}

class GenerateSessionTitleResponse extends $pb.GeneratedMessage {
  factory GenerateSessionTitleResponse({
    SessionInfo? session,
    $core.bool? generated,
  }) {
    final result = GenerateSessionTitleResponse._();
    if (session != null) result.session = session;
    if (generated != null) result.generated = generated;
    return result;
  }

  GenerateSessionTitleResponse._();

  factory GenerateSessionTitleResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GenerateSessionTitleResponse()..mergeFromBuffer(data, registry);
  factory GenerateSessionTitleResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GenerateSessionTitleResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GenerateSessionTitleResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GenerateSessionTitleResponse.$_createMessage)
    ..aOM<SessionInfo>(1, _omitFieldNames ? '' : 'session',
        subBuilder: SessionInfo.$_createMessage)
    ..aOB(2, _omitFieldNames ? '' : 'generated')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateSessionTitleResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GenerateSessionTitleResponse copyWith(
          void Function(GenerateSessionTitleResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GenerateSessionTitleResponse))
          as GenerateSessionTitleResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GenerateSessionTitleResponse() / GenerateSessionTitleResponse.new instead')
  static GenerateSessionTitleResponse create() =>
      GenerateSessionTitleResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GenerateSessionTitleResponse._();
  @$core.override
  GenerateSessionTitleResponse createEmptyInstance() =>
      GenerateSessionTitleResponse._();
  @$core.pragma('dart2js:noInline')
  static GenerateSessionTitleResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GenerateSessionTitleResponse>(
          GenerateSessionTitleResponse.$_createMessage);
  static GenerateSessionTitleResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionInfo get session => $_getN(0);
  @$pb.TagNumber(1)
  set session(SessionInfo value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearSession() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionInfo ensureSession() => $_ensure(0);

  /// True when this call derived and persisted a new title; false when an
  /// effective title already existed or was set concurrently.
  @$pb.TagNumber(2)
  $core.bool get generated => $_getBF(1);
  @$pb.TagNumber(2)
  set generated($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasGenerated() => $_has(1);
  @$pb.TagNumber(2)
  void clearGenerated() => $_clearField(2);
}

class MarkSessionReadRequest extends $pb.GeneratedMessage {
  factory MarkSessionReadRequest({
    $core.String? appName,
    $core.String? userId,
    $core.String? sessionId,
    $core.String? lastReadEventId,
  }) {
    final result = MarkSessionReadRequest._();
    if (appName != null) result.appName = appName;
    if (userId != null) result.userId = userId;
    if (sessionId != null) result.sessionId = sessionId;
    if (lastReadEventId != null) result.lastReadEventId = lastReadEventId;
    return result;
  }

  MarkSessionReadRequest._();

  factory MarkSessionReadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkSessionReadRequest()..mergeFromBuffer(data, registry);
  factory MarkSessionReadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkSessionReadRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkSessionReadRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MarkSessionReadRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appName')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'sessionId')
    ..aOS(4, _omitFieldNames ? '' : 'lastReadEventId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkSessionReadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkSessionReadRequest copyWith(
          void Function(MarkSessionReadRequest) updates) =>
      super.copyWith((message) => updates(message as MarkSessionReadRequest))
          as MarkSessionReadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MarkSessionReadRequest() / MarkSessionReadRequest.new instead')
  static MarkSessionReadRequest create() => MarkSessionReadRequest._();
  static $pb.GeneratedMessage $_createMessage() => MarkSessionReadRequest._();
  @$core.override
  MarkSessionReadRequest createEmptyInstance() => MarkSessionReadRequest._();
  @$core.pragma('dart2js:noInline')
  static MarkSessionReadRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MarkSessionReadRequest>(
          MarkSessionReadRequest.$_createMessage);
  static MarkSessionReadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appName => $_getSZ(0);
  @$pb.TagNumber(1)
  set appName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppName() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get sessionId => $_getSZ(2);
  @$pb.TagNumber(3)
  set sessionId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSessionId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSessionId() => $_clearField(3);

  /// The event ID the client has rendered up to. The server records the
  /// timestamp: clients that send this only when the session is the active
  /// conversation, the document is visible, and the latest known event or
  /// terminal invocation state has been rendered.
  @$pb.TagNumber(4)
  $core.String get lastReadEventId => $_getSZ(3);
  @$pb.TagNumber(4)
  set lastReadEventId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLastReadEventId() => $_has(3);
  @$pb.TagNumber(4)
  void clearLastReadEventId() => $_clearField(4);
}

class MarkSessionReadResponse extends $pb.GeneratedMessage {
  factory MarkSessionReadResponse({
    SessionInfo? session,
  }) {
    final result = MarkSessionReadResponse._();
    if (session != null) result.session = session;
    return result;
  }

  MarkSessionReadResponse._();

  factory MarkSessionReadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkSessionReadResponse()..mergeFromBuffer(data, registry);
  factory MarkSessionReadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MarkSessionReadResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MarkSessionReadResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MarkSessionReadResponse.$_createMessage)
    ..aOM<SessionInfo>(1, _omitFieldNames ? '' : 'session',
        subBuilder: SessionInfo.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkSessionReadResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MarkSessionReadResponse copyWith(
          void Function(MarkSessionReadResponse) updates) =>
      super.copyWith((message) => updates(message as MarkSessionReadResponse))
          as MarkSessionReadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MarkSessionReadResponse() / MarkSessionReadResponse.new instead')
  static MarkSessionReadResponse create() => MarkSessionReadResponse._();
  static $pb.GeneratedMessage $_createMessage() => MarkSessionReadResponse._();
  @$core.override
  MarkSessionReadResponse createEmptyInstance() => MarkSessionReadResponse._();
  @$core.pragma('dart2js:noInline')
  static MarkSessionReadResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MarkSessionReadResponse>(
          MarkSessionReadResponse.$_createMessage);
  static MarkSessionReadResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionInfo get session => $_getN(0);
  @$pb.TagNumber(1)
  set session(SessionInfo value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearSession() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionInfo ensureSession() => $_ensure(0);
}

/// SessionEvent is a serializable representation of a session event.
class SessionEvent extends $pb.GeneratedMessage {
  factory SessionEvent({
    $core.String? eventId,
    $core.String? invocationId,
    $core.String? author,
    $core.String? branch,
    $core.String? contentJson,
    $1.Timestamp? timestamp,
    $core.String? traceId,
    $core.String? traceUrl,
  }) {
    final result = SessionEvent._();
    if (eventId != null) result.eventId = eventId;
    if (invocationId != null) result.invocationId = invocationId;
    if (author != null) result.author = author;
    if (branch != null) result.branch = branch;
    if (contentJson != null) result.contentJson = contentJson;
    if (timestamp != null) result.timestamp = timestamp;
    if (traceId != null) result.traceId = traceId;
    if (traceUrl != null) result.traceUrl = traceUrl;
    return result;
  }

  SessionEvent._();

  factory SessionEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionEvent()..mergeFromBuffer(data, registry);
  factory SessionEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionEvent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SessionEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SessionEvent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'eventId')
    ..aOS(2, _omitFieldNames ? '' : 'invocationId')
    ..aOS(3, _omitFieldNames ? '' : 'author')
    ..aOS(4, _omitFieldNames ? '' : 'branch')
    ..aOS(5, _omitFieldNames ? '' : 'contentJson')
    ..aOM<$1.Timestamp>(6, _omitFieldNames ? '' : 'timestamp',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOS(7, _omitFieldNames ? '' : 'traceId')
    ..aOS(8, _omitFieldNames ? '' : 'traceUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionEvent copyWith(void Function(SessionEvent) updates) =>
      super.copyWith((message) => updates(message as SessionEvent))
          as SessionEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SessionEvent() / SessionEvent.new instead')
  static SessionEvent create() => SessionEvent._();
  static $pb.GeneratedMessage $_createMessage() => SessionEvent._();
  @$core.override
  SessionEvent createEmptyInstance() => SessionEvent._();
  @$core.pragma('dart2js:noInline')
  static SessionEvent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SessionEvent>(
          SessionEvent.$_createMessage);
  static SessionEvent? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get eventId => $_getSZ(0);
  @$pb.TagNumber(1)
  set eventId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEventId() => $_has(0);
  @$pb.TagNumber(1)
  void clearEventId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get invocationId => $_getSZ(1);
  @$pb.TagNumber(2)
  set invocationId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInvocationId() => $_has(1);
  @$pb.TagNumber(2)
  void clearInvocationId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get author => $_getSZ(2);
  @$pb.TagNumber(3)
  set author($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAuthor() => $_has(2);
  @$pb.TagNumber(3)
  void clearAuthor() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get branch => $_getSZ(3);
  @$pb.TagNumber(4)
  set branch($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBranch() => $_has(3);
  @$pb.TagNumber(4)
  void clearBranch() => $_clearField(4);

  /// Content serialized as JSON string (genai.Content).
  @$pb.TagNumber(5)
  $core.String get contentJson => $_getSZ(4);
  @$pb.TagNumber(5)
  set contentJson($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasContentJson() => $_has(4);
  @$pb.TagNumber(5)
  void clearContentJson() => $_clearField(5);

  @$pb.TagNumber(6)
  $1.Timestamp get timestamp => $_getN(5);
  @$pb.TagNumber(6)
  set timestamp($1.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasTimestamp() => $_has(5);
  @$pb.TagNumber(6)
  void clearTimestamp() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Timestamp ensureTimestamp() => $_ensure(5);

  /// Trace id propagated to observability backends (e.g. Langfuse). Currently
  /// mirrors invocation_id; consumers should treat it as opaque.
  @$pb.TagNumber(7)
  $core.String get traceId => $_getSZ(6);
  @$pb.TagNumber(7)
  set traceId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasTraceId() => $_has(6);
  @$pb.TagNumber(7)
  void clearTraceId() => $_clearField(7);

  /// Fully-qualified URL to the trace view in the configured observability
  /// backend, if one is wired. Empty otherwise.
  @$pb.TagNumber(8)
  $core.String get traceUrl => $_getSZ(7);
  @$pb.TagNumber(8)
  set traceUrl($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasTraceUrl() => $_has(7);
  @$pb.TagNumber(8)
  void clearTraceUrl() => $_clearField(8);
}

/// AgentService manages Agent configurations.
class AgentServiceApi {
  final $pb.RpcClient _client;

  AgentServiceApi(this._client);

  $async.Future<ListAgentsResponse> listAgents(
          $pb.ClientContext? ctx, ListAgentsRequest request) =>
      _client.invoke<ListAgentsResponse>(
          ctx, 'AgentService', 'ListAgents', request, ListAgentsResponse());
  $async.Future<GetAgentResponse> getAgent(
          $pb.ClientContext? ctx, GetAgentRequest request) =>
      _client.invoke<GetAgentResponse>(
          ctx, 'AgentService', 'GetAgent', request, GetAgentResponse());
  $async.Future<CreateAgentResponse> createAgent(
          $pb.ClientContext? ctx, CreateAgentRequest request) =>
      _client.invoke<CreateAgentResponse>(
          ctx, 'AgentService', 'CreateAgent', request, CreateAgentResponse());
  $async.Future<UpdateAgentResponse> updateAgent(
          $pb.ClientContext? ctx, UpdateAgentRequest request) =>
      _client.invoke<UpdateAgentResponse>(
          ctx, 'AgentService', 'UpdateAgent', request, UpdateAgentResponse());
  $async.Future<DeleteAgentResponse> deleteAgent(
          $pb.ClientContext? ctx, DeleteAgentRequest request) =>
      _client.invoke<DeleteAgentResponse>(
          ctx, 'AgentService', 'DeleteAgent', request, DeleteAgentResponse());

  /// InvokeAgent runs an agent once with the given input. If session_id is empty
  /// an ephemeral session id is generated; otherwise the existing session is used.
  $async.Future<InvokeAgentResponse> invokeAgent(
          $pb.ClientContext? ctx, InvokeAgentRequest request) =>
      _client.invoke<InvokeAgentResponse>(
          ctx, 'AgentService', 'InvokeAgent', request, InvokeAgentResponse());

  /// ListAgentInvocations returns recorded invocations, optionally filtered by
  /// agent name, with simple opaque-token pagination.
  $async.Future<ListAgentInvocationsResponse> listAgentInvocations(
          $pb.ClientContext? ctx, ListAgentInvocationsRequest request) =>
      _client.invoke<ListAgentInvocationsResponse>(ctx, 'AgentService',
          'ListAgentInvocations', request, ListAgentInvocationsResponse());

  /// ReloadAgents drops cached runners and re-resolves agents/MCP/remote-agents
  /// from the current configuration. Exposes the runtime hot-reload trigger
  /// used internally by config mutations to the dashboard "Hot-reload" button.
  $async.Future<ReloadAgentsResponse> reloadAgents(
          $pb.ClientContext? ctx, ReloadAgentsRequest request) =>
      _client.invoke<ReloadAgentsResponse>(
          ctx, 'AgentService', 'ReloadAgents', request, ReloadAgentsResponse());

  /// GetAgentRuntimeStatus returns the latest invocation state for an agent,
  /// suitable for the Agents table Status / Last Run columns.
  $async.Future<GetAgentRuntimeStatusResponse> getAgentRuntimeStatus(
          $pb.ClientContext? ctx, GetAgentRuntimeStatusRequest request) =>
      _client.invoke<GetAgentRuntimeStatusResponse>(ctx, 'AgentService',
          'GetAgentRuntimeStatus', request, GetAgentRuntimeStatusResponse());

  /// ListAgentRuntimeStatuses returns runtime statuses for the given agent
  /// names in a single round trip. When `names` is empty, statuses for all
  /// configured agents are returned.
  $async.Future<ListAgentRuntimeStatusesResponse> listAgentRuntimeStatuses(
          $pb.ClientContext? ctx, ListAgentRuntimeStatusesRequest request) =>
      _client.invoke<ListAgentRuntimeStatusesResponse>(
          ctx,
          'AgentService',
          'ListAgentRuntimeStatuses',
          request,
          ListAgentRuntimeStatusesResponse());

  /// CancelAgentInvocation cancels an in-flight invocation by its ID. The
  /// invocation transitions to FAILED with a cancellation error.
  $async.Future<CancelAgentInvocationResponse> cancelAgentInvocation(
          $pb.ClientContext? ctx, CancelAgentInvocationRequest request) =>
      _client.invoke<CancelAgentInvocationResponse>(ctx, 'AgentService',
          'CancelAgentInvocation', request, CancelAgentInvocationResponse());

  /// StreamAgent runs an agent and streams progress events back to the
  /// caller. The dashboard chat UI uses this Connect server-stream RPC
  /// (replaces the removed POST /api/chat/stream SSE endpoint). One of the
  /// StreamAgentResponse.event oneof variants is sent for each runner event
  /// (started, mid-stream text delta, full ADK event, final response).
  /// Errors bubble up as the RPC's terminal connect.Error so the client
  /// gets typed cancellation / failed-precondition codes instead of an
  /// ad-hoc string payload.
  $async.Future<StreamAgentResponse> streamAgent(
          $pb.ClientContext? ctx, StreamAgentRequest request) =>
      _client.invoke<StreamAgentResponse>(
          ctx, 'AgentService', 'StreamAgent', request, StreamAgentResponse());

  /// AssignAgentID is retired (issue #241): the Agent ID rollout is complete,
  /// every Agent carries an immutable agent_id assigned at creation, and the
  /// repository no longer resolves Agents by name. Always returns
  /// Unimplemented.
  @$core.Deprecated('This method is deprecated')
  $async.Future<AssignAgentIDResponse> assignAgentID(
          $pb.ClientContext? ctx, AssignAgentIDRequest request) =>
      _client.invoke<AssignAgentIDResponse>(ctx, 'AgentService',
          'AssignAgentID', request, AssignAgentIDResponse());

  /// GetMigrationReadiness is retired (issue #241). Use VerifyAgentIDCutover
  /// for post-cutover diagnostics. Always returns Unimplemented.
  @$core.Deprecated('This method is deprecated')
  $async.Future<GetMigrationReadinessResponse> getMigrationReadiness(
          $pb.ClientContext? ctx, GetMigrationReadinessRequest request) =>
      _client.invoke<GetMigrationReadinessResponse>(ctx, 'AgentService',
          'GetMigrationReadiness', request, GetMigrationReadinessResponse());

  /// MigrateAgentsV2 is retired (issue #241): the V2 migration observation
  /// window is closed. Use VerifyAgentIDCutover for post-cutover diagnostics.
  /// Always returns Unimplemented.
  @$core.Deprecated('This method is deprecated')
  $async.Future<MigrateAgentsV2Response> migrateAgentsV2(
          $pb.ClientContext? ctx, MigrateAgentsV2Request request) =>
      _client.invoke<MigrateAgentsV2Response>(ctx, 'AgentService',
          'MigrateAgentsV2', request, MigrateAgentsV2Response());

  /// VerifyAgentIDCutover runs the read-only final cutover verifier (issue
  /// #241) across every workspace: it confirms all Agents carry valid,
  /// workspace-unique Agent IDs, no legacy embedded sub_agents or name-based
  /// workflow references remain, all child/workflow references resolve, no
  /// Agent is stuck in MIGRATION_REQUIRED, consumer records (channels, cron
  /// jobs, automations, forum threads) carry Agent IDs, and no runtime-name
  /// conflict violates the runner's naming constraints. Global admins only.
  $async.Future<VerifyAgentIDCutoverResponse> verifyAgentIDCutover(
          $pb.ClientContext? ctx, VerifyAgentIDCutoverRequest request) =>
      _client.invoke<VerifyAgentIDCutoverResponse>(ctx, 'AgentService',
          'VerifyAgentIDCutover', request, VerifyAgentIDCutoverResponse());

  /// UpdateAgentConfiguration is the composite-save command: it coordinates an
  /// Agent operational patch (DB-owned) with Agent Content changes (Git-owned)
  /// under a single durable operation while keeping ownership separate. The
  /// patch is applied with optimistic concurrency (expected_agent_version); a
  /// mismatch aborts. Returns the operation record for status/retry.
  $async.Future<UpdateAgentConfigurationResponse> updateAgentConfiguration(
          $pb.ClientContext? ctx, UpdateAgentConfigurationRequest request) =>
      _client.invoke<UpdateAgentConfigurationResponse>(
          ctx,
          'AgentService',
          'UpdateAgentConfiguration',
          request,
          UpdateAgentConfigurationResponse());

  /// RestoreAgent reactivates a soft-deleted (tombstoned) Agent from its
  /// retained configuration and Agent Content, flipping it from DELETED back to
  /// ACTIVE and re-publishing the retained content.
  $async.Future<RestoreAgentResponse> restoreAgent(
          $pb.ClientContext? ctx, RestoreAgentRequest request) =>
      _client.invoke<RestoreAgentResponse>(
          ctx, 'AgentService', 'RestoreAgent', request, RestoreAgentResponse());

  /// GetAgentInvocation returns the authoritative state of one invocation by
  /// its ID, scoped by workspace and user ownership.
  $async.Future<GetAgentInvocationResponse> getAgentInvocation(
          $pb.ClientContext? ctx, GetAgentInvocationRequest request) =>
      _client.invoke<GetAgentInvocationResponse>(ctx, 'AgentService',
          'GetAgentInvocation', request, GetAgentInvocationResponse());

  /// GetAgentOperation returns a durable lifecycle operation record by ID.
  $async.Future<GetAgentOperationResponse> getAgentOperation(
          $pb.ClientContext? ctx, GetAgentOperationRequest request) =>
      _client.invoke<GetAgentOperationResponse>(ctx, 'AgentService',
          'GetAgentOperation', request, GetAgentOperationResponse());

  /// ListAgentOperations lists lifecycle operations in the workspace, optionally
  /// filtered by status, with opaque-token pagination.
  $async.Future<ListAgentOperationsResponse> listAgentOperations(
          $pb.ClientContext? ctx, ListAgentOperationsRequest request) =>
      _client.invoke<ListAgentOperationsResponse>(ctx, 'AgentService',
          'ListAgentOperations', request, ListAgentOperationsResponse());

  /// RetryAgentOperation resumes a FAILED lifecycle operation from its first
  /// unfinished step. Steps that already succeeded are skipped (idempotent).
  $async.Future<RetryAgentOperationResponse> retryAgentOperation(
          $pb.ClientContext? ctx, RetryAgentOperationRequest request) =>
      _client.invoke<RetryAgentOperationResponse>(ctx, 'AgentService',
          'RetryAgentOperation', request, RetryAgentOperationResponse());
}

/// MCPServerService manages MCP Server configurations.
class MCPServerServiceApi {
  final $pb.RpcClient _client;

  MCPServerServiceApi(this._client);

  $async.Future<ListMCPServersResponse> listMCPServers(
          $pb.ClientContext? ctx, ListMCPServersRequest request) =>
      _client.invoke<ListMCPServersResponse>(ctx, 'MCPServerService',
          'ListMCPServers', request, ListMCPServersResponse());
  $async.Future<GetMCPServerResponse> getMCPServer(
          $pb.ClientContext? ctx, GetMCPServerRequest request) =>
      _client.invoke<GetMCPServerResponse>(ctx, 'MCPServerService',
          'GetMCPServer', request, GetMCPServerResponse());
  $async.Future<CreateMCPServerResponse> createMCPServer(
          $pb.ClientContext? ctx, CreateMCPServerRequest request) =>
      _client.invoke<CreateMCPServerResponse>(ctx, 'MCPServerService',
          'CreateMCPServer', request, CreateMCPServerResponse());
  $async.Future<UpdateMCPServerResponse> updateMCPServer(
          $pb.ClientContext? ctx, UpdateMCPServerRequest request) =>
      _client.invoke<UpdateMCPServerResponse>(ctx, 'MCPServerService',
          'UpdateMCPServer', request, UpdateMCPServerResponse());
  $async.Future<DeleteMCPServerResponse> deleteMCPServer(
          $pb.ClientContext? ctx, DeleteMCPServerRequest request) =>
      _client.invoke<DeleteMCPServerResponse>(ctx, 'MCPServerService',
          'DeleteMCPServer', request, DeleteMCPServerResponse());

  /// GetMCPServerStatus returns the current runtime status of a configured
  /// MCP server. Implemented in dashboard.proto.
  $async.Future<$8.GetMCPServerStatusResponse> getMCPServerStatus(
          $pb.ClientContext? ctx, $8.GetMCPServerStatusRequest request) =>
      _client.invoke<$8.GetMCPServerStatusResponse>(ctx, 'MCPServerService',
          'GetMCPServerStatus', request, $8.GetMCPServerStatusResponse());

  /// ListMCPTools enumerates tools exposed by the configured MCP servers.
  /// When server_id is empty, all servers are probed.
  $async.Future<$8.ListMCPToolsResponse> listMCPTools(
          $pb.ClientContext? ctx, $8.ListMCPToolsRequest request) =>
      _client.invoke<$8.ListMCPToolsResponse>(ctx, 'MCPServerService',
          'ListMCPTools', request, $8.ListMCPToolsResponse());

  /// StartMCPServerOAuth prepares a workspace-scoped OAuth2 authorization
  /// flow and returns the user-facing authorization URL.
  $async.Future<StartMCPServerOAuthResponse> startMCPServerOAuth(
          $pb.ClientContext? ctx, StartMCPServerOAuthRequest request) =>
      _client.invoke<StartMCPServerOAuthResponse>(ctx, 'MCPServerService',
          'StartMCPServerOAuth', request, StartMCPServerOAuthResponse());

  /// CompleteMCPServerOAuth completes an OAuth2 authorization code callback.
  $async.Future<CompleteMCPServerOAuthResponse> completeMCPServerOAuth(
          $pb.ClientContext? ctx, CompleteMCPServerOAuthRequest request) =>
      _client.invoke<CompleteMCPServerOAuthResponse>(ctx, 'MCPServerService',
          'CompleteMCPServerOAuth', request, CompleteMCPServerOAuthResponse());

  /// GetMCPServerOAuthStatus returns non-secret OAuth connection state.
  $async.Future<GetMCPServerOAuthStatusResponse> getMCPServerOAuthStatus(
          $pb.ClientContext? ctx, GetMCPServerOAuthStatusRequest request) =>
      _client.invoke<GetMCPServerOAuthStatusResponse>(
          ctx,
          'MCPServerService',
          'GetMCPServerOAuthStatus',
          request,
          GetMCPServerOAuthStatusResponse());

  /// DisconnectMCPServerOAuth removes local OAuth2 credentials for a server.
  $async.Future<DisconnectMCPServerOAuthResponse> disconnectMCPServerOAuth(
          $pb.ClientContext? ctx, DisconnectMCPServerOAuthRequest request) =>
      _client.invoke<DisconnectMCPServerOAuthResponse>(
          ctx,
          'MCPServerService',
          'DisconnectMCPServerOAuth',
          request,
          DisconnectMCPServerOAuthResponse());
}

/// GlobalMCPServerService manages workspace-agnostic MCP server presets that
/// admins curate and end users install into their workspace. Browse / install
/// is open to authenticated callers; admin role is required for the
/// create / update / delete mutations.
class GlobalMCPServerServiceApi {
  final $pb.RpcClient _client;

  GlobalMCPServerServiceApi(this._client);

  $async.Future<ListGlobalMCPServersResponse> listGlobalMCPServers(
          $pb.ClientContext? ctx, ListGlobalMCPServersRequest request) =>
      _client.invoke<ListGlobalMCPServersResponse>(
          ctx,
          'GlobalMCPServerService',
          'ListGlobalMCPServers',
          request,
          ListGlobalMCPServersResponse());
  $async.Future<CreateGlobalMCPServerResponse> createGlobalMCPServer(
          $pb.ClientContext? ctx, CreateGlobalMCPServerRequest request) =>
      _client.invoke<CreateGlobalMCPServerResponse>(
          ctx,
          'GlobalMCPServerService',
          'CreateGlobalMCPServer',
          request,
          CreateGlobalMCPServerResponse());
  $async.Future<UpdateGlobalMCPServerResponse> updateGlobalMCPServer(
          $pb.ClientContext? ctx, UpdateGlobalMCPServerRequest request) =>
      _client.invoke<UpdateGlobalMCPServerResponse>(
          ctx,
          'GlobalMCPServerService',
          'UpdateGlobalMCPServer',
          request,
          UpdateGlobalMCPServerResponse());
  $async.Future<DeleteGlobalMCPServerResponse> deleteGlobalMCPServer(
          $pb.ClientContext? ctx, DeleteGlobalMCPServerRequest request) =>
      _client.invoke<DeleteGlobalMCPServerResponse>(
          ctx,
          'GlobalMCPServerService',
          'DeleteGlobalMCPServer',
          request,
          DeleteGlobalMCPServerResponse());

  /// InstallGlobalMCPServer clones a preset into the caller's workspace.
  /// Admins may pass a different target workspace via `workspace_id`; the
  /// server audit-logs cross-workspace installs.
  $async.Future<InstallGlobalMCPServerResponse> installGlobalMCPServer(
          $pb.ClientContext? ctx, InstallGlobalMCPServerRequest request) =>
      _client.invoke<InstallGlobalMCPServerResponse>(
          ctx,
          'GlobalMCPServerService',
          'InstallGlobalMCPServer',
          request,
          InstallGlobalMCPServerResponse());
}

/// ModelProviderService manages LLM provider configurations.
class ModelProviderServiceApi {
  final $pb.RpcClient _client;

  ModelProviderServiceApi(this._client);

  $async.Future<ListModelProvidersResponse> listModelProviders(
          $pb.ClientContext? ctx, ListModelProvidersRequest request) =>
      _client.invoke<ListModelProvidersResponse>(ctx, 'ModelProviderService',
          'ListModelProviders', request, ListModelProvidersResponse());
  $async.Future<GetModelProviderResponse> getModelProvider(
          $pb.ClientContext? ctx, GetModelProviderRequest request) =>
      _client.invoke<GetModelProviderResponse>(ctx, 'ModelProviderService',
          'GetModelProvider', request, GetModelProviderResponse());
  $async.Future<CreateModelProviderResponse> createModelProvider(
          $pb.ClientContext? ctx, CreateModelProviderRequest request) =>
      _client.invoke<CreateModelProviderResponse>(ctx, 'ModelProviderService',
          'CreateModelProvider', request, CreateModelProviderResponse());
  $async.Future<UpdateModelProviderResponse> updateModelProvider(
          $pb.ClientContext? ctx, UpdateModelProviderRequest request) =>
      _client.invoke<UpdateModelProviderResponse>(ctx, 'ModelProviderService',
          'UpdateModelProvider', request, UpdateModelProviderResponse());
  $async.Future<DeleteModelProviderResponse> deleteModelProvider(
          $pb.ClientContext? ctx, DeleteModelProviderRequest request) =>
      _client.invoke<DeleteModelProviderResponse>(ctx, 'ModelProviderService',
          'DeleteModelProvider', request, DeleteModelProviderResponse());
}

/// NotifyGroupService manages outbound notification groups.
class NotifyGroupServiceApi {
  final $pb.RpcClient _client;

  NotifyGroupServiceApi(this._client);

  $async.Future<ListNotifyGroupsResponse> listNotifyGroups(
          $pb.ClientContext? ctx, ListNotifyGroupsRequest request) =>
      _client.invoke<ListNotifyGroupsResponse>(ctx, 'NotifyGroupService',
          'ListNotifyGroups', request, ListNotifyGroupsResponse());
  $async.Future<GetNotifyGroupResponse> getNotifyGroup(
          $pb.ClientContext? ctx, GetNotifyGroupRequest request) =>
      _client.invoke<GetNotifyGroupResponse>(ctx, 'NotifyGroupService',
          'GetNotifyGroup', request, GetNotifyGroupResponse());
  $async.Future<CreateNotifyGroupResponse> createNotifyGroup(
          $pb.ClientContext? ctx, CreateNotifyGroupRequest request) =>
      _client.invoke<CreateNotifyGroupResponse>(ctx, 'NotifyGroupService',
          'CreateNotifyGroup', request, CreateNotifyGroupResponse());
  $async.Future<UpdateNotifyGroupResponse> updateNotifyGroup(
          $pb.ClientContext? ctx, UpdateNotifyGroupRequest request) =>
      _client.invoke<UpdateNotifyGroupResponse>(ctx, 'NotifyGroupService',
          'UpdateNotifyGroup', request, UpdateNotifyGroupResponse());
  $async.Future<DeleteNotifyGroupResponse> deleteNotifyGroup(
          $pb.ClientContext? ctx, DeleteNotifyGroupRequest request) =>
      _client.invoke<DeleteNotifyGroupResponse>(ctx, 'NotifyGroupService',
          'DeleteNotifyGroup', request, DeleteNotifyGroupResponse());
}

/// RemoteAgentService manages Remote Agent configurations.
class RemoteAgentServiceApi {
  final $pb.RpcClient _client;

  RemoteAgentServiceApi(this._client);

  $async.Future<ListRemoteAgentsResponse> listRemoteAgents(
          $pb.ClientContext? ctx, ListRemoteAgentsRequest request) =>
      _client.invoke<ListRemoteAgentsResponse>(ctx, 'RemoteAgentService',
          'ListRemoteAgents', request, ListRemoteAgentsResponse());
  $async.Future<GetRemoteAgentResponse> getRemoteAgent(
          $pb.ClientContext? ctx, GetRemoteAgentRequest request) =>
      _client.invoke<GetRemoteAgentResponse>(ctx, 'RemoteAgentService',
          'GetRemoteAgent', request, GetRemoteAgentResponse());
  $async.Future<CreateRemoteAgentResponse> createRemoteAgent(
          $pb.ClientContext? ctx, CreateRemoteAgentRequest request) =>
      _client.invoke<CreateRemoteAgentResponse>(ctx, 'RemoteAgentService',
          'CreateRemoteAgent', request, CreateRemoteAgentResponse());
  $async.Future<UpdateRemoteAgentResponse> updateRemoteAgent(
          $pb.ClientContext? ctx, UpdateRemoteAgentRequest request) =>
      _client.invoke<UpdateRemoteAgentResponse>(ctx, 'RemoteAgentService',
          'UpdateRemoteAgent', request, UpdateRemoteAgentResponse());
  $async.Future<DeleteRemoteAgentResponse> deleteRemoteAgent(
          $pb.ClientContext? ctx, DeleteRemoteAgentRequest request) =>
      _client.invoke<DeleteRemoteAgentResponse>(ctx, 'RemoteAgentService',
          'DeleteRemoteAgent', request, DeleteRemoteAgentResponse());

  /// GetRemoteAgentStatus probes the configured remote agent endpoint and
  /// returns its liveness state (A2A: HTTP agent.json; daemon: registry).
  $async.Future<$8.GetRemoteAgentStatusResponse> getRemoteAgentStatus(
          $pb.ClientContext? ctx, $8.GetRemoteAgentStatusRequest request) =>
      _client.invoke<$8.GetRemoteAgentStatusResponse>(ctx, 'RemoteAgentService',
          'GetRemoteAgentStatus', request, $8.GetRemoteAgentStatusResponse());
}

/// ChannelService manages AgentChannel configurations.
class ChannelServiceApi {
  final $pb.RpcClient _client;

  ChannelServiceApi(this._client);

  $async.Future<ListChannelsResponse> listChannels(
          $pb.ClientContext? ctx, ListChannelsRequest request) =>
      _client.invoke<ListChannelsResponse>(ctx, 'ChannelService',
          'ListChannels', request, ListChannelsResponse());
  $async.Future<GetChannelResponse> getChannel(
          $pb.ClientContext? ctx, GetChannelRequest request) =>
      _client.invoke<GetChannelResponse>(
          ctx, 'ChannelService', 'GetChannel', request, GetChannelResponse());
  $async.Future<CreateChannelResponse> createChannel(
          $pb.ClientContext? ctx, CreateChannelRequest request) =>
      _client.invoke<CreateChannelResponse>(ctx, 'ChannelService',
          'CreateChannel', request, CreateChannelResponse());
  $async.Future<UpdateChannelResponse> updateChannel(
          $pb.ClientContext? ctx, UpdateChannelRequest request) =>
      _client.invoke<UpdateChannelResponse>(ctx, 'ChannelService',
          'UpdateChannel', request, UpdateChannelResponse());
  $async.Future<DeleteChannelResponse> deleteChannel(
          $pb.ClientContext? ctx, DeleteChannelRequest request) =>
      _client.invoke<DeleteChannelResponse>(ctx, 'ChannelService',
          'DeleteChannel', request, DeleteChannelResponse());

  /// GetChannelStatus returns the current runtime status of a channel poller.
  /// Implemented in dashboard.proto.
  $async.Future<$8.GetChannelStatusResponse> getChannelStatus(
          $pb.ClientContext? ctx, $8.GetChannelStatusRequest request) =>
      _client.invoke<$8.GetChannelStatusResponse>(ctx, 'ChannelService',
          'GetChannelStatus', request, $8.GetChannelStatusResponse());

  /// RestartChannel reloads the channel manager. Effectively bounces all pollers.
  $async.Future<RestartChannelResponse> restartChannel(
          $pb.ClientContext? ctx, RestartChannelRequest request) =>
      _client.invoke<RestartChannelResponse>(ctx, 'ChannelService',
          'RestartChannel', request, RestartChannelResponse());

  /// PauseChannel disables the named channel and reloads the channel manager so
  /// its poller stops. The configuration `enabled` flag is set to false.
  $async.Future<PauseChannelResponse> pauseChannel(
          $pb.ClientContext? ctx, PauseChannelRequest request) =>
      _client.invoke<PauseChannelResponse>(ctx, 'ChannelService',
          'PauseChannel', request, PauseChannelResponse());

  /// ResumeChannel re-enables the named channel and reloads the channel manager.
  $async.Future<ResumeChannelResponse> resumeChannel(
          $pb.ClientContext? ctx, ResumeChannelRequest request) =>
      _client.invoke<ResumeChannelResponse>(ctx, 'ChannelService',
          'ResumeChannel', request, ResumeChannelResponse());
}

class SessionServiceApi {
  final $pb.RpcClient _client;

  SessionServiceApi(this._client);

  /// CreateSession creates a new session for the given agent.
  $async.Future<CreateSessionResponse> createSession(
          $pb.ClientContext? ctx, CreateSessionRequest request) =>
      _client.invoke<CreateSessionResponse>(ctx, 'SessionService',
          'CreateSession', request, CreateSessionResponse());

  /// GetSession retrieves a session by ID, optionally with recent events.
  $async.Future<GetSessionResponse> getSession(
          $pb.ClientContext? ctx, GetSessionRequest request) =>
      _client.invoke<GetSessionResponse>(
          ctx, 'SessionService', 'GetSession', request, GetSessionResponse());

  /// ListSessions lists sessions for a given app and user.
  $async.Future<ListSessionsResponse> listSessions(
          $pb.ClientContext? ctx, ListSessionsRequest request) =>
      _client.invoke<ListSessionsResponse>(ctx, 'SessionService',
          'ListSessions', request, ListSessionsResponse());

  /// DeleteSession deletes a session and its events.
  $async.Future<DeleteSessionResponse> deleteSession(
          $pb.ClientContext? ctx, DeleteSessionRequest request) =>
      _client.invoke<DeleteSessionResponse>(ctx, 'SessionService',
          'DeleteSession', request, DeleteSessionResponse());

  /// ReplySession sends a user message to an existing session and returns the agent response.
  $async.Future<ReplySessionResponse> replySession(
          $pb.ClientContext? ctx, ReplySessionRequest request) =>
      _client.invoke<ReplySessionResponse>(ctx, 'SessionService',
          'ReplySession', request, ReplySessionResponse());

  /// UpdateSessionTitle sets a first-class title on a session.
  $async.Future<UpdateSessionTitleResponse> updateSessionTitle(
          $pb.ClientContext? ctx, UpdateSessionTitleRequest request) =>
      _client.invoke<UpdateSessionTitleResponse>(ctx, 'SessionService',
          'UpdateSessionTitle', request, UpdateSessionTitleResponse());

  /// GenerateSessionTitle derives a best-effort title from the first
  /// conversation turn and persists it only while no effective title exists.
  /// Manual, legacy, and concurrent titles always win. Returns the effective
  /// SessionInfo and whether this call actually generated a new title.
  $async.Future<GenerateSessionTitleResponse> generateSessionTitle(
          $pb.ClientContext? ctx, GenerateSessionTitleRequest request) =>
      _client.invoke<GenerateSessionTitleResponse>(ctx, 'SessionService',
          'GenerateSessionTitle', request, GenerateSessionTitleResponse());

  /// MarkSessionRead records that the authenticated user has viewed the
  /// session up to (and including) the given event. The server persists the
  /// read cursor; subsequent ListSessions / GetSession responses reflect it
  /// through SessionInfo.unread. Only the session owner may mark read;
  /// another user's attempt returns NotFound.
  $async.Future<MarkSessionReadResponse> markSessionRead(
          $pb.ClientContext? ctx, MarkSessionReadRequest request) =>
      _client.invoke<MarkSessionReadResponse>(ctx, 'SessionService',
          'MarkSessionRead', request, MarkSessionReadResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
