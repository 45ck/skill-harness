from __future__ import annotations

import argparse
from dataclasses import asdict
from dataclasses import dataclass
import json
import os
from pathlib import Path
import re
import sys


REPO_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_INTAKE_ROOT = REPO_ROOT.parent / "skill-intake"
DEFAULT_LOCAL_SKILLS_ROOT = Path.home() / ".agents" / "skills"

SKIP_DIR_NAMES = {
    ".git",
    ".hg",
    ".svn",
    ".venv",
    "__pycache__",
    "dist",
    "node_modules",
}
TEXT_SUFFIXES = {
    "",
    ".bat",
    ".cjs",
    ".cmd",
    ".js",
    ".json",
    ".jsx",
    ".md",
    ".mjs",
    ".ps1",
    ".py",
    ".sh",
    ".toml",
    ".ts",
    ".tsx",
    ".txt",
    ".yaml",
    ".yml",
}
EXECUTABLE_SUFFIXES = {".bat", ".cmd", ".cjs", ".js", ".mjs", ".ps1", ".py", ".sh", ".ts"}
PLUGIN_DIR_NAMES = {
    ".claude-plugin",
    ".codex-plugin",
    ".cursor-plugin",
    ".gemini-plugin",
    ".opencode-plugin",
}
LOCK_OR_MANIFEST_NAMES = {
    "skills_index.json",
    "skills-lock.json",
    "skill-lock.json",
    "package-lock.json",
    "pnpm-lock.yaml",
    "yarn.lock",
    "uv.lock",
}
LICENSE_NAMES = {
    "license",
    "license.md",
    "license.txt",
    "license-content",
    "license-content.md",
    "license-content.txt",
    "third_party_notices.md",
    "thirdpartynoticetext.txt",
}
BLOCKED_PATTERNS = {
    "approval-bypass": re.compile(r"(?i)dangerously-bypass|bypass[-_\s]+approvals?|disable[-_\s]+sandbox"),
    "remote-shell-exec": re.compile(r"(?i)(curl|wget)\s+[^|]+[|]\s*(bash|sh)|irm\s+[^|]+[|]\s*iex"),
    "secret-exfiltration": re.compile(r"(?i)(print|echo|send|upload|post).{0,80}(token|secret|api[_-]?key|password)"),
}
REVIEW_PATTERNS = {
    "token-reference": re.compile(r"(?i)(github_personal_access_token|api[_-]?key|bearer[_-]?token|secret|password)"),
    "mcp-or-tool-config": re.compile(r"(?i)(\.mcp\.json|mcp_servers?|modelcontextprotocol|tool[_-]?approval)"),
    "install-script": re.compile(r"(?i)(install\.sh|install\.ps1|setup\.sh|setup\.ps1|postinstall)"),
}


@dataclass
class RepoSummary:
    name: str
    path: Path
    skill_names: list[str]
    agent_file_count: int
    cursor_rule_count: int
    host_instruction_count: int
    copilot_instruction_count: int
    plugin_manifest_count: int
    plugin_surface_count: int
    openai_metadata_count: int
    mcp_config_count: int
    lock_metadata_count: int
    license_file_count: int
    executable_file_count: int
    blocked_flags: list[str]
    review_flags: list[str]
    trust_recommendation: str
    install_surface: str
    overlap_names: list[str]
    unique_names: list[str]


def iter_repo_files(repo_root: Path):
    for dirpath, dirnames, filenames in os.walk(repo_root):
        dirnames[:] = [name for name in dirnames if name not in SKIP_DIR_NAMES]
        current = Path(dirpath)
        for filename in filenames:
            yield current / filename


def relative_posix(repo_root: Path, path: Path) -> str:
    return path.relative_to(repo_root).as_posix()


def is_text_candidate(path: Path) -> bool:
    return path.suffix.lower() in TEXT_SUFFIXES


def read_text_sample(path: Path, max_bytes: int = 200_000) -> str:
    try:
        data = path.read_bytes()[:max_bytes]
    except OSError:
        return ""
    return data.decode("utf-8", errors="ignore")


def find_skill_names(repo_root: Path) -> list[str]:
    return sorted(
        {
            skill_md.parent.name
            for skill_md in iter_repo_files(repo_root)
            if skill_md.name == "SKILL.md"
            if skill_md.is_file()
        }
    )


