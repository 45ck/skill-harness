---
artifactType: research-synthesis
artifactId: public-agentic-ecosystem-research-2026-06-19
owner: codex
issue: skill-harness-4r6
status: ready
reviewSurface: generated/review/research/public-agentic-ecosystem-research-2026-06-19.html
evidenceLinks:
  - docs/third-party-skill-intake.md
  - docs/dependency-provenance.md
  - docs/repo-placement.md
  - docs/agent-operating-skills.md
  - scripts/external_skill_intake.py
  - scripts/dependencies.json
  - https://www.youtube.com/watch?v=ZK3JhU73W18
  - https://github.com/openai/codex
  - https://github.com/openai/openai-cookbook
  - https://github.com/gastownhall/beads
  - https://github.com/gastownhall/gastown
  - https://github.com/bmad-code-org/bmad-method
  - https://github.com/BuilderIO/skills
freshness:
  generatedAt: 2026-06-19
  sourceFirst: true
---

# Public Agentic Ecosystem Research

## Purpose

Map the public skill, agent, task-memory, Cursor/Codex rule, swarm, and agent-workflow ecosystem so `skill-harness` can decide what to copy, adapt, fixture, or ignore.

This expands the earlier BuilderIO scan. It now includes Beads and Gas Town from the Gas Town Hall ecosystem, Codex cookbook and record/replay skill creation, Cursor rules/MDC repositories, Claude/Anthropic skill and subagent surfaces, agency-style frameworks, MCP tooling, and methodology-led systems such as BMad.

## Method Note

This artifact separates discovery metadata from adoption evidence.

Discovery metadata is a starting point: stars, update dates, license signal, and public positioning. Adoption evidence requires cloning or pinned snapshots, running `scripts/external_skill_intake.py`, checking license/provenance, and creating fixtures. That intake work has not been completed yet, so adoption recommendations here are hypotheses.

Current confidence levels:

| Evidence Type | Confidence | Use |
| --- | --- | --- |
| Official docs and official repos | high | Host surface design, naming, install conventions, and fixture targets. |
| GitHub API metadata from 2026-06-19 | medium | Discovery and prioritization only. |
| Community catalog README claims | medium-low | Taxonomy and install-surface discovery. |
| Stars and forks | low | Popularity signal only, not quality or safety. |
| Unscanned third-party skill bodies | low | Do not copy until intake review passes. |

## Short Answer

Research is worth continuing, but the useful comparison is by operating model, not by stars alone. The landscape has at least nine distinct schools of thought:

1. Record-and-replay skill creation.
2. Agent task memory and issue graphs.
3. Multi-agent workspace managers.
4. Agency and swarm orchestration frameworks.
5. Skill and subagent marketplaces.
6. Always-on instruction and repo doctrine.
7. Cursor/rules-first IDE configuration.
8. MCP and tool-boundary capability packs.
9. Spec-driven or methodology-led delivery frameworks.

For `skill-harness`, the safe path remains intake-first:

- Use public repos as pattern references, fixtures, and comparison material.
- Keep public repos outside the live install path.
- Promote only reviewed first-party rewrites into `packs/`.
- Add fixture coverage by install posture, not by repo brand.
- Treat stars as discovery signal, not proof.

## Discovery Metadata

Metadata below was collected on 2026-06-19 from GitHub CLI/API, repository pages, YouTube metadata via `yt-dlp`, and search results. Counts are approximate point-in-time signals.

