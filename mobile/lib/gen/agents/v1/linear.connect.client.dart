//
//  Generated code. Do not modify.
//  source: agents/v1/linear.proto
//

import "package:connectrpc/connect.dart" as connect;
import "linear.pb.dart" as agentsv1linear;
import "linear.connect.spec.dart" as specs;

/// LinearAppService manages a workspace's Linear Apps. Members may read;
/// workspace owners and admins manage.
extension type LinearAppServiceClient (connect.Transport _transport) {
  Future<agentsv1linear.ListLinearAppsResponse> listLinearApps(
    agentsv1linear.ListLinearAppsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.listLinearApps,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1linear.GetLinearAppResponse> getLinearApp(
    agentsv1linear.GetLinearAppRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.getLinearApp,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateLinearApp registers a Linear OAuth app routed to one Agent.
  Future<agentsv1linear.CreateLinearAppResponse> createLinearApp(
    agentsv1linear.CreateLinearAppRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.createLinearApp,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateLinearApp replaces the mutable fields under the optimistic
  /// revision check. The ID and client ID are immutable.
  Future<agentsv1linear.UpdateLinearAppResponse> updateLinearApp(
    agentsv1linear.UpdateLinearAppRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.updateLinearApp,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// PutLinearAppCredentials sets or clears the write-only secrets.
  Future<agentsv1linear.PutLinearAppCredentialsResponse> putLinearAppCredentials(
    agentsv1linear.PutLinearAppCredentialsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.putLinearAppCredentials,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteLinearApp removes the App together with its installations.
  Future<agentsv1linear.DeleteLinearAppResponse> deleteLinearApp(
    agentsv1linear.DeleteLinearAppRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.deleteLinearApp,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// BeginLinearInstall starts installing the App into a Linear organization:
  /// it returns Linear's authorize URL, carrying a single-use state that the
  /// public OAuth callback consumes.
  Future<agentsv1linear.BeginLinearInstallResponse> beginLinearInstall(
    agentsv1linear.BeginLinearInstallRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.beginLinearInstall,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1linear.ListLinearInstallationsResponse> listLinearInstallations(
    agentsv1linear.ListLinearInstallationsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.listLinearInstallations,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteLinearInstallation removes an installation, revoking its token at
  /// Linear as a best effort.
  Future<agentsv1linear.DeleteLinearInstallationResponse> deleteLinearInstallation(
    agentsv1linear.DeleteLinearInstallationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAppService.deleteLinearInstallation,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// =============================================================================
/// LinearAdminService manages platform-level Linear settings.
/// The public base URL is global, not per-workspace: it names the public
/// address of this deployment. Only global administrators may read or change
/// it, so workspace input can never redirect Linear webhooks or OAuth
/// callbacks.
/// =============================================================================
extension type LinearAdminServiceClient (connect.Transport _transport) {
  Future<agentsv1linear.GetLinearSettingsResponse> getLinearSettings(
    agentsv1linear.GetLinearSettingsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAdminService.getLinearSettings,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1linear.UpdateLinearSettingsResponse> updateLinearSettings(
    agentsv1linear.UpdateLinearSettingsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearAdminService.updateLinearSettings,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// LinearProcessingService exposes processing records. Members read;
/// owners and admins resend a reply that was produced but not posted. There
/// is deliberately no rerun action.
extension type LinearProcessingServiceClient (connect.Transport _transport) {
  Future<agentsv1linear.ListLinearProcessingRecordsResponse> listLinearProcessingRecords(
    agentsv1linear.ListLinearProcessingRecordsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearProcessingService.listLinearProcessingRecords,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1linear.GetLinearProcessingRecordResponse> getLinearProcessingRecord(
    agentsv1linear.GetLinearProcessingRecordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearProcessingService.getLinearProcessingRecord,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ResendLinearReply posts the persisted reply of a FAILED record. It never
  /// invokes the Agent, and it is refused when no reply was persisted.
  Future<agentsv1linear.ResendLinearReplyResponse> resendLinearReply(
    agentsv1linear.ResendLinearReplyRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.LinearProcessingService.resendLinearReply,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
