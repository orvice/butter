// This is a generated file - do not edit.
//
// Generated from agents/v1/repobinding.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'repobinding.pb.dart' as $1;
import 'repobinding.pbjson.dart';

export 'repobinding.pb.dart';

abstract class WorkspaceRepoBindingServiceBase extends $pb.GeneratedService {
  $async.Future<$1.GetWorkspaceRepoBindingResponse> getWorkspaceRepoBinding(
      $pb.ServerContext ctx, $1.GetWorkspaceRepoBindingRequest request);
  $async.Future<$1.PutWorkspaceRepoBindingResponse> putWorkspaceRepoBinding(
      $pb.ServerContext ctx, $1.PutWorkspaceRepoBindingRequest request);
  $async.Future<$1.DeleteWorkspaceRepoBindingResponse>
      deleteWorkspaceRepoBinding(
          $pb.ServerContext ctx, $1.DeleteWorkspaceRepoBindingRequest request);
  $async.Future<$1.OnboardWorkspaceRepositoryResponse>
      onboardWorkspaceRepository(
          $pb.ServerContext ctx, $1.OnboardWorkspaceRepositoryRequest request);
  $async.Future<$1.SetWorkspaceRepoBindingCredentialResponse>
      setWorkspaceRepoBindingCredential($pb.ServerContext ctx,
          $1.SetWorkspaceRepoBindingCredentialRequest request);
  $async.Future<$1.ValidateWorkspaceRepoBindingResponse>
      validateWorkspaceRepoBinding($pb.ServerContext ctx,
          $1.ValidateWorkspaceRepoBindingRequest request);
  $async.Future<$1.SyncWorkspaceRepositoryResponse> syncWorkspaceRepository(
      $pb.ServerContext ctx, $1.SyncWorkspaceRepositoryRequest request);
  $async.Future<$1.GetRepositorySyncStatusResponse> getRepositorySyncStatus(
      $pb.ServerContext ctx, $1.GetRepositorySyncStatusRequest request);
  $async.Future<$1.ListRepositoryEntriesResponse> listRepositoryEntries(
      $pb.ServerContext ctx, $1.ListRepositoryEntriesRequest request);
  $async.Future<$1.GetRepositoryFileResponse> getRepositoryFile(
      $pb.ServerContext ctx, $1.GetRepositoryFileRequest request);
  $async.Future<$1.PublishWorkspaceRepositoryResponse>
      publishWorkspaceRepository(
          $pb.ServerContext ctx, $1.PublishWorkspaceRepositoryRequest request);
  $async.Future<$1.ConfigureWebhookSecretResponse> configureWebhookSecret(
      $pb.ServerContext ctx, $1.ConfigureWebhookSecretRequest request);
  $async.Future<$1.AcceptRepositoryBaselineResponse> acceptRepositoryBaseline(
      $pb.ServerContext ctx, $1.AcceptRepositoryBaselineRequest request);
  $async.Future<$1.CommitAgentContentResponse> commitAgentContent(
      $pb.ServerContext ctx, $1.CommitAgentContentRequest request);
  $async.Future<$1.RollbackAgentContentResponse> rollbackAgentContent(
      $pb.ServerContext ctx, $1.RollbackAgentContentRequest request);
  $async.Future<$1.PurgeAgentContentResponse> purgeAgentContent(
      $pb.ServerContext ctx, $1.PurgeAgentContentRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetWorkspaceRepoBinding':
        return $1.GetWorkspaceRepoBindingRequest();
      case 'PutWorkspaceRepoBinding':
        return $1.PutWorkspaceRepoBindingRequest();
      case 'DeleteWorkspaceRepoBinding':
        return $1.DeleteWorkspaceRepoBindingRequest();
      case 'OnboardWorkspaceRepository':
        return $1.OnboardWorkspaceRepositoryRequest();
      case 'SetWorkspaceRepoBindingCredential':
        return $1.SetWorkspaceRepoBindingCredentialRequest();
      case 'ValidateWorkspaceRepoBinding':
        return $1.ValidateWorkspaceRepoBindingRequest();
      case 'SyncWorkspaceRepository':
        return $1.SyncWorkspaceRepositoryRequest();
      case 'GetRepositorySyncStatus':
        return $1.GetRepositorySyncStatusRequest();
      case 'ListRepositoryEntries':
        return $1.ListRepositoryEntriesRequest();
      case 'GetRepositoryFile':
        return $1.GetRepositoryFileRequest();
      case 'PublishWorkspaceRepository':
        return $1.PublishWorkspaceRepositoryRequest();
      case 'ConfigureWebhookSecret':
        return $1.ConfigureWebhookSecretRequest();
      case 'AcceptRepositoryBaseline':
        return $1.AcceptRepositoryBaselineRequest();
      case 'CommitAgentContent':
        return $1.CommitAgentContentRequest();
      case 'RollbackAgentContent':
        return $1.RollbackAgentContentRequest();
      case 'PurgeAgentContent':
        return $1.PurgeAgentContentRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetWorkspaceRepoBinding':
        return getWorkspaceRepoBinding(
            ctx, request as $1.GetWorkspaceRepoBindingRequest);
      case 'PutWorkspaceRepoBinding':
        return putWorkspaceRepoBinding(
            ctx, request as $1.PutWorkspaceRepoBindingRequest);
      case 'DeleteWorkspaceRepoBinding':
        return deleteWorkspaceRepoBinding(
            ctx, request as $1.DeleteWorkspaceRepoBindingRequest);
      case 'OnboardWorkspaceRepository':
        return onboardWorkspaceRepository(
            ctx, request as $1.OnboardWorkspaceRepositoryRequest);
      case 'SetWorkspaceRepoBindingCredential':
        return setWorkspaceRepoBindingCredential(
            ctx, request as $1.SetWorkspaceRepoBindingCredentialRequest);
      case 'ValidateWorkspaceRepoBinding':
        return validateWorkspaceRepoBinding(
            ctx, request as $1.ValidateWorkspaceRepoBindingRequest);
      case 'SyncWorkspaceRepository':
        return syncWorkspaceRepository(
            ctx, request as $1.SyncWorkspaceRepositoryRequest);
      case 'GetRepositorySyncStatus':
        return getRepositorySyncStatus(
            ctx, request as $1.GetRepositorySyncStatusRequest);
      case 'ListRepositoryEntries':
        return listRepositoryEntries(
            ctx, request as $1.ListRepositoryEntriesRequest);
      case 'GetRepositoryFile':
        return getRepositoryFile(ctx, request as $1.GetRepositoryFileRequest);
      case 'PublishWorkspaceRepository':
        return publishWorkspaceRepository(
            ctx, request as $1.PublishWorkspaceRepositoryRequest);
      case 'ConfigureWebhookSecret':
        return configureWebhookSecret(
            ctx, request as $1.ConfigureWebhookSecretRequest);
      case 'AcceptRepositoryBaseline':
        return acceptRepositoryBaseline(
            ctx, request as $1.AcceptRepositoryBaselineRequest);
      case 'CommitAgentContent':
        return commitAgentContent(ctx, request as $1.CommitAgentContentRequest);
      case 'RollbackAgentContent':
        return rollbackAgentContent(
            ctx, request as $1.RollbackAgentContentRequest);
      case 'PurgeAgentContent':
        return purgeAgentContent(ctx, request as $1.PurgeAgentContentRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      WorkspaceRepoBindingServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => WorkspaceRepoBindingServiceBase$messageJson;
}
