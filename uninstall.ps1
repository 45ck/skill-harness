$agents = @('requirements-analyst','requirements-analyst-beads','ux-researcher','system-modeler','system-modeler-beads','software-architect','software-architect-beads','web-engineer','backend-engineer','test-designer','test-designer-beads','qa-automation-engineer','quality-reviewer','security-reviewer','security-reviewer-beads','pentest-reviewer','delivery-manager','delivery-manager-beads','research-writer')
foreach ($agent in $agents) {
  Remove-Item -LiteralPath (Join-Path $HOME ".claude\agents\$agent.md") -Force -ErrorAction SilentlyContinue
  Remove-Item -LiteralPath (Join-Path $HOME ".codex\agents\$agent.toml") -Force -ErrorAction SilentlyContinue
}
