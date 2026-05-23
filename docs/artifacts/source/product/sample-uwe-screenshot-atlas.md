---
artifactType: e2e-product-system-atlas
family: product
owner: system-modeler
reviewRequired: true
evidenceLinks:
  - docs/artifacts/source/product/e2e-product-system-atlas-workflow-2026-05-24.md
  - docs/artifacts/source/models/sample-uwe-navigation.md
  - generated/review/evidence/sample-uwe-atlas/landing.svg
  - generated/review/evidence/sample-uwe-atlas/auth.svg
  - generated/review/evidence/sample-uwe-atlas/dashboard.svg
  - generated/review/evidence/sample-uwe-atlas/channels.svg
  - generated/review/evidence/sample-uwe-atlas/settings.svg
  - generated/review/evidence/sample-uwe-atlas/denied.svg
---

# Sample UWE Screenshot Atlas

This is a synthetic Vibecord-style example showing how a UWE navigation model can become a screenshot-backed inspection artifact. It is not evidence from the real Vibecord app.

## Purpose

Demonstrate the reusable shape: navigation nodes are the spine, screenshots make each node inspectable, actions show process behavior, and side effects connect UI actions to runtime or data changes.

## Renderer Contract

This atlas is intentionally OSS-first. The source graph is rendered first as a Cytoscape.js + dagre workspace for pan, zoom, package focus, and node inspection. The same source is also converted to DOT and rendered with Graphviz through `@viz-js/viz`; Skill Harness only injects screenshot evidence into UML node compartments after the renderer has produced SVG. The HTML review surface opts into the reviewed `reviewed-uwe-workspace` lane so the bundled graph viewer works without CDN scripts or network access.

## Scope

| Field | Value |
| --- | --- |
| Product boundary | synthetic community app example |
| Environment | generated sample evidence only |
| Roles | anonymous, member, admin |
| Included entry points | landing, sign-in, dashboard |
| Exclusions | real Vibecord routes, production data, real auth secrets |
| Data safety | screenshots are synthetic SVGs |

## UWE Navigation Nodes

| Node ID | Route or state | UWE facet | Role(s) | Screenshot evidence | Primary actions | Expected side effects |
| --- | --- | --- | --- | --- | --- | --- |
| landing | `/` | navigation + presentation | anonymous | `generated/review/evidence/sample-uwe-atlas/landing.svg` | sign in, get started | route transition, no session mutation |
| auth | `/login` | access + process | anonymous | `generated/review/evidence/sample-uwe-atlas/auth.svg` | submit credentials, recover password | session token on success, validation errors on failure |
| dashboard | `/app` | content + navigation + presentation | member, admin | `generated/review/evidence/sample-uwe-atlas/dashboard.svg` | open channels, view activity, open settings | API fetch, membership read, activity query |
| channels | `/app/channels` | navigation + content | member | `generated/review/evidence/sample-uwe-atlas/channels.svg` | inspect rooms, select channel | channel list query, selected-room read |
| settings | `/app/settings` | access + presentation | admin | `generated/review/evidence/sample-uwe-atlas/settings.svg` | change settings, save | authorization check, settings update |
| denied | access-denied state | access + adaptation | member | `generated/review/evidence/sample-uwe-atlas/denied.svg` | back to app | authorization denial, no settings mutation |

## Navigation Links

```artifact-infographic
{
  "title": "Sample UWE Navigation Graph",
  "tool": "graphviz",
  "kind": "uwe-navigation",
  "summary": "A bounded UWE navigation graph with navigation classes and screenshots embedded inside the navigational nodes.",
  "navigationClasses": [
    "Visitor acquisition and access",
    "Authenticated workspace flow",
    "Workspace utilities and admin"
  ],
  "nodes": [
    {
      "id": "Landing",
      "label": "Landing",
      "route": "/",
      "navigationClass": "Visitor acquisition and access",
      "facet": "navigation",
      "role": "anonymous",
      "actions": "sign in, get started",
      "effect": "route transition only",
      "screenshot": "generated/review/evidence/sample-uwe-atlas/landing.svg"
    },
    {
      "id": "Auth",
      "label": "Auth",
      "route": "/login",
      "navigationClass": "Visitor acquisition and access",
      "facet": "access",
      "role": "anonymous",
      "actions": "submit credentials, recover password",
      "effect": "session token on success",
      "screenshot": "generated/review/evidence/sample-uwe-atlas/auth.svg"
    },
    {
      "id": "Dashboard",
      "label": "Dashboard",
      "route": "/app",
      "navigationClass": "Authenticated workspace flow",
      "facet": "content",
      "role": "member",
      "actions": "open channels, view activity, open settings",
      "effect": "membership and activity read",
      "screenshot": "generated/review/evidence/sample-uwe-atlas/dashboard.svg"
    },
    {
      "id": "Channels",
      "label": "Channels",
      "route": "/app/channels",
      "navigationClass": "Authenticated workspace flow",
      "facet": "navigation",
      "role": "member",
      "actions": "inspect rooms, select channel",
      "effect": "channel list query",
      "screenshot": "generated/review/evidence/sample-uwe-atlas/channels.svg"
    },
    {
      "id": "Settings",
      "label": "Settings",
      "route": "/app/settings",
      "navigationClass": "Workspace utilities and admin",
      "facet": "access",
      "role": "admin",
      "actions": "change settings, save",
      "effect": "authorization check",
      "screenshot": "generated/review/evidence/sample-uwe-atlas/settings.svg"
    },
    {
      "id": "Denied",
      "label": "Access denied",
      "route": "access-denied",
      "navigationClass": "Workspace utilities and admin",
      "facet": "adaptation",
      "role": "member",
      "actions": "back to app",
      "effect": "blocked settings mutation",
      "screenshot": "generated/review/evidence/sample-uwe-atlas/denied.svg"
    }
  ],
  "edges": [
    ["Landing", "Auth", "sign in"],
    ["Auth", "Dashboard", "valid session"],
    ["Dashboard", "Channels", "open channels"],
    ["Dashboard", "Settings", "admin only"],
    ["Dashboard", "Denied", "member denied"],
    ["Denied", "Dashboard", "back"],
    ["Settings", "Dashboard", "back"]
  ]
}
```

