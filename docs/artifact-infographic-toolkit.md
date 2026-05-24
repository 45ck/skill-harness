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
2. Use `kind: "uwe-navigation"` and `htmlInteractionLane: "reviewed-uwe-workspace"` on the manifest entry when the artifact needs the interactive workspace.
3. Render structure with established OSS renderers: Cytoscape.js plus dagre for the primary workspace, and Graphviz through `@viz-js/viz` for the UML fallback.
4. Inject screenshots into generated UML node compartments as an evidence extension step.
5. Render red focus boxes, crops, and callouts from structured evidence metadata; do not bake annotations into screenshots when the source can carry them.
6. Keep Skill Harness glue thin: source parsing, screenshot data URLs, renderer invocation, policy checks, and manifest wiring.

The renderer used by downstream product modeling artifacts is this Skill Harness UWE evidence renderer. Repos should not maintain separate custom UWE HTML renderers just because a one-off artifact looked better. Promote useful visual behavior into this contract and regenerate from source.

## UWE Evidence Contract

A UWE screenshot atlas uses UWE for semantics and a separate evidence layer for screenshots, red boxes, crops, and callouts:

- `packages`: UWE package/lane labels used to group nodes.
- `nodes[].id`: stable node id.
- `nodes[].stereotype`: explicit UWE stereotype such as `navigationClass`, `menu`, `index`, `query`, `processClass`, or `externalNode`.
- `nodes[].facets`: descriptive concerns such as `navigation`, `presentation`, `process`, `access`, or `adaptation`; these are not substitutes for `stereotype`.
- `nodes[].evidenceRefs`: evidence ids that provide the node screenshot, focus annotation, or crop.
- `edges[].id`: stable edge id so annotations can point at links.
- `edges[].stereotype`: `navigationLink` or `processLink`.
- `edges[].guard`: optional role, feature flag, state, or validation condition.
- `edges[].evidenceRefs`: evidence ids that prove the link trigger or result.
- `evidence[]`: screenshot or crop objects; this is the only place where red focus/crop metadata should live.
- `evidence[].annotations[].bounds`: normalized `{x,y,w,h}` coordinates relative to the displayed screenshot.
- `evidence[].annotations[].relatesTo`: `{nodeId, edgeId, actionId}` reference.
- `evidence[].annotations[].semantics`: use `evidence-only`; annotations never redefine UWE semantics.

Recommended screenshot treatment:

- Use red only for evidence highlights, not page decoration.
- Pair every red box with a number and short label.
- Show the full screenshot plus a zoomed crop in the inspector.
- Keep Graphviz/UML nodes semantically clean; detailed red callouts belong in the inspector and lightbox.
- For `processClass` and `processLink`, prefer trigger/result evidence over reusing only a generic full-screen navigation screenshot.

## Downstream Atlas Alignment

When a downstream repo already has a one-off UWE screenshot atlas, convert the durable version into a Skill Harness `artifact-infographic` source instead of preserving bespoke HTML:

- Package group maps -> `packages` plus `nodes[].package`.
- Screenshot focus maps -> `evidence[].annotations[]` with normalized bounds.
- Existing graph node source -> `nodes[]` with explicit `stereotype`.
- Existing graph edge source -> `edges[]` with explicit `id`, `stereotype`, `guard`, and `evidenceRefs`.
- Product-specific proof tags -> `facets`, namespaced notes, or separate evidence metadata; keep official UWE stereotypes intact.
- Generated output -> `generated/review/product/<app>-e2e-product-system-atlas.html` from `node scripts/generate-artifact-review.mjs`.

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

UWE screenshot specs should use the evidence contract:

````markdown
```artifact-infographic
{
  "title": "App UWE Navigation Atlas",
  "tool": "graphviz",
  "kind": "uwe-navigation",
  "packages": ["Public", "Authenticated", "Runtime"],
  "nodes": [
    {
      "id": "Settings",
      "label": "Settings",
      "route": "/app/settings",
      "package": "Authenticated",
      "stereotype": "navigationClass",
      "facets": ["access", "presentation"],
      "role": "admin",
      "actions": "save settings",
      "effect": "authorization check; settings update",
      "evidenceRefs": ["ev-settings-save"]
    }
  ],
  "edges": [
    {
      "id": "edge-settings-save",
      "from": "Settings",
      "to": "SaveSettings",
      "label": "save",
      "stereotype": "processLink",
      "guard": "admin role",
      "evidenceRefs": ["ev-settings-save"]
    }
  ],
  "evidence": [
    {
      "id": "ev-settings-save",
      "kind": "screenshot",
      "path": "generated/review/evidence/app/settings.png",
      "primaryFor": ["Settings"],
      "caption": "Save control starts the settings persistence process.",
      "annotations": [
        {
          "id": "ann-settings-save",
          "kind": "highlight",
          "bounds": {"x": 0.30, "y": 0.68, "w": 0.14, "h": 0.09},
          "label": "Save starts process",
          "relatesTo": {"edgeId": "edge-settings-save", "actionId": "ACT-005"},
          "semantics": "evidence-only"
        }
      ]
    }
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
