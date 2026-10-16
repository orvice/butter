//
//  Generated code. Do not modify.
//  source: agents/v1/telegram.proto
//

import "package:connectrpc/connect.dart" as connect;
import "telegram.pb.dart" as agentsv1telegram;

/// =============================================================================
/// TelegramChannelService
/// Reads are open to every workspace member (the auth middleware already
/// verified membership for the X-Workspace-ID header). Mutations require the
/// caller to hold the workspace "owner" or "admin" role; global admins bypass.
/// No RPC ever returns a Bot Token.
/// =============================================================================
abstract final class TelegramChannelService {
  /// Fully-qualified name of the TelegramChannelService service.
  static const name = 'agents.v1.TelegramChannelService';

  /// ListTelegramChannels returns every Channel in the current workspace.
  static const listTelegramChannels = connect.Spec(
    '/$name/ListTelegramChannels',
    connect.StreamType.unary,
    agentsv1telegram.ListTelegramChannelsRequest.new,
    agentsv1telegram.ListTelegramChannelsResponse.new,
  );

  /// GetTelegramChannel returns one Channel by ID.
  static const getTelegramChannel = connect.Spec(
    '/$name/GetTelegramChannel',
    connect.StreamType.unary,
    agentsv1telegram.GetTelegramChannelRequest.new,
    agentsv1telegram.GetTelegramChannelResponse.new,
  );

  /// CreateTelegramChannel validates the supplied Bot Token with Telegram
  /// `getMe` before any Channel or credential record is committed, then stores
  /// the resolved immutable Bot identity. Channels are always created with
  /// inbound and outbound disabled; enablement is a separate action.
  static const createTelegramChannel = connect.Spec(
    '/$name/CreateTelegramChannel',
    connect.StreamType.unary,
    agentsv1telegram.CreateTelegramChannelRequest.new,
    agentsv1telegram.CreateTelegramChannelResponse.new,
  );

  /// UpdateTelegramChannel changes the mutable configuration (name, receive
  /// mode). It requires the expected revision and rejects changes to the
  /// immutable key and Bot identity.
  static const updateTelegramChannel = connect.Spec(
    '/$name/UpdateTelegramChannel',
    connect.StreamType.unary,
    agentsv1telegram.UpdateTelegramChannelRequest.new,
    agentsv1telegram.UpdateTelegramChannelResponse.new,
  );

  /// RotateTelegramChannelCredential validates the new Bot Token first and
  /// accepts it only when it resolves to the same immutable Bot ID. The prior
  /// credential stays active if validation or persistence fails.
  static const rotateTelegramChannelCredential = connect.Spec(
    '/$name/RotateTelegramChannelCredential',
    connect.StreamType.unary,
    agentsv1telegram.RotateTelegramChannelCredentialRequest.new,
    agentsv1telegram.RotateTelegramChannelCredentialResponse.new,
  );

  /// SetTelegramChannelEnabled applies the desired inbound/outbound state
  /// after a strict preflight. Enabling inbound without outbound is rejected.
  static const setTelegramChannelEnabled = connect.Spec(
    '/$name/SetTelegramChannelEnabled',
    connect.StreamType.unary,
    agentsv1telegram.SetTelegramChannelEnabledRequest.new,
    agentsv1telegram.SetTelegramChannelEnabledResponse.new,
  );

  /// DeleteTelegramChannel removes a Channel and its stored credential. It is
  /// refused while any Destination still references the Channel.
  static const deleteTelegramChannel = connect.Spec(
    '/$name/DeleteTelegramChannel',
    connect.StreamType.unary,
    agentsv1telegram.DeleteTelegramChannelRequest.new,
    agentsv1telegram.DeleteTelegramChannelResponse.new,
  );

  /// GetTelegramChannelStatus returns configured desired state plus derived
  /// runtime readiness. Runtime availability is never persisted as
  /// configuration.
  static const getTelegramChannelStatus = connect.Spec(
    '/$name/GetTelegramChannelStatus',
    connect.StreamType.unary,
    agentsv1telegram.GetTelegramChannelStatusRequest.new,
    agentsv1telegram.GetTelegramChannelStatusResponse.new,
  );
}
/// =============================================================================
/// TelegramDestinationService
/// =============================================================================
abstract final class TelegramDestinationService {
  /// Fully-qualified name of the TelegramDestinationService service.
  static const name = 'agents.v1.TelegramDestinationService';

  /// ListTelegramDestinations returns the workspace's Destinations, optionally
  /// narrowed to one Channel.
  static const listTelegramDestinations = connect.Spec(
    '/$name/ListTelegramDestinations',
    connect.StreamType.unary,
    agentsv1telegram.ListTelegramDestinationsRequest.new,
    agentsv1telegram.ListTelegramDestinationsResponse.new,
  );

