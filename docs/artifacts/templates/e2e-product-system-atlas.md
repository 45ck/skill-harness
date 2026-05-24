# E2E Product System Atlas: [App Name]

**Status:** Draft
**Artifact type:** e2e-product-system-atlas
**Family:** product
**Model method:** UWE navigation model with screenshot-backed evidence
**Canonical source:** docs/artifacts/source/product/[app]-e2e-product-system-atlas.md
**Generated review:** generated/review/product/[app]-e2e-product-system-atlas.html
**Owner agents:** requirements-analyst, system-modeler, ux-researcher, test-designer, web-engineer, backend-engineer, software-architect, security-reviewer, quality-reviewer

## Purpose

Create a source-first atlas for inspecting the whole app from landing page to deployed workload behavior. The review surface should show the UWE navigation structure, screenshots for navigable nodes, red evidence-only highlights/crops for important controls, manual QA evidence for actions, and runtime side effects.

## Scope

- Product boundary:
- Deployed target or environment:
- User roles:
- Included entry points:
- Excluded or unreachable areas:
- Authorization and data-safety limits:

## UWE Navigation Nodes

| Node ID | Route or state | Role(s) | Screenshot evidence | Primary actions | Expected side effects |
| --- | --- | --- | --- | --- | --- |
| landing | / | anonymous | generated/review/evidence/[app]/landing.png | sign in, sign up, browse | session unchanged |

## Navigation Links

artifact-infographic:

    {
      "title": "UWE Navigation Graph",
      "tool": "graphviz",
      "kind": "uwe-navigation",
      "summary": "Navigable app nodes grouped by UWE navigation class with screenshots embedded inside the UWE navigation nodes. Keep this bounded, not a giant whole-system UML diagram.",
      "packages": [
        "Visitor acquisition and access",
        "Authenticated app flow",
        "Utilities and admin"
      ],
      "nodes": [
        {
          "id": "Landing",
          "label": "Landing",
          "route": "/",
          "package": "Visitor acquisition and access",
          "stereotype": "navigationClass",
          "facets": ["navigation", "presentation"],
          "role": "anonymous",
          "effect": "session unchanged",
          "evidenceRefs": ["ev-landing-primary-cta"],
          "screenshot": "generated/review/evidence/[app]/landing.png"
        },
        {
          "id": "Auth",
          "label": "Auth",
          "route": "/login",
          "package": "Visitor acquisition and access",
          "stereotype": "navigationClass",
          "facets": ["access", "presentation"],
          "role": "anonymous",
          "effect": "session created on success",
          "evidenceRefs": ["ev-auth-submit"],
          "screenshot": "generated/review/evidence/[app]/auth.png"
        },
        {
          "id": "Dashboard",
          "label": "Dashboard",
          "route": "/app",
          "package": "Authenticated app flow",
          "stereotype": "navigationClass",
          "facets": ["content", "navigation", "presentation"],
          "role": "member",
          "effect": "account/project data read",
          "evidenceRefs": ["ev-dashboard-primary-action"],
          "screenshot": "generated/review/evidence/[app]/dashboard.png"
        }
      ],
      "edges": [
        {
          "id": "edge-landing-auth",
          "from": "Landing",
          "to": "Auth",
          "label": "sign in",
          "stereotype": "navigationLink",
          "evidenceRefs": ["ev-landing-primary-cta"]
        },
        {
          "id": "edge-auth-dashboard",
          "from": "Auth",
          "to": "Dashboard",
          "label": "valid session",
          "stereotype": "navigationLink",
          "guard": "valid credentials",
          "evidenceRefs": ["ev-auth-submit"]
        }
      ],
      "evidence": [
        {
          "id": "ev-landing-primary-cta",
          "kind": "screenshot",
          "path": "generated/review/evidence/[app]/landing.png",
          "state": "visitor-entry",
          "viewport": "desktop",
          "caption": "Primary CTA moves the visitor into auth.",
          "primaryFor": ["Landing"],
          "annotations": [
            {
              "id": "ann-landing-cta",
              "kind": "highlight",
              "bounds": { "x": 0.07, "y": 0.48, "w": 0.15, "h": 0.09 },
              "label": "Primary CTA",
              "relatesTo": { "edgeId": "edge-landing-auth", "actionId": "ACT-001" },
              "semantics": "evidence-only"
            }
          ]
        },
        {
          "id": "ev-auth-submit",
          "kind": "screenshot",
          "path": "generated/review/evidence/[app]/auth.png",
          "state": "auth-submit",
          "viewport": "desktop",
          "caption": "Submit control starts credential validation.",
          "primaryFor": ["Auth"],
          "annotations": [
            {
              "id": "ann-auth-submit",
              "kind": "highlight",
              "bounds": { "x": 0.30, "y": 0.64, "w": 0.18, "h": 0.09 },
              "label": "Submit credentials",
              "relatesTo": { "edgeId": "edge-auth-dashboard", "actionId": "ACT-002" },
              "semantics": "evidence-only"
            }
          ]
        }
      ]
    }

