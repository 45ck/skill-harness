# UML-Backed Developer Artifacts Research

Date: 2026-05-23

Beads issue: `skill-harness-per`

Status: proposal

## Purpose

This note investigates whether `skill-harness` should support a stronger UML/UWE-style modeling workflow for target repos such as `macquariecollege`, where agents can maintain canonical model sources, regenerate review surfaces, and show before/after model changes during feature work.

The short answer is yes, but it should be source-first model engineering, not diagram-first documentation. UML should become a governed developer artifact specialization over the existing `docs/artifacts` scaffold.

## Evidence Base

- `skill-harness` already scaffolds source-backed developer artifacts with canonical Markdown/TOON/specgraph sources, generated static HTML review surfaces, and artifact provenance metadata.
- The current model policy already allows Mermaid, Markdown, TOON, and PlantUML notations, and model kinds such as use-case, activity, sequence, state, class, domain, context, container, component, dynamic, deployment, dependency, and architecture-space.
- OMG UML 2.5.1 is the current formal UML specification and defines UML as a language for visualizing, specifying, constructing, and documenting distributed object system artifacts: https://www.omg.org/spec/UML/
- UWE is a UML-based Web Engineering approach that uses a UML profile, model-driven methodology, and web-domain models such as content, navigation, presentation, process, adaptation, and architecture: https://uwe.pst.ifi.lmu.de/aboutUwe.html
- UWE's own tutorial separates navigation, presentation, and process models, which maps well to artifact facets rather than a single diagram type: https://uwe.pst.ifi.lmu.de/teachingTutorialNavigation.html and https://uwe.pst.ifi.lmu.de/teachingTutorialProcess.html
- The C4 model recommends context, container, component, and code diagrams, but says teams do not need every level; context and container diagrams are enough for many teams: https://c4model.com/diagrams
- Structurizr DSL is a text-based way to define C4 architecture models and views, with export paths to PlantUML, Mermaid, SVG, PNG, and static sites: https://docs.structurizr.com/dsl
- PlantUML supports repo-friendly text-to-diagram workflows where use case diagrams and their relationships live next to code in version control: https://plantuml.com/use-case-diagram
- Mermaid supports text-based diagram sources but has renderer-specific syntax caveats, so generated HTML should embed pre-rendered output instead of relying on a browser runtime: https://mermaid.js.org/intro/syntax-reference
- GitHub pull requests are branch-based review artifacts, and `gh pr create` supports titles, bodies, base/head branches, draft PRs, labels, and reviewers: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request

## Core Position

Do not add a separate `uml` artifact universe. Add first-class modeling support as an opt-in specialization of developer artifacts.

Canonical truth:

- Markdown model artifacts
- TOON scenario/model fragments
- specgraph or `agent-docs` specs
- Mermaid source blocks
- PlantUML source files
- optional Structurizr DSL for C4-heavy architecture repos

Generated review surfaces:

- static HTML under `generated/review/`
- pre-rendered inline SVG or static markup
- before/after comparison pages
- dependency or runtime graphs generated from code, traces, or package metadata

Never treat generated HTML, PNG, SVG, or screenshots as the authoritative model.

## Model Taxonomy

Keep the current top-level model kinds and add UWE as metadata/facets.

Durable source by default:

- `use-case`: actors, goals, system responsibilities, acceptance links
- `activity`: workflows, business processes, navigation/process flows
- `sequence`: cross-boundary interactions and request lifecycles
- `state`: lifecycle-heavy entities, sessions, enrollment states, workflow states
- `domain`: ubiquitous language, entities, concepts, relationships, invariants
- `context`: system boundary and external actors/systems
- `container`: deployable/runtime units and data stores
- `component`: meaningful internal components when the team actually uses them
- `deployment`: environments, hosting, trust boundaries, data movement
- `architecture-space`: ownership, decisions, quality attributes, risks, dependencies

Durable only when intentionally designed:

