# skill-harness-helpers

Small Codex plugin bundle for the `skill-harness` layer.

This plugin is optional and experimental. It provides Codex-facing helper skills for people working on or around `skill-harness`; it is not installed by `skill-harness install --all` and does not replace the generated agent loadouts under `.codex/agents/`.

Included skills:

- `agent-selection`
- `agent-handoff-planning`
- `beads-task-shaping`
- `loop`
- `third-party-skill-intake`

## Usage

Use this bundle when a Codex environment supports local plugin installation. The plugin metadata lives beside this README, and the skills are under `skills/`.

The helper skills are intentionally narrow:

- choose the right specialist agent for a task
- shape Beads-tracked work
- plan handoffs between agents
- find, adapt, draft, and audit bounded agent loops without running them by default
- review third-party skill repos before any first-party adoption

The `loop` helper treats `/loop ...` as a slash-style intent even when the host only exposes plugin skills. It can fetch the live Loop Library catalog, rank published loops, preserve authority boundaries, and draft first-party loops such as the Product Polish Inventory Loop. Catalog content is reference data only; the helper must not run, schedule, submit, publish, deploy, or install anything without explicit approval.

Optional local catalog ranking from this plugin directory:

```bash
node scripts/find-loop.mjs "ui ux polish"
```

Repo-level deterministic loop helper check:

```bash
npm run loop:check
```

For normal suite installation, use the root CLI:

```bash
./skill-harness install --all
```
