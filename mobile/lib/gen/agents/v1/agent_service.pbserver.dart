// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_service.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'agent_service.pb.dart' as $10;
import 'agent_service.pbjson.dart';
import 'dashboard.pb.dart' as $8;

export 'agent_service.pb.dart';

abstract class AgentServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListAgentsResponse> listAgents(
      $pb.ServerContext ctx, $10.ListAgentsRequest request);
  $async.Future<$10.GetAgentResponse> getAgent(
      $pb.ServerContext ctx, $10.GetAgentRequest request);
  $async.Future<$10.CreateAgentResponse> createAgent(
      $pb.ServerContext ctx, $10.CreateAgentRequest request);
  $async.Future<$10.UpdateAgentResponse> updateAgent(
      $pb.ServerContext ctx, $10.UpdateAgentRequest request);
  $async.Future<$10.DeleteAgentResponse> deleteAgent(
      $pb.ServerContext ctx, $10.DeleteAgentRequest request);
  $async.Future<$10.InvokeAgentResponse> invokeAgent(
      $pb.ServerContext ctx, $10.InvokeAgentRequest request);
  $async.Future<$10.ListAgentInvocationsResponse> listAgentInvocations(
      $pb.ServerContext ctx, $10.ListAgentInvocationsRequest request);
  $async.Future<$10.ReloadAgentsResponse> reloadAgents(
      $pb.ServerContext ctx, $10.ReloadAgentsRequest request);
  $async.Future<$10.GetAgentRuntimeStatusResponse> getAgentRuntimeStatus(
      $pb.ServerContext ctx, $10.GetAgentRuntimeStatusRequest request);
  $async.Future<$10.ListAgentRuntimeStatusesResponse> listAgentRuntimeStatuses(
      $pb.ServerContext ctx, $10.ListAgentRuntimeStatusesRequest request);
  $async.Future<$10.CancelAgentInvocationResponse> cancelAgentInvocation(
      $pb.ServerContext ctx, $10.CancelAgentInvocationRequest request);
  $async.Future<$10.StreamAgentResponse> streamAgent(
      $pb.ServerContext ctx, $10.StreamAgentRequest request);
  $async.Future<$10.AssignAgentIDResponse> assignAgentID(
      $pb.ServerContext ctx, $10.AssignAgentIDRequest request);
  $async.Future<$10.GetMigrationReadinessResponse> getMigrationReadiness(
      $pb.ServerContext ctx, $10.GetMigrationReadinessRequest request);
  $async.Future<$10.MigrateAgentsV2Response> migrateAgentsV2(
      $pb.ServerContext ctx, $10.MigrateAgentsV2Request request);
  $async.Future<$10.VerifyAgentIDCutoverResponse> verifyAgentIDCutover(
      $pb.ServerContext ctx, $10.VerifyAgentIDCutoverRequest request);
  $async.Future<$10.UpdateAgentConfigurationResponse> updateAgentConfiguration(
      $pb.ServerContext ctx, $10.UpdateAgentConfigurationRequest request);
  $async.Future<$10.RestoreAgentResponse> restoreAgent(
      $pb.ServerContext ctx, $10.RestoreAgentRequest request);
  $async.Future<$10.GetAgentInvocationResponse> getAgentInvocation(
      $pb.ServerContext ctx, $10.GetAgentInvocationRequest request);
  $async.Future<$10.GetAgentOperationResponse> getAgentOperation(
      $pb.ServerContext ctx, $10.GetAgentOperationRequest request);
  $async.Future<$10.ListAgentOperationsResponse> listAgentOperations(
      $pb.ServerContext ctx, $10.ListAgentOperationsRequest request);
  $async.Future<$10.RetryAgentOperationResponse> retryAgentOperation(
      $pb.ServerContext ctx, $10.RetryAgentOperationRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListAgents':
        return $10.ListAgentsRequest();
      case 'GetAgent':
        return $10.GetAgentRequest();
      case 'CreateAgent':
        return $10.CreateAgentRequest();
      case 'UpdateAgent':
        return $10.UpdateAgentRequest();
      case 'DeleteAgent':
        return $10.DeleteAgentRequest();
      case 'InvokeAgent':
        return $10.InvokeAgentRequest();
      case 'ListAgentInvocations':
        return $10.ListAgentInvocationsRequest();
      case 'ReloadAgents':
        return $10.ReloadAgentsRequest();
      case 'GetAgentRuntimeStatus':
        return $10.GetAgentRuntimeStatusRequest();
      case 'ListAgentRuntimeStatuses':
        return $10.ListAgentRuntimeStatusesRequest();
      case 'CancelAgentInvocation':
        return $10.CancelAgentInvocationRequest();
      case 'StreamAgent':
        return $10.StreamAgentRequest();
      case 'AssignAgentID':
        return $10.AssignAgentIDRequest();
      case 'GetMigrationReadiness':
        return $10.GetMigrationReadinessRequest();
      case 'MigrateAgentsV2':
        return $10.MigrateAgentsV2Request();
      case 'VerifyAgentIDCutover':
        return $10.VerifyAgentIDCutoverRequest();
      case 'UpdateAgentConfiguration':
        return $10.UpdateAgentConfigurationRequest();
      case 'RestoreAgent':
        return $10.RestoreAgentRequest();
      case 'GetAgentInvocation':
        return $10.GetAgentInvocationRequest();
      case 'GetAgentOperation':
        return $10.GetAgentOperationRequest();
      case 'ListAgentOperations':
        return $10.ListAgentOperationsRequest();
      case 'RetryAgentOperation':
        return $10.RetryAgentOperationRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListAgents':
        return listAgents(ctx, request as $10.ListAgentsRequest);
      case 'GetAgent':
        return getAgent(ctx, request as $10.GetAgentRequest);
      case 'CreateAgent':
        return createAgent(ctx, request as $10.CreateAgentRequest);
      case 'UpdateAgent':
        return updateAgent(ctx, request as $10.UpdateAgentRequest);
      case 'DeleteAgent':
        return deleteAgent(ctx, request as $10.DeleteAgentRequest);
      case 'InvokeAgent':
        return invokeAgent(ctx, request as $10.InvokeAgentRequest);
      case 'ListAgentInvocations':
        return listAgentInvocations(
            ctx, request as $10.ListAgentInvocationsRequest);
      case 'ReloadAgents':
        return reloadAgents(ctx, request as $10.ReloadAgentsRequest);
      case 'GetAgentRuntimeStatus':
        return getAgentRuntimeStatus(
            ctx, request as $10.GetAgentRuntimeStatusRequest);
      case 'ListAgentRuntimeStatuses':
        return listAgentRuntimeStatuses(
            ctx, request as $10.ListAgentRuntimeStatusesRequest);
      case 'CancelAgentInvocation':
        return cancelAgentInvocation(
            ctx, request as $10.CancelAgentInvocationRequest);
      case 'StreamAgent':
        return streamAgent(ctx, request as $10.StreamAgentRequest);
      case 'AssignAgentID':
        return assignAgentID(ctx, request as $10.AssignAgentIDRequest);
      case 'GetMigrationReadiness':
        return getMigrationReadiness(
            ctx, request as $10.GetMigrationReadinessRequest);
      case 'MigrateAgentsV2':
        return migrateAgentsV2(ctx, request as $10.MigrateAgentsV2Request);
      case 'VerifyAgentIDCutover':
        return verifyAgentIDCutover(
            ctx, request as $10.VerifyAgentIDCutoverRequest);
      case 'UpdateAgentConfiguration':
        return updateAgentConfiguration(
            ctx, request as $10.UpdateAgentConfigurationRequest);
      case 'RestoreAgent':
        return restoreAgent(ctx, request as $10.RestoreAgentRequest);
      case 'GetAgentInvocation':
        return getAgentInvocation(
            ctx, request as $10.GetAgentInvocationRequest);
      case 'GetAgentOperation':
        return getAgentOperation(ctx, request as $10.GetAgentOperationRequest);
      case 'ListAgentOperations':
        return listAgentOperations(
            ctx, request as $10.ListAgentOperationsRequest);
      case 'RetryAgentOperation':
        return retryAgentOperation(
            ctx, request as $10.RetryAgentOperationRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => AgentServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => AgentServiceBase$messageJson;
}

abstract class MCPServerServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListMCPServersResponse> listMCPServers(
      $pb.ServerContext ctx, $10.ListMCPServersRequest request);
  $async.Future<$10.GetMCPServerResponse> getMCPServer(
      $pb.ServerContext ctx, $10.GetMCPServerRequest request);
  $async.Future<$10.CreateMCPServerResponse> createMCPServer(
      $pb.ServerContext ctx, $10.CreateMCPServerRequest request);
  $async.Future<$10.UpdateMCPServerResponse> updateMCPServer(
      $pb.ServerContext ctx, $10.UpdateMCPServerRequest request);
  $async.Future<$10.DeleteMCPServerResponse> deleteMCPServer(
      $pb.ServerContext ctx, $10.DeleteMCPServerRequest request);
  $async.Future<$8.GetMCPServerStatusResponse> getMCPServerStatus(
      $pb.ServerContext ctx, $8.GetMCPServerStatusRequest request);
  $async.Future<$8.ListMCPToolsResponse> listMCPTools(
      $pb.ServerContext ctx, $8.ListMCPToolsRequest request);
  $async.Future<$10.StartMCPServerOAuthResponse> startMCPServerOAuth(
      $pb.ServerContext ctx, $10.StartMCPServerOAuthRequest request);
  $async.Future<$10.CompleteMCPServerOAuthResponse> completeMCPServerOAuth(
      $pb.ServerContext ctx, $10.CompleteMCPServerOAuthRequest request);
  $async.Future<$10.GetMCPServerOAuthStatusResponse> getMCPServerOAuthStatus(
      $pb.ServerContext ctx, $10.GetMCPServerOAuthStatusRequest request);
  $async.Future<$10.DisconnectMCPServerOAuthResponse> disconnectMCPServerOAuth(
      $pb.ServerContext ctx, $10.DisconnectMCPServerOAuthRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListMCPServers':
        return $10.ListMCPServersRequest();
      case 'GetMCPServer':
        return $10.GetMCPServerRequest();
      case 'CreateMCPServer':
        return $10.CreateMCPServerRequest();
      case 'UpdateMCPServer':
        return $10.UpdateMCPServerRequest();
      case 'DeleteMCPServer':
        return $10.DeleteMCPServerRequest();
      case 'GetMCPServerStatus':
        return $8.GetMCPServerStatusRequest();
      case 'ListMCPTools':
        return $8.ListMCPToolsRequest();
      case 'StartMCPServerOAuth':
        return $10.StartMCPServerOAuthRequest();
      case 'CompleteMCPServerOAuth':
        return $10.CompleteMCPServerOAuthRequest();
      case 'GetMCPServerOAuthStatus':
        return $10.GetMCPServerOAuthStatusRequest();
      case 'DisconnectMCPServerOAuth':
        return $10.DisconnectMCPServerOAuthRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListMCPServers':
        return listMCPServers(ctx, request as $10.ListMCPServersRequest);
      case 'GetMCPServer':
        return getMCPServer(ctx, request as $10.GetMCPServerRequest);
      case 'CreateMCPServer':
        return createMCPServer(ctx, request as $10.CreateMCPServerRequest);
      case 'UpdateMCPServer':
        return updateMCPServer(ctx, request as $10.UpdateMCPServerRequest);
      case 'DeleteMCPServer':
        return deleteMCPServer(ctx, request as $10.DeleteMCPServerRequest);
      case 'GetMCPServerStatus':
        return getMCPServerStatus(ctx, request as $8.GetMCPServerStatusRequest);
      case 'ListMCPTools':
        return listMCPTools(ctx, request as $8.ListMCPToolsRequest);
      case 'StartMCPServerOAuth':
        return startMCPServerOAuth(
            ctx, request as $10.StartMCPServerOAuthRequest);
      case 'CompleteMCPServerOAuth':
        return completeMCPServerOAuth(
            ctx, request as $10.CompleteMCPServerOAuthRequest);
      case 'GetMCPServerOAuthStatus':
        return getMCPServerOAuthStatus(
            ctx, request as $10.GetMCPServerOAuthStatusRequest);
      case 'DisconnectMCPServerOAuth':
        return disconnectMCPServerOAuth(
            ctx, request as $10.DisconnectMCPServerOAuthRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => MCPServerServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => MCPServerServiceBase$messageJson;
}

abstract class GlobalMCPServerServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListGlobalMCPServersResponse> listGlobalMCPServers(
      $pb.ServerContext ctx, $10.ListGlobalMCPServersRequest request);
  $async.Future<$10.CreateGlobalMCPServerResponse> createGlobalMCPServer(
      $pb.ServerContext ctx, $10.CreateGlobalMCPServerRequest request);
  $async.Future<$10.UpdateGlobalMCPServerResponse> updateGlobalMCPServer(
      $pb.ServerContext ctx, $10.UpdateGlobalMCPServerRequest request);
  $async.Future<$10.DeleteGlobalMCPServerResponse> deleteGlobalMCPServer(
      $pb.ServerContext ctx, $10.DeleteGlobalMCPServerRequest request);
  $async.Future<$10.InstallGlobalMCPServerResponse> installGlobalMCPServer(
      $pb.ServerContext ctx, $10.InstallGlobalMCPServerRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListGlobalMCPServers':
        return $10.ListGlobalMCPServersRequest();
      case 'CreateGlobalMCPServer':
        return $10.CreateGlobalMCPServerRequest();
      case 'UpdateGlobalMCPServer':
        return $10.UpdateGlobalMCPServerRequest();
      case 'DeleteGlobalMCPServer':
        return $10.DeleteGlobalMCPServerRequest();
      case 'InstallGlobalMCPServer':
        return $10.InstallGlobalMCPServerRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListGlobalMCPServers':
        return listGlobalMCPServers(
            ctx, request as $10.ListGlobalMCPServersRequest);
      case 'CreateGlobalMCPServer':
        return createGlobalMCPServer(
            ctx, request as $10.CreateGlobalMCPServerRequest);
      case 'UpdateGlobalMCPServer':
        return updateGlobalMCPServer(
            ctx, request as $10.UpdateGlobalMCPServerRequest);
      case 'DeleteGlobalMCPServer':
        return deleteGlobalMCPServer(
            ctx, request as $10.DeleteGlobalMCPServerRequest);
      case 'InstallGlobalMCPServer':
        return installGlobalMCPServer(
            ctx, request as $10.InstallGlobalMCPServerRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      GlobalMCPServerServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => GlobalMCPServerServiceBase$messageJson;
}

abstract class ModelProviderServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListModelProvidersResponse> listModelProviders(
      $pb.ServerContext ctx, $10.ListModelProvidersRequest request);
  $async.Future<$10.GetModelProviderResponse> getModelProvider(
      $pb.ServerContext ctx, $10.GetModelProviderRequest request);
  $async.Future<$10.CreateModelProviderResponse> createModelProvider(
      $pb.ServerContext ctx, $10.CreateModelProviderRequest request);
  $async.Future<$10.UpdateModelProviderResponse> updateModelProvider(
      $pb.ServerContext ctx, $10.UpdateModelProviderRequest request);
  $async.Future<$10.DeleteModelProviderResponse> deleteModelProvider(
      $pb.ServerContext ctx, $10.DeleteModelProviderRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListModelProviders':
        return $10.ListModelProvidersRequest();
      case 'GetModelProvider':
        return $10.GetModelProviderRequest();
      case 'CreateModelProvider':
        return $10.CreateModelProviderRequest();
      case 'UpdateModelProvider':
        return $10.UpdateModelProviderRequest();
      case 'DeleteModelProvider':
        return $10.DeleteModelProviderRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListModelProviders':
        return listModelProviders(
            ctx, request as $10.ListModelProvidersRequest);
      case 'GetModelProvider':
        return getModelProvider(ctx, request as $10.GetModelProviderRequest);
      case 'CreateModelProvider':
        return createModelProvider(
            ctx, request as $10.CreateModelProviderRequest);
      case 'UpdateModelProvider':
        return updateModelProvider(
            ctx, request as $10.UpdateModelProviderRequest);
      case 'DeleteModelProvider':
        return deleteModelProvider(
            ctx, request as $10.DeleteModelProviderRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      ModelProviderServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ModelProviderServiceBase$messageJson;
}

abstract class NotifyGroupServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListNotifyGroupsResponse> listNotifyGroups(
      $pb.ServerContext ctx, $10.ListNotifyGroupsRequest request);
  $async.Future<$10.GetNotifyGroupResponse> getNotifyGroup(
      $pb.ServerContext ctx, $10.GetNotifyGroupRequest request);
  $async.Future<$10.CreateNotifyGroupResponse> createNotifyGroup(
      $pb.ServerContext ctx, $10.CreateNotifyGroupRequest request);
  $async.Future<$10.UpdateNotifyGroupResponse> updateNotifyGroup(
      $pb.ServerContext ctx, $10.UpdateNotifyGroupRequest request);
  $async.Future<$10.DeleteNotifyGroupResponse> deleteNotifyGroup(
      $pb.ServerContext ctx, $10.DeleteNotifyGroupRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListNotifyGroups':
        return $10.ListNotifyGroupsRequest();
      case 'GetNotifyGroup':
        return $10.GetNotifyGroupRequest();
      case 'CreateNotifyGroup':
        return $10.CreateNotifyGroupRequest();
      case 'UpdateNotifyGroup':
        return $10.UpdateNotifyGroupRequest();
      case 'DeleteNotifyGroup':
        return $10.DeleteNotifyGroupRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListNotifyGroups':
        return listNotifyGroups(ctx, request as $10.ListNotifyGroupsRequest);
      case 'GetNotifyGroup':
        return getNotifyGroup(ctx, request as $10.GetNotifyGroupRequest);
      case 'CreateNotifyGroup':
        return createNotifyGroup(ctx, request as $10.CreateNotifyGroupRequest);
      case 'UpdateNotifyGroup':
        return updateNotifyGroup(ctx, request as $10.UpdateNotifyGroupRequest);
      case 'DeleteNotifyGroup':
        return deleteNotifyGroup(ctx, request as $10.DeleteNotifyGroupRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      NotifyGroupServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => NotifyGroupServiceBase$messageJson;
}

abstract class RemoteAgentServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListRemoteAgentsResponse> listRemoteAgents(
      $pb.ServerContext ctx, $10.ListRemoteAgentsRequest request);
  $async.Future<$10.GetRemoteAgentResponse> getRemoteAgent(
      $pb.ServerContext ctx, $10.GetRemoteAgentRequest request);
  $async.Future<$10.CreateRemoteAgentResponse> createRemoteAgent(
      $pb.ServerContext ctx, $10.CreateRemoteAgentRequest request);
  $async.Future<$10.UpdateRemoteAgentResponse> updateRemoteAgent(
      $pb.ServerContext ctx, $10.UpdateRemoteAgentRequest request);
  $async.Future<$10.DeleteRemoteAgentResponse> deleteRemoteAgent(
      $pb.ServerContext ctx, $10.DeleteRemoteAgentRequest request);
  $async.Future<$8.GetRemoteAgentStatusResponse> getRemoteAgentStatus(
      $pb.ServerContext ctx, $8.GetRemoteAgentStatusRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListRemoteAgents':
        return $10.ListRemoteAgentsRequest();
      case 'GetRemoteAgent':
        return $10.GetRemoteAgentRequest();
      case 'CreateRemoteAgent':
        return $10.CreateRemoteAgentRequest();
      case 'UpdateRemoteAgent':
        return $10.UpdateRemoteAgentRequest();
      case 'DeleteRemoteAgent':
        return $10.DeleteRemoteAgentRequest();
      case 'GetRemoteAgentStatus':
        return $8.GetRemoteAgentStatusRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListRemoteAgents':
        return listRemoteAgents(ctx, request as $10.ListRemoteAgentsRequest);
      case 'GetRemoteAgent':
        return getRemoteAgent(ctx, request as $10.GetRemoteAgentRequest);
      case 'CreateRemoteAgent':
        return createRemoteAgent(ctx, request as $10.CreateRemoteAgentRequest);
      case 'UpdateRemoteAgent':
        return updateRemoteAgent(ctx, request as $10.UpdateRemoteAgentRequest);
      case 'DeleteRemoteAgent':
        return deleteRemoteAgent(ctx, request as $10.DeleteRemoteAgentRequest);
      case 'GetRemoteAgentStatus':
        return getRemoteAgentStatus(
            ctx, request as $8.GetRemoteAgentStatusRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      RemoteAgentServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => RemoteAgentServiceBase$messageJson;
}

abstract class ChannelServiceBase extends $pb.GeneratedService {
  $async.Future<$10.ListChannelsResponse> listChannels(
      $pb.ServerContext ctx, $10.ListChannelsRequest request);
  $async.Future<$10.GetChannelResponse> getChannel(
      $pb.ServerContext ctx, $10.GetChannelRequest request);
  $async.Future<$10.CreateChannelResponse> createChannel(
      $pb.ServerContext ctx, $10.CreateChannelRequest request);
  $async.Future<$10.UpdateChannelResponse> updateChannel(
      $pb.ServerContext ctx, $10.UpdateChannelRequest request);
  $async.Future<$10.DeleteChannelResponse> deleteChannel(
      $pb.ServerContext ctx, $10.DeleteChannelRequest request);
  $async.Future<$8.GetChannelStatusResponse> getChannelStatus(
      $pb.ServerContext ctx, $8.GetChannelStatusRequest request);
  $async.Future<$10.RestartChannelResponse> restartChannel(
      $pb.ServerContext ctx, $10.RestartChannelRequest request);
  $async.Future<$10.PauseChannelResponse> pauseChannel(
      $pb.ServerContext ctx, $10.PauseChannelRequest request);
  $async.Future<$10.ResumeChannelResponse> resumeChannel(
      $pb.ServerContext ctx, $10.ResumeChannelRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListChannels':
        return $10.ListChannelsRequest();
      case 'GetChannel':
        return $10.GetChannelRequest();
      case 'CreateChannel':
        return $10.CreateChannelRequest();
      case 'UpdateChannel':
        return $10.UpdateChannelRequest();
      case 'DeleteChannel':
        return $10.DeleteChannelRequest();
      case 'GetChannelStatus':
        return $8.GetChannelStatusRequest();
      case 'RestartChannel':
        return $10.RestartChannelRequest();
      case 'PauseChannel':
        return $10.PauseChannelRequest();
      case 'ResumeChannel':
        return $10.ResumeChannelRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListChannels':
        return listChannels(ctx, request as $10.ListChannelsRequest);
      case 'GetChannel':
        return getChannel(ctx, request as $10.GetChannelRequest);
      case 'CreateChannel':
        return createChannel(ctx, request as $10.CreateChannelRequest);
      case 'UpdateChannel':
        return updateChannel(ctx, request as $10.UpdateChannelRequest);
      case 'DeleteChannel':
        return deleteChannel(ctx, request as $10.DeleteChannelRequest);
      case 'GetChannelStatus':
        return getChannelStatus(ctx, request as $8.GetChannelStatusRequest);
      case 'RestartChannel':
        return restartChannel(ctx, request as $10.RestartChannelRequest);
      case 'PauseChannel':
        return pauseChannel(ctx, request as $10.PauseChannelRequest);
      case 'ResumeChannel':
        return resumeChannel(ctx, request as $10.ResumeChannelRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ChannelServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ChannelServiceBase$messageJson;
}

abstract class SessionServiceBase extends $pb.GeneratedService {
  $async.Future<$10.CreateSessionResponse> createSession(
      $pb.ServerContext ctx, $10.CreateSessionRequest request);
  $async.Future<$10.GetSessionResponse> getSession(
      $pb.ServerContext ctx, $10.GetSessionRequest request);
  $async.Future<$10.ListSessionsResponse> listSessions(
      $pb.ServerContext ctx, $10.ListSessionsRequest request);
  $async.Future<$10.DeleteSessionResponse> deleteSession(
      $pb.ServerContext ctx, $10.DeleteSessionRequest request);
  $async.Future<$10.ReplySessionResponse> replySession(
      $pb.ServerContext ctx, $10.ReplySessionRequest request);
  $async.Future<$10.UpdateSessionTitleResponse> updateSessionTitle(
      $pb.ServerContext ctx, $10.UpdateSessionTitleRequest request);
  $async.Future<$10.GenerateSessionTitleResponse> generateSessionTitle(
      $pb.ServerContext ctx, $10.GenerateSessionTitleRequest request);
  $async.Future<$10.MarkSessionReadResponse> markSessionRead(
      $pb.ServerContext ctx, $10.MarkSessionReadRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateSession':
        return $10.CreateSessionRequest();
      case 'GetSession':
        return $10.GetSessionRequest();
      case 'ListSessions':
        return $10.ListSessionsRequest();
      case 'DeleteSession':
        return $10.DeleteSessionRequest();
      case 'ReplySession':
        return $10.ReplySessionRequest();
      case 'UpdateSessionTitle':
        return $10.UpdateSessionTitleRequest();
      case 'GenerateSessionTitle':
        return $10.GenerateSessionTitleRequest();
      case 'MarkSessionRead':
        return $10.MarkSessionReadRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateSession':
        return createSession(ctx, request as $10.CreateSessionRequest);
      case 'GetSession':
        return getSession(ctx, request as $10.GetSessionRequest);
      case 'ListSessions':
        return listSessions(ctx, request as $10.ListSessionsRequest);
      case 'DeleteSession':
        return deleteSession(ctx, request as $10.DeleteSessionRequest);
      case 'ReplySession':
        return replySession(ctx, request as $10.ReplySessionRequest);
      case 'UpdateSessionTitle':
        return updateSessionTitle(
            ctx, request as $10.UpdateSessionTitleRequest);
      case 'GenerateSessionTitle':
        return generateSessionTitle(
            ctx, request as $10.GenerateSessionTitleRequest);
      case 'MarkSessionRead':
        return markSessionRead(ctx, request as $10.MarkSessionReadRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => SessionServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => SessionServiceBase$messageJson;
}
