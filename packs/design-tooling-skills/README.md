# design-tooling-skills

Embedded pack for design-to-code execution and design system alignment.

Design work follows the suite's visual-source-first artifact policy: keep design briefs, interaction state specs, token mappings, and prototype sources as durable source, then generate high-fidelity human review surfaces when people need to inspect UI, product, or workflow behavior. Low-fidelity sketches are scratch unless explicitly captured as evidence.

Use this pack when a task needs Figma implementation planning, design token alignment, component-state review, or source-backed UI artifact planning.

Install it through the harness:

```bash
./skill-harness install --packs=design-tooling-skills --packs-only
```

Prerequisites depend on the selected skill. Figma work requires access to the relevant design source or exported evidence.

Included skills:

- `figma-implementation-planning`
- `design-token-alignment`
