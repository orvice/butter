import {
  LinearAppCredentialState,
  LinearInstallationCredentialState,
  LinearProcessingStatus,
  type LinearApp,
  type LinearInstallation,
  type LinearProcessingRecord,
} from '@/gen/agents/v1/linear_pb'
import { timestampDate, type Timestamp } from '@bufbuild/protobuf/wkt'
import { Code, ConnectError } from '@connectrpc/connect'

export const EVIDENCE_WINDOW_SIZE = 200

const STATUS_SUCCEEDED = LinearProcessingStatus.SUCCEEDED
const STATUS_FAILED = LinearProcessingStatus.FAILED
const STATUS_FAILED_UNCERTAIN = LinearProcessingStatus.FAILED_UNCERTAIN
const STATUS_CANCELLED = LinearProcessingStatus.CANCELLED
const STATUS_QUEUED = LinearProcessingStatus.QUEUED
const STATUS_RECEIVED = LinearProcessingStatus.RECEIVED
const STATUS_PROCESSING = LinearProcessingStatus.PROCESSING
const STATUS_READY_TO_DELIVER = LinearProcessingStatus.READY_TO_DELIVER

function millis(ts?: Timestamp): number {
  return ts ? timestampDate(ts).getTime() : 0
}

export function formatTime(ts?: Timestamp): string {
  return ts ? timestampDate(ts).toLocaleString() : 'unknown time'
}

export function pluralize(
  count: number,
  singular: string,
  plural = `${singular}s`
) {
  return `${count} ${count === 1 ? singular : plural}`
}

// --- Configuration and credential facts -------------------------------------

export type LinearAppCondition = {
  level: 'incomplete' | 'inbound-disabled' | 'enabled'
  title: string
  detail: string
  nextStep: string
}

/**
 * appCondition states what is configured and what is only the desired
 * inbound state. It never calls the App connected, healthy, or verified.
 */
export function appCondition(app: LinearApp): LinearAppCondition {
  const missing: string[] = []
  if (!app.clientSecretSet) missing.push('the OAuth client secret')
  if (!app.webhookSecretSet) missing.push('the webhook signing secret')
  if (!app.callbackUrl) missing.push('the public base URL (callback URL)')
  if (
    app.credentialState === LinearAppCredentialState.INCOMPLETE ||
    missing.length > 0
  ) {
    return {
      level: 'incomplete',
      title: 'Configuration incomplete',
      detail: `Missing ${missing.length > 0 ? missing.join(', ') : 'a required credential'}.`,
      nextStep:
        'Set the missing secret or the platform public base URL, then install the App before expecting deliveries.',
    }
  }
  if (!app.inboundEnabled) {
    return {
      level: 'inbound-disabled',
      title: 'Configured, inbound disabled',
      detail:
        'Linear deliveries are not accepted. Existing evidence stays visible.',
      nextStep:
        'Turn inbound on when the App is installed in the target organization. Turning it off never interrupts a turn that is already running.',
    }
  }
  return {
    level: 'enabled',
    title: 'Inbound enabled (desired state)',
    detail:
      'Butter will accept new Linear deliveries for this App. This is the desired state, not a health signal: check the evidence below for what actually ran.',
    nextStep:
      'Send a delegated issue or @mention from Linear and watch the evidence below.',
  }
}

export type InstallationCredentialFact = {
  needsReinstall: boolean
  /** True while no credential outcome has been recorded at all. */
  unrecorded: boolean
  /** Short badge label; never claims more than the recorded state. */
  label: string
  title: string
  detail: string
  nextStep?: string
}

/**
 * installationCredentialFact reports only what Butter recorded. "Credential
 * last known usable" is shown for VALID alone: an unspecified state, a
 * future enum value, or a reinstall requirement must never read as a usable
 * credential.
 */
