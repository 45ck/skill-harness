# Skill Harness

Use the narrowest specialist agent that can own the work end to end. Treat this repo as the suite entrypoint and project setup repo for the 45ck stack, including embedded packs under `packs/`.

## UML-First Developer Artifacts

- Auto-detect model impact for every engineering change. When code, API, workflow, dependency, deployment, UI structure, or agent behavior changes, update the relevant model source or record why no model change is needed.
- Fresh developer-artifact setups use `--modeling-mode auto`, which resolves to `uml-first` for new repos and preserves existing repos. Use `--modeling-mode off|baseline|uml-first` or `--skip-modeling` only when that is intentional.
- Keep canonical UML/UWE/C4/evidence model sources in repo-relative text files. Prefer `docs/artifacts/source/models/` when no better domain docs path exists.
- Keep `docs/artifacts/source/models/model-inventory.md` and `docs/artifacts/artifacts.manifest.json` current with model ids, owners, methods, source paths, evidence, and generated review surfaces.
- Human model artifacts belong in static HTML under `generated/review/models/`. Generate them with `node scripts/generate-model-review.mjs`; validate with `node scripts/check-model-artifact-policy.mjs` and `node scripts/check-artifact-html-policy.mjs`.
- Treat generated HTML, SVG, PNG, screenshots, and comparison pages as review surfaces only. Canonical truth stays in source artifacts and model diffs.

## Session Rules

- Use Beads for task tracking in this repo.
- Run the relevant Go tests and generated artifact checks after code changes.
- Push completed work to the remote before ending the session.
