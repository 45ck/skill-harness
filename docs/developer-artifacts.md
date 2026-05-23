# Developer Artifacts

`skill-harness setup-project` scaffolds developer artifact guidance by default. The capability is intentionally project-local: it shapes how a target repo records plans, decisions, evidence, and generated review surfaces without creating new global agent variants.

## Position

Developer artifacts use this source-of-truth split:

- canonical source: Markdown, TOON, JSON/YAML, specgraph / `agent-docs`, model text, structured data, or existing project docs
- generated review surface: static HTML under `generated/review/`
- handoff evidence: linked files, issue tracker entries, tests, reports, logs, screenshots, or runtime proof
- artifact provenance: `docs/artifacts/artifacts.manifest.json`

HTML is never the only durable source for a decision. Edit the source first, then regenerate or discard the review surface.

## Setup Behavior

Default setup:

```bash
./skill-harness setup-project --dir ../my-project
```

This creates:

- `.skill-harness/project.json`
- `.skill-harness/setup-proof.json`
- `docs/artifacts/artifacts.manifest.json`
- `docs/artifacts/source/`
- `docs/artifacts/source/product/`
- `docs/artifacts/source/business/`
- `docs/artifacts/source/data/`
- `docs/artifacts/source/research/`
- `docs/artifacts/source/ux/`
- `docs/artifacts/templates/`
- `generated/review/`
- `generated/review/product/`
- `generated/review/business/`
- `generated/review/data/`
- `generated/review/research/`
- `generated/review/ux/`
- `scripts/check-artifact-manifest.mjs`
- `scripts/check-artifact-html-policy.mjs`
- `scripts/open-artifact-review.mjs`

It also adds `generated/review/` to `.gitignore` and adds package scripts when applicable:

- `docs:check`
- `docs:generate`
- `docs:report`
- `artifacts:check`
- `artifacts:manifest:check`
- `artifacts:html:check`
- `artifacts:open`

If `--skip-agent-docs` is used, the artifact scaffold still works, but it does not add scripts that call `agent-docs`.

Modeling mode defaults to `auto`. Fresh developer-artifact setups resolve `auto` to UML-first modeling; existing harnessed repos preserve their current behavior unless explicitly migrated. Use explicit modes when needed:

```bash
./skill-harness setup-project --dir ../my-project --modeling-mode auto
./skill-harness setup-project --dir ../my-project --modeling-mode uml-first
./skill-harness setup-project --dir ../my-project --modeling-mode baseline
./skill-harness setup-project --dir ../my-project --skip-modeling
```

This is not a separate profile. It keeps the existing developer artifact profile and adds source-first modeling conveniences: `docs/artifacts/source/models/`, `docs/artifacts/source/models/model-inventory.md`, `generated/review/models/`, `docs/artifacts/templates/model-diff-artifact.md`, `scripts/generate-model-review.mjs`, `scripts/open-artifact-review.mjs`, `scripts/check-model-artifact-policy.mjs`, model-aware package scripts, and setup-proof check entries. The base scaffold already supports `model-view`; UML-first adds stricter UML/UWE/C4/evidence metadata, `model-diff` lineage checks, and paired human HTML review expectations. `--enable-modeling` remains as a legacy alias for `--modeling-mode uml-first`.

`.skill-harness/setup-proof.json` is the install evidence record for the setup run. It records the resolved target and operation directories, package manager, requested and effective artifact profile, tool initialization statuses, skipped capabilities, generated paths, Beads mode, and available check commands. Use it to distinguish a repo that merely has copied files from one where the harness recorded what it set up.

## Profiles

Use `--developer-artifacts-profile` to select the target workflow:

| Profile | Effective Mode | Intended Use |
|---|---|---|
| `auto` | `dual` | Default source-first workflow with optional generated review surfaces |
| `codex-app` | `html` | Codex app workflows where file-backed previews are useful |
| `claude-desktop` | `html` | Desktop preview workflows where a generated HTML artifact helps review |
| `cli` | `markdown` | Terminal-heavy work where paths and Markdown are the primary interface |
| `tui` | `markdown` | TUI work where HTML should remain secondary |
| `media` | `dual` | Demo, QA, and generated media workflows where source-backed media and review surfaces are useful |
| `agent-loop` | `dual` | Governed self-improving agent workflows with trace receipts, eval summaries, and learning proposals |
| `markdown` | `markdown` | Alias for canonical-source-only workflows |
| `html` | `html` | Alias for generated HTML review workflows |
| `dual` | `dual` | Explicit source plus generated review workflow |
| `none` | none | Disable scaffold creation |