| Repo Or Resource | Stars | License Signal | Latest Signal | Lane | Adoption Disposition |
| --- | ---: | --- | --- | --- | --- |
| `anthropics/skills` | 152.5k | mixed/no repo-wide license signal | 2026-06-09 | official skills | Reference and fixture; copy only per-skill after license review. |
| `openai/codex` | 92.0k | Apache-2.0 | 2026-06-18 | Codex host/runtime | Reference for host behavior and AGENTS conventions. |
| `modelcontextprotocol/servers` | 87.4k | NOASSERTION | 2026-06-17 | MCP/tool boundary | Fixture/reference for MCP config lane. |
| `openai/openai-cookbook` | 74.2k | MIT | 2026-06-18 | examples/cookbook | Reference and examples; not baseline content. |
| `microsoft/autogen` | 59.1k | CC-BY-4.0 | 2026-04-15 | agent framework | Historical/runtime reference only. |
| `shanraisshan/claude-code-best-practice` | 58.3k | MIT | 2026-06-18 | operating doctrine | Reference for doctrine patterns. |
| `crewAIInc/crewAI` | 53.9k | MIT | 2026-06-18 | agency framework | Runtime reference only. |
| `bmad-code-org/bmad-method` | 49.3k | NOASSERTION | 2026-06-18 | methodology framework | Stage-gate comparison; fixture candidate. |
| `PatrickJS/awesome-cursorrules` | 40.0k | CC0-1.0 | 2026-05-30 | Cursor rules catalog | Cursor rule fixture/index reference. |
| Cursor docs: Rules, Skills, Subagents | n/a | product docs | 2026-06-19 reviewed | Cursor host surface | High-confidence source for `.cursor/rules/*.mdc`, `AGENTS.md`, skills, and subagent compatibility behavior. |
| `agno-agi/agno` | 40.8k | Apache-2.0 | 2026-06-18 | agent platform | Runtime/control-plane reference only. |
| `wshobson/agents` | 36.9k | MIT | 2026-06-17 | agent marketplace | Marketplace fixture/reference. |
| `github/awesome-copilot` | 35.3k | MIT | 2026-06-18 | marketplace/index | Discovery and fixture source. |
| `langchain-ai/langgraph` | 35.1k | MIT | 2026-06-18 | workflow graph | Runtime reference only. |
| `cursor/cursor` | 33.0k | no license detected | 2026-05-12 | Cursor host | Host behavior reference, not copy target. |
| `github/github-mcp-server` | 30.8k | MIT | 2026-06-18 | MCP/tool boundary | Narrow official MCP fixture/reference. |
| `eyaltoledano/claude-task-master` | 27.6k | NOASSERTION | 2026-04-28 | task planning | Planning/task-state fixture candidate. |
| `openai/openai-agents-python` | 27.2k | MIT | 2026-06-18 | agent SDK | Vocabulary/reference for orchestration. |
| `gastownhall/beads` | 24.6k | MIT | 2026-06-18 | task memory | First-class peer; already aligned with repo workflow. |
| `OthmanAdi/planning-with-files` | 23.6k | MIT | 2026-06-16 | file planning | Lightweight planning fallback reference. |
| `humanlayer/12-factor-agents` | 23.4k | NOASSERTION | 2025-09-21 | production doctrine | Reference only until license clarity. |
| `SuperClaude-Org/SuperClaude_Framework` | 23.3k | MIT | 2026-06-13 | command/persona framework | Meta-framework reference, not dependency. |
| `agentsmd/agents.md` | 22.3k | MIT | 2026-03-12 | instruction standard | Canonical instruction-format reference. |
| `VoltAgent/awesome-claude-code-subagents` | 22.1k | MIT | 2026-06-16 | subagent catalog | Install-surface fixture candidate. |
| `openai/swarm` | 21.6k | MIT | 2026-04-15 | educational orchestration | Handoff model reference. |
| `gastownhall/gastown` | 16.0k | MIT | 2026-06-17 | workspace manager | Agent-loop research; not default runtime. |
| `can1357/oh-my-pi` | 13.4k | MIT | 2026-06-18 | coding agent runtime | Runtime reference and fixture candidate. |
| `contains-studio/agents` | 12.4k | no license detected | 2025-07-28 | subagent pack | Fixture/reference only until license review. |
| `lastmile-ai/mcp-agent` | 8.4k | Apache-2.0 | 2026-01-25 | MCP orchestration | MCP composition reference. |
| `microsoft/SkillOpt` | 8.3k | MIT | 2026-06-17 | skill optimization | Future evidence-loop research. |
| `snarktank/ai-dev-tasks` | 7.8k | Apache-2.0 | 2025-11-05 | AI task files | Task-planning fixture candidate. |
| `HKUDS/ClawTeam` | 5.3k | MIT | 2026-05-09 | swarm workspace | Runtime reference and fixture candidate. |
| `VRSEN/agency-swarm` | 4.5k | MIT | 2026-06-18 | agency framework | Role/handoff reference. |
| `sanjeed5/awesome-cursor-rules-mdc` | 3.5k | no license detected in API | 2026-05-19 | Cursor MDC catalog | Cursor `.mdc` fixture candidate. |
| `vanzan01/cursor-memory-bank` | 3.0k | unclear in page scrape | 2026-01-07 | Cursor memory bank | Persistent markdown state and custom-mode school; fixture/reference only. |
| `gotalab/cc-sdd` | 3.5k | MIT | 2026-05-20 | spec-driven delivery | Workflow-profile comparison. |
| `openai/openai-agents-js` | 3.2k | MIT | 2026-06-18 | agent SDK | JS/TS orchestration reference. |
| `microsoft/skills` | 2.6k | MIT | 2026-06-18 | mixed skills/MCP/agents | Mixed-host fixture/reference. |
| `FrancyJGLisboa/agent-skill-creator` | 1.5k | MIT | 2026-06-13 | skill generator | Cross-host packaging reference. |
| `BuilderIO/skills` | 1.2k | MIT | 2026-06-17 | skill pack | Clean small skill-pack reference. |
| `desplega-ai/agent-swarm` | 0.5k | MIT | 2026-06-18 | company swarm | Runtime reference only. |
| `mxyhi/ok-skills` | 0.4k | Apache-2.0 | 2026-06-18 | curated skill pack | Small direct skill reference. |
| `jabrena/cursor-rules-java` | 0.4k | Apache-2.0 | 2026-06-17 | Cursor Java workflow | Opinionated enterprise Cursor rules/skills/MCP reference. |
| `matank001/cursor-security-rules` | 0.4k | MIT | 2025-08-27 | Cursor security rules | Narrow scoped security-rule reference. |
| `shinpr/sub-agents-skills` | 45 | MIT | 2026-06-12 | cross-host agents/skills | Explicit Codex/Claude/Cursor/Gemini routing fixture. |
| `tecnomanu/agent-rules-kit` | 34 | ISC | 2025-09-14 | Cursor mirror docs | Rule/docs sync reference. |
| `yu-iskw/cursor-experiments` | 25 | unclear in page scrape | 2026-02-27 | Cursor multi-agent demo | Experimental orchestration reference only. |
| `lastmile-ai/mcp-eval` | 24 | Apache-2.0 | 2025-11-19 | MCP evals | Future MCP conformance reference. |
| `launchdarkly-labs/cursor-rules` | 2 | MIT | 2026-04-01 | vendor Cursor rules | Scoped vendor-rule fixture candidate. |
| `alipajand/agent-context-doctor` | 0 | MIT | 2026-06-12 | context audit | Interesting tiny fixture: audits AGENTS, CLAUDE, Cursor rules, Copilot instructions. |
| `dataeclipse/skilldrift` | 0 | MIT | 2026-05-09 | drift check | Interesting tiny fixture: detects drift between Cursor rules and Claude/Codex skills. |
| `junjunup/skillops-forge` | 0 | MIT | 2026-05-28 | skill lint | Interesting tiny fixture: offline SKILL/CLAUDE/Cursor risk scoring. |
| YouTube: `Record & Replay in Codex` | n/a | YouTube content | 2026-06-18 | record/replay skill creation | Demonstrates user-recorded process converted into reusable Codex skill. |

