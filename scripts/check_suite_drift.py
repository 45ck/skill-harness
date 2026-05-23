#!/usr/bin/env python3
"""Check that generated suite docs match the canonical graph inputs."""

from __future__ import annotations

import argparse
import difflib

from render_suite_docs import OUT_PATH, render_markdown
from suite_graph import ROOT


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="fail when docs are stale")
    args = parser.parse_args()

    expected = render_markdown()
    actual = OUT_PATH.read_text(encoding="utf-8") if OUT_PATH.exists() else ""
    if actual == expected:
        print("Suite docs drift check passed")
        return 0

    if args.check:
        print("Suite docs drift check failed:")
        diff = difflib.unified_diff(
            actual.splitlines(),
            expected.splitlines(),
            fromfile=OUT_PATH.relative_to(ROOT).as_posix(),
            tofile="generated",
            lineterm="",
        )
        for line in diff:
            print(line)
        return 1

    OUT_PATH.write_text(expected, encoding="utf-8")
    print(f"Rendered {OUT_PATH.relative_to(ROOT).as_posix()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
