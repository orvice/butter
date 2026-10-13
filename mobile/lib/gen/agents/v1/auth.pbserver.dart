// This is a generated file - do not edit.
//
// Generated from agents/v1/auth.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'auth.pb.dart' as $2;
import 'auth.pbjson.dart';

export 'auth.pb.dart';

abstract class AuthServiceBase extends $pb.GeneratedService {
  $async.Future<$2.LoginResponse> login(
      $pb.ServerContext ctx, $2.LoginRequest request);
  $async.Future<$2.MeResponse> me($pb.ServerContext ctx, $2.MeRequest request);
  $async.Future<$2.LogoutResponse> logout(
      $pb.ServerContext ctx, $2.LogoutRequest request);
  $async.Future<$2.ListUsersResponse> listUsers(
      $pb.ServerContext ctx, $2.ListUsersRequest request);
  $async.Future<$2.CreateUserResponse> createUser(
      $pb.ServerContext ctx, $2.CreateUserRequest request);
  $async.Future<$2.UpdateUserPasswordResponse> updateUserPassword(
      $pb.ServerContext ctx, $2.UpdateUserPasswordRequest request);
  $async.Future<$2.SetUserDisabledResponse> setUserDisabled(
      $pb.ServerContext ctx, $2.SetUserDisabledRequest request);
  $async.Future<$2.UpdateProfileResponse> updateProfile(
      $pb.ServerContext ctx, $2.UpdateProfileRequest request);
  $async.Future<$2.ChangePasswordResponse> changePassword(
      $pb.ServerContext ctx, $2.ChangePasswordRequest request);
  $async.Future<$2.ListOAuthProvidersResponse> listOAuthProviders(
      $pb.ServerContext ctx, $2.ListOAuthProvidersRequest request);
  $async.Future<$2.BeginOAuthFlowResponse> beginOAuthFlow(
      $pb.ServerContext ctx, $2.BeginOAuthFlowRequest request);
  $async.Future<$2.CompleteOAuthFlowResponse> completeOAuthFlow(
      $pb.ServerContext ctx, $2.CompleteOAuthFlowRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'Login':
        return $2.LoginRequest();
      case 'Me':
        return $2.MeRequest();
      case 'Logout':
        return $2.LogoutRequest();
      case 'ListUsers':
        return $2.ListUsersRequest();
      case 'CreateUser':
        return $2.CreateUserRequest();
      case 'UpdateUserPassword':
        return $2.UpdateUserPasswordRequest();
      case 'SetUserDisabled':
        return $2.SetUserDisabledRequest();
      case 'UpdateProfile':
        return $2.UpdateProfileRequest();
      case 'ChangePassword':
        return $2.ChangePasswordRequest();
      case 'ListOAuthProviders':
        return $2.ListOAuthProvidersRequest();
      case 'BeginOAuthFlow':
        return $2.BeginOAuthFlowRequest();
      case 'CompleteOAuthFlow':
        return $2.CompleteOAuthFlowRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'Login':
        return login(ctx, request as $2.LoginRequest);
      case 'Me':
        return me(ctx, request as $2.MeRequest);
      case 'Logout':
        return logout(ctx, request as $2.LogoutRequest);
      case 'ListUsers':
        return listUsers(ctx, request as $2.ListUsersRequest);
      case 'CreateUser':
        return createUser(ctx, request as $2.CreateUserRequest);
      case 'UpdateUserPassword':
        return updateUserPassword(ctx, request as $2.UpdateUserPasswordRequest);
      case 'SetUserDisabled':
        return setUserDisabled(ctx, request as $2.SetUserDisabledRequest);
      case 'UpdateProfile':
        return updateProfile(ctx, request as $2.UpdateProfileRequest);
      case 'ChangePassword':
        return changePassword(ctx, request as $2.ChangePasswordRequest);
      case 'ListOAuthProviders':
        return listOAuthProviders(ctx, request as $2.ListOAuthProvidersRequest);
      case 'BeginOAuthFlow':
        return beginOAuthFlow(ctx, request as $2.BeginOAuthFlowRequest);
      case 'CompleteOAuthFlow':
        return completeOAuthFlow(ctx, request as $2.CompleteOAuthFlowRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => AuthServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => AuthServiceBase$messageJson;
}
