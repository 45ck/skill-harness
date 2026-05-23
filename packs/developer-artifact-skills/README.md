# developer-artifact-skills

Embedded pack for creating durable developer artifacts and generated review surfaces.

Core policy:

- canonical source stays in Markdown, TOON, specgraph, or project docs
- HTML is a generated review surface for scanning, comparison, diagrams, prototypes, and desktop previews
- projects can opt out or switch profiles through `skill-harness setup-project`

Included skills:

- `developer-artifact-shaper` - choose the right artifact type, owner, source path, and review surface
- `html-review-artifact` - create safe, self-contained HTML review artifacts from canonical sources
- `model-review-artifact` - shape source-backed Mermaid, C4, UML-style, dependency, and architecture-space review models
- `artifact-evidence-gate` - check that an artifact is grounded in evidence and safe to hand off
- `artifact-handoff-pack` - assemble the minimal source, generated view, evidence, and next-action bundle

