# Adoption audit: skill-harness usage

Date: 2026-04-29

This audit checked active 45ck repositories for actual `skill-harness` usage.

## Verdict

`skill-harness` is not generally consumed as an imported package. It is used as a setup repo and generator. Downstream repos usually have copied `.claude/agents` and `.codex/agents` files or their own repo-local harness scripts.

That means a repo can contain skill-harness output without having a live `skill-harness` runtime. Treat those as generated assets unless the repo also runs `skill-harness check`, `skill-harness install`, or a documented setup-project flow.

## Active or partial users

| Repo | Observed usage | Status |
| --- | --- | --- |
| skill-harness | Go CLI owns `install`, `setup-project`, `check`, `beads-worktrees`, and `uninstall`. | Active source repo. |
| hydra-reach | Copied Claude and Codex agent definitions matching skill-harness output. | Generated-output-only. |
| content-machine | Has its own `scripts/harness` and JSON-stdio runtime. | Active repo-local harness, not this package. |
| video-evaluator | Has its own `agent/run-tool.mjs`, skills, and harness scripts. | Active repo-local harness, not this package. |
| demo-machine | Consumes `@45ck/video-evaluator` and tests generated skill scaffolding. | Indirect consumer, not skill-harness. |

## Weak spots found

- Generated Codex agent source files are not always the installed files Codex actually loads. Installing into the user Codex agent directory is a separate step.
- Existing Go tests mainly cover setup context, package-manager detection, scaffold generation, artifact checks, and hermetic setup behavior. They do not yet prove every install, check, and dependency-copy path end to end.
- Content-machine has a separate harness/catalog mismatch: at least one runnable skill documents an entrypoint in body text but not in frontmatter, so catalog output reports `entrypoint: null`.

## Adoption checks

For a repo claiming active skill-harness adoption, require one of these:

```sh
skill-harness check --all
skill-harness install --all
```

or a documented setup-project run:

```sh
skill-harness setup-project --dir .
```

If a repo only commits copied agent definitions, describe that state as "generated agent definitions" rather than "skill-harness installed."

## Follow-up

1. Add integration tests around `install`, `check`, and dependency copying against temporary home directories.
2. Document the difference between source agent templates and installed Codex agent configs.
3. Keep benchmark claims tied to controlled experiments, not broad downstream enforcement, unless consumer repos prove active wiring.