## Presentation Evidence

The manifest lists three synthetic screenshots. A real repo should replace these with Playwright captures from an authorized test environment:

| Screenshot | Node | Purpose |
| --- | --- | --- |
| `landing.svg` | landing | anonymous entry state |
| `auth.svg` | auth | access/process state |
| `dashboard.svg` | dashboard | authenticated navigation state |
| `channels.svg` | channels | channel list and selected-room state |
| `settings.svg` | settings | admin settings branch |
| `denied.svg` | denied | role-sensitive denied branch |

## Action And Side-Effect Matrix

| Action ID | Node | Trigger | Expected UI result | Data effect | Runtime effect | Evidence | Verdict |
| --- | --- | --- | --- | --- | --- | --- | --- |
| ACT-001 | landing | click sign in | auth node visible | none | route transition | landing + auth screenshots | pass |
| ACT-002 | auth | submit valid credentials | dashboard visible | session created | auth API call, session cookie/token | auth + dashboard screenshots | synthetic |
| ACT-003 | dashboard | open channels | channel list visible | membership/channel read | API fetch and cache update | dashboard + channels screenshots | synthetic |
| ACT-004 | dashboard | open settings as member | access denied visible | no settings mutation | authorization check | dashboard + denied screenshots | synthetic |
| ACT-005 | dashboard | open settings as admin | settings visible | settings read/update available | authorization check | dashboard + settings screenshots | synthetic |

## Access And Adaptation

| Branch | Expected behavior | Evidence state |
| --- | --- | --- |
| anonymous | can reach landing and auth only | captured |
| member | can reach dashboard and member channels; denied from admin settings | dashboard, channels, denied captured |
| admin | can reach dashboard, members, moderation, settings | settings branch captured |
| mobile viewport | navigation should collapse without hiding primary actions | not captured |
| feature flag disabled | flagged routes should be absent or disabled | not captured |

## Runtime Evidence

| Area | Expected evidence in a real app | Sample state |
| --- | --- | --- |
| Deployed URL | app URL, commit, build ID | omitted |
| Health | smoke check or health endpoint | omitted |
| Data stores | test user, community, channel rows | synthetic only |
| Jobs/events | activity event after login or channel action | synthetic only |
| Integrations | email/webhook/analytics evidence when relevant | omitted |

## Manual QA Sequence

1. Capture landing in desktop and mobile viewports.
2. Navigate to auth and capture empty, invalid, and valid states.
3. Sign in with a test member and capture the dashboard.
4. Exercise all visible dashboard actions once with safe test data.
5. Repeat role-sensitive paths for admin and denied member branches.
6. Record side effects for API calls, session state, data rows, jobs, and integrations.
7. Mark any missing branch `untested` or `inconclusive`.

```artifact-infographic
{
  "title": "Sample Coverage",
  "tool": "vega-lite",
  "kind": "bar",
  "summary": "Synthetic coverage values show what a real atlas should make visible.",
  "values": [
    {"label": "Nodes", "value": 3},
    {"label": "Screens", "value": 3},
    {"label": "Actions", "value": 4},
    {"label": "Roles", "value": 2},
    {"label": "Runtime", "value": 1}
  ]
}
```

## Reuse Across Repos

To use this pattern in another repo:

1. Run `skill-harness setup-project --modeling-mode uml-first`.
2. Copy `docs/artifacts/templates/e2e-product-system-atlas.md` into a new product source.
3. Capture screenshots under `generated/review/evidence/<app>/`.
4. Add screenshots to the manifest under `screenshots`, `images`, or `visualEvidence`.
5. Add or update the paired UWE model source under `docs/artifacts/source/models/`.
6. Run `npm run artifacts:review` and `npm run models:review`.

## Readiness

This sample is ready as a demonstration artifact, not as product evidence. Real app atlases must replace synthetic screenshots and sample verdicts with authorized captures, manual QA output, and runtime evidence.
