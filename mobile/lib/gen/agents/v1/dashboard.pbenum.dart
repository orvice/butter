// This is a generated file - do not edit.
//
// Generated from agents/v1/dashboard.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class GetActivityMetricsRequest_Range extends $pb.ProtobufEnum {
  static const GetActivityMetricsRequest_Range RANGE_UNSPECIFIED =
      GetActivityMetricsRequest_Range._(
          0, _omitEnumNames ? '' : 'RANGE_UNSPECIFIED');
  static const GetActivityMetricsRequest_Range RANGE_1D =
      GetActivityMetricsRequest_Range._(1, _omitEnumNames ? '' : 'RANGE_1D');
  static const GetActivityMetricsRequest_Range RANGE_7D =
      GetActivityMetricsRequest_Range._(2, _omitEnumNames ? '' : 'RANGE_7D');
  static const GetActivityMetricsRequest_Range RANGE_30D =
      GetActivityMetricsRequest_Range._(3, _omitEnumNames ? '' : 'RANGE_30D');

  static const $core.List<GetActivityMetricsRequest_Range> values =
      <GetActivityMetricsRequest_Range>[
    RANGE_UNSPECIFIED,
    RANGE_1D,
    RANGE_7D,
    RANGE_30D,
  ];

  static final $core.List<GetActivityMetricsRequest_Range?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static GetActivityMetricsRequest_Range? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const GetActivityMetricsRequest_Range._(super.value, super.name);
}

class GetCronExecutionTimeseriesRequest_Range extends $pb.ProtobufEnum {
  static const GetCronExecutionTimeseriesRequest_Range RANGE_UNSPECIFIED =
      GetCronExecutionTimeseriesRequest_Range._(
          0, _omitEnumNames ? '' : 'RANGE_UNSPECIFIED');
  static const GetCronExecutionTimeseriesRequest_Range RANGE_1D =
      GetCronExecutionTimeseriesRequest_Range._(
          1, _omitEnumNames ? '' : 'RANGE_1D');
  static const GetCronExecutionTimeseriesRequest_Range RANGE_7D =
      GetCronExecutionTimeseriesRequest_Range._(
          2, _omitEnumNames ? '' : 'RANGE_7D');
  static const GetCronExecutionTimeseriesRequest_Range RANGE_30D =
      GetCronExecutionTimeseriesRequest_Range._(
          3, _omitEnumNames ? '' : 'RANGE_30D');

  static const $core.List<GetCronExecutionTimeseriesRequest_Range> values =
      <GetCronExecutionTimeseriesRequest_Range>[
    RANGE_UNSPECIFIED,
    RANGE_1D,
    RANGE_7D,
    RANGE_30D,
  ];

  static final $core.List<GetCronExecutionTimeseriesRequest_Range?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static GetCronExecutionTimeseriesRequest_Range? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const GetCronExecutionTimeseriesRequest_Range._(super.value, super.name);
}

class ComponentHealth_Status extends $pb.ProtobufEnum {
  static const ComponentHealth_Status STATUS_UNSPECIFIED =
      ComponentHealth_Status._(0, _omitEnumNames ? '' : 'STATUS_UNSPECIFIED');
  static const ComponentHealth_Status STATUS_HEALTHY =
      ComponentHealth_Status._(1, _omitEnumNames ? '' : 'STATUS_HEALTHY');
  static const ComponentHealth_Status STATUS_DEGRADED =
      ComponentHealth_Status._(2, _omitEnumNames ? '' : 'STATUS_DEGRADED');
  static const ComponentHealth_Status STATUS_DOWN =
      ComponentHealth_Status._(3, _omitEnumNames ? '' : 'STATUS_DOWN');

  static const $core.List<ComponentHealth_Status> values =
      <ComponentHealth_Status>[
    STATUS_UNSPECIFIED,
    STATUS_HEALTHY,
    STATUS_DEGRADED,
    STATUS_DOWN,
  ];

  static final $core.List<ComponentHealth_Status?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static ComponentHealth_Status? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ComponentHealth_Status._(super.value, super.name);
}

class DaemonStatus_State extends $pb.ProtobufEnum {
  static const DaemonStatus_State STATE_UNSPECIFIED =
      DaemonStatus_State._(0, _omitEnumNames ? '' : 'STATE_UNSPECIFIED');
  static const DaemonStatus_State STATE_ONLINE =
      DaemonStatus_State._(1, _omitEnumNames ? '' : 'STATE_ONLINE');
  static const DaemonStatus_State STATE_IDLE =
      DaemonStatus_State._(2, _omitEnumNames ? '' : 'STATE_IDLE');
  static const DaemonStatus_State STATE_OFFLINE =
      DaemonStatus_State._(3, _omitEnumNames ? '' : 'STATE_OFFLINE');

