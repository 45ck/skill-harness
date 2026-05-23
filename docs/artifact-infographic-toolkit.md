# Artifact Infographic Toolkit

Human-facing non-model artifacts can use open-source infographic tools as source/spec renderers while keeping generated HTML static and self-contained.

## Policy

- Use Mermaid for authored architecture, process, sequence, and model diagrams that already fit Mermaid text.
- Use Graphviz or PlantUML-style source as the baseline for UWE navigation atlases; Skill Harness may add screenshot compartments after generation, but it must not replace the renderer with a custom diagram engine.
- Use Vega-Lite as the default chart grammar for metrics, comparisons, evidence strength, and decision dashboards.
- Use Observable Plot for compact exploratory charts and statistical views.
- Use D3 for bespoke infographic layouts that do not fit a simpler grammar.
- Use Graphviz for node-edge dependency, lineage, and relationship maps.
- Use Apache ECharts only as a generation-time renderer or static equivalent, not as a browser runtime.
- Use RAWGraphs for design-led exported SVGs backed by tabular source data.
- Use Chart.js only through server-rendered/static output or an equivalent static chart.

Generated review HTML must not load these libraries in the browser by default. Render or export static SVG, static HTML, or data-url images before handoff. The only current exception is a manifest-explicit `htmlInteractionLane: reviewed-svg-pan-zoom` review surface, which may embed the vendored `svg-pan-zoom` runtime plus the approved initializer so large inline SVG models can be panned and zoomed without network access.

## UWE Screenshot Atlases

UWE screenshot atlases are for whole-app inspection: routes, links, buttons, role branches, side effects, and screenshots should be visible in one navigable model surface. They must stay OSS-first:

1. Author the canonical atlas as source data in an `artifact-infographic` JSON fence.
2. Render structure with an established renderer, currently Graphviz through `@viz-js/viz`; PlantUML-compatible source can be added as another generation backend later.
3. Inject screenshots into generated UML node compartments as an extension step.
4. Use `svg-pan-zoom` for pan/zoom when the artifact opts into `reviewed-svg-pan-zoom`.
5. Keep Skill Harness glue thin: source parsing, screenshot data URLs, renderer invocation, policy checks, and manifest wiring.

## Source Spec

The generic artifact renderer recognizes JSON fenced blocks named `artifact-infographic` in canonical source files:

````markdown
```artifact-infographic
{
  "title": "Evidence Coverage",
  "tool": "vega-lite",
  "kind": "bar",
  "values": [
    {"label": "Source", "value": 3},
    {"label": "Tests", "value": 2},
    {"label": "Docs", "value": 4}
  ]
}
```
````

Graph specs can use structured nodes and edges:

````markdown
```artifact-infographic
{
  "title": "Review Flow",
  "tool": "graphviz",
  "kind": "graph",
  "edges": [
    ["Source", "Manifest"],
    ["Manifest", "Generated HTML"],
    ["Generated HTML", "Policy Check"]
  ]
}
```
````

The checked-in generic renderer includes a static fallback for bar, line, and relationship graph specs. Projects can add generation-time adapters later, but the output contract stays the same: no external scripts, no network calls, and no browser runtime dependency.

## Handoff

For human-facing discovery, planning, research, product, business, data, UX, or mockup artifacts:

1. Create canonical source under `docs/artifacts/source/<family>/`.
2. Add `artifact-infographic` specs when charts or diagrams improve review.
3. Add a manifest entry with `reviewRequired: true`.
4. Run `npm run artifacts:review`.
5. Open `generated/review/index.html` or the artifact-specific review surface.