```artifact-infographic
{
  "title": "Ecosystem Lanes In Current Scan",
  "tool": "vega-lite",
  "kind": "bar",
  "summary": "The sample intentionally spans operating models rather than only SKILL.md repositories.",
  "values": [
    { "label": "Skill and subagent packs", "value": 11 },
    { "label": "Agent frameworks", "value": 8 },
    { "label": "Task memory and planning", "value": 5 },
    { "label": "MCP and tool boundaries", "value": 5 },
    { "label": "Instruction doctrine", "value": 5 },
    { "label": "Cursor/rules systems", "value": 11 },
    { "label": "Workspace/swarm managers", "value": 4 },
    { "label": "Indexes/marketplaces", "value": 4 },
    { "label": "Record/replay skill creation", "value": 1 }
  ]
}
```

## Schools Of Thought

### 1. Record-And-Replay Skill Creation

Representative sources:

- YouTube: `Record & Replay in Codex`, uploaded by OpenAI on 2026-06-18.
- `openai/codex`
- `openai/skills`

OpenAI's record/replay demo shows a user recording a manual workflow, Codex reviewing the recording, and turning the learned process into a reusable skill. The example workflow is a YouTube publishing process: read metadata from a spreadsheet, find assets, add title/description/thumbnail/captions, save private, and verify. The same pattern is presented as applying to pull request formatting or calendar-invite setup.

