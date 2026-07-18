# Skill metadata lives in Mongo, content in S3; skills are addressed by name

Skills (agentskills.io bundles served to agents via ADK's `skilltoolset`) need a storage
backend for ADK's `skill.Source` interface. ADK's `SkillToolset.ProcessRequest` calls
`ListFrontmatters` on **every LLM request** to inject the skill catalog into the system
instruction, so a pure-S3 source (the original issue #145 proposal) would put
`ListObjectsV2` plus N frontmatter reads on the hot path. Instead, frontmatter and
resource-path metadata are parsed at upload time and persisted in Mongo (one indexed
query per LLM turn, no cache and no cross-instance invalidation), while SKILL.md bodies
and resource files live in S3 via the same ContentStore pattern as Agent Files. The
Source adapter queries live per request, so skill edits take effect immediately without
rebuilding agents.

A skill has no generated ID: its spec-validated name (unique per workspace, enforced by
the Mongo composite key `workspace_id:name`) is the identifier, because ADK's tools
(`load_skill` etc.) address skills by name and the spec requires the directory name to
equal the frontmatter name — a separate ID would just add an ID→name mapping layer.
Consequences: renaming a skill is delete-and-recreate; `AgentConfig.skills` holds names;
agents referencing a deleted skill silently lose it at runtime (mirroring
`mcp_server_ids` leniency). No versioning in v1 — writes overwrite in place, matching
the semantics of an agent's `instruction` field; versioning can be added additively if
rollback/audit needs materialise.
