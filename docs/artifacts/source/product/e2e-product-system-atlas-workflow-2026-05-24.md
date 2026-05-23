---
artifactType: e2e-product-system-atlas
family: product
owner: system-modeler
reviewRequired: true
evidenceLinks:
  - docs/developer-artifacts.md
  - docs/artifacts/templates/e2e-product-system-atlas.md
  - docs/artifacts/source/models/sh-uwe-navigation-atlas.md
  - scripts/generate-artifact-review.mjs
  - cmd/skill-harness/main.go
---

# E2E Product System Atlas Workflow

An E2E Product System Atlas is a UWE navigation model with screenshots, manual QA evidence, and runtime side-effect evidence attached to each meaningful user-reachable node and action.

## Purpose

Give humans and agents a high-fidelity way to inspect a whole app without pretending that one giant UML diagram can explain the system. The atlas starts from navigation because navigation is what users can actually traverse, then hangs screenshots, actions, data effects, background work, deployment evidence, and test verdicts from those navigation nodes.

## Product Shape

The atlas is useful when an app has enough screens, roles, and deployed side effects that code search or route lists are not enough. It should answer:

- what a user can reach from the landing page
- what each role can see and do
- what screenshot evidence exists for each navigable state
- what actions change UI, session, data, jobs, integrations, or deployment state
- which paths passed manual QA and which remain untested
- what runtime or deployment evidence backs the claims

It is not a replacement for automated tests, logs, specs, or source models. It is a review surface that makes those sources inspectable.

## UWE Navigation Core

Use UWE vocabulary as the organizing model:

| UWE element | Atlas meaning | Evidence |
| --- | --- | --- |
| Content | durable product concepts, records, entities, documents, media, and user-visible data | schema notes, API responses, fixture IDs, screenshots |
| Navigation | reachable nodes, route states, role-specific links, menus, redirects, and guarded branches | route inventory, sitemap, screenshots, browser traces |
| Presentation | screenshots, responsive states, empty/loading/error states, component state evidence | local image evidence listed in the manifest |
| Process | multi-step flows, form submissions, checkout/enrollment/onboarding/admin tasks | manual QA sequence, sequence/activity diagrams, test notes |
| Access | roles, permissions, auth gates, denied states, tenant boundaries | auth tests, role matrix, security review notes |
| Adaptation | personalization, feature flags, tenant variants, device/browser variants | config evidence, screenshots per variant |

The navigation model is the spine. Other facets attach to nodes and actions.

## Canonical Sources

For each target app, create:

| Source | Purpose |
| --- | --- |
| `docs/artifacts/source/product/<app>-e2e-product-system-atlas.md` | product-facing atlas, scope, route/action matrix, QA evidence, deployment evidence |
| `docs/artifacts/source/models/<app>-uwe-navigation.md` | canonical UWE navigation model source |
| `generated/review/product/<app>-e2e-product-system-atlas.html` | screenshot-rich human review artifact |
| `generated/review/models/<app>-uwe-navigation.html` | model review surface for the UWE navigation graph |
| `generated/review/evidence/<app>/` | screenshots and small generated evidence images referenced by manifest metadata |

HTML, screenshots, SVG, PNG, comparison pages, and video remain review surfaces only. Markdown/model source and linked evidence remain canonical.

## Agent Team

| Agent | Owns |
| --- | --- |
| requirements-analyst | product boundary, roles, user goals, acceptance criteria, exclusions |
| system-modeler | UWE navigation model, use-case/activity/sequence/state slices, model inventory updates |
| ux-researcher | screenshot coverage, journey friction, presentation states, responsive checks |
| web-engineer | route/action discovery, browser capture, UI state evidence |
| test-designer or qa-automation-engineer | manual QA sequence, action matrix, oracle quality, regression candidates |
| backend-engineer | API/data side effects, persistence, jobs, events, integration effects |
| software-architect | deployment/runtime view, workload boundaries, health checks, operational evidence |
| security-reviewer | auth/access-control boundaries, redaction, private data and secret leakage risks |
| quality-reviewer | evidence gate, stale review detection, readiness verdict |

Use a team when the target app crosses those boundaries. For small apps, one agent may fill multiple roles, but the artifact should still separate ownership.

## Manual QA Contract

The atlas should cover the bounded user action space, not an infinite theoretical state space:

