import type { RunStatus } from "@/components/butter/primitives";

// Takes the raw enum string rather than the CronExecutionStatus union: the
// frontend union lags the proto (no RUNNING/WAITING_INPUT/FAILED members yet),
// and switching on those literals against the union would be a type error.
export function cronExecStatus(status: string): RunStatus {
  switch (status) {
    case "CRON_EXECUTION_STATUS_RUNNING":
      return "running";
    case "CRON_EXECUTION_STATUS_SUCCESS":
    case "CRON_EXECUTION_STATUS_SUCCEEDED":
      return "success";
    case "CRON_EXECUTION_STATUS_ERROR":
    case "CRON_EXECUTION_STATUS_FAILED":
    case "CRON_EXECUTION_STATUS_CANCELLED":
      return "failed";
    case "CRON_EXECUTION_STATUS_WAITING_INPUT":
      return "waiting";
    default:
      return "never";
  }
}
