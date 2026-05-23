# Developer Artifacts

Requested profile: dual
Effective profile: dual

Use this directory for durable developer artifacts and generated review surfaces.

## Source Of Truth

- Keep canonical decisions, specs, investigations, product briefs, business notes, data definitions, research syntheses, UX flows, and handoff notes in Markdown, TOON, JSON/YAML, or specgraph-compatible sources.
- Treat HTML as a generated review surface for scanning, comparison, diagrams, dashboards, prototypes, mockups, and desktop app previews.
- Do not make generated HTML the only durable source for a decision.
- Record source-backed review surfaces in artifacts.manifest.json so agents and humans can detect stale output.
- For human-facing discovery, planning, product, business, data, research, UX, and mockup artifacts, set reviewRequired: true in artifacts.manifest.json and generate infographic HTML with node scripts/generate-artifact-review.mjs.

## Layout

- source/ - canonical artifact sources when they do not belong in a domain-specific docs folder.
- source/product/ - product briefs, feature maps, roadmaps, acceptance matrices, and E2E product system atlases.
- source/business/ - business models, pricing assumptions, stakeholder maps, and risk registers.
- source/data/ - schemas, data dictionaries, metric definitions, lineage, and quality rules.
- source/research/ - claim-evidence matrices, literature maps, interviews, and assumption registers.
- source/ux/ - design briefs, interaction flows, component states, and prototype sources.
- templates/ - local templates for recurring artifact types.
- artifacts.manifest.json - provenance and freshness index for source-backed review artifacts.
- ../../generated/review/ - generated HTML or rich review artifacts for humans.
- ../../generated/review/product/ - generated product review surfaces.
- ../../generated/review/business/ - generated business review surfaces.
- ../../generated/review/data/ - generated data review surfaces.
- ../../generated/review/research/ - generated research review surfaces.
- ../../generated/review/ux/ - generated UX mockups, prototypes, and state boards.
- ../../generated/media/ - generated demo media for media profile projects.
- ../../generated/agent-runs/ - generated trace receipts and eval summaries for agent-loop profile projects.

## Visual Source-First Policy

- Use visual-source-first artifacts for product, business, data, research, and UX work when humans need to inspect structure, evidence, states, tradeoffs, or mockups.
- Keep source artifacts agent-readable and diffable. Generated HTML, screenshots, videos, SVGs, PNGs, and comparison pages are review surfaces only.
- High-fidelity HTML is the default human review surface for UI, customer-facing workflow, product, and mockup reviews. Low-fidelity sketches are scratch only and should not become canonical approval surfaces.
- Visual review surfaces should show realistic data, states, error paths, assumptions, evidence strength, source links, and freshness metadata.
- Use an E2E Product System Atlas when whole-app inspection needs a UWE navigation model with screenshots, manual QA verdicts, runtime side effects, and access/adaptation branches.
- Non-UI human review surfaces should be infographic-style by default: summary metrics, charts or timelines, evidence/freshness panels, review verdicts, and links back to source.
- Label synthetic user, simulated customer, or agent-generated evidence separately from real user or customer evidence.
- Use a team of agents when ownership crosses a real boundary: requirements-analyst for product intent, delivery-manager for business constraints, backend-engineer for data shape, research-writer for evidence, ux-researcher for high-fidelity UX review, system-modeler for structural impact, and quality-reviewer for readiness gates.
- Record every durable generated visual artifact in artifacts.manifest.json with source, reviewSurface, owner, evidenceLinks, status, and freshness.

## Model And Diagram Policy

- Keep Mermaid, C4, UML-style, and architecture-space sources in Markdown, TOON, or specgraph-compatible source artifacts.
- Pre-render diagrams into generated HTML as inline SVG or static markup; do not load a browser Mermaid runtime by default.
- Treat Mermaid C4 as a review notation and record the level explicitly: context, container, component, dynamic, or deployment.
- Treat dependency graphs as generated evidence unless the project has a separate model source of truth.
- Link every generated model view back to its source artifact, issue, and evidence.
- For screenshot-backed whole-app inspection, treat UWE navigation as the model spine and attach screenshots, action QA, side effects, access rules, and adaptation variants to navigation nodes.

## Source-First Modeling

- Modeling mode: uml-first.
- Auto-detect model impact for engineering changes. If code, API, workflow, dependency, deployment, or UX structure changes, update the relevant model source or record why no model change is needed.
- Use source/models/ as the default home for canonical UML/UWE/C4/evidence model sources when no domain-specific docs folder is better.
- Use generated/review/models/ for generated human HTML before/after review surfaces.
- Keep model-view and model-diff entries in artifacts.manifest.json; the manifest carries modelId, method, facets, lineage, diff metadata, evidence links, renderer data, and source hashes.
- Treat source diffs as canonical. HTML, SVG, PNG, and screenshots are review surfaces only.
- UWE facets are content, navigation, presentation, process, access, and adaptation. Access is the local security/access-control facet; adaptation covers personalization/context variation.
- Keep model-inventory.md current as the canonical index of model ids, owners, sources, generated reviews, and implementation touchpoints.
- Run node scripts/generate-model-review.mjs to refresh static HTML review pages for humans.
- Run node scripts/check-model-artifact-policy.mjs before handing off model-backed engineering artifacts.


## Media And Demo Policy

- Keep .demo.yaml, QA flows, reports, and source notes as canonical artifacts.
- Treat MP4s, GIF/WebP previews, poster frames, and frame strips as generated outputs.
- Store generated media under generated/media/ for media profile projects and keep it out of git by default.
- Do not turn failed or inconclusive QA evidence into an approved product demo.
- Exclude raw traces, HAR/network dumps, console logs, page errors, secrets, and customer data from handoff bundles unless explicitly redacted and approved.

## Agent Loop Policy

- Start from a Beads issue or an explicit human request before changing files.
- Keep the durable loop playbook in source/agent-loop-playbook.md for agent-loop profile projects.
- Store trace receipts, eval summaries, and run evidence under generated/agent-runs/ and keep them out of git by default.
- Treat learning outputs as proposals until tests, policy checks, and the human DRI approve high-risk changes.
- Record reusable lessons with the project memory mechanism instead of unmanaged memory files.
- Do not expand tool permissions, publish, deploy, or run irreversible actions without explicit human approval.

## HTML Review Policy

- Self-contained static HTML only by default.
- No external scripts, external assets, or network calls unless the project explicitly opts in.
- Default interaction is CSS/HTML only: radio tabs, details/summary, anchor navigation, and inline SVG states.
- Do not use inline JavaScript. The reviewed inline-JS lane is reserved until manifest metadata, CSP/checker support, and human approval requirements are implemented together.
- Every HTML review artifact must include the required CSP meta tag from .skill-harness/project.json.
- Use semantic headings, landmarks, meaningful link text, and alt text for embedded images.
- No secrets, credentials, tokens, private logs, or customer data.
- Link back to the canonical source artifact and issue.
- Regenerate or discard HTML when the source changes.
- Open generated HTML with the best human review surface for the current environment: Codex Browser plugin in Codex app, Claude desktop preview/browser in Claude desktop, or node scripts/open-artifact-review.mjs for CLI/system-browser fallback.

Run this policy check before handing off generated HTML:

    node scripts/check-artifact-manifest.mjs
    node scripts/generate-artifact-review.mjs --check
    node scripts/check-artifact-html-policy.mjs
    node scripts/open-artifact-review.mjs --print
