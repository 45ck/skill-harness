# Group B Setup — Baseline (Raw Claude)

Group B runs with no tooling. The agent receives only the task prompt and nothing else. Follow these steps exactly.

## Prerequisites

- Node.js >= 22 (required to run the output, not for any tooling)
- npm >= 10

## Step 1: Create the session directory

Use `run-session.sh` to scaffold the directory:

```bash
./experiments/run-session.sh b <session-name>
# Example:
./experiments/run-session.sh b auth-module-01
```

This creates `experiments/sessions/b-<session-name>/` as a plain empty directory.

## Step 2: Verify nothing is installed

The session directory must be clean:

- No `package.json` (the agent will create one)
- No `.specgraph/` or `.agent-docs/` directory
- No `AGENTS.md` or `CLAUDE.md`
- No `~/.claude/skills/` entries loaded for this session (do not install skill packs)
- No noslop hooks

If you previously ran a Group A session and installed skills globally into `~/.claude/skills/`, you must temporarily remove them for a clean Group B comparison:

```bash
# Temporarily move skills out of the way
mv ~/.claude/skills ~/.claude/skills.bak
# Restore after the session
mv ~/.claude/skills.bak ~/.claude/skills
```

## Step 3: Start the session

Open a fresh Claude Code conversation in `experiments/sessions/b-<session-name>/`.

Paste the task prompt from `session-task.md` as the first and only message.

Do not send any follow-up messages. Let the agent run to completion.

## What the agent has access to

- The task prompt
- Claude's built-in knowledge
- Any standard CLI tools available in the shell (git, node, npm, etc.)
- Nothing else

## What the agent does NOT have access to

- specgraph / agent-docs
- noslop
- skill-harness skills
- AGENTS.md guidance
- Any pre-defined workflow instructions

## Fairness note

The baseline condition isolates the value of the toolkit. If Group B produces comparable results, the overhead of the toolkit may not be worth it for simple tasks. If Group A scores significantly higher, the toolkit is delivering value. Either outcome is informative.

Do not help the agent or suggest approaches. Score the output as-is when it declares completion.