export function installationCredentialFact(
  installation: LinearInstallation
): InstallationCredentialFact {
  switch (installation.credentialState) {
    case LinearInstallationCredentialState.VALID:
      return {
        needsReinstall: false,
        unrecorded: false,
        label: 'Credential last known usable',
        title: 'Credential: last known usable',
        detail:
          'The stored token worked, or was refreshed, on its last use. This is a credential fact, not a claim that the organization is reachable now.',
      }
    case LinearInstallationCredentialState.NEEDS_REINSTALL:
      return {
        needsReinstall: true,
        unrecorded: false,
        label: 'Reinstall needed',
        title: 'Reinstall needed',
        detail:
          installation.lastCredentialError ||
          'Linear refused the stored token or its refresh, so this installation cannot authenticate.',
        nextStep:
          'Install this App into the organization again from the Installations card. Earlier records stay visible as history.',
      }
    default:
      // UNSPECIFIED (and any state this build does not know) records no
      // outcome, so the copy says exactly that instead of guessing.
      return {
        needsReinstall: false,
        unrecorded: true,
        label: 'Credential state not recorded',
        title: 'Credential state not recorded',
        detail:
          'No token use or refresh outcome has been recorded for this installation yet. This is a missing fact, not a health claim.',
        nextStep:
          'Delegate an issue to the App or @mention it in Linear; the next authenticated call records the credential outcome.',
      }
  }
}

// --- One delivery record ----------------------------------------------------

export type RecordTone =
  | 'success'
  | 'rejected'
  | 'failed'
  | 'uncertain'
  | 'stopped'
  | 'pending'
  | 'neutral'

export type RecordDiagnosis = {
  tone: RecordTone
  title: string
  /** What the record says happened. Never a credential. */
  reason?: string
  /** What a person can do next. */
  nextStep?: string
}

/**
 * isAgentRunEvidence is the only definition of "the Agent really ran and
 * answered" used on this page:
 * - the record belongs to this App and (when given) this Installation,
 * - it was produced by the App's current Agent, so a record from an earlier
 *   binding is not evidence for the current one,
 * - it is a created or prompted delivery that SUCCEEDED with an invocation
 *   ID, a persisted reply, and that reply reached Linear.
 *
 * A Stop delivery never starts Agent work, so it has no invocation ID and is
 * never counted here; a record that only answered without running the Agent
 * (an allowlist rejection, a disabled App, a Stop acknowledgement) is also
 * excluded. This is historical evidence inside the read window, not current
 * health.
 */
export function isAgentRunEvidence(
  record: LinearProcessingRecord,
  app: LinearApp,
  installationId?: string
): boolean {
  if (record.appId !== app.id) return false
  if (installationId !== undefined && record.installationId !== installationId)
    return false
  if (record.agentId !== app.agentId) return false
  if (record.action !== 'created' && record.action !== 'prompted') return false
  if (record.status !== STATUS_SUCCEEDED) return false
  if (record.outputType === 'error') return false
  if (!record.invocationId) return false
  if (!record.delivered) return false
  if (!record.output) return false
  return true
}

/**
 * diagnoseRecord explains one delivery and its next step. It distinguishes
 * a delivery that was never observed from one that was observed and then
 * rejected, failed, stopped, or is still running.
 */
