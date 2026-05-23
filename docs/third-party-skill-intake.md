# Third-Party Skill Intake

This repo should review external skill ecosystems, but it should not absorb them as raw dependencies.

## Why

- `skill-harness` owns install tooling, shared workflow agents, embedded packs, and harness-specific helper skills.
- Third-party repos are best treated as intake material until a specific pattern, skill, or install idea earns a first-party home under `packs/` or another harness-local surface.

This matches the repo-placement rule in [docs/repo-placement.md](./repo-placement.md): shared cross-pack workflow composition stays here, and incubating reusable skill content can now live under repo-local `packs/`.

## Current shortlist

Clone third-party repos into a sibling workspace, not inside this repo and not into the live install cache under `~/.skill-harness/packs`.

Suggested sibling workspace:

```bash
../skill-intake
```

Current high-signal shortlist:

- `openai/skills`
- `anthropics/skills`
- `vercel-labs/skills`
- `ComposioHQ/awesome-codex-skills`
- `ComposioHQ/awesome-claude-skills`
- `troykelly/codex-skills`
- `Dimillian/Skills`
- `alirezarezvani/claude-skills`
- `sickn33/antigravity-awesome-skills`
- `VoltAgent/awesome-agent-skills`

## Intake rule

Do this:

- clone outside the harness repo
- inspect repo structure, license, install surface, skill layout, and any helper scripts
- compare with the existing 45ck catalog before proposing imports
- map every worthwhile idea to a first-party destination

Do not do this:

- add third-party repos directly to `scripts/dependencies.json`
- treat stars or repo size as proof of quality
- install third-party skills into the shared suite without provenance review
- copy entire external catalogs into `skill-harness`

## What to look for

Prioritize four buckets:

- official format guidance and reference implementations
- install/distribution patterns we may want to reuse
- tool-specific skills we do not already cover
- reusable helper patterns for repo intake, agent handoffs, or multi-tool packaging

### Highest-value gaps found in the March 2026 pass

These are the clearest opportunities relative to the current 45ck catalog.

1. Tool-specific GitHub workflow skills
   Suggested destination: the embedded `packs/coding-workflow-skills` pack.
   Source examples: `gh-address-comments`, `gh-fix-ci`, `github`, `pr-creation`, `ci-monitoring`.

2. Figma and design-tool execution skills
   Suggested destination: the embedded `packs/design-tooling-skills` pack.
   Source examples: `figma-*`, `canvas-design`, `theme-factory`.

3. MCP builder and app-connection skills
   Suggested destination: the embedded `packs/integration-tooling-skills` pack.
   Source examples: `mcp-builder`, `connect`, `connect-apps`.

4. Repo and skill audit skills
   Suggested destination: this repo only if the skill is harness-specific.
   Source examples: `project-skill-audit`, external repo catalogs, installer CLIs.

5. Distribution ideas
   Suggested destination: harness tooling.
   Source examples: plugin marketplaces, installer CLIs, bundle docs, update checks, project-scoped installs.

## Embedded adoption from this pass

The current repo now embeds the first adoption wave directly under `packs/`:

- `packs/coding-workflow-skills`
- `packs/design-tooling-skills`
- `packs/integration-tooling-skills`

These are first-party rewrites inspired by the external review, not raw third-party vendoring.

## Current conclusions

### Strong keep

- Keep `skill-harness` focused on installer logic, workflow agents, embedded packs, and harness-specific helper skills.
- Keep third-party repo clones in a sibling intake workspace.

### Worth reusing

- The official repos are useful as format and structure references.
- The large catalogs are useful for discovery and gap scanning.
- `vercel-labs/skills` is useful as a packaging and install-surface reference.
- `troykelly/codex-skills` is useful as a Codex-native workflow reference, but not as a direct dependency.

### Not recommended

- Do not wire giant public catalogs into the default install flow.
- Do not promote third-party skills into shared installs until they are rewritten or curated into embedded first-party packs.
- Do not let third-party helper scripts run inside the shared install path without review.

## Security and provenance review

Treat third-party skills like code dependencies.

Review:

- license and provenance
- prompt injection or tool-poisoning patterns
- hidden network, filesystem, or destructive command assumptions
- unsafe autonomy defaults such as bypassed approvals or sandbox disabling
- weak install stories that blur project-local and user-global state

If a skill is useful but risky, port the idea manually into a first-party pack and rewrite the unsafe parts.

## How to rerun the scan

Use the repo-local intake script after cloning or updating the shortlist:

```powershell
python scripts/external_skill_intake.py --output docs/external-skill-intake-report.md
```

Or point it at explicit repos:

```powershell
python scripts/external_skill_intake.py `
  "..\skill-intake\openai-skills" `
  "..\skill-intake\anthropics-skills" `
  "..\skill-intake\troykelly-codex-skills"
```

On Unix-like shells, pass the same sibling paths with `/` separators.

The script only inventories repo structure and skill names. Adoption decisions still need human review.

## Intake gate output

The intake script now reports the fields that matter before any external repo can be considered for adoption:

- `SKILL.md` skill count and overlap with the local 45ck catalog
- agent files and plugin surfaces such as `.claude-plugin`, `.codex-plugin`, and `plugin.json`
- Codex metadata files such as `agents/openai.yaml`
- MCP config files such as `.mcp.json`
- lock, index, or manifest files such as `skills_index.json`
- license and third-party notice files
- executable helpers such as shell, PowerShell, Python, JavaScript, and TypeScript files
- blocked flags for approval bypass, remote shell execution, or secret exfiltration patterns
- review flags for token references, MCP/tool config, install scripts, and executable helpers
- a trust recommendation: `intake-candidate`, `manual-review-*`, or `quarantine`

Use machine-readable output when turning a review into tracked work:

```powershell
python scripts/external_skill_intake.py `
  --output docs/external-skill-intake-report.md `
  --json-output docs/external-skill-intake-report.json
```

Use the blocking mode in CI or before proposing a new dependency:

```powershell
python scripts/external_skill_intake.py `
  "..\skill-intake\some-public-skill-repo" `
  --fail-on-blocked
```

`--fail-on-blocked` is intentionally narrow. It catches patterns that should not flow into shared installs, but it does not replace license review, script review, MCP review, or a first-party rewrite decision.
