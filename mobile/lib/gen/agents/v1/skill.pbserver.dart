// This is a generated file - do not edit.
//
// Generated from agents/v1/skill.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'skill.pb.dart' as $1;
import 'skill.pbjson.dart';

export 'skill.pb.dart';

abstract class SkillServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListSkillsResponse> listSkills(
      $pb.ServerContext ctx, $1.ListSkillsRequest request);
  $async.Future<$1.GetSkillResponse> getSkill(
      $pb.ServerContext ctx, $1.GetSkillRequest request);
  $async.Future<$1.CreateSkillResponse> createSkill(
      $pb.ServerContext ctx, $1.CreateSkillRequest request);
  $async.Future<$1.UpdateSkillResponse> updateSkill(
      $pb.ServerContext ctx, $1.UpdateSkillRequest request);
  $async.Future<$1.DeleteSkillResponse> deleteSkill(
      $pb.ServerContext ctx, $1.DeleteSkillRequest request);
  $async.Future<$1.ListSkillResourcesResponse> listSkillResources(
      $pb.ServerContext ctx, $1.ListSkillResourcesRequest request);
  $async.Future<$1.GetSkillResourceResponse> getSkillResource(
      $pb.ServerContext ctx, $1.GetSkillResourceRequest request);
  $async.Future<$1.PutSkillResourceResponse> putSkillResource(
      $pb.ServerContext ctx, $1.PutSkillResourceRequest request);
  $async.Future<$1.DeleteSkillResourceResponse> deleteSkillResource(
      $pb.ServerContext ctx, $1.DeleteSkillResourceRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListSkills':
        return $1.ListSkillsRequest();
      case 'GetSkill':
        return $1.GetSkillRequest();
      case 'CreateSkill':
        return $1.CreateSkillRequest();
      case 'UpdateSkill':
        return $1.UpdateSkillRequest();
      case 'DeleteSkill':
        return $1.DeleteSkillRequest();
      case 'ListSkillResources':
        return $1.ListSkillResourcesRequest();
      case 'GetSkillResource':
        return $1.GetSkillResourceRequest();
      case 'PutSkillResource':
        return $1.PutSkillResourceRequest();
      case 'DeleteSkillResource':
        return $1.DeleteSkillResourceRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListSkills':
        return listSkills(ctx, request as $1.ListSkillsRequest);
      case 'GetSkill':
        return getSkill(ctx, request as $1.GetSkillRequest);
      case 'CreateSkill':
        return createSkill(ctx, request as $1.CreateSkillRequest);
      case 'UpdateSkill':
        return updateSkill(ctx, request as $1.UpdateSkillRequest);
      case 'DeleteSkill':
        return deleteSkill(ctx, request as $1.DeleteSkillRequest);
      case 'ListSkillResources':
        return listSkillResources(ctx, request as $1.ListSkillResourcesRequest);
      case 'GetSkillResource':
        return getSkillResource(ctx, request as $1.GetSkillResourceRequest);
      case 'PutSkillResource':
        return putSkillResource(ctx, request as $1.PutSkillResourceRequest);
      case 'DeleteSkillResource':
        return deleteSkillResource(
            ctx, request as $1.DeleteSkillResourceRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => SkillServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => SkillServiceBase$messageJson;
}
