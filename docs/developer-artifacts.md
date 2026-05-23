# Developer Artifacts

`skill-harness setup-project` scaffolds developer artifact guidance by default. The capability is intentionally project-local: it shapes how a target repo records plans, decisions, evidence, and generated review surfaces without creating new global agent variants.

## Position

Developer artifacts use this source-of-truth split:

- canonical source: Markdown, TOON, specgraph / `agent-docs`, or existing project docs
- generated review surface: static HTML under `generated/review/`
- handoff evidence: linked files, Beads issues, tests, reports, logs, screenshots, or runtime proof

HTML is never the only durable source for a decision. Edit the source first, then regenerate or discard the review surface.

## Setup Behavior

Default setup:

```bash
./skill-harness setup-project --dir ../my-project
```

This creates:

- `.skill-harness/project.json`
- `docs/artifacts/source/`
- `docs/artifacts/templates/`
- `generated/review/`
- `scripts/check-artifact-html-policy.mjs`

It also adds `generated/review/` to `.gitignore` and adds package scripts when applicable:

- `docs:check`
- `docs:generate`
- `docs:report`
- `artifacts:html:check`

If `--skip-agent-docs` is used, the artifact scaffold still works, but it does not add scripts that call `agent-docs`.

## Profiles

Use `--developer-artifacts-profile` to select the target workflow:

| Profile | Effective Mode | Intended Use |
|---|---|---|
| `auto` | `dual` | Default source-first workflow with optional generated review surfaces |
| `codex-app` | `html` | Codex app workflows where file-backed previews are useful |
| `claude-desktop` | `html` | Desktop preview workflows where a generated HTML artifact helps review |
| `cli` | `markdown` | Terminal-heavy work where paths and Markdown are the primary interface |
| `tui` | `markdown` | TUI work where HTML should remain secondary |
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

- no external scripts
- no external assets
- no network calls
- no inline JavaScript unless explicitly reviewed and allowed by the project
- required CSP meta tag
- semantic headings, landmarks, meaningful link text, and alt text
- no secrets, tokens, credentials, private logs, customer data, or large opaque blobs

Run:

```bash
node scripts/check-artifact-html-policy.mjs
```

The checker rejects common unsafe constructs including `<script>`, iframes, object/embed/form tags, meta refresh, external `src` / `href` / `action` references, and browser APIs such as `fetch`, `XMLHttpRequest`, `WebSocket`, `EventSource`, `sendBeacon`, `serviceWorker`, `document.cookie`, `localStorage`, and `sessionStorage`.

## Skill Pack

The embedded `developer-artifact-skills` pack provides:

- `developer-artifact-shaper`: choose artifact type, canonical source, and review surface
- `html-review-artifact`: create safe generated HTML review artifacts
- `artifact-evidence-gate`: check source links, evidence, freshness, and safety
- `artifact-handoff-pack`: assemble the minimal handoff bundle

These skills are wired into the author, reviewer, delivery, research, and workflow loadouts where artifact decisions naturally happen.

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
- `README.md`
- `AGENT_INSTRUCTIONS.md`

Run:

```bash
go test ./cmd/skill-harness
node -e "JSON.parse(require('fs').readFileSync('scripts/dependencies.json','utf8')); JSON.parse(require('fs').readFileSync('scripts/agent_loadouts.json','utf8'))"
```

