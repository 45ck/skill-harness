# Model-To-Code Human Artifacts Plan

Date: 2026-05-23

Beads issue: `skill-harness-ra5`

Status: planning artifact

Generated review surface: `generated/review/model-to-code-human-artifacts-plan-2026-05-23.html`

## Purpose

This artifact plans how `skill-harness`, and optionally `agent-docs` / specgraph, can keep canonical software models mapped to code while producing human-readable Markdown and browser review artifacts.

The immediate goal is not full round-trip UML. The useful goal is a governed loop:

1. Keep model sources as durable text.
2. Map model elements to code, tests, specs, issues, and generated evidence.
3. Generate Markdown and browser review surfaces from those sources.
4. Fail checks when mappings, generated docs, or review surfaces drift.

## Diagnosis

`skill-harness` already contains much of the downstream scaffold logic:

- `setup-project` supports developer artifacts, generated review surfaces, manifest policy, and modelling modes.
- `--modeling-mode` and `--enable-modeling` provide stricter UML/UWE/C4 setup for target repos.
- model-related policy and tests exist in `cmd/skill-harness/main.go` and `cmd/skill-harness/main_test.go`.
- source notes already exist under `docs/artifacts/source/`.

The repo root itself is not fully self-scaffolded as a developer artifact project:

- `docs/artifacts/artifacts.manifest.json` was missing before this artifact.
- `.skill-harness/project.json` is missing at the root.
- `.skill-harness/setup-proof.json` is missing at the root.
- `scripts/check-artifact-manifest.mjs` is missing at the root.
- `scripts/check-artifact-html-policy.mjs` is missing at the root.
- `scripts/check-model-artifact-policy.mjs` is missing at the root.
- `docs/artifacts/source/models/` and `generated/review/models/` are missing at the root.

That is why browser-facing human artifacts are not naturally produced here unless an agent creates them explicitly. The harness can scaffold this for target repos, but the harness repo has not fully applied its own artifact contract to itself.

## Core Decision

Use a hybrid source-of-truth model:

- `skill-harness` owns project setup, scaffold generation, artifact policy templates, and suite-level drift checks.
- `agent-docs` / specgraph owns requirements/spec traceability and documentation gates where a target repo already uses it.
- target repos own their canonical model sources and their code evidence.
- generated HTML is a review surface only.

Do not create a separate UML universe. Keep UML/C4/model artifacts inside the existing developer artifact system.

## Recommended Architecture

### 1. Computed Suite Graph In `skill-harness`

Add a repo-local computed graph that loads:

- `scripts/dependencies.json`
- `scripts/agent_loadouts.json`
- `.claude/agents/`
- `.codex/agents/`
- embedded pack metadata under `packs/`
- selected docs projections such as `docs/agent-loadouts.md`

The graph should normalize:

- packs
- skills
- agents
- agent loadouts
- generated agent templates
- artifact profiles
- modelling modes
- setup-project outputs

This graph becomes the internal model for harness consistency. It should be computed from existing sources first, not stored as a new giant catalog in the first slice.

### 2. Hybrid Trace Layer With `agent-docs` / specgraph

Use `skill-harness` for scaffold policy and artifact generation, and use `agent-docs` / specgraph for trace queries when a target repo has it installed.

Recommended split:

| Area | Owns | Should Not Own |
|---|---|---|
| `skill-harness` | scaffold contract, templates, package scripts, policy checks, migration modes, review generators | project-specific model content, deep code indexing, long-lived graph state |
| `agent-docs` / specgraph | code/spec/model trace graph, impact queries, evidence linking, changed-file to model resolution | HTML rendering, project scaffold layout, generated review files |
| scaffolded project repo | canonical model sources, touchpoints, owners, evidence, repo-local commands | global install logic, cross-repo policy defaults |
| `generated/review/models/` | static HTML and generated review Markdown for humans | canonical truth |

If specgraph is available, add model nodes and edges:

- `MODEL`
- `IMPLEMENTS_MODEL`
- `TOUCHES_MODEL`
- `EVIDENCES_MODEL`
- `REVIEWS_MODEL`

The local scaffold should still work without specgraph by using file/path/source-hash checks. Specgraph improves impact queries; it should not be required for basic artifact hygiene.

### 3. Model Inventory In Target Repos

For repos using model-aware artifacts, scaffold:

- `docs/artifacts/source/models/system-context.md`
- `docs/artifacts/source/models/domain-language.md`
- `docs/artifacts/source/models/use-case-index.md`
- `docs/artifacts/source/models/update-matrix.md`
- `generated/review/models/`
- `docs/artifacts/artifacts.manifest.json`

The first durable models should be:

- context
- domain
- use-case
- selected sequence/activity/state models for real workflows

Avoid one giant whole-system UML diagram.

### 4. Model-To-Code Mapping Contract

Every canonical model artifact should declare stable model IDs and explicit evidence mappings.

Suggested source front matter:

