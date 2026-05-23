# Skill Harness Opportunity Review

Date: 2026-05-23

Beads issue: `skill-harness-22q`

Generated review surface: `generated/review/skill-harness-opportunity-review-2026-05-23.html`

## Purpose

This source note captures candidate additions to `skill-harness` after reviewing the public `45ck` GitHub organization, local adjacent public repos, and five specialist agent brainstorms. It is an opportunity review, not an accepted roadmap.

Private or internal repo details are intentionally omitted from this tracked source file. They were considered only as broad usage signals for app bootstrap, desktop agent workflows, and review artifacts.

## Evidence Base

- `skill-harness` is currently the umbrella installer and generator for the shared skill pack suite, project setup, Beads, noslop, agent-docs, and developer artifact scaffolding.
- The public org shows a mature skill-pack cluster already represented in `scripts/dependencies.json`.
- The public org also shows adjacent companion tools that are not yet first-class harness capabilities: `demo-machine`, `content-machine`, `video-evaluator`, `manual-qa-machine`, `prompt-language`, `Portarium`, `agent-docs`, and `noslop`.
- Local `demo-machine` docs show a browser capture and render pipeline with post-run artifacts such as `events.json`, `quality.json`, analyzer bundles, storyboards, and review prompts.
- Local `manual-qa-machine` docs show deterministic QA flows with screenshots, network, console, accessibility, performance, Markdown, and JSON reports.
- Local `prompt-language` docs show a verification-first execution runtime with deterministic loops, gates, and parallel agents.

## Recommended Opportunity Order

### 1. Capability Profiles And Suite Catalog

Create one machine-readable suite catalog that classifies repos and embedded packs as skill packs, doctrine, companion tools, overlays, runtime integrations, generated agents, and profiles.

First profiles should be conservative:

- `minimal`: core agents and skill packs only.
- `default`: current setup-project behavior.
- `desktop-review`: developer artifacts, generated HTML review surfaces, and app-browser-friendly defaults.
- `cli`: Markdown/TOON-first artifacts, no generated HTML by default.
- `media`: demo-machine, manual QA, video analysis, and demo packaging skills.
- `governed-agent`: prompt-language and Portarium-aligned governance guidance.

Why: this converts one-off flags into understandable install choices and keeps optional features like developer artifacts and media workflows default-on only where they fit.

### 2. Demo And Media Production Pack

Add an embedded incubating pack for demo production workflows. The pack should not become a generic video editor. It should specialize in structured post-capture outputs from `demo-machine`, `manual-qa-machine`, and video analysis artifacts.

Candidate skills:

- `demo-social-cut`: create short silent cuts from an existing demo run.
- `demo-slideshow-edit`: produce no-caption slideshow style edits from screenshots, storyboard frames, and selected spans.
- `demo-review-surface`: build a static local review page that compares spec, media, evidence, and quality.
- `qa-to-demo`: convert manual QA findings or flows into a demo-machine spec or reproducible clip plan.
- `demo-release-packager`: assemble approved media, source spec, and evidence for docs or marketing handoff.

Why: this directly matches the existing public tool cluster around `demo-machine`, `content-machine`, and `video-evaluator`, and it gives the user's no-caption polished video idea a narrow, useful shape.

### 3. Install Proof Artifact

After `setup-project`, emit a machine-readable proof artifact that records:

- detected package manager and setup scope
- selected profile and skipped capabilities
- installed tools
- Beads status
- noslop doctor output summary
- agent-docs/specgraph gate status where available
- developer artifact profile and policy checker status
- rendered agent paths

Why: `skill-harness` should prove what it installed instead of relying on narrative setup logs.

### 4. Integration Test Harness

Add end-to-end tests for the mutation-heavy commands using temp homes, temp repos, and fixture packs:

- `install`
- `render`
- `check`
- `setup-project`
- `beads-worktrees`

Why: the current suite has good unit coverage for setup parsing and artifact scaffold behavior, but the highest-risk behavior mutates user-home and repo-local surfaces.

### 5. Workflow Automation And CI

Add repo-level CI and drift checks:

- `go test ./...`
- build the CLI
- compile Python scripts
- run dependency and loadout checks
- smoke a temp `setup-project`
- validate docs/loadout/dependency drift

Why: this repo installs discipline into other repos, so it should have its own visible discipline.

### 6. Cross-Repo Adoption Audit

Turn the existing adoption audit idea into a command that scans local workspaces or the public org metadata and reports:

- harness-installed repos
- generated-only repos
- repos missing `.skill-harness/project.json`
- repos with stale rendered agents
- repos with missing descriptions or topics
- candidate profiles per repo

Why: the org has many related repos and uneven metadata. A harness-owned audit command would make suite drift visible.

### 7. Prompt-Language Bridge

Add a companion profile or pack that teaches installed agents how to run under `prompt-language` flows with real completion gates, child agents, retries, and verification loops.

Why: this is the natural runtime complement to static skill and agent installation.

### 8. Governed-Agent Profile

Add Portarium-aligned guidance for policy, approvals, audit evidence, and human-in-the-loop action boundaries.

Why: the org already has a public governed operations control plane. `skill-harness` can prepare repos to use those patterns without pulling the whole runtime into every project.

### 9. Metadata And Branding Hygiene

Use repo-branding and org metadata conventions to propose descriptions, topics, social previews, and README badges for repos with sparse public metadata.

Why: discovery and consistency are becoming part of the suite's product surface.

### 10. Consumer Conformance Kit

Create small fixture repos for `minimal`, `default`, `desktop-review`, `cli`, and `media` profiles. Each fixture should have expected proof artifacts and quality outputs.

Why: this turns the experiment story into repeatable regression evidence.

## Demo-Machine Video Skill Shape

The strongest version of the user's idea is a `demo-production-skills` embedded pack with a `media` profile.

Inputs:

- `.demo.yaml`
- completed `demo-machine` output directory
- `events.json`
- `quality.json`
- analyzer outputs such as storyboard and segment evidence
- optional `manual-qa-machine` report or finding

Outputs:

- silent short cut
- slideshow-style MP4
- poster frames
- frame strip
- local HTML review artifact
- handoff bundle with source spec and evidence summary

Default behavior:

- In desktop apps such as Codex app or Claude Desktop, prefer generated HTML review surfaces because users can inspect media, specs, and evidence visually.
- In CLI/TUI workflows, prefer Markdown/TOON source and a plain handoff bundle. HTML remains opt-in.
- Keep the source spec durable and regenerate media/review surfaces from it.

Guardrails:

- Do not load external media or scripts into review artifacts.
- Do not present arbitrary video editing as the goal.
- Do not make silent cuts the default replacement for narrated demos.
- Require evidence links back to source runs.
- Keep generated media out of git unless explicitly promoted.

## First Implementation Slice

The best first slice is:

1. Add `media` as a named profile in planning docs.
2. Add embedded `packs/demo-production-skills` with `demo-social-cut`, `demo-review-surface`, and `qa-to-demo`.
3. Add a generated HTML review artifact template for demo runs.
4. Add docs that explain desktop versus CLI defaults for media artifacts.
5. Add follow-up issues for install-proof artifacts and integration tests.

This gives the demo/video idea a concrete home while keeping the harness from becoming a broad media application.
