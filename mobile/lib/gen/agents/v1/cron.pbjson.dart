// This is a generated file - do not edit.
//
// Generated from agents/v1/cron.proto.

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
    as $0;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pbjson.dart'
    as $1;

@$core.Deprecated('Use cronConcurrencyPolicyDescriptor instead')
const CronConcurrencyPolicy$json = {
  '1': 'CronConcurrencyPolicy',
  '2': [
    {'1': 'CRON_CONCURRENCY_POLICY_UNSPECIFIED', '2': 0},
    {'1': 'CRON_CONCURRENCY_POLICY_SKIP', '2': 1},
    {'1': 'CRON_CONCURRENCY_POLICY_QUEUE', '2': 2},
    {'1': 'CRON_CONCURRENCY_POLICY_REPLACE', '2': 3},
    {'1': 'CRON_CONCURRENCY_POLICY_ALLOW', '2': 4},
  ],
};

/// Descriptor for `CronConcurrencyPolicy`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List cronConcurrencyPolicyDescriptor = $convert.base64Decode(
    'ChVDcm9uQ29uY3VycmVuY3lQb2xpY3kSJwojQ1JPTl9DT05DVVJSRU5DWV9QT0xJQ1lfVU5TUE'
    'VDSUZJRUQQABIgChxDUk9OX0NPTkNVUlJFTkNZX1BPTElDWV9TS0lQEAESIQodQ1JPTl9DT05D'
    'VVJSRU5DWV9QT0xJQ1lfUVVFVUUQAhIjCh9DUk9OX0NPTkNVUlJFTkNZX1BPTElDWV9SRVBMQU'
    'NFEAMSIQodQ1JPTl9DT05DVVJSRU5DWV9QT0xJQ1lfQUxMT1cQBA==');

@$core.Deprecated('Use cronNotifyOnDescriptor instead')
const CronNotifyOn$json = {
  '1': 'CronNotifyOn',
  '2': [
    {'1': 'CRON_NOTIFY_ON_UNSPECIFIED', '2': 0},
    {'1': 'CRON_NOTIFY_ON_ALWAYS', '2': 1},
    {'1': 'CRON_NOTIFY_ON_FAILURE', '2': 2},
    {'1': 'CRON_NOTIFY_ON_SUCCESS', '2': 3},
  ],
};

/// Descriptor for `CronNotifyOn`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List cronNotifyOnDescriptor = $convert.base64Decode(
    'CgxDcm9uTm90aWZ5T24SHgoaQ1JPTl9OT1RJRllfT05fVU5TUEVDSUZJRUQQABIZChVDUk9OX0'
    '5PVElGWV9PTl9BTFdBWVMQARIaChZDUk9OX05PVElGWV9PTl9GQUlMVVJFEAISGgoWQ1JPTl9O'
    'T1RJRllfT05fU1VDQ0VTUxAD');

@$core.Deprecated('Use cronDeliveryTypeDescriptor instead')
const CronDeliveryType$json = {
  '1': 'CronDeliveryType',
  '2': [
    {'1': 'CRON_DELIVERY_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'CRON_DELIVERY_TYPE_LOG', '2': 1},
    {'1': 'CRON_DELIVERY_TYPE_WEBHOOK', '2': 2},
    {'1': 'CRON_DELIVERY_TYPE_CHANNEL', '2': 3},
    {'1': 'CRON_DELIVERY_TYPE_NOTIFY_GROUP', '2': 4},
    {'1': 'CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION', '2': 5},
  ],
};

/// Descriptor for `CronDeliveryType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List cronDeliveryTypeDescriptor = $convert.base64Decode(
    'ChBDcm9uRGVsaXZlcnlUeXBlEiIKHkNST05fREVMSVZFUllfVFlQRV9VTlNQRUNJRklFRBAAEh'
    'oKFkNST05fREVMSVZFUllfVFlQRV9MT0cQARIeChpDUk9OX0RFTElWRVJZX1RZUEVfV0VCSE9P'
    'SxACEh4KGkNST05fREVMSVZFUllfVFlQRV9DSEFOTkVMEAMSIwofQ1JPTl9ERUxJVkVSWV9UWV'
    'BFX05PVElGWV9HUk9VUBAEEisKJ0NST05fREVMSVZFUllfVFlQRV9URUxFR1JBTV9ERVNUSU5B'
    'VElPThAF');

