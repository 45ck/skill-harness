---
artifactType: implementation-plan
family: product
owner: system-modeler
reviewRequired: true
evidenceLinks:
  - docs/artifact-infographic-toolkit.md
  - docs/artifacts/templates/e2e-product-system-atlas.md
  - docs/artifacts/source/product/sample-uwe-screenshot-atlas.md
  - scripts/generate-artifact-review.mjs
  - scripts/uwe-workspace-runtime.js
  - C:/Dev/VibeCoord/qa-artifacts/e2e-product-modeling/vibecoord-uwe-navigation-source-2026-05-24.json
  - C:/Dev/VibeCoord/qa-artifacts/html/vibecoord-e2e-product-modeling-2026-05-24.html
---

# VibeCoord UWE Renderer Adoption

VibeCoord and Skill Harness should use the same UWE screenshot evidence renderer. The VibeCoord artifact proved the right visual language: package-organized UWE graph, screenshot-backed nodes, red focus boxes, zoom crops, and an inspector for action/effect evidence. Skill Harness now owns that reusable renderer contract so future VibeCoord artifacts should be source-first inputs to `scripts/generate-artifact-review.mjs`, not hand-authored one-off HTML.

## Decision

Use Skill Harness `artifact-infographic` specs with `kind: "uwe-navigation"` and manifest `htmlInteractionLane: "reviewed-uwe-workspace"` as the canonical rendering path for VibeCoord-style product modeling.

The VibeCoord HTML can remain historical evidence, but the next durable atlas should be generated from a canonical source file with:

- UWE packages in `packages`
- explicit `nodes[].stereotype`
- descriptive `nodes[].facets`
- stable `edges[].id`
- `edges[].stereotype` as `navigationLink` or `processLink`
- top-level `evidence[]`
- normalized red-box annotations in `evidence[].annotations[].bounds`
- `semantics: "evidence-only"` on every annotation

## Crosswalk

| VibeCoord current source | Skill Harness renderer source | Reason |
| --- | --- | --- |
| `uwePackageGroups` / package maps | `packages` and `nodes[].package` | Keeps the model organized as UWE packages instead of a generic network. |
| screenshot-backed graph node list | `nodes[]` | Preserves reachable UI states as UWE nodes. |
| node type labels | `nodes[].stereotype` | Prevents screenshots or AI overlays from redefining UWE semantics. |
| visual focus rectangles | `evidence[].annotations[]` | Makes red boxes reusable, inspectable, and source-diffable. |
| focus crop behavior | `evidence[].crop` or first annotation bounds | Lets the generic inspector render zoom crops. |
| action/process descriptions | `edges[]` plus action matrix rows | Links user actions to navigation/process effects. |
| VibeCoord AI/proof overlay notes | namespaced evidence metadata or facets | Keeps official UWE profile clean while retaining VibeCoord-specific proof context. |
| custom artifact HTML | generated review surface | Avoids maintaining a second renderer. |

## Migration Steps

1. Create `docs/artifacts/source/product/vibecoord-e2e-product-system-atlas-2026-05-24.md` in VibeCoord.
2. Convert `qa-artifacts/e2e-product-modeling/vibecoord-uwe-navigation-source-2026-05-24.json` into an `artifact-infographic` fence using the Skill Harness UWE evidence contract.
3. Move each screenshot focus rectangle into `evidence[].annotations[]`.
4. Give every UWE edge a stable `id`.
5. Add or update the VibeCoord artifact manifest entry with `reviewRequired: true`, `reviewSurface`, and `htmlInteractionLane: "reviewed-uwe-workspace"`.
6. Generate the review with the Skill Harness renderer.
7. Keep the old VibeCoord HTML artifact only as historical QA evidence until the generated atlas is visually accepted.

## Acceptance Criteria

| ID | Criterion |
| --- | --- |
| AC1 | VibeCoord has a canonical source atlas that uses the Skill Harness UWE evidence contract. |
| AC2 | The generated VibeCoord review surface uses the same Cytoscape/dagre workspace, Graphviz fallback, red focus boxes, crop panel, and annotated lightbox as the Skill Harness sample. |
| AC3 | No VibeCoord-specific custom UWE renderer remains the default path for new atlases. |
| AC4 | Red focus boxes are `evidence-only` and never change UWE stereotypes. |
| AC5 | Process links and guarded branches have edge IDs and evidence references. |

```artifact-infographic
{
  "title": "VibeCoord Atlas Renderer Unification",
  "tool": "graphviz",
  "kind": "graph",
  "summary": "VibeCoord keeps its product-specific screenshots and proof model, but Skill Harness owns the reusable UWE evidence rendering path.",
  "edges": [
    ["VibeCoord QA screenshots", "VibeCoord atlas source"],
    ["VibeCoord focus rectangles", "UWE evidence annotations"],
    ["VibeCoord atlas source", "Skill Harness renderer"],
    ["Skill Harness renderer", "Generated VibeCoord atlas"],
    ["Generated VibeCoord atlas", "Founder review"],
    ["Renderer improvements", "Skill Harness contract"]
  ]
}
```

```artifact-infographic
{
  "title": "Adoption Readiness",
  "tool": "vega-lite",
  "kind": "bar",
  "summary": "Most of the renderer behavior already exists; the remaining work is converting VibeCoord's current bespoke source into the shared contract.",
  "values": [
    {"label": "Renderer", "value": 5},
    {"label": "UWE semantics", "value": 5},
    {"label": "Red highlights", "value": 5},
    {"label": "VibeCoord source conversion", "value": 2},
    {"label": "VibeCoord manifest wiring", "value": 1}
  ]
}
```

## Copy Back For VibeCoord

Use this note in the VibeCoord repo:

```md
Future UWE screenshot atlases must use the Skill Harness `artifact-infographic` `kind: "uwe-navigation"` contract and the `reviewed-uwe-workspace` renderer. Do not create new one-off UWE HTML renderers. Convert screenshot focus boxes into `evidence[].annotations[]`, keep official UWE stereotypes in `nodes[].stereotype`, and attach VibeCoord-specific AI/proof details as namespaced evidence metadata or facets.
```

