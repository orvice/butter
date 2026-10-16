//
//  Generated code. Do not modify.
//  source: agents/v1/telegram.proto
//

import "package:connectrpc/connect.dart" as connect;
import "telegram.pb.dart" as agentsv1telegram;
import "telegram.connect.spec.dart" as specs;

/// =============================================================================
/// TelegramChannelService
/// Reads are open to every workspace member (the auth middleware already
/// verified membership for the X-Workspace-ID header). Mutations require the
/// caller to hold the workspace "owner" or "admin" role; global admins bypass.
/// No RPC ever returns a Bot Token.
/// =============================================================================
extension type TelegramChannelServiceClient (connect.Transport _transport) {
  /// ListTelegramChannels returns every Channel in the current workspace.
  Future<agentsv1telegram.ListTelegramChannelsResponse> listTelegramChannels(
    agentsv1telegram.ListTelegramChannelsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.listTelegramChannels,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetTelegramChannel returns one Channel by ID.
  Future<agentsv1telegram.GetTelegramChannelResponse> getTelegramChannel(
    agentsv1telegram.GetTelegramChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.getTelegramChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateTelegramChannel validates the supplied Bot Token with Telegram
  /// `getMe` before any Channel or credential record is committed, then stores
  /// the resolved immutable Bot identity. Channels are always created with
  /// inbound and outbound disabled; enablement is a separate action.
  Future<agentsv1telegram.CreateTelegramChannelResponse> createTelegramChannel(
    agentsv1telegram.CreateTelegramChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.createTelegramChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateTelegramChannel changes the mutable configuration (name, receive
  /// mode). It requires the expected revision and rejects changes to the
  /// immutable key and Bot identity.
  Future<agentsv1telegram.UpdateTelegramChannelResponse> updateTelegramChannel(
    agentsv1telegram.UpdateTelegramChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.updateTelegramChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RotateTelegramChannelCredential validates the new Bot Token first and
  /// accepts it only when it resolves to the same immutable Bot ID. The prior
  /// credential stays active if validation or persistence fails.
  Future<agentsv1telegram.RotateTelegramChannelCredentialResponse> rotateTelegramChannelCredential(
    agentsv1telegram.RotateTelegramChannelCredentialRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.rotateTelegramChannelCredential,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SetTelegramChannelEnabled applies the desired inbound/outbound state
  /// after a strict preflight. Enabling inbound without outbound is rejected.
  Future<agentsv1telegram.SetTelegramChannelEnabledResponse> setTelegramChannelEnabled(
    agentsv1telegram.SetTelegramChannelEnabledRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.setTelegramChannelEnabled,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteTelegramChannel removes a Channel and its stored credential. It is
  /// refused while any Destination still references the Channel.
  Future<agentsv1telegram.DeleteTelegramChannelResponse> deleteTelegramChannel(
    agentsv1telegram.DeleteTelegramChannelRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.deleteTelegramChannel,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetTelegramChannelStatus returns configured desired state plus derived
  /// runtime readiness. Runtime availability is never persisted as
  /// configuration.
  Future<agentsv1telegram.GetTelegramChannelStatusResponse> getTelegramChannelStatus(
    agentsv1telegram.GetTelegramChannelStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramChannelService.getTelegramChannelStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// =============================================================================
/// TelegramDestinationService
/// =============================================================================
extension type TelegramDestinationServiceClient (connect.Transport _transport) {
  /// ListTelegramDestinations returns the workspace's Destinations, optionally
  /// narrowed to one Channel.
  Future<agentsv1telegram.ListTelegramDestinationsResponse> listTelegramDestinations(
    agentsv1telegram.ListTelegramDestinationsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramDestinationService.listTelegramDestinations,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetTelegramDestination returns one Destination by ID.
  Future<agentsv1telegram.GetTelegramDestinationResponse> getTelegramDestination(
    agentsv1telegram.GetTelegramDestinationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramDestinationService.getTelegramDestination,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateTelegramDestination registers one exact Telegram address. Creating
  /// a Destination sends no Telegram message.
  Future<agentsv1telegram.CreateTelegramDestinationResponse> createTelegramDestination(
    agentsv1telegram.CreateTelegramDestinationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramDestinationService.createTelegramDestination,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateTelegramDestination changes name, enablement, and policy. It
  /// requires the expected revision and rejects any change to the immutable
  /// address; redirecting traffic requires creating a new Destination.
  Future<agentsv1telegram.UpdateTelegramDestinationResponse> updateTelegramDestination(
    agentsv1telegram.UpdateTelegramDestinationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramDestinationService.updateTelegramDestination,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteTelegramDestination removes a Destination. It is refused while a
  /// Notify Group or Cron job still references it.
  Future<agentsv1telegram.DeleteTelegramDestinationResponse> deleteTelegramDestination(
    agentsv1telegram.DeleteTelegramDestinationRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramDestinationService.deleteTelegramDestination,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SendTelegramTestMessage delivers a message to the Destination through the
  /// unified sender. It is the explicit action that proves an address works:
  /// creating a Destination never sends anything. A successful send marks an
  /// unverified Destination verified.
  Future<agentsv1telegram.SendTelegramTestMessageResponse> sendTelegramTestMessage(
    agentsv1telegram.SendTelegramTestMessageRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramDestinationService.sendTelegramTestMessage,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// =============================================================================
/// TelegramAdminService manages platform-level Telegram settings.
/// The Webhook base URL is global, not per-workspace: it names the public
/// address of this deployment behind its load balancer, which is a platform
/// fact rather than a tenant choice. Only global administrators may read or
/// change it, so workspace input can never redirect Telegram callbacks.
/// =============================================================================
extension type TelegramAdminServiceClient (connect.Transport _transport) {
  /// GetTelegramSettings returns the platform Telegram settings.
  Future<agentsv1telegram.GetTelegramSettingsResponse> getTelegramSettings(
    agentsv1telegram.GetTelegramSettingsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramAdminService.getTelegramSettings,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateTelegramSettings replaces the platform Telegram settings.
  Future<agentsv1telegram.UpdateTelegramSettingsResponse> updateTelegramSettings(
    agentsv1telegram.UpdateTelegramSettingsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramAdminService.updateTelegramSettings,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// =============================================================================
/// TelegramProcessingService
/// Workspace members may read processing records and redacted failures for
/// their workspace; owners/admins may resend an already-produced reply.
/// There is deliberately no Agent rerun action: once Agent work may have
/// produced side effects, repeating it is an operator decision made with
/// context this API does not have.
/// =============================================================================
extension type TelegramProcessingServiceClient (connect.Transport _transport) {
  Future<agentsv1telegram.ListTelegramProcessingRecordsResponse> listTelegramProcessingRecords(
    agentsv1telegram.ListTelegramProcessingRecordsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramProcessingService.listTelegramProcessingRecords,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1telegram.GetTelegramProcessingRecordResponse> getTelegramProcessingRecord(
    agentsv1telegram.GetTelegramProcessingRecordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramProcessingService.getTelegramProcessingRecord,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ResendTelegramReply re-delivers unsent or failed segments of a response
  /// that was already produced. It never invokes the Agent, and it is refused
  /// when no complete output exists.
  Future<agentsv1telegram.ResendTelegramReplyResponse> resendTelegramReply(
    agentsv1telegram.ResendTelegramReplyRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TelegramProcessingService.resendTelegramReply,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