```yaml
artifactType: model-view
modelId: setup-project-flow
modelKind: activity
method: uml
notation: markdown-mermaid
abstractionLevel: design
owner: system-modeler
codeMappings:
  - kind: command
    path: cmd/skill-harness/main.go
    symbol: runSetupProject
  - kind: config
    path: scripts/agent_loadouts.json
evidenceLinks:
  - cmd/skill-harness/main_test.go
  - README.md
updateTriggers:
  - setup-project behavior changes
  - artifact profile changes
  - modelling mode changes
```

The mapping checker should verify:

- mapped files exist
- mapped symbols exist where a language parser or stable text check is available
- expected generated docs match current model source
- source hashes are current
- generated HTML review paths exist when required by profile and status
- model diffs reference valid before/after artifacts

## Artifact Pipeline

```text
Model source
  -> model parser / lightweight metadata reader
  -> suite graph or project graph
  -> Markdown docs
  -> generated HTML review surface
  -> manifest freshness record
  -> CI drift checks
```

The same source should be able to produce:

- human Markdown for repo browsing
- generated HTML for browser review
- machine-readable manifest entries
- CI failures when stale

## Where Things Belong

| Concern | Home | Reason |
|---|---|---|
| project setup flags and scaffold files | `skill-harness` | It already owns setup-project and artifact profiles. |
| suite graph over packs and agents | `skill-harness` | The harness owns pack/agent installation and rendering. |
| requirements/spec traceability | `agent-docs` / specgraph | It already owns spec-oriented documentation gates. |
| model-to-code mapping schema | shared, scaffolded by `skill-harness` | Target repos need the schema without importing harness at runtime. |
| generated HTML policy | `skill-harness` scaffold | The harness already emits artifact policy checkers. |
| repo-specific diagrams and docs | target repo | The target repo owns its domain and code truth. |
| dependency and import rules | target repo CI | Rules differ by language and architecture. |

## Skill Harness Baseline Model Taxonomy

Use stable `modelId` values in kebab case. Do not embed dates, versions, or issue IDs in `modelId`.

| Model ID | Kind / Method | Level | Canonical Source | Primary Touchpoints |
|---|---|---|---|---|
| `sh-system-context` | `context` / `c4` | runtime | `docs/artifacts/source/models/sh-system-context.md` | `README.md`, `cmd/skill-harness/main.go`, `scripts/dependencies.json`, `packs/` |
| `sh-domain-artifact-governance` | `domain` / `uml` | domain | `docs/artifacts/source/models/sh-domain-artifact-governance.md` | `docs/developer-artifacts.md`, `AGENTS.md`, `AGENT_INSTRUCTIONS.md`, `cmd/skill-harness/main.go` |
| `sh-usecase-cli-workflows` | `use-case` / `uml` | domain | `docs/artifacts/source/models/sh-usecase-cli-workflows.md` | `README.md`, `cmd/skill-harness/main.go` |
| `sh-activity-setup-project` | `activity` / `uml` | runtime | `docs/artifacts/source/models/sh-activity-setup-project.md` | `cmd/skill-harness/main.go`, `README.md` |
| `sh-component-scaffold-engine` | `component` / `c4` | design | `docs/artifacts/source/models/sh-component-scaffold-engine.md` | `cmd/skill-harness/main.go`, `scripts/*.py`, `scripts/*.json` |
| `sh-state-artifact-freshness` | `state` / `uml` | decision | `docs/artifacts/source/models/sh-state-artifact-freshness.md` | `docs/developer-artifacts.md`, manifest checker code, model policy checker code |
| `sh-sequence-setup-project-toolchain` | `sequence` / `uml` | runtime | `docs/artifacts/source/models/sh-sequence-setup-project-toolchain.md` | `setup-project`, `agent-docs`, `noslop`, `bd`, proof writing |
| `sh-deployment-local-toolchain` | `deployment` / `c4` | deployment | `docs/artifacts/source/models/sh-deployment-local-toolchain.md` | install scripts, home-dir agent output, target repo scaffold paths |
| `sh-loadout-pack-mapping` | `architecture-space` / `evidence` | runtime | `docs/artifacts/source/models/sh-loadout-pack-mapping.md` | `scripts/dependencies.json`, `scripts/agent_loadouts.json`, `docs/agent-loadouts.md` |

Required baseline models:

- `sh-system-context`
- `sh-domain-artifact-governance`
- `sh-usecase-cli-workflows`
- `sh-activity-setup-project`
- `sh-component-scaffold-engine`
- `sh-state-artifact-freshness`

The remaining models become required when the touched change crosses tool orchestration, install topology, or agent/pack mapping.

## Drift Verdicts

Keep `artifact.status` as the lifecycle field. Add a derived `driftVerdict` in review output and handoff notes.