@$core.Deprecated('Use cronExecutionStatusDescriptor instead')
const CronExecutionStatus$json = {
  '1': 'CronExecutionStatus',
  '2': [
    {'1': 'CRON_EXECUTION_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'CRON_EXECUTION_STATUS_SUCCESS', '2': 1},
    {'1': 'CRON_EXECUTION_STATUS_ERROR', '2': 2},
    {'1': 'CRON_EXECUTION_STATUS_SKIPPED', '2': 3},
    {'1': 'CRON_EXECUTION_STATUS_CANCELLED', '2': 4},
    {'1': 'CRON_EXECUTION_STATUS_WAITING_INPUT', '2': 5},
  ],
};

/// Descriptor for `CronExecutionStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List cronExecutionStatusDescriptor = $convert.base64Decode(
    'ChNDcm9uRXhlY3V0aW9uU3RhdHVzEiUKIUNST05fRVhFQ1VUSU9OX1NUQVRVU19VTlNQRUNJRk'
    'lFRBAAEiEKHUNST05fRVhFQ1VUSU9OX1NUQVRVU19TVUNDRVNTEAESHwobQ1JPTl9FWEVDVVRJ'
    'T05fU1RBVFVTX0VSUk9SEAISIQodQ1JPTl9FWEVDVVRJT05fU1RBVFVTX1NLSVBQRUQQAxIjCh'
    '9DUk9OX0VYRUNVVElPTl9TVEFUVVNfQ0FOQ0VMTEVEEAQSJwojQ1JPTl9FWEVDVVRJT05fU1RB'
    'VFVTX1dBSVRJTkdfSU5QVVQQBQ==');

@$core.Deprecated('Use cronExecutionTriggerTypeDescriptor instead')
const CronExecutionTriggerType$json = {
  '1': 'CronExecutionTriggerType',
  '2': [
    {'1': 'CRON_EXECUTION_TRIGGER_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE', '2': 1},
    {'1': 'CRON_EXECUTION_TRIGGER_TYPE_MANUAL', '2': 2},
  ],
};

/// Descriptor for `CronExecutionTriggerType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List cronExecutionTriggerTypeDescriptor = $convert.base64Decode(
    'ChhDcm9uRXhlY3V0aW9uVHJpZ2dlclR5cGUSKwonQ1JPTl9FWEVDVVRJT05fVFJJR0dFUl9UWV'
    'BFX1VOU1BFQ0lGSUVEEAASKAokQ1JPTl9FWEVDVVRJT05fVFJJR0dFUl9UWVBFX1NDSEVEVUxF'
    'EAESJgoiQ1JPTl9FWEVDVVRJT05fVFJJR0dFUl9UWVBFX01BTlVBTBAC');

@$core.Deprecated('Use cronJobDescriptor instead')
const CronJob$json = {
  '1': 'CronJob',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'schedule', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'schedule'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'agent_id', '3': 14, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'input', '3': 4, '4': 1, '5': 9, '10': 'input'},
    {'1': 'timezone', '3': 5, '4': 1, '5': 9, '10': 'timezone'},
    {'1': 'enabled', '3': 6, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'delivery',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronDelivery',
      '10': 'delivery'
    },
    {
      '1': 'timeout',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'timeout'
    },
    {
      '1': 'retry',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronRetryPolicy',
      '10': 'retry'
    },
    {
      '1': 'metadata',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.CronJob.MetadataEntry',
      '10': 'metadata'
    },
    {
      '1': 'concurrency_policy',
      '3': 11,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.CronConcurrencyPolicy',
      '10': 'concurrencyPolicy'
    },
    {
      '1': 'notify_on',
      '3': 12,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.CronNotifyOn',
      '10': 'notifyOn'
    },
    {'1': 'max_output_bytes', '3': 13, '4': 1, '5': 5, '10': 'maxOutputBytes'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [CronJob_MetadataEntry$json],
};

