# specgraph-skills

Embedded pack for [specgraph](https://github.com/45ck/agent-docs) (`github:45ck/agent-docs`) day-to-day workflow.

Included skills:

- `annotation-writer` — Add `@spec`/`@implements` JSDoc annotations to source files to produce E0 evidence
- `evidence-gap-review` — Analyse `specgraph verify` output and produce a prioritised action list
- `spec-writer` — Write spec documents with correct YAML frontmatter and evidence requirements
- `verify-interpreter` — Interpret `specgraph verify` output and recommend next steps
- `waiver-writer` — Write justified evidence waivers with expiry and approval fields

## Prerequisites

specgraph must be installed in the project:

```bash
npm install --save-dev github:45ck/agent-docs
npx agent-docs init
```

Or bootstrap an entire project (installs specgraph, Beads, noslop, git hooks):

```bash
./skill-harness setup-project --dir ./my-project
```

The binary is available as both `specgraph` and `agent-docs`.

Config file lives at `.specgraph/config.json` (falls back to `.agent-docs/config.json` for legacy projects).

## Installation

### Via skill-harness

```bash
./skill-harness install --packs=specgraph-skills --packs-only
```

### Manual

```bash
cp -r packs/specgraph-skills/skills/* ~/.claude/skills/
```

## Quick reference

| Skill | Invoke when |
|---|---|
| `spec-writer` | Creating a new spec doc in `docs/` |
| `evidence-gap-review` | Analysing `specgraph verify` output to find lowest-effort fixes |
| `waiver-writer` | A requirement cannot currently be satisfied and needs formal bypass |
| `verify-interpreter` | Interpreting verification output and deciding next steps |
| `annotation-writer` | Adding `@spec`/`@implements` tags to source or test files |
