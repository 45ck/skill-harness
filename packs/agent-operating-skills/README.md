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

Current skills:

- `self-improving-agent-loop` - design governed sense/model/plan/act/gate/learn loops.
- `agent-task-shaping` - convert vague work into a bounded agent task.
- `context-engineering-planner` - plan source context, evidence, memory, retrieval, and exclusions.
- `autonomy-boundary-checker` - decide where the agent can act, must ask, or must stop.
- `tool-permission-planner` - design least-privilege tool access and approval gates.
- `agent-memory-design-reviewer` - review durable memory, retrieval, staleness, privacy, and poisoning risk.
- `multi-agent-workflow-reviewer` - assign ownership, handoffs, conflict rules, and gates.
- `agent-run-evidence-reviewer` - review traces, logs, eval summaries, and learning proposals.

Keep this pack general. Product-specific control planes, domain agents, finance workflows, genomics workflows, Discord operations, and media pipelines should live in separate optional packs unless the pattern is broadly reusable across the suite.