@$core.Deprecated('Use cronJobDescriptor instead')
const CronJob_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `CronJob`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cronJobDescriptor = $convert.base64Decode(
    'CgdDcm9uSm9iEhsKBG5hbWUYASABKAlCB/pCBHICEAFSBG5hbWUSIwoIc2NoZWR1bGUYAiABKA'
    'lCB/pCBHICEAFSCHNjaGVkdWxlEh0KCmFnZW50X25hbWUYAyABKAlSCWFnZW50TmFtZRIZCghh'
    'Z2VudF9pZBgOIAEoCVIHYWdlbnRJZBIUCgVpbnB1dBgEIAEoCVIFaW5wdXQSGgoIdGltZXpvbm'
    'UYBSABKAlSCHRpbWV6b25lEhgKB2VuYWJsZWQYBiABKAhSB2VuYWJsZWQSMwoIZGVsaXZlcnkY'
    'ByABKAsyFy5hZ2VudHMudjEuQ3JvbkRlbGl2ZXJ5UghkZWxpdmVyeRIzCgd0aW1lb3V0GAggAS'
    'gLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9uUgd0aW1lb3V0EjAKBXJldHJ5GAkgASgLMhou'
    'YWdlbnRzLnYxLkNyb25SZXRyeVBvbGljeVIFcmV0cnkSPAoIbWV0YWRhdGEYCiADKAsyIC5hZ2'
    'VudHMudjEuQ3JvbkpvYi5NZXRhZGF0YUVudHJ5UghtZXRhZGF0YRJPChJjb25jdXJyZW5jeV9w'
    'b2xpY3kYCyABKA4yIC5hZ2VudHMudjEuQ3JvbkNvbmN1cnJlbmN5UG9saWN5UhFjb25jdXJyZW'
    '5jeVBvbGljeRI0Cglub3RpZnlfb24YDCABKA4yFy5hZ2VudHMudjEuQ3Jvbk5vdGlmeU9uUghu'
    'b3RpZnlPbhIoChBtYXhfb3V0cHV0X2J5dGVzGA0gASgFUg5tYXhPdXRwdXRCeXRlcxIhCgx3b3'
    'Jrc3BhY2VfaWQYZCABKAlSC3dvcmtzcGFjZUlkGjsKDU1ldGFkYXRhRW50cnkSEAoDa2V5GAEg'
    'ASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use cronRetryPolicyDescriptor instead')
const CronRetryPolicy$json = {
  '1': 'CronRetryPolicy',
  '2': [
    {'1': 'max_attempts', '3': 1, '4': 1, '5': 5, '10': 'maxAttempts'},
    {
      '1': 'backoff',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'backoff'
    },
  ],
};

/// Descriptor for `CronRetryPolicy`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cronRetryPolicyDescriptor = $convert.base64Decode(
    'Cg9Dcm9uUmV0cnlQb2xpY3kSIQoMbWF4X2F0dGVtcHRzGAEgASgFUgttYXhBdHRlbXB0cxIzCg'
    'diYWNrb2ZmGAIgASgLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9uUgdiYWNrb2Zm');

@$core.Deprecated('Use cronDeliveryDescriptor instead')
const CronDelivery$json = {
  '1': 'CronDelivery',
  '2': [
    {
      '1': 'type',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.CronDeliveryType',
      '10': 'type'
    },
    {'1': 'webhook_url', '3': 2, '4': 1, '5': 9, '10': 'webhookUrl'},
    {'1': 'channel_name', '3': 3, '4': 1, '5': 9, '10': 'channelName'},
    {'1': 'chat_id', '3': 4, '4': 1, '5': 9, '10': 'chatId'},
    {'1': 'notify_group_name', '3': 5, '4': 1, '5': 9, '10': 'notifyGroupName'},
    {
      '1': 'telegram_destination_id',
      '3': 6,
      '4': 1,
      '5': 9,
      '10': 'telegramDestinationId'
    },
  ],
};

/// Descriptor for `CronDelivery`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cronDeliveryDescriptor = $convert.base64Decode(
    'CgxDcm9uRGVsaXZlcnkSLwoEdHlwZRgBIAEoDjIbLmFnZW50cy52MS5Dcm9uRGVsaXZlcnlUeX'
    'BlUgR0eXBlEh8KC3dlYmhvb2tfdXJsGAIgASgJUgp3ZWJob29rVXJsEiEKDGNoYW5uZWxfbmFt'
    'ZRgDIAEoCVILY2hhbm5lbE5hbWUSFwoHY2hhdF9pZBgEIAEoCVIGY2hhdElkEioKEW5vdGlmeV'
    '9ncm91cF9uYW1lGAUgASgJUg9ub3RpZnlHcm91cE5hbWUSNgoXdGVsZWdyYW1fZGVzdGluYXRp'
    'b25faWQYBiABKAlSFXRlbGVncmFtRGVzdGluYXRpb25JZA==');

