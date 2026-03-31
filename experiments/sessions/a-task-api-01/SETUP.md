# Group A Setup — Full Toolkit

Group A runs with specgraph, noslop, and skill-harness skills active. Follow these steps exactly before starting the Claude Code conversation.

## Prerequisites

- Node.js >= 22
- npm >= 10
- skill-harness binary available (`skill-harness` or `./skill-harness`)
- Beads (`bd`) installed and available on PATH (optional but recommended for issue tracking)

## Step 1: Create the session directory

Use `run-session.sh` to scaffold the directory:

```bash
./experiments/run-session.sh a <session-name>
# Example:
./experiments/run-session.sh a auth-module-01
```

This creates `experiments/sessions/a-<session-name>/` and changes into it.

## Step 2: Initialize npm

```bash
cd experiments/sessions/a-<session-name>
npm init -y
```

## Step 3: Install specgraph

```bash
npm install @45ck/agent-docs
```

This installs the `specgraph` binary (also aliased as `agent-docs`). Verify:

```bash
npx specgraph --version
```

## Step 4: Install noslop

```bash
npm install @45ck/noslop
npx noslop install
```

`noslop install` sets up the git pre-commit hook, CI check stubs, and Claude Code guardrails. Accept all defaults when prompted.

Verify hooks are in place:

```bash
cat .git/hooks/pre-commit
```

## Step 5: Install skill packs

From the skill-harness root, install both packs into `~/.claude/skills/`:

```bash
./skill-harness install --packs specgraph-skills noslop-skills --packs-only
```

Or install manually:

```bash
cp -r packs/specgraph-skills/skills/* ~/.claude/skills/
cp -r packs/noslop-skills/skills/* ~/.claude/skills/
```

## Step 6: Initialize specgraph config

```bash
npx specgraph init
```

This creates `.specgraph/config.json`. Accept defaults. Commit the config:

```bash
git init
git add .specgraph/config.json package.json package-lock.json
git commit -m "chore: init project with specgraph and noslop"
```

## Step 7: Create AGENTS.md

Create `AGENTS.md` at the project root with the following content:

```markdown
# Agent instructions

This project uses specgraph for spec tracking and noslop for quality gates.

## Before writing any code

1. Run `/spec-writer` to create a spec document in `docs/` for the feature you are implementing.
2. Confirm the spec is committed before proceeding.

## While writing code

3. Annotate every source file with `@spec` and `@implements` JSDoc tags using `/annotation-writer`.
4. Run `npx specgraph verify` after each significant change and review output with `/verify-interpreter`.

## Before committing

5. Run `/noslop-commit-gate` and resolve all failures before committing.
6. Run `npx specgraph verify` one final time. All requirements must have E1 or higher evidence.

## If a requirement cannot be satisfied

Use `/waiver-writer` to write a justified waiver with an expiry date.
```

## Step 8: Start the session

Open a fresh Claude Code conversation in `experiments/sessions/a-<session-name>/`.

Paste the task prompt from `session-task.md` as the first and only message.

Do not send any follow-up messages. Let the agent run to completion.

## What to expect from a well-run Group A session

- The agent invokes `/spec-writer` early and creates a spec document under `docs/`.
- Source files contain `@spec`/`@implements` annotations.
- `npx specgraph verify` passes or shows only justified waivers.
- Commits are gated by noslop hooks.
- The agent uses `/evidence-gap-review` if verification reports gaps.
- `npm test` exits 0 at the end.