Opt out entirely:

```bash
./skill-harness setup-project --dir ../my-project --skip-developer-artifacts
```

`--skip-artifacts` is kept as a shorter alias.

## HTML Policy

Generated HTML review artifacts must be static and self-contained by default:

- clear human review structure with overview, visuals, source, evidence, and diff/comparison sections when relevant
- tabs or equivalent navigation for dense artifacts, implemented without external runtimes
- screenshots or evidence images embedded as data URLs when they are listed in the manifest and safe to include
- no external scripts
- no external assets
- no network calls
- no inline JavaScript unless explicitly reviewed and allowed by the project
- required CSP meta tag
- semantic headings, landmarks, meaningful link text, and alt text
- no secrets, tokens, credentials, private logs, customer data, or large opaque blobs

Open generated HTML with the best available human surface:

- Codex app: use the Browser plugin for local HTML when available.
- Claude desktop: use the built-in browser or preview when available.
- CLI-only: use `node scripts/open-artifact-review.mjs` to open the system default browser.
- Headless or CI: use `node scripts/open-artifact-review.mjs --print` and pass the printed file URL to the human or host UI.
- Host integrations: use `node scripts/open-artifact-review.mjs --json --print` when an agent or app needs a machine-readable target path, file URL, and preferred host action.
- In-app browsers that block `file://` URLs should serve the generated review directory with a local static server and open the resulting `http://127.0.0.1:<port>/...` URL in the browser surface. The generated HTML remains static and source-backed; the server is only a preview transport.

Run:

```bash
node scripts/check-artifact-html-policy.mjs
node scripts/open-artifact-review.mjs --print
node scripts/open-artifact-review.mjs --json --print
```

The checker rejects common unsafe constructs including `<script>`, iframes, object/embed/form tags, meta refresh, external `src` / `href` / `action` references, and browser APIs such as `fetch`, `XMLHttpRequest`, `WebSocket`, `EventSource`, `sendBeacon`, `serviceWorker`, `document.cookie`, `localStorage`, and `sessionStorage`.

The checker is not a secret scanner, data loss prevention tool, prompt-injection detector, or provenance reviewer. Review and redact generated artifacts before opening them in a browser, sharing them, committing them, or serving them through a local preview. Add a dedicated secret scan if a project needs enforceable leakage detection.

## Visual Source-First Artifacts

Use visual-source-first artifacts when product, business, data, research, UX, or mockup work needs human visual inspection and agent-readable continuity.

| Family | Canonical Sources | Human Review Surfaces | Primary Agents |
|---|---|---|---|
| Product | PRD, opportunity brief, feature map, roadmap, acceptance criteria | product brief, feature map, decision dashboard | requirements-analyst, delivery-manager |
| Business | business model, pricing assumptions, stakeholder map, risk register | strategy review, assumption dashboard, stakeholder map | delivery-manager, requirements-analyst |
| Data | schema, data dictionary, metric definition, lineage, quality rules | schema map, metric dashboard, data quality review | backend-engineer, test-designer |
| Research | claim-evidence matrix, literature theme map, interview synthesis, assumption register | research board, evidence map, confidence dashboard | research-writer, ux-researcher |
| UX | design brief, component state spec, interaction flow, prototype source | high-fidelity prototype, state board, journey map, accessibility review | ux-researcher, web-engineer |

Rules:

