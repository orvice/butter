import { Link } from '@tanstack/react-router'
import type {
  LinearApp,
  LinearProcessingRecord,
} from '@/gen/agents/v1/linear_pb'
import {
  CircleAlert,
  CircleCheck,
  Clock,
  RefreshCw,
  TriangleAlert,
} from 'lucide-react'
import {
  useLinearApp,
  useLinearInstallations,
  useLinearProcessingRecords,
} from '@/api/linear'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  EVIDENCE_WINDOW_SIZE,
  appCondition,
  diagnoseRecord,
  evidenceWindow,
  formatTime,
  installationEvidence,
  pluralize,
  readErrorCopy,
  type InstallationEvidence,
  type LinearAppCondition,
  type RecordDiagnosis,
} from './evidence'

/**
 * The verification view reads the same processing records the receiver
 * writes: at most the newest 200, kept for 30 days. Everything below says
 * what that window observed; nothing here is a health check, a permanent
 * verification, or a claim about the current Linear organization.
 */
// --- Components -------------------------------------------------------------

function ConditionIcon({ level }: { level: LinearAppCondition['level'] }) {
  if (level === 'enabled') {
    return <CircleCheck className='mt-0.5 size-4 text-emerald-600' />
  }
  if (level === 'inbound-disabled') {
    return <CircleAlert className='mt-0.5 size-4 text-amber-600' />
  }
  return <CircleAlert className='mt-0.5 size-4 text-destructive' />
}

export function LinearAppConditionLine({ app }: { app: LinearApp }) {
  const condition = appCondition(app)
  return (
    <div className='space-y-1' data-testid='linear-app-condition'>
      <div className='flex items-start gap-2 text-sm font-medium'>
        <ConditionIcon level={condition.level} />
        {condition.title}
      </div>
      <p className='pl-6 text-sm text-muted-foreground'>{condition.detail}</p>
      <p className='pl-6 text-sm text-muted-foreground'>
        Next: {condition.nextStep}
      </p>
    </div>
  )
}

function RecordRow({ diagnosis }: { diagnosis: RecordDiagnosis }) {
  const tone =
    diagnosis.tone === 'failed' || diagnosis.tone === 'uncertain'
      ? 'text-destructive'
      : diagnosis.tone === 'success'
        ? 'text-emerald-700'
        : 'text-muted-foreground'
  return (
    <li className='space-y-1 border-l-2 border-muted pl-3'>
      <div className={`text-sm font-medium ${tone}`}>{diagnosis.title}</div>
      {diagnosis.reason && (
        <p className='text-sm break-words whitespace-pre-wrap text-muted-foreground'>
          {diagnosis.reason}
        </p>
      )}
      {diagnosis.nextStep && (
        <p className='text-sm'>Next: {diagnosis.nextStep}</p>
      )}
    </li>
  )
}

