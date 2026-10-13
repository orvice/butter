// This is a generated file - do not edit.
//
// Generated from agents/v1/dashboard.proto.

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

import 'package:protobuf/well_known_types/google/protobuf/duration.pbjson.dart'
    as $1;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pbjson.dart'
    as $0;

import 'api_token.pbjson.dart' as $2;

@$core.Deprecated('Use getActivityMetricsRequestDescriptor instead')
const GetActivityMetricsRequest$json = {
  '1': 'GetActivityMetricsRequest',
  '2': [
    {
      '1': 'range',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.GetActivityMetricsRequest.Range',
      '10': 'range'
    },
  ],
  '4': [GetActivityMetricsRequest_Range$json],
};

@$core.Deprecated('Use getActivityMetricsRequestDescriptor instead')
const GetActivityMetricsRequest_Range$json = {
  '1': 'Range',
  '2': [
    {'1': 'RANGE_UNSPECIFIED', '2': 0},
    {'1': 'RANGE_1D', '2': 1},
    {'1': 'RANGE_7D', '2': 2},
    {'1': 'RANGE_30D', '2': 3},
  ],
};

/// Descriptor for `GetActivityMetricsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getActivityMetricsRequestDescriptor = $convert.base64Decode(
    'ChlHZXRBY3Rpdml0eU1ldHJpY3NSZXF1ZXN0EkAKBXJhbmdlGAEgASgOMiouYWdlbnRzLnYxLk'
    'dldEFjdGl2aXR5TWV0cmljc1JlcXVlc3QuUmFuZ2VSBXJhbmdlIkkKBVJhbmdlEhUKEVJBTkdF'
    'X1VOU1BFQ0lGSUVEEAASDAoIUkFOR0VfMUQQARIMCghSQU5HRV83RBACEg0KCVJBTkdFXzMwRB'
    'AD');

@$core.Deprecated('Use getActivityMetricsResponseDescriptor instead')
const GetActivityMetricsResponse$json = {
  '1': 'GetActivityMetricsResponse',
  '2': [
    {'1': 'agent_runs', '3': 1, '4': 1, '5': 5, '10': 'agentRuns'},
    {'1': 'agent_runs_failed', '3': 2, '4': 1, '5': 5, '10': 'agentRunsFailed'},
    {'1': 'automation_runs', '3': 3, '4': 1, '5': 5, '10': 'automationRuns'},
    {
      '1': 'automation_runs_failed',
      '3': 4,
      '4': 1,
      '5': 5,
      '10': 'automationRunsFailed'
    },
    {
      '1': 'window_start',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'windowStart'
    },
    {
      '1': 'window_end',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'windowEnd'
    },
  ],
};

/// Descriptor for `GetActivityMetricsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getActivityMetricsResponseDescriptor = $convert.base64Decode(
    'ChpHZXRBY3Rpdml0eU1ldHJpY3NSZXNwb25zZRIdCgphZ2VudF9ydW5zGAEgASgFUglhZ2VudF'
    'J1bnMSKgoRYWdlbnRfcnVuc19mYWlsZWQYAiABKAVSD2FnZW50UnVuc0ZhaWxlZBInCg9hdXRv'
    'bWF0aW9uX3J1bnMYAyABKAVSDmF1dG9tYXRpb25SdW5zEjQKFmF1dG9tYXRpb25fcnVuc19mYW'
    'lsZWQYBCABKAVSFGF1dG9tYXRpb25SdW5zRmFpbGVkEj0KDHdpbmRvd19zdGFydBgFIAEoCzIa'
    'Lmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSC3dpbmRvd1N0YXJ0EjkKCndpbmRvd19lbmQYBi'
    'ABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl3aW5kb3dFbmQ=');

@$core.Deprecated('Use getCronExecutionTimeseriesRequestDescriptor instead')
const GetCronExecutionTimeseriesRequest$json = {
  '1': 'GetCronExecutionTimeseriesRequest',
  '2': [
    {
      '1': 'range',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.GetCronExecutionTimeseriesRequest.Range',
      '10': 'range'
    },
    {'1': 'job_name', '3': 2, '4': 1, '5': 9, '10': 'jobName'},
  ],
  '4': [GetCronExecutionTimeseriesRequest_Range$json],
};

@$core.Deprecated('Use getCronExecutionTimeseriesRequestDescriptor instead')
const GetCronExecutionTimeseriesRequest_Range$json = {
  '1': 'Range',
  '2': [
    {'1': 'RANGE_UNSPECIFIED', '2': 0},
    {'1': 'RANGE_1D', '2': 1},
    {'1': 'RANGE_7D', '2': 2},
    {'1': 'RANGE_30D', '2': 3},
  ],
};

/// Descriptor for `GetCronExecutionTimeseriesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCronExecutionTimeseriesRequestDescriptor =
    $convert.base64Decode(
        'CiFHZXRDcm9uRXhlY3V0aW9uVGltZXNlcmllc1JlcXVlc3QSSAoFcmFuZ2UYASABKA4yMi5hZ2'
        'VudHMudjEuR2V0Q3JvbkV4ZWN1dGlvblRpbWVzZXJpZXNSZXF1ZXN0LlJhbmdlUgVyYW5nZRIZ'
        'Cghqb2JfbmFtZRgCIAEoCVIHam9iTmFtZSJJCgVSYW5nZRIVChFSQU5HRV9VTlNQRUNJRklFRB'
        'AAEgwKCFJBTkdFXzFEEAESDAoIUkFOR0VfN0QQAhINCglSQU5HRV8zMEQQAw==');

@$core.Deprecated('Use getCronExecutionTimeseriesResponseDescriptor instead')
const GetCronExecutionTimeseriesResponse$json = {
  '1': 'GetCronExecutionTimeseriesResponse',
  '2': [
    {
      '1': 'buckets',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.CronExecutionBucket',
      '10': 'buckets'
    },
  ],
};

/// Descriptor for `GetCronExecutionTimeseriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCronExecutionTimeseriesResponseDescriptor =
    $convert.base64Decode(
        'CiJHZXRDcm9uRXhlY3V0aW9uVGltZXNlcmllc1Jlc3BvbnNlEjgKB2J1Y2tldHMYASADKAsyHi'
        '5hZ2VudHMudjEuQ3JvbkV4ZWN1dGlvbkJ1Y2tldFIHYnVja2V0cw==');

@$core.Deprecated('Use cronExecutionBucketDescriptor instead')
const CronExecutionBucket$json = {
  '1': 'CronExecutionBucket',
  '2': [
    {
      '1': 'start',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'start'
    },
    {'1': 'success', '3': 2, '4': 1, '5': 5, '10': 'success'},
    {'1': 'error', '3': 3, '4': 1, '5': 5, '10': 'error'},
  ],
};

/// Descriptor for `CronExecutionBucket`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cronExecutionBucketDescriptor = $convert.base64Decode(
    'ChNDcm9uRXhlY3V0aW9uQnVja2V0EjAKBXN0YXJ0GAEgASgLMhouZ29vZ2xlLnByb3RvYnVmLl'
    'RpbWVzdGFtcFIFc3RhcnQSGAoHc3VjY2VzcxgCIAEoBVIHc3VjY2VzcxIUCgVlcnJvchgDIAEo'
    'BVIFZXJyb3I=');

@$core.Deprecated('Use getActivityFeedRequestDescriptor instead')
const GetActivityFeedRequest$json = {
  '1': 'GetActivityFeedRequest',
  '2': [
    {'1': 'limit', '3': 1, '4': 1, '5': 5, '10': 'limit'},
    {'1': 'page_token', '3': 2, '4': 1, '5': 9, '10': 'pageToken'},
  ],
};

/// Descriptor for `GetActivityFeedRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getActivityFeedRequestDescriptor =
    $convert.base64Decode(
        'ChZHZXRBY3Rpdml0eUZlZWRSZXF1ZXN0EhQKBWxpbWl0GAEgASgFUgVsaW1pdBIdCgpwYWdlX3'
        'Rva2VuGAIgASgJUglwYWdlVG9rZW4=');

