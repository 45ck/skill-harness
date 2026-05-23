# Security Policy

`skill-harness` installs agent workflows, copies skills, runs setup commands in target repositories, and can generate review artifacts. Treat changes here as supply-chain-sensitive.

## Reporting A Vulnerability

Do not disclose suspected vulnerabilities in a public issue.

Use GitHub private vulnerability reporting for `45ck/skill-harness` when it is available. If private reporting is not available, open a public issue that requests a private security contact without including exploit details.

Useful report details:

- affected command, script, pack, or generated artifact
- reproduction steps in a disposable repo or environment
- expected and actual behavior
- whether secrets, credentials, private files, network calls, or destructive filesystem operations are involved
- affected operating systems

## Scope

In scope:

- unsafe install or uninstall behavior
- unexpected writes outside the requested target directory or documented user-level install locations
- untrusted skill or pack intake paths
- generated review artifacts that execute scripts, load external assets, leak secrets, or make network calls
- dependency or installer behavior that weakens target repo security
- prompt-injection or tool-poisoning paths in bundled skills and agent instructions

Out of scope:

- issues in third-party repositories before they are adopted into this repo
- intentionally local experiment fixtures under [experiments](experiments)
- missing hardening in downstream target repos unless caused by `skill-harness` setup behavior

## Supported Versions

Security fixes target the default branch first. Until formal releases are published, use the latest commit from the default branch or a current release bundle built from it.

## Security Expectations For Contributions

- Do not commit secrets, private logs, raw traces, tokens, credentials, customer data, or machine-local paths that reveal sensitive information.
- Keep generated HTML review artifacts static and self-contained unless a maintainer explicitly approves an exception.
- Do not add external scripts, external assets, telemetry, network calls, or hidden executable helpers to generated artifacts.
- Review license, provenance, install scripts, MCP configuration, helper executables, and autonomy assumptions before adopting third-party skill content.

See [docs/dependency-provenance.md](docs/dependency-provenance.md) for dependency source and pinning expectations.
