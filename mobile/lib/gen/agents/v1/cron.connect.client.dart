//
//  Generated code. Do not modify.
//  source: agents/v1/cron.proto
//

import "package:connectrpc/connect.dart" as connect;
import "cron.pb.dart" as agentsv1cron;
import "cron.connect.spec.dart" as specs;

/// CronJobService manages cron jobs and their execution history.
extension type CronJobServiceClient (connect.Transport _transport) {
  /// ListCronJobs returns all cron jobs.
  Future<agentsv1cron.ListCronJobsResponse> listCronJobs(
    agentsv1cron.ListCronJobsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.listCronJobs,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetCronJob returns a single cron job by name.
  Future<agentsv1cron.GetCronJobResponse> getCronJob(
    agentsv1cron.GetCronJobRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.getCronJob,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateCronJob creates a new cron job and schedules it.
  Future<agentsv1cron.CreateCronJobResponse> createCronJob(
    agentsv1cron.CreateCronJobRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.createCronJob,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateCronJob updates an existing cron job and reschedules it.
  Future<agentsv1cron.UpdateCronJobResponse> updateCronJob(
    agentsv1cron.UpdateCronJobRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.updateCronJob,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteCronJob removes a cron job and unschedules it.
  Future<agentsv1cron.DeleteCronJobResponse> deleteCronJob(
    agentsv1cron.DeleteCronJobRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.deleteCronJob,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListCronExecutions returns execution records, optionally filtered by job name.
  Future<agentsv1cron.ListCronExecutionsResponse> listCronExecutions(
    agentsv1cron.ListCronExecutionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.listCronExecutions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RunCronJobNow triggers an immediate execution of the cron job, bypassing
  /// the schedule. The execution is recorded the same way scheduled runs are.
  Future<agentsv1cron.RunCronJobNowResponse> runCronJobNow(
    agentsv1cron.RunCronJobNowRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.CronJobService.runCronJobNow,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