- `class`: useful for deliberate design; risky when it is just reverse-engineered code
- `dynamic`: useful for key feature scenarios; generated trace projections should be evidence

Generated evidence by default:

- `dependency`: package, import, route, module, database, and API dependency graphs
- reverse-engineered class diagrams
- runtime topology snapshots
- trace-derived interaction graphs

## UWE Mapping

Represent UWE as structured metadata and source sections, not only diagram labels.

Recommended source metadata:

```yaml
method: uwe
uwe_facets:
  - content
  - navigation
  - presentation
  - process
  - access
```

Practical mapping:

- `content` -> `domain` model with entities, content objects, relationships, invariants
- `navigation` -> `activity` or `use-case` source with explicit navigation nodes and links
- `presentation` -> Markdown tables plus optional state/component projections for pages/forms
- `process` -> `activity` or `sequence` source for actions, service calls, and validations
- `access` or personalization -> actor/role/access matrix linked to auth and policy evidence

If this grows, add `navigation` as a model kind later. Do not start there; a new kind needs checker rules, templates, manifest support, and clear evidence standards.

## Before/After Model Changes

The canonical before/after diff should be the text diff in git. The visual review should be generated from that text.

Recommended artifact shape:

- artifact type: `model-view` for a single model state
- artifact type: `model-diff` for before/after review
- canonical source path: model source
- generated review path: comparison HTML
- evidence links: Beads issue, specs, code files, tests, traces, PR
- freshness: source hash, renderer, renderer version, generated timestamp

Phase 1 diff:

- compare canonical source text
- render before and after diagrams
- generate a static HTML review page showing source summary, old render, new render, evidence, and residual risks

Phase 2 diff:

- parse selected model sources into a simple intermediate representation
- compare model entities, relationships, messages, states, guards, actors, and responsibilities
- keep SVG/HTML comparison as presentation only

Avoid image-pixel diffing as the main signal. Layout changes are noisy and often unrelated to model meaning.

## Agent Workflow

Use multiple agents only where there is a real ownership boundary.

Default loop:

1. `system-modeler-beads` owns canonical model source updates.
2. `software-architect` owns architecture-level model choices, C4/Structurizr decisions, and tradeoffs.
3. `workflow-engineer` owns branch, PR, Beads issue state, checks, and handoff packaging.
4. `quality-reviewer` or `agent-run-evidence-reviewer` gates evidence, freshness, and redaction before claiming the model is ready.

Conflict rules:

- Beads issue scope beats ad hoc chat scope.
- Canonical model source beats generated HTML.
- Code, tests, and traces beat stale diagrams.
- If model and implementation disagree, mark the artifact `needs-evidence` or `inconclusive`.
- Only one agent edits a given model source file at a time.
- Reviewers can work in parallel on evidence and critique, not on the same source artifact.

Agent update rule during feature work:

- user goal or workflow changes -> update `use-case`, `activity`, and maybe `sequence`
- business concepts or rules change -> update `domain` and maybe `state`
- service/component boundaries change -> update `context`, `container`, `component`, and maybe `deployment`
- internal refactor only -> usually regenerate dependency evidence, not durable UML
- generated review is stale after source changes until checks regenerate it

## GitHub Governance For Shared Repos

For a repo the developer can access through GitHub CLI but does not own, default to least privilege.

Allowed without extra approval:

- read repo metadata
- inspect issues, PRs, checks, actions, files, and branches
- prepare local branches and local artifacts
- generate PR body text locally

Requires explicit human approval:

- `gh pr create`
- pushing to upstream rather than a fork
- requesting reviewers, labels, or milestone changes
- posting review comments
- changing repo settings or branch protection
- approving, merging, closing, or deleting branches
- expanding GitHub token scopes

Default recommendation for `macquariecollege`-style shared repos:

- push to a personal fork or a clearly named feature branch allowed by repo policy
- open draft PRs only after the user approves
- link model diffs in the PR body
- never auto-merge or approve own PR
- keep raw agent traces out of git; promote only redacted summaries

## Skill-Harness Implementation Shape