The important theory is "show, do not just prompt." Instead of hand-writing a large prompt for every repeated process, the user records one execution trace and uses it as skill-creation input.

Implication for `skill-harness`: this is highly relevant. The repo should treat record/replay as a future source type for skills:

- `recording -> learned skill -> source review -> installed skill`
- generated skill needs provenance and redaction review
- the replay target may use browser use, computer use, plugins, or MCP
- record/replay output should not bypass the same intake and manifest gates as hand-authored skills

### 2. Host-Native Layering: AGENTS, Skills, MCP, Subagents

Representative sources:

- OpenAI Codex docs and `openai/codex`
- `openai/openai-cookbook`
- `openai/openai-agents-python`
- `openai/openai-agents-js`
- Anthropic skills, subagents, teams, workflows, and cookbook docs

The OpenAI/Codex and Anthropic/Claude lanes converge on a layered model:

- small repo instructions for always-on behavior
- triggered skills for reusable procedures
- MCP/apps/plugins for external tools and data
- subagents or managed agents for isolated delegation
- SDK/runtime layers only when building agent applications

Implication for `skill-harness`: copy the architecture, not the bodies. Keep root instructions small, push procedures into skills, keep MCP/plugin surfaces separate, and use runtime SDKs as references unless an explicit agent-runtime profile is introduced.

### 3. Task Memory Beats Loose Chat State

Representative repos:

- `gastownhall/beads`
- `eyaltoledano/claude-task-master`
- `OthmanAdi/planning-with-files`
- `snarktank/ai-dev-tasks`

Beads is the strongest signal in this lane and already matters locally. It is a distributed graph issue tracker for AI agents, powered by Dolt, and positions itself as structured memory for agents. Task Master emphasizes PRD parsing, dependency graphs, MCP tools, editor rules, and local task execution. planning-with-files is simpler: persistent files that survive context loss.

Implication for `skill-harness`: Beads should remain the primary task-state integration, but fixture tests should cover lighter fallback plan files and Task Master-style project task graphs.

### 4. Multi-Agent Workspace Managers

Representative repos:

- `gastownhall/gastown`
- `HKUDS/ClawTeam`
- `desplega-ai/agent-swarm`

Gas Town appears to be the repo family the user called "gashouse". Its theory is workspace orchestration: agents have identities, handoffs, workspaces, hooks, and a shared Beads ledger. ClawTeam and desplega-ai/agent-swarm push harder into lead/worker execution, worktrees, tmux, Docker, inboxes, dashboards, and integrations.

Implication for `skill-harness`: useful for agent-loop research, not default install. The risk is category error: once `skill-harness` owns persistent agent identities, sessions, mailboxes, and merge queues, it stops being a harness and becomes a runtime platform.

### 5. Agency And Swarm Frameworks

Representative repos:

- `openai/openai-agents-python`
- `openai/openai-agents-js`
- `openai/swarm`
- `VRSEN/agency-swarm`
- `crewAIInc/crewAI`
- `langchain-ai/langgraph`
- `agno-agi/agno`
- `microsoft/autogen`

The independent agent-framework review was clear: `skill-harness` should standardize a taxonomy, not adopt a runtime. The OpenAI Agents SDK is the best current vocabulary for tools, guardrails, sessions, handoffs, hosted/local tools, and MCP. OpenAI Swarm is useful as a minimal teaching model. Agency Swarm and CrewAI demonstrate role/org metaphors. LangGraph represents durable graph workflows. Agno represents a control-plane platform. AutoGen is historically important but no longer the best forward anchor.

Implication for `skill-harness`: use these for language and fixtures. Do not install them in the default baseline.

### 6. Skill, Subagent, And Marketplace Repos

Representative repos:

- `anthropics/skills`
- `openai/skills`
- `BuilderIO/skills`
- `wshobson/agents`
- `github/awesome-copilot`
- `VoltAgent/awesome-claude-code-subagents`
- `contains-studio/agents`
- `microsoft/skills`
- `mxyhi/ok-skills`
- `FrancyJGLisboa/agent-skill-creator`