@$core.Deprecated('Use getActivityFeedResponseDescriptor instead')
const GetActivityFeedResponse$json = {
  '1': 'GetActivityFeedResponse',
  '2': [
    {
      '1': 'events',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ActivityEvent',
      '10': 'events'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `GetActivityFeedResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getActivityFeedResponseDescriptor = $convert.base64Decode(
    'ChdHZXRBY3Rpdml0eUZlZWRSZXNwb25zZRIwCgZldmVudHMYASADKAsyGC5hZ2VudHMudjEuQW'
    'N0aXZpdHlFdmVudFIGZXZlbnRzEiYKD25leHRfcGFnZV90b2tlbhgCIAEoCVINbmV4dFBhZ2VU'
    'b2tlbg==');

@$core.Deprecated('Use activityEventDescriptor instead')
const ActivityEvent$json = {
  '1': 'ActivityEvent',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'kind', '3': 2, '4': 1, '5': 9, '10': 'kind'},
    {'1': 'actor', '3': 3, '4': 1, '5': 9, '10': 'actor'},
    {'1': 'message', '3': 4, '4': 1, '5': 9, '10': 'message'},
    {
      '1': 'timestamp',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'timestamp'
    },
  ],
};

/// Descriptor for `ActivityEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List activityEventDescriptor = $convert.base64Decode(
    'Cg1BY3Rpdml0eUV2ZW50Eg4KAmlkGAEgASgJUgJpZBISCgRraW5kGAIgASgJUgRraW5kEhQKBW'
    'FjdG9yGAMgASgJUgVhY3RvchIYCgdtZXNzYWdlGAQgASgJUgdtZXNzYWdlEjgKCXRpbWVzdGFt'
    'cBgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXRpbWVzdGFtcA==');

@$core.Deprecated('Use getOverviewRequestDescriptor instead')
const GetOverviewRequest$json = {
  '1': 'GetOverviewRequest',
  '2': [
    {'1': 'environment', '3': 1, '4': 1, '5': 9, '10': 'environment'},
  ],
};

/// Descriptor for `GetOverviewRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getOverviewRequestDescriptor = $convert.base64Decode(
    'ChJHZXRPdmVydmlld1JlcXVlc3QSIAoLZW52aXJvbm1lbnQYASABKAlSC2Vudmlyb25tZW50');

@$core.Deprecated('Use getOverviewResponseDescriptor instead')
const GetOverviewResponse$json = {
  '1': 'GetOverviewResponse',
  '2': [
    {
      '1': 'counts',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.OverviewCounts',
      '10': 'counts'
    },
    {
      '1': 'health',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.HealthSummary',
      '10': 'health'
    },
    {
      '1': 'latest_daemon_handshake',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonHandshake',
      '10': 'latestDaemonHandshake'
    },
  ],
};

/// Descriptor for `GetOverviewResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getOverviewResponseDescriptor = $convert.base64Decode(
    'ChNHZXRPdmVydmlld1Jlc3BvbnNlEjEKBmNvdW50cxgBIAEoCzIZLmFnZW50cy52MS5PdmVydm'
    'lld0NvdW50c1IGY291bnRzEjAKBmhlYWx0aBgCIAEoCzIYLmFnZW50cy52MS5IZWFsdGhTdW1t'
    'YXJ5UgZoZWFsdGgSUgoXbGF0ZXN0X2RhZW1vbl9oYW5kc2hha2UYAyABKAsyGi5hZ2VudHMudj'
    'EuRGFlbW9uSGFuZHNoYWtlUhVsYXRlc3REYWVtb25IYW5kc2hha2U=');

@$core.Deprecated('Use overviewCountsDescriptor instead')
const OverviewCounts$json = {
  '1': 'OverviewCounts',
  '2': [
    {'1': 'active_agents', '3': 1, '4': 1, '5': 5, '10': 'activeAgents'},
    {'1': 'mcp_servers', '3': 2, '4': 1, '5': 5, '10': 'mcpServers'},
    {
      '1': 'connected_daemons',
      '3': 3,
      '4': 1,
      '5': 5,
      '10': 'connectedDaemons'
    },
    {'1': 'remote_agents', '3': 4, '4': 1, '5': 5, '10': 'remoteAgents'},
    {'1': 'channels', '3': 5, '4': 1, '5': 5, '10': 'channels'},
    {'1': 'cron_jobs', '3': 6, '4': 1, '5': 5, '10': 'cronJobs'},
    {'1': 'active_sessions', '3': 7, '4': 1, '5': 5, '10': 'activeSessions'},
  ],
};

/// Descriptor for `OverviewCounts`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List overviewCountsDescriptor = $convert.base64Decode(
    'Cg5PdmVydmlld0NvdW50cxIjCg1hY3RpdmVfYWdlbnRzGAEgASgFUgxhY3RpdmVBZ2VudHMSHw'
    'oLbWNwX3NlcnZlcnMYAiABKAVSCm1jcFNlcnZlcnMSKwoRY29ubmVjdGVkX2RhZW1vbnMYAyAB'
    'KAVSEGNvbm5lY3RlZERhZW1vbnMSIwoNcmVtb3RlX2FnZW50cxgEIAEoBVIMcmVtb3RlQWdlbn'
    'RzEhoKCGNoYW5uZWxzGAUgASgFUghjaGFubmVscxIbCgljcm9uX2pvYnMYBiABKAVSCGNyb25K'
    'b2JzEicKD2FjdGl2ZV9zZXNzaW9ucxgHIAEoBVIOYWN0aXZlU2Vzc2lvbnM=');

@$core.Deprecated('Use healthSummaryDescriptor instead')
const HealthSummary$json = {
  '1': 'HealthSummary',
  '2': [
    {
      '1': 'mongodb',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ComponentHealth',
      '10': 'mongodb'
    },
    {
      '1': 'redis',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ComponentHealth',
      '10': 'redis'
    },
    {
      '1': 'runner',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ComponentHealth',
      '10': 'runner'
    },
  ],
};

/// Descriptor for `HealthSummary`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List healthSummaryDescriptor = $convert.base64Decode(
    'Cg1IZWFsdGhTdW1tYXJ5EjQKB21vbmdvZGIYASABKAsyGi5hZ2VudHMudjEuQ29tcG9uZW50SG'
    'VhbHRoUgdtb25nb2RiEjAKBXJlZGlzGAIgASgLMhouYWdlbnRzLnYxLkNvbXBvbmVudEhlYWx0'
    'aFIFcmVkaXMSMgoGcnVubmVyGAMgASgLMhouYWdlbnRzLnYxLkNvbXBvbmVudEhlYWx0aFIGcn'
    'VubmVy');

@$core.Deprecated('Use componentHealthDescriptor instead')
const ComponentHealth$json = {
  '1': 'ComponentHealth',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.ComponentHealth.Status',
      '10': 'status'
    },
    {'1': 'detail', '3': 2, '4': 1, '5': 9, '10': 'detail'},
    {
      '1': 'checked_at',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'checkedAt'
    },
    {'1': 'latency_ms', '3': 4, '4': 1, '5': 3, '10': 'latencyMs'},
  ],
  '4': [ComponentHealth_Status$json],
};

@$core.Deprecated('Use componentHealthDescriptor instead')
const ComponentHealth_Status$json = {
  '1': 'Status',
  '2': [
    {'1': 'STATUS_UNSPECIFIED', '2': 0},
    {'1': 'STATUS_HEALTHY', '2': 1},
    {'1': 'STATUS_DEGRADED', '2': 2},
    {'1': 'STATUS_DOWN', '2': 3},
  ],
};

/// Descriptor for `ComponentHealth`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List componentHealthDescriptor = $convert.base64Decode(
    'Cg9Db21wb25lbnRIZWFsdGgSOQoGc3RhdHVzGAEgASgOMiEuYWdlbnRzLnYxLkNvbXBvbmVudE'
    'hlYWx0aC5TdGF0dXNSBnN0YXR1cxIWCgZkZXRhaWwYAiABKAlSBmRldGFpbBI5CgpjaGVja2Vk'
    'X2F0GAMgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJY2hlY2tlZEF0Eh0KCmxhdG'
    'VuY3lfbXMYBCABKANSCWxhdGVuY3lNcyJaCgZTdGF0dXMSFgoSU1RBVFVTX1VOU1BFQ0lGSUVE'
    'EAASEgoOU1RBVFVTX0hFQUxUSFkQARITCg9TVEFUVVNfREVHUkFERUQQAhIPCgtTVEFUVVNfRE'
    '9XThAD');