@$core.Deprecated('Use cronExecutionDescriptor instead')
const CronExecution$json = {
  '1': 'CronExecution',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'job_name', '3': 2, '4': 1, '5': 9, '10': 'jobName'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'agent_id', '3': 18, '4': 1, '5': 9, '10': 'agentId'},
    {
      '1': 'status',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.CronExecutionStatus',
      '10': 'status'
    },
    {'1': 'input', '3': 5, '4': 1, '5': 9, '10': 'input'},
    {'1': 'output', '3': 6, '4': 1, '5': 9, '10': 'output'},
    {
      '1': 'started_at',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startedAt'
    },
    {
      '1': 'finished_at',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'finishedAt'
    },
    {'1': 'error', '3': 9, '4': 1, '5': 9, '10': 'error'},
    {'1': 'duration_ms', '3': 10, '4': 1, '5': 3, '10': 'durationMs'},
    {'1': 'attempt_count', '3': 11, '4': 1, '5': 5, '10': 'attemptCount'},
    {
      '1': 'trigger_type',
      '3': 12,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.CronExecutionTriggerType',
      '10': 'triggerType'
    },
    {'1': 'skipped_reason', '3': 13, '4': 1, '5': 9, '10': 'skippedReason'},
    {'1': 'truncated', '3': 14, '4': 1, '5': 8, '10': 'truncated'},
    {'1': 'session_app_name', '3': 15, '4': 1, '5': 9, '10': 'sessionAppName'},
    {'1': 'session_user_id', '3': 16, '4': 1, '5': 9, '10': 'sessionUserId'},
    {'1': 'session_id', '3': 17, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `CronExecution`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cronExecutionDescriptor = $convert.base64Decode(
    'Cg1Dcm9uRXhlY3V0aW9uEg4KAmlkGAEgASgJUgJpZBIZCghqb2JfbmFtZRgCIAEoCVIHam9iTm'
    'FtZRIdCgphZ2VudF9uYW1lGAMgASgJUglhZ2VudE5hbWUSGQoIYWdlbnRfaWQYEiABKAlSB2Fn'
    'ZW50SWQSNgoGc3RhdHVzGAQgASgOMh4uYWdlbnRzLnYxLkNyb25FeGVjdXRpb25TdGF0dXNSBn'
    'N0YXR1cxIUCgVpbnB1dBgFIAEoCVIFaW5wdXQSFgoGb3V0cHV0GAYgASgJUgZvdXRwdXQSOQoK'
    'c3RhcnRlZF9hdBgHIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXN0YXJ0ZWRBdB'
    'I7CgtmaW5pc2hlZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmZpbmlz'
    'aGVkQXQSFAoFZXJyb3IYCSABKAlSBWVycm9yEh8KC2R1cmF0aW9uX21zGAogASgDUgpkdXJhdG'
    'lvbk1zEiMKDWF0dGVtcHRfY291bnQYCyABKAVSDGF0dGVtcHRDb3VudBJGCgx0cmlnZ2VyX3R5'
    'cGUYDCABKA4yIy5hZ2VudHMudjEuQ3JvbkV4ZWN1dGlvblRyaWdnZXJUeXBlUgt0cmlnZ2VyVH'
    'lwZRIlCg5za2lwcGVkX3JlYXNvbhgNIAEoCVINc2tpcHBlZFJlYXNvbhIcCgl0cnVuY2F0ZWQY'
    'DiABKAhSCXRydW5jYXRlZBIoChBzZXNzaW9uX2FwcF9uYW1lGA8gASgJUg5zZXNzaW9uQXBwTm'
    'FtZRImCg9zZXNzaW9uX3VzZXJfaWQYECABKAlSDXNlc3Npb25Vc2VySWQSHQoKc2Vzc2lvbl9p'
    'ZBgRIAEoCVIJc2Vzc2lvbklkEiEKDHdvcmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYWNlSWQ=');

@$core.Deprecated('Use listCronJobsRequestDescriptor instead')
const ListCronJobsRequest$json = {
  '1': 'ListCronJobsRequest',
};

/// Descriptor for `ListCronJobsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCronJobsRequestDescriptor =
    $convert.base64Decode('ChNMaXN0Q3JvbkpvYnNSZXF1ZXN0');

@$core.Deprecated('Use listCronJobsResponseDescriptor instead')
const ListCronJobsResponse$json = {
  '1': 'ListCronJobsResponse',
  '2': [
    {
      '1': 'cron_jobs',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJobs'
    },
  ],
};

/// Descriptor for `ListCronJobsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCronJobsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0Q3JvbkpvYnNSZXNwb25zZRIvCgljcm9uX2pvYnMYASADKAsyEi5hZ2VudHMudjEuQ3'
    'JvbkpvYlIIY3JvbkpvYnM=');