This is the closest direct lane for `skill-harness`. BuilderIO is small and clean. Anthropic/OpenAI official surfaces define format expectations. Microsoft, wshobson, VoltAgent, GitHub Copilot catalogs, and smaller cross-agent tools show the marketplace direction.

The main risk is install posture: global copies, plugin installs, shell helpers, curl-piped installers, MCP config changes, or large no-license prompt bodies.

Implication for `skill-harness`: strengthen `scripts/external_skill_intake.py` and create synthetic fixtures for skill-only, plugin, script, curl, global-copy, no-license, and mixed MCP cases.

### 7. Always-On Instruction Doctrine

Representative repos:

- `agentsmd/agents.md`
- `shanraisshan/claude-code-best-practice`
- `humanlayer/12-factor-agents`
- `SuperClaude-Org/SuperClaude_Framework`

These repos treat agent behavior as persistent operating doctrine, not triggered task skills. AGENTS.md is the simplest standard. SuperClaude and Claude Code best-practice repos are more opinionated around commands, personas, modes, and workflow doctrine. 12-factor-agents is broader production doctrine for LLM software.

Implication for `skill-harness`: keep `AGENTS.md`, `CLAUDE.md`, `AGENT_INSTRUCTIONS.md`, `.codex`, `.claude`, `.cursor`, and Copilot instruction surfaces aligned from one canonical source. Avoid copying large always-on prompt bodies.

### 8. Cursor Rules And MDC-First IDE Configuration

Representative sources:

- Cursor docs for Rules and Cloud Agents.
- Cursor docs for Skills and Subagents.
- `PatrickJS/awesome-cursorrules`
- `sanjeed5/awesome-cursor-rules-mdc`
- `vanzan01/cursor-memory-bank`
- `jabrena/cursor-rules-java`
- `matank001/cursor-security-rules`
- `blefnk/awesome-cursor-rules`
- `tecnomanu/agent-rules-kit`
- `shinpr/sub-agents-skills`
- `yu-iskw/cursor-experiments`
- `launchdarkly-labs/cursor-rules`
- `alipajand/agent-context-doctor`
- `dataeclipse/skilldrift`
- `junjunup/skillops-forge`

Cursor's current theory is not giant prompt blobs. It is a versioned host stack of rules, `AGENTS.md`, skills, and native subagents. The high-confidence official surface is:

- project rules live as `.cursor/rules/*.mdc`; plain `.md` files in that folder are ignored
- `AGENTS.md` is the simple repository instruction alternative
- skills can be loaded from `.agents/skills`, `.cursor/skills`, `.claude/skills`, and `.codex/skills`
- subagents can be loaded from `.cursor/agents`, with compatibility for `.claude/agents` and `.codex/agents`
- Cursor Cloud Agents add a host-native remote execution surface

Community Cursor repos split into three schools. The first is the catalog school: big lists such as `awesome-cursorrules` and `awesome-cursor-rules-mdc`. The second is the memory-bank/custom-modes school, represented by `cursor-memory-bank`, which is useful but partly legacy. The third is cross-host portability: shared assets consumed by Cursor, Claude, Codex, Gemini, or Copilot, with small host-specific overlays only where necessary.

Implication for `skill-harness`: Cursor should be a first-class render target and fixture lane, but not a separate content silo. The interesting small repos are not necessarily popular; `skilldrift`, `agent-context-doctor`, and `skillops-forge` show the exact problem `skill-harness` has: keeping `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/*.mdc`, Copilot instructions, and `SKILL.md` from drifting apart.

### 9. MCP And Tool-Boundary Capability Packs

Representative repos:

- `modelcontextprotocol/servers`
- `github/github-mcp-server`
- `lastmile-ai/mcp-agent`
- `lastmile-ai/mcp-eval`

MCP changes the adoption question. Skills and prompts are text inputs; MCP servers are tool and auth surfaces. The GitHub MCP server is a narrow official capability pack. `mcp-agent` wraps MCP servers into agent workflows and durable execution. `mcp-eval` points toward testing real MCP tool use.

Implication for `skill-harness`: MCP descriptors and plugins need their own manifest lane and approval gates. They should not be treated as ordinary skills.

