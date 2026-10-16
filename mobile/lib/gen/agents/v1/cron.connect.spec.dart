//
//  Generated code. Do not modify.
//  source: agents/v1/cron.proto
//

import "package:connectrpc/connect.dart" as connect;
import "cron.pb.dart" as agentsv1cron;

/// CronJobService manages cron jobs and their execution history.
abstract final class CronJobService {
  /// Fully-qualified name of the CronJobService service.
  static const name = 'agents.v1.CronJobService';

  /// ListCronJobs returns all cron jobs.
  static const listCronJobs = connect.Spec(
    '/$name/ListCronJobs',
    connect.StreamType.unary,
    agentsv1cron.ListCronJobsRequest.new,
    agentsv1cron.ListCronJobsResponse.new,
  );

  /// GetCronJob returns a single cron job by name.
  static const getCronJob = connect.Spec(
    '/$name/GetCronJob',
    connect.StreamType.unary,
    agentsv1cron.GetCronJobRequest.new,
    agentsv1cron.GetCronJobResponse.new,
  );

  /// CreateCronJob creates a new cron job and schedules it.
  static const createCronJob = connect.Spec(
    '/$name/CreateCronJob',
    connect.StreamType.unary,
    agentsv1cron.CreateCronJobRequest.new,
    agentsv1cron.CreateCronJobResponse.new,
  );

  /// UpdateCronJob updates an existing cron job and reschedules it.
  static const updateCronJob = connect.Spec(
    '/$name/UpdateCronJob',
    connect.StreamType.unary,
    agentsv1cron.UpdateCronJobRequest.new,
    agentsv1cron.UpdateCronJobResponse.new,
  );

  /// DeleteCronJob removes a cron job and unschedules it.
  static const deleteCronJob = connect.Spec(
    '/$name/DeleteCronJob',
    connect.StreamType.unary,
    agentsv1cron.DeleteCronJobRequest.new,
    agentsv1cron.DeleteCronJobResponse.new,
  );

  /// ListCronExecutions returns execution records, optionally filtered by job name.
  static const listCronExecutions = connect.Spec(
    '/$name/ListCronExecutions',
    connect.StreamType.unary,
    agentsv1cron.ListCronExecutionsRequest.new,
    agentsv1cron.ListCronExecutionsResponse.new,
  );

  /// RunCronJobNow triggers an immediate execution of the cron job, bypassing
  /// the schedule. The execution is recorded the same way scheduled runs are.
  static const runCronJobNow = connect.Spec(
    '/$name/RunCronJobNow',
    connect.StreamType.unary,
    agentsv1cron.RunCronJobNowRequest.new,
    agentsv1cron.RunCronJobNowResponse.new,
  );
}
