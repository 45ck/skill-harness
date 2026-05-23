# Demo Production Media

`skill-harness` now carries an embedded `demo-production-skills` pack and a narrow `media` developer-artifact profile.

This is not a general video editor. It is an orchestration layer for source-backed demo and QA media where the durable truth stays in `.demo.yaml`, QA flows, QA reports, Markdown/TOON source notes, and artifact manifests.

## Primary Tool Boundaries

- `demo-machine` owns browser demo capture, event timelines, render output, quality reports, and analyzer handoff artifacts.
- `manual-qa-machine` owns QA flow execution, verdicts, screenshots, console/network/page-error/accessibility/performance evidence, and QA reports.
- `video-evaluator` owns reusable video analysis concepts such as storyboard, shot, segment, layout-safety, and review-prompt evidence.
- `skill-harness` owns installable skill guidance, project artifact scaffolding, source/review separation, manifest checks, and conservative handoff rules.

## Media Profile

Use:

```bash
./skill-harness setup-project --dir ../my-project --developer-artifacts-profile media
```

The profile resolves to `dual` and adds:

- `generated/media/`
- `docs/artifacts/templates/demo-artifact.md`
- media output policy in `.skill-harness/project.json`
- `.gitignore` coverage for generated media

Generated media stays out of git by default. Promote MP4s, GIF/WebP previews, poster frames, and frame strips only when the repo intentionally treats them as release assets.

## Skill Pack

The embedded `demo-production-skills` pack includes:

- `demo-story-packager`
- `demo-social-cut`
- `demo-slideshow-edit`
- `demo-review-surface`
- `qa-to-demo`
- `demo-release-packager`

These skills are wired into UX, QA automation, delivery, research, and workflow loadouts where demo or evidence packaging naturally appears.

## Evidence Rules

- Do not approve a demo from screenshots or polished video alone.
- Preserve upstream QA verdicts. `fail` and `inconclusive` inputs can create repro or draft packages, not approved product demos.
- Missing analyzer evidence downgrades status to `needs-evidence` or `inconclusive`.
- Record source, run directory, output media, evidence links, owner, status, and source hash in `docs/artifacts/artifacts.manifest.json` when available.
- Exclude raw traces, HAR/network dumps, console logs, page errors, secrets, and customer data from release bundles unless explicitly redacted and approved.

## Research Notes

- Remotion is useful when a project wants React/TypeScript-defined video compositions and programmatic MP4 rendering.
- FFmpeg remains the lowest-level mechanism for concat/filter pipelines, fades, frame strips, and format conversion.
- Playwright traces are useful review/debug evidence because they preserve action timelines, DOM snapshots, screenshots, and related browser state.
- Mermaid C4 and model diagrams remain source-backed review artifacts; media profile projects should still pre-render diagrams instead of loading browser runtimes.

Sources reviewed:

- Remotion docs: https://www.remotiondocs.com/
- FFmpeg filters docs: https://ffmpeg.org/ffmpeg-filters.html
- Playwright Trace Viewer docs: https://playwright.dev/docs/trace-viewer-intro
- Mermaid C4 docs: https://mermaid.js.org/syntax/c4.html
- 45ck video-evaluator: https://github.com/45ck/video-evaluator
