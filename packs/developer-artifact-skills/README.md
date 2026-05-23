# developer-artifact-skills

Embedded pack for creating durable developer artifacts and generated review surfaces.

Core policy:

- canonical source stays in Markdown, TOON, specgraph, or project docs
- HTML is a generated review surface for scanning, comparison, diagrams, prototypes, and desktop previews
- product, business, data, research, UX, and mockup work should use visual-source-first pairs when humans need rich review: agent-readable source plus generated visual surface
- high-fidelity is the default for UI, customer-facing workflow, product, and mockup review; low-fidelity sketches are scratch unless captured as evidence
- projects can opt out or switch profiles through `skill-harness setup-project`

Included skills:

- `developer-artifact-shaper` - choose the right artifact type, owner, source path, and review surface
- `visual-source-artifact` - shape product, business, data, research, UX, and mockup artifacts as source-backed visual review surfaces
- `html-review-artifact` - create safe, self-contained HTML review artifacts from canonical sources
- `model-review-artifact` - shape source-backed Mermaid, C4, UML-style, dependency, and architecture-space review models
- `artifact-evidence-gate` - check that an artifact is grounded in evidence and safe to hand off
- `artifact-handoff-pack` - assemble the minimal source, generated view, evidence, and next-action bundle