@$core.Deprecated('Use daemonHandshakeDescriptor instead')
const DaemonHandshake$json = {
  '1': 'DaemonHandshake',
  '2': [
    {'1': 'daemon_runtime_id', '3': 1, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'acp_runtimes', '3': 3, '4': 3, '5': 9, '10': 'acpRuntimes'},
    {
      '1': 'connected_at',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'connectedAt'
    },
    {'1': 'os', '3': 5, '4': 1, '5': 9, '10': 'os'},
  ],
};

/// Descriptor for `DaemonHandshake`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List daemonHandshakeDescriptor = $convert.base64Decode(
    'Cg9EYWVtb25IYW5kc2hha2USKgoRZGFlbW9uX3J1bnRpbWVfaWQYASABKAlSD2RhZW1vblJ1bn'
    'RpbWVJZBISCgRuYW1lGAIgASgJUgRuYW1lEiEKDGFjcF9ydW50aW1lcxgDIAMoCVILYWNwUnVu'
    'dGltZXMSPQoMY29ubmVjdGVkX2F0GAQgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcF'
    'ILY29ubmVjdGVkQXQSDgoCb3MYBSABKAlSAm9z');

@$core.Deprecated('Use getBridgeDiagnosticsRequestDescriptor instead')
const GetBridgeDiagnosticsRequest$json = {
  '1': 'GetBridgeDiagnosticsRequest',
  '2': [
    {
      '1': 'window',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'window'
    },
  ],
};

/// Descriptor for `GetBridgeDiagnosticsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBridgeDiagnosticsRequestDescriptor =
    $convert.base64Decode(
        'ChtHZXRCcmlkZ2VEaWFnbm9zdGljc1JlcXVlc3QSMQoGd2luZG93GAEgASgLMhkuZ29vZ2xlLn'
        'Byb3RvYnVmLkR1cmF0aW9uUgZ3aW5kb3c=');

@$core.Deprecated('Use getBridgeDiagnosticsResponseDescriptor instead')
const GetBridgeDiagnosticsResponse$json = {
  '1': 'GetBridgeDiagnosticsResponse',
  '2': [
    {
      '1': 'diagnostics',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.BridgeDiagnostics',
      '10': 'diagnostics'
    },
  ],
};

/// Descriptor for `GetBridgeDiagnosticsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBridgeDiagnosticsResponseDescriptor =
    $convert.base64Decode(
        'ChxHZXRCcmlkZ2VEaWFnbm9zdGljc1Jlc3BvbnNlEj4KC2RpYWdub3N0aWNzGAEgASgLMhwuYW'
        'dlbnRzLnYxLkJyaWRnZURpYWdub3N0aWNzUgtkaWFnbm9zdGljcw==');

@$core.Deprecated('Use bridgeDiagnosticsDescriptor instead')
const BridgeDiagnostics$json = {
  '1': 'BridgeDiagnostics',
  '2': [
    {'1': 'cpu_percent', '3': 1, '4': 1, '5': 1, '10': 'cpuPercent'},
    {'1': 'memory_used_bytes', '3': 2, '4': 1, '5': 3, '10': 'memoryUsedBytes'},
    {
      '1': 'memory_limit_bytes',
      '3': 3,
      '4': 1,
      '5': 3,
      '10': 'memoryLimitBytes'
    },
    {'1': 'goroutines', '3': 4, '4': 1, '5': 5, '10': 'goroutines'},
    {
      '1': 'checked_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'checkedAt'
    },
    {
      '1': 'latency',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.LatencyPoint',
      '10': 'latency'
    },
  ],
};

/// Descriptor for `BridgeDiagnostics`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List bridgeDiagnosticsDescriptor = $convert.base64Decode(
    'ChFCcmlkZ2VEaWFnbm9zdGljcxIfCgtjcHVfcGVyY2VudBgBIAEoAVIKY3B1UGVyY2VudBIqCh'
    'FtZW1vcnlfdXNlZF9ieXRlcxgCIAEoA1IPbWVtb3J5VXNlZEJ5dGVzEiwKEm1lbW9yeV9saW1p'
    'dF9ieXRlcxgDIAEoA1IQbWVtb3J5TGltaXRCeXRlcxIeCgpnb3JvdXRpbmVzGAQgASgFUgpnb3'
    'JvdXRpbmVzEjkKCmNoZWNrZWRfYXQYBSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1w'
    'UgljaGVja2VkQXQSMQoHbGF0ZW5jeRgGIAMoCzIXLmFnZW50cy52MS5MYXRlbmN5UG9pbnRSB2'
    'xhdGVuY3k=');

@$core.Deprecated('Use latencyPointDescriptor instead')
const LatencyPoint$json = {
  '1': 'LatencyPoint',
  '2': [
    {
      '1': 'timestamp',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'timestamp'
    },
    {'1': 'latency_ms', '3': 2, '4': 1, '5': 3, '10': 'latencyMs'},
  ],
};

/// Descriptor for `LatencyPoint`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List latencyPointDescriptor = $convert.base64Decode(
    'CgxMYXRlbmN5UG9pbnQSOAoJdGltZXN0YW1wGAEgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbW'
    'VzdGFtcFIJdGltZXN0YW1wEh0KCmxhdGVuY3lfbXMYAiABKANSCWxhdGVuY3lNcw==');

@$core.Deprecated('Use daemonRuntimeDescriptor instead')
const DaemonRuntime$json = {
  '1': 'DaemonRuntime',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {
      '1': 'labels',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime.LabelsEntry',
      '10': 'labels'
    },
    {
      '1': 'created_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {'1': 'created_by', '3': 7, '4': 1, '5': 9, '10': 'createdBy'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [DaemonRuntime_LabelsEntry$json],
  '9': [
    {'1': 4, '2': 5},
  ],
  '10': ['allowed_capabilities'],
};

@$core.Deprecated('Use daemonRuntimeDescriptor instead')
const DaemonRuntime_LabelsEntry$json = {
  '1': 'LabelsEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `DaemonRuntime`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List daemonRuntimeDescriptor = $convert.base64Decode(
    'Cg1EYWVtb25SdW50aW1lEg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEiAKC2'
    'Rlc2NyaXB0aW9uGAMgASgJUgtkZXNjcmlwdGlvbhI8CgZsYWJlbHMYBSADKAsyJC5hZ2VudHMu'
    'djEuRGFlbW9uUnVudGltZS5MYWJlbHNFbnRyeVIGbGFiZWxzEjkKCmNyZWF0ZWRfYXQYBiABKA'
    'syGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgljcmVhdGVkQXQSHQoKY3JlYXRlZF9ieRgH'
    'IAEoCVIJY3JlYXRlZEJ5EiEKDHdvcmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYWNlSWQaOQoLTG'
    'FiZWxzRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AUoE'
    'CAQQBVIUYWxsb3dlZF9jYXBhYmlsaXRpZXM=');

@$core.Deprecated('Use listDaemonRuntimesRequestDescriptor instead')
const ListDaemonRuntimesRequest$json = {
  '1': 'ListDaemonRuntimesRequest',
};

/// Descriptor for `ListDaemonRuntimesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDaemonRuntimesRequestDescriptor =
    $convert.base64Decode('ChlMaXN0RGFlbW9uUnVudGltZXNSZXF1ZXN0');

@$core.Deprecated('Use listDaemonRuntimesResponseDescriptor instead')
const ListDaemonRuntimesResponse$json = {
  '1': 'ListDaemonRuntimesResponse',
  '2': [
    {
      '1': 'runtimes',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime',
      '10': 'runtimes'
    },
  ],
};

/// Descriptor for `ListDaemonRuntimesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDaemonRuntimesResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0RGFlbW9uUnVudGltZXNSZXNwb25zZRI0CghydW50aW1lcxgBIAMoCzIYLmFnZW50cy'
        '52MS5EYWVtb25SdW50aW1lUghydW50aW1lcw==');

@$core.Deprecated('Use getDaemonRuntimeRequestDescriptor instead')
const GetDaemonRuntimeRequest$json = {
  '1': 'GetDaemonRuntimeRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetDaemonRuntimeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDaemonRuntimeRequestDescriptor = $convert
    .base64Decode('ChdHZXREYWVtb25SdW50aW1lUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use getDaemonRuntimeResponseDescriptor instead')
const GetDaemonRuntimeResponse$json = {
  '1': 'GetDaemonRuntimeResponse',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime',
      '10': 'runtime'
    },
  ],
};

