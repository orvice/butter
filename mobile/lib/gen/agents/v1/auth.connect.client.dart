//
//  Generated code. Do not modify.
//  source: agents/v1/auth.proto
//

import "package:connectrpc/connect.dart" as connect;
import "auth.pb.dart" as agentsv1auth;
import "auth.connect.spec.dart" as specs;

/// AuthService manages dashboard user authentication.
extension type AuthServiceClient (connect.Transport _transport) {
  /// Login is public and exchanges username/password for a session token.
  Future<agentsv1auth.LoginResponse> login(
    agentsv1auth.LoginRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.login,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// Me returns the authenticated user for the current session.
  Future<agentsv1auth.MeResponse> me(
    agentsv1auth.MeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.me,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// Logout revokes the current session token.
  Future<agentsv1auth.LogoutResponse> logout(
    agentsv1auth.LogoutRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.logout,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListUsers lists dashboard users. Admin only.
  Future<agentsv1auth.ListUsersResponse> listUsers(
    agentsv1auth.ListUsersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.listUsers,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CreateUser creates a dashboard user. Admin only.
  Future<agentsv1auth.CreateUserResponse> createUser(
    agentsv1auth.CreateUserRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.createUser,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateUserPassword changes a user's password. Admin only.
  Future<agentsv1auth.UpdateUserPasswordResponse> updateUserPassword(
    agentsv1auth.UpdateUserPasswordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.updateUserPassword,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SetUserDisabled enables or disables a user. Admin only.
  Future<agentsv1auth.SetUserDisabledResponse> setUserDisabled(
    agentsv1auth.SetUserDisabledRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.setUserDisabled,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// UpdateProfile updates the authenticated user's display name.
  Future<agentsv1auth.UpdateProfileResponse> updateProfile(
    agentsv1auth.UpdateProfileRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.updateProfile,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ChangePassword changes the authenticated user's password.
  Future<agentsv1auth.ChangePasswordResponse> changePassword(
    agentsv1auth.ChangePasswordRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.changePassword,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListOAuthProviders is public and returns the OAuth providers configured
  /// on the server. Frontends use it to decide which "Sign in with X" buttons
  /// to render on the login page.
  Future<agentsv1auth.ListOAuthProvidersResponse> listOAuthProviders(
    agentsv1auth.ListOAuthProvidersRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.listOAuthProviders,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// BeginOAuthFlow is public and returns the provider authorize URL plus an
  /// opaque state token. The client redirects the user to authorize_url.
  Future<agentsv1auth.BeginOAuthFlowResponse> beginOAuthFlow(
    agentsv1auth.BeginOAuthFlowRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.beginOAuthFlow,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CompleteOAuthFlow is public and exchanges the provider's authorization
  /// code for a session token. The response mirrors LoginResponse.
  Future<agentsv1auth.CompleteOAuthFlowResponse> completeOAuthFlow(
    agentsv1auth.CompleteOAuthFlowRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.completeOAuthFlow,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RedeemOAuthHandoff is public and completes a native (mobile) OAuth login.
  /// The browser leg ends with a one-time handoff code delivered to the App's
  /// return URI; the App proves it began the flow with the PKCE code_verifier
  /// and receives the session token directly, never through a URL.
  Future<agentsv1auth.RedeemOAuthHandoffResponse> redeemOAuthHandoff(
    agentsv1auth.RedeemOAuthHandoffRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AuthService.redeemOAuthHandoff,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