Add modeling as an opt-in specialization, not a replacement for current artifact profiles.

Possible setup surface:

```bash
./skill-harness setup-project --dir ../macquariecollege --enable-modeling
```

Scaffold additions:

- `docs/artifacts/source/models/`
- `docs/artifacts/templates/model-diff-artifact.md`
- `generated/review/models/`
- `scripts/check-model-artifact-policy.mjs`
- package scripts:
  - `models:check`
  - `models:diff:check`
  - `models:review`

Project config additions under `.skill-harness/project.json`:

```json
{
  "modelPolicy": {
    "canonicalSource": true,
    "generatedReviewOnly": true,
    "renderDiagramsOffline": true,
    "uml": {
      "enabled": true,
      "methods": ["uml", "uwe", "c4"],
      "allowedSourceExtensions": [".md", ".toon", ".mmd", ".puml", ".dsl"],
      "defaultDiffMethod": "source",
      "semanticDiff": false
    }
  }
}
```

Manifest additions:

- `modelId`
- `modelKind`
- `notation`
- `method`
- `facets`
- `abstractionLevel`
- `owner`
- `render.engine`
- `render.engineVersion`
- `render.generatedAt`
- `lineage.derivedFrom`
- `lineage.supersedes`
- `diff.beforeArtifactId`
- `diff.afterArtifactId`
- `diff.method`
- `diff.reviewSurface`

Checker additions:

- model source must be repo-relative and text-based
- rendered HTML must live under `generated/review/`
- `sourceHash` must match when present
- `model-diff` must reference valid before/after artifacts
- generated review cannot be marked canonical
- ready model artifacts need evidence links
- UWE facets must be known values

## Rollout Plan

Phase 0: use existing capability

- Create canonical Markdown model artifacts manually.
- Use Mermaid or PlantUML source blocks.
- Record evidence and source hashes in the existing manifest where target repos have it.
- Do not generate HTML unless the target repo already has artifact checks.

Phase 1: harness support

- Add `--enable-modeling`.
- Scaffold model directories, model diff template, and checker.
- Extend `.skill-harness/project.json` and setup proof.
- Add tests for generated project config and checker behavior.

Phase 2: review surfaces

- Add static before/after model review generator.
- Prefer inline SVG outputs from Mermaid/PlantUML/Structurizr export.
- Keep generated review out of git by default unless the target repo opts in.

Phase 3: semantic model diffs

- Define a small model intermediate representation.
- Support semantic diffs for use-case, sequence, activity, state, domain/class, and C4 views.
- Gate semantic claims with parser tests and fixture models.

Phase 4: repo adoption

- Run `setup-project --enable-modeling` in a pilot repo.
- Create an initial system context, domain model, and one feature sequence/activity pair.
- Require every model PR to show what changed, why, evidence, and stale risks.

## Anti-Patterns

- Making generated HTML or SVG the source of truth.
- Maintaining one giant "whole system UML" diagram.
- Hand-maintaining dependency graphs that can be generated from code.
- Treating reverse-engineered class diagrams as intended design.
- Encoding UWE semantics only inside diagram labels.
- Letting agents update code and models silently without a linked issue.
- Accepting a diagram because it looks plausible without tests, traces, specs, or code links.
- Adding a separate UML profile when existing artifact profiles already express how humans review artifacts.

## Recommended First Slice

Implement `--enable-modeling` in `skill-harness` with:

1. model source and review directories
2. model diff template
3. model policy config in `.skill-harness/project.json`
4. `model-diff` artifact type
5. model checker script
6. tests for scaffold output and checker failures

For `macquariecollege`, start smaller:

1. create `docs/artifacts/source/models/system-context.md`
2. create `docs/artifacts/source/models/domain-language.md`
3. create one feature-level `activity` or `sequence` model for the next real change
4. ask agents to update only those model files when the change invalidates them
5. include model before/after notes in the PR body

This gives you visual engineering leverage without turning UML into stale wallpaper.
