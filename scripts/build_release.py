from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import tarfile
from pathlib import Path
import zipfile


ROOT = Path(__file__).resolve().parents[1]
DIST = ROOT / "dist"

INCLUDE_PATHS = [
    Path(".claude"),
    Path(".codex"),
    Path(".agents"),
    Path("cmd"),
    Path("docs"),
    Path("plugins"),
    Path("scripts"),
    Path("README.md"),
    Path("AGENT_INSTRUCTIONS.md"),
    Path("AGENTS.md"),
    Path("LICENSE"),
    Path("install.sh"),
    Path("install.ps1"),
    Path("uninstall.sh"),
    Path("uninstall.ps1"),
    Path("go.mod"),
    Path("logo.svg"),
    Path("banner.svg"),
    Path("social-preview.svg"),
]

TARGETS = [
    ("windows", "amd64"),
    ("windows", "arm64"),
    ("linux", "amd64"),
    ("linux", "arm64"),
    ("darwin", "amd64"),
    ("darwin", "arm64"),
]


def run(cmd: list[str], cwd: Path | None = None, env: dict[str, str] | None = None) -> None:
    subprocess.run(cmd, cwd=cwd or ROOT, env=env, check=True)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", default="dev", help="Version label for output folders and archives.")
    parser.add_argument(
        "--target",
        action="append",
        default=[],
        help="Optional target in GOOS/GOARCH form, for example windows/amd64.",
    )
    return parser.parse_args()


def selected_targets(raw_targets: list[str]) -> list[tuple[str, str]]:
    if not raw_targets:
        return TARGETS
    parsed: list[tuple[str, str]] = []
    for item in raw_targets:
        goos, goarch = item.split("/", 1)
        parsed.append((goos, goarch))
    return parsed


def copy_tree(src: Path, dst: Path) -> None:
    if src.is_dir():
        shutil.copytree(src, dst, dirs_exist_ok=True)
    else:
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)


def build_binary(bundle_root: Path, goos: str, goarch: str) -> str:
    binary_name = "skill-harness.exe" if goos == "windows" else "skill-harness"
    target_path = bundle_root / binary_name
    env = dict(os.environ, GOOS=goos, GOARCH=goarch, CGO_ENABLED="0")
    run(
        [
            shutil.which("go") or "go",
            "build",
            "-o",
            str(target_path),
            "./cmd/skill-harness",
        ],
        env=env,
    )
    return binary_name


def write_quickstart(bundle_root: Path, binary_name: str, goos: str) -> None:
    command = f".\\{binary_name}" if goos == "windows" else f"./{binary_name}"
    content = (
        "# skill-harness release bundle\n\n"
        "This bundle includes the CLI plus the required repo files.\n\n"
        "Quick start:\n\n"
        f"- Run `{command} install --all`\n"
        f"- Or run `{command} install --interactive`\n"
        f"- Or read `AGENT_INSTRUCTIONS.md`\n"
    )
    (bundle_root / "QUICKSTART.md").write_text(content, encoding="utf-8")


def make_archives(bundle_root: Path, archive_base: Path, goos: str) -> None:
    zip_path = archive_base.parent / f"{archive_base.name}.zip"
    with zipfile.ZipFile(zip_path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for file in bundle_root.rglob("*"):
            if file.is_file():
                zf.write(file, file.relative_to(bundle_root.parent))

    if goos != "windows":
        tar_path = archive_base.parent / f"{archive_base.name}.tar.gz"
        with tarfile.open(tar_path, "w:gz") as tf:
            tf.add(bundle_root, arcname=bundle_root.name)


def main() -> None:
    args = parse_args()
    version_dir = DIST / args.version
    if version_dir.exists():
        shutil.rmtree(version_dir)
    version_dir.mkdir(parents=True, exist_ok=True)

    for goos, goarch in selected_targets(args.target):
        bundle_name = f"skill-harness_{args.version}_{goos}_{goarch}"
        bundle_root = version_dir / bundle_name
        bundle_root.mkdir(parents=True, exist_ok=True)
        for rel in INCLUDE_PATHS:
            copy_tree(ROOT / rel, bundle_root / rel)
        binary_name = build_binary(bundle_root, goos, goarch)
        write_quickstart(bundle_root, binary_name, goos)
        make_archives(bundle_root, version_dir / bundle_name, goos)
        print(f"built {bundle_name}")


if __name__ == "__main__":
    main()