export function diagnoseRecord(
  record: LinearProcessingRecord,
  app?: LinearApp
): RecordDiagnosis {
  switch (record.status) {
    case STATUS_QUEUED:
      return {
        tone: 'pending',
        title: 'Queued behind the current turn',
        reason: 'The session was busy, so this prompt waits for the next turn.',
        nextStep: 'No action needed; the reply will be posted when it runs.',
      }
    case STATUS_RECEIVED:
      return {
        tone: 'pending',
        title: 'Accepted, not started',
        reason: 'The delivery is claimed but Agent work has not started.',
        nextStep: 'No action needed unless it stays in this state.',
      }
    case STATUS_PROCESSING:
      return {
        tone: 'pending',
        title: 'Agent may be running',
        reason: 'Agent work started; a crash from here is not retried.',
        nextStep: 'Wait for the reply. A crashed run is settled as uncertain.',
      }
    case STATUS_READY_TO_DELIVER:
      return {
        tone: 'pending',
        title: 'Reply ready, posting',
        reason: 'The reply is persisted and only the post remains.',
        nextStep:
          'If it stays here, check the installation credential and resend from Deliveries.',
      }
    case STATUS_SUCCEEDED:
      if (record.outputType === 'error') {
        // A delivered error reply either refused the delivery before Agent
        // work (no invocation) or reported a run that already happened. The
        // invocation ID is what tells them apart.
        const ran = Boolean(record.invocationId)
        return {
          tone: ran ? 'failed' : 'rejected',
          title: ran
            ? 'Ran, and replied with an error'
            : 'Observed and answered without running the Agent',
          reason:
            record.output ||
            (ran
              ? 'The Agent run ended with an error reply.'
              : 'The delivery was refused before Agent work started.'),
          nextStep: ran
            ? 'Read the error reply in Linear, fix the cause, then prompt again.'
            : rejectionNextStep(record, app),
        }
      }
      if (!record.invocationId) {
        return {
          tone: 'neutral',
          title: 'Handled without running the Agent',
          reason:
            record.output ||
            'No Agent work was needed, for example a Stop request or an acknowledgement.',
          nextStep:
            'There is no Agent run to verify here. Send a new prompt in Linear to run the Agent.',
        }
      }
      if (!record.delivered) {
        return {
          tone: 'neutral',
          title: 'Reply not confirmed in Linear',
          reason:
            'The record is marked succeeded but the post is not recorded.',
          nextStep: 'Check the installation credential, then resend the reply.',
        }
      }
      return {
        tone: 'success',
        title: 'Ran and replied',
        reason: 'The Agent run succeeded and its reply reached Linear.',
      }
    case STATUS_FAILED:
      if (record.output && !record.delivered) {
        return {
          tone: 'failed',
          title: 'Delivery failed after the reply was produced',
          reason: record.error || 'Linear did not accept the reply.',
          nextStep:
            'Resend the persisted reply from Deliveries. Resending never reruns the Agent.',
        }
      }
      return {
        tone: 'failed',
        title: 'Failed before the Agent ran',
        reason:
          record.error || 'The delivery failed before Agent work started.',
        nextStep:
          'Fix the cause, then trigger the event again from Linear. This delivery is safe to retry.',
      }
    case STATUS_FAILED_UNCERTAIN:
      return {
        tone: 'uncertain',
        title: 'Stopped mid-run, result uncertain',
        reason:
          record.error ||
          'The Agent may already have taken action before the run ended.',
        nextStep:
          'This is never retried automatically. Check the likely side effects, then prompt again in Linear if appropriate.',
      }
    case STATUS_CANCELLED:
      return {
        tone: 'stopped',
        title: 'Stopped',
        reason:
          record.error ||
          'The turn was stopped by the user, or the queued follow-up was discarded by a stop.',
        nextStep:
          'The Agent did not finish this turn. Send a new prompt in Linear to continue.',
      }
    default:
      return {
        tone: 'neutral',
        title: 'Recorded',
        reason: record.error || 'No further detail was recorded.',
      }
  }
}

function rejectionNextStep(
  record: LinearProcessingRecord,
  app?: LinearApp
): string {
  if (app && !app.inboundEnabled) {
    return 'Inbound is off for this App. An admin can turn it on, then trigger the event again from Linear.'
  }
  if (
    app &&
    app.allowedUserIds.length > 0 &&
    record.promptingUserId &&
    !app.allowedUserIds.includes(record.promptingUserId)
  ) {
    return `The prompting Linear user (${record.promptingUserId}) is not in this App's allowlist. An admin can add that user ID, then prompt again.`
  }
  return 'Check that the routed Agent is active and the App is enabled, then prompt again in Linear.'
}

