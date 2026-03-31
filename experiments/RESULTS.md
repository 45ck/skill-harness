# Experiment Results

Record one section per session pair. A session pair is one Group A run and one Group B run on the same task. Score independently before comparing.

---

## Cumulative summary (6 sessions)

| # | Experiment | Task | Group A | Group B | Delta |
|---|-----------|------|:-------:|:-------:|:-----:|
| 1 | Greenfield small (rep 1) | auth-module-01 | 32/35 | 20/35 | +12 |
| 2 | Greenfield small (rep 2) | auth-module-02 | 31/35 | 19/35 | +12 |
| 3 | Greenfield large | task-api-01 | 32/35 | 19/35 | +13 |
| 4 | Maintenance/handoff | handoff-01 | 33/35 | 19/35 | +14 |
| 7 | Ambiguous brief | ambiguous-01 | TBD | TBD | TBD |
| **Avg (1-4)** | | | **32** | **19.25** | **+12.75** |

**Signal is consistent and strengthening.** The gap is stable at +12–14 across all task types. Largest gap on the maintenance/handoff experiment — the toolkit's paper trail had measurable utility for a fresh agent. Exp 7 (ambiguous brief) pending.

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

## Session: 2026-04-01 (run 2)
## Task: User Authentication Module (`experiments/session-task.md`)

| Metric | Group A (toolkit) | Group B (baseline) |
|--------|:----------------:|:-----------------:|
| Spec compliance | 5 / 5 | 0 / 5 |
| Evidence quality | 3 / 5 | 0 / 5 |
| Output correctness | 5 / 5 | 5 / 5 |
| Over-engineering *(inverted)* | 4 / 5 | 4 / 5 |
| Drift *(inverted)* | 5 / 5 | 5 / 5 |
| Quality gate adherence | 4 / 5 | 3 / 5 |
| Documentation quality | 5 / 5 | 2 / 5 |
| **Total** | **31 / 35** | **19 / 35** |

**Delta: +12 in favour of Group A.** Consistent with session 1.

## Observations

_Group A notable behaviors:_

- Spec written first (`docs/AUTH-001.md`) with two evidence dimensions: `implementation: E0` and `test_coverage: E0`.
- Spec committed before any code. 5 `@implements` annotations; specgraph registered **9 implementation claims** at E0.
- Added `activeSessionCount()` utility (not required — minor over-engineering, -1 on metric 4).
- `npx specgraph verify`: `1 WARN` (advisory — same Beads-unavailable condition as session 1). Waiver recorded.
- `npm test`: **20/20 pass** across 5 suites.
- Two-commit workflow: spec-first, then implementation + tests + README + waiver.

_Group B notable behaviors:_

- No spec, no docs, no annotations — direct implementation from task prompt.
- **22/22 tests pass**, no failures encountered (no bug to catch and fix this run).
- `createCredential()` factory function — minor abstraction, same score as session 1 on over-engineering.
- Lazy session sweep on every `login` call (unrequested but reasonable).
- README present with usage examples. Strong JSDoc.
- No deliberate final-state verification step observed (-1 vs session 1 on quality gate adherence).

## Reviewer notes

**Results replicate cleanly.** Session 2 delta (+12) matches session 1 (+12) exactly. Group A scores 31–32; Group B scores 19–20. The spread is stable.

**Spec compliance and evidence quality account for the entire gap.** Group A scores 5+3=8 on these two metrics each session; Group B scores 0+0=0. All other metrics are within 1 point of each other across both sessions. The toolkit adds traceability; it adds nothing to functional output.

**Group A over-engineering crept in.** Session 1 Group A scored 5 (nothing unrequested); session 2 scored 4 (`activeSessionCount` added). Small variance — different random seed, same model. Does not indicate a pattern.

**Group B quality gate adherence dropped 4→3.** Session 1 Group B caught and fixed a real boundary condition bug, demonstrating active quality discipline. Session 2 had no failures to fix — the agent just ran tests once, they passed, and it declared completion. Slightly less evidence of deliberate quality checking.

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

---

## Experiment 3: Greenfield Large Project (2026-04-01)
## Task: Task Management REST API (3 modules, 6 HTTP routes)

| Metric | Group A (toolkit) | Group B (baseline) |
|--------|:----------------:|:-----------------:|
| Spec compliance | 5 / 5 | 0 / 5 |
| Evidence quality | 3 / 5 | 0 / 5 |
| Output correctness | 5 / 5 | 5 / 5 |
| Over-engineering *(inverted)* | 5 / 5 | 4 / 5 |
| Drift *(inverted)* | 5 / 5 | 5 / 5 |
| Quality gate adherence | 4 / 5 | 3 / 5 |
| Documentation quality | 5 / 5 | 2 / 5 |
| **Total** | **32 / 35** | **19 / 35** |

