//
//  Generated code. Do not modify.
//  source: agents/v1/dashboard.proto
//

import "package:connectrpc/connect.dart" as connect;
import "dashboard.pb.dart" as agentsv1dashboard;

abstract final class DashboardService {
  /// Fully-qualified name of the DashboardService service.
  static const name = 'agents.v1.DashboardService';

  /// GetOverview returns aggregate counts, component health and the latest
  /// daemon handshake for the dashboard Overview screen.
  static const getOverview = connect.Spec(
    '/$name/GetOverview',
    connect.StreamType.unary,
    agentsv1dashboard.GetOverviewRequest.new,
    agentsv1dashboard.GetOverviewResponse.new,
  );

  /// GetActivityFeed returns the most recent invocation activity, oldest event
  /// first stripped, suitable for the dashboard Activity Feed widget.
  static const getActivityFeed = connect.Spec(
    '/$name/GetActivityFeed',
    connect.StreamType.unary,
    agentsv1dashboard.GetActivityFeedRequest.new,
    agentsv1dashboard.GetActivityFeedResponse.new,
  );

  /// GetCronExecutionTimeseries aggregates cron_executions into time buckets,
  /// suitable for the Overview screen "Cron Executions" chart.
  static const getCronExecutionTimeseries = connect.Spec(
    '/$name/GetCronExecutionTimeseries',
    connect.StreamType.unary,
    agentsv1dashboard.GetCronExecutionTimeseriesRequest.new,
    agentsv1dashboard.GetCronExecutionTimeseriesResponse.new,
  );

  /// GetActivityMetrics returns run-volume counts over a rolling time window,
  /// suitable for the Overview screen "Activity" metric cards (7d / 30d toggle).
  static const getActivityMetrics = connect.Spec(
    '/$name/GetActivityMetrics',
    connect.StreamType.unary,
    agentsv1dashboard.GetActivityMetricsRequest.new,
    agentsv1dashboard.GetActivityMetricsResponse.new,
  );
}
abstract final class DaemonService {
  /// Fully-qualified name of the DaemonService service.
  static const name = 'agents.v1.DaemonService';

  /// Workspace-scoped daemon runtime configuration. A runtime must exist before
  /// a worker credential can be issued or a runtime connection accepted.
  static const listDaemonRuntimes = connect.Spec(
    '/$name/ListDaemonRuntimes',
    connect.StreamType.unary,
    agentsv1dashboard.ListDaemonRuntimesRequest.new,
    agentsv1dashboard.ListDaemonRuntimesResponse.new,
  );

  static const getDaemonRuntime = connect.Spec(
    '/$name/GetDaemonRuntime',
    connect.StreamType.unary,
    agentsv1dashboard.GetDaemonRuntimeRequest.new,
    agentsv1dashboard.GetDaemonRuntimeResponse.new,
  );

  static const createDaemonRuntime = connect.Spec(
    '/$name/CreateDaemonRuntime',
    connect.StreamType.unary,
    agentsv1dashboard.CreateDaemonRuntimeRequest.new,
    agentsv1dashboard.CreateDaemonRuntimeResponse.new,
  );

  static const updateDaemonRuntime = connect.Spec(
    '/$name/UpdateDaemonRuntime',
    connect.StreamType.unary,
    agentsv1dashboard.UpdateDaemonRuntimeRequest.new,
    agentsv1dashboard.UpdateDaemonRuntimeResponse.new,
  );

  static const deleteDaemonRuntime = connect.Spec(
    '/$name/DeleteDaemonRuntime',
    connect.StreamType.unary,
    agentsv1dashboard.DeleteDaemonRuntimeRequest.new,
    agentsv1dashboard.DeleteDaemonRuntimeResponse.new,
  );

  static const createDaemonRuntimeToken = connect.Spec(
    '/$name/CreateDaemonRuntimeToken',
    connect.StreamType.unary,
    agentsv1dashboard.CreateDaemonRuntimeTokenRequest.new,
    agentsv1dashboard.CreateDaemonRuntimeTokenResponse.new,
  );

  static const listDaemons = connect.Spec(
    '/$name/ListDaemons',
    connect.StreamType.unary,
    agentsv1dashboard.ListDaemonsRequest.new,
    agentsv1dashboard.ListDaemonsResponse.new,
  );

  static const getDaemon = connect.Spec(
    '/$name/GetDaemon',
    connect.StreamType.unary,
    agentsv1dashboard.GetDaemonRequest.new,
    agentsv1dashboard.GetDaemonResponse.new,
  );

  /// CancelDaemonTask requests cancellation of a running task on any connected
  /// daemon. Returns NotFound if no online daemon is tracking the task.
  static const cancelDaemonTask = connect.Spec(
    '/$name/CancelDaemonTask',
    connect.StreamType.unary,
    agentsv1dashboard.CancelDaemonTaskRequest.new,
    agentsv1dashboard.CancelDaemonTaskResponse.new,
  );

  /// ListDaemonTasks returns all tasks currently in flight on connected daemons.
  static const listDaemonTasks = connect.Spec(
    '/$name/ListDaemonTasks',
    connect.StreamType.unary,
    agentsv1dashboard.ListDaemonTasksRequest.new,
    agentsv1dashboard.ListDaemonTasksResponse.new,
  );

  /// GetBridgeDiagnostics returns router-side resource usage and a recent
  /// bridge-latency sample series, suitable for the Bridge Diagnostics panel.
  static const getBridgeDiagnostics = connect.Spec(
    '/$name/GetBridgeDiagnostics',
    connect.StreamType.unary,
    agentsv1dashboard.GetBridgeDiagnosticsRequest.new,
    agentsv1dashboard.GetBridgeDiagnosticsResponse.new,
  );
}