function isIssue(record: LinearProcessingRecord): boolean {
  return (
    record.status === STATUS_FAILED ||
    record.status === STATUS_FAILED_UNCERTAIN ||
    record.status === STATUS_CANCELLED ||
    (record.status === STATUS_SUCCEEDED && record.outputType === 'error')
  )
}

// --- Installation evidence --------------------------------------------------

export type InstallationEvidence = {
  installation: LinearInstallation
  credential: InstallationCredentialFact
  /** Every record in the read window for this installation. */
  records: LinearProcessingRecord[]
  /** Records produced by the App's current Agent. */
  currentAgentRecords: LinearProcessingRecord[]
  /** Records from an earlier Agent binding; never counted as evidence. */
  otherAgentCount: number
  latestSuccess?: LinearProcessingRecord
  /** The newest issue newer than the latest success, when there is one. */
  latestIssueAfterSuccess?: LinearProcessingRecord
  recentIssues: LinearProcessingRecord[]
  /** Whether any delivery at all was observed for this installation. */
  observed: boolean
}

export function installationEvidence(
  app: LinearApp,
  installation: LinearInstallation,
  records: LinearProcessingRecord[]
): InstallationEvidence {
  const forInstallation = records.filter(
    (record) => record.installationId === installation.id
  )
  const currentAgentRecords = forInstallation.filter(
    (record) => record.agentId === app.agentId
  )
  const otherAgentCount = forInstallation.length - currentAgentRecords.length
  const latestSuccess = currentAgentRecords.find((record) =>
    isAgentRunEvidence(record, app, installation.id)
  )
  const latestIssueAfterSuccess = latestSuccess
    ? currentAgentRecords.find(
        (record) =>
          isIssue(record) &&
          millis(record.createdAt) > millis(latestSuccess.createdAt)
      )
    : undefined
  return {
    installation,
    credential: installationCredentialFact(installation),
    records: forInstallation,
    currentAgentRecords,
    otherAgentCount,
    latestSuccess,
    latestIssueAfterSuccess,
    recentIssues: currentAgentRecords.filter(isIssue).slice(0, 5),
    observed: forInstallation.length > 0,
  }
}

// --- Read window ------------------------------------------------------------

export type EvidenceWindow = {
  records: LinearProcessingRecord[]
  /** The read returned the full cap, so older records were not read. */
  capped: boolean
  newestAt?: Timestamp
  oldestAt?: Timestamp
}

export function evidenceWindow(
  records: LinearProcessingRecord[]
): EvidenceWindow {
  return {
    records,
    capped: records.length >= EVIDENCE_WINDOW_SIZE,
    newestAt: records[0]?.createdAt,
    oldestAt: records[records.length - 1]?.createdAt,
  }
}

// --- Read failures ----------------------------------------------------------

export type ReadErrorCopy = {
  title: string
  detail: string
  nextStep: string
}

/**
 * readErrorCopy explains why the evidence could not be read. A permission or
 * authentication failure is an access limit of the viewer, not evidence that
 * no delivery ever ran, so it must never be shown as "no events".
 */
export function readErrorCopy(error: unknown): ReadErrorCopy {
  if (
    error instanceof ConnectError &&
    (error.code === Code.PermissionDenied ||
      error.code === Code.Unauthenticated)
  ) {
    return {
      title: 'Delivery records are not visible to this account',
      detail:
        'Reading Linear processing history needs workspace access to Linear. This is a read-scope limit, not evidence that nothing ran.',
      nextStep:
        'Ask a workspace owner for Linear access, then reopen this page.',
    }
  }
  return {
    title: 'Delivery evidence could not be read',
    detail:
      error instanceof Error && error.message
        ? error.message
        : 'The read failed without a detail.',
    nextStep:
      'Read it again. If it keeps failing, check that the Linear runtime is running and that this account can read the workspace.',
  }
}
