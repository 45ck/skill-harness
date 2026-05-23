# Dependency Provenance

`skill-harness` installs from a mix of local embedded packs, Git repositories, and package manager sources. Treat these as supply-chain inputs.

## Machine-Readable Source

[scripts/dependencies.json](../scripts/dependencies.json) is the machine-readable source for shared pack and agent dependency mapping.

Repository entries use one of these forms:

- `url`: a Git remote that `skill-harness` can clone or update
- `path`: a repo-local embedded pack under [packs](../packs)

Project setup package installs are currently implemented in [cmd/skill-harness/main.go](../cmd/skill-harness/main.go):

- `@45ck/noslop` from the configured package manager registry
- `github:45ck/agent-docs` as a development dependency
- `github.com/steveyegge/beads/cmd/bd@latest` through `go install` when possible
- upstream Beads install scripts as a fallback when `bd` and Go installation are unavailable

## Pinning Policy

The current default is moving-head trust for external pack Git repositories and latest-compatible package manager resolution for project setup dependencies. This keeps the suite easy to update, but it is not a reproducible lockfile policy.

For higher assurance:

- review [scripts/dependencies.json](../scripts/dependencies.json) before `install --all`
- install Beads manually before `setup-project`, or pass `--skip-beads`
- run `setup-project --install-only` before initialization when you want to inspect installed packages first
- use `--scope workspace` to avoid monorepo-root lifting
- pin dependencies in the target repo after setup when reproducibility matters

## Embedded Packs

Embedded packs are first-party repo-local sources:

- `agent-operating-skills`
- `coding-workflow-skills`
- `demo-production-skills`
- `design-tooling-skills`
- `developer-artifact-skills`
- `integration-tooling-skills`
- `noslop-skills`
- `specgraph-skills`

## Third-Party Intake

Do not add third-party skill catalogs directly to the shared install flow. Follow [third-party skill intake](third-party-skill-intake.md), review license and provenance, inspect helper scripts and tool configuration, and prefer curated first-party rewrites under `packs/`.

