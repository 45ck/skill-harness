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
- `developer-artifact-skills` — developer artifact shaping, generated HTML review surfaces, evidence gates, and handoff bundles