### 10. Spec-Driven And Methodology-Led Delivery

Representative repos:

- `bmad-code-org/bmad-method`
- `gotalab/cc-sdd`
- `humanlayer/12-factor-agents`
- `microsoft/SkillOpt`

BMad is especially relevant because it is not just a prompt repo. It has an installer, modules, agents, skills architecture, and end-to-end workflows from analyst to product brief, PRD, architecture, stories, and implementation. `cc-sdd` is narrower: approved specs to implementation. SkillOpt is different again: training and optimizing natural-language skills from trajectories.

Implication for `skill-harness`: keep methodology as first-party doctrine/playbook content. Copying BMad wholesale would conflict with `skill-harness`' suite-entrypoint role, but its stage gates and installer ergonomics are worth deeper comparison.

## Six Agent Opinions

| Lane | Independent Opinion | Harness Implication |
| --- | --- | --- |
| OpenAI/Codex | Copy the layered architecture: small `AGENTS.md`, skills for richer workflows, MCP/plugins for external systems, subagents after that. Treat cookbook as examples, not baseline. | Adapt OpenAI skill structure and Codex-as-MCP concepts; keep SDKs optional. |
| Claude/Anthropic | Official surfaces are stronger than community catalogs. Skills, subagents, teams, workflows, and plugins form a clear ladder. | Emit first-party packages targeting official Claude surfaces; do not vendor community packs. |
| Beads/Gas Town/task memory | Beads is real signal because it matches durable task state already used here. Gas Town is workspace/runtime research, not core harness. | Strengthen Beads profile and add planning/task-memory fixtures. |
| Agency/swarm frameworks | `skill-harness` should standardize taxonomy, not adopt a runtime. Agents SDK and MCP are useful vocabulary; CrewAI/LangGraph/Agno are reference-only. | Keep runtime frameworks out of default install. |
| Quality review | Current research is still metadata-heavy and broad. Adoption evidence requires intake scans, pinned commits, and fixture matrix. | Split discovery from adoption and mark recommendations as hypotheses. |
| Cursor/rules lane | Cursor is now rules, `AGENTS.md`, skills, and native subagents, with compatibility for `.claude` and `.codex`. Small drift/audit tools may matter more than big catalogs. | Add Cursor as render target/fixture lane; prioritize drift checks across host instruction files. |

## Copy, Adapt, Fixture, Ignore

| Action | Good Candidates | Rule |
| --- | --- | --- |
| Copy small structure only | BuilderIO folder layout, AGENTS.md precedence examples, plugin manifest field shapes, Cursor `.mdc` shape | Keep license attribution when any protected expression is copied. |
| Adapt into first-party packs | workflow gates, task-shaping prompts, agent handoff packets, Beads profile ideas, Cursor render targets, fixture schemas | Rewrite in 45ck voice and safety model. |
| Fixture only | global-copy installers, curl installers, no-license catalogs, MCP descriptors, plugin bundles, record/replay outputs | Use synthetic repos or pinned scan output, not live install dependencies. |
| Reference only | AutoGen, LangGraph, CrewAI, Agno, Gas Town runtime topology, OpenAI Agents SDK runtime, Cursor Cloud Agents | Inform architecture; do not bundle runtime frameworks by default. |
| Reject for default install | no-license prompt bodies, shell-piped installers, global agent mutation, opaque hosted swarms, unsandboxed record/replay skills | Keep outside shared install flow. |

```artifact-infographic
{
  "title": "Recommended Adoption Disposition",
  "tool": "observable-plot",
  "kind": "bar",
  "summary": "Most high-signal repos should inform design or fixtures rather than become dependencies.",
  "values": [
    { "label": "First-party rewrite", "value": 10 },
    { "label": "Fixture only", "value": 14 },
    { "label": "Architecture reference", "value": 12 },
    { "label": "Discovery index", "value": 5 },
    { "label": "Copy directly", "value": 1 }
  ]
}
```

## Intake Fixture Matrix

