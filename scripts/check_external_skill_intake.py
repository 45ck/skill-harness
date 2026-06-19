#!/usr/bin/env python3
"""Check synthetic external skill intake fixtures."""

from __future__ import annotations

from external_skill_intake import REPO_ROOT
from external_skill_intake import format_report
from external_skill_intake import summarize_repo


FIXTURE_ROOT = REPO_ROOT / "tests" / "fixtures" / "external-skill-intake"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def main() -> int:
    repos = sorted([path for path in FIXTURE_ROOT.iterdir() if path.is_dir()], key=lambda path: path.name)
    summaries = {summary.name: summary for summary in (summarize_repo(path, set()) for path in repos)}

    clean = summaries["clean-cross-host-pack"]
    require(len(clean.skill_names) == 1, "clean fixture should expose one SKILL.md")
    require(clean.cursor_rule_count == 1, "clean fixture should expose one Cursor .mdc rule")
    require(clean.host_instruction_count == 2, "clean fixture should expose AGENTS.md and Copilot instructions")
    require(clean.copilot_instruction_count == 1, "clean fixture should expose one Copilot instruction file")
    require(clean.plugin_manifest_count == 1, "clean fixture should expose one plugin manifest")
    require(clean.plugin_surface_count == 1, "clean fixture should expose one plugin surface directory")
    require(clean.mcp_config_count == 1, "clean fixture should expose one MCP config")
    require(clean.license_file_count == 1, "clean fixture should expose one license file")
    require(clean.trust_recommendation == "manual-review-executable-or-mcp", "MCP fixture should require manual review")
    require("cursor-rules" in clean.install_surface, "clean fixture should classify Cursor rule install surface")
    require("host-instructions" in clean.install_surface, "clean fixture should classify host instruction surface")

    risky = summaries["risky-script-pack"]
    require(risky.license_file_count == 0, "risky fixture should have no license")
    require("remote-shell-exec" in risky.blocked_flags, "risky fixture should block remote shell execution")
    require(risky.trust_recommendation == "quarantine", "risky fixture should be quarantined")

    report = format_report(list(summaries.values()), REPO_ROOT / ".agents" / "skills", sample_size=3)
    require("Cursor rules" in report, "report should include Cursor rule column")
    require("Host instructions" in report, "report should include host instruction column")
    print("External skill intake fixture check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
