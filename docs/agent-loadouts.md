# Agent Loadouts

This document maps each workflow agent to its curated skill set.

## requirements-analyst

- Skills: problem-statement-refiner, assumptions-constraints-log, requirements-elicitation, requirements-interrogator, acceptance-criteria-writer, requirements-prioritizer, stakeholder-analysis, spec-writer

## requirements-analyst-beads

- Skills: problem-statement-refiner, assumptions-constraints-log, requirements-elicitation, requirements-interrogator, acceptance-criteria-writer, requirements-prioritizer, stakeholder-analysis, spec-writer

## ux-researcher

- Skills: persona-synthesizer, task-analysis-writer, usability-test-planner, representative-task-writer, wireframe-critic, field-study-planner, satisfaction-questionnaire-writer, evaluation-report-writer

## system-modeler

- Skills: use-case-modeler, use-case-description-writer, sequence-diagram-builder, activity-diagram-builder, state-model-builder, domain-class-modeler, model-consistency-checker, scenario-to-uml-transformer

## system-modeler-beads

- Skills: use-case-modeler, use-case-description-writer, sequence-diagram-builder, activity-diagram-builder, state-model-builder, domain-class-modeler, model-consistency-checker, scenario-to-uml-transformer

## software-architect

- Skills: adr-writer, architecture-option-generator, tradeoff-analysis-writer, quality-attribute-scenario-writer, service-decomposition-advisor, runtime-view-writer, deployment-view-writer, integration-boundary-mapper

## software-architect-beads

- Skills: adr-writer, architecture-option-generator, tradeoff-analysis-writer, quality-attribute-scenario-writer, service-decomposition-advisor, runtime-view-writer, deployment-view-writer, integration-boundary-mapper

## web-engineer

- Skills: route-and-controller-planner, client-server-responsibility-splitter, form-validation-designer, request-response-flow-mapper, session-state-strategy-reviewer, server-side-validation-enforcer, mvc-boundary-checker, class-responsibility-checker

## backend-engineer

- Skills: entity-model-designer, repository-layer-designer, transaction-boundary-checker, api-contract-writer, api-error-model-designer, schema-normalizer, data-structure-selector, class-responsibility-checker

## test-designer

- Skills: equivalence-partitioning-generator, boundary-value-generator, decision-table-builder, state-transition-test-designer, test-oracle-writer, coverage-goal-planner, testability-reviewer, nfr-evidence-matrix-builder

## test-designer-beads

- Skills: equivalence-partitioning-generator, boundary-value-generator, decision-table-builder, state-transition-test-designer, test-oracle-writer, coverage-goal-planner, testability-reviewer, nfr-evidence-matrix-builder

## qa-automation-engineer

- Skills: unit-test-writer, integration-test-writer, api-test-suite-builder, ui-test-scenario-writer, fixture-and-test-data-builder, regression-suite-curator, test-readability-reviewer, test-plan-writer

## quality-reviewer

- Skills: maintainability-reviewer, technical-debt-auditor, code-review-checklist-runner, review-severity-scorer, code-smell-detector, refactoring-candidate-ranker, quality-risk-register-builder, rework-plan-writer

## security-reviewer

- Skills: threat-surface-mapper, trust-boundary-identifier, secure-by-design-reviewer, secrets-handling-checker, authn-authz-separator, token-auth-design-reviewer, prompt-injection-reviewer, tool-permission-boundary-checker

## security-reviewer-beads

- Skills: threat-surface-mapper, trust-boundary-identifier, secure-by-design-reviewer, secrets-handling-checker, authn-authz-separator, token-auth-design-reviewer, prompt-injection-reviewer, tool-permission-boundary-checker

## pentest-reviewer

- Skills: authorized-scope-checker, recon-plan-writer, api-attack-surface-mapper, owasp-wstg-checklist-runner, report-finding-writer, repro-steps-writer, remediation-priority-ranker, input-filter-bypass-checker

## delivery-manager

- Skills: sprint-goal-writer, backlog-groomer, project-charter-writer, risk-register-builder, milestone-planner, go-live-readiness-reviewer, rollback-readiness-checker, maintenance-triage-helper

## delivery-manager-beads

- Skills: sprint-goal-writer, backlog-groomer, project-charter-writer, risk-register-builder, milestone-planner, go-live-readiness-reviewer, rollback-readiness-checker, maintenance-triage-helper

## research-writer

- Skills: search-string-designer, paper-screening-checker, related-work-synthesizer, evidence-strength-scorer, gap-finder, critical-analysis-writer, reference-integrity-checker, evaluation-report-writer

## Dependency model

- The agent-to-pack mapping is declared in [scripts/dependencies.json](../scripts/dependencies.json).
- Installers bootstrap those pack repos automatically before installing or rendering agents.