/// Descriptor for `GetDaemonRuntimeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDaemonRuntimeResponseDescriptor =
    $convert.base64Decode(
        'ChhHZXREYWVtb25SdW50aW1lUmVzcG9uc2USMgoHcnVudGltZRgBIAEoCzIYLmFnZW50cy52MS'
        '5EYWVtb25SdW50aW1lUgdydW50aW1l');

@$core.Deprecated('Use createDaemonRuntimeRequestDescriptor instead')
const CreateDaemonRuntimeRequest$json = {
  '1': 'CreateDaemonRuntimeRequest',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime',
      '10': 'runtime'
    },
  ],
};

/// Descriptor for `CreateDaemonRuntimeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDaemonRuntimeRequestDescriptor =
    $convert.base64Decode(
        'ChpDcmVhdGVEYWVtb25SdW50aW1lUmVxdWVzdBIyCgdydW50aW1lGAEgASgLMhguYWdlbnRzLn'
        'YxLkRhZW1vblJ1bnRpbWVSB3J1bnRpbWU=');

@$core.Deprecated('Use createDaemonRuntimeResponseDescriptor instead')
const CreateDaemonRuntimeResponse$json = {
  '1': 'CreateDaemonRuntimeResponse',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime',
      '10': 'runtime'
    },
  ],
};

/// Descriptor for `CreateDaemonRuntimeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDaemonRuntimeResponseDescriptor =
    $convert.base64Decode(
        'ChtDcmVhdGVEYWVtb25SdW50aW1lUmVzcG9uc2USMgoHcnVudGltZRgBIAEoCzIYLmFnZW50cy'
        '52MS5EYWVtb25SdW50aW1lUgdydW50aW1l');

@$core.Deprecated('Use updateDaemonRuntimeRequestDescriptor instead')
const UpdateDaemonRuntimeRequest$json = {
  '1': 'UpdateDaemonRuntimeRequest',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime',
      '10': 'runtime'
    },
  ],
};

/// Descriptor for `UpdateDaemonRuntimeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateDaemonRuntimeRequestDescriptor =
    $convert.base64Decode(
        'ChpVcGRhdGVEYWVtb25SdW50aW1lUmVxdWVzdBIyCgdydW50aW1lGAEgASgLMhguYWdlbnRzLn'
        'YxLkRhZW1vblJ1bnRpbWVSB3J1bnRpbWU=');

@$core.Deprecated('Use updateDaemonRuntimeResponseDescriptor instead')
const UpdateDaemonRuntimeResponse$json = {
  '1': 'UpdateDaemonRuntimeResponse',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonRuntime',
      '10': 'runtime'
    },
  ],
};

/// Descriptor for `UpdateDaemonRuntimeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateDaemonRuntimeResponseDescriptor =
    $convert.base64Decode(
        'ChtVcGRhdGVEYWVtb25SdW50aW1lUmVzcG9uc2USMgoHcnVudGltZRgBIAEoCzIYLmFnZW50cy'
        '52MS5EYWVtb25SdW50aW1lUgdydW50aW1l');

@$core.Deprecated('Use deleteDaemonRuntimeRequestDescriptor instead')
const DeleteDaemonRuntimeRequest$json = {
  '1': 'DeleteDaemonRuntimeRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteDaemonRuntimeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteDaemonRuntimeRequestDescriptor =
    $convert.base64Decode(
        'ChpEZWxldGVEYWVtb25SdW50aW1lUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use deleteDaemonRuntimeResponseDescriptor instead')
const DeleteDaemonRuntimeResponse$json = {
  '1': 'DeleteDaemonRuntimeResponse',
};

/// Descriptor for `DeleteDaemonRuntimeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteDaemonRuntimeResponseDescriptor =
    $convert.base64Decode('ChtEZWxldGVEYWVtb25SdW50aW1lUmVzcG9uc2U=');

@$core.Deprecated('Use createDaemonRuntimeTokenRequestDescriptor instead')
const CreateDaemonRuntimeTokenRequest$json = {
  '1': 'CreateDaemonRuntimeTokenRequest',
  '2': [
    {'1': 'daemon_runtime_id', '3': 1, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'ttl',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'ttl'
    },
  ],
};

/// Descriptor for `CreateDaemonRuntimeTokenRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDaemonRuntimeTokenRequestDescriptor =
    $convert.base64Decode(
        'Ch9DcmVhdGVEYWVtb25SdW50aW1lVG9rZW5SZXF1ZXN0EioKEWRhZW1vbl9ydW50aW1lX2lkGA'
        'EgASgJUg9kYWVtb25SdW50aW1lSWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIrCgN0dGwYAyABKAsy'
        'GS5nb29nbGUucHJvdG9idWYuRHVyYXRpb25SA3R0bA==');

@$core.Deprecated('Use createDaemonRuntimeTokenResponseDescriptor instead')
const CreateDaemonRuntimeTokenResponse$json = {
  '1': 'CreateDaemonRuntimeTokenResponse',
  '2': [
    {
      '1': 'token',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.APIToken',
      '10': 'token'
    },
    {'1': 'secret', '3': 2, '4': 1, '5': 9, '10': 'secret'},
  ],
};

/// Descriptor for `CreateDaemonRuntimeTokenResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDaemonRuntimeTokenResponseDescriptor =
    $convert.base64Decode(
        'CiBDcmVhdGVEYWVtb25SdW50aW1lVG9rZW5SZXNwb25zZRIpCgV0b2tlbhgBIAEoCzITLmFnZW'
        '50cy52MS5BUElUb2tlblIFdG9rZW4SFgoGc2VjcmV0GAIgASgJUgZzZWNyZXQ=');

@$core.Deprecated('Use listDaemonsRequestDescriptor instead')
const ListDaemonsRequest$json = {
  '1': 'ListDaemonsRequest',
};

/// Descriptor for `ListDaemonsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDaemonsRequestDescriptor =
    $convert.base64Decode('ChJMaXN0RGFlbW9uc1JlcXVlc3Q=');

@$core.Deprecated('Use listDaemonsResponseDescriptor instead')
const ListDaemonsResponse$json = {
  '1': 'ListDaemonsResponse',
  '2': [
    {
      '1': 'daemons',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.DaemonStatus',
      '10': 'daemons'
    },
  ],
};

/// Descriptor for `ListDaemonsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDaemonsResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0RGFlbW9uc1Jlc3BvbnNlEjEKB2RhZW1vbnMYASADKAsyFy5hZ2VudHMudjEuRGFlbW'
    '9uU3RhdHVzUgdkYWVtb25z');

@$core.Deprecated('Use getDaemonRequestDescriptor instead')
const GetDaemonRequest$json = {
  '1': 'GetDaemonRequest',
  '2': [
    {'1': 'daemon_runtime_id', '3': 1, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
  ],
};

/// Descriptor for `GetDaemonRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDaemonRequestDescriptor = $convert.base64Decode(
    'ChBHZXREYWVtb25SZXF1ZXN0EioKEWRhZW1vbl9ydW50aW1lX2lkGAEgASgJUg9kYWVtb25SdW'
    '50aW1lSWQ=');

@$core.Deprecated('Use getDaemonResponseDescriptor instead')
const GetDaemonResponse$json = {
  '1': 'GetDaemonResponse',
  '2': [
    {
      '1': 'daemon',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DaemonStatus',
      '10': 'daemon'
    },
  ],
};

/// Descriptor for `GetDaemonResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDaemonResponseDescriptor = $convert.base64Decode(
    'ChFHZXREYWVtb25SZXNwb25zZRIvCgZkYWVtb24YASABKAsyFy5hZ2VudHMudjEuRGFlbW9uU3'
    'RhdHVzUgZkYWVtb24=');