def count_agent_files(repo_root: Path) -> int:
    count = 0
    for path in iter_repo_files(repo_root):
        if path.suffix.lower() not in {".md", ".toml"}:
            continue
        normalized = relative_posix(repo_root, path).lower()
        if (
            "/.claude/agents/" in normalized
            or "/.codex/agents/" in normalized
            or normalized.endswith("/agents.md")
            or "/agents/" in normalized
        ):
            count += 1
    return count


def count_cursor_rules(repo_root: Path) -> int:
    count = 0
    for path in iter_repo_files(repo_root):
        normalized = relative_posix(repo_root, path).lower()
        if normalized.startswith(".cursor/rules/") and path.suffix.lower() == ".mdc":
            count += 1
    return count


def count_host_instruction_files(repo_root: Path) -> int:
    names = {
        "agents.md",
        "agent_instructions.md",
        "claude.md",
        "gemini.md",
        "llms.txt",
    }
    count = 0
    for path in iter_repo_files(repo_root):
        normalized = relative_posix(repo_root, path).lower()
        if path.name.lower() in names:
            count += 1
        elif normalized == ".github/copilot-instructions.md":
            count += 1
    return count


def count_copilot_instruction_files(repo_root: Path) -> int:
    return sum(
        1
        for path in iter_repo_files(repo_root)
        if relative_posix(repo_root, path).lower() == ".github/copilot-instructions.md"
    )


def count_plugin_manifests(repo_root: Path) -> int:
    return sum(1 for path in iter_repo_files(repo_root) if path.name == "plugin.json")


def count_plugin_surfaces(repo_root: Path) -> int:
    count = 0
    for dirpath, dirnames, _filenames in os.walk(repo_root):
        kept = []
        for dirname in dirnames:
            if dirname in SKIP_DIR_NAMES:
                continue
            if dirname in PLUGIN_DIR_NAMES:
                count += 1
            kept.append(dirname)
        dirnames[:] = kept
    return count


def count_named_files(repo_root: Path, names: set[str]) -> int:
    normalized = {name.lower() for name in names}
    return sum(1 for path in iter_repo_files(repo_root) if path.name.lower() in normalized)


def count_path_suffix(repo_root: Path, suffix: str) -> int:
    return sum(1 for path in iter_repo_files(repo_root) if relative_posix(repo_root, path).endswith(suffix))


def count_executable_files(repo_root: Path) -> int:
    return sum(1 for path in iter_repo_files(repo_root) if path.suffix.lower() in EXECUTABLE_SUFFIXES)


def scan_risk_flags(repo_root: Path) -> tuple[list[str], list[str]]:
    blocked: set[str] = set()
    review: set[str] = set()
    for path in iter_repo_files(repo_root):
        rel = relative_posix(repo_root, path).lower()
        if path.suffix.lower() in EXECUTABLE_SUFFIXES:
            review.add("executable-helper")
        if path.name.lower() in LICENSE_NAMES:
            continue
        if not is_text_candidate(path):
            continue
        text = rel + "\n" + read_text_sample(path)
        for name, pattern in BLOCKED_PATTERNS.items():
            if pattern.search(text):
                blocked.add(name)
        for name, pattern in REVIEW_PATTERNS.items():
            if pattern.search(text):
                review.add(name)
    return sorted(blocked), sorted(review)


def trust_recommendation(summary: RepoSummary) -> str:
    if summary.blocked_flags:
        return "quarantine"
    if summary.license_file_count == 0:
        return "manual-review-missing-license"
    if summary.executable_file_count > 0 or summary.mcp_config_count > 0:
        return "manual-review-executable-or-mcp"
    if summary.plugin_surface_count > 0 or summary.plugin_manifest_count > 0:
        return "manual-review-plugin"
    return "intake-candidate"


def readme_text(repo_root: Path) -> str:
    for name in ("README.md", "readme.md"):
        candidate = repo_root / name
        if candidate.exists():
            return candidate.read_text(encoding="utf-8", errors="ignore")
    return ""


