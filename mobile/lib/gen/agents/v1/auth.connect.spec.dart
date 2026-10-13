//
//  Generated code. Do not modify.
//  source: agents/v1/auth.proto
//

import "package:connectrpc/connect.dart" as connect;
import "auth.pb.dart" as agentsv1auth;

/// AuthService manages dashboard user authentication.
abstract final class AuthService {
  /// Fully-qualified name of the AuthService service.
  static const name = 'agents.v1.AuthService';

  /// Login is public and exchanges username/password for a session token.
  static const login = connect.Spec(
    '/$name/Login',
    connect.StreamType.unary,
    agentsv1auth.LoginRequest.new,
    agentsv1auth.LoginResponse.new,
  );

  /// Me returns the authenticated user for the current session.
  static const me = connect.Spec(
    '/$name/Me',
    connect.StreamType.unary,
    agentsv1auth.MeRequest.new,
    agentsv1auth.MeResponse.new,
  );

  /// Logout revokes the current session token.
  static const logout = connect.Spec(
    '/$name/Logout',
    connect.StreamType.unary,
    agentsv1auth.LogoutRequest.new,
    agentsv1auth.LogoutResponse.new,
  );

  /// ListUsers lists dashboard users. Admin only.
  static const listUsers = connect.Spec(
    '/$name/ListUsers',
    connect.StreamType.unary,
    agentsv1auth.ListUsersRequest.new,
    agentsv1auth.ListUsersResponse.new,
  );

  /// CreateUser creates a dashboard user. Admin only.
  static const createUser = connect.Spec(
    '/$name/CreateUser',
    connect.StreamType.unary,
    agentsv1auth.CreateUserRequest.new,
    agentsv1auth.CreateUserResponse.new,
  );

  /// UpdateUserPassword changes a user's password. Admin only.
  static const updateUserPassword = connect.Spec(
    '/$name/UpdateUserPassword',
    connect.StreamType.unary,
    agentsv1auth.UpdateUserPasswordRequest.new,
    agentsv1auth.UpdateUserPasswordResponse.new,
  );

  /// SetUserDisabled enables or disables a user. Admin only.
  static const setUserDisabled = connect.Spec(
    '/$name/SetUserDisabled',
    connect.StreamType.unary,
    agentsv1auth.SetUserDisabledRequest.new,
    agentsv1auth.SetUserDisabledResponse.new,
  );

  /// UpdateProfile updates the authenticated user's display name.
  static const updateProfile = connect.Spec(
    '/$name/UpdateProfile',
    connect.StreamType.unary,
    agentsv1auth.UpdateProfileRequest.new,
    agentsv1auth.UpdateProfileResponse.new,
  );

  /// ChangePassword changes the authenticated user's password.
  static const changePassword = connect.Spec(
    '/$name/ChangePassword',
    connect.StreamType.unary,
    agentsv1auth.ChangePasswordRequest.new,
    agentsv1auth.ChangePasswordResponse.new,
  );

  /// ListOAuthProviders is public and returns the OAuth providers configured
  /// on the server. Frontends use it to decide which "Sign in with X" buttons
  /// to render on the login page.
  static const listOAuthProviders = connect.Spec(
    '/$name/ListOAuthProviders',
    connect.StreamType.unary,
    agentsv1auth.ListOAuthProvidersRequest.new,
    agentsv1auth.ListOAuthProvidersResponse.new,
  );

  /// BeginOAuthFlow is public and returns the provider authorize URL plus an
  /// opaque state token. The client redirects the user to authorize_url.
  static const beginOAuthFlow = connect.Spec(
    '/$name/BeginOAuthFlow',
    connect.StreamType.unary,
    agentsv1auth.BeginOAuthFlowRequest.new,
    agentsv1auth.BeginOAuthFlowResponse.new,
  );

  /// CompleteOAuthFlow is public and exchanges the provider's authorization
  /// code for a session token. The response mirrors LoginResponse.
  static const completeOAuthFlow = connect.Spec(
    '/$name/CompleteOAuthFlow',
    connect.StreamType.unary,
    agentsv1auth.CompleteOAuthFlowRequest.new,
    agentsv1auth.CompleteOAuthFlowResponse.new,
  );

  /// RedeemOAuthHandoff is public and completes a native (mobile) OAuth login.
  /// The browser leg ends with a one-time handoff code delivered to the App's
  /// return URI; the App proves it began the flow with the PKCE code_verifier
  /// and receives the session token directly, never through a URL.
  static const redeemOAuthHandoff = connect.Spec(
    '/$name/RedeemOAuthHandoff',
    connect.StreamType.unary,
    agentsv1auth.RedeemOAuthHandoffRequest.new,
    agentsv1auth.RedeemOAuthHandoffResponse.new,
  );
}
