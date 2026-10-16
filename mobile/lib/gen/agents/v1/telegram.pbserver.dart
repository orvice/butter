// This is a generated file - do not edit.
//
// Generated from agents/v1/telegram.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'telegram.pb.dart' as $1;
import 'telegram.pbjson.dart';

export 'telegram.pb.dart';

abstract class TelegramChannelServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListTelegramChannelsResponse> listTelegramChannels(
      $pb.ServerContext ctx, $1.ListTelegramChannelsRequest request);
  $async.Future<$1.GetTelegramChannelResponse> getTelegramChannel(
      $pb.ServerContext ctx, $1.GetTelegramChannelRequest request);
  $async.Future<$1.CreateTelegramChannelResponse> createTelegramChannel(
      $pb.ServerContext ctx, $1.CreateTelegramChannelRequest request);
  $async.Future<$1.UpdateTelegramChannelResponse> updateTelegramChannel(
      $pb.ServerContext ctx, $1.UpdateTelegramChannelRequest request);
  $async.Future<$1.RotateTelegramChannelCredentialResponse>
      rotateTelegramChannelCredential($pb.ServerContext ctx,
          $1.RotateTelegramChannelCredentialRequest request);
  $async.Future<$1.SetTelegramChannelEnabledResponse> setTelegramChannelEnabled(
      $pb.ServerContext ctx, $1.SetTelegramChannelEnabledRequest request);
  $async.Future<$1.DeleteTelegramChannelResponse> deleteTelegramChannel(
      $pb.ServerContext ctx, $1.DeleteTelegramChannelRequest request);
  $async.Future<$1.GetTelegramChannelStatusResponse> getTelegramChannelStatus(
      $pb.ServerContext ctx, $1.GetTelegramChannelStatusRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListTelegramChannels':
        return $1.ListTelegramChannelsRequest();
      case 'GetTelegramChannel':
        return $1.GetTelegramChannelRequest();
      case 'CreateTelegramChannel':
        return $1.CreateTelegramChannelRequest();
      case 'UpdateTelegramChannel':
        return $1.UpdateTelegramChannelRequest();
      case 'RotateTelegramChannelCredential':
        return $1.RotateTelegramChannelCredentialRequest();
      case 'SetTelegramChannelEnabled':
        return $1.SetTelegramChannelEnabledRequest();
      case 'DeleteTelegramChannel':
        return $1.DeleteTelegramChannelRequest();
      case 'GetTelegramChannelStatus':
        return $1.GetTelegramChannelStatusRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListTelegramChannels':
        return listTelegramChannels(
            ctx, request as $1.ListTelegramChannelsRequest);
      case 'GetTelegramChannel':
        return getTelegramChannel(ctx, request as $1.GetTelegramChannelRequest);
      case 'CreateTelegramChannel':
        return createTelegramChannel(
            ctx, request as $1.CreateTelegramChannelRequest);
      case 'UpdateTelegramChannel':
        return updateTelegramChannel(
            ctx, request as $1.UpdateTelegramChannelRequest);
      case 'RotateTelegramChannelCredential':
        return rotateTelegramChannelCredential(
            ctx, request as $1.RotateTelegramChannelCredentialRequest);
      case 'SetTelegramChannelEnabled':
        return setTelegramChannelEnabled(
            ctx, request as $1.SetTelegramChannelEnabledRequest);
      case 'DeleteTelegramChannel':
        return deleteTelegramChannel(
            ctx, request as $1.DeleteTelegramChannelRequest);
      case 'GetTelegramChannelStatus':
        return getTelegramChannelStatus(
            ctx, request as $1.GetTelegramChannelStatusRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      TelegramChannelServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => TelegramChannelServiceBase$messageJson;
}

abstract class TelegramDestinationServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListTelegramDestinationsResponse> listTelegramDestinations(
      $pb.ServerContext ctx, $1.ListTelegramDestinationsRequest request);
  $async.Future<$1.GetTelegramDestinationResponse> getTelegramDestination(
      $pb.ServerContext ctx, $1.GetTelegramDestinationRequest request);
  $async.Future<$1.CreateTelegramDestinationResponse> createTelegramDestination(
      $pb.ServerContext ctx, $1.CreateTelegramDestinationRequest request);
  $async.Future<$1.UpdateTelegramDestinationResponse> updateTelegramDestination(
      $pb.ServerContext ctx, $1.UpdateTelegramDestinationRequest request);
  $async.Future<$1.DeleteTelegramDestinationResponse> deleteTelegramDestination(
      $pb.ServerContext ctx, $1.DeleteTelegramDestinationRequest request);
  $async.Future<$1.SendTelegramTestMessageResponse> sendTelegramTestMessage(
      $pb.ServerContext ctx, $1.SendTelegramTestMessageRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListTelegramDestinations':
        return $1.ListTelegramDestinationsRequest();
      case 'GetTelegramDestination':
        return $1.GetTelegramDestinationRequest();
      case 'CreateTelegramDestination':
        return $1.CreateTelegramDestinationRequest();
      case 'UpdateTelegramDestination':
        return $1.UpdateTelegramDestinationRequest();
      case 'DeleteTelegramDestination':
        return $1.DeleteTelegramDestinationRequest();
      case 'SendTelegramTestMessage':
        return $1.SendTelegramTestMessageRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListTelegramDestinations':
        return listTelegramDestinations(
            ctx, request as $1.ListTelegramDestinationsRequest);
      case 'GetTelegramDestination':
        return getTelegramDestination(
            ctx, request as $1.GetTelegramDestinationRequest);
      case 'CreateTelegramDestination':
        return createTelegramDestination(
            ctx, request as $1.CreateTelegramDestinationRequest);
      case 'UpdateTelegramDestination':
        return updateTelegramDestination(
            ctx, request as $1.UpdateTelegramDestinationRequest);
      case 'DeleteTelegramDestination':
        return deleteTelegramDestination(
            ctx, request as $1.DeleteTelegramDestinationRequest);
      case 'SendTelegramTestMessage':
        return sendTelegramTestMessage(
            ctx, request as $1.SendTelegramTestMessageRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      TelegramDestinationServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => TelegramDestinationServiceBase$messageJson;
}

abstract class TelegramAdminServiceBase extends $pb.GeneratedService {
  $async.Future<$1.GetTelegramSettingsResponse> getTelegramSettings(
      $pb.ServerContext ctx, $1.GetTelegramSettingsRequest request);
  $async.Future<$1.UpdateTelegramSettingsResponse> updateTelegramSettings(
      $pb.ServerContext ctx, $1.UpdateTelegramSettingsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetTelegramSettings':
        return $1.GetTelegramSettingsRequest();
      case 'UpdateTelegramSettings':
        return $1.UpdateTelegramSettingsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetTelegramSettings':
        return getTelegramSettings(
            ctx, request as $1.GetTelegramSettingsRequest);
      case 'UpdateTelegramSettings':
        return updateTelegramSettings(
            ctx, request as $1.UpdateTelegramSettingsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      TelegramAdminServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => TelegramAdminServiceBase$messageJson;
}

abstract class TelegramProcessingServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListTelegramProcessingRecordsResponse>
      listTelegramProcessingRecords($pb.ServerContext ctx,
          $1.ListTelegramProcessingRecordsRequest request);
  $async.Future<$1.GetTelegramProcessingRecordResponse>
      getTelegramProcessingRecord(
          $pb.ServerContext ctx, $1.GetTelegramProcessingRecordRequest request);
  $async.Future<$1.ResendTelegramReplyResponse> resendTelegramReply(
      $pb.ServerContext ctx, $1.ResendTelegramReplyRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListTelegramProcessingRecords':
        return $1.ListTelegramProcessingRecordsRequest();
      case 'GetTelegramProcessingRecord':
        return $1.GetTelegramProcessingRecordRequest();
      case 'ResendTelegramReply':
        return $1.ResendTelegramReplyRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListTelegramProcessingRecords':
        return listTelegramProcessingRecords(
            ctx, request as $1.ListTelegramProcessingRecordsRequest);
      case 'GetTelegramProcessingRecord':
        return getTelegramProcessingRecord(
            ctx, request as $1.GetTelegramProcessingRecordRequest);
      case 'ResendTelegramReply':
        return resendTelegramReply(
            ctx, request as $1.ResendTelegramReplyRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      TelegramProcessingServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => TelegramProcessingServiceBase$messageJson;
}
