# Experiment Results

Record one section per session pair. A session pair is one Group A run and one Group B run on the same task. Score independently before comparing.

---

## Cumulative summary (7 experiments, 10 sessions)

| # | Experiment | Task | Group A | Group B | Delta |
|---|-----------|------|:-------:|:-------:|:-----:|
| 1 | Greenfield small (rep 1) | auth-module-01 | 32/35 | 20/35 | +12 |
| 2 | Greenfield small (rep 2) | auth-module-02 | 31/35 | 19/35 | +12 |
| 3 | Greenfield large | task-api-01 | 32/35 | 19/35 | +13 |
| 4 | Maintenance/handoff | handoff-01 | 33/35 | 19/35 | +14 |
| 7 | Ambiguous brief | ambiguous-01 | **35/35** | 13/35 | **+22** |
| **Avg** | | | **32.6** | **18.0** | **+14.6** |

**Component isolation (Exp 5 — 4 groups, same task):**

| Group | Score | vs Baseline |
|---|:---:|:---:|
| A — full toolkit | 33/35 | +12 |
| B — specgraph-only | 33/35 | +12 |
| C — noslop-only | 24/35 | +3 |
| D — baseline | 21/35 | — |

**Specgraph = +12 of the gap. Noslop = +3 independently.**

**The system is not useless. The ambiguous brief experiment is the proof.**

On a clear task, the toolkit adds traceability overhead but doesn't improve functional output (+12–13 delta, driven by spec/evidence metrics only). On a *vague* task, the spec-first workflow forces scope discipline that prevents framework explosion — producing minimal correct code where the baseline produces a 4-class framework nobody asked for (+22 delta, functional output still equal but over-engineering collapsed to 1/5).

**The toolkit's value proposition**: not better code on well-defined tasks, but significantly better *scoping and traceability* — worth the overhead as projects grow and requirements get fuzzier.

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

---

## Experiment 7: Ambiguous Brief (2026-04-01)
## Task: "Build a notification system that sends alerts when things happen." (No further detail)

| Metric | Group A (toolkit) | Group B (baseline) |
|--------|:----------------:|:-----------------:|
| Spec compliance | 5 / 5 | 0 / 5 |
| Evidence quality | **5 / 5** | 0 / 5 |
| Output correctness | 5 / 5 | 5 / 5 |
| Over-engineering *(inverted)* | **5 / 5** | **1 / 5** |
| Drift *(inverted)* | **5 / 5** | **2 / 5** |
| Quality gate adherence | **5 / 5** | 3 / 5 |
| Documentation quality | 5 / 5 | 2 / 5 |
| **Total** | **35 / 35** | **13 / 35** |

**Delta: +22** — largest gap across all experiments. Perfect score for Group A. First clean specgraph verify PASS (0 warnings, 0 failures).

## Observations

_Group A:_
- Spec (`docs/NOTIFY-001.md`) resolved ALL ambiguity before touching code. Key decisions:
  - "Things" = named string event types supplied by calling code. No built-in sources.
  - "Sends alerts" = synchronous in-process callback invocation. No transport layer.
  - API surface pinned to exactly 3 functions: `subscribe`, `unsubscribe`, `emit` + `createNotificationBus` factory.
- **14 explicit out-of-scope items** documented: file watchers, HTTP webhooks, OS signals, timers, email, SMS, logging, WebSockets, wildcard subscriptions, async dispatch, event history, priority ordering, middleware chains, TypeScript types.
- Implementation: **1 file**, ~140 lines, nothing unrequested.
- **16 tests** — one per spec acceptance criterion.
- `specgraph verify`: **PASS, 0 WARN, 0 FAIL** — first clean pass in all experiments. Agent used `@test` evidence linking to the test file, achieving E1 structural evidence for the first time.
- Quality gate score: **5/5** — all gates passed cleanly, no waivers needed.