  /// GetTelegramDestination returns one Destination by ID.
  static const getTelegramDestination = connect.Spec(
    '/$name/GetTelegramDestination',
    connect.StreamType.unary,
    agentsv1telegram.GetTelegramDestinationRequest.new,
    agentsv1telegram.GetTelegramDestinationResponse.new,
  );

  /// CreateTelegramDestination registers one exact Telegram address. Creating
  /// a Destination sends no Telegram message.
  static const createTelegramDestination = connect.Spec(
    '/$name/CreateTelegramDestination',
    connect.StreamType.unary,
    agentsv1telegram.CreateTelegramDestinationRequest.new,
    agentsv1telegram.CreateTelegramDestinationResponse.new,
  );

  /// UpdateTelegramDestination changes name, enablement, and policy. It
  /// requires the expected revision and rejects any change to the immutable
  /// address; redirecting traffic requires creating a new Destination.
  static const updateTelegramDestination = connect.Spec(
    '/$name/UpdateTelegramDestination',
    connect.StreamType.unary,
    agentsv1telegram.UpdateTelegramDestinationRequest.new,
    agentsv1telegram.UpdateTelegramDestinationResponse.new,
  );

  /// DeleteTelegramDestination removes a Destination. It is refused while a
  /// Notify Group or Cron job still references it.
  static const deleteTelegramDestination = connect.Spec(
    '/$name/DeleteTelegramDestination',
    connect.StreamType.unary,
    agentsv1telegram.DeleteTelegramDestinationRequest.new,
    agentsv1telegram.DeleteTelegramDestinationResponse.new,
  );

  /// SendTelegramTestMessage delivers a message to the Destination through the
  /// unified sender. It is the explicit action that proves an address works:
  /// creating a Destination never sends anything. A successful send marks an
  /// unverified Destination verified.
  static const sendTelegramTestMessage = connect.Spec(
    '/$name/SendTelegramTestMessage',
    connect.StreamType.unary,
    agentsv1telegram.SendTelegramTestMessageRequest.new,
    agentsv1telegram.SendTelegramTestMessageResponse.new,
  );
}
/// =============================================================================
/// TelegramAdminService manages platform-level Telegram settings.
/// The Webhook base URL is global, not per-workspace: it names the public
/// address of this deployment behind its load balancer, which is a platform
/// fact rather than a tenant choice. Only global administrators may read or
/// change it, so workspace input can never redirect Telegram callbacks.
/// =============================================================================
abstract final class TelegramAdminService {
  /// Fully-qualified name of the TelegramAdminService service.
  static const name = 'agents.v1.TelegramAdminService';

  /// GetTelegramSettings returns the platform Telegram settings.
  static const getTelegramSettings = connect.Spec(
    '/$name/GetTelegramSettings',
    connect.StreamType.unary,
    agentsv1telegram.GetTelegramSettingsRequest.new,
    agentsv1telegram.GetTelegramSettingsResponse.new,
  );

  /// UpdateTelegramSettings replaces the platform Telegram settings.
  static const updateTelegramSettings = connect.Spec(
    '/$name/UpdateTelegramSettings',
    connect.StreamType.unary,
    agentsv1telegram.UpdateTelegramSettingsRequest.new,
    agentsv1telegram.UpdateTelegramSettingsResponse.new,
  );
}
/// =============================================================================
/// TelegramProcessingService
/// Workspace members may read processing records and redacted failures for
/// their workspace; owners/admins may resend an already-produced reply.
/// There is deliberately no Agent rerun action: once Agent work may have
/// produced side effects, repeating it is an operator decision made with
/// context this API does not have.
/// =============================================================================
abstract final class TelegramProcessingService {
  /// Fully-qualified name of the TelegramProcessingService service.
  static const name = 'agents.v1.TelegramProcessingService';

  static const listTelegramProcessingRecords = connect.Spec(
    '/$name/ListTelegramProcessingRecords',
    connect.StreamType.unary,
    agentsv1telegram.ListTelegramProcessingRecordsRequest.new,
    agentsv1telegram.ListTelegramProcessingRecordsResponse.new,
  );

  static const getTelegramProcessingRecord = connect.Spec(
    '/$name/GetTelegramProcessingRecord',
    connect.StreamType.unary,
    agentsv1telegram.GetTelegramProcessingRecordRequest.new,
    agentsv1telegram.GetTelegramProcessingRecordResponse.new,
  );

  /// ResendTelegramReply re-delivers unsent or failed segments of a response
  /// that was already produced. It never invokes the Agent, and it is refused
  /// when no complete output exists.
  static const resendTelegramReply = connect.Spec(
    '/$name/ResendTelegramReply',
    connect.StreamType.unary,
    agentsv1telegram.ResendTelegramReplyRequest.new,
    agentsv1telegram.ResendTelegramReplyResponse.new,
  );
}
