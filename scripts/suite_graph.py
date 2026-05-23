#!/usr/bin/env python3
"""Build a deterministic repo-local graph of agents, skills, packs, and templates."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def pack_skills() -> dict[str, list[str]]:
    packs: dict[str, list[str]] = {}
    packs_root = ROOT / "packs"
    if not packs_root.exists():
        return packs
    for pack_dir in sorted(path for path in packs_root.iterdir() if path.is_dir()):
        skills_dir = pack_dir / "skills"
        skills = []
        if skills_dir.exists():
            skills = sorted(path.name for path in skills_dir.iterdir() if (path / "SKILL.md").exists())
        packs[pack_dir.name] = skills
    return packs


def templates() -> dict[str, list[str]]:
    return {
        "claude": sorted(path.stem for path in (ROOT / ".claude" / "agents").glob("*.md")),
        "codex": sorted(path.stem for path in (ROOT / ".codex" / "agents").glob("*.toml")),
    }


def build_graph() -> dict:
    dependencies = read_json(ROOT / "scripts" / "dependencies.json")
    loadouts = read_json(ROOT / "scripts" / "agent_loadouts.json")
    return {
        "version": 1,
        "agents": {
            name: {
                "skills": list(data.get("skills", [])),
                "repos": list(dependencies.get("agents", {}).get(name, {}).get("repos", [])),
                "templates": {
                    "claude": name in templates()["claude"],
                    "codex": name in templates()["codex"],
                },
            }
            for name, data in loadouts.items()
        },
        "repos": dependencies.get("repos", {}),
        "packs": pack_skills(),
        "templates": templates(),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--format", choices=["json"], default="json")
    args = parser.parse_args()
    graph = build_graph()
    if args.format == "json":
        print(json.dumps(graph, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
