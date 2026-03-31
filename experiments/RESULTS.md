# Experiment Results

Record one section per session pair. A session pair is one Group A run and one Group B run on the same task. Score independently before comparing.

---

## Session: 2026-04-01
## Task: User Authentication Module (`experiments/session-task.md`)

| Metric | Group A (toolkit) | Group B (baseline) |
|--------|:----------------:|:-----------------:|
| Spec compliance | 5 / 5 | 0 / 5 |
| Evidence quality | 3 / 5 | 0 / 5 |
| Output correctness | 5 / 5 | 5 / 5 |
| Over-engineering *(inverted)* | 5 / 5 | 4 / 5 |
| Drift *(inverted)* | 5 / 5 | 5 / 5 |
| Quality gate adherence | 4 / 5 | 4 / 5 |
| Documentation quality | 5 / 5 | 2 / 5 |
| **Total** | **32 / 35** | **20 / 35** |

**Delta: +12 in favour of Group A.** Exceeds the 7-point meaningful-signal threshold.

## Observations

_Group A notable behaviors (skills invoked, specgraph verify output, noslop gate results):_

- Invoked `spec-writer` skill as **first action** — wrote `docs/AUTH-001.md` before any code.
- Spec contained correct YAML frontmatter (`id`, `title`, `state: in_progress`, `kind: functional`, `required_evidence: implementation: E0`), 7 requirements, acceptance criteria, and an out-of-scope boundary list.
- Spec committed before implementation began.
- Every function in `auth.js` annotated with `@spec AUTH-001`, `@implements`, and `@evidence E0`.
- specgraph registered **9 implementation claims** at E0.
- `npx specgraph verify` output: `1 WARN` (advisory — no `VERIFIED_BY` cross-reference claims; Beads not available in session).
- Waiver written inline in spec frontmatter with justification and `expires: 2026-07-01`.
- `npm test`: **19/19 pass**. No failures to fix.
- noslop pre-commit hooks were NOT active (`noslop install` did not wire `.git/hooks/pre-commit` in this environment — only sample hooks present).
- All files committed in two commits: spec-first commit, then implementation + tests + README + verify-waiver.

_Group B notable behaviors (what the agent did without guidance):_

- No spec, no docs, no annotations — code written directly from task prompt.
- Caught and fixed a real **boundary condition bug**: initial expiry check used `> SESSION_TTL_MS` instead of `>= SESSION_TTL_MS`, meaning a token aged exactly 30 minutes was incorrectly valid. Fixed before declaring completion.
- Produced **25 tests** (vs Group A's 19) with broader edge-case coverage across 4 suites.
- Added `getSessionUser()` (undocumented bonus function, not required) and `buildUserStore()` (minor abstraction for test isolation).
- `npm test`: **25/25 pass**.
- README present with code examples for all three operations.
- Strong inline JSDoc — `@param`, `@returns`, `@throws` on every public function.

## Reviewer notes

**Output correctness was equal (both 5/5).** The toolkit produced no improvement in functional quality. Both implementations are correct, handle edge cases, and use appropriate stdlib APIs. The toolkit's value is entirely in traceability and process — not raw code quality.

**Group B wrote more tests.** Without the overhead of writing specs and running verify, Group B had more cognitive budget available for test-writing. This is a real tradeoff: the toolkit trades raw test volume for a formal requirements record and an evidence chain.

**Quality gate adherence tied (both 4/5).** Group A ran specgraph verify and wrote a waiver; Group B caught a boundary condition via testing. Different mechanisms, same adherence quality. Group A would likely score 5 in an environment where noslop hooks are properly wired.

**Evidence quality limited by environment (Group A: 3/5).** Beads issue tracking was unavailable, preventing E1 evidence. In a project with Beads, this metric would reach 4. The waiver handling was correct and well-reasoned.

**The toolkit enforces a paper trail.** The most concrete difference: after Group A's session, you can answer "which requirement does this function implement and what evidence exists?" After Group B's session, you cannot. For production systems, auditability, or onboarding, Group A's output is substantially more useful.

**Meaningful signal, but single session.** +12 exceeds the threshold but should be replicated across 3–5 sessions before drawing firm conclusions. Variability between runs (same model, different random seeds) may be significant.

---

## Scoring reference

| Metric | 0 | 5 |
|--------|---|---|
| Spec compliance | No specs written | Full coverage; all requirements traceable |
| Evidence quality | No evidence | Full chain: annotations + tests + clean verify |
| Output correctness | Syntax errors / doesn't run | All acceptance criteria met; tests pass |
| Over-engineering | Severe bloat | Minimal, precise — exactly what was asked |
| Drift | Task abandoned | Laser-focused; every action served the task |
| Quality gate adherence | No checks performed | All gates passed before completion |
| Documentation quality | No documentation | README + specs + inline docs all accurate |

Full rubric: `experiments/methodology.md`

---

<!-- Copy the session block above for each new session pair -->
