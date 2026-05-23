# Agent Loadouts

This document maps each shared `skill-harness` agent to its curated skill set.

## requirements-analyst

- Skills: problem-statement-refiner, assumptions-constraints-log, requirements-elicitation, requirements-interrogator, acceptance-criteria-writer, requirements-prioritizer, stakeholder-analysis, spec-writer, frontier-model-context, digital-expert-test, pre-ai-scarcity-thinking

## requirements-analyst-beads

- Skills: problem-statement-refiner, assumptions-constraints-log, requirements-elicitation, requirements-interrogator, acceptance-criteria-writer, requirements-prioritizer, stakeholder-analysis, spec-writer, frontier-model-context, digital-expert-test, pre-ai-scarcity-thinking

## ux-researcher

- Skills: persona-synthesizer, task-analysis-writer, usability-test-planner, representative-task-writer, wireframe-critic, field-study-planner, satisfaction-questionnaire-writer, evaluation-report-writer, frontier-model-context, digital-expert-test, pre-ai-scarcity-thinking

## system-modeler

- Skills: use-case-modeler, use-case-description-writer, sequence-diagram-builder, activity-diagram-builder, state-model-builder, domain-class-modeler, model-consistency-checker, scenario-to-uml-transformer, frontier-model-context, digital-expert-test

## system-modeler-beads

- Skills: use-case-modeler, use-case-description-writer, sequence-diagram-builder, activity-diagram-builder, state-model-builder, domain-class-modeler, model-consistency-checker, scenario-to-uml-transformer, frontier-model-context, digital-expert-test

## software-architect

- Skills: adr-writer, architecture-option-generator, tradeoff-analysis-writer, quality-attribute-scenario-writer, service-decomposition-advisor, runtime-view-writer, deployment-view-writer, integration-boundary-mapper, public-private-hybrid-cloud-reviewer, mcp-server-planning, app-integration-shaping, frontier-model-context, digital-expert-test, legacy-automation-fallback, llm-default-architecture, verification-first-delivery, developer-artifact-shaper, html-review-artifact, artifact-evidence-gate

## software-architect-beads

- Skills: adr-writer, architecture-option-generator, tradeoff-analysis-writer, quality-attribute-scenario-writer, service-decomposition-advisor, runtime-view-writer, deployment-view-writer, integration-boundary-mapper, public-private-hybrid-cloud-reviewer, mcp-server-planning, app-integration-shaping, frontier-model-context, digital-expert-test, legacy-automation-fallback, llm-default-architecture, verification-first-delivery, developer-artifact-shaper, html-review-artifact, artifact-evidence-gate

## web-engineer

- Skills: route-and-controller-planner, client-server-responsibility-splitter, form-validation-designer, request-response-flow-mapper, session-state-strategy-reviewer, server-side-validation-enforcer, mvc-boundary-checker, class-responsibility-checker, figma-implementation-planning, design-token-alignment, frontier-model-context, digital-expert-test, legacy-automation-fallback, llm-default-architecture, verification-first-delivery

## backend-engineer

- Skills: entity-model-designer, repository-layer-designer, transaction-boundary-checker, api-contract-writer, api-error-model-designer, schema-normalizer, data-structure-selector, class-responsibility-checker, frontier-model-context, digital-expert-test, legacy-automation-fallback, llm-default-architecture, verification-first-delivery

## test-designer

- Skills: equivalence-partitioning-generator, boundary-value-generator, decision-table-builder, state-transition-test-designer, test-oracle-writer, coverage-goal-planner, testability-reviewer, nfr-evidence-matrix-builder, frontier-model-context, digital-expert-test, verification-first-delivery

## test-designer-beads

- Skills: equivalence-partitioning-generator, boundary-value-generator, decision-table-builder, state-transition-test-designer, test-oracle-writer, coverage-goal-planner, testability-reviewer, nfr-evidence-matrix-builder, frontier-model-context, digital-expert-test, verification-first-delivery

## qa-automation-engineer

- Skills: unit-test-writer, integration-test-writer, api-test-suite-builder, ui-test-scenario-writer, fixture-and-test-data-builder, regression-suite-curator, test-readability-reviewer, test-plan-writer, frontier-model-context, digital-expert-test, legacy-automation-fallback, verification-first-delivery

## quality-reviewer

- Skills: fagan-inspection, maintainability-reviewer, technical-debt-auditor, code-review-checklist-runner, review-severity-scorer, code-smell-detector, refactoring-candidate-ranker, quality-risk-register-builder, rework-plan-writer, frontier-model-context, digital-expert-test, legacy-automation-fallback, artifact-evidence-gate, artifact-handoff-pack

## security-reviewer

- Skills: threat-surface-mapper, trust-boundary-identifier, secure-by-design-reviewer, secrets-handling-checker, authn-authz-separator, token-auth-design-reviewer, prompt-injection-reviewer, tool-permission-boundary-checker, frontier-model-context, digital-expert-test, verification-first-delivery

## security-reviewer-beads

- Skills: threat-surface-mapper, trust-boundary-identifier, secure-by-design-reviewer, secrets-handling-checker, authn-authz-separator, token-auth-design-reviewer, prompt-injection-reviewer, tool-permission-boundary-checker, frontier-model-context, digital-expert-test, verification-first-delivery

## pentest-reviewer

- Skills: authorized-scope-checker, recon-plan-writer, api-attack-surface-mapper, owasp-wstg-checklist-runner, report-finding-writer, repro-steps-writer, remediation-priority-ranker, input-filter-bypass-checker, frontier-model-context, digital-expert-test, verification-first-delivery

## delivery-manager

- Skills: sprint-goal-writer, backlog-groomer, project-charter-writer, risk-register-builder, milestone-planner, go-live-readiness-reviewer, rollback-readiness-checker, maintenance-triage-helper, update-rollback-planner, frontier-model-context, digital-expert-test, pre-ai-scarcity-thinking, verification-first-delivery, developer-artifact-shaper, artifact-handoff-pack

## delivery-manager-beads

- Skills: sprint-goal-writer, backlog-groomer, project-charter-writer, risk-register-builder, milestone-planner, go-live-readiness-reviewer, rollback-readiness-checker, maintenance-triage-helper, update-rollback-planner, frontier-model-context, digital-expert-test, pre-ai-scarcity-thinking, verification-first-delivery, developer-artifact-shaper, artifact-handoff-pack

## research-writer

- Skills: search-string-designer, paper-screening-checker, related-work-synthesizer, evidence-strength-scorer, gap-finder, critical-analysis-writer, reference-integrity-checker, evaluation-report-writer, frontier-model-context, digital-expert-test, pre-ai-scarcity-thinking, developer-artifact-shaper, html-review-artifact, artifact-evidence-gate

## workflow-engineer

- Skills: issue-driven-delivery, gh-review-followthrough, gh-actions-failure-triage, review-ready-check, frontier-model-context, digital-expert-test, legacy-automation-fallback, llm-default-architecture, verification-first-delivery, developer-artifact-shaper, artifact-handoff-pack

## Dependency model

- The agent-to-repo mapping is declared in [scripts/dependencies.json](../scripts/dependencies.json).
- Installers bootstrap those dependency repos automatically before installing or rendering agents.
