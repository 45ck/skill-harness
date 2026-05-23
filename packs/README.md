# Embedded Packs

This directory holds suite-local packs that ship directly with `skill-harness`.

Use embedded packs when:

- a capability is new or still being curated
- the pack is tightly coupled to the harness workflow
- separate repository ownership would add more coordination cost than value

Current embedded packs:

- `coding-workflow-skills`
- `design-tooling-skills`
- `integration-tooling-skills`
- `specgraph-skills` — 5 specgraph workflow skills (requires `@45ck/agent-docs`)
- `noslop-skills` — 3 noslop quality gate skills (requires `@45ck/noslop`)
- `agent-operating-skills` — frontier-agent task shaping, context engineering, autonomy boundaries, tool permissions, memory review, multi-agent workflow review, and run evidence checks
- `developer-artifact-skills` — developer artifact shaping, source-backed model views, generated HTML review surfaces, evidence gates, and handoff bundles
- `demo-production-skills` — source-backed demo, QA, silent cut, slideshow, review surface, and release media workflows

See [docs/developer-artifacts.md](../docs/developer-artifacts.md) for the setup-project scaffold, profiles, generated HTML policy, manifest policy, model views, and maintenance rules.
See [docs/agent-operating-skills.md](../docs/agent-operating-skills.md) for the agent operating pack and governed agent-loop profile.
