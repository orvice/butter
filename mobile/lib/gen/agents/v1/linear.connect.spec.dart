//
//  Generated code. Do not modify.
//  source: agents/v1/linear.proto
//

import "package:connectrpc/connect.dart" as connect;
import "linear.pb.dart" as agentsv1linear;

/// LinearAppService manages a workspace's Linear Apps. Members may read;
/// workspace owners and admins manage.
abstract final class LinearAppService {
  /// Fully-qualified name of the LinearAppService service.
  static const name = 'agents.v1.LinearAppService';

  static const listLinearApps = connect.Spec(
    '/$name/ListLinearApps',
    connect.StreamType.unary,
    agentsv1linear.ListLinearAppsRequest.new,
    agentsv1linear.ListLinearAppsResponse.new,
  );

  static const getLinearApp = connect.Spec(
    '/$name/GetLinearApp',
    connect.StreamType.unary,
    agentsv1linear.GetLinearAppRequest.new,
    agentsv1linear.GetLinearAppResponse.new,
  );

  /// CreateLinearApp registers a Linear OAuth app routed to one Agent.
  static const createLinearApp = connect.Spec(
    '/$name/CreateLinearApp',
    connect.StreamType.unary,
    agentsv1linear.CreateLinearAppRequest.new,
    agentsv1linear.CreateLinearAppResponse.new,
  );

  /// UpdateLinearApp replaces the mutable fields under the optimistic
  /// revision check. The ID and client ID are immutable.
  static const updateLinearApp = connect.Spec(
    '/$name/UpdateLinearApp',
    connect.StreamType.unary,
    agentsv1linear.UpdateLinearAppRequest.new,
    agentsv1linear.UpdateLinearAppResponse.new,
  );

  /// PutLinearAppCredentials sets or clears the write-only secrets.
  static const putLinearAppCredentials = connect.Spec(
    '/$name/PutLinearAppCredentials',
    connect.StreamType.unary,
    agentsv1linear.PutLinearAppCredentialsRequest.new,
    agentsv1linear.PutLinearAppCredentialsResponse.new,
  );

  /// DeleteLinearApp removes the App together with its installations.
  static const deleteLinearApp = connect.Spec(
    '/$name/DeleteLinearApp',
    connect.StreamType.unary,
    agentsv1linear.DeleteLinearAppRequest.new,
    agentsv1linear.DeleteLinearAppResponse.new,
  );

  /// BeginLinearInstall starts installing the App into a Linear organization:
  /// it returns Linear's authorize URL, carrying a single-use state that the
  /// public OAuth callback consumes.
  static const beginLinearInstall = connect.Spec(
    '/$name/BeginLinearInstall',
    connect.StreamType.unary,
    agentsv1linear.BeginLinearInstallRequest.new,
    agentsv1linear.BeginLinearInstallResponse.new,
  );

  static const listLinearInstallations = connect.Spec(
    '/$name/ListLinearInstallations',
    connect.StreamType.unary,
    agentsv1linear.ListLinearInstallationsRequest.new,
    agentsv1linear.ListLinearInstallationsResponse.new,
  );

  /// DeleteLinearInstallation removes an installation, revoking its token at
  /// Linear as a best effort.
  static const deleteLinearInstallation = connect.Spec(
    '/$name/DeleteLinearInstallation',
    connect.StreamType.unary,
    agentsv1linear.DeleteLinearInstallationRequest.new,
    agentsv1linear.DeleteLinearInstallationResponse.new,
  );
}

/// =============================================================================
/// LinearAdminService manages platform-level Linear settings.
/// The public base URL is global, not per-workspace: it names the public
/// address of this deployment. Only global administrators may read or change
/// it, so workspace input can never redirect Linear webhooks or OAuth
/// callbacks.
/// =============================================================================
abstract final class LinearAdminService {
  /// Fully-qualified name of the LinearAdminService service.
  static const name = 'agents.v1.LinearAdminService';

  static const getLinearSettings = connect.Spec(
    '/$name/GetLinearSettings',
    connect.StreamType.unary,
    agentsv1linear.GetLinearSettingsRequest.new,
    agentsv1linear.GetLinearSettingsResponse.new,
  );

  static const updateLinearSettings = connect.Spec(
    '/$name/UpdateLinearSettings',
    connect.StreamType.unary,
    agentsv1linear.UpdateLinearSettingsRequest.new,
    agentsv1linear.UpdateLinearSettingsResponse.new,
  );
}

/// LinearProcessingService exposes processing records. Members read;
/// owners and admins resend a reply that was produced but not posted. There
/// is deliberately no rerun action.
abstract final class LinearProcessingService {
  /// Fully-qualified name of the LinearProcessingService service.
  static const name = 'agents.v1.LinearProcessingService';

  static const listLinearProcessingRecords = connect.Spec(
    '/$name/ListLinearProcessingRecords',
    connect.StreamType.unary,
    agentsv1linear.ListLinearProcessingRecordsRequest.new,
    agentsv1linear.ListLinearProcessingRecordsResponse.new,
  );

  static const getLinearProcessingRecord = connect.Spec(
    '/$name/GetLinearProcessingRecord',
    connect.StreamType.unary,
    agentsv1linear.GetLinearProcessingRecordRequest.new,
    agentsv1linear.GetLinearProcessingRecordResponse.new,
  );

  /// ResendLinearReply posts the persisted reply of a FAILED record. It never
  /// invokes the Agent, and it is refused when no reply was persisted.
  static const resendLinearReply = connect.Spec(
    '/$name/ResendLinearReply',
    connect.StreamType.unary,
    agentsv1linear.ResendLinearReplyRequest.new,
    agentsv1linear.ResendLinearReplyResponse.new,
  );
}
