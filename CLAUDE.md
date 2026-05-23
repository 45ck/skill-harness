# Skill Harness

<!-- bd-doctor-divergence: ok -->

Use the narrowest specialist agent that can own the work end to end. Treat this repo as the suite entrypoint and project setup repo for the 45ck stack, including embedded packs under `packs/`.

This is maintainer and agent operating guidance. Public contributors should start with `CONTRIBUTING.md`, `SUPPORT.md`, and `SECURITY.md`; fork-based contributors do not need Beads locally and should use pull requests instead of direct pushes.

## UML-First Developer Artifacts

- Auto-detect model impact for every engineering change. When code, API, workflow, dependency, deployment, UI structure, or agent behavior changes, update the relevant model source or record why no model change is needed.
- Fresh developer-artifact setups use `--modeling-mode auto`, which resolves to `uml-first` for new repos and preserves existing repos. Use `--modeling-mode off|baseline|uml-first` or `--skip-modeling` only when that is intentional.
- Keep canonical UML/UWE/C4/evidence model sources in repo-relative text files. Prefer `docs/artifacts/source/models/` when no better domain docs path exists.
- Keep `docs/artifacts/source/models/model-inventory.md` and `docs/artifacts/artifacts.manifest.json` current with model ids, owners, methods, source paths, evidence, and generated review surfaces.
- Human model artifacts belong in static HTML under `generated/review/models/`. Generate them with `node scripts/generate-model-review.mjs`; validate with `node scripts/check-model-artifact-policy.mjs` and `node scripts/check-artifact-html-policy.mjs`.
- Human-facing discovery, planning, research, product, business, data, and UX artifacts belong in source-backed infographic HTML under `generated/review/<family>/`. Create the canonical source first, mark the manifest entry `reviewRequired: true`, generate with `node scripts/generate-artifact-review.mjs`, then surface/open the HTML review page.
- Default generated review interaction is CSS-only: radio tabs, details/summary, anchor navigation, and inline SVG states. Do not add inline JavaScript; the reviewed inline-JS lane is reserved until manifest-aware checker and CSP support are implemented.
- Open generated HTML in the best available human review surface. In Codex app, prefer the Browser plugin for local HTML. In Claude desktop, prefer the built-in browser or preview. In CLI-only contexts, use `node scripts/open-artifact-review.mjs` to open the system default browser or print the file URL in headless/CI contexts.
- Use `node scripts/open-artifact-review.mjs --json --print` when a host workflow needs the resolved review target and preferred open action before choosing a browser, preview, or local HTTP fallback.
- Treat generated HTML, SVG, PNG, screenshots, and comparison pages as review surfaces only. Canonical truth stays in source artifacts and model diffs.

## Visual Source-First Artifacts

- Use source-first artifacts for product, business, data, research, UX, and mockup work: canonical agent-readable source first, generated visual human review surface second.
- Prefer `docs/artifacts/source/product/`, `business/`, `data/`, `research/`, and `ux/` for canonical sources and matching `generated/review/` subfolders for human surfaces.
- High-fidelity HTML/prototype review is the default for UI, product, customer-facing workflow, and mockup artifacts. Low-fidelity sketches are scratch only unless captured as explicit research evidence.
- Use Mermaid, Vega-Lite, Observable Plot, D3, Graphviz, Apache ECharts, RAWGraphs, and Chart.js as source/spec or generation-time infographic renderers only. Generated HTML must embed static SVG/HTML/data-url output and must not load browser chart runtimes.
- Add `artifact-infographic` JSON fences or manifest `infographics` entries when a non-model human artifact needs charts, graphs, or other infographic panels.
- Record generated visual artifacts in `docs/artifacts/artifacts.manifest.json` with source, review surface, owner, evidence links, status, and freshness. Keep synthetic user or agent-simulation evidence distinct from real user/customer evidence.
- Use multiple specialist agents when ownership crosses real boundaries: product, business, data, research, UX, structural modeling, and readiness review.

## Session Rules

- Use Beads for task tracking in this repo.
- Run the relevant Go tests and generated artifact checks after code changes.
- Keep public contribution, support, security, and conduct docs aligned when changing OSS-facing workflows.
- Push completed work to the remote before ending the session.