- every public landing or marketing entry point
- every authenticated route reachable by each role
- primary CTAs, forms, menus, destructive actions, and recovery actions
- important invalid, empty, denied, loading, offline, conflict, and success states
- side effects in data stores, sessions, queues, emails, webhooks, files, analytics, and deployed workloads
- role or tenant branches that materially change navigation or data access

Each action row should include trigger, expected UI result, data effect, runtime effect, evidence, verdict, and follow-up if untested.

## Screenshot Contract

Screenshots should be captured from authorized environments with non-sensitive test data. Store durable review images under `generated/review/evidence/<app>/` and reference them from `docs/artifacts/artifacts.manifest.json`:

```json
{
  "screenshots": [
    {
      "path": "generated/review/evidence/app/landing.png",
      "caption": "Landing page before authentication",
      "alt": "Landing page screenshot"
    }
  ]
}
```

The generic artifact generator embeds small local images as data URLs so the generated HTML review surface stays self-contained. Large captures, videos, traces, raw logs, HAR files, customer data, and unredacted secrets must not be embedded.

## Runtime Evidence

Deployment claims need concrete backing. Useful evidence includes:

- deployed URL, build ID, commit SHA, environment name
- health check or smoke test output
- relevant API response shape with test data
- sanitized log or trace excerpts
- database row or event evidence for test records
- queue, cron, webhook, email, storage, or background job observation
- monitoring dashboard screenshot when redacted and approved

If evidence cannot be safely included, record the evidence class and why the artifact does not embed it.

## Review Surface

The generated product review should include:

- status, evidence count, source depth, freshness
- UWE navigation graph from `artifact-infographic` or model source
- screenshot gallery from manifest `screenshots`, `images`, or `visualEvidence`
- action and side-effect matrix in the canonical source
- tabs for overview, evidence, source, and metadata
- links to the paired model review and any QA/runtime evidence

The generated model review should include:

- UWE navigation graph
- facets and access/process notes
- linked screenshots when available
- evidence links and update triggers

## Acceptance Criteria

| ID | Criterion |
| --- | --- |
| AC1 | Scaffolded repos include an E2E Product System Atlas template. |
| AC2 | The artifact type `e2e-product-system-atlas` is accepted by the developer artifact policy. |
| AC3 | Generic generated artifact HTML can embed local screenshot/image evidence listed in manifest metadata. |
| AC4 | The canonical model guidance frames the atlas as UWE navigation plus facets, not one giant whole-system UML diagram. |
| AC5 | Readiness requires explicit untested/inconclusive branches instead of implied full coverage. |
| AC6 | HTML review output remains static, self-contained, source-backed, and policy checked. |

```artifact-infographic
{
  "title": "Atlas Evidence Coverage",
  "tool": "vega-lite",
  "kind": "bar",
  "summary": "A ready atlas balances navigation, screenshots, QA, runtime evidence, and access review.",
  "values": [
    {"label": "Navigation", "value": 5},
    {"label": "Screenshots", "value": 5},
    {"label": "QA", "value": 4},
    {"label": "Runtime", "value": 4},
    {"label": "Access", "value": 4}
  ]
}
```

```artifact-infographic
{
  "title": "UWE Atlas Source Flow",
  "tool": "graphviz",
  "kind": "graph",
  "summary": "Canonical source and evidence generate separate product and model review surfaces.",
  "edges": [
    ["Product Source", "Product HTML"],
    ["UWE Model Source", "Model HTML"],
    ["Screenshots", "Product HTML"],
    ["Manual QA", "Product Source"],
    ["Runtime Evidence", "Product Source"],
    ["Manifest", "Freshness Checks"]
  ]
}
```

## Readiness Gate

Use `artifact-evidence-gate` before handoff:

- source exists and names product scope, roles, exclusions, and authorization limits
- UWE navigation nodes and links reflect discovered routes and role-specific reachable states
- screenshots are local, redacted, referenced from manifest metadata, and small enough to embed safely
- manual QA rows identify pass, fail, untested, or inconclusive verdicts
- runtime side effects have concrete evidence or an explicit evidence gap
- paired model source and product atlas source are both linked in the manifest
- `npm run artifacts:review` and `npm run models:review` pass

## Freshness

Update the atlas when routes, roles, navigation, UI states, forms, workflows, APIs, persistence side effects, background jobs, integrations, deployment topology, or screenshot evidence changes.
