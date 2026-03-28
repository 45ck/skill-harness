#!/usr/bin/env bash
set -euo pipefail
AGENTS=(requirements-analyst requirements-analyst-beads ux-researcher system-modeler system-modeler-beads software-architect software-architect-beads web-engineer backend-engineer test-designer test-designer-beads qa-automation-engineer quality-reviewer security-reviewer security-reviewer-beads pentest-reviewer delivery-manager delivery-manager-beads research-writer)
for agent in "${AGENTS[@]}"; do
  rm -f "$HOME/.claude/agents/$agent.md" "$HOME/.codex/agents/$agent.toml"
done