  static const $core.List<DaemonStatus_State> values = <DaemonStatus_State>[
    STATE_UNSPECIFIED,
    STATE_ONLINE,
    STATE_IDLE,
    STATE_OFFLINE,
  ];

  static final $core.List<DaemonStatus_State?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static DaemonStatus_State? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DaemonStatus_State._(super.value, super.name);
}

class RemoteAgentStatus_State extends $pb.ProtobufEnum {
  static const RemoteAgentStatus_State STATE_UNSPECIFIED =
      RemoteAgentStatus_State._(0, _omitEnumNames ? '' : 'STATE_UNSPECIFIED');
  static const RemoteAgentStatus_State STATE_CONFIGURED =
      RemoteAgentStatus_State._(1, _omitEnumNames ? '' : 'STATE_CONFIGURED');
  static const RemoteAgentStatus_State STATE_ACTIVE =
      RemoteAgentStatus_State._(2, _omitEnumNames ? '' : 'STATE_ACTIVE');
  static const RemoteAgentStatus_State STATE_IDLE =
      RemoteAgentStatus_State._(3, _omitEnumNames ? '' : 'STATE_IDLE');
  static const RemoteAgentStatus_State STATE_UNREACHABLE =
      RemoteAgentStatus_State._(4, _omitEnumNames ? '' : 'STATE_UNREACHABLE');
  static const RemoteAgentStatus_State STATE_ERROR =
      RemoteAgentStatus_State._(5, _omitEnumNames ? '' : 'STATE_ERROR');

  static const $core.List<RemoteAgentStatus_State> values =
      <RemoteAgentStatus_State>[
    STATE_UNSPECIFIED,
    STATE_CONFIGURED,
    STATE_ACTIVE,
    STATE_IDLE,
    STATE_UNREACHABLE,
    STATE_ERROR,
  ];

  static final $core.List<RemoteAgentStatus_State?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static RemoteAgentStatus_State? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RemoteAgentStatus_State._(super.value, super.name);
}

class MCPServerStatus_State extends $pb.ProtobufEnum {
  static const MCPServerStatus_State STATE_UNSPECIFIED =
      MCPServerStatus_State._(0, _omitEnumNames ? '' : 'STATE_UNSPECIFIED');
  static const MCPServerStatus_State STATE_CONFIGURED =
      MCPServerStatus_State._(1, _omitEnumNames ? '' : 'STATE_CONFIGURED');
  static const MCPServerStatus_State STATE_CONNECTED =
      MCPServerStatus_State._(2, _omitEnumNames ? '' : 'STATE_CONNECTED');
  static const MCPServerStatus_State STATE_DISCONNECTED =
      MCPServerStatus_State._(3, _omitEnumNames ? '' : 'STATE_DISCONNECTED');
  static const MCPServerStatus_State STATE_ERROR =
      MCPServerStatus_State._(4, _omitEnumNames ? '' : 'STATE_ERROR');

  static const $core.List<MCPServerStatus_State> values =
      <MCPServerStatus_State>[
    STATE_UNSPECIFIED,
    STATE_CONFIGURED,
    STATE_CONNECTED,
    STATE_DISCONNECTED,
    STATE_ERROR,
  ];

  static final $core.List<MCPServerStatus_State?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static MCPServerStatus_State? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MCPServerStatus_State._(super.value, super.name);
}

class ChannelStatus_State extends $pb.ProtobufEnum {
  static const ChannelStatus_State STATE_UNSPECIFIED =
      ChannelStatus_State._(0, _omitEnumNames ? '' : 'STATE_UNSPECIFIED');
  static const ChannelStatus_State STATE_LIVE =
      ChannelStatus_State._(1, _omitEnumNames ? '' : 'STATE_LIVE');
  static const ChannelStatus_State STATE_PAUSED =
      ChannelStatus_State._(2, _omitEnumNames ? '' : 'STATE_PAUSED');
  static const ChannelStatus_State STATE_DISABLED =
      ChannelStatus_State._(3, _omitEnumNames ? '' : 'STATE_DISABLED');
  static const ChannelStatus_State STATE_ERROR =
      ChannelStatus_State._(4, _omitEnumNames ? '' : 'STATE_ERROR');

  static const $core.List<ChannelStatus_State> values = <ChannelStatus_State>[
    STATE_UNSPECIFIED,
    STATE_LIVE,
    STATE_PAUSED,
    STATE_DISABLED,
    STATE_ERROR,
  ];

  static final $core.List<ChannelStatus_State?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static ChannelStatus_State? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ChannelStatus_State._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