**Delta: +13** (highest gap on a greenfield task; gap held and widened vs small tasks).

## Observations

_Group A:_
- 3 specs written and committed before a single line of implementation: `AUTH-001.md`, `TASKS-001.md`, `API-001.md` — one per component.
- 16 implementation claims registered at E0 across all 3 modules.
- `specgraph verify`: `3 WARN` (advisory, all waived — one per spec, same Beads-unavailable condition). 0 FAIL. Waivers recorded in each spec's frontmatter.
- `npm test`: **28/28 pass** covering all 6 routes, auth, scoping, 401/404/400 cases.
- README with full curl examples for every endpoint.
- Two-commit workflow: specs first, then all implementation.

_Group B:_
- Direct implementation from task prompt. No specs.
- `getUsernameFromToken` extra export (minor), `pruneExpiredSessions` lazy call on every login/check.
- **27/27 pass** — one fewer test than Group A despite no spec overhead.
- README present. Strong JSDoc. No spec docs.
- Ran tests once, clean pass, declared done (no deliberate final verification).

## Reviewer notes

**Gap widened slightly on larger scope (+13 vs +12).** The additional modules and spec docs mean Group A's structural advantage compounds. Three spec docs with clear requirements lists means every design decision is recorded; Group B's three-module implementation has no documented rationale anywhere.

**Group B test count was lower (27 vs 28) on the larger task** — opposite of small tasks where Group B wrote more tests. Hypothesis: at higher complexity, the spec-first workflow gives Group A a structured test target list (one test per acceptance criterion), while Group B is working from memory of what it built.

**Quality gate discipline diverged more (+4 vs +3).** Group A ran specgraph verify + npm test with documented waivers. Group B ran tests once and stopped.

---

## Experiment 4: Maintenance / Handoff Test (2026-04-01)
## Task: Add password reset to existing auth module (fresh agent, no prior context)

| Metric | Group A (toolkit) | Group B (baseline) |
|--------|:----------------:|:-----------------:|
| Spec compliance | 5 / 5 | 0 / 5 |
| Evidence quality | 4 / 5 | 0 / 5 |
| Output correctness | 5 / 5 | 5 / 5 |
| Over-engineering *(inverted)* | 5 / 5 | 3 / 5 |
| Drift *(inverted)* | 5 / 5 | 5 / 5 |
| Quality gate adherence | 4 / 5 | 4 / 5 |
| Documentation quality | 5 / 5 | 2 / 5 |
| **Total** | **33 / 35** | **19 / 35** |

**Delta: +14** — largest gap across all experiments. The paper trail proved its value on first use.

## Observations

_Group A:_
- Agent explicitly reported: *"The spec doc was genuinely useful — it covered requirements, acceptance criteria, out-of-scope items, and assumptions in enough detail that I needed no guesswork about intent."*
- Navigated the codebase via `docs/AUTH-001.md` + existing annotations. No surprises.
- Added requirements 8–12 to spec before adding code. Annotated new functions.
- `specgraph verify`: same pre-existing advisory warn, pre-existing waiver still valid. 0 new FAIL.
- **26/26 pass** (19 original + 7 new tests).
- Evidence quality scored **4** (highest across all experiments) — starting from an established E0 claim chain means the evidence dimension was already partially satisfied.

_Group B:_
- Agent navigated by reading `auth.js` and `auth.test.js` directly — no spec to consult.
- Hit a **hidden shared-mutable-state trap**: `USER_STORE` is module-level. `resetPassword` mutates it. Existing tests did not reset it in `beforeEach` (never needed to before). First `npm test` run failed.
- Had to add `_resetUserStore` helper, extract `DEFAULT_USERS` constant, update `beforeEach` — extra discovery and remediation work not needed by Group A.
- **36/36 pass** (25 original + 11 new tests) — more tests than Group A's 7 new, but 4 of those extra tests were specifically to cover the trap they discovered.
- Over-engineering score: **3** (forced to add `_resetUserStore` + `DEFAULT_USERS` refactor due to undocumented state — not bloat by choice, but undocumented complexity that compelled extra scaffolding).

## Reviewer notes

**The paper trail proved useful on first actual use.** This is the most important result across all experiments. Both agents succeeded, but Group A had zero friction; Group B spent non-trivial effort discovering a hidden implementation detail that a spec document, a comment, or even a README note would have prevented.

**The hidden-state trap is a proxy for real maintenance cost.** In a production codebase with tens of modules, undocumented mutable state is a major source of bugs. Group A's annotations pointed new agents to the spec; Group B's code provided no such guidance. The discovery cost here was a few extra functions and a test rewrite — at larger scale it would be a bug.

**Evidence quality reached 4/5 for the first time** in Group A. The maintenance context means starting from a working evidence chain rather than zero — iterative spec maintenance compounds positively.

<!-- Copy the session block above for each new session pair -->