@$core.Deprecated('Use getCronJobRequestDescriptor instead')
const GetCronJobRequest$json = {
  '1': 'GetCronJobRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `GetCronJobRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCronJobRequestDescriptor = $convert
    .base64Decode('ChFHZXRDcm9uSm9iUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use createCronJobRequestDescriptor instead')
const CreateCronJobRequest$json = {
  '1': 'CreateCronJobRequest',
  '2': [
    {
      '1': 'cron_job',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJob'
    },
  ],
};

/// Descriptor for `CreateCronJobRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCronJobRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVDcm9uSm9iUmVxdWVzdBItCghjcm9uX2pvYhgBIAEoCzISLmFnZW50cy52MS5Dcm'
    '9uSm9iUgdjcm9uSm9i');

@$core.Deprecated('Use updateCronJobRequestDescriptor instead')
const UpdateCronJobRequest$json = {
  '1': 'UpdateCronJobRequest',
  '2': [
    {
      '1': 'cron_job',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJob'
    },
  ],
};

/// Descriptor for `UpdateCronJobRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCronJobRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVDcm9uSm9iUmVxdWVzdBItCghjcm9uX2pvYhgBIAEoCzISLmFnZW50cy52MS5Dcm'
    '9uSm9iUgdjcm9uSm9i');

@$core.Deprecated('Use deleteCronJobRequestDescriptor instead')
const DeleteCronJobRequest$json = {
  '1': 'DeleteCronJobRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `DeleteCronJobRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCronJobRequestDescriptor = $convert
    .base64Decode('ChREZWxldGVDcm9uSm9iUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use getCronJobResponseDescriptor instead')
const GetCronJobResponse$json = {
  '1': 'GetCronJobResponse',
  '2': [
    {
      '1': 'cron_job',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJob'
    },
  ],
};

/// Descriptor for `GetCronJobResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCronJobResponseDescriptor = $convert.base64Decode(
    'ChJHZXRDcm9uSm9iUmVzcG9uc2USLQoIY3Jvbl9qb2IYASABKAsyEi5hZ2VudHMudjEuQ3Jvbk'
    'pvYlIHY3JvbkpvYg==');

@$core.Deprecated('Use createCronJobResponseDescriptor instead')
const CreateCronJobResponse$json = {
  '1': 'CreateCronJobResponse',
  '2': [
    {
      '1': 'cron_job',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJob'
    },
  ],
};

/// Descriptor for `CreateCronJobResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCronJobResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVDcm9uSm9iUmVzcG9uc2USLQoIY3Jvbl9qb2IYASABKAsyEi5hZ2VudHMudjEuQ3'
    'JvbkpvYlIHY3JvbkpvYg==');

@$core.Deprecated('Use updateCronJobResponseDescriptor instead')
const UpdateCronJobResponse$json = {
  '1': 'UpdateCronJobResponse',
  '2': [
    {
      '1': 'cron_job',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJob'
    },
  ],
};

/// Descriptor for `UpdateCronJobResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateCronJobResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVDcm9uSm9iUmVzcG9uc2USLQoIY3Jvbl9qb2IYASABKAsyEi5hZ2VudHMudjEuQ3'
    'JvbkpvYlIHY3JvbkpvYg==');

@$core.Deprecated('Use deleteCronJobResponseDescriptor instead')
const DeleteCronJobResponse$json = {
  '1': 'DeleteCronJobResponse',
  '2': [
    {
      '1': 'cron_job',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronJob',
      '10': 'cronJob'
    },
  ],
};

