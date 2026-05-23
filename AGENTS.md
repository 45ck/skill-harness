# Skill Harness

This repository provides the umbrella install and setup layer for the 45ck stack.

Guidelines:
- Use the narrowest agent that can own the task end to end.
- Escalate to another specialist when the work crosses a real boundary, not just for vocabulary changes.
- Prefer the curated skills listed in each agent before reaching for unrelated skills.
- Treat packs as the capability library and agents as the workflow loadout.
- Treat this repo as the suite entrypoint and project setup repo, including embedded packs under `packs/`.
- Treat `frontier-agent-playbook` as the shared doctrine companion when a task needs frontier-capability priors, agentic thinking, or anti-fallback framing.

UML-first artifacts:
- Auto-detect model impact for every engineering change. If code, API, workflow, dependency, deployment, UI structure, or agent behavior changes, update the relevant canonical model source or record why no model change is required.
- Use the modeling system by default for fresh developer-artifact setups. `--modeling-mode auto` preserves legacy repos, defaults fresh repos to `uml-first`, and can be overridden with `--modeling-mode off|baseline|uml-first` or `--skip-modeling`.
- Keep canonical UML/UWE/C4/evidence model sources in repo-relative text files, preferring `docs/artifacts/source/models/` when no domain-specific docs path is better.
- Keep `docs/artifacts/source/models/model-inventory.md` and `docs/artifacts/artifacts.manifest.json` aligned with model ids, owners, methods, source paths, evidence, and generated review surfaces.
- Human review artifacts for models must be static HTML under `generated/review/models/`, generated from source with `node scripts/generate-model-review.mjs`, and validated with the model and HTML policy checks.
- Open generated HTML in the best human review surface for the current environment. In Codex app, use the Browser plugin for local HTML when available; in Claude desktop, use the built-in browser/preview when available; in CLI-only contexts, use `node scripts/open-artifact-review.mjs` to open the system default browser or print the file URL in headless/CI contexts.
- Use `node scripts/open-artifact-review.mjs --json --print` when an agent or host app needs to auto-detect the artifact target and choose between Codex Browser, Claude preview, system browser, or a local HTTP preview fallback.
- Treat HTML, SVG, PNG, screenshots, and generated comparison pages as review surfaces only. Source files and model diffs remain canonical.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