| Drift Verdict | Meaning | Lifecycle Mapping |
|---|---|---|
| `aligned` | source, mapping, evidence, and review surface match code/docs | `ready` |
| `source-missing` | behavior changed but no canonical model source exists | `needs-source` |
| `inventory-mismatch` | inventory, manifest, and touchpoints disagree | `needs-source` |
| `review-stale` | source or code changed after HTML generation | `stale` |
| `evidence-missing` | claims are under-supported | `needs-evidence` |
| `policy-failed` | manifest, model, or HTML policy checks fail | `unsafe` |
| `inconclusive` | impact cannot yet be determined | `inconclusive` |

## Implementation Slices

### Slice 1: Self-Scaffold Harness Artifacts

Create the root artifact contract for this repo:

- `.skill-harness/project.json`
- `.skill-harness/setup-proof.json`
- `docs/artifacts/artifacts.manifest.json`
- `scripts/check-artifact-manifest.mjs`
- `scripts/check-artifact-html-policy.mjs`
- `docs/artifacts/source/models/`
- `generated/review/models/`

Acceptance:

- this planning artifact is listed in the manifest
- generated review HTML links back to canonical Markdown
- local artifact checks can run without network access

### Slice 2: Computed Suite Graph

Add:

- `scripts/suite_graph.py`
- `scripts/render_suite_docs.py`
- `scripts/check_suite_drift.py`

Acceptance:

- the graph enumerates configured packs, embedded skills, agents, loadouts, and templates
- `docs/agent-loadouts.md` can be regenerated deterministically
- manual drift in generated docs causes `check_suite_drift.py --check` to fail

### Slice 3: Model Inventory Scaffold

Extend model-aware setup to create:

- system context model
- domain language model
- use-case index
- update matrix
- examples of codeMappings and evidenceLinks

Acceptance:

- `setup-project --modeling-mode uml-first` creates non-empty starter model sources
- every starter model has stable IDs and update triggers

### Slice 4: Mapping And Freshness Checks

Add model mapping validation:

- path existence checks
- source hash checks
- review surface path checks
- optional symbol text checks
- model status verdicts

Acceptance:

- removing a mapped source path fails the check
- stale source hash fails the check
- ready model artifacts without evidence fail the check
- generated evidence model kinds are not treated as authored canonical truth unless explicitly marked

### Slice 5: Browser Human Artifact Workflow

Add a local command or documented script:

```bash
npm run artifacts:review
```

or:

```bash
skill-harness artifacts review --open
```

Acceptance:

- generated HTML is static and self-contained
- browser review surfaces are created under `generated/review/`
- desktop agents can open the generated page
- CLI agents still have Markdown source as the durable handoff

### Slice 6: CI Gates

Add CI checks:

- `go test ./...`
- Go build
- Python script compile
- suite graph drift check
- artifact manifest check
- HTML safety check
- temp repo `setup-project` smoke
- temp home render smoke

Acceptance:

- CI does not depend on user-home state
- generated docs drift fails CI
- artifact safety failures fail CI

### Slice 7: Specgraph Model Impact Indexing

Add optional `agent-docs` / specgraph support for model impact queries.

Acceptance:

- model metadata can be ingested as graph nodes
- changed files can resolve to impacted models
- `models:impact` can print `none`, `evidence-only`, or `canonical-model-update-required`
- repos without specgraph still pass basic path/hash/manifest checks

## Change Update Rules

| Change Type | Required Model Action |
|---|---|
| user goal or workflow changes | update use-case and activity; maybe sequence |
| domain concepts or business rules change | update domain; maybe state |
| service or component boundary changes | update context/container/component |
| deployment or trust boundary changes | update deployment and security evidence |
| internal refactor only | regenerate dependency evidence; durable UML usually unchanged |
| generated review is stale | regenerate from source or mark artifact not ready |

## Browser Artifact Policy

Generated browser artifacts should:

- be self-contained HTML
- use inline CSS
- avoid external scripts, external assets, and network calls
- link to the canonical source artifact
- include issue ID, owner, status, evidence links, and freshness state
- use semantic HTML and accessible contrast
- stay out of canonical truth

The current artifact is intentionally dual:

- canonical source: this Markdown file
- generated human review: `generated/review/model-to-code-human-artifacts-plan-2026-05-23.html`

## Open Questions

1. Should `skill-harness` self-scaffold immediately, or should that be part of `skill-harness-wqp`?
2. Should model mappings live only in artifact Markdown front matter, or also in a normalized `.skill-harness/model-index.json`?
3. Should `agent-docs` own model/spec linkage once a repo has specgraph, or should harness checkers stay independent?
4. Should generated browser artifacts be committed in this repo, or kept generated-only except for selected review reports?
5. Should the first implementation target `skill-harness` itself or a pilot downstream project?

## Recommended Next Move

Implement `skill-harness-wqp` first enough to self-scaffold this repo's artifact loop, then implement `skill-harness-wdv` as the reusable suite graph and drift check.

The browser artifact gap should be treated as both a workflow issue and a repo setup issue:

- workflow issue: agents should produce generated review surfaces when the user asks for human artifacts
- setup issue: the root repo does not yet have the scaffolded manifest/check scripts that would make this behavior automatic