def detect_install_surface(repo_root: Path, readme: str) -> str:
    normalized = readme.lower()
    surfaces: list[str] = []

    if "plugin marketplace" in normalized or "/plugin install" in normalized:
        surfaces.append("plugin")
    if count_plugin_surfaces(repo_root) or count_plugin_manifests(repo_root):
        surfaces.append("plugin")
    if "npx skills" in normalized or "agent-skills-cli" in normalized:
        surfaces.append("cli")
    if "skill-installer" in normalized:
        surfaces.append("catalog")
    if "install.sh" in normalized or (repo_root / "install.sh").exists():
        surfaces.append("script")
    if (
        ".codex/skills" in normalized
        or ".claude/skills" in normalized
        or "$codex_home/skills" in normalized
        or "$codeX_home/skills".lower() in normalized
    ):
        surfaces.append("copy")
    if ".cursor/rules" in normalized or count_cursor_rules(repo_root):
        surfaces.append("cursor-rules")
    if "copilot-instructions.md" in normalized or count_copilot_instruction_files(repo_root):
        surfaces.append("copilot-instructions")
    if count_host_instruction_files(repo_root):
        surfaces.append("host-instructions")
    if "curated list" in normalized or "awesome" in normalized:
        surfaces.append("index")

    ordered = []
    for value in surfaces:
        if value not in ordered:
            ordered.append(value)
    return ",".join(ordered) if ordered else "unknown"


def local_skill_names(local_root: Path) -> list[str]:
    if not local_root.exists():
        return []
    return sorted(
        {
            path.name
            for path in local_root.iterdir()
            if path.is_dir() and not path.name.startswith(".")
        }
    )


def summarize_repo(repo_root: Path, local_names: set[str]) -> RepoSummary:
    skills = find_skill_names(repo_root)
    overlap = sorted([name for name in skills if name in local_names])
    unique = sorted([name for name in skills if name not in local_names])
    readme = readme_text(repo_root)
    blocked_flags, review_flags = scan_risk_flags(repo_root)
    summary = RepoSummary(
        name=repo_root.name,
        path=repo_root,
        skill_names=skills,
        agent_file_count=count_agent_files(repo_root),
        cursor_rule_count=count_cursor_rules(repo_root),
        host_instruction_count=count_host_instruction_files(repo_root),
        copilot_instruction_count=count_copilot_instruction_files(repo_root),
        plugin_manifest_count=count_plugin_manifests(repo_root),
        plugin_surface_count=count_plugin_surfaces(repo_root),
        openai_metadata_count=count_path_suffix(repo_root, "/agents/openai.yaml"),
        mcp_config_count=count_named_files(repo_root, {".mcp.json"}),
        lock_metadata_count=count_named_files(repo_root, LOCK_OR_MANIFEST_NAMES),
        license_file_count=count_named_files(repo_root, LICENSE_NAMES),
        executable_file_count=count_executable_files(repo_root),
        blocked_flags=blocked_flags,
        review_flags=review_flags,
        trust_recommendation="",
        install_surface=detect_install_surface(repo_root, readme),
        overlap_names=overlap,
        unique_names=unique,
    )
    summary.trust_recommendation = trust_recommendation(summary)
    return summary


def find_repo_roots(paths: list[str], intake_root: Path) -> list[Path]:
    if paths:
        return [Path(path).resolve() for path in paths]
    if not intake_root.exists():
        return []
    return sorted([path for path in intake_root.iterdir() if path.is_dir()], key=lambda p: p.name.lower())


def markdown_table(summaries: list[RepoSummary]) -> list[str]:
    lines = [
        "| Repo | Skills | Agents | Cursor rules | Host instructions | Plugins | OpenAI metadata | MCP configs | Executables | License files | Blocked flags | Recommendation | Install surface |",
        "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- |",
    ]
    for summary in summaries:
        blocked = ", ".join(summary.blocked_flags) if summary.blocked_flags else "none"
        lines.append(
            f"| `{summary.name}` | {len(summary.skill_names)} | {summary.agent_file_count} | "
            f"{summary.cursor_rule_count} | {summary.host_instruction_count} | "
            f"{summary.plugin_manifest_count + summary.plugin_surface_count} | "
            f"{summary.openai_metadata_count} | {summary.mcp_config_count} | "
            f"{summary.executable_file_count} | {summary.license_file_count} | "
            f"{blocked} | `{summary.trust_recommendation}` | {summary.install_surface} |"
        )
    return lines