function InstallationDiagnostic({
  app,
  evidence,
  appRevision,
  readable,
}: {
  app: LinearApp
  evidence: InstallationEvidence
  appRevision: bigint
  /** False when the record read failed; never claim "no events" then. */
  readable: boolean
}) {
  const { installation } = evidence
  const success = evidence.latestSuccess
  const successRevision = success ? success.appRevision : 0n
  return (
    <div
      className='space-y-3 rounded-md border p-3'
      data-testid={`linear-installation-diagnostic-${installation.id}`}
    >
      <div className='flex flex-wrap items-center gap-2'>
        <span className='font-medium'>
          {installation.organizationName || installation.organizationId}
        </span>
        {evidence.credential.needsReinstall ? (
          <Badge variant='destructive'>Reinstall needed</Badge>
        ) : (
          <Badge variant='outline'>Credential last known usable</Badge>
        )}
      </div>
      <div className='text-xs text-muted-foreground'>
        App user {installation.appUserId || 'unknown'} · installed{' '}
        {formatTime(installation.installedAt)}
      </div>
      {installation.scopes.length > 0 && (
        <div className='text-xs text-muted-foreground'>
          Granted Linear scopes: {installation.scopes.join(', ')}
        </div>
      )}
      {installation.scopes.length > 0 &&
        !installation.scopes.includes('read') && (
          <p className='text-xs text-amber-700'>
            This installation does not include Linear&apos;s read scope. The
            evidence here only reflects what Butter recorded; some Linear data
            cannot be read.
          </p>
        )}

      <div className='space-y-1 text-sm'>
        <div className='font-medium'>{evidence.credential.title}</div>
        <p className='text-muted-foreground'>{evidence.credential.detail}</p>
        {evidence.credential.nextStep && (
          <p>Next: {evidence.credential.nextStep}</p>
        )}
      </div>

      <div className='space-y-1 text-sm'>
        {!readable ? (
          <p className='text-muted-foreground'>
            No evidence was read for this installation.
          </p>
        ) : success ? (
          <div className='space-y-1' data-testid='linear-recent-success'>
            <div className='flex items-center gap-2 font-medium text-emerald-700'>
              <CircleCheck className='size-4' />
              Recent successful run
            </div>
            <div className='font-mono text-xs text-muted-foreground'>
              {success.issueIdentifier || success.agentSessionId || 'session'} ·{' '}
              {success.action} · {formatTime(success.createdAt)} · invocation{' '}
              {success.invocationId}
            </div>
            <p className='text-xs text-muted-foreground'>
              Historical evidence from the read window, not current health. A
              later run may still fail.
            </p>
            {successRevision !== appRevision && (
              <p className='text-xs text-muted-foreground'>
                This run was accepted under App revision{' '}
                {String(successRevision)}; the current configuration is revision{' '}
                {String(appRevision)}.
              </p>
            )}
          </div>
        ) : (
          <div className='space-y-1' data-testid='linear-no-success'>
            <div className='font-medium'>
              No run evidence in the current record range
            </div>
            <p className='text-muted-foreground'>
              {evidence.observed
                ? `Observed ${pluralize(evidence.records.length, 'delivery', 'deliveries')} for this installation, but none of them ran this Agent and reached Linear with a reply.`
                : 'No events were observed for this installation in the current record range.'}
            </p>
            <p>
              Next:{' '}
              {evidence.observed
                ? 'Check the reasons below, fix the cause, then prompt again from Linear.'
                : 'Delegate an issue to the App or @mention it in Linear, then check this page again.'}
            </p>
          </div>
        )}
      </div>

      {readable && evidence.otherAgentCount > 0 && (
        <p className='text-xs text-muted-foreground'>
          {pluralize(evidence.otherAgentCount, 'record')} from an earlier Agent
          binding (or without an Agent stamp) are not counted as evidence for
          this Agent.
        </p>
      )}

      {readable && evidence.latestIssueAfterSuccess && (
        <Alert variant='destructive'>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>The latest delivery did not succeed</AlertTitle>
          <AlertDescription>
            A successful run exists earlier in the window, but the newest issue
            came after it. The last success is history, not the current outcome.
          </AlertDescription>
        </Alert>
      )}

      {readable && evidence.recentIssues.length > 0 && (
        <div className='space-y-2'>
          <div className='text-sm font-medium'>Recent failures and stops</div>
          <ul className='space-y-2' data-testid='linear-recent-issues'>
            {evidence.recentIssues.map((record) => (
              <RecordRow
                key={record.id}
                diagnosis={diagnoseRecord(record, app)}
              />
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

/**
 * LinearDiagnosticsCard is the verification view shared by the App page and
 * the guided setup. With `live`, it re-reads the window every 5 seconds only
 * while the page is visible; leaving the page unmounts it and stops the
 * polling. It is read-only.
 */
export function LinearDiagnosticsCard({
  appId,
  live = false,
  showHeading = true,
}: {
  appId: string
  live?: boolean
  showHeading?: boolean
}) {
  const appQuery = useLinearApp(appId)
  const installationsQuery = useLinearInstallations(appId)
  const recordsQuery = useLinearProcessingRecords(
    { appId, pageSize: EVIDENCE_WINDOW_SIZE },
    { live }
  )

  if (appQuery.isLoading) {
    return (
      <Skeleton className='h-40' data-testid='linear-diagnostics-loading' />
    )
  }

  const app = appQuery.data
  if (!app) {
    const copy = readErrorCopy(appQuery.error)
    return (
      <Card data-testid='linear-diagnostics'>
        <CardContent className='py-4'>
          <Alert variant='destructive' data-testid='linear-read-error'>
            <TriangleAlert className='h-4 w-4' />
            <AlertTitle>{copy.title}</AlertTitle>
            <AlertDescription>
              {copy.detail} Next: {copy.nextStep}
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  if (installationsQuery.isLoading || recordsQuery.isLoading) {
    return (
      <Skeleton className='h-40' data-testid='linear-diagnostics-loading' />
    )
  }

  const installations = installationsQuery.data
  const records = recordsQuery.data
  const readError = recordsQuery.error ?? installationsQuery.error
  const readErrorInfo = readError ? readErrorCopy(readError) : undefined
  // A failed read that still has the previous records is shown as stale
  // evidence; a failed read with no data never reads as "no events".
  const readable = records !== undefined
  const window = evidenceWindow(records ?? [])
  const evidence = (installations ?? []).map((installation) =>
    installationEvidence(app, installation, readable ? window.records : [])
  )
  const unstampedRecords = readable
    ? window.records.filter((record) => !record.installationId)
    : []
  const removedInstallationRecords = readable
    ? window.records.filter(
        (record) =>
          record.installationId &&
          !(installations ?? []).some(
            (installation) => installation.id === record.installationId
          )
      )
    : []

  return (
    <Card data-testid='linear-diagnostics'>
      <CardHeader className='pb-2'>
        <CardTitle className='flex items-center gap-2 text-base'>
          <RefreshCw className='h-4 w-4' />
          {showHeading ? 'Status and history' : 'Delivery evidence'}
        </CardTitle>
        <CardDescription>
          Configuration facts, credential facts, and delivery evidence from the
          newest {EVIDENCE_WINDOW_SIZE} processing records kept for 30 days.
          This is not a connected, health, or permanent verification state.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <LinearAppConditionLine app={app} />

        {readErrorInfo && (
          <Alert variant='destructive' data-testid='linear-read-error'>
            <TriangleAlert className='h-4 w-4' />
            <AlertTitle>{readErrorInfo.title}</AlertTitle>
            <AlertDescription>
              {readErrorInfo.detail}
              {readable
                ? ' The evidence below is from the last successful read.'
                : ''}{' '}
              Next: {readErrorInfo.nextStep}
            </AlertDescription>
          </Alert>
        )}

        {readable && (
          <div
            className='space-y-1 text-xs text-muted-foreground'
            data-testid='linear-read-window'
          >
            <div className='flex items-center gap-1.5'>
              <Clock className='size-3.5' />
              Read window: {pluralize(window.records.length, 'record')}
              {window.records.length > 0 &&
                `, newest ${formatTime(window.newestAt)}, oldest ${formatTime(window.oldestAt)}`}
              .
            </div>
            {window.capped && (
              <p className='text-amber-700'>
                The read window is full ({EVIDENCE_WINDOW_SIZE} records), so
                older records are outside this evidence. A lack of success here
                may only mean the evidence is out of range.
              </p>
            )}
            {live && (
              <p>
                Refreshes every 5 seconds while this page is visible; leaving
                the page stops the refresh.
              </p>
            )}
          </div>
        )}

        {evidence.length === 0 && installations !== undefined ? (
          <p
            className='text-sm text-muted-foreground'
            data-testid='linear-no-installation'
          >
            No Linear organization is installed yet. Install the App before
            expecting any delivery evidence.
          </p>
        ) : (
          evidence.map((item) => (
            <InstallationDiagnostic
              key={item.installation.id}
              app={app}
              evidence={item}
              appRevision={app.revision}
              readable={readable}
            />
          ))
        )}

        {unstampedRecords.length > 0 && (
          <p className='text-xs text-muted-foreground'>
            {pluralize(unstampedRecords.length, 'record')} were not stamped with
            an installation and are not counted as evidence for a current
            installation.
          </p>
        )}

        {removedInstallationRecords.length > 0 && (
          <p className='text-xs text-muted-foreground'>
            {pluralize(removedInstallationRecords.length, 'record')} belong to
            an installation that is no longer listed. They stay as history but
            are not evidence for a current installation.
          </p>
        )}

        <div className='flex flex-wrap items-center gap-3 text-xs text-muted-foreground'>
          <span data-testid='linear-diagnostics-updated'>
            Read at{' '}
            {recordsQuery.dataUpdatedAt
              ? new Date(recordsQuery.dataUpdatedAt).toLocaleTimeString()
              : '—'}
          </span>
          <Link
            to='/linear-deliveries'
            className='underline underline-offset-2'
          >
            Open Deliveries
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}

/**
 * LinearRecordDiagnosis shows the cause and next step of one record. It is
 * used by Deliveries, where the records are already loaded.
 */
export function LinearRecordDiagnosis({
  record,
  app,
}: {
  record: LinearProcessingRecord
  app?: LinearApp
}) {
  const diagnosis = diagnoseRecord(record, app)
  const tone =
    diagnosis.tone === 'failed' || diagnosis.tone === 'uncertain'
      ? 'text-destructive'
      : diagnosis.tone === 'success'
        ? 'text-emerald-700'
        : ''
  return (
    <div
      className='space-y-1'
      data-testid={`linear-record-diagnosis-${record.id}`}
    >
      <div className={`text-sm font-medium ${tone}`}>{diagnosis.title}</div>
      {diagnosis.reason && (
        <p className='text-sm break-words whitespace-pre-wrap text-muted-foreground'>
          {diagnosis.reason}
        </p>
      )}
      {diagnosis.nextStep && (
        <p className='text-sm'>Next: {diagnosis.nextStep}</p>
      )}
    </div>
  )
}

/** A record-level note about the Agent or configuration that produced it. */
export function LinearRecordContextNote({
  record,
  app,
}: {
  record: LinearProcessingRecord
  app?: LinearApp
}) {
  if (!app) return null
  const notes: string[] = []
  if (record.agentId && record.agentId !== app.agentId) {
    notes.push(
      'This record belongs to an Agent this App routed to earlier, not to the current Agent.'
    )
  }
  if (record.appRevision > 0n && record.appRevision !== app.revision) {
    notes.push(
      `Accepted under App revision ${String(record.appRevision)}; the current configuration is revision ${String(app.revision)}.`
    )
  }
  if (!record.installationId) {
    notes.push('No installation was recorded for this delivery.')
  }
  if (notes.length === 0) return null
  return (
    <ul className='list-disc pl-5 text-xs text-muted-foreground'>
      {notes.map((note) => (
        <li key={note}>{note}</li>
      ))}
    </ul>
  )
}
