# Agent Operating Skills

General skills for designing, running, reviewing, and improving frontier-agent workflows.

Use this embedded pack for reusable agent operating practices:

- designing governed self-improving agent loops
- shaping ambiguous work into bounded agent tasks
- planning context, memory, artifacts, and tool surfaces
- defining autonomy and human approval boundaries
- reviewing tool permissions and memory behavior
- coordinating multi-agent workflows
- checking run evidence before adopting workflow changes
- reviewing external skill, agent, rule, plugin, MCP, record/replay, and task-memory ecosystems before first-party adoption

Current skills:

- `self-improving-agent-loop` - design governed sense/model/plan/act/gate/learn loops.
- `agent-task-shaping` - convert vague work into a bounded agent task.
- `context-engineering-planner` - plan source context, evidence, memory, retrieval, and exclusions.
- `autonomy-boundary-checker` - decide where the agent can act, must ask, or must stop.
- `tool-permission-planner` - design least-privilege tool access and approval gates.
- `agent-memory-design-reviewer` - review durable memory, retrieval, staleness, privacy, and poisoning risk.
- `multi-agent-workflow-reviewer` - assign ownership, handoffs, conflict rules, and gates.
- `agent-run-evidence-reviewer` - review traces, logs, eval summaries, and learning proposals.
- `third-party-skill-intake` - classify public skill, agent, rule, plugin, MCP, and workflow repos before adoption.
- `skill-provenance-reviewer` - review license, source, helper scripts, tool permissions, and supply-chain risk.
- `external-skill-fixture-builder` - turn external ecosystem patterns into safe fixture coverage.
- `host-instruction-drift-checker` - compare AGENTS, Claude, Codex, Cursor, Copilot, and skill surfaces for behavioral drift.
- `record-replay-skill-reviewer` - review recorded/generated skills for redaction, provenance, and replay safety.
- `task-memory-profile-planner` - choose Beads, lightweight task files, external trackers, or no durable memory yet.

Keep this pack general. Product-specific control planes, domain agents, finance workflows, genomics workflows, Discord operations, and media pipelines should live in separate optional packs unless the pattern is broadly reusable across the suite.
