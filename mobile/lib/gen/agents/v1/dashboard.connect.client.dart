//
//  Generated code. Do not modify.
//  source: agents/v1/dashboard.proto
//

import "package:connectrpc/connect.dart" as connect;
import "dashboard.pb.dart" as agentsv1dashboard;
import "dashboard.connect.spec.dart" as specs;

extension type DashboardServiceClient (connect.Transport _transport) {
  /// GetOverview returns aggregate counts, component health and the latest
  /// daemon handshake for the dashboard Overview screen.
  Future<agentsv1dashboard.GetOverviewResponse> getOverview(
    agentsv1dashboard.GetOverviewRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DashboardService.getOverview,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetActivityFeed returns the most recent invocation activity, oldest event
  /// first stripped, suitable for the dashboard Activity Feed widget.
  Future<agentsv1dashboard.GetActivityFeedResponse> getActivityFeed(
    agentsv1dashboard.GetActivityFeedRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DashboardService.getActivityFeed,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCronExecutionTimeseries aggregates cron_executions into time buckets,
  /// suitable for the Overview screen "Cron Executions" chart.
  Future<agentsv1dashboard.GetCronExecutionTimeseriesResponse> getCronExecutionTimeseries(
    agentsv1dashboard.GetCronExecutionTimeseriesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DashboardService.getCronExecutionTimeseries,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetActivityMetrics returns run-volume counts over a rolling time window,
  /// suitable for the Overview screen "Activity" metric cards (7d / 30d toggle).
  Future<agentsv1dashboard.GetActivityMetricsResponse> getActivityMetrics(
    agentsv1dashboard.GetActivityMetricsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DashboardService.getActivityMetrics,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
extension type DaemonServiceClient (connect.Transport _transport) {
  /// Workspace-scoped daemon runtime configuration. A runtime must exist before
  /// a worker credential can be issued or a runtime connection accepted.
  Future<agentsv1dashboard.ListDaemonRuntimesResponse> listDaemonRuntimes(
    agentsv1dashboard.ListDaemonRuntimesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.listDaemonRuntimes,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.GetDaemonRuntimeResponse> getDaemonRuntime(
    agentsv1dashboard.GetDaemonRuntimeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.getDaemonRuntime,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.CreateDaemonRuntimeResponse> createDaemonRuntime(
    agentsv1dashboard.CreateDaemonRuntimeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.createDaemonRuntime,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.UpdateDaemonRuntimeResponse> updateDaemonRuntime(
    agentsv1dashboard.UpdateDaemonRuntimeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.updateDaemonRuntime,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.DeleteDaemonRuntimeResponse> deleteDaemonRuntime(
    agentsv1dashboard.DeleteDaemonRuntimeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.deleteDaemonRuntime,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.CreateDaemonRuntimeTokenResponse> createDaemonRuntimeToken(
    agentsv1dashboard.CreateDaemonRuntimeTokenRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.createDaemonRuntimeToken,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.ListDaemonsResponse> listDaemons(
    agentsv1dashboard.ListDaemonsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.listDaemons,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1dashboard.GetDaemonResponse> getDaemon(
    agentsv1dashboard.GetDaemonRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.getDaemon,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelDaemonTask requests cancellation of a running task on any connected
  /// daemon. Returns NotFound if no online daemon is tracking the task.
  Future<agentsv1dashboard.CancelDaemonTaskResponse> cancelDaemonTask(
    agentsv1dashboard.CancelDaemonTaskRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.cancelDaemonTask,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListDaemonTasks returns all tasks currently in flight on connected daemons.
  Future<agentsv1dashboard.ListDaemonTasksResponse> listDaemonTasks(
    agentsv1dashboard.ListDaemonTasksRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.listDaemonTasks,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetBridgeDiagnostics returns router-side resource usage and a recent
  /// bridge-latency sample series, suitable for the Bridge Diagnostics panel.
  Future<agentsv1dashboard.GetBridgeDiagnosticsResponse> getBridgeDiagnostics(
    agentsv1dashboard.GetBridgeDiagnosticsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DaemonService.getBridgeDiagnostics,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