| Fixture Archetype | Example Sources | Expected Gate |
| --- | --- | --- |
| Pure `SKILL.md` pack | BuilderIO, OpenAI/Anthropic skills | Parse and compare with local catalog; no global install. |
| Skills with helper scripts | community skill packs | Flag executable helpers for manual review. |
| Claude subagent global-copy pack | contains-studio, VoltAgent | Flag global mutation and missing license where applicable. |
| Plugin manifest | Anthropic/OpenAI plugin surfaces, VoltAgent install patterns | Separate plugin lane from ordinary skills. |
| MCP config/server | GitHub MCP server, mcp-agent, MCP reference servers | Separate tool/auth lane with approval gates. |
| Cursor `.mdc` rules | awesome-cursorrules, awesome-cursor-rules-mdc | Validate as host instruction/rules surface, not SKILL.md. |
| Cursor skills/subagents compatibility | Cursor docs, shinpr/sub-agents-skills | Validate shared `.agents`, `.claude`, `.codex`, and `.cursor` loading behavior. |
| Copilot instructions/prompts | github/awesome-copilot | Validate as host instruction/prompt surface. |
| Record/replay generated skill | OpenAI record/replay demo | Require provenance, redaction, and human review before install. |
| Task-memory repo | Beads, Task Master, planning-with-files | Validate state ownership, lock/sync behavior, and tracker boundary. |
| Curl/shell installer | community catalogs, SuperClaude-style installers | Quarantine until script review passes. |
| No-license catalog | contains-studio, some official/reference repos | Inspiration only until license clarity. |

## Priority Shortlist

### Tier 1: Study Next

1. `openai/skills` and OpenAI Codex record/replay.
   - Harness question: should `skill-harness` define a learned-skill provenance model?

2. `gastownhall/beads`.
   - Harness question: should `skill-harness` expose stronger Beads profiles and task-memory fixtures?

3. Cursor rules and MDC repositories.
   - Harness question: should `.cursor/rules` become a first-class render target with drift checks?

4. `bmad-code-org/bmad-method`.
   - Harness question: should `skill-harness` document an alternative-method comparison or add BMad-like stage-gate fixtures?

5. `VoltAgent/awesome-claude-code-subagents`.
   - Harness question: can `external_skill_intake.py` safely classify plugin/manual/script/curl installs?

6. `mcp-agent` plus `github-mcp-server`.
   - Harness question: should plugin/MCP descriptors get their own manifest lane?

### Tier 2: Monitor

- `gastownhall/gastown`
- `wshobson/agents`
- `microsoft/skills`
- `mxyhi/ok-skills`
- `FrancyJGLisboa/agent-skill-creator`
- `gotalab/cc-sdd`
- `desplega-ai/agent-swarm`
- `HKUDS/ClawTeam`
- `dataeclipse/skilldrift`
- `alipajand/agent-context-doctor`
- `junjunup/skillops-forge`

### Tier 3: Reference Only

- `crewAIInc/crewAI`
- `langchain-ai/langgraph`
- `microsoft/autogen`
- `agno-agi/agno`
- `openai/swarm`
- `openai/openai-agents-python`
- `openai/openai-agents-js`

## Gaps In Current Research

- The shortlist has not yet been cloned into `../skill-intake`.
- `scripts/external_skill_intake.py` has not yet been run across the shortlist.
- Repo rows do not yet record commit SHA pins.
- License details for no-license or NOASSERTION repos need repo-level review.
- YouTube record/replay claims are based on verified title/metadata and available English subtitles, but not on source code or product docs.
- The Cursor lane still needs deeper official-doc extraction beyond repository metadata.
- Stars over-represent awareness, timing, and brand; they do not measure skill quality or safety.
- We have not yet created synthetic intake fixtures for the major install surfaces found here.

## Recommended Next Work

1. Clone Tier 1 repos into `../skill-intake`, pinned by commit SHA.
2. Run `python scripts/external_skill_intake.py --json-output ...` across the cloned shortlist.
3. Add a third-party-intake scorecard:
   - license clarity
   - source commit SHA
   - install surface
   - global mutation risk
   - executable helper risk
   - MCP/plugin risk
   - prompt/doctrine copy risk
   - record/replay provenance risk
   - first-party rewrite target
4. Create synthetic fixture repos for each fixture archetype in the matrix above.
5. Update `docs/third-party-skill-intake.md` with the taxonomy once the fixture suite exists.
6. Consider a `host-instruction-drift` check covering `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/*.mdc`, Copilot instructions, and `SKILL.md`.

