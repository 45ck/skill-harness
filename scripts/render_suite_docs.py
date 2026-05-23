#!/usr/bin/env python3
"""Render deterministic suite documentation from agent loadout data."""

from __future__ import annotations

from pathlib import Path

from suite_graph import ROOT, build_graph


OUT_PATH = ROOT / "docs" / "agent-loadouts.md"


def render_markdown() -> str:
    graph = build_graph()
    lines = [
        "# Agent Loadouts",
        "",
        "This document maps each shared `skill-harness` agent to its curated skill set.",
        "",
    ]
    for agent, data in graph["agents"].items():
        skills = ", ".join(data.get("skills", []))
        lines.extend([
            f"## {agent}",
            "",
            f"- Skills: {skills}",
            "",
        ])
    return "\n".join(lines).rstrip() + "\n"


def main() -> int:
    OUT_PATH.write_text(render_markdown(), encoding="utf-8")
    print(f"Rendered {OUT_PATH.relative_to(ROOT).as_posix()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