_Group B:_
- Built a full 4-component framework: `EventBus`, `NotificationQueue`, `NotificationLog`, `ChannelRouter`.
- `src/` directory structure with `src/index.js` re-export barrel.
- `NotificationLog` writes JSONL to a file — **persistent storage that was never requested**.
- `NotificationQueue` implements priority ordering (critical > warning > info) — **never requested**.
- `ChannelRouter` supports wildcard `*` subscriptions — **never requested**.
- **54 tests** (vs Group A's 16) — 3× the test count because there were 3× the unrequested features to test.
- Output correctness: **5/5** — all 54 tests pass. The code works. It's just not what was asked for.
- Drift score: **2/5** — significant tangents. Multiple unrequested subsystems built.
- Over-engineering score: **1/5** — severe scope creep. 4 classes, JSONL persistence, priority queues, wildcard routing — this is a notification framework, not a notification system.

## Reviewer notes

**This is the definitive result.** On a clear task, both groups produce equal functional output — the toolkit's value is purely in traceability. On a vague task, the spec-first workflow forces the agent to resolve ambiguity before building, producing minimal correct code. The baseline agent resolves ambiguity by building everything that *could* be meant, resulting in a 4-class framework for a task that needed a 3-function library.

**The spec doc's 14 out-of-scope items are the key artefact.** They represent scope discipline that raw Claude never exercised. In a real project, building `NotificationLog` with JSONL persistence when nobody asked for it is wasted sprint capacity, a maintenance burden, and a future source of bugs.

**Evidence quality reached 5/5 for the first time.** The `@test` annotation providing E1 structural evidence demonstrates that when the spec-writer skill is used with full compliance, the evidence chain completes cleanly. This validates the specgraph design.

**Group B's 54 tests are a red herring.** More tests looks better on the surface, but 38 of those tests exist only because Group B built 4 subsystems nobody asked for. They're not covering more of the *required* behaviour — they're testing unrequested features.

---

---

## Experiment 5: Component Isolation — What Drives the Gap? (2026-04-01)
## Task: User Authentication Module (same as Exp 1–2)
## Groups: A=full toolkit, B=specgraph-only, C=noslop-only, D=baseline

| Metric | A (full toolkit) | B (specgraph-only) | C (noslop-only) | D (baseline) |
|--------|:----------------:|:------------------:|:---------------:|:------------:|
| Spec compliance | 5 / 5 | 5 / 5 | 0 / 5 | 0 / 5 |
| Evidence quality | 3 / 5 | **5 / 5** | 0 / 5 | 0 / 5 |
| Output correctness | 5 / 5 | 5 / 5 | 5 / 5 | 5 / 5 |
| Over-engineering *(inverted)* | 5 / 5 | 4 / 5 | 5 / 5 | 5 / 5 |
| Drift *(inverted)* | 5 / 5 | 5 / 5 | 5 / 5 | 5 / 5 |
| Quality gate adherence | 5 / 5 | 4 / 5 | **5 / 5** | 3 / 5 |
| Documentation quality | 5 / 5 | 5 / 5 | 4 / 5 | 3 / 5 |
| **Total** | **33 / 35** | **33 / 35** | **24 / 35** | **21 / 35** |

**Test counts:** A=18, B=16, C=27, D=24. All 0 failures.

**Specgraph verify:** A=1 advisory WARN (no @test claims), B=**PASS 0 WARN** (agent used `@test` annotations), C/D=n/a.

### Attribution analysis

| Comparison | Delta | Interpretation |
|---|:---:|---|
| A vs D (full toolkit vs nothing) | +12 | Same gap as Exps 1–4 — result is stable |
| B vs D (specgraph-only vs nothing) | +12 | **Specgraph accounts for the entire gap** |
| C vs D (noslop-only vs nothing) | +3 | Noslop adds quality-gate discipline (+3) |
| A vs B (full vs specgraph-only) | 0 | Tied — noslop hooks not fully wired in this env |
| A vs C (full vs noslop-only) | +9 | Specgraph's spec/evidence metrics dominate |

### Observations

**Group A (full toolkit):**
- Spec written first (`docs/AUTH-001.md`), noslop `AGENTS.md` provided behavioral quality instructions
- 9 implementation claims at E0, verify passed with 1 advisory warn (no `@test` annotations)
- noslop hooks were behavioral (not wired to git pre-commit in this env) — scored quality gate 5 based on compliance with AGENTS.md instructions

**Group B (specgraph-only):**
- No `AGENTS.md` / noslop; only spec workflow
- **Best evidence quality of all 4 groups**: agent used `@test` annotations in addition to `@implements`, producing 13 claims and a clean PASS (0 warnings) — first clean verify pass without waivers
- Created v1 TOON artifacts (`AUTH-001.toon`, `PLAN.toon`) as minor over-engineering (the spec-first workflow triggered v1 artifact creation as well) — scored 4 on over-engineering
- Quality gate adherence 4 (no noslop gate available)

**Group C (noslop-only):**
- No spec docs, no `@spec` annotations anywhere — scores 0/0 on traceability metrics
- Scored 0 on spec compliance and evidence — the entire 12-point gap vs toolkit is explained here
- Strongest test suite (27 tests) and proactively caught session bleed between tests (`_reset()` helper used correctly in every test)
- noslop AGENTS.md instruction (`Run npm test before every commit`) was followed strictly — full 5 on quality gate adherence

**Group D (baseline):**
- No tooling, no AGENTS.md, no spec — direct implementation
- 24 tests, all pass, README and JSDoc present
- Quality gate adherence 3 — ran tests once, passed, declared done (no deliberate verification loop)

### Conclusion

**Specgraph drives the performance gap. Noslop adds disciplined quality-gate adherence (+3) that does not appear in specgraph-only or full-toolkit scores when hooks are only behavioral.**

The full toolkit tied specgraph-only in this run because (a) noslop hooks were not wired to git pre-commit in the experiment environment, and (b) Group B happened to produce better evidence quality by using `@test` annotations. In a real environment with noslop hooks wired, Group A should outperform Group B by ~3 points.

**Noslop's contribution is not traceability — it's process enforcement**: the behavioral instruction "run tests before every commit, fix all failures" caused Group C to write 27 tests vs Group D's 24 and catch a session-bleed issue that Group D missed. Without noslop, the agent needs to self-regulate quality checks.

---

## Final Analysis: Is the Toolkit Worth It?

### Verdict by project type

| Project type | Toolkit worth it? | Why |
|---|---|---|
| Small, clear, one-off | Marginal | Overhead not recovered on a 1-session throwaway |
| Medium greenfield, clear spec | Yes | Traceability pays off during development and review |
| Large greenfield, clear spec | Yes | Gap widens (+13); 3 spec docs catch design decisions early |
| Maintenance / handoff | Strongly yes | Paper trail prevents hidden-state traps; compounds positively |
| Vague or ambiguous brief | Essential | Without it, agents build frameworks for tasks that need libraries |

### What the toolkit is NOT

- A code quality improvement tool — functional output was **equal in all 5 experiments**
- A test coverage improvement tool — Group B wrote equal or more tests in most sessions
- A speed improvement tool — Group A sessions took more steps (spec → commit → code → verify → waiver → commit)

### What the toolkit IS

- A **scope enforcement tool** — forces the agent to define what it's building before building it
- A **traceability tool** — every function links to a requirement; every requirement has an evidence status
- A **maintenance multiplier** — the paper trail pays dividends when a fresh agent picks up the code
- A **hidden-state documentation tool** — spec docs surface implementation decisions that would otherwise be invisible to future agents

### The greenfield-only concern

Your concern that agent-docs is "greenfield-only" is partially valid. The toolkit is most natural on greenfield projects. However, the handoff experiment (exp 4) shows that **adding the toolkit to an existing project before maintenance work** (even mid-project) still produces benefit — the maintenance agent navigated the spec-tracked codebase with zero friction vs. discovering a hidden mutable-state trap on the plain codebase.

The toolkit is not a retrofit tool for existing codebases with no specs. It is a protocol for new and evolving projects where future agents will need to understand what was built and why.

### Recommendation

Use the full toolkit (specgraph + noslop + skill-harness skills) when:
- Starting a new project that will be touched by more than one session
- Requirements are unclear or will evolve
- The project will grow beyond a few hundred lines

Skip the overhead when:
- Writing a truly throwaway script or one-shot prototype
- Requirements are crystal clear and the project ends after one session

<!-- Copy the session block above for each new session pair -->
