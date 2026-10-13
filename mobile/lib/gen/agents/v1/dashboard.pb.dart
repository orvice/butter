// This is a generated file - do not edit.
//
// Generated from agents/v1/dashboard.proto.

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

import 'agent.pbenum.dart' as $3;
import 'agentchannel.pbenum.dart' as $4;
import 'api_token.pb.dart' as $2;
import 'dashboard.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'dashboard.pbenum.dart';

class GetActivityMetricsRequest extends $pb.GeneratedMessage {
  factory GetActivityMetricsRequest({
    GetActivityMetricsRequest_Range? range,
  }) {
    final result = GetActivityMetricsRequest._();
    if (range != null) result.range = range;
    return result;
  }

  GetActivityMetricsRequest._();

  factory GetActivityMetricsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityMetricsRequest()..mergeFromBuffer(data, registry);
  factory GetActivityMetricsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityMetricsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetActivityMetricsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetActivityMetricsRequest.$_createMessage)
    ..aE<GetActivityMetricsRequest_Range>(1, _omitFieldNames ? '' : 'range',
        enumValues: GetActivityMetricsRequest_Range.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityMetricsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityMetricsRequest copyWith(
          void Function(GetActivityMetricsRequest) updates) =>
      super.copyWith((message) => updates(message as GetActivityMetricsRequest))
          as GetActivityMetricsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetActivityMetricsRequest() / GetActivityMetricsRequest.new instead')
  static GetActivityMetricsRequest create() => GetActivityMetricsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetActivityMetricsRequest._();
  @$core.override
  GetActivityMetricsRequest createEmptyInstance() =>
      GetActivityMetricsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetActivityMetricsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetActivityMetricsRequest>(
          GetActivityMetricsRequest.$_createMessage);
  static GetActivityMetricsRequest? _defaultInstance;

  /// Rolling window the counts cover. Unspecified is treated as RANGE_7D.
  @$pb.TagNumber(1)
  GetActivityMetricsRequest_Range get range => $_getN(0);
  @$pb.TagNumber(1)
  set range(GetActivityMetricsRequest_Range value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRange() => $_has(0);
  @$pb.TagNumber(1)
  void clearRange() => $_clearField(1);
}

class GetActivityMetricsResponse extends $pb.GeneratedMessage {
  factory GetActivityMetricsResponse({
    $core.int? agentRuns,
    $core.int? agentRunsFailed,
    $core.int? automationRuns,
    $core.int? automationRunsFailed,
    $0.Timestamp? windowStart,
    $0.Timestamp? windowEnd,
  }) {
    final result = GetActivityMetricsResponse._();
    if (agentRuns != null) result.agentRuns = agentRuns;
    if (agentRunsFailed != null) result.agentRunsFailed = agentRunsFailed;
    if (automationRuns != null) result.automationRuns = automationRuns;
    if (automationRunsFailed != null)
      result.automationRunsFailed = automationRunsFailed;
    if (windowStart != null) result.windowStart = windowStart;
    if (windowEnd != null) result.windowEnd = windowEnd;
    return result;
  }

  GetActivityMetricsResponse._();

  factory GetActivityMetricsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityMetricsResponse()..mergeFromBuffer(data, registry);
  factory GetActivityMetricsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityMetricsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetActivityMetricsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetActivityMetricsResponse.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'agentRuns')
    ..aI(2, _omitFieldNames ? '' : 'agentRunsFailed')
    ..aI(3, _omitFieldNames ? '' : 'automationRuns')
    ..aI(4, _omitFieldNames ? '' : 'automationRunsFailed')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'windowStart',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'windowEnd',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityMetricsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityMetricsResponse copyWith(
          void Function(GetActivityMetricsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetActivityMetricsResponse))
          as GetActivityMetricsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetActivityMetricsResponse() / GetActivityMetricsResponse.new instead')
  static GetActivityMetricsResponse create() => GetActivityMetricsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetActivityMetricsResponse._();
  @$core.override
  GetActivityMetricsResponse createEmptyInstance() =>
      GetActivityMetricsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetActivityMetricsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetActivityMetricsResponse>(
          GetActivityMetricsResponse.$_createMessage);
  static GetActivityMetricsResponse? _defaultInstance;

  /// Agent invocations that started within the window (across all workspaces).
  @$pb.TagNumber(1)
  $core.int get agentRuns => $_getIZ(0);
  @$pb.TagNumber(1)
  set agentRuns($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentRuns() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentRuns() => $_clearField(1);

  /// Subset of agent_runs that ended in failure.
  @$pb.TagNumber(2)
  $core.int get agentRunsFailed => $_getIZ(1);
  @$pb.TagNumber(2)
  set agentRunsFailed($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentRunsFailed() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentRunsFailed() => $_clearField(2);

  /// Cron / automation executions that started within the window.
  @$pb.TagNumber(3)
  $core.int get automationRuns => $_getIZ(2);
  @$pb.TagNumber(3)
  set automationRuns($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAutomationRuns() => $_has(2);
  @$pb.TagNumber(3)
  void clearAutomationRuns() => $_clearField(3);

  /// Subset of automation_runs that ended in error.
  @$pb.TagNumber(4)
  $core.int get automationRunsFailed => $_getIZ(3);
  @$pb.TagNumber(4)
  set automationRunsFailed($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAutomationRunsFailed() => $_has(3);
  @$pb.TagNumber(4)
  void clearAutomationRunsFailed() => $_clearField(4);

  /// Half-open window [window_start, window_end) the counts cover.
  @$pb.TagNumber(5)
  $0.Timestamp get windowStart => $_getN(4);
  @$pb.TagNumber(5)
  set windowStart($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasWindowStart() => $_has(4);
  @$pb.TagNumber(5)
  void clearWindowStart() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureWindowStart() => $_ensure(4);

  @$pb.TagNumber(6)
  $0.Timestamp get windowEnd => $_getN(5);
  @$pb.TagNumber(6)
  set windowEnd($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasWindowEnd() => $_has(5);
  @$pb.TagNumber(6)
  void clearWindowEnd() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureWindowEnd() => $_ensure(5);
}

class GetCronExecutionTimeseriesRequest extends $pb.GeneratedMessage {
  factory GetCronExecutionTimeseriesRequest({
    GetCronExecutionTimeseriesRequest_Range? range,
    $core.String? jobName,
  }) {
    final result = GetCronExecutionTimeseriesRequest._();
    if (range != null) result.range = range;
    if (jobName != null) result.jobName = jobName;
    return result;
  }

  GetCronExecutionTimeseriesRequest._();

  factory GetCronExecutionTimeseriesRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronExecutionTimeseriesRequest()..mergeFromBuffer(data, registry);
  factory GetCronExecutionTimeseriesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronExecutionTimeseriesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCronExecutionTimeseriesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetCronExecutionTimeseriesRequest.$_createMessage)
    ..aE<GetCronExecutionTimeseriesRequest_Range>(
        1, _omitFieldNames ? '' : 'range',
        enumValues: GetCronExecutionTimeseriesRequest_Range.values)
    ..aOS(2, _omitFieldNames ? '' : 'jobName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronExecutionTimeseriesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronExecutionTimeseriesRequest copyWith(
          void Function(GetCronExecutionTimeseriesRequest) updates) =>
      super.copyWith((message) =>
              updates(message as GetCronExecutionTimeseriesRequest))
          as GetCronExecutionTimeseriesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCronExecutionTimeseriesRequest() / GetCronExecutionTimeseriesRequest.new instead')
  static GetCronExecutionTimeseriesRequest create() =>
      GetCronExecutionTimeseriesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetCronExecutionTimeseriesRequest._();
  @$core.override
  GetCronExecutionTimeseriesRequest createEmptyInstance() =>
      GetCronExecutionTimeseriesRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCronExecutionTimeseriesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCronExecutionTimeseriesRequest>(
          GetCronExecutionTimeseriesRequest.$_createMessage);
  static GetCronExecutionTimeseriesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  GetCronExecutionTimeseriesRequest_Range get range => $_getN(0);
  @$pb.TagNumber(1)
  set range(GetCronExecutionTimeseriesRequest_Range value) =>
      $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRange() => $_has(0);
  @$pb.TagNumber(1)
  void clearRange() => $_clearField(1);

  /// Optional: filter by job name.
  @$pb.TagNumber(2)
  $core.String get jobName => $_getSZ(1);
  @$pb.TagNumber(2)
  set jobName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasJobName() => $_has(1);
  @$pb.TagNumber(2)
  void clearJobName() => $_clearField(2);
}

class GetCronExecutionTimeseriesResponse extends $pb.GeneratedMessage {
  factory GetCronExecutionTimeseriesResponse({
    $core.Iterable<CronExecutionBucket>? buckets,
  }) {
    final result = GetCronExecutionTimeseriesResponse._();
    if (buckets != null) result.buckets.addAll(buckets);
    return result;
  }

  GetCronExecutionTimeseriesResponse._();

  factory GetCronExecutionTimeseriesResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronExecutionTimeseriesResponse()..mergeFromBuffer(data, registry);
  factory GetCronExecutionTimeseriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronExecutionTimeseriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCronExecutionTimeseriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetCronExecutionTimeseriesResponse.$_createMessage)
    ..pPM<CronExecutionBucket>(1, _omitFieldNames ? '' : 'buckets',
        subBuilder: CronExecutionBucket.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronExecutionTimeseriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronExecutionTimeseriesResponse copyWith(
          void Function(GetCronExecutionTimeseriesResponse) updates) =>
      super.copyWith((message) =>
              updates(message as GetCronExecutionTimeseriesResponse))
          as GetCronExecutionTimeseriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetCronExecutionTimeseriesResponse() / GetCronExecutionTimeseriesResponse.new instead')
  static GetCronExecutionTimeseriesResponse create() =>
      GetCronExecutionTimeseriesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetCronExecutionTimeseriesResponse._();
  @$core.override
  GetCronExecutionTimeseriesResponse createEmptyInstance() =>
      GetCronExecutionTimeseriesResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCronExecutionTimeseriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCronExecutionTimeseriesResponse>(
          GetCronExecutionTimeseriesResponse.$_createMessage);
  static GetCronExecutionTimeseriesResponse? _defaultInstance;

  /// Oldest bucket first.
  @$pb.TagNumber(1)
  $pb.PbList<CronExecutionBucket> get buckets => $_getList(0);
}

class CronExecutionBucket extends $pb.GeneratedMessage {
  factory CronExecutionBucket({
    $0.Timestamp? start,
    $core.int? success,
    $core.int? error,
  }) {
    final result = CronExecutionBucket._();
    if (start != null) result.start = start;
    if (success != null) result.success = success;
    if (error != null) result.error = error;
    return result;
  }

  CronExecutionBucket._();

  factory CronExecutionBucket.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronExecutionBucket()..mergeFromBuffer(data, registry);
  factory CronExecutionBucket.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronExecutionBucket()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CronExecutionBucket',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CronExecutionBucket.$_createMessage)
    ..aOM<$0.Timestamp>(1, _omitFieldNames ? '' : 'start',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'success')
    ..aI(3, _omitFieldNames ? '' : 'error')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronExecutionBucket clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronExecutionBucket copyWith(void Function(CronExecutionBucket) updates) =>
      super.copyWith((message) => updates(message as CronExecutionBucket))
          as CronExecutionBucket;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CronExecutionBucket() / CronExecutionBucket.new instead')
  static CronExecutionBucket create() => CronExecutionBucket._();
  static $pb.GeneratedMessage $_createMessage() => CronExecutionBucket._();
  @$core.override
  CronExecutionBucket createEmptyInstance() => CronExecutionBucket._();
  @$core.pragma('dart2js:noInline')
  static CronExecutionBucket getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CronExecutionBucket>(
          CronExecutionBucket.$_createMessage);
  static CronExecutionBucket? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Timestamp get start => $_getN(0);
  @$pb.TagNumber(1)
  set start($0.Timestamp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStart() => $_has(0);
  @$pb.TagNumber(1)
  void clearStart() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Timestamp ensureStart() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.int get success => $_getIZ(1);
  @$pb.TagNumber(2)
  set success($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSuccess() => $_has(1);
  @$pb.TagNumber(2)
  void clearSuccess() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get error => $_getIZ(2);
  @$pb.TagNumber(3)
  set error($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasError() => $_has(2);
  @$pb.TagNumber(3)
  void clearError() => $_clearField(3);
}

class GetActivityFeedRequest extends $pb.GeneratedMessage {
  factory GetActivityFeedRequest({
    $core.int? limit,
    $core.String? pageToken,
  }) {
    final result = GetActivityFeedRequest._();
    if (limit != null) result.limit = limit;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  GetActivityFeedRequest._();

  factory GetActivityFeedRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityFeedRequest()..mergeFromBuffer(data, registry);
  factory GetActivityFeedRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityFeedRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetActivityFeedRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetActivityFeedRequest.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'limit')
    ..aOS(2, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityFeedRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityFeedRequest copyWith(
          void Function(GetActivityFeedRequest) updates) =>
      super.copyWith((message) => updates(message as GetActivityFeedRequest))
          as GetActivityFeedRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetActivityFeedRequest() / GetActivityFeedRequest.new instead')
  static GetActivityFeedRequest create() => GetActivityFeedRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetActivityFeedRequest._();
  @$core.override
  GetActivityFeedRequest createEmptyInstance() => GetActivityFeedRequest._();
  @$core.pragma('dart2js:noInline')
  static GetActivityFeedRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetActivityFeedRequest>(
          GetActivityFeedRequest.$_createMessage);
  static GetActivityFeedRequest? _defaultInstance;

  /// Maximum records to return. Defaults to 20, capped at 200.
  @$pb.TagNumber(1)
  $core.int get limit => $_getIZ(0);
  @$pb.TagNumber(1)
  set limit($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLimit() => $_has(0);
  @$pb.TagNumber(1)
  void clearLimit() => $_clearField(1);

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

class GetActivityFeedResponse extends $pb.GeneratedMessage {
  factory GetActivityFeedResponse({
    $core.Iterable<ActivityEvent>? events,
    $core.String? nextPageToken,
  }) {
    final result = GetActivityFeedResponse._();
    if (events != null) result.events.addAll(events);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  GetActivityFeedResponse._();

  factory GetActivityFeedResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityFeedResponse()..mergeFromBuffer(data, registry);
  factory GetActivityFeedResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetActivityFeedResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetActivityFeedResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetActivityFeedResponse.$_createMessage)
    ..pPM<ActivityEvent>(1, _omitFieldNames ? '' : 'events',
        subBuilder: ActivityEvent.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityFeedResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetActivityFeedResponse copyWith(
          void Function(GetActivityFeedResponse) updates) =>
      super.copyWith((message) => updates(message as GetActivityFeedResponse))
          as GetActivityFeedResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetActivityFeedResponse() / GetActivityFeedResponse.new instead')
  static GetActivityFeedResponse create() => GetActivityFeedResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetActivityFeedResponse._();
  @$core.override
  GetActivityFeedResponse createEmptyInstance() => GetActivityFeedResponse._();
  @$core.pragma('dart2js:noInline')
  static GetActivityFeedResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetActivityFeedResponse>(
          GetActivityFeedResponse.$_createMessage);
  static GetActivityFeedResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ActivityEvent> get events => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);
}

class ActivityEvent extends $pb.GeneratedMessage {
  factory ActivityEvent({
    $core.String? id,
    $core.String? kind,
    $core.String? actor,
    $core.String? message,
    $0.Timestamp? timestamp,
  }) {
    final result = ActivityEvent._();
    if (id != null) result.id = id;
    if (kind != null) result.kind = kind;
    if (actor != null) result.actor = actor;
    if (message != null) result.message = message;
    if (timestamp != null) result.timestamp = timestamp;
    return result;
  }

  ActivityEvent._();

  factory ActivityEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ActivityEvent()..mergeFromBuffer(data, registry);
  factory ActivityEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ActivityEvent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ActivityEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ActivityEvent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'kind')
    ..aOS(3, _omitFieldNames ? '' : 'actor')
    ..aOS(4, _omitFieldNames ? '' : 'message')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'timestamp',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ActivityEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ActivityEvent copyWith(void Function(ActivityEvent) updates) =>
      super.copyWith((message) => updates(message as ActivityEvent))
          as ActivityEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ActivityEvent() / ActivityEvent.new instead')
  static ActivityEvent create() => ActivityEvent._();
  static $pb.GeneratedMessage $_createMessage() => ActivityEvent._();
  @$core.override
  ActivityEvent createEmptyInstance() => ActivityEvent._();
  @$core.pragma('dart2js:noInline')
  static ActivityEvent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ActivityEvent>(
          ActivityEvent.$_createMessage);
  static ActivityEvent? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Kind of event: invocation | execution_completed | error.
  @$pb.TagNumber(2)
  $core.String get kind => $_getSZ(1);
  @$pb.TagNumber(2)
  set kind($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  /// Agent / cron job / webhook id that produced the event.
  @$pb.TagNumber(3)
  $core.String get actor => $_getSZ(2);
  @$pb.TagNumber(3)
  set actor($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasActor() => $_has(2);
  @$pb.TagNumber(3)
  void clearActor() => $_clearField(3);

  /// Short human-readable message.
  @$pb.TagNumber(4)
  $core.String get message => $_getSZ(3);
  @$pb.TagNumber(4)
  set message($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMessage() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessage() => $_clearField(4);

  @$pb.TagNumber(5)
  $0.Timestamp get timestamp => $_getN(4);
  @$pb.TagNumber(5)
  set timestamp($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasTimestamp() => $_has(4);
  @$pb.TagNumber(5)
  void clearTimestamp() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureTimestamp() => $_ensure(4);
}

class GetOverviewRequest extends $pb.GeneratedMessage {
  factory GetOverviewRequest({
    $core.String? environment,
  }) {
    final result = GetOverviewRequest._();
    if (environment != null) result.environment = environment;
    return result;
  }

  GetOverviewRequest._();

  factory GetOverviewRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOverviewRequest()..mergeFromBuffer(data, registry);
  factory GetOverviewRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOverviewRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetOverviewRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetOverviewRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'environment')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOverviewRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOverviewRequest copyWith(void Function(GetOverviewRequest) updates) =>
      super.copyWith((message) => updates(message as GetOverviewRequest))
          as GetOverviewRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetOverviewRequest() / GetOverviewRequest.new instead')
  static GetOverviewRequest create() => GetOverviewRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetOverviewRequest._();
  @$core.override
  GetOverviewRequest createEmptyInstance() => GetOverviewRequest._();
  @$core.pragma('dart2js:noInline')
  static GetOverviewRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetOverviewRequest>(
          GetOverviewRequest.$_createMessage);
  static GetOverviewRequest? _defaultInstance;

  /// Optional environment label echoed back to the client. The server does not
  /// currently filter on this field.
  @$pb.TagNumber(1)
  $core.String get environment => $_getSZ(0);
  @$pb.TagNumber(1)
  set environment($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnvironment() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnvironment() => $_clearField(1);
}

class GetOverviewResponse extends $pb.GeneratedMessage {
  factory GetOverviewResponse({
    OverviewCounts? counts,
    HealthSummary? health,
    DaemonHandshake? latestDaemonHandshake,
  }) {
    final result = GetOverviewResponse._();
    if (counts != null) result.counts = counts;
    if (health != null) result.health = health;
    if (latestDaemonHandshake != null)
      result.latestDaemonHandshake = latestDaemonHandshake;
    return result;
  }

  GetOverviewResponse._();

  factory GetOverviewResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOverviewResponse()..mergeFromBuffer(data, registry);
  factory GetOverviewResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetOverviewResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetOverviewResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetOverviewResponse.$_createMessage)
    ..aOM<OverviewCounts>(1, _omitFieldNames ? '' : 'counts',
        subBuilder: OverviewCounts.$_createMessage)
    ..aOM<HealthSummary>(2, _omitFieldNames ? '' : 'health',
        subBuilder: HealthSummary.$_createMessage)
    ..aOM<DaemonHandshake>(3, _omitFieldNames ? '' : 'latestDaemonHandshake',
        subBuilder: DaemonHandshake.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOverviewResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetOverviewResponse copyWith(void Function(GetOverviewResponse) updates) =>
      super.copyWith((message) => updates(message as GetOverviewResponse))
          as GetOverviewResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetOverviewResponse() / GetOverviewResponse.new instead')
  static GetOverviewResponse create() => GetOverviewResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetOverviewResponse._();
  @$core.override
  GetOverviewResponse createEmptyInstance() => GetOverviewResponse._();
  @$core.pragma('dart2js:noInline')
  static GetOverviewResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetOverviewResponse>(
          GetOverviewResponse.$_createMessage);
  static GetOverviewResponse? _defaultInstance;

  @$pb.TagNumber(1)
  OverviewCounts get counts => $_getN(0);
  @$pb.TagNumber(1)
  set counts(OverviewCounts value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCounts() => $_has(0);
  @$pb.TagNumber(1)
  void clearCounts() => $_clearField(1);
  @$pb.TagNumber(1)
  OverviewCounts ensureCounts() => $_ensure(0);

  @$pb.TagNumber(2)
  HealthSummary get health => $_getN(1);
  @$pb.TagNumber(2)
  set health(HealthSummary value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasHealth() => $_has(1);
  @$pb.TagNumber(2)
  void clearHealth() => $_clearField(2);
  @$pb.TagNumber(2)
  HealthSummary ensureHealth() => $_ensure(1);

  @$pb.TagNumber(3)
  DaemonHandshake get latestDaemonHandshake => $_getN(2);
  @$pb.TagNumber(3)
  set latestDaemonHandshake(DaemonHandshake value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasLatestDaemonHandshake() => $_has(2);
  @$pb.TagNumber(3)
  void clearLatestDaemonHandshake() => $_clearField(3);
  @$pb.TagNumber(3)
  DaemonHandshake ensureLatestDaemonHandshake() => $_ensure(2);
}

class OverviewCounts extends $pb.GeneratedMessage {
  factory OverviewCounts({
    $core.int? activeAgents,
    $core.int? mcpServers,
    $core.int? connectedDaemons,
    $core.int? remoteAgents,
    $core.int? channels,
    $core.int? cronJobs,
    $core.int? activeSessions,
  }) {
    final result = OverviewCounts._();
    if (activeAgents != null) result.activeAgents = activeAgents;
    if (mcpServers != null) result.mcpServers = mcpServers;
    if (connectedDaemons != null) result.connectedDaemons = connectedDaemons;
    if (remoteAgents != null) result.remoteAgents = remoteAgents;
    if (channels != null) result.channels = channels;
    if (cronJobs != null) result.cronJobs = cronJobs;
    if (activeSessions != null) result.activeSessions = activeSessions;
    return result;
  }

  OverviewCounts._();

  factory OverviewCounts.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OverviewCounts()..mergeFromBuffer(data, registry);
  factory OverviewCounts.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OverviewCounts()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OverviewCounts',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: OverviewCounts.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'activeAgents')
    ..aI(2, _omitFieldNames ? '' : 'mcpServers')
    ..aI(3, _omitFieldNames ? '' : 'connectedDaemons')
    ..aI(4, _omitFieldNames ? '' : 'remoteAgents')
    ..aI(5, _omitFieldNames ? '' : 'channels')
    ..aI(6, _omitFieldNames ? '' : 'cronJobs')
    ..aI(7, _omitFieldNames ? '' : 'activeSessions')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OverviewCounts clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OverviewCounts copyWith(void Function(OverviewCounts) updates) =>
      super.copyWith((message) => updates(message as OverviewCounts))
          as OverviewCounts;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use OverviewCounts() / OverviewCounts.new instead')
  static OverviewCounts create() => OverviewCounts._();
  static $pb.GeneratedMessage $_createMessage() => OverviewCounts._();
  @$core.override
  OverviewCounts createEmptyInstance() => OverviewCounts._();
  @$core.pragma('dart2js:noInline')
  static OverviewCounts getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<OverviewCounts>(
          OverviewCounts.$_createMessage);
  static OverviewCounts? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get activeAgents => $_getIZ(0);
  @$pb.TagNumber(1)
  set activeAgents($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasActiveAgents() => $_has(0);
  @$pb.TagNumber(1)
  void clearActiveAgents() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get mcpServers => $_getIZ(1);
  @$pb.TagNumber(2)
  set mcpServers($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMcpServers() => $_has(1);
  @$pb.TagNumber(2)
  void clearMcpServers() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get connectedDaemons => $_getIZ(2);
  @$pb.TagNumber(3)
  set connectedDaemons($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasConnectedDaemons() => $_has(2);
  @$pb.TagNumber(3)
  void clearConnectedDaemons() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get remoteAgents => $_getIZ(3);
  @$pb.TagNumber(4)
  set remoteAgents($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRemoteAgents() => $_has(3);
  @$pb.TagNumber(4)
  void clearRemoteAgents() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get channels => $_getIZ(4);
  @$pb.TagNumber(5)
  set channels($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasChannels() => $_has(4);
  @$pb.TagNumber(5)
  void clearChannels() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get cronJobs => $_getIZ(5);
  @$pb.TagNumber(6)
  set cronJobs($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCronJobs() => $_has(5);
  @$pb.TagNumber(6)
  void clearCronJobs() => $_clearField(6);

  /// Total ADK sessions persisted in MongoDB. 0 if the session backend does
  /// not implement counting.
  @$pb.TagNumber(7)
  $core.int get activeSessions => $_getIZ(6);
  @$pb.TagNumber(7)
  set activeSessions($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasActiveSessions() => $_has(6);
  @$pb.TagNumber(7)
  void clearActiveSessions() => $_clearField(7);
}

class HealthSummary extends $pb.GeneratedMessage {
  factory HealthSummary({
    ComponentHealth? mongodb,
    ComponentHealth? redis,
    ComponentHealth? runner,
  }) {
    final result = HealthSummary._();
    if (mongodb != null) result.mongodb = mongodb;
    if (redis != null) result.redis = redis;
    if (runner != null) result.runner = runner;
    return result;
  }

  HealthSummary._();

  factory HealthSummary.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HealthSummary()..mergeFromBuffer(data, registry);
  factory HealthSummary.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HealthSummary()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HealthSummary',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: HealthSummary.$_createMessage)
    ..aOM<ComponentHealth>(1, _omitFieldNames ? '' : 'mongodb',
        subBuilder: ComponentHealth.$_createMessage)
    ..aOM<ComponentHealth>(2, _omitFieldNames ? '' : 'redis',
        subBuilder: ComponentHealth.$_createMessage)
    ..aOM<ComponentHealth>(3, _omitFieldNames ? '' : 'runner',
        subBuilder: ComponentHealth.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HealthSummary clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HealthSummary copyWith(void Function(HealthSummary) updates) =>
      super.copyWith((message) => updates(message as HealthSummary))
          as HealthSummary;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HealthSummary() / HealthSummary.new instead')
  static HealthSummary create() => HealthSummary._();
  static $pb.GeneratedMessage $_createMessage() => HealthSummary._();
  @$core.override
  HealthSummary createEmptyInstance() => HealthSummary._();
  @$core.pragma('dart2js:noInline')
  static HealthSummary getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HealthSummary>(
          HealthSummary.$_createMessage);
  static HealthSummary? _defaultInstance;

  @$pb.TagNumber(1)
  ComponentHealth get mongodb => $_getN(0);
  @$pb.TagNumber(1)
  set mongodb(ComponentHealth value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMongodb() => $_has(0);
  @$pb.TagNumber(1)
  void clearMongodb() => $_clearField(1);
  @$pb.TagNumber(1)
  ComponentHealth ensureMongodb() => $_ensure(0);

  @$pb.TagNumber(2)
  ComponentHealth get redis => $_getN(1);
  @$pb.TagNumber(2)
  set redis(ComponentHealth value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRedis() => $_has(1);
  @$pb.TagNumber(2)
  void clearRedis() => $_clearField(2);
  @$pb.TagNumber(2)
  ComponentHealth ensureRedis() => $_ensure(1);

  @$pb.TagNumber(3)
  ComponentHealth get runner => $_getN(2);
  @$pb.TagNumber(3)
  set runner(ComponentHealth value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasRunner() => $_has(2);
  @$pb.TagNumber(3)
  void clearRunner() => $_clearField(3);
  @$pb.TagNumber(3)
  ComponentHealth ensureRunner() => $_ensure(2);
}

class ComponentHealth extends $pb.GeneratedMessage {
  factory ComponentHealth({
    ComponentHealth_Status? status,
    $core.String? detail,
    $0.Timestamp? checkedAt,
    $fixnum.Int64? latencyMs,
  }) {
    final result = ComponentHealth._();
    if (status != null) result.status = status;
    if (detail != null) result.detail = detail;
    if (checkedAt != null) result.checkedAt = checkedAt;
    if (latencyMs != null) result.latencyMs = latencyMs;
    return result;
  }

  ComponentHealth._();

  factory ComponentHealth.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ComponentHealth()..mergeFromBuffer(data, registry);
  factory ComponentHealth.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ComponentHealth()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ComponentHealth',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ComponentHealth.$_createMessage)
    ..aE<ComponentHealth_Status>(1, _omitFieldNames ? '' : 'status',
        enumValues: ComponentHealth_Status.values)
    ..aOS(2, _omitFieldNames ? '' : 'detail')
    ..aOM<$0.Timestamp>(3, _omitFieldNames ? '' : 'checkedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aInt64(4, _omitFieldNames ? '' : 'latencyMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ComponentHealth clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ComponentHealth copyWith(void Function(ComponentHealth) updates) =>
      super.copyWith((message) => updates(message as ComponentHealth))
          as ComponentHealth;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ComponentHealth() / ComponentHealth.new instead')
  static ComponentHealth create() => ComponentHealth._();
  static $pb.GeneratedMessage $_createMessage() => ComponentHealth._();
  @$core.override
  ComponentHealth createEmptyInstance() => ComponentHealth._();
  @$core.pragma('dart2js:noInline')
  static ComponentHealth getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ComponentHealth>(
          ComponentHealth.$_createMessage);
  static ComponentHealth? _defaultInstance;

  @$pb.TagNumber(1)
  ComponentHealth_Status get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(ComponentHealth_Status value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get detail => $_getSZ(1);
  @$pb.TagNumber(2)
  set detail($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDetail() => $_has(1);
  @$pb.TagNumber(2)
  void clearDetail() => $_clearField(2);

  @$pb.TagNumber(3)
  $0.Timestamp get checkedAt => $_getN(2);
  @$pb.TagNumber(3)
  set checkedAt($0.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasCheckedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearCheckedAt() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Timestamp ensureCheckedAt() => $_ensure(2);

  @$pb.TagNumber(4)
  $fixnum.Int64 get latencyMs => $_getI64(3);
  @$pb.TagNumber(4)
  set latencyMs($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasLatencyMs() => $_has(3);
  @$pb.TagNumber(4)
  void clearLatencyMs() => $_clearField(4);
}

class DaemonHandshake extends $pb.GeneratedMessage {
  factory DaemonHandshake({
    $core.String? daemonRuntimeId,
    $core.String? name,
    $core.Iterable<$core.String>? acpRuntimes,
    $0.Timestamp? connectedAt,
    $core.String? os,
  }) {
    final result = DaemonHandshake._();
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    if (name != null) result.name = name;
    if (acpRuntimes != null) result.acpRuntimes.addAll(acpRuntimes);
    if (connectedAt != null) result.connectedAt = connectedAt;
    if (os != null) result.os = os;
    return result;
  }

  DaemonHandshake._();

  factory DaemonHandshake.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonHandshake()..mergeFromBuffer(data, registry);
  factory DaemonHandshake.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonHandshake()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DaemonHandshake',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DaemonHandshake.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..pPS(3, _omitFieldNames ? '' : 'acpRuntimes')
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'connectedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'os')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonHandshake clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonHandshake copyWith(void Function(DaemonHandshake) updates) =>
      super.copyWith((message) => updates(message as DaemonHandshake))
          as DaemonHandshake;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DaemonHandshake() / DaemonHandshake.new instead')
  static DaemonHandshake create() => DaemonHandshake._();
  static $pb.GeneratedMessage $_createMessage() => DaemonHandshake._();
  @$core.override
  DaemonHandshake createEmptyInstance() => DaemonHandshake._();
  @$core.pragma('dart2js:noInline')
  static DaemonHandshake getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DaemonHandshake>(
          DaemonHandshake.$_createMessage);
  static DaemonHandshake? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get daemonRuntimeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set daemonRuntimeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemonRuntimeId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemonRuntimeId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get acpRuntimes => $_getList(2);

  @$pb.TagNumber(4)
  $0.Timestamp get connectedAt => $_getN(3);
  @$pb.TagNumber(4)
  set connectedAt($0.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasConnectedAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearConnectedAt() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Timestamp ensureConnectedAt() => $_ensure(3);

  /// OS identifier reported by the daemon at handshake (e.g. "linux-amd64").
  @$pb.TagNumber(5)
  $core.String get os => $_getSZ(4);
  @$pb.TagNumber(5)
  set os($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOs() => $_has(4);
  @$pb.TagNumber(5)
  void clearOs() => $_clearField(5);
}

class GetBridgeDiagnosticsRequest extends $pb.GeneratedMessage {
  factory GetBridgeDiagnosticsRequest({
    $1.Duration? window,
  }) {
    final result = GetBridgeDiagnosticsRequest._();
    if (window != null) result.window = window;
    return result;
  }

  GetBridgeDiagnosticsRequest._();

  factory GetBridgeDiagnosticsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBridgeDiagnosticsRequest()..mergeFromBuffer(data, registry);
  factory GetBridgeDiagnosticsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBridgeDiagnosticsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBridgeDiagnosticsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetBridgeDiagnosticsRequest.$_createMessage)
    ..aOM<$1.Duration>(1, _omitFieldNames ? '' : 'window',
        subBuilder: $1.Duration.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBridgeDiagnosticsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBridgeDiagnosticsRequest copyWith(
          void Function(GetBridgeDiagnosticsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetBridgeDiagnosticsRequest))
          as GetBridgeDiagnosticsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetBridgeDiagnosticsRequest() / GetBridgeDiagnosticsRequest.new instead')
  static GetBridgeDiagnosticsRequest create() =>
      GetBridgeDiagnosticsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetBridgeDiagnosticsRequest._();
  @$core.override
  GetBridgeDiagnosticsRequest createEmptyInstance() =>
      GetBridgeDiagnosticsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetBridgeDiagnosticsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetBridgeDiagnosticsRequest>(
          GetBridgeDiagnosticsRequest.$_createMessage);
  static GetBridgeDiagnosticsRequest? _defaultInstance;

  /// Reserved for future windowing (e.g. last 1h / 24h). Currently ignored —
  /// the server returns a rolling in-memory sample window.
  @$pb.TagNumber(1)
  $1.Duration get window => $_getN(0);
  @$pb.TagNumber(1)
  set window($1.Duration value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWindow() => $_has(0);
  @$pb.TagNumber(1)
  void clearWindow() => $_clearField(1);
  @$pb.TagNumber(1)
  $1.Duration ensureWindow() => $_ensure(0);
}

class GetBridgeDiagnosticsResponse extends $pb.GeneratedMessage {
  factory GetBridgeDiagnosticsResponse({
    BridgeDiagnostics? diagnostics,
  }) {
    final result = GetBridgeDiagnosticsResponse._();
    if (diagnostics != null) result.diagnostics = diagnostics;
    return result;
  }

  GetBridgeDiagnosticsResponse._();

  factory GetBridgeDiagnosticsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBridgeDiagnosticsResponse()..mergeFromBuffer(data, registry);
  factory GetBridgeDiagnosticsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetBridgeDiagnosticsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetBridgeDiagnosticsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetBridgeDiagnosticsResponse.$_createMessage)
    ..aOM<BridgeDiagnostics>(1, _omitFieldNames ? '' : 'diagnostics',
        subBuilder: BridgeDiagnostics.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBridgeDiagnosticsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetBridgeDiagnosticsResponse copyWith(
          void Function(GetBridgeDiagnosticsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetBridgeDiagnosticsResponse))
          as GetBridgeDiagnosticsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetBridgeDiagnosticsResponse() / GetBridgeDiagnosticsResponse.new instead')
  static GetBridgeDiagnosticsResponse create() =>
      GetBridgeDiagnosticsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetBridgeDiagnosticsResponse._();
  @$core.override
  GetBridgeDiagnosticsResponse createEmptyInstance() =>
      GetBridgeDiagnosticsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetBridgeDiagnosticsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetBridgeDiagnosticsResponse>(
          GetBridgeDiagnosticsResponse.$_createMessage);
  static GetBridgeDiagnosticsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  BridgeDiagnostics get diagnostics => $_getN(0);
  @$pb.TagNumber(1)
  set diagnostics(BridgeDiagnostics value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDiagnostics() => $_has(0);
  @$pb.TagNumber(1)
  void clearDiagnostics() => $_clearField(1);
  @$pb.TagNumber(1)
  BridgeDiagnostics ensureDiagnostics() => $_ensure(0);
}

class BridgeDiagnostics extends $pb.GeneratedMessage {
  factory BridgeDiagnostics({
    $core.double? cpuPercent,
    $fixnum.Int64? memoryUsedBytes,
    $fixnum.Int64? memoryLimitBytes,
    $core.int? goroutines,
    $0.Timestamp? checkedAt,
    $core.Iterable<LatencyPoint>? latency,
  }) {
    final result = BridgeDiagnostics._();
    if (cpuPercent != null) result.cpuPercent = cpuPercent;
    if (memoryUsedBytes != null) result.memoryUsedBytes = memoryUsedBytes;
    if (memoryLimitBytes != null) result.memoryLimitBytes = memoryLimitBytes;
    if (goroutines != null) result.goroutines = goroutines;
    if (checkedAt != null) result.checkedAt = checkedAt;
    if (latency != null) result.latency.addAll(latency);
    return result;
  }

  BridgeDiagnostics._();

  factory BridgeDiagnostics.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BridgeDiagnostics()..mergeFromBuffer(data, registry);
  factory BridgeDiagnostics.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BridgeDiagnostics()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BridgeDiagnostics',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: BridgeDiagnostics.$_createMessage)
    ..aD(1, _omitFieldNames ? '' : 'cpuPercent')
    ..aInt64(2, _omitFieldNames ? '' : 'memoryUsedBytes')
    ..aInt64(3, _omitFieldNames ? '' : 'memoryLimitBytes')
    ..aI(4, _omitFieldNames ? '' : 'goroutines')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'checkedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..pPM<LatencyPoint>(6, _omitFieldNames ? '' : 'latency',
        subBuilder: LatencyPoint.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BridgeDiagnostics clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BridgeDiagnostics copyWith(void Function(BridgeDiagnostics) updates) =>
      super.copyWith((message) => updates(message as BridgeDiagnostics))
          as BridgeDiagnostics;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BridgeDiagnostics() / BridgeDiagnostics.new instead')
  static BridgeDiagnostics create() => BridgeDiagnostics._();
  static $pb.GeneratedMessage $_createMessage() => BridgeDiagnostics._();
  @$core.override
  BridgeDiagnostics createEmptyInstance() => BridgeDiagnostics._();
  @$core.pragma('dart2js:noInline')
  static BridgeDiagnostics getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BridgeDiagnostics>(
          BridgeDiagnostics.$_createMessage);
  static BridgeDiagnostics? _defaultInstance;

  /// Process-level metrics for the butter server (the "router").
  @$pb.TagNumber(1)
  $core.double get cpuPercent => $_getN(0);
  @$pb.TagNumber(1)
  set cpuPercent($core.double value) => $_setDouble(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCpuPercent() => $_has(0);
  @$pb.TagNumber(1)
  void clearCpuPercent() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get memoryUsedBytes => $_getI64(1);
  @$pb.TagNumber(2)
  set memoryUsedBytes($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMemoryUsedBytes() => $_has(1);
  @$pb.TagNumber(2)
  void clearMemoryUsedBytes() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get memoryLimitBytes => $_getI64(2);
  @$pb.TagNumber(3)
  set memoryLimitBytes($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMemoryLimitBytes() => $_has(2);
  @$pb.TagNumber(3)
  void clearMemoryLimitBytes() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get goroutines => $_getIZ(3);
  @$pb.TagNumber(4)
  set goroutines($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasGoroutines() => $_has(3);
  @$pb.TagNumber(4)
  void clearGoroutines() => $_clearField(4);

  @$pb.TagNumber(5)
  $0.Timestamp get checkedAt => $_getN(4);
  @$pb.TagNumber(5)
  set checkedAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCheckedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearCheckedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureCheckedAt() => $_ensure(4);

  /// Rolling samples of bridge invocation latency (oldest first).
  @$pb.TagNumber(6)
  $pb.PbList<LatencyPoint> get latency => $_getList(5);
}

class LatencyPoint extends $pb.GeneratedMessage {
  factory LatencyPoint({
    $0.Timestamp? timestamp,
    $fixnum.Int64? latencyMs,
  }) {
    final result = LatencyPoint._();
    if (timestamp != null) result.timestamp = timestamp;
    if (latencyMs != null) result.latencyMs = latencyMs;
    return result;
  }

  LatencyPoint._();

  factory LatencyPoint.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LatencyPoint()..mergeFromBuffer(data, registry);
  factory LatencyPoint.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LatencyPoint()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LatencyPoint',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LatencyPoint.$_createMessage)
    ..aOM<$0.Timestamp>(1, _omitFieldNames ? '' : 'timestamp',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aInt64(2, _omitFieldNames ? '' : 'latencyMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LatencyPoint clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LatencyPoint copyWith(void Function(LatencyPoint) updates) =>
      super.copyWith((message) => updates(message as LatencyPoint))
          as LatencyPoint;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LatencyPoint() / LatencyPoint.new instead')
  static LatencyPoint create() => LatencyPoint._();
  static $pb.GeneratedMessage $_createMessage() => LatencyPoint._();
  @$core.override
  LatencyPoint createEmptyInstance() => LatencyPoint._();
  @$core.pragma('dart2js:noInline')
  static LatencyPoint getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LatencyPoint>(
          LatencyPoint.$_createMessage);
  static LatencyPoint? _defaultInstance;

  @$pb.TagNumber(1)
  $0.Timestamp get timestamp => $_getN(0);
  @$pb.TagNumber(1)
  set timestamp($0.Timestamp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTimestamp() => $_has(0);
  @$pb.TagNumber(1)
  void clearTimestamp() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Timestamp ensureTimestamp() => $_ensure(0);

  @$pb.TagNumber(2)
  $fixnum.Int64 get latencyMs => $_getI64(1);
  @$pb.TagNumber(2)
  set latencyMs($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLatencyMs() => $_has(1);
  @$pb.TagNumber(2)
  void clearLatencyMs() => $_clearField(2);
}

class DaemonRuntime extends $pb.GeneratedMessage {
  factory DaemonRuntime({
    $core.String? id,
    $core.String? name,
    $core.String? description,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? labels,
    $0.Timestamp? createdAt,
    $core.String? createdBy,
    $core.String? workspaceId,
  }) {
    final result = DaemonRuntime._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (labels != null) result.labels.addEntries(labels);
    if (createdAt != null) result.createdAt = createdAt;
    if (createdBy != null) result.createdBy = createdBy;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  DaemonRuntime._();

  factory DaemonRuntime.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonRuntime()..mergeFromBuffer(data, registry);
  factory DaemonRuntime.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonRuntime()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DaemonRuntime',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DaemonRuntime.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..m<$core.String, $core.String>(5, _omitFieldNames ? '' : 'labels',
        entryClassName: 'DaemonRuntime.LabelsEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(7, _omitFieldNames ? '' : 'createdBy')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonRuntime clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonRuntime copyWith(void Function(DaemonRuntime) updates) =>
      super.copyWith((message) => updates(message as DaemonRuntime))
          as DaemonRuntime;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DaemonRuntime() / DaemonRuntime.new instead')
  static DaemonRuntime create() => DaemonRuntime._();
  static $pb.GeneratedMessage $_createMessage() => DaemonRuntime._();
  @$core.override
  DaemonRuntime createEmptyInstance() => DaemonRuntime._();
  @$core.pragma('dart2js:noInline')
  static DaemonRuntime getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DaemonRuntime>(
          DaemonRuntime.$_createMessage);
  static DaemonRuntime? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  @$pb.TagNumber(5)
  $pb.PbMap<$core.String, $core.String> get labels => $_getMap(3);

  @$pb.TagNumber(6)
  $0.Timestamp get createdAt => $_getN(4);
  @$pb.TagNumber(6)
  set createdAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCreatedAt() => $_has(4);
  @$pb.TagNumber(6)
  void clearCreatedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureCreatedAt() => $_ensure(4);

  @$pb.TagNumber(7)
  $core.String get createdBy => $_getSZ(5);
  @$pb.TagNumber(7)
  set createdBy($core.String value) => $_setString(5, value);
  @$pb.TagNumber(7)
  $core.bool hasCreatedBy() => $_has(5);
  @$pb.TagNumber(7)
  void clearCreatedBy() => $_clearField(7);

  /// ID of the workspace this daemon runtime belongs to.
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(6);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(6);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListDaemonRuntimesRequest extends $pb.GeneratedMessage {
  factory ListDaemonRuntimesRequest() => ListDaemonRuntimesRequest._();

  ListDaemonRuntimesRequest._();

  factory ListDaemonRuntimesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonRuntimesRequest()..mergeFromBuffer(data, registry);
  factory ListDaemonRuntimesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonRuntimesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDaemonRuntimesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListDaemonRuntimesRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonRuntimesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonRuntimesRequest copyWith(
          void Function(ListDaemonRuntimesRequest) updates) =>
      super.copyWith((message) => updates(message as ListDaemonRuntimesRequest))
          as ListDaemonRuntimesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListDaemonRuntimesRequest() / ListDaemonRuntimesRequest.new instead')
  static ListDaemonRuntimesRequest create() => ListDaemonRuntimesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListDaemonRuntimesRequest._();
  @$core.override
  ListDaemonRuntimesRequest createEmptyInstance() =>
      ListDaemonRuntimesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListDaemonRuntimesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDaemonRuntimesRequest>(
          ListDaemonRuntimesRequest.$_createMessage);
  static ListDaemonRuntimesRequest? _defaultInstance;
}

class ListDaemonRuntimesResponse extends $pb.GeneratedMessage {
  factory ListDaemonRuntimesResponse({
    $core.Iterable<DaemonRuntime>? runtimes,
  }) {
    final result = ListDaemonRuntimesResponse._();
    if (runtimes != null) result.runtimes.addAll(runtimes);
    return result;
  }

  ListDaemonRuntimesResponse._();

  factory ListDaemonRuntimesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonRuntimesResponse()..mergeFromBuffer(data, registry);
  factory ListDaemonRuntimesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonRuntimesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDaemonRuntimesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListDaemonRuntimesResponse.$_createMessage)
    ..pPM<DaemonRuntime>(1, _omitFieldNames ? '' : 'runtimes',
        subBuilder: DaemonRuntime.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonRuntimesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonRuntimesResponse copyWith(
          void Function(ListDaemonRuntimesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListDaemonRuntimesResponse))
          as ListDaemonRuntimesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListDaemonRuntimesResponse() / ListDaemonRuntimesResponse.new instead')
  static ListDaemonRuntimesResponse create() => ListDaemonRuntimesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListDaemonRuntimesResponse._();
  @$core.override
  ListDaemonRuntimesResponse createEmptyInstance() =>
      ListDaemonRuntimesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListDaemonRuntimesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDaemonRuntimesResponse>(
          ListDaemonRuntimesResponse.$_createMessage);
  static ListDaemonRuntimesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<DaemonRuntime> get runtimes => $_getList(0);
}

class GetDaemonRuntimeRequest extends $pb.GeneratedMessage {
  factory GetDaemonRuntimeRequest({
    $core.String? id,
  }) {
    final result = GetDaemonRuntimeRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetDaemonRuntimeRequest._();

  factory GetDaemonRuntimeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonRuntimeRequest()..mergeFromBuffer(data, registry);
  factory GetDaemonRuntimeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonRuntimeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDaemonRuntimeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetDaemonRuntimeRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonRuntimeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonRuntimeRequest copyWith(
          void Function(GetDaemonRuntimeRequest) updates) =>
      super.copyWith((message) => updates(message as GetDaemonRuntimeRequest))
          as GetDaemonRuntimeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetDaemonRuntimeRequest() / GetDaemonRuntimeRequest.new instead')
  static GetDaemonRuntimeRequest create() => GetDaemonRuntimeRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetDaemonRuntimeRequest._();
  @$core.override
  GetDaemonRuntimeRequest createEmptyInstance() => GetDaemonRuntimeRequest._();
  @$core.pragma('dart2js:noInline')
  static GetDaemonRuntimeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetDaemonRuntimeRequest>(
          GetDaemonRuntimeRequest.$_createMessage);
  static GetDaemonRuntimeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetDaemonRuntimeResponse extends $pb.GeneratedMessage {
  factory GetDaemonRuntimeResponse({
    DaemonRuntime? runtime,
  }) {
    final result = GetDaemonRuntimeResponse._();
    if (runtime != null) result.runtime = runtime;
    return result;
  }

  GetDaemonRuntimeResponse._();

  factory GetDaemonRuntimeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonRuntimeResponse()..mergeFromBuffer(data, registry);
  factory GetDaemonRuntimeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonRuntimeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDaemonRuntimeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetDaemonRuntimeResponse.$_createMessage)
    ..aOM<DaemonRuntime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: DaemonRuntime.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonRuntimeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonRuntimeResponse copyWith(
          void Function(GetDaemonRuntimeResponse) updates) =>
      super.copyWith((message) => updates(message as GetDaemonRuntimeResponse))
          as GetDaemonRuntimeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetDaemonRuntimeResponse() / GetDaemonRuntimeResponse.new instead')
  static GetDaemonRuntimeResponse create() => GetDaemonRuntimeResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetDaemonRuntimeResponse._();
  @$core.override
  GetDaemonRuntimeResponse createEmptyInstance() =>
      GetDaemonRuntimeResponse._();
  @$core.pragma('dart2js:noInline')
  static GetDaemonRuntimeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetDaemonRuntimeResponse>(
          GetDaemonRuntimeResponse.$_createMessage);
  static GetDaemonRuntimeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  DaemonRuntime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(DaemonRuntime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  DaemonRuntime ensureRuntime() => $_ensure(0);
}

class CreateDaemonRuntimeRequest extends $pb.GeneratedMessage {
  factory CreateDaemonRuntimeRequest({
    DaemonRuntime? runtime,
  }) {
    final result = CreateDaemonRuntimeRequest._();
    if (runtime != null) result.runtime = runtime;
    return result;
  }

  CreateDaemonRuntimeRequest._();

  factory CreateDaemonRuntimeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeRequest()..mergeFromBuffer(data, registry);
  factory CreateDaemonRuntimeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDaemonRuntimeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateDaemonRuntimeRequest.$_createMessage)
    ..aOM<DaemonRuntime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: DaemonRuntime.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeRequest copyWith(
          void Function(CreateDaemonRuntimeRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateDaemonRuntimeRequest))
          as CreateDaemonRuntimeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDaemonRuntimeRequest() / CreateDaemonRuntimeRequest.new instead')
  static CreateDaemonRuntimeRequest create() => CreateDaemonRuntimeRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateDaemonRuntimeRequest._();
  @$core.override
  CreateDaemonRuntimeRequest createEmptyInstance() =>
      CreateDaemonRuntimeRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateDaemonRuntimeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDaemonRuntimeRequest>(
          CreateDaemonRuntimeRequest.$_createMessage);
  static CreateDaemonRuntimeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  DaemonRuntime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(DaemonRuntime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  DaemonRuntime ensureRuntime() => $_ensure(0);
}

class CreateDaemonRuntimeResponse extends $pb.GeneratedMessage {
  factory CreateDaemonRuntimeResponse({
    DaemonRuntime? runtime,
  }) {
    final result = CreateDaemonRuntimeResponse._();
    if (runtime != null) result.runtime = runtime;
    return result;
  }

  CreateDaemonRuntimeResponse._();

  factory CreateDaemonRuntimeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeResponse()..mergeFromBuffer(data, registry);
  factory CreateDaemonRuntimeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDaemonRuntimeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateDaemonRuntimeResponse.$_createMessage)
    ..aOM<DaemonRuntime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: DaemonRuntime.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeResponse copyWith(
          void Function(CreateDaemonRuntimeResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateDaemonRuntimeResponse))
          as CreateDaemonRuntimeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDaemonRuntimeResponse() / CreateDaemonRuntimeResponse.new instead')
  static CreateDaemonRuntimeResponse create() =>
      CreateDaemonRuntimeResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateDaemonRuntimeResponse._();
  @$core.override
  CreateDaemonRuntimeResponse createEmptyInstance() =>
      CreateDaemonRuntimeResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateDaemonRuntimeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDaemonRuntimeResponse>(
          CreateDaemonRuntimeResponse.$_createMessage);
  static CreateDaemonRuntimeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  DaemonRuntime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(DaemonRuntime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  DaemonRuntime ensureRuntime() => $_ensure(0);
}

class UpdateDaemonRuntimeRequest extends $pb.GeneratedMessage {
  factory UpdateDaemonRuntimeRequest({
    DaemonRuntime? runtime,
  }) {
    final result = UpdateDaemonRuntimeRequest._();
    if (runtime != null) result.runtime = runtime;
    return result;
  }

  UpdateDaemonRuntimeRequest._();

  factory UpdateDaemonRuntimeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDaemonRuntimeRequest()..mergeFromBuffer(data, registry);
  factory UpdateDaemonRuntimeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDaemonRuntimeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateDaemonRuntimeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateDaemonRuntimeRequest.$_createMessage)
    ..aOM<DaemonRuntime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: DaemonRuntime.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDaemonRuntimeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDaemonRuntimeRequest copyWith(
          void Function(UpdateDaemonRuntimeRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateDaemonRuntimeRequest))
          as UpdateDaemonRuntimeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateDaemonRuntimeRequest() / UpdateDaemonRuntimeRequest.new instead')
  static UpdateDaemonRuntimeRequest create() => UpdateDaemonRuntimeRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateDaemonRuntimeRequest._();
  @$core.override
  UpdateDaemonRuntimeRequest createEmptyInstance() =>
      UpdateDaemonRuntimeRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateDaemonRuntimeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateDaemonRuntimeRequest>(
          UpdateDaemonRuntimeRequest.$_createMessage);
  static UpdateDaemonRuntimeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  DaemonRuntime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(DaemonRuntime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  DaemonRuntime ensureRuntime() => $_ensure(0);
}

class UpdateDaemonRuntimeResponse extends $pb.GeneratedMessage {
  factory UpdateDaemonRuntimeResponse({
    DaemonRuntime? runtime,
  }) {
    final result = UpdateDaemonRuntimeResponse._();
    if (runtime != null) result.runtime = runtime;
    return result;
  }

  UpdateDaemonRuntimeResponse._();

  factory UpdateDaemonRuntimeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDaemonRuntimeResponse()..mergeFromBuffer(data, registry);
  factory UpdateDaemonRuntimeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateDaemonRuntimeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateDaemonRuntimeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateDaemonRuntimeResponse.$_createMessage)
    ..aOM<DaemonRuntime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: DaemonRuntime.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDaemonRuntimeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateDaemonRuntimeResponse copyWith(
          void Function(UpdateDaemonRuntimeResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateDaemonRuntimeResponse))
          as UpdateDaemonRuntimeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateDaemonRuntimeResponse() / UpdateDaemonRuntimeResponse.new instead')
  static UpdateDaemonRuntimeResponse create() =>
      UpdateDaemonRuntimeResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateDaemonRuntimeResponse._();
  @$core.override
  UpdateDaemonRuntimeResponse createEmptyInstance() =>
      UpdateDaemonRuntimeResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateDaemonRuntimeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateDaemonRuntimeResponse>(
          UpdateDaemonRuntimeResponse.$_createMessage);
  static UpdateDaemonRuntimeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  DaemonRuntime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(DaemonRuntime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  DaemonRuntime ensureRuntime() => $_ensure(0);
}

class DeleteDaemonRuntimeRequest extends $pb.GeneratedMessage {
  factory DeleteDaemonRuntimeRequest({
    $core.String? id,
  }) {
    final result = DeleteDaemonRuntimeRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteDaemonRuntimeRequest._();

  factory DeleteDaemonRuntimeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDaemonRuntimeRequest()..mergeFromBuffer(data, registry);
  factory DeleteDaemonRuntimeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDaemonRuntimeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteDaemonRuntimeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteDaemonRuntimeRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDaemonRuntimeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDaemonRuntimeRequest copyWith(
          void Function(DeleteDaemonRuntimeRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteDaemonRuntimeRequest))
          as DeleteDaemonRuntimeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteDaemonRuntimeRequest() / DeleteDaemonRuntimeRequest.new instead')
  static DeleteDaemonRuntimeRequest create() => DeleteDaemonRuntimeRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteDaemonRuntimeRequest._();
  @$core.override
  DeleteDaemonRuntimeRequest createEmptyInstance() =>
      DeleteDaemonRuntimeRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteDaemonRuntimeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteDaemonRuntimeRequest>(
          DeleteDaemonRuntimeRequest.$_createMessage);
  static DeleteDaemonRuntimeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteDaemonRuntimeResponse extends $pb.GeneratedMessage {
  factory DeleteDaemonRuntimeResponse() => DeleteDaemonRuntimeResponse._();

  DeleteDaemonRuntimeResponse._();

  factory DeleteDaemonRuntimeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDaemonRuntimeResponse()..mergeFromBuffer(data, registry);
  factory DeleteDaemonRuntimeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteDaemonRuntimeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteDaemonRuntimeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteDaemonRuntimeResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDaemonRuntimeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDaemonRuntimeResponse copyWith(
          void Function(DeleteDaemonRuntimeResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteDaemonRuntimeResponse))
          as DeleteDaemonRuntimeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteDaemonRuntimeResponse() / DeleteDaemonRuntimeResponse.new instead')
  static DeleteDaemonRuntimeResponse create() =>
      DeleteDaemonRuntimeResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteDaemonRuntimeResponse._();
  @$core.override
  DeleteDaemonRuntimeResponse createEmptyInstance() =>
      DeleteDaemonRuntimeResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteDaemonRuntimeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteDaemonRuntimeResponse>(
          DeleteDaemonRuntimeResponse.$_createMessage);
  static DeleteDaemonRuntimeResponse? _defaultInstance;
}

class CreateDaemonRuntimeTokenRequest extends $pb.GeneratedMessage {
  factory CreateDaemonRuntimeTokenRequest({
    $core.String? daemonRuntimeId,
    $core.String? name,
    $1.Duration? ttl,
  }) {
    final result = CreateDaemonRuntimeTokenRequest._();
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    if (name != null) result.name = name;
    if (ttl != null) result.ttl = ttl;
    return result;
  }

  CreateDaemonRuntimeTokenRequest._();

  factory CreateDaemonRuntimeTokenRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeTokenRequest()..mergeFromBuffer(data, registry);
  factory CreateDaemonRuntimeTokenRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeTokenRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDaemonRuntimeTokenRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateDaemonRuntimeTokenRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOM<$1.Duration>(3, _omitFieldNames ? '' : 'ttl',
        subBuilder: $1.Duration.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeTokenRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeTokenRequest copyWith(
          void Function(CreateDaemonRuntimeTokenRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateDaemonRuntimeTokenRequest))
          as CreateDaemonRuntimeTokenRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDaemonRuntimeTokenRequest() / CreateDaemonRuntimeTokenRequest.new instead')
  static CreateDaemonRuntimeTokenRequest create() =>
      CreateDaemonRuntimeTokenRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateDaemonRuntimeTokenRequest._();
  @$core.override
  CreateDaemonRuntimeTokenRequest createEmptyInstance() =>
      CreateDaemonRuntimeTokenRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateDaemonRuntimeTokenRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDaemonRuntimeTokenRequest>(
          CreateDaemonRuntimeTokenRequest.$_createMessage);
  static CreateDaemonRuntimeTokenRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get daemonRuntimeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set daemonRuntimeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemonRuntimeId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemonRuntimeId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $1.Duration get ttl => $_getN(2);
  @$pb.TagNumber(3)
  set ttl($1.Duration value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasTtl() => $_has(2);
  @$pb.TagNumber(3)
  void clearTtl() => $_clearField(3);
  @$pb.TagNumber(3)
  $1.Duration ensureTtl() => $_ensure(2);
}

class CreateDaemonRuntimeTokenResponse extends $pb.GeneratedMessage {
  factory CreateDaemonRuntimeTokenResponse({
    $2.APIToken? token,
    $core.String? secret,
  }) {
    final result = CreateDaemonRuntimeTokenResponse._();
    if (token != null) result.token = token;
    if (secret != null) result.secret = secret;
    return result;
  }

  CreateDaemonRuntimeTokenResponse._();

  factory CreateDaemonRuntimeTokenResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeTokenResponse()..mergeFromBuffer(data, registry);
  factory CreateDaemonRuntimeTokenResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateDaemonRuntimeTokenResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateDaemonRuntimeTokenResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateDaemonRuntimeTokenResponse.$_createMessage)
    ..aOM<$2.APIToken>(1, _omitFieldNames ? '' : 'token',
        subBuilder: $2.APIToken.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'secret')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeTokenResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateDaemonRuntimeTokenResponse copyWith(
          void Function(CreateDaemonRuntimeTokenResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateDaemonRuntimeTokenResponse))
          as CreateDaemonRuntimeTokenResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateDaemonRuntimeTokenResponse() / CreateDaemonRuntimeTokenResponse.new instead')
  static CreateDaemonRuntimeTokenResponse create() =>
      CreateDaemonRuntimeTokenResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateDaemonRuntimeTokenResponse._();
  @$core.override
  CreateDaemonRuntimeTokenResponse createEmptyInstance() =>
      CreateDaemonRuntimeTokenResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateDaemonRuntimeTokenResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateDaemonRuntimeTokenResponse>(
          CreateDaemonRuntimeTokenResponse.$_createMessage);
  static CreateDaemonRuntimeTokenResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $2.APIToken get token => $_getN(0);
  @$pb.TagNumber(1)
  set token($2.APIToken value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);
  @$pb.TagNumber(1)
  $2.APIToken ensureToken() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get secret => $_getSZ(1);
  @$pb.TagNumber(2)
  set secret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearSecret() => $_clearField(2);
}

class ListDaemonsRequest extends $pb.GeneratedMessage {
  factory ListDaemonsRequest() => ListDaemonsRequest._();

  ListDaemonsRequest._();

  factory ListDaemonsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonsRequest()..mergeFromBuffer(data, registry);
  factory ListDaemonsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDaemonsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListDaemonsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonsRequest copyWith(void Function(ListDaemonsRequest) updates) =>
      super.copyWith((message) => updates(message as ListDaemonsRequest))
          as ListDaemonsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListDaemonsRequest() / ListDaemonsRequest.new instead')
  static ListDaemonsRequest create() => ListDaemonsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListDaemonsRequest._();
  @$core.override
  ListDaemonsRequest createEmptyInstance() => ListDaemonsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListDaemonsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDaemonsRequest>(
          ListDaemonsRequest.$_createMessage);
  static ListDaemonsRequest? _defaultInstance;
}

class ListDaemonsResponse extends $pb.GeneratedMessage {
  factory ListDaemonsResponse({
    $core.Iterable<DaemonStatus>? daemons,
  }) {
    final result = ListDaemonsResponse._();
    if (daemons != null) result.daemons.addAll(daemons);
    return result;
  }

  ListDaemonsResponse._();

  factory ListDaemonsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonsResponse()..mergeFromBuffer(data, registry);
  factory ListDaemonsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDaemonsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListDaemonsResponse.$_createMessage)
    ..pPM<DaemonStatus>(1, _omitFieldNames ? '' : 'daemons',
        subBuilder: DaemonStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonsResponse copyWith(void Function(ListDaemonsResponse) updates) =>
      super.copyWith((message) => updates(message as ListDaemonsResponse))
          as ListDaemonsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListDaemonsResponse() / ListDaemonsResponse.new instead')
  static ListDaemonsResponse create() => ListDaemonsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListDaemonsResponse._();
  @$core.override
  ListDaemonsResponse createEmptyInstance() => ListDaemonsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListDaemonsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDaemonsResponse>(
          ListDaemonsResponse.$_createMessage);
  static ListDaemonsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<DaemonStatus> get daemons => $_getList(0);
}

class GetDaemonRequest extends $pb.GeneratedMessage {
  factory GetDaemonRequest({
    $core.String? daemonRuntimeId,
  }) {
    final result = GetDaemonRequest._();
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    return result;
  }

  GetDaemonRequest._();

  factory GetDaemonRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonRequest()..mergeFromBuffer(data, registry);
  factory GetDaemonRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDaemonRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetDaemonRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonRequest copyWith(void Function(GetDaemonRequest) updates) =>
      super.copyWith((message) => updates(message as GetDaemonRequest))
          as GetDaemonRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetDaemonRequest() / GetDaemonRequest.new instead')
  static GetDaemonRequest create() => GetDaemonRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetDaemonRequest._();
  @$core.override
  GetDaemonRequest createEmptyInstance() => GetDaemonRequest._();
  @$core.pragma('dart2js:noInline')
  static GetDaemonRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetDaemonRequest>(
          GetDaemonRequest.$_createMessage);
  static GetDaemonRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get daemonRuntimeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set daemonRuntimeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemonRuntimeId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemonRuntimeId() => $_clearField(1);
}

class GetDaemonResponse extends $pb.GeneratedMessage {
  factory GetDaemonResponse({
    DaemonStatus? daemon,
  }) {
    final result = GetDaemonResponse._();
    if (daemon != null) result.daemon = daemon;
    return result;
  }

  GetDaemonResponse._();

  factory GetDaemonResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonResponse()..mergeFromBuffer(data, registry);
  factory GetDaemonResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetDaemonResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDaemonResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetDaemonResponse.$_createMessage)
    ..aOM<DaemonStatus>(1, _omitFieldNames ? '' : 'daemon',
        subBuilder: DaemonStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonResponse copyWith(void Function(GetDaemonResponse) updates) =>
      super.copyWith((message) => updates(message as GetDaemonResponse))
          as GetDaemonResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetDaemonResponse() / GetDaemonResponse.new instead')
  static GetDaemonResponse create() => GetDaemonResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetDaemonResponse._();
  @$core.override
  GetDaemonResponse createEmptyInstance() => GetDaemonResponse._();
  @$core.pragma('dart2js:noInline')
  static GetDaemonResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetDaemonResponse>(
          GetDaemonResponse.$_createMessage);
  static GetDaemonResponse? _defaultInstance;

  @$pb.TagNumber(1)
  DaemonStatus get daemon => $_getN(0);
  @$pb.TagNumber(1)
  set daemon(DaemonStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemon() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemon() => $_clearField(1);
  @$pb.TagNumber(1)
  DaemonStatus ensureDaemon() => $_ensure(0);
}

class CancelDaemonTaskRequest extends $pb.GeneratedMessage {
  factory CancelDaemonTaskRequest({
    $core.String? taskId,
    $core.String? daemonRuntimeId,
  }) {
    final result = CancelDaemonTaskRequest._();
    if (taskId != null) result.taskId = taskId;
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    return result;
  }

  CancelDaemonTaskRequest._();

  factory CancelDaemonTaskRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDaemonTaskRequest()..mergeFromBuffer(data, registry);
  factory CancelDaemonTaskRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDaemonTaskRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelDaemonTaskRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CancelDaemonTaskRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'taskId')
    ..aOS(2, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDaemonTaskRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDaemonTaskRequest copyWith(
          void Function(CancelDaemonTaskRequest) updates) =>
      super.copyWith((message) => updates(message as CancelDaemonTaskRequest))
          as CancelDaemonTaskRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelDaemonTaskRequest() / CancelDaemonTaskRequest.new instead')
  static CancelDaemonTaskRequest create() => CancelDaemonTaskRequest._();
  static $pb.GeneratedMessage $_createMessage() => CancelDaemonTaskRequest._();
  @$core.override
  CancelDaemonTaskRequest createEmptyInstance() => CancelDaemonTaskRequest._();
  @$core.pragma('dart2js:noInline')
  static CancelDaemonTaskRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelDaemonTaskRequest>(
          CancelDaemonTaskRequest.$_createMessage);
  static CancelDaemonTaskRequest? _defaultInstance;

  /// Task id to cancel.
  @$pb.TagNumber(1)
  $core.String get taskId => $_getSZ(0);
  @$pb.TagNumber(1)
  set taskId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTaskId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTaskId() => $_clearField(1);

  /// Optional daemon runtime id hint; if empty the registry is scanned.
  @$pb.TagNumber(2)
  $core.String get daemonRuntimeId => $_getSZ(1);
  @$pb.TagNumber(2)
  set daemonRuntimeId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDaemonRuntimeId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDaemonRuntimeId() => $_clearField(2);
}

class CancelDaemonTaskResponse extends $pb.GeneratedMessage {
  factory CancelDaemonTaskResponse({
    $core.String? daemonRuntimeId,
  }) {
    final result = CancelDaemonTaskResponse._();
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    return result;
  }

  CancelDaemonTaskResponse._();

  factory CancelDaemonTaskResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDaemonTaskResponse()..mergeFromBuffer(data, registry);
  factory CancelDaemonTaskResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CancelDaemonTaskResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelDaemonTaskResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CancelDaemonTaskResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDaemonTaskResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelDaemonTaskResponse copyWith(
          void Function(CancelDaemonTaskResponse) updates) =>
      super.copyWith((message) => updates(message as CancelDaemonTaskResponse))
          as CancelDaemonTaskResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CancelDaemonTaskResponse() / CancelDaemonTaskResponse.new instead')
  static CancelDaemonTaskResponse create() => CancelDaemonTaskResponse._();
  static $pb.GeneratedMessage $_createMessage() => CancelDaemonTaskResponse._();
  @$core.override
  CancelDaemonTaskResponse createEmptyInstance() =>
      CancelDaemonTaskResponse._();
  @$core.pragma('dart2js:noInline')
  static CancelDaemonTaskResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelDaemonTaskResponse>(
          CancelDaemonTaskResponse.$_createMessage);
  static CancelDaemonTaskResponse? _defaultInstance;

  /// Runtime id that received the cancel signal.
  @$pb.TagNumber(1)
  $core.String get daemonRuntimeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set daemonRuntimeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemonRuntimeId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemonRuntimeId() => $_clearField(1);
}

class ListDaemonTasksRequest extends $pb.GeneratedMessage {
  factory ListDaemonTasksRequest({
    $core.String? daemonRuntimeId,
  }) {
    final result = ListDaemonTasksRequest._();
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    return result;
  }

  ListDaemonTasksRequest._();

  factory ListDaemonTasksRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonTasksRequest()..mergeFromBuffer(data, registry);
  factory ListDaemonTasksRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonTasksRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDaemonTasksRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListDaemonTasksRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonTasksRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonTasksRequest copyWith(
          void Function(ListDaemonTasksRequest) updates) =>
      super.copyWith((message) => updates(message as ListDaemonTasksRequest))
          as ListDaemonTasksRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListDaemonTasksRequest() / ListDaemonTasksRequest.new instead')
  static ListDaemonTasksRequest create() => ListDaemonTasksRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListDaemonTasksRequest._();
  @$core.override
  ListDaemonTasksRequest createEmptyInstance() => ListDaemonTasksRequest._();
  @$core.pragma('dart2js:noInline')
  static ListDaemonTasksRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDaemonTasksRequest>(
          ListDaemonTasksRequest.$_createMessage);
  static ListDaemonTasksRequest? _defaultInstance;

  /// Optional: only return tasks served by this daemon runtime id.
  @$pb.TagNumber(1)
  $core.String get daemonRuntimeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set daemonRuntimeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemonRuntimeId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemonRuntimeId() => $_clearField(1);
}

class ListDaemonTasksResponse extends $pb.GeneratedMessage {
  factory ListDaemonTasksResponse({
    $core.Iterable<DaemonTaskInFlight>? tasks,
  }) {
    final result = ListDaemonTasksResponse._();
    if (tasks != null) result.tasks.addAll(tasks);
    return result;
  }

  ListDaemonTasksResponse._();

  factory ListDaemonTasksResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonTasksResponse()..mergeFromBuffer(data, registry);
  factory ListDaemonTasksResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListDaemonTasksResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDaemonTasksResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListDaemonTasksResponse.$_createMessage)
    ..pPM<DaemonTaskInFlight>(1, _omitFieldNames ? '' : 'tasks',
        subBuilder: DaemonTaskInFlight.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonTasksResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDaemonTasksResponse copyWith(
          void Function(ListDaemonTasksResponse) updates) =>
      super.copyWith((message) => updates(message as ListDaemonTasksResponse))
          as ListDaemonTasksResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListDaemonTasksResponse() / ListDaemonTasksResponse.new instead')
  static ListDaemonTasksResponse create() => ListDaemonTasksResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListDaemonTasksResponse._();
  @$core.override
  ListDaemonTasksResponse createEmptyInstance() => ListDaemonTasksResponse._();
  @$core.pragma('dart2js:noInline')
  static ListDaemonTasksResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDaemonTasksResponse>(
          ListDaemonTasksResponse.$_createMessage);
  static ListDaemonTasksResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<DaemonTaskInFlight> get tasks => $_getList(0);
}

class DaemonTaskInFlight extends $pb.GeneratedMessage {
  factory DaemonTaskInFlight({
    $core.String? taskId,
    $core.String? daemonRuntimeId,
    $core.String? daemonName,
    $core.String? acpRuntime,
    $0.Timestamp? startedAt,
    $1.Duration? elapsed,
    $core.String? currentStep,
    $core.int? progress,
    $core.String? agentName,
    $core.String? workspaceId,
    $core.String? agentId,
  }) {
    final result = DaemonTaskInFlight._();
    if (taskId != null) result.taskId = taskId;
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    if (daemonName != null) result.daemonName = daemonName;
    if (acpRuntime != null) result.acpRuntime = acpRuntime;
    if (startedAt != null) result.startedAt = startedAt;
    if (elapsed != null) result.elapsed = elapsed;
    if (currentStep != null) result.currentStep = currentStep;
    if (progress != null) result.progress = progress;
    if (agentName != null) result.agentName = agentName;
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  DaemonTaskInFlight._();

  factory DaemonTaskInFlight.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonTaskInFlight()..mergeFromBuffer(data, registry);
  factory DaemonTaskInFlight.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonTaskInFlight()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DaemonTaskInFlight',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DaemonTaskInFlight.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'taskId')
    ..aOS(2, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..aOS(3, _omitFieldNames ? '' : 'daemonName')
    ..aOS(4, _omitFieldNames ? '' : 'acpRuntime')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$1.Duration>(6, _omitFieldNames ? '' : 'elapsed',
        subBuilder: $1.Duration.$_createMessage)
    ..aOS(7, _omitFieldNames ? '' : 'currentStep')
    ..aI(8, _omitFieldNames ? '' : 'progress')
    ..aOS(9, _omitFieldNames ? '' : 'agentName')
    ..aOS(10, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(11, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonTaskInFlight clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonTaskInFlight copyWith(void Function(DaemonTaskInFlight) updates) =>
      super.copyWith((message) => updates(message as DaemonTaskInFlight))
          as DaemonTaskInFlight;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DaemonTaskInFlight() / DaemonTaskInFlight.new instead')
  static DaemonTaskInFlight create() => DaemonTaskInFlight._();
  static $pb.GeneratedMessage $_createMessage() => DaemonTaskInFlight._();
  @$core.override
  DaemonTaskInFlight createEmptyInstance() => DaemonTaskInFlight._();
  @$core.pragma('dart2js:noInline')
  static DaemonTaskInFlight getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DaemonTaskInFlight>(
          DaemonTaskInFlight.$_createMessage);
  static DaemonTaskInFlight? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get taskId => $_getSZ(0);
  @$pb.TagNumber(1)
  set taskId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTaskId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTaskId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get daemonRuntimeId => $_getSZ(1);
  @$pb.TagNumber(2)
  set daemonRuntimeId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDaemonRuntimeId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDaemonRuntimeId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get daemonName => $_getSZ(2);
  @$pb.TagNumber(3)
  set daemonName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDaemonName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDaemonName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get acpRuntime => $_getSZ(3);
  @$pb.TagNumber(4)
  set acpRuntime($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAcpRuntime() => $_has(3);
  @$pb.TagNumber(4)
  void clearAcpRuntime() => $_clearField(4);

  /// When the task was dispatched to the daemon.
  @$pb.TagNumber(5)
  $0.Timestamp get startedAt => $_getN(4);
  @$pb.TagNumber(5)
  set startedAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStartedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearStartedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureStartedAt() => $_ensure(4);

  /// Wallclock elapsed since started_at.
  @$pb.TagNumber(6)
  $1.Duration get elapsed => $_getN(5);
  @$pb.TagNumber(6)
  set elapsed($1.Duration value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasElapsed() => $_has(5);
  @$pb.TagNumber(6)
  void clearElapsed() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Duration ensureElapsed() => $_ensure(5);

  /// Latest progress label reported by the daemon, if any.
  @$pb.TagNumber(7)
  $core.String get currentStep => $_getSZ(6);
  @$pb.TagNumber(7)
  set currentStep($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCurrentStep() => $_has(6);
  @$pb.TagNumber(7)
  void clearCurrentStep() => $_clearField(7);

  /// Latest progress percentage in [0, 100]; 0 when unknown.
  @$pb.TagNumber(8)
  $core.int get progress => $_getIZ(7);
  @$pb.TagNumber(8)
  set progress($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasProgress() => $_has(7);
  @$pb.TagNumber(8)
  void clearProgress() => $_clearField(8);

  /// Legacy name of the agent that triggered the dispatch.
  @$pb.TagNumber(9)
  $core.String get agentName => $_getSZ(8);
  @$pb.TagNumber(9)
  set agentName($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasAgentName() => $_has(8);
  @$pb.TagNumber(9)
  void clearAgentName() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get workspaceId => $_getSZ(9);
  @$pb.TagNumber(10)
  set workspaceId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasWorkspaceId() => $_has(9);
  @$pb.TagNumber(10)
  void clearWorkspaceId() => $_clearField(10);

  /// Immutable agent_id of the agent that triggered the dispatch, when known.
  @$pb.TagNumber(11)
  $core.String get agentId => $_getSZ(10);
  @$pb.TagNumber(11)
  set agentId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasAgentId() => $_has(10);
  @$pb.TagNumber(11)
  void clearAgentId() => $_clearField(11);
}

class DaemonStatus extends $pb.GeneratedMessage {
  factory DaemonStatus({
    $core.String? daemonRuntimeId,
    $core.String? name,
    $core.Iterable<$core.String>? acpRuntimes,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? labels,
    DaemonStatus_State? state,
    $0.Timestamp? connectedAt,
    $1.Duration? uptime,
    $core.int? activeTasks,
    $core.String? version,
    $core.String? os,
    $core.Iterable<$core.String>? executors,
    $core.String? remoteAddr,
    $core.String? workspaceId,
  }) {
    final result = DaemonStatus._();
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    if (name != null) result.name = name;
    if (acpRuntimes != null) result.acpRuntimes.addAll(acpRuntimes);
    if (labels != null) result.labels.addEntries(labels);
    if (state != null) result.state = state;
    if (connectedAt != null) result.connectedAt = connectedAt;
    if (uptime != null) result.uptime = uptime;
    if (activeTasks != null) result.activeTasks = activeTasks;
    if (version != null) result.version = version;
    if (os != null) result.os = os;
    if (executors != null) result.executors.addAll(executors);
    if (remoteAddr != null) result.remoteAddr = remoteAddr;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  DaemonStatus._();

  factory DaemonStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonStatus()..mergeFromBuffer(data, registry);
  factory DaemonStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DaemonStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DaemonStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DaemonStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..pPS(3, _omitFieldNames ? '' : 'acpRuntimes')
    ..m<$core.String, $core.String>(4, _omitFieldNames ? '' : 'labels',
        entryClassName: 'DaemonStatus.LabelsEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aE<DaemonStatus_State>(5, _omitFieldNames ? '' : 'state',
        enumValues: DaemonStatus_State.values)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'connectedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$1.Duration>(7, _omitFieldNames ? '' : 'uptime',
        subBuilder: $1.Duration.$_createMessage)
    ..aI(8, _omitFieldNames ? '' : 'activeTasks')
    ..aOS(9, _omitFieldNames ? '' : 'version')
    ..aOS(10, _omitFieldNames ? '' : 'os')
    ..pPS(11, _omitFieldNames ? '' : 'executors')
    ..aOS(12, _omitFieldNames ? '' : 'remoteAddr')
    ..aOS(13, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DaemonStatus copyWith(void Function(DaemonStatus) updates) =>
      super.copyWith((message) => updates(message as DaemonStatus))
          as DaemonStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DaemonStatus() / DaemonStatus.new instead')
  static DaemonStatus create() => DaemonStatus._();
  static $pb.GeneratedMessage $_createMessage() => DaemonStatus._();
  @$core.override
  DaemonStatus createEmptyInstance() => DaemonStatus._();
  @$core.pragma('dart2js:noInline')
  static DaemonStatus getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DaemonStatus>(
          DaemonStatus.$_createMessage);
  static DaemonStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get daemonRuntimeId => $_getSZ(0);
  @$pb.TagNumber(1)
  set daemonRuntimeId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDaemonRuntimeId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaemonRuntimeId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get acpRuntimes => $_getList(2);

  @$pb.TagNumber(4)
  $pb.PbMap<$core.String, $core.String> get labels => $_getMap(3);

  @$pb.TagNumber(5)
  DaemonStatus_State get state => $_getN(4);
  @$pb.TagNumber(5)
  set state(DaemonStatus_State value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get connectedAt => $_getN(5);
  @$pb.TagNumber(6)
  set connectedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasConnectedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearConnectedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureConnectedAt() => $_ensure(5);

  @$pb.TagNumber(7)
  $1.Duration get uptime => $_getN(6);
  @$pb.TagNumber(7)
  set uptime($1.Duration value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasUptime() => $_has(6);
  @$pb.TagNumber(7)
  void clearUptime() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Duration ensureUptime() => $_ensure(6);

  @$pb.TagNumber(8)
  $core.int get activeTasks => $_getIZ(7);
  @$pb.TagNumber(8)
  set activeTasks($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasActiveTasks() => $_has(7);
  @$pb.TagNumber(8)
  void clearActiveTasks() => $_clearField(8);

  /// Reported by the daemon at handshake.
  @$pb.TagNumber(9)
  $core.String get version => $_getSZ(8);
  @$pb.TagNumber(9)
  set version($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasVersion() => $_has(8);
  @$pb.TagNumber(9)
  void clearVersion() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get os => $_getSZ(9);
  @$pb.TagNumber(10)
  set os($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasOs() => $_has(9);
  @$pb.TagNumber(10)
  void clearOs() => $_clearField(10);

  @$pb.TagNumber(11)
  $pb.PbList<$core.String> get executors => $_getList(10);

  /// Captured by the server from the gRPC peer (best-effort, may be empty).
  @$pb.TagNumber(12)
  $core.String get remoteAddr => $_getSZ(11);
  @$pb.TagNumber(12)
  set remoteAddr($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasRemoteAddr() => $_has(11);
  @$pb.TagNumber(12)
  void clearRemoteAddr() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.String get workspaceId => $_getSZ(12);
  @$pb.TagNumber(13)
  set workspaceId($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasWorkspaceId() => $_has(12);
  @$pb.TagNumber(13)
  void clearWorkspaceId() => $_clearField(13);
}

class GetMCPServerStatusRequest extends $pb.GeneratedMessage {
  factory GetMCPServerStatusRequest({
    $core.String? id,
  }) {
    final result = GetMCPServerStatusRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetMCPServerStatusRequest._();

  factory GetMCPServerStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerStatusRequest()..mergeFromBuffer(data, registry);
  factory GetMCPServerStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMCPServerStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMCPServerStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerStatusRequest copyWith(
          void Function(GetMCPServerStatusRequest) updates) =>
      super.copyWith((message) => updates(message as GetMCPServerStatusRequest))
          as GetMCPServerStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMCPServerStatusRequest() / GetMCPServerStatusRequest.new instead')
  static GetMCPServerStatusRequest create() => GetMCPServerStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetMCPServerStatusRequest._();
  @$core.override
  GetMCPServerStatusRequest createEmptyInstance() =>
      GetMCPServerStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetMCPServerStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMCPServerStatusRequest>(
          GetMCPServerStatusRequest.$_createMessage);
  static GetMCPServerStatusRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetMCPServerStatusResponse extends $pb.GeneratedMessage {
  factory GetMCPServerStatusResponse({
    MCPServerStatus? status,
  }) {
    final result = GetMCPServerStatusResponse._();
    if (status != null) result.status = status;
    return result;
  }

  GetMCPServerStatusResponse._();

  factory GetMCPServerStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerStatusResponse()..mergeFromBuffer(data, registry);
  factory GetMCPServerStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetMCPServerStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMCPServerStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetMCPServerStatusResponse.$_createMessage)
    ..aOM<MCPServerStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: MCPServerStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMCPServerStatusResponse copyWith(
          void Function(GetMCPServerStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetMCPServerStatusResponse))
          as GetMCPServerStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetMCPServerStatusResponse() / GetMCPServerStatusResponse.new instead')
  static GetMCPServerStatusResponse create() => GetMCPServerStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetMCPServerStatusResponse._();
  @$core.override
  GetMCPServerStatusResponse createEmptyInstance() =>
      GetMCPServerStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetMCPServerStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMCPServerStatusResponse>(
          GetMCPServerStatusResponse.$_createMessage);
  static GetMCPServerStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  MCPServerStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(MCPServerStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  MCPServerStatus ensureStatus() => $_ensure(0);
}

class GetRemoteAgentStatusRequest extends $pb.GeneratedMessage {
  factory GetRemoteAgentStatusRequest({
    $core.String? id,
  }) {
    final result = GetRemoteAgentStatusRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetRemoteAgentStatusRequest._();

  factory GetRemoteAgentStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentStatusRequest()..mergeFromBuffer(data, registry);
  factory GetRemoteAgentStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRemoteAgentStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRemoteAgentStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentStatusRequest copyWith(
          void Function(GetRemoteAgentStatusRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetRemoteAgentStatusRequest))
          as GetRemoteAgentStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRemoteAgentStatusRequest() / GetRemoteAgentStatusRequest.new instead')
  static GetRemoteAgentStatusRequest create() =>
      GetRemoteAgentStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRemoteAgentStatusRequest._();
  @$core.override
  GetRemoteAgentStatusRequest createEmptyInstance() =>
      GetRemoteAgentStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetRemoteAgentStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRemoteAgentStatusRequest>(
          GetRemoteAgentStatusRequest.$_createMessage);
  static GetRemoteAgentStatusRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetRemoteAgentStatusResponse extends $pb.GeneratedMessage {
  factory GetRemoteAgentStatusResponse({
    RemoteAgentStatus? status,
  }) {
    final result = GetRemoteAgentStatusResponse._();
    if (status != null) result.status = status;
    return result;
  }

  GetRemoteAgentStatusResponse._();

  factory GetRemoteAgentStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentStatusResponse()..mergeFromBuffer(data, registry);
  factory GetRemoteAgentStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRemoteAgentStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRemoteAgentStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRemoteAgentStatusResponse.$_createMessage)
    ..aOM<RemoteAgentStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: RemoteAgentStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRemoteAgentStatusResponse copyWith(
          void Function(GetRemoteAgentStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetRemoteAgentStatusResponse))
          as GetRemoteAgentStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRemoteAgentStatusResponse() / GetRemoteAgentStatusResponse.new instead')
  static GetRemoteAgentStatusResponse create() =>
      GetRemoteAgentStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRemoteAgentStatusResponse._();
  @$core.override
  GetRemoteAgentStatusResponse createEmptyInstance() =>
      GetRemoteAgentStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetRemoteAgentStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRemoteAgentStatusResponse>(
          GetRemoteAgentStatusResponse.$_createMessage);
  static GetRemoteAgentStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  RemoteAgentStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(RemoteAgentStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  RemoteAgentStatus ensureStatus() => $_ensure(0);
}

class RemoteAgentStatus extends $pb.GeneratedMessage {
  factory RemoteAgentStatus({
    $core.String? id,
    $3.RemoteAgentProtocol? protocol,
    RemoteAgentStatus_State? state,
    $core.String? detail,
    $core.String? servingDaemonRuntimeId,
    $0.Timestamp? checkedAt,
    $fixnum.Int64? latencyMs,
  }) {
    final result = RemoteAgentStatus._();
    if (id != null) result.id = id;
    if (protocol != null) result.protocol = protocol;
    if (state != null) result.state = state;
    if (detail != null) result.detail = detail;
    if (servingDaemonRuntimeId != null)
      result.servingDaemonRuntimeId = servingDaemonRuntimeId;
    if (checkedAt != null) result.checkedAt = checkedAt;
    if (latencyMs != null) result.latencyMs = latencyMs;
    return result;
  }

  RemoteAgentStatus._();

  factory RemoteAgentStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoteAgentStatus()..mergeFromBuffer(data, registry);
  factory RemoteAgentStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoteAgentStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RemoteAgentStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RemoteAgentStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aE<$3.RemoteAgentProtocol>(2, _omitFieldNames ? '' : 'protocol',
        enumValues: $3.RemoteAgentProtocol.values)
    ..aE<RemoteAgentStatus_State>(3, _omitFieldNames ? '' : 'state',
        enumValues: RemoteAgentStatus_State.values)
    ..aOS(4, _omitFieldNames ? '' : 'detail')
    ..aOS(5, _omitFieldNames ? '' : 'servingDaemonRuntimeId')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'checkedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aInt64(7, _omitFieldNames ? '' : 'latencyMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoteAgentStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoteAgentStatus copyWith(void Function(RemoteAgentStatus) updates) =>
      super.copyWith((message) => updates(message as RemoteAgentStatus))
          as RemoteAgentStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RemoteAgentStatus() / RemoteAgentStatus.new instead')
  static RemoteAgentStatus create() => RemoteAgentStatus._();
  static $pb.GeneratedMessage $_createMessage() => RemoteAgentStatus._();
  @$core.override
  RemoteAgentStatus createEmptyInstance() => RemoteAgentStatus._();
  @$core.pragma('dart2js:noInline')
  static RemoteAgentStatus getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RemoteAgentStatus>(
          RemoteAgentStatus.$_createMessage);
  static RemoteAgentStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $3.RemoteAgentProtocol get protocol => $_getN(1);
  @$pb.TagNumber(2)
  set protocol($3.RemoteAgentProtocol value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasProtocol() => $_has(1);
  @$pb.TagNumber(2)
  void clearProtocol() => $_clearField(2);

  @$pb.TagNumber(3)
  RemoteAgentStatus_State get state => $_getN(2);
  @$pb.TagNumber(3)
  set state(RemoteAgentStatus_State value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get detail => $_getSZ(3);
  @$pb.TagNumber(4)
  set detail($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDetail() => $_has(3);
  @$pb.TagNumber(4)
  void clearDetail() => $_clearField(4);

  /// For DAEMON protocol only: id of the runtime currently serving the
  /// configured daemon_runtime_id, if one is online.
  @$pb.TagNumber(5)
  $core.String get servingDaemonRuntimeId => $_getSZ(4);
  @$pb.TagNumber(5)
  set servingDaemonRuntimeId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasServingDaemonRuntimeId() => $_has(4);
  @$pb.TagNumber(5)
  void clearServingDaemonRuntimeId() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get checkedAt => $_getN(5);
  @$pb.TagNumber(6)
  set checkedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCheckedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCheckedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureCheckedAt() => $_ensure(5);

  @$pb.TagNumber(7)
  $fixnum.Int64 get latencyMs => $_getI64(6);
  @$pb.TagNumber(7)
  set latencyMs($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasLatencyMs() => $_has(6);
  @$pb.TagNumber(7)
  void clearLatencyMs() => $_clearField(7);
}

class ListMCPToolsRequest extends $pb.GeneratedMessage {
  factory ListMCPToolsRequest({
    $core.String? serverId,
  }) {
    final result = ListMCPToolsRequest._();
    if (serverId != null) result.serverId = serverId;
    return result;
  }

  ListMCPToolsRequest._();

  factory ListMCPToolsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPToolsRequest()..mergeFromBuffer(data, registry);
  factory ListMCPToolsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPToolsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMCPToolsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListMCPToolsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'serverId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPToolsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPToolsRequest copyWith(void Function(ListMCPToolsRequest) updates) =>
      super.copyWith((message) => updates(message as ListMCPToolsRequest))
          as ListMCPToolsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListMCPToolsRequest() / ListMCPToolsRequest.new instead')
  static ListMCPToolsRequest create() => ListMCPToolsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListMCPToolsRequest._();
  @$core.override
  ListMCPToolsRequest createEmptyInstance() => ListMCPToolsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListMCPToolsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMCPToolsRequest>(
          ListMCPToolsRequest.$_createMessage);
  static ListMCPToolsRequest? _defaultInstance;

  /// Optional: when set, list tools for this server only.
  @$pb.TagNumber(1)
  $core.String get serverId => $_getSZ(0);
  @$pb.TagNumber(1)
  set serverId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasServerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearServerId() => $_clearField(1);
}

class ListMCPToolsResponse extends $pb.GeneratedMessage {
  factory ListMCPToolsResponse({
    $core.Iterable<MCPTool>? tools,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? errors,
  }) {
    final result = ListMCPToolsResponse._();
    if (tools != null) result.tools.addAll(tools);
    if (errors != null) result.errors.addEntries(errors);
    return result;
  }

  ListMCPToolsResponse._();

  factory ListMCPToolsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPToolsResponse()..mergeFromBuffer(data, registry);
  factory ListMCPToolsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListMCPToolsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListMCPToolsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListMCPToolsResponse.$_createMessage)
    ..pPM<MCPTool>(1, _omitFieldNames ? '' : 'tools',
        subBuilder: MCPTool.$_createMessage)
    ..m<$core.String, $core.String>(2, _omitFieldNames ? '' : 'errors',
        entryClassName: 'ListMCPToolsResponse.ErrorsEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPToolsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListMCPToolsResponse copyWith(void Function(ListMCPToolsResponse) updates) =>
      super.copyWith((message) => updates(message as ListMCPToolsResponse))
          as ListMCPToolsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListMCPToolsResponse() / ListMCPToolsResponse.new instead')
  static ListMCPToolsResponse create() => ListMCPToolsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListMCPToolsResponse._();
  @$core.override
  ListMCPToolsResponse createEmptyInstance() => ListMCPToolsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListMCPToolsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListMCPToolsResponse>(
          ListMCPToolsResponse.$_createMessage);
  static ListMCPToolsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<MCPTool> get tools => $_getList(0);

  /// Servers that failed probing, keyed by id → error.
  @$pb.TagNumber(2)
  $pb.PbMap<$core.String, $core.String> get errors => $_getMap(1);
}

class MCPTool extends $pb.GeneratedMessage {
  factory MCPTool({
    $core.String? name,
    $core.String? description,
    $core.String? serverId,
    $core.String? serverName,
    $core.bool? allowed,
  }) {
    final result = MCPTool._();
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (serverId != null) result.serverId = serverId;
    if (serverName != null) result.serverName = serverName;
    if (allowed != null) result.allowed = allowed;
    return result;
  }

  MCPTool._();

  factory MCPTool.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPTool()..mergeFromBuffer(data, registry);
  factory MCPTool.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPTool()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MCPTool',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MCPTool.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..aOS(3, _omitFieldNames ? '' : 'serverId')
    ..aOS(4, _omitFieldNames ? '' : 'serverName')
    ..aOB(5, _omitFieldNames ? '' : 'allowed')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPTool clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPTool copyWith(void Function(MCPTool) updates) =>
      super.copyWith((message) => updates(message as MCPTool)) as MCPTool;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MCPTool() / MCPTool.new instead')
  static MCPTool create() => MCPTool._();
  static $pb.GeneratedMessage $_createMessage() => MCPTool._();
  @$core.override
  MCPTool createEmptyInstance() => MCPTool._();
  @$core.pragma('dart2js:noInline')
  static MCPTool getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MCPTool>(MCPTool.$_createMessage);
  static MCPTool? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get serverId => $_getSZ(2);
  @$pb.TagNumber(3)
  set serverId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasServerId() => $_has(2);
  @$pb.TagNumber(3)
  void clearServerId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get serverName => $_getSZ(3);
  @$pb.TagNumber(4)
  set serverName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasServerName() => $_has(3);
  @$pb.TagNumber(4)
  void clearServerName() => $_clearField(4);

  /// True if not filtered out by the server's tool_filter whitelist.
  @$pb.TagNumber(5)
  $core.bool get allowed => $_getBF(4);
  @$pb.TagNumber(5)
  set allowed($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAllowed() => $_has(4);
  @$pb.TagNumber(5)
  void clearAllowed() => $_clearField(5);
}

class MCPServerStatus extends $pb.GeneratedMessage {
  factory MCPServerStatus({
    $core.String? id,
    $core.String? name,
    MCPServerStatus_State? state,
    $core.int? toolCount,
    $core.String? detail,
    $0.Timestamp? checkedAt,
  }) {
    final result = MCPServerStatus._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (state != null) result.state = state;
    if (toolCount != null) result.toolCount = toolCount;
    if (detail != null) result.detail = detail;
    if (checkedAt != null) result.checkedAt = checkedAt;
    return result;
  }

  MCPServerStatus._();

  factory MCPServerStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServerStatus()..mergeFromBuffer(data, registry);
  factory MCPServerStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServerStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MCPServerStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MCPServerStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aE<MCPServerStatus_State>(3, _omitFieldNames ? '' : 'state',
        enumValues: MCPServerStatus_State.values)
    ..aI(4, _omitFieldNames ? '' : 'toolCount')
    ..aOS(5, _omitFieldNames ? '' : 'detail')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'checkedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServerStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServerStatus copyWith(void Function(MCPServerStatus) updates) =>
      super.copyWith((message) => updates(message as MCPServerStatus))
          as MCPServerStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MCPServerStatus() / MCPServerStatus.new instead')
  static MCPServerStatus create() => MCPServerStatus._();
  static $pb.GeneratedMessage $_createMessage() => MCPServerStatus._();
  @$core.override
  MCPServerStatus createEmptyInstance() => MCPServerStatus._();
  @$core.pragma('dart2js:noInline')
  static MCPServerStatus getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MCPServerStatus>(
          MCPServerStatus.$_createMessage);
  static MCPServerStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  MCPServerStatus_State get state => $_getN(2);
  @$pb.TagNumber(3)
  set state(MCPServerStatus_State value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => $_clearField(3);

  /// tool_count is taken from the configured tool_filter when present;
  /// a live probe count will be added when MCP probing is implemented.
  @$pb.TagNumber(4)
  $core.int get toolCount => $_getIZ(3);
  @$pb.TagNumber(4)
  set toolCount($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasToolCount() => $_has(3);
  @$pb.TagNumber(4)
  void clearToolCount() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get detail => $_getSZ(4);
  @$pb.TagNumber(5)
  set detail($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDetail() => $_has(4);
  @$pb.TagNumber(5)
  void clearDetail() => $_clearField(5);

  @$pb.TagNumber(6)
  $0.Timestamp get checkedAt => $_getN(5);
  @$pb.TagNumber(6)
  set checkedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCheckedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCheckedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureCheckedAt() => $_ensure(5);
}

class GetChannelStatusRequest extends $pb.GeneratedMessage {
  factory GetChannelStatusRequest({
    $core.String? name,
  }) {
    final result = GetChannelStatusRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetChannelStatusRequest._();

  factory GetChannelStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelStatusRequest()..mergeFromBuffer(data, registry);
  factory GetChannelStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetChannelStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetChannelStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelStatusRequest copyWith(
          void Function(GetChannelStatusRequest) updates) =>
      super.copyWith((message) => updates(message as GetChannelStatusRequest))
          as GetChannelStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetChannelStatusRequest() / GetChannelStatusRequest.new instead')
  static GetChannelStatusRequest create() => GetChannelStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetChannelStatusRequest._();
  @$core.override
  GetChannelStatusRequest createEmptyInstance() => GetChannelStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetChannelStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetChannelStatusRequest>(
          GetChannelStatusRequest.$_createMessage);
  static GetChannelStatusRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetChannelStatusResponse extends $pb.GeneratedMessage {
  factory GetChannelStatusResponse({
    ChannelStatus? status,
  }) {
    final result = GetChannelStatusResponse._();
    if (status != null) result.status = status;
    return result;
  }

  GetChannelStatusResponse._();

  factory GetChannelStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelStatusResponse()..mergeFromBuffer(data, registry);
  factory GetChannelStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetChannelStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetChannelStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetChannelStatusResponse.$_createMessage)
    ..aOM<ChannelStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: ChannelStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetChannelStatusResponse copyWith(
          void Function(GetChannelStatusResponse) updates) =>
      super.copyWith((message) => updates(message as GetChannelStatusResponse))
          as GetChannelStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetChannelStatusResponse() / GetChannelStatusResponse.new instead')
  static GetChannelStatusResponse create() => GetChannelStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetChannelStatusResponse._();
  @$core.override
  GetChannelStatusResponse createEmptyInstance() =>
      GetChannelStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetChannelStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetChannelStatusResponse>(
          GetChannelStatusResponse.$_createMessage);
  static GetChannelStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ChannelStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(ChannelStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  ChannelStatus ensureStatus() => $_ensure(0);
}

class ChannelStatus extends $pb.GeneratedMessage {
  factory ChannelStatus({
    $core.String? name,
    $4.AgentChannelPlatform? platform,
    ChannelStatus_State? state,
    $0.Timestamp? lastPollAt,
    $core.String? detail,
  }) {
    final result = ChannelStatus._();
    if (name != null) result.name = name;
    if (platform != null) result.platform = platform;
    if (state != null) result.state = state;
    if (lastPollAt != null) result.lastPollAt = lastPollAt;
    if (detail != null) result.detail = detail;
    return result;
  }

  ChannelStatus._();

  factory ChannelStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChannelStatus()..mergeFromBuffer(data, registry);
  factory ChannelStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChannelStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChannelStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ChannelStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aE<$4.AgentChannelPlatform>(2, _omitFieldNames ? '' : 'platform',
        enumValues: $4.AgentChannelPlatform.values)
    ..aE<ChannelStatus_State>(3, _omitFieldNames ? '' : 'state',
        enumValues: ChannelStatus_State.values)
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'lastPollAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'detail')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChannelStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChannelStatus copyWith(void Function(ChannelStatus) updates) =>
      super.copyWith((message) => updates(message as ChannelStatus))
          as ChannelStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ChannelStatus() / ChannelStatus.new instead')
  static ChannelStatus create() => ChannelStatus._();
  static $pb.GeneratedMessage $_createMessage() => ChannelStatus._();
  @$core.override
  ChannelStatus createEmptyInstance() => ChannelStatus._();
  @$core.pragma('dart2js:noInline')
  static ChannelStatus getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ChannelStatus>(
          ChannelStatus.$_createMessage);
  static ChannelStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $4.AgentChannelPlatform get platform => $_getN(1);
  @$pb.TagNumber(2)
  set platform($4.AgentChannelPlatform value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPlatform() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlatform() => $_clearField(2);

  @$pb.TagNumber(3)
  ChannelStatus_State get state => $_getN(2);
  @$pb.TagNumber(3)
  set state(ChannelStatus_State value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Timestamp get lastPollAt => $_getN(3);
  @$pb.TagNumber(4)
  set lastPollAt($0.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasLastPollAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearLastPollAt() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Timestamp ensureLastPollAt() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get detail => $_getSZ(4);
  @$pb.TagNumber(5)
  set detail($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDetail() => $_has(4);
  @$pb.TagNumber(5)
  void clearDetail() => $_clearField(5);
}

class DashboardServiceApi {
  final $pb.RpcClient _client;

  DashboardServiceApi(this._client);

  /// GetOverview returns aggregate counts, component health and the latest
  /// daemon handshake for the dashboard Overview screen.
  $async.Future<GetOverviewResponse> getOverview(
          $pb.ClientContext? ctx, GetOverviewRequest request) =>
      _client.invoke<GetOverviewResponse>(ctx, 'DashboardService',
          'GetOverview', request, GetOverviewResponse());

  /// GetActivityFeed returns the most recent invocation activity, oldest event
  /// first stripped, suitable for the dashboard Activity Feed widget.
  $async.Future<GetActivityFeedResponse> getActivityFeed(
          $pb.ClientContext? ctx, GetActivityFeedRequest request) =>
      _client.invoke<GetActivityFeedResponse>(ctx, 'DashboardService',
          'GetActivityFeed', request, GetActivityFeedResponse());

  /// GetCronExecutionTimeseries aggregates cron_executions into time buckets,
  /// suitable for the Overview screen "Cron Executions" chart.
  $async.Future<GetCronExecutionTimeseriesResponse> getCronExecutionTimeseries(
          $pb.ClientContext? ctx, GetCronExecutionTimeseriesRequest request) =>
      _client.invoke<GetCronExecutionTimeseriesResponse>(
          ctx,
          'DashboardService',
          'GetCronExecutionTimeseries',
          request,
          GetCronExecutionTimeseriesResponse());

  /// GetActivityMetrics returns run-volume counts over a rolling time window,
  /// suitable for the Overview screen "Activity" metric cards (7d / 30d toggle).
  $async.Future<GetActivityMetricsResponse> getActivityMetrics(
          $pb.ClientContext? ctx, GetActivityMetricsRequest request) =>
      _client.invoke<GetActivityMetricsResponse>(ctx, 'DashboardService',
          'GetActivityMetrics', request, GetActivityMetricsResponse());
}

class DaemonServiceApi {
  final $pb.RpcClient _client;

  DaemonServiceApi(this._client);

  /// Workspace-scoped daemon runtime configuration. A runtime must exist before
  /// a worker credential can be issued or a runtime connection accepted.
  $async.Future<ListDaemonRuntimesResponse> listDaemonRuntimes(
          $pb.ClientContext? ctx, ListDaemonRuntimesRequest request) =>
      _client.invoke<ListDaemonRuntimesResponse>(ctx, 'DaemonService',
          'ListDaemonRuntimes', request, ListDaemonRuntimesResponse());
  $async.Future<GetDaemonRuntimeResponse> getDaemonRuntime(
          $pb.ClientContext? ctx, GetDaemonRuntimeRequest request) =>
      _client.invoke<GetDaemonRuntimeResponse>(ctx, 'DaemonService',
          'GetDaemonRuntime', request, GetDaemonRuntimeResponse());
  $async.Future<CreateDaemonRuntimeResponse> createDaemonRuntime(
          $pb.ClientContext? ctx, CreateDaemonRuntimeRequest request) =>
      _client.invoke<CreateDaemonRuntimeResponse>(ctx, 'DaemonService',
          'CreateDaemonRuntime', request, CreateDaemonRuntimeResponse());
  $async.Future<UpdateDaemonRuntimeResponse> updateDaemonRuntime(
          $pb.ClientContext? ctx, UpdateDaemonRuntimeRequest request) =>
      _client.invoke<UpdateDaemonRuntimeResponse>(ctx, 'DaemonService',
          'UpdateDaemonRuntime', request, UpdateDaemonRuntimeResponse());
  $async.Future<DeleteDaemonRuntimeResponse> deleteDaemonRuntime(
          $pb.ClientContext? ctx, DeleteDaemonRuntimeRequest request) =>
      _client.invoke<DeleteDaemonRuntimeResponse>(ctx, 'DaemonService',
          'DeleteDaemonRuntime', request, DeleteDaemonRuntimeResponse());
  $async.Future<CreateDaemonRuntimeTokenResponse> createDaemonRuntimeToken(
          $pb.ClientContext? ctx, CreateDaemonRuntimeTokenRequest request) =>
      _client.invoke<CreateDaemonRuntimeTokenResponse>(
          ctx,
          'DaemonService',
          'CreateDaemonRuntimeToken',
          request,
          CreateDaemonRuntimeTokenResponse());
  $async.Future<ListDaemonsResponse> listDaemons(
          $pb.ClientContext? ctx, ListDaemonsRequest request) =>
      _client.invoke<ListDaemonsResponse>(
          ctx, 'DaemonService', 'ListDaemons', request, ListDaemonsResponse());
  $async.Future<GetDaemonResponse> getDaemon(
          $pb.ClientContext? ctx, GetDaemonRequest request) =>
      _client.invoke<GetDaemonResponse>(
          ctx, 'DaemonService', 'GetDaemon', request, GetDaemonResponse());

  /// CancelDaemonTask requests cancellation of a running task on any connected
  /// daemon. Returns NotFound if no online daemon is tracking the task.
  $async.Future<CancelDaemonTaskResponse> cancelDaemonTask(
          $pb.ClientContext? ctx, CancelDaemonTaskRequest request) =>
      _client.invoke<CancelDaemonTaskResponse>(ctx, 'DaemonService',
          'CancelDaemonTask', request, CancelDaemonTaskResponse());

  /// ListDaemonTasks returns all tasks currently in flight on connected daemons.
  $async.Future<ListDaemonTasksResponse> listDaemonTasks(
          $pb.ClientContext? ctx, ListDaemonTasksRequest request) =>
      _client.invoke<ListDaemonTasksResponse>(ctx, 'DaemonService',
          'ListDaemonTasks', request, ListDaemonTasksResponse());

  /// GetBridgeDiagnostics returns router-side resource usage and a recent
  /// bridge-latency sample series, suitable for the Bridge Diagnostics panel.
  $async.Future<GetBridgeDiagnosticsResponse> getBridgeDiagnostics(
          $pb.ClientContext? ctx, GetBridgeDiagnosticsRequest request) =>
      _client.invoke<GetBridgeDiagnosticsResponse>(ctx, 'DaemonService',
          'GetBridgeDiagnostics', request, GetBridgeDiagnosticsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