/// Descriptor for `DeleteCronJobResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteCronJobResponseDescriptor = $convert.base64Decode(
    'ChVEZWxldGVDcm9uSm9iUmVzcG9uc2USLQoIY3Jvbl9qb2IYASABKAsyEi5hZ2VudHMudjEuQ3'
    'JvbkpvYlIHY3JvbkpvYg==');

@$core.Deprecated('Use listCronExecutionsRequestDescriptor instead')
const ListCronExecutionsRequest$json = {
  '1': 'ListCronExecutionsRequest',
  '2': [
    {'1': 'job_name', '3': 1, '4': 1, '5': 9, '10': 'jobName'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 3, '4': 1, '5': 9, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListCronExecutionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCronExecutionsRequestDescriptor = $convert.base64Decode(
    'ChlMaXN0Q3JvbkV4ZWN1dGlvbnNSZXF1ZXN0EhkKCGpvYl9uYW1lGAEgASgJUgdqb2JOYW1lEh'
    'sKCXBhZ2Vfc2l6ZRgCIAEoBVIIcGFnZVNpemUSHQoKcGFnZV90b2tlbhgDIAEoCVIJcGFnZVRv'
    'a2Vu');

@$core.Deprecated('Use runCronJobNowRequestDescriptor instead')
const RunCronJobNowRequest$json = {
  '1': 'RunCronJobNowRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
  ],
};

/// Descriptor for `RunCronJobNowRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runCronJobNowRequestDescriptor =
    $convert.base64Decode(
        'ChRSdW5Dcm9uSm9iTm93UmVxdWVzdBIbCgRuYW1lGAEgASgJQgf6QgRyAhABUgRuYW1l');

@$core.Deprecated('Use runCronJobNowResponseDescriptor instead')
const RunCronJobNowResponse$json = {
  '1': 'RunCronJobNowResponse',
  '2': [
    {
      '1': 'execution',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CronExecution',
      '10': 'execution'
    },
  ],
};

/// Descriptor for `RunCronJobNowResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runCronJobNowResponseDescriptor = $convert.base64Decode(
    'ChVSdW5Dcm9uSm9iTm93UmVzcG9uc2USNgoJZXhlY3V0aW9uGAEgASgLMhguYWdlbnRzLnYxLk'
    'Nyb25FeGVjdXRpb25SCWV4ZWN1dGlvbg==');

@$core.Deprecated('Use listCronExecutionsResponseDescriptor instead')
const ListCronExecutionsResponse$json = {
  '1': 'ListCronExecutionsResponse',
  '2': [
    {
      '1': 'executions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.CronExecution',
      '10': 'executions'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListCronExecutionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCronExecutionsResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0Q3JvbkV4ZWN1dGlvbnNSZXNwb25zZRI4CgpleGVjdXRpb25zGAEgAygLMhguYWdlbn'
        'RzLnYxLkNyb25FeGVjdXRpb25SCmV4ZWN1dGlvbnMSJgoPbmV4dF9wYWdlX3Rva2VuGAIgASgJ'
        'Ug1uZXh0UGFnZVRva2Vu');

const $core.Map<$core.String, $core.dynamic> CronJobServiceBase$json = {
  '1': 'CronJobService',
  '2': [
    {
      '1': 'ListCronJobs',
      '2': '.agents.v1.ListCronJobsRequest',
      '3': '.agents.v1.ListCronJobsResponse'
    },
    {
      '1': 'GetCronJob',
      '2': '.agents.v1.GetCronJobRequest',
      '3': '.agents.v1.GetCronJobResponse'
    },
    {
      '1': 'CreateCronJob',
      '2': '.agents.v1.CreateCronJobRequest',
      '3': '.agents.v1.CreateCronJobResponse'
    },
    {
      '1': 'UpdateCronJob',
      '2': '.agents.v1.UpdateCronJobRequest',
      '3': '.agents.v1.UpdateCronJobResponse'
    },
    {
      '1': 'DeleteCronJob',
      '2': '.agents.v1.DeleteCronJobRequest',
      '3': '.agents.v1.DeleteCronJobResponse'
    },
    {
      '1': 'ListCronExecutions',
      '2': '.agents.v1.ListCronExecutionsRequest',
      '3': '.agents.v1.ListCronExecutionsResponse'
    },
    {
      '1': 'RunCronJobNow',
      '2': '.agents.v1.RunCronJobNowRequest',
      '3': '.agents.v1.RunCronJobNowResponse'
    },
  ],
};

@$core.Deprecated('Use cronJobServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    CronJobServiceBase$messageJson = {
  '.agents.v1.ListCronJobsRequest': ListCronJobsRequest$json,
  '.agents.v1.ListCronJobsResponse': ListCronJobsResponse$json,
  '.agents.v1.CronJob': CronJob$json,
  '.agents.v1.CronDelivery': CronDelivery$json,
  '.google.protobuf.Duration': $0.Duration$json,
  '.agents.v1.CronRetryPolicy': CronRetryPolicy$json,
  '.agents.v1.CronJob.MetadataEntry': CronJob_MetadataEntry$json,
  '.agents.v1.GetCronJobRequest': GetCronJobRequest$json,
  '.agents.v1.GetCronJobResponse': GetCronJobResponse$json,
  '.agents.v1.CreateCronJobRequest': CreateCronJobRequest$json,
  '.agents.v1.CreateCronJobResponse': CreateCronJobResponse$json,
  '.agents.v1.UpdateCronJobRequest': UpdateCronJobRequest$json,
  '.agents.v1.UpdateCronJobResponse': UpdateCronJobResponse$json,
  '.agents.v1.DeleteCronJobRequest': DeleteCronJobRequest$json,
  '.agents.v1.DeleteCronJobResponse': DeleteCronJobResponse$json,
  '.agents.v1.ListCronExecutionsRequest': ListCronExecutionsRequest$json,
  '.agents.v1.ListCronExecutionsResponse': ListCronExecutionsResponse$json,
  '.agents.v1.CronExecution': CronExecution$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
  '.agents.v1.RunCronJobNowRequest': RunCronJobNowRequest$json,
  '.agents.v1.RunCronJobNowResponse': RunCronJobNowResponse$json,
};

/// Descriptor for `CronJobService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List cronJobServiceDescriptor = $convert.base64Decode(
    'Cg5Dcm9uSm9iU2VydmljZRJPCgxMaXN0Q3JvbkpvYnMSHi5hZ2VudHMudjEuTGlzdENyb25Kb2'
    'JzUmVxdWVzdBofLmFnZW50cy52MS5MaXN0Q3JvbkpvYnNSZXNwb25zZRJJCgpHZXRDcm9uSm9i'
    'EhwuYWdlbnRzLnYxLkdldENyb25Kb2JSZXF1ZXN0Gh0uYWdlbnRzLnYxLkdldENyb25Kb2JSZX'
    'Nwb25zZRJSCg1DcmVhdGVDcm9uSm9iEh8uYWdlbnRzLnYxLkNyZWF0ZUNyb25Kb2JSZXF1ZXN0'
    'GiAuYWdlbnRzLnYxLkNyZWF0ZUNyb25Kb2JSZXNwb25zZRJSCg1VcGRhdGVDcm9uSm9iEh8uYW'
    'dlbnRzLnYxLlVwZGF0ZUNyb25Kb2JSZXF1ZXN0GiAuYWdlbnRzLnYxLlVwZGF0ZUNyb25Kb2JS'
    'ZXNwb25zZRJSCg1EZWxldGVDcm9uSm9iEh8uYWdlbnRzLnYxLkRlbGV0ZUNyb25Kb2JSZXF1ZX'
    'N0GiAuYWdlbnRzLnYxLkRlbGV0ZUNyb25Kb2JSZXNwb25zZRJhChJMaXN0Q3JvbkV4ZWN1dGlv'
    'bnMSJC5hZ2VudHMudjEuTGlzdENyb25FeGVjdXRpb25zUmVxdWVzdBolLmFnZW50cy52MS5MaX'
    'N0Q3JvbkV4ZWN1dGlvbnNSZXNwb25zZRJSCg1SdW5Dcm9uSm9iTm93Eh8uYWdlbnRzLnYxLlJ1'
    'bkNyb25Kb2JOb3dSZXF1ZXN0GiAuYWdlbnRzLnYxLlJ1bkNyb25Kb2JOb3dSZXNwb25zZQ==');
