from pathlib import Path
import re
import sys


def escape_double_quotes(text: str) -> str:
    return text.replace("\\", "\\\\").replace('"', '\\"')


def yaml_quote(text: str) -> str:
    return f'"{escape_double_quotes(text)}"'


def find_section_line(body: str, heading: str) -> str | None:
    pattern = rf"(?im)^##\s+{re.escape(heading)}\s*$"
    match = re.search(pattern, body)
    if not match:
        return None
    rest = body[match.end():].splitlines()
    for line in rest:
        stripped = line.strip()
        if not stripped:
            continue
        if stripped.startswith("## "):
            break
        if stripped.startswith("- "):
            continue
        return stripped
    return None


def derive_description(skill_name: str, body: str, frontmatter: str | None) -> str:
    if frontmatter:
        purpose_match = re.search(r'(?im)^purpose:\s*"?(.+?)"?\s*$', frontmatter)
        if purpose_match:
            return purpose_match.group(1).strip().rstrip(".")
    for heading in ["Purpose", "Use when", "Trigger this skill when"]:
        line = find_section_line(body, heading)
        if line:
            return line.rstrip(".")
    return f"Skill for {skill_name.replace('-', ' ')}"


def normalize_file(path: Path) -> bool:
    text = path.read_text(encoding="utf-8")
    skill_name = path.parent.name
    changed = False

    if text.startswith("---\n"):
        end = text.find("\n---\n", 4)
        if end != -1:
            frontmatter = text[4:end]
            body = text[end + 5 :]
            desired_name_line = f"name: {yaml_quote(skill_name)}"
            if re.search(r"(?im)^name:\s*", frontmatter):
                updated_frontmatter = re.sub(
                    r"(?im)^name:\s*.*$",
                    desired_name_line,
                    frontmatter,
                    count=1,
                )
                if updated_frontmatter != frontmatter:
                    frontmatter = updated_frontmatter
                    changed = True
            if not re.search(r"(?im)^description:\s*", frontmatter):
                description = derive_description(skill_name, body, frontmatter)
                frontmatter = frontmatter.rstrip() + f"\ndescription: {yaml_quote(description)}\n"
                changed = True
            else:
                description = derive_description(skill_name, body, frontmatter)
                updated_frontmatter = re.sub(
                    r"(?im)^description:\s*.*$",
                    f"description: {yaml_quote(description)}",
                    frontmatter,
                    count=1,
                )
                if updated_frontmatter != frontmatter:
                    frontmatter = updated_frontmatter
                    changed = True
            text = f"---\n{frontmatter.rstrip()}\n---\n{body.lstrip()}"
        else:
            frontmatter = None
            body = text
    else:
        frontmatter = None
        body = text

    if frontmatter is None:
        description = derive_description(skill_name, body, None)
        normalized = (
            f"---\nname: {yaml_quote(skill_name)}\ndescription: {yaml_quote(description)}\n---\n\n"
            + text.lstrip()
        )
        text = normalized
        changed = True

    if changed:
        path.write_text(text, encoding="utf-8", newline="\n")
    return changed


def normalize_tree(root: Path) -> tuple[int, int]:
    total = 0
    changed = 0
    if not root.exists():
        return total, changed
    for skill_md in root.glob("*/SKILL.md"):
        total += 1
        if normalize_file(skill_md):
            changed += 1
    return total, changed


def main(argv: list[str]) -> int:
    roots = [Path(arg) for arg in argv[1:]]
    if not roots:
        roots = [Path.home() / ".agents" / "skills", Path.home() / ".claude" / "skills"]
    grand_total = 0
    grand_changed = 0
    for root in roots:
        total, changed = normalize_tree(root)
        grand_total += total
        grand_changed += changed
        print(f"{root}: normalized {changed} of {total} skills")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
