# Agent Operating Skills

`agent-operating-skills` is the embedded pack for frontier-agent workflow design. It is intentionally general: the pack captures reusable operating patterns for agent work, while product-specific control planes, domain agents, finance workflows, genomics workflows, Discord operations, and media pipelines stay in optional repos or separate packs.

## Design Position

Frontier-agent doctrine stays in `frontier-agent-playbook`. This embedded pack turns that doctrine into operational workflow skills that can be installed with the shared harness:

- shape ambiguous requests into bounded tasks
- plan context, memory, artifacts, and tool surfaces
- decide autonomy and approval boundaries
- review tool permissions and memory behavior
- coordinate multi-agent work only when there is a real ownership boundary
- review agent run evidence before accepting self-improvement claims
- review public skill, agent, rule, plugin, MCP, record/replay, and task-memory ecosystems before they affect the shared suite

The pack should bias toward frontier-model capability without removing governance. Agents can plan, synthesize, inspect, test, and propose improvements directly when the digital surfaces are available. Deterministic scripts, manifests, policy checks, and tests remain the scaffold for repeatability.

## Included Skills

- `self-improving-agent-loop`: design governed sense/model/plan/act/gate/learn loops
- `agent-task-shaping`: convert vague work into a bounded agent task
- `context-engineering-planner`: plan source context, evidence, memory, retrieval, and exclusions
- `autonomy-boundary-checker`: decide where the agent can act, must ask, or must stop
- `tool-permission-planner`: design least-privilege tool access and approval gates
- `agent-memory-design-reviewer`: review durable memory, retrieval, staleness, privacy, and poisoning risk
- `multi-agent-workflow-reviewer`: assign ownership, handoffs, conflict rules, and gates
- `agent-run-evidence-reviewer`: review traces, logs, eval summaries, and learning proposals
- `third-party-skill-intake`: classify public skill, agent, rule, plugin, MCP, and workflow repos before adoption
- `skill-provenance-reviewer`: review license, source, helper scripts, tool permissions, and supply-chain risk
- `external-skill-fixture-builder`: turn external ecosystem patterns into safe fixture coverage
- `host-instruction-drift-checker`: compare AGENTS, Claude, Codex, Cursor, Copilot, and skill surfaces for behavioral drift
- `record-replay-skill-reviewer`: review recorded/generated skills for redaction, provenance, and replay safety
- `task-memory-profile-planner`: choose Beads, lightweight task files, external trackers, or no durable memory yet

## Ecosystem Intake Position

External agentic repos should enter the harness through reviewed patterns and fixtures, not default dependencies.

- `research-writer` owns discovery and synthesis.
- `security-reviewer` owns provenance, permission, and generated-skill risk.
- `quality-reviewer` owns fixture coverage and evidence gates.
- `software-architect` owns installer/profile impact.
- `delivery-manager` owns task-memory profile decisions.
- `workflow-engineer` owns Beads, update flow, host drift, and CI/check integration.

The first-party adoption path is:

1. classify the external repo with `third-party-skill-intake`
2. review license and permission risk with `skill-provenance-reviewer`
3. build or update a synthetic fixture with `external-skill-fixture-builder`
4. check host-surface behavior with `host-instruction-drift-checker`
5. promote only rewritten, reviewed skills into `packs/`

## Setup Profile

Use `--developer-artifacts-profile agent-loop` for projects that need a governed self-improving loop scaffold:

```bash
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile agent-loop
```

This resolves to the normal `dual` artifact mode and adds:

- `generated/agent-runs/`
- `docs/artifacts/source/agent-loop-playbook.md`
- `docs/artifacts/templates/agent-loop-artifact.md`
- `scripts/check-agent-loop-policy.mjs`
- `agent-loop:check` and `agent-loop:review` package scripts
- agent-loop policy metadata in `.skill-harness/project.json`

Generated run receipts stay out of git by default. Promote only summarized and redacted evidence into durable docs, issue tracker entries, or project memory.

## Loadout Wiring

The pack is wired into the agents that naturally operate or review agent workflows:

- requirements analysts for task shaping and context planning
- system modelers and architects for workflow/model boundaries
- QA, quality, security, delivery, research, and workflow agents for evidence, approval, memory, and tool governance
- security, quality, research, delivery, architecture, and workflow agents for external ecosystem intake and profile impact

Do not wire this pack into every loadout by default. Use it where agent workflow behavior is part of the job, not just because the work happens to be performed by an agent.