## Screenshot Evidence Contract

Use the Skill Harness UWE evidence renderer rather than a repo-local bespoke HTML diagram:

- nodes[].stereotype carries official UWE meaning.
- nodes[].facets carries descriptive concerns only.
- nodes[].evidenceRefs and edges[].evidenceRefs attach screenshot evidence to modeled elements.
- edges[].id is required when annotations reference a navigation or process link.
- evidence[].annotations[].bounds uses normalized coordinates from 0 to 1.
- evidence[].annotations[].semantics must be evidence-only.
- Red boxes and zoom crops are rendered by the review surface inspector/lightbox; do not bake them into screenshots unless no source metadata is available.
- When migrating an existing one-off atlas, map screenshot focus data to evidence[].annotations[] and package groups to packages plus nodes[].package.

## Action And Side-Effect Matrix

| Action ID | Node | Trigger | Expected UI result | Data effect | Runtime effect | Evidence | Verdict |
| --- | --- | --- | --- | --- | --- | --- | --- |
| ACT-001 | landing | click sign in | auth form visible | none | route transition only | screenshot + manual QA note | untested |

## Manual QA Sequence

1. Inventory public routes and capture desktop/mobile screenshots.
2. Authenticate with each authorized role and capture post-login navigation nodes.
3. Exercise every primary visible action once with safe test data.
4. Exercise important invalid, empty, denied, and recovery paths.
5. Record data, event, job, email, webhook, or deployed workload side effects.
6. Mark untested branches explicitly instead of implying full coverage.

## Deployment And Runtime Evidence

| Area | Evidence to capture | Notes |
| --- | --- | --- |
| Deployed URL | URL, commit, build id | avoid secrets |
| Health | health check, uptime check, smoke result | link logs only after redaction |
| Data stores | tables/collections touched | use test data |
| Jobs/events | queue/event/log observation | no raw private logs |
| Integrations | outbound calls/webhooks/emails | redact tokens and customer data |

## Screenshot Manifest

List screenshots in docs/artifacts/artifacts.manifest.json under screenshots, images, or visualEvidence. Generated HTML embeds small local images as data URLs.

json:

    {
      "screenshots": [
        {
          "path": "generated/review/evidence/[app]/landing.png",
          "caption": "Landing page",
          "alt": "Landing page screenshot"
        }
      ]
    }

## Readiness Gate

- Canonical source exists and names scope, roles, and exclusions.
- UWE navigation nodes cover all known routable or user-reachable states.
- Each primary action has expected UI result, side effect, evidence, and verdict.
- Screenshots are local, redacted, and linked from the manifest.
- Runtime claims are backed by logs, traces, health checks, tests, or deployment metadata.
- Untested branches are labelled untested or inconclusive.
- Generated HTML passes manifest, drift, and HTML policy checks.