@$core.Deprecated('Use cancelDaemonTaskRequestDescriptor instead')
const CancelDaemonTaskRequest$json = {
  '1': 'CancelDaemonTaskRequest',
  '2': [
    {'1': 'task_id', '3': 1, '4': 1, '5': 9, '10': 'taskId'},
    {'1': 'daemon_runtime_id', '3': 2, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
  ],
};

/// Descriptor for `CancelDaemonTaskRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelDaemonTaskRequestDescriptor =
    $convert.base64Decode(
        'ChdDYW5jZWxEYWVtb25UYXNrUmVxdWVzdBIXCgd0YXNrX2lkGAEgASgJUgZ0YXNrSWQSKgoRZG'
        'FlbW9uX3J1bnRpbWVfaWQYAiABKAlSD2RhZW1vblJ1bnRpbWVJZA==');

@$core.Deprecated('Use cancelDaemonTaskResponseDescriptor instead')
const CancelDaemonTaskResponse$json = {
  '1': 'CancelDaemonTaskResponse',
  '2': [
    {'1': 'daemon_runtime_id', '3': 1, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
  ],
};

/// Descriptor for `CancelDaemonTaskResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelDaemonTaskResponseDescriptor =
    $convert.base64Decode(
        'ChhDYW5jZWxEYWVtb25UYXNrUmVzcG9uc2USKgoRZGFlbW9uX3J1bnRpbWVfaWQYASABKAlSD2'
        'RhZW1vblJ1bnRpbWVJZA==');

@$core.Deprecated('Use listDaemonTasksRequestDescriptor instead')
const ListDaemonTasksRequest$json = {
  '1': 'ListDaemonTasksRequest',
  '2': [
    {'1': 'daemon_runtime_id', '3': 1, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
  ],
};

/// Descriptor for `ListDaemonTasksRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDaemonTasksRequestDescriptor =
    $convert.base64Decode(
        'ChZMaXN0RGFlbW9uVGFza3NSZXF1ZXN0EioKEWRhZW1vbl9ydW50aW1lX2lkGAEgASgJUg9kYW'
        'Vtb25SdW50aW1lSWQ=');

@$core.Deprecated('Use listDaemonTasksResponseDescriptor instead')
const ListDaemonTasksResponse$json = {
  '1': 'ListDaemonTasksResponse',
  '2': [
    {
      '1': 'tasks',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.DaemonTaskInFlight',
      '10': 'tasks'
    },
  ],
};

/// Descriptor for `ListDaemonTasksResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDaemonTasksResponseDescriptor =
    $convert.base64Decode(
        'ChdMaXN0RGFlbW9uVGFza3NSZXNwb25zZRIzCgV0YXNrcxgBIAMoCzIdLmFnZW50cy52MS5EYW'
        'Vtb25UYXNrSW5GbGlnaHRSBXRhc2tz');

@$core.Deprecated('Use daemonTaskInFlightDescriptor instead')
const DaemonTaskInFlight$json = {
  '1': 'DaemonTaskInFlight',
  '2': [
    {'1': 'task_id', '3': 1, '4': 1, '5': 9, '10': 'taskId'},
    {'1': 'daemon_runtime_id', '3': 2, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
    {'1': 'daemon_name', '3': 3, '4': 1, '5': 9, '10': 'daemonName'},
    {'1': 'acp_runtime', '3': 4, '4': 1, '5': 9, '10': 'acpRuntime'},
    {
      '1': 'started_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startedAt'
    },
    {
      '1': 'elapsed',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'elapsed'
    },
    {'1': 'current_step', '3': 7, '4': 1, '5': 9, '10': 'currentStep'},
    {'1': 'progress', '3': 8, '4': 1, '5': 5, '10': 'progress'},
    {'1': 'agent_name', '3': 9, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'workspace_id', '3': 10, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'agent_id', '3': 11, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `DaemonTaskInFlight`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List daemonTaskInFlightDescriptor = $convert.base64Decode(
    'ChJEYWVtb25UYXNrSW5GbGlnaHQSFwoHdGFza19pZBgBIAEoCVIGdGFza0lkEioKEWRhZW1vbl'
    '9ydW50aW1lX2lkGAIgASgJUg9kYWVtb25SdW50aW1lSWQSHwoLZGFlbW9uX25hbWUYAyABKAlS'
    'CmRhZW1vbk5hbWUSHwoLYWNwX3J1bnRpbWUYBCABKAlSCmFjcFJ1bnRpbWUSOQoKc3RhcnRlZF'
    '9hdBgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXN0YXJ0ZWRBdBIzCgdlbGFw'
    'c2VkGAYgASgLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9uUgdlbGFwc2VkEiEKDGN1cnJlbn'
    'Rfc3RlcBgHIAEoCVILY3VycmVudFN0ZXASGgoIcHJvZ3Jlc3MYCCABKAVSCHByb2dyZXNzEh0K'
    'CmFnZW50X25hbWUYCSABKAlSCWFnZW50TmFtZRIhCgx3b3Jrc3BhY2VfaWQYCiABKAlSC3dvcm'
    'tzcGFjZUlkEhkKCGFnZW50X2lkGAsgASgJUgdhZ2VudElk');

@$core.Deprecated('Use daemonStatusDescriptor instead')
const DaemonStatus$json = {
  '1': 'DaemonStatus',
  '2': [
    {'1': 'daemon_runtime_id', '3': 1, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'acp_runtimes', '3': 3, '4': 3, '5': 9, '10': 'acpRuntimes'},
    {
      '1': 'labels',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.DaemonStatus.LabelsEntry',
      '10': 'labels'
    },
    {
      '1': 'state',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.DaemonStatus.State',
      '10': 'state'
    },
    {
      '1': 'connected_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'connectedAt'
    },
    {
      '1': 'uptime',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'uptime'
    },
    {'1': 'active_tasks', '3': 8, '4': 1, '5': 5, '10': 'activeTasks'},
    {'1': 'version', '3': 9, '4': 1, '5': 9, '10': 'version'},
    {'1': 'os', '3': 10, '4': 1, '5': 9, '10': 'os'},
    {'1': 'executors', '3': 11, '4': 3, '5': 9, '10': 'executors'},
    {'1': 'remote_addr', '3': 12, '4': 1, '5': 9, '10': 'remoteAddr'},
    {'1': 'workspace_id', '3': 13, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [DaemonStatus_LabelsEntry$json],
  '4': [DaemonStatus_State$json],
};

@$core.Deprecated('Use daemonStatusDescriptor instead')
const DaemonStatus_LabelsEntry$json = {
  '1': 'LabelsEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

@$core.Deprecated('Use daemonStatusDescriptor instead')
const DaemonStatus_State$json = {
  '1': 'State',
  '2': [
    {'1': 'STATE_UNSPECIFIED', '2': 0},
    {'1': 'STATE_ONLINE', '2': 1},
    {'1': 'STATE_IDLE', '2': 2},
    {'1': 'STATE_OFFLINE', '2': 3},
  ],
};

/// Descriptor for `DaemonStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List daemonStatusDescriptor = $convert.base64Decode(
    'CgxEYWVtb25TdGF0dXMSKgoRZGFlbW9uX3J1bnRpbWVfaWQYASABKAlSD2RhZW1vblJ1bnRpbW'
    'VJZBISCgRuYW1lGAIgASgJUgRuYW1lEiEKDGFjcF9ydW50aW1lcxgDIAMoCVILYWNwUnVudGlt'
    'ZXMSOwoGbGFiZWxzGAQgAygLMiMuYWdlbnRzLnYxLkRhZW1vblN0YXR1cy5MYWJlbHNFbnRyeV'
    'IGbGFiZWxzEjMKBXN0YXRlGAUgASgOMh0uYWdlbnRzLnYxLkRhZW1vblN0YXR1cy5TdGF0ZVIF'
    'c3RhdGUSPQoMY29ubmVjdGVkX2F0GAYgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcF'
    'ILY29ubmVjdGVkQXQSMQoGdXB0aW1lGAcgASgLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9u'
    'UgZ1cHRpbWUSIQoMYWN0aXZlX3Rhc2tzGAggASgFUgthY3RpdmVUYXNrcxIYCgd2ZXJzaW9uGA'
    'kgASgJUgd2ZXJzaW9uEg4KAm9zGAogASgJUgJvcxIcCglleGVjdXRvcnMYCyADKAlSCWV4ZWN1'
    'dG9ycxIfCgtyZW1vdGVfYWRkchgMIAEoCVIKcmVtb3RlQWRkchIhCgx3b3Jrc3BhY2VfaWQYDS'
    'ABKAlSC3dvcmtzcGFjZUlkGjkKC0xhYmVsc0VudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZh'
    'bHVlGAIgASgJUgV2YWx1ZToCOAEiUwoFU3RhdGUSFQoRU1RBVEVfVU5TUEVDSUZJRUQQABIQCg'
    'xTVEFURV9PTkxJTkUQARIOCgpTVEFURV9JRExFEAISEQoNU1RBVEVfT0ZGTElORRAD');

@$core.Deprecated('Use getMCPServerStatusRequestDescriptor instead')
const GetMCPServerStatusRequest$json = {
  '1': 'GetMCPServerStatusRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetMCPServerStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMCPServerStatusRequestDescriptor =
    $convert.base64Decode(
        'ChlHZXRNQ1BTZXJ2ZXJTdGF0dXNSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use getMCPServerStatusResponseDescriptor instead')
const GetMCPServerStatusResponse$json = {
  '1': 'GetMCPServerStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServerStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetMCPServerStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMCPServerStatusResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRNQ1BTZXJ2ZXJTdGF0dXNSZXNwb25zZRIyCgZzdGF0dXMYASABKAsyGi5hZ2VudHMudj'
        'EuTUNQU2VydmVyU3RhdHVzUgZzdGF0dXM=');

@$core.Deprecated('Use getRemoteAgentStatusRequestDescriptor instead')
const GetRemoteAgentStatusRequest$json = {
  '1': 'GetRemoteAgentStatusRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetRemoteAgentStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRemoteAgentStatusRequestDescriptor =
    $convert.base64Decode(
        'ChtHZXRSZW1vdGVBZ2VudFN0YXR1c1JlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getRemoteAgentStatusResponseDescriptor instead')
const GetRemoteAgentStatusResponse$json = {
  '1': 'GetRemoteAgentStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RemoteAgentStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetRemoteAgentStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRemoteAgentStatusResponseDescriptor =
    $convert.base64Decode(
        'ChxHZXRSZW1vdGVBZ2VudFN0YXR1c1Jlc3BvbnNlEjQKBnN0YXR1cxgBIAEoCzIcLmFnZW50cy'
        '52MS5SZW1vdGVBZ2VudFN0YXR1c1IGc3RhdHVz');

@$core.Deprecated('Use remoteAgentStatusDescriptor instead')
const RemoteAgentStatus$json = {
  '1': 'RemoteAgentStatus',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'protocol',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RemoteAgentProtocol',
      '10': 'protocol'
    },
    {
      '1': 'state',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RemoteAgentStatus.State',
      '10': 'state'
    },
    {'1': 'detail', '3': 4, '4': 1, '5': 9, '10': 'detail'},
    {
      '1': 'serving_daemon_runtime_id',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'servingDaemonRuntimeId'
    },
    {
      '1': 'checked_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'checkedAt'
    },
    {'1': 'latency_ms', '3': 7, '4': 1, '5': 3, '10': 'latencyMs'},
  ],
  '4': [RemoteAgentStatus_State$json],
};

@$core.Deprecated('Use remoteAgentStatusDescriptor instead')
const RemoteAgentStatus_State$json = {
  '1': 'State',
  '2': [
    {'1': 'STATE_UNSPECIFIED', '2': 0},
    {'1': 'STATE_CONFIGURED', '2': 1},
    {'1': 'STATE_ACTIVE', '2': 2},
    {'1': 'STATE_IDLE', '2': 3},
    {'1': 'STATE_UNREACHABLE', '2': 4},
    {'1': 'STATE_ERROR', '2': 5},
  ],
};

/// Descriptor for `RemoteAgentStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List remoteAgentStatusDescriptor = $convert.base64Decode(
    'ChFSZW1vdGVBZ2VudFN0YXR1cxIOCgJpZBgBIAEoCVICaWQSOgoIcHJvdG9jb2wYAiABKA4yHi'
    '5hZ2VudHMudjEuUmVtb3RlQWdlbnRQcm90b2NvbFIIcHJvdG9jb2wSOAoFc3RhdGUYAyABKA4y'
    'Ii5hZ2VudHMudjEuUmVtb3RlQWdlbnRTdGF0dXMuU3RhdGVSBXN0YXRlEhYKBmRldGFpbBgEIA'
    'EoCVIGZGV0YWlsEjkKGXNlcnZpbmdfZGFlbW9uX3J1bnRpbWVfaWQYBSABKAlSFnNlcnZpbmdE'
    'YWVtb25SdW50aW1lSWQSOQoKY2hlY2tlZF9hdBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW'
    '1lc3RhbXBSCWNoZWNrZWRBdBIdCgpsYXRlbmN5X21zGAcgASgDUglsYXRlbmN5TXMifgoFU3Rh'
    'dGUSFQoRU1RBVEVfVU5TUEVDSUZJRUQQABIUChBTVEFURV9DT05GSUdVUkVEEAESEAoMU1RBVE'
    'VfQUNUSVZFEAISDgoKU1RBVEVfSURMRRADEhUKEVNUQVRFX1VOUkVBQ0hBQkxFEAQSDwoLU1RB'
    'VEVfRVJST1IQBQ==');

@$core.Deprecated('Use listMCPToolsRequestDescriptor instead')
const ListMCPToolsRequest$json = {
  '1': 'ListMCPToolsRequest',
  '2': [
    {'1': 'server_id', '3': 1, '4': 1, '5': 9, '10': 'serverId'},
  ],
};

/// Descriptor for `ListMCPToolsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMCPToolsRequestDescriptor =
    $convert.base64Decode(
        'ChNMaXN0TUNQVG9vbHNSZXF1ZXN0EhsKCXNlcnZlcl9pZBgBIAEoCVIIc2VydmVySWQ=');

@$core.Deprecated('Use listMCPToolsResponseDescriptor instead')
const ListMCPToolsResponse$json = {
  '1': 'ListMCPToolsResponse',
  '2': [
    {
      '1': 'tools',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MCPTool',
      '10': 'tools'
    },
    {
      '1': 'errors',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ListMCPToolsResponse.ErrorsEntry',
      '10': 'errors'
    },
  ],
  '3': [ListMCPToolsResponse_ErrorsEntry$json],
};

@$core.Deprecated('Use listMCPToolsResponseDescriptor instead')
const ListMCPToolsResponse_ErrorsEntry$json = {
  '1': 'ErrorsEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `ListMCPToolsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMCPToolsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0TUNQVG9vbHNSZXNwb25zZRIoCgV0b29scxgBIAMoCzISLmFnZW50cy52MS5NQ1BUb2'
    '9sUgV0b29scxJDCgZlcnJvcnMYAiADKAsyKy5hZ2VudHMudjEuTGlzdE1DUFRvb2xzUmVzcG9u'
    'c2UuRXJyb3JzRW50cnlSBmVycm9ycxo5CgtFcnJvcnNFbnRyeRIQCgNrZXkYASABKAlSA2tleR'
    'IUCgV2YWx1ZRgCIAEoCVIFdmFsdWU6AjgB');

@$core.Deprecated('Use mCPToolDescriptor instead')
const MCPTool$json = {
  '1': 'MCPTool',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    {'1': 'server_id', '3': 3, '4': 1, '5': 9, '10': 'serverId'},
    {'1': 'server_name', '3': 4, '4': 1, '5': 9, '10': 'serverName'},
    {'1': 'allowed', '3': 5, '4': 1, '5': 8, '10': 'allowed'},
  ],
};

/// Descriptor for `MCPTool`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mCPToolDescriptor = $convert.base64Decode(
    'CgdNQ1BUb29sEhIKBG5hbWUYASABKAlSBG5hbWUSIAoLZGVzY3JpcHRpb24YAiABKAlSC2Rlc2'
    'NyaXB0aW9uEhsKCXNlcnZlcl9pZBgDIAEoCVIIc2VydmVySWQSHwoLc2VydmVyX25hbWUYBCAB'
    'KAlSCnNlcnZlck5hbWUSGAoHYWxsb3dlZBgFIAEoCFIHYWxsb3dlZA==');

@$core.Deprecated('Use mCPServerStatusDescriptor instead')
const MCPServerStatus$json = {
  '1': 'MCPServerStatus',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'state',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MCPServerStatus.State',
      '10': 'state'
    },
    {'1': 'tool_count', '3': 4, '4': 1, '5': 5, '10': 'toolCount'},
    {'1': 'detail', '3': 5, '4': 1, '5': 9, '10': 'detail'},
    {
      '1': 'checked_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'checkedAt'
    },
  ],
  '4': [MCPServerStatus_State$json],
};

@$core.Deprecated('Use mCPServerStatusDescriptor instead')
const MCPServerStatus_State$json = {
  '1': 'State',
  '2': [
    {'1': 'STATE_UNSPECIFIED', '2': 0},
    {'1': 'STATE_CONFIGURED', '2': 1},
    {'1': 'STATE_CONNECTED', '2': 2},
    {'1': 'STATE_DISCONNECTED', '2': 3},
    {'1': 'STATE_ERROR', '2': 4},
  ],
};

/// Descriptor for `MCPServerStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mCPServerStatusDescriptor = $convert.base64Decode(
    'Cg9NQ1BTZXJ2ZXJTdGF0dXMSDgoCaWQYASABKAlSAmlkEhIKBG5hbWUYAiABKAlSBG5hbWUSNg'
    'oFc3RhdGUYAyABKA4yIC5hZ2VudHMudjEuTUNQU2VydmVyU3RhdHVzLlN0YXRlUgVzdGF0ZRId'
    'Cgp0b29sX2NvdW50GAQgASgFUgl0b29sQ291bnQSFgoGZGV0YWlsGAUgASgJUgZkZXRhaWwSOQ'
    'oKY2hlY2tlZF9hdBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWNoZWNrZWRB'
    'dCJyCgVTdGF0ZRIVChFTVEFURV9VTlNQRUNJRklFRBAAEhQKEFNUQVRFX0NPTkZJR1VSRUQQAR'
    'ITCg9TVEFURV9DT05ORUNURUQQAhIWChJTVEFURV9ESVNDT05ORUNURUQQAxIPCgtTVEFURV9F'
    'UlJPUhAE');

@$core.Deprecated('Use getChannelStatusRequestDescriptor instead')
const GetChannelStatusRequest$json = {
  '1': 'GetChannelStatusRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `GetChannelStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getChannelStatusRequestDescriptor =
    $convert.base64Decode(
        'ChdHZXRDaGFubmVsU3RhdHVzUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use getChannelStatusResponseDescriptor instead')
const GetChannelStatusResponse$json = {
  '1': 'GetChannelStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ChannelStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetChannelStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getChannelStatusResponseDescriptor =
    $convert.base64Decode(
        'ChhHZXRDaGFubmVsU3RhdHVzUmVzcG9uc2USMAoGc3RhdHVzGAEgASgLMhguYWdlbnRzLnYxLk'
        'NoYW5uZWxTdGF0dXNSBnN0YXR1cw==');

@$core.Deprecated('Use channelStatusDescriptor instead')
const ChannelStatus$json = {
  '1': 'ChannelStatus',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'platform',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentChannelPlatform',
      '10': 'platform'
    },
    {
      '1': 'state',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.ChannelStatus.State',
      '10': 'state'
    },
    {
      '1': 'last_poll_at',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastPollAt'
    },
    {'1': 'detail', '3': 5, '4': 1, '5': 9, '10': 'detail'},
  ],
  '4': [ChannelStatus_State$json],
};

@$core.Deprecated('Use channelStatusDescriptor instead')
const ChannelStatus_State$json = {
  '1': 'State',
  '2': [
    {'1': 'STATE_UNSPECIFIED', '2': 0},
    {'1': 'STATE_LIVE', '2': 1},
    {'1': 'STATE_PAUSED', '2': 2},
    {'1': 'STATE_DISABLED', '2': 3},
    {'1': 'STATE_ERROR', '2': 4},
  ],
};

/// Descriptor for `ChannelStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List channelStatusDescriptor = $convert.base64Decode(
    'Cg1DaGFubmVsU3RhdHVzEhIKBG5hbWUYASABKAlSBG5hbWUSOwoIcGxhdGZvcm0YAiABKA4yHy'
    '5hZ2VudHMudjEuQWdlbnRDaGFubmVsUGxhdGZvcm1SCHBsYXRmb3JtEjQKBXN0YXRlGAMgASgO'
    'Mh4uYWdlbnRzLnYxLkNoYW5uZWxTdGF0dXMuU3RhdGVSBXN0YXRlEjwKDGxhc3RfcG9sbF9hdB'
    'gEIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmxhc3RQb2xsQXQSFgoGZGV0YWls'
    'GAUgASgJUgZkZXRhaWwiZQoFU3RhdGUSFQoRU1RBVEVfVU5TUEVDSUZJRUQQABIOCgpTVEFURV'
    '9MSVZFEAESEAoMU1RBVEVfUEFVU0VEEAISEgoOU1RBVEVfRElTQUJMRUQQAxIPCgtTVEFURV9F'
    'UlJPUhAE');

const $core.Map<$core.String, $core.dynamic> DashboardServiceBase$json = {
  '1': 'DashboardService',
  '2': [
    {
      '1': 'GetOverview',
      '2': '.agents.v1.GetOverviewRequest',
      '3': '.agents.v1.GetOverviewResponse'
    },
    {
      '1': 'GetActivityFeed',
      '2': '.agents.v1.GetActivityFeedRequest',
      '3': '.agents.v1.GetActivityFeedResponse'
    },
    {
      '1': 'GetCronExecutionTimeseries',
      '2': '.agents.v1.GetCronExecutionTimeseriesRequest',
      '3': '.agents.v1.GetCronExecutionTimeseriesResponse'
    },
    {
      '1': 'GetActivityMetrics',
      '2': '.agents.v1.GetActivityMetricsRequest',
      '3': '.agents.v1.GetActivityMetricsResponse'
    },
  ],
};

@$core.Deprecated('Use dashboardServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DashboardServiceBase$messageJson = {
  '.agents.v1.GetOverviewRequest': GetOverviewRequest$json,
  '.agents.v1.GetOverviewResponse': GetOverviewResponse$json,
  '.agents.v1.OverviewCounts': OverviewCounts$json,
  '.agents.v1.HealthSummary': HealthSummary$json,
  '.agents.v1.ComponentHealth': ComponentHealth$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.DaemonHandshake': DaemonHandshake$json,
  '.agents.v1.GetActivityFeedRequest': GetActivityFeedRequest$json,
  '.agents.v1.GetActivityFeedResponse': GetActivityFeedResponse$json,
  '.agents.v1.ActivityEvent': ActivityEvent$json,
  '.agents.v1.GetCronExecutionTimeseriesRequest':
      GetCronExecutionTimeseriesRequest$json,
  '.agents.v1.GetCronExecutionTimeseriesResponse':
      GetCronExecutionTimeseriesResponse$json,
  '.agents.v1.CronExecutionBucket': CronExecutionBucket$json,
  '.agents.v1.GetActivityMetricsRequest': GetActivityMetricsRequest$json,
  '.agents.v1.GetActivityMetricsResponse': GetActivityMetricsResponse$json,
};

/// Descriptor for `DashboardService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List dashboardServiceDescriptor = $convert.base64Decode(
    'ChBEYXNoYm9hcmRTZXJ2aWNlEkwKC0dldE92ZXJ2aWV3Eh0uYWdlbnRzLnYxLkdldE92ZXJ2aW'
    'V3UmVxdWVzdBoeLmFnZW50cy52MS5HZXRPdmVydmlld1Jlc3BvbnNlElgKD0dldEFjdGl2aXR5'
    'RmVlZBIhLmFnZW50cy52MS5HZXRBY3Rpdml0eUZlZWRSZXF1ZXN0GiIuYWdlbnRzLnYxLkdldE'
    'FjdGl2aXR5RmVlZFJlc3BvbnNlEnkKGkdldENyb25FeGVjdXRpb25UaW1lc2VyaWVzEiwuYWdl'
    'bnRzLnYxLkdldENyb25FeGVjdXRpb25UaW1lc2VyaWVzUmVxdWVzdBotLmFnZW50cy52MS5HZX'
    'RDcm9uRXhlY3V0aW9uVGltZXNlcmllc1Jlc3BvbnNlEmEKEkdldEFjdGl2aXR5TWV0cmljcxIk'
    'LmFnZW50cy52MS5HZXRBY3Rpdml0eU1ldHJpY3NSZXF1ZXN0GiUuYWdlbnRzLnYxLkdldEFjdG'
    'l2aXR5TWV0cmljc1Jlc3BvbnNl');

const $core.Map<$core.String, $core.dynamic> DaemonServiceBase$json = {
  '1': 'DaemonService',
  '2': [
    {
      '1': 'ListDaemonRuntimes',
      '2': '.agents.v1.ListDaemonRuntimesRequest',
      '3': '.agents.v1.ListDaemonRuntimesResponse'
    },
    {
      '1': 'GetDaemonRuntime',
      '2': '.agents.v1.GetDaemonRuntimeRequest',
      '3': '.agents.v1.GetDaemonRuntimeResponse'
    },
    {
      '1': 'CreateDaemonRuntime',
      '2': '.agents.v1.CreateDaemonRuntimeRequest',
      '3': '.agents.v1.CreateDaemonRuntimeResponse'
    },
    {
      '1': 'UpdateDaemonRuntime',
      '2': '.agents.v1.UpdateDaemonRuntimeRequest',
      '3': '.agents.v1.UpdateDaemonRuntimeResponse'
    },
    {
      '1': 'DeleteDaemonRuntime',
      '2': '.agents.v1.DeleteDaemonRuntimeRequest',
      '3': '.agents.v1.DeleteDaemonRuntimeResponse'
    },
    {
      '1': 'CreateDaemonRuntimeToken',
      '2': '.agents.v1.CreateDaemonRuntimeTokenRequest',
      '3': '.agents.v1.CreateDaemonRuntimeTokenResponse'
    },
    {
      '1': 'ListDaemons',
      '2': '.agents.v1.ListDaemonsRequest',
      '3': '.agents.v1.ListDaemonsResponse'
    },
    {
      '1': 'GetDaemon',
      '2': '.agents.v1.GetDaemonRequest',
      '3': '.agents.v1.GetDaemonResponse'
    },
    {
      '1': 'CancelDaemonTask',
      '2': '.agents.v1.CancelDaemonTaskRequest',
      '3': '.agents.v1.CancelDaemonTaskResponse'
    },
    {
      '1': 'ListDaemonTasks',
      '2': '.agents.v1.ListDaemonTasksRequest',
      '3': '.agents.v1.ListDaemonTasksResponse'
    },
    {
      '1': 'GetBridgeDiagnostics',
      '2': '.agents.v1.GetBridgeDiagnosticsRequest',
      '3': '.agents.v1.GetBridgeDiagnosticsResponse'
    },
  ],
};

@$core.Deprecated('Use daemonServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DaemonServiceBase$messageJson = {
  '.agents.v1.ListDaemonRuntimesRequest': ListDaemonRuntimesRequest$json,
  '.agents.v1.ListDaemonRuntimesResponse': ListDaemonRuntimesResponse$json,
  '.agents.v1.DaemonRuntime': DaemonRuntime$json,
  '.agents.v1.DaemonRuntime.LabelsEntry': DaemonRuntime_LabelsEntry$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetDaemonRuntimeRequest': GetDaemonRuntimeRequest$json,
  '.agents.v1.GetDaemonRuntimeResponse': GetDaemonRuntimeResponse$json,
  '.agents.v1.CreateDaemonRuntimeRequest': CreateDaemonRuntimeRequest$json,
  '.agents.v1.CreateDaemonRuntimeResponse': CreateDaemonRuntimeResponse$json,
  '.agents.v1.UpdateDaemonRuntimeRequest': UpdateDaemonRuntimeRequest$json,
  '.agents.v1.UpdateDaemonRuntimeResponse': UpdateDaemonRuntimeResponse$json,
  '.agents.v1.DeleteDaemonRuntimeRequest': DeleteDaemonRuntimeRequest$json,
  '.agents.v1.DeleteDaemonRuntimeResponse': DeleteDaemonRuntimeResponse$json,
  '.agents.v1.CreateDaemonRuntimeTokenRequest':
      CreateDaemonRuntimeTokenRequest$json,
  '.google.protobuf.Duration': $1.Duration$json,
  '.agents.v1.CreateDaemonRuntimeTokenResponse':
      CreateDaemonRuntimeTokenResponse$json,
  '.agents.v1.APIToken': $2.APIToken$json,
  '.agents.v1.ListDaemonsRequest': ListDaemonsRequest$json,
  '.agents.v1.ListDaemonsResponse': ListDaemonsResponse$json,
  '.agents.v1.DaemonStatus': DaemonStatus$json,
  '.agents.v1.DaemonStatus.LabelsEntry': DaemonStatus_LabelsEntry$json,
  '.agents.v1.GetDaemonRequest': GetDaemonRequest$json,
  '.agents.v1.GetDaemonResponse': GetDaemonResponse$json,
  '.agents.v1.CancelDaemonTaskRequest': CancelDaemonTaskRequest$json,
  '.agents.v1.CancelDaemonTaskResponse': CancelDaemonTaskResponse$json,
  '.agents.v1.ListDaemonTasksRequest': ListDaemonTasksRequest$json,
  '.agents.v1.ListDaemonTasksResponse': ListDaemonTasksResponse$json,
  '.agents.v1.DaemonTaskInFlight': DaemonTaskInFlight$json,
  '.agents.v1.GetBridgeDiagnosticsRequest': GetBridgeDiagnosticsRequest$json,
  '.agents.v1.GetBridgeDiagnosticsResponse': GetBridgeDiagnosticsResponse$json,
  '.agents.v1.BridgeDiagnostics': BridgeDiagnostics$json,
  '.agents.v1.LatencyPoint': LatencyPoint$json,
};

/// Descriptor for `DaemonService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List daemonServiceDescriptor = $convert.base64Decode(
    'Cg1EYWVtb25TZXJ2aWNlEmEKEkxpc3REYWVtb25SdW50aW1lcxIkLmFnZW50cy52MS5MaXN0RG'
    'FlbW9uUnVudGltZXNSZXF1ZXN0GiUuYWdlbnRzLnYxLkxpc3REYWVtb25SdW50aW1lc1Jlc3Bv'
    'bnNlElsKEEdldERhZW1vblJ1bnRpbWUSIi5hZ2VudHMudjEuR2V0RGFlbW9uUnVudGltZVJlcX'
    'Vlc3QaIy5hZ2VudHMudjEuR2V0RGFlbW9uUnVudGltZVJlc3BvbnNlEmQKE0NyZWF0ZURhZW1v'
    'blJ1bnRpbWUSJS5hZ2VudHMudjEuQ3JlYXRlRGFlbW9uUnVudGltZVJlcXVlc3QaJi5hZ2VudH'
    'MudjEuQ3JlYXRlRGFlbW9uUnVudGltZVJlc3BvbnNlEmQKE1VwZGF0ZURhZW1vblJ1bnRpbWUS'
    'JS5hZ2VudHMudjEuVXBkYXRlRGFlbW9uUnVudGltZVJlcXVlc3QaJi5hZ2VudHMudjEuVXBkYX'
    'RlRGFlbW9uUnVudGltZVJlc3BvbnNlEmQKE0RlbGV0ZURhZW1vblJ1bnRpbWUSJS5hZ2VudHMu'
    'djEuRGVsZXRlRGFlbW9uUnVudGltZVJlcXVlc3QaJi5hZ2VudHMudjEuRGVsZXRlRGFlbW9uUn'
    'VudGltZVJlc3BvbnNlEnMKGENyZWF0ZURhZW1vblJ1bnRpbWVUb2tlbhIqLmFnZW50cy52MS5D'
    'cmVhdGVEYWVtb25SdW50aW1lVG9rZW5SZXF1ZXN0GisuYWdlbnRzLnYxLkNyZWF0ZURhZW1vbl'
    'J1bnRpbWVUb2tlblJlc3BvbnNlEkwKC0xpc3REYWVtb25zEh0uYWdlbnRzLnYxLkxpc3REYWVt'
    'b25zUmVxdWVzdBoeLmFnZW50cy52MS5MaXN0RGFlbW9uc1Jlc3BvbnNlEkYKCUdldERhZW1vbh'
    'IbLmFnZW50cy52MS5HZXREYWVtb25SZXF1ZXN0GhwuYWdlbnRzLnYxLkdldERhZW1vblJlc3Bv'
    'bnNlElsKEENhbmNlbERhZW1vblRhc2sSIi5hZ2VudHMudjEuQ2FuY2VsRGFlbW9uVGFza1JlcX'
    'Vlc3QaIy5hZ2VudHMudjEuQ2FuY2VsRGFlbW9uVGFza1Jlc3BvbnNlElgKD0xpc3REYWVtb25U'
    'YXNrcxIhLmFnZW50cy52MS5MaXN0RGFlbW9uVGFza3NSZXF1ZXN0GiIuYWdlbnRzLnYxLkxpc3'
    'REYWVtb25UYXNrc1Jlc3BvbnNlEmcKFEdldEJyaWRnZURpYWdub3N0aWNzEiYuYWdlbnRzLnYx'
    'LkdldEJyaWRnZURpYWdub3N0aWNzUmVxdWVzdBonLmFnZW50cy52MS5HZXRCcmlkZ2VEaWFnbm'
    '9zdGljc1Jlc3BvbnNl');
