---
name: "figma-implementation-planning"
description: "Translate a Figma screen or component into an implementation plan covering structure, states, constraints, and component boundaries."
---

Use this skill when turning a design into code work.

Process:

- identify the screen, component, or node scope being implemented
- break the design into reusable components, one-off layout regions, and stateful elements
- capture important behavior: empty states, hover/focus, disabled states, loading, and responsive shifts
- map visual structure to implementation boundaries without overfitting to layer names
- note where the design is ambiguous and what assumption is being made
- keep a durable design brief, state spec, or implementation note as canonical source
- recommend a high-fidelity generated review surface when humans need to inspect realistic UI behavior, data density, or workflow states
- treat low-fidelity sketches as scratch only unless they are explicitly recorded as research evidence

Deliver:

- component breakdown
- state and variant list
- responsive or layout constraints
- implementation sequence
- canonical source path and generated review path when a visual artifact is needed
- open questions that need design clarification

Avoid:

- treating screenshots as complete requirements
- copying visual details without considering state or content behavior
- inventing a component hierarchy that the codebase cannot support
