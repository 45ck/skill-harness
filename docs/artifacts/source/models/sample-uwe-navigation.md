---
artifactType: model-view
modelId: sample-uwe-navigation
modelKind: component
method: uwe
notation: mermaid
abstractionLevel: design
owner: system-modeler
facets:
  - content
  - navigation
  - presentation
  - process
  - access
  - adaptation
implementationTouchpoints:
  - docs/artifacts/source/product/sample-uwe-screenshot-atlas.md
  - generated/review/evidence/sample-uwe-atlas/landing.svg
  - generated/review/evidence/sample-uwe-atlas/auth.svg
  - generated/review/evidence/sample-uwe-atlas/dashboard.svg
  - generated/review/evidence/sample-uwe-atlas/channels.svg
  - generated/review/evidence/sample-uwe-atlas/settings.svg
  - generated/review/evidence/sample-uwe-atlas/denied.svg
docTouchpoints:
  - docs/artifacts/templates/e2e-product-system-atlas.md
  - docs/artifacts/source/models/model-inventory.md
evidenceLinks:
  - docs/artifacts/source/product/sample-uwe-screenshot-atlas.md
  - generated/review/evidence/sample-uwe-atlas/landing.svg
  - generated/review/evidence/sample-uwe-atlas/auth.svg
  - generated/review/evidence/sample-uwe-atlas/dashboard.svg
  - generated/review/evidence/sample-uwe-atlas/channels.svg
  - generated/review/evidence/sample-uwe-atlas/settings.svg
  - generated/review/evidence/sample-uwe-atlas/denied.svg
reviewRequired: true
updateTriggers:
  - sample atlas source changes
  - screenshot evidence changes
  - UWE atlas template changes
driftVerdict: aligned
---

# Sample UWE Navigation Model

This synthetic model shows how a screenshot-backed UWE atlas can represent navigable app states, access branches, and process actions.

## Purpose

Provide a small working example of the UWE navigation pattern that can be copied into target repos and replaced with real app routes and screenshots.

## Source Model

```mermaid
flowchart LR
  Landing["landing\n/"]
  Auth["auth\n/login"]
  Dashboard["dashboard\n/app"]
  Channels["channels\n/app/channels"]
  Settings["settings\n/app/settings"]
  Denied["access denied"]

  Landing -->|sign in|getAuth[Auth]
  getAuth --> Auth
  Auth -->|valid member| Dashboard
  Dashboard --> Channels
  Dashboard -->|admin| Settings
  Dashboard -->|member denied| Denied
  Settings --> Dashboard
```

## UWE Facet Mapping

| Node | Content | Navigation | Presentation | Process | Access | Adaptation |
| --- | --- | --- | --- | --- | --- | --- |
| landing | marketing copy | public entry | landing screenshot | get started | anonymous | desktop/mobile variants |
| auth | credential fields | login route | auth screenshot | authenticate | anonymous to member | password recovery branch |
| dashboard | community/channel data | authenticated hub | dashboard screenshot | open channels/settings | member/admin | role-specific nav |
| channels | channel list and room data | member branch | channels screenshot | select channel | member | responsive channel list |
| settings | settings data | admin branch | settings screenshot | update settings | admin only | feature flags |
| denied | denial message | restricted branch | denied screenshot | return to app | member denied | role-specific fallback |

## Screenshot Evidence

| Node | Evidence |
| --- | --- |
| landing | `generated/review/evidence/sample-uwe-atlas/landing.svg` |
| auth | `generated/review/evidence/sample-uwe-atlas/auth.svg` |
| dashboard | `generated/review/evidence/sample-uwe-atlas/dashboard.svg` |
| channels | `generated/review/evidence/sample-uwe-atlas/channels.svg` |
| settings | `generated/review/evidence/sample-uwe-atlas/settings.svg` |
| denied | `generated/review/evidence/sample-uwe-atlas/denied.svg` |

## Invariants

- Navigation nodes should map to user-reachable route or UI states.
- Each node should have screenshot evidence or an explicit gap.
- Role-sensitive links must show allowed and denied branches.
- Process actions should link to expected side effects in the product atlas.
- Synthetic evidence must be labelled as synthetic.

## Freshness

Update this model when the sample product atlas, screenshot evidence, or UWE atlas template changes.
