# skill-harness-helpers

Small Codex plugin bundle for the `skill-harness` layer.

This plugin is optional and experimental. It provides Codex-facing helper skills for people working on or around `skill-harness`; it is not installed by `skill-harness install --all` and does not replace the generated agent loadouts under `.codex/agents/`.

Included skills:

- `agent-selection`
- `agent-handoff-planning`
- `beads-task-shaping`
- `third-party-skill-intake`

## Usage

Use this bundle when a Codex environment supports local plugin installation. The plugin metadata lives beside this README, and the skills are under `skills/`.

The helper skills are intentionally narrow:

- choose the right specialist agent for a task
- shape Beads-tracked work
- plan handoffs between agents
- review third-party skill repos before any first-party adoption

For normal suite installation, use the root CLI:

```bash
./skill-harness install --all
```
