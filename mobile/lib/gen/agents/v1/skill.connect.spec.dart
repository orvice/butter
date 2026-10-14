//
//  Generated code. Do not modify.
//  source: agents/v1/skill.proto
//

import "package:connectrpc/connect.dart" as connect;
import "skill.pb.dart" as agentsv1skill;

abstract final class SkillService {
  /// Fully-qualified name of the SkillService service.
  static const name = 'agents.v1.SkillService';

  static const listSkills = connect.Spec(
    '/$name/ListSkills',
    connect.StreamType.unary,
    agentsv1skill.ListSkillsRequest.new,
    agentsv1skill.ListSkillsResponse.new,
  );

  static const getSkill = connect.Spec(
    '/$name/GetSkill',
    connect.StreamType.unary,
    agentsv1skill.GetSkillRequest.new,
    agentsv1skill.GetSkillResponse.new,
  );

  static const createSkill = connect.Spec(
    '/$name/CreateSkill',
    connect.StreamType.unary,
    agentsv1skill.CreateSkillRequest.new,
    agentsv1skill.CreateSkillResponse.new,
  );

  static const updateSkill = connect.Spec(
    '/$name/UpdateSkill',
    connect.StreamType.unary,
    agentsv1skill.UpdateSkillRequest.new,
    agentsv1skill.UpdateSkillResponse.new,
  );

  static const deleteSkill = connect.Spec(
    '/$name/DeleteSkill',
    connect.StreamType.unary,
    agentsv1skill.DeleteSkillRequest.new,
    agentsv1skill.DeleteSkillResponse.new,
  );

  static const listSkillResources = connect.Spec(
    '/$name/ListSkillResources',
    connect.StreamType.unary,
    agentsv1skill.ListSkillResourcesRequest.new,
    agentsv1skill.ListSkillResourcesResponse.new,
  );

  static const getSkillResource = connect.Spec(
    '/$name/GetSkillResource',
    connect.StreamType.unary,
    agentsv1skill.GetSkillResourceRequest.new,
    agentsv1skill.GetSkillResourceResponse.new,
  );

  static const putSkillResource = connect.Spec(
    '/$name/PutSkillResource',
    connect.StreamType.unary,
    agentsv1skill.PutSkillResourceRequest.new,
    agentsv1skill.PutSkillResourceResponse.new,
  );

  static const deleteSkillResource = connect.Spec(
    '/$name/DeleteSkillResource',
    connect.StreamType.unary,
    agentsv1skill.DeleteSkillResourceRequest.new,
    agentsv1skill.DeleteSkillResourceResponse.new,
  );
}
