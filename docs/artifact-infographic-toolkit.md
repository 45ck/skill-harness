# Artifact Infographic Toolkit

Human-facing non-model artifacts can use open-source infographic tools as source/spec renderers while keeping generated HTML static and self-contained.

## Policy

- Use Mermaid for authored architecture, process, sequence, and model diagrams that already fit Mermaid text.
- Use Vega-Lite as the default chart grammar for metrics, comparisons, evidence strength, and decision dashboards.
- Use Observable Plot for compact exploratory charts and statistical views.
- Use D3 for bespoke infographic layouts that do not fit a simpler grammar.
- Use Graphviz for node-edge dependency, lineage, and relationship maps.
- Use Apache ECharts only as a generation-time renderer or static equivalent, not as a browser runtime.
- Use RAWGraphs for design-led exported SVGs backed by tabular source data.
- Use Chart.js only through server-rendered/static output or an equivalent static chart.

Generated review HTML must not load these libraries in the browser. Render or export static SVG, static HTML, or data-url images before handoff.

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