## Preliminary Decision

Do not copy any large public agentic repo wholesale.

Do adapt these ideas:

- OpenAI/Codex record-and-replay as a future skill provenance source.
- Official Codex/Claude layered architecture: instructions, skills, tools/plugins/MCP, subagents.
- Beads-style durable task graph as first-class context.
- Cursor `.mdc` rules as a host-specific render target and drift surface.
- Gas Town-style identity, handoff, and supervision vocabulary as future agent-loop research, not default runtime.
- BMad-style stage gates as comparison material for workflow profiles.
- Marketplace-style packaging fixtures for intake tests.
- MCP/tool-boundary separation as a separate manifest lane.
- SkillOpt-style trajectory-to-skill optimization as future evidence-loop research.

## Source Index

- OpenAI Codex: https://github.com/openai/codex
- OpenAI Cookbook: https://github.com/openai/openai-cookbook
- OpenAI Agents Python SDK: https://github.com/openai/openai-agents-python
- OpenAI Agents JS SDK: https://github.com/openai/openai-agents-js
- OpenAI Swarm: https://github.com/openai/swarm
- OpenAI record/replay video: https://www.youtube.com/watch?v=ZK3JhU73W18
- `gastownhall/beads`: https://github.com/gastownhall/beads
- `gastownhall/gastown`: https://github.com/gastownhall/gastown
- Steve Yegge Beads article: https://steve-yegge.medium.com/introducing-beads-a-coding-agent-memory-system-637d7d92514a
- `bmad-code-org/bmad-method`: https://github.com/bmad-code-org/bmad-method
- `eyaltoledano/claude-task-master`: https://github.com/eyaltoledano/claude-task-master
- `PatrickJS/awesome-cursorrules`: https://github.com/PatrickJS/awesome-cursorrules
- `sanjeed5/awesome-cursor-rules-mdc`: https://github.com/sanjeed5/awesome-cursor-rules-mdc
- `vanzan01/cursor-memory-bank`: https://github.com/vanzan01/cursor-memory-bank
- `jabrena/cursor-rules-java`: https://github.com/jabrena/cursor-rules-java
- `matank001/cursor-security-rules`: https://github.com/matank001/cursor-security-rules
- `shinpr/sub-agents-skills`: https://github.com/shinpr/sub-agents-skills
- `yu-iskw/cursor-experiments`: https://github.com/yu-iskw/cursor-experiments
- `launchdarkly-labs/cursor-rules`: https://github.com/launchdarkly-labs/cursor-rules
- `tecnomanu/agent-rules-kit`: https://github.com/tecnomanu/agent-rules-kit
- `dataeclipse/skilldrift`: https://github.com/dataeclipse/skilldrift
- `alipajand/agent-context-doctor`: https://github.com/alipajand/agent-context-doctor
- `junjunup/skillops-forge`: https://github.com/junjunup/skillops-forge
- `VRSEN/agency-swarm`: https://github.com/VRSEN/agency-swarm
- `crewAIInc/crewAI`: https://github.com/crewAIInc/crewAI
- `microsoft/autogen`: https://github.com/microsoft/autogen
- `langchain-ai/langgraph`: https://github.com/langchain-ai/langgraph
- `agno-agi/agno`: https://github.com/agno-agi/agno
- `HKUDS/ClawTeam`: https://github.com/HKUDS/ClawTeam
- `desplega-ai/agent-swarm`: https://github.com/desplega-ai/agent-swarm
- `anthropics/skills`: https://github.com/anthropics/skills
- `BuilderIO/skills`: https://github.com/BuilderIO/skills
- `wshobson/agents`: https://github.com/wshobson/agents
- `github/awesome-copilot`: https://github.com/github/awesome-copilot
- `VoltAgent/awesome-claude-code-subagents`: https://github.com/VoltAgent/awesome-claude-code-subagents
- `contains-studio/agents`: https://github.com/contains-studio/agents
- `modelcontextprotocol/servers`: https://github.com/modelcontextprotocol/servers
- `github/github-mcp-server`: https://github.com/github/github-mcp-server
- `lastmile-ai/mcp-agent`: https://github.com/lastmile-ai/mcp-agent
- `lastmile-ai/mcp-eval`: https://github.com/lastmile-ai/mcp-eval