def repo_section(summary: RepoSummary, sample_size: int) -> list[str]:
    unique_sample = ", ".join(summary.unique_names[:sample_size]) if summary.unique_names else "none"
    overlap_sample = ", ".join(summary.overlap_names[:sample_size]) if summary.overlap_names else "none"
    return [
        f"### {summary.name}",
        f"- Path: `{summary.path}`",
        f"- Skills: {len(summary.skill_names)}",
        f"- Agent files: {summary.agent_file_count}",
        f"- Cursor rules: {summary.cursor_rule_count}",
        f"- Host instruction files: {summary.host_instruction_count}",
        f"- Copilot instruction files: {summary.copilot_instruction_count}",
        f"- Plugin manifests: {summary.plugin_manifest_count}",
        f"- Plugin surfaces: {summary.plugin_surface_count}",
        f"- OpenAI metadata files: {summary.openai_metadata_count}",
        f"- MCP configs: {summary.mcp_config_count}",
        f"- Lock or index metadata files: {summary.lock_metadata_count}",
        f"- License files: {summary.license_file_count}",
        f"- Executable helpers: {summary.executable_file_count}",
        f"- Blocked flags: {', '.join(summary.blocked_flags) if summary.blocked_flags else 'none'}",
        f"- Review flags: {', '.join(summary.review_flags) if summary.review_flags else 'none'}",
        f"- Trust recommendation: `{summary.trust_recommendation}`",
        f"- Install surface: `{summary.install_surface}`",
        f"- Sample unique skills: {unique_sample}",
        f"- Sample overlapping skills: {overlap_sample}",
        "",
    ]


def format_report(summaries: list[RepoSummary], local_root: Path, sample_size: int) -> str:
    lines = [
        "# External Skill Intake Report",
        "",
        f"- Local comparison root: `{local_root}`",
        f"- Repos scanned: {len(summaries)}",
        "",
        "## Summary",
        "",
    ]
    lines.extend(markdown_table(summaries))
    lines.append("")
    lines.append("## Repo Notes")
    lines.append("")
    for summary in summaries:
        lines.extend(repo_section(summary, sample_size))
    return "\n".join(lines).rstrip() + "\n"


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Scan external skill repos and compare their skill names with the local 45ck-installed catalog."
    )
    parser.add_argument(
        "repos",
        nargs="*",
        help="Repo roots to scan. Defaults to every directory under --intake-root.",
    )
    parser.add_argument(
        "--intake-root",
        default=str(DEFAULT_INTAKE_ROOT),
        help="Fallback directory containing cloned external repos.",
    )
    parser.add_argument(
        "--local-skills-root",
        default=str(DEFAULT_LOCAL_SKILLS_ROOT),
        help="Local skills directory used for overlap checks.",
    )
    parser.add_argument(
        "--sample-size",
        type=int,
        default=12,
        help="How many sample skill names to print per repo.",
    )
    parser.add_argument(
        "--output",
        help="Write the markdown report to this file instead of stdout only.",
    )
    parser.add_argument(
        "--json-output",
        help="Write machine-readable repo summaries to this JSON file.",
    )
    parser.add_argument(
        "--fail-on-blocked",
        action="store_true",
        help="Exit with status 2 when any scanned repo has blocked risk flags.",
    )
    return parser.parse_args(argv[1:])


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    intake_root = Path(args.intake_root).resolve()
    local_root = Path(args.local_skills_root).resolve()

    repo_roots = find_repo_roots(args.repos, intake_root)
    if not repo_roots:
        print("no repos found to scan", file=sys.stderr)
        return 1

    local_names = set(local_skill_names(local_root))
    summaries = [summarize_repo(repo_root, local_names) for repo_root in repo_roots]
    report = format_report(summaries, local_root, sample_size=max(args.sample_size, 1))
    print(report, end="")

    if args.output:
        output_path = Path(args.output).resolve()
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(report, encoding="utf-8", newline="\n")

    if args.json_output:
        json_path = Path(args.json_output).resolve()
        json_path.parent.mkdir(parents=True, exist_ok=True)
        json_ready = []
        for summary in summaries:
            item = asdict(summary)
            item["path"] = str(summary.path)
            json_ready.append(item)
        json_path.write_text(json.dumps(json_ready, indent=2) + "\n", encoding="utf-8", newline="\n")

    if args.fail_on_blocked and any(summary.blocked_flags for summary in summaries):
        return 2

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