- Keep source artifacts under `docs/artifacts/source/<family>/` unless a domain-specific docs path is better.
- Put generated visual review surfaces under `generated/review/<family>/`.
- High-fidelity HTML/prototype review is the default for UI, product, customer-facing workflow, and mockup artifacts. Low-fidelity sketches are scratch unless explicitly captured as research evidence.
- Visual review surfaces should expose realistic states, data density, error paths, assumptions, evidence strength, source links, and freshness metadata.
- Label synthetic user, simulated customer, or agent-generated evidence separately from real user/customer evidence.
- Record durable generated visual artifacts in `docs/artifacts/artifacts.manifest.json` with source, reviewSurface, owner, evidenceLinks, status, and freshness.
- Use a team of agents when boundaries cross: requirements for product intent, delivery for business constraints, backend for data shape, research for evidence, UX for high-fidelity review, system-modeler for structure/workflow impact, and quality-reviewer for readiness gates.

## Model And Diagram Artifacts

Mermaid, C4, UML-style, UWE-inspired, dependency, and architecture-space views fit the developer artifact model when they stay source-backed:

- auto-detect model impact for every engineering change; if code, API, workflow, dependency, deployment, UI structure, or agent behavior changes, update the relevant model source or record why no model change is needed
- keep canonical model source in Markdown, TOON, specgraph, Mermaid text, PlantUML text, or existing project docs
- generate static human HTML review surfaces from that source under `generated/review/models/`, with tabs for overview, visuals, source, evidence, and before/after diff context
- pre-render diagrams as inline SVG or static markup; do not load a browser Mermaid runtime by default
- record model kind, notation, abstraction level, source path, generated review path, owner, evidence links, renderer, and source hash in the manifest
- keep `docs/artifacts/source/models/model-inventory.md` current as the canonical index of model ids, owners, sources, review paths, and implementation touchpoints
- treat Mermaid C4 as a useful review notation, but mark the C4 level explicitly: context, container, component, dynamic, or deployment
- treat dependency graphs as generated evidence unless the project has a separate model source of truth
- keep UWE UML semantics in structured source; generate simplified review diagrams only when they help humans inspect the workflow

Run the manifest check before handing off model-backed artifacts:

```bash
node scripts/check-artifact-manifest.mjs
```

When modeling mode is `baseline` or `uml-first`, model-backed artifacts get an additional policy layer:

- `model-view` and `model-diff` entries require `modelId`, `modelKind`, `notation`, `method`, `abstractionLevel`, `owner`, source path, review path, and freshness metadata when present
- valid methods are `uml`, `uwe`, `c4`, and `evidence`; each method has allowed model kinds
- UWE facets are `content`, `navigation`, `presentation`, `process`, `access`, and `adaptation`; `access` is the local access-control facet and `adaptation` covers personalization/context variation
- `model-diff` entries must reference valid before/after artifact ids and declare `diff.method`
- HTML, SVG, PNG, screenshots, and generated comparison pages are review surfaces only; the source diff remains canonical
- generated model review HTML should live under `generated/review/models/`

Run:

```bash
node scripts/check-model-artifact-policy.mjs
node scripts/generate-model-review.mjs
node scripts/generate-model-review.mjs --check
node scripts/check-artifact-html-policy.mjs
```

Use `models:review` as the write path and `models:drift` or `artifacts:model:drift` as the read-only generated HTML drift gate.

For local human review, use `npm run models:open` when a system browser is appropriate. In Codex or another app-hosted workflow, resolve the artifact with `node scripts/open-artifact-review.mjs --json --print`, then open the URL through the host browser. If the host blocks file URLs, serve `generated/review/` locally and open `/models/index.html`.

## Media And Demo Artifacts

Use `--developer-artifacts-profile media` for repos that produce reviewable demos, QA evidence reels, silent cuts, slideshow-style MP4s, poster frames, or release demo bundles.

The media profile keeps the normal source-first artifact model and adds:

- `generated/media/`
- `docs/artifacts/templates/demo-artifact.md`
- `.skill-harness/project.json` media output policy
- `.gitignore` coverage for generated media

Media outputs are generated artifacts, not canonical truth. Keep `.demo.yaml`, QA flows, QA reports, Markdown/TOON source notes, and manifest entries as the durable source. Failed or inconclusive QA evidence may become a repro or draft plan, but not an approved product demo. Exclude raw traces, HAR/network dumps, console logs, page errors, secrets, and customer data from demo handoff bundles unless they have been explicitly redacted and approved.

## Agent Loop Artifacts

Use `--developer-artifacts-profile agent-loop` for repos that want a governed self-improving agent workflow.

The agent-loop profile keeps the normal source-first artifact model and adds:

- `generated/agent-runs/`
- `docs/artifacts/source/agent-loop-playbook.md`
- `docs/artifacts/templates/agent-loop-artifact.md`
- `scripts/check-agent-loop-policy.mjs`
- `agent-loop:check` and `agent-loop:review` package scripts
- `.skill-harness/project.json` agent-loop policy
- `.gitignore` coverage for generated run receipts

The loop uses two agents by default: a research/model agent to gather evidence and frame the gap, and a workflow/loop agent to implement a reversible slice, run gates, and propose durable learning. The human DRI remains responsible for scope, permission expansion, production data access, destructive actions, publishing, deployment, and final adoption.

Generated traces and eval summaries stay out of git by default. Promote only redacted, source-backed summaries into durable docs, issue tracker entries, or the project memory mechanism.

## Skill Pack

The embedded `developer-artifact-skills` pack provides:

- `developer-artifact-shaper`: choose artifact type, canonical source, and review surface
- `visual-source-artifact`: shape product, business, data, research, UX, and mockup artifacts as source-backed visual review surfaces
- `html-review-artifact`: create safe generated HTML review artifacts
- `model-review-artifact`: shape source-backed Mermaid, C4, UML-style, dependency, and architecture-space model views
- `artifact-evidence-gate`: check source links, evidence, freshness, and safety
- `artifact-handoff-pack`: assemble the minimal handoff bundle

These skills are wired into the author, reviewer, delivery, research, and workflow loadouts where artifact decisions naturally happen.

The embedded `demo-production-skills` pack provides:

- `demo-story-packager`: package completed demo runs with source, media, evidence, and handoff notes
- `demo-social-cut`: plan short silent cuts from evidence-backed demo runs
- `demo-slideshow-edit`: plan no-caption slideshow reels from frames, storyboards, and selected spans
- `demo-review-surface`: create static review surfaces for demo media and evidence
- `qa-to-demo`: convert QA evidence into demo specs or clip plans without changing verdict semantics
- `demo-release-packager`: assemble approved demo media into safe handoff bundles

The embedded `agent-operating-skills` pack provides:

- `self-improving-agent-loop`: design governed sense/model/plan/act/gate/learn loops
- `agent-task-shaping`: convert vague work into a bounded agent task
- `context-engineering-planner`: plan source context, evidence, memory, retrieval, and exclusions
- `autonomy-boundary-checker`: decide where the agent can act, must ask, or must stop
- `tool-permission-planner`: design least-privilege tool access and approval gates
- `agent-memory-design-reviewer`: review durable memory, retrieval, staleness, privacy, and poisoning risk
- `multi-agent-workflow-reviewer`: assign ownership, handoffs, conflict rules, and gates
- `agent-run-evidence-reviewer`: review traces, logs, eval summaries, and learning proposals

## Evidence Rules

Artifact readiness should be conservative:

- `ready`: source, evidence, and safety checks are sufficient
- `needs-source`: generated view exists but source is missing or stale
- `needs-evidence`: claims are under-supported
- `inconclusive`: evidence is missing or ambiguous
- `unsafe`: the artifact leaks sensitive data or violates HTML policy

Screenshots and prose summaries are useful review evidence, but they are not launch proof by themselves. Prefer automated or runtime evidence when the artifact is used to support delivery decisions.

## Maintenance

When changing this capability, update all of these together:

- `cmd/skill-harness/main.go`
- `cmd/skill-harness/main_test.go`
- `scripts/dependencies.json`
- `scripts/agent_loadouts.json`
- `docs/agent-loadouts.md`
- `packs/developer-artifact-skills/`
- `packs/demo-production-skills/`
- `packs/agent-operating-skills/`
- `docs/agent-operating-skills.md`
- `README.md`
- `AGENT_INSTRUCTIONS.md`

Run:

```bash
go test ./cmd/skill-harness
node -e "JSON.parse(require('fs').readFileSync('scripts/dependencies.json','utf8')); JSON.parse(require('fs').readFileSync('scripts/agent_loadouts.json','utf8'))"
```
