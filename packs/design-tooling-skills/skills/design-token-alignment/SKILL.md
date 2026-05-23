---
name: "design-token-alignment"
description: "Align design artifacts with code tokens and component APIs so spacing, color, type, and states stay consistent through implementation."
---

Use this skill when a design needs to map cleanly onto an existing component system or token set.

Checklist:

- identify which design values already map to existing tokens
- flag values that are near-matches and should be normalized instead of copied literally
- note any missing tokens, variant gaps, or component API mismatches
- keep the recommendation biased toward reuse before extension
- produce a concise token and component mapping table
- keep token decisions and component-state mappings in a durable source artifact
- generate a visual review surface when humans need to compare token impact, states, or before/after UI behavior

Good output:

- existing tokens reused where possible
- missing tokens called out explicitly
- API gaps separated from pure styling gaps
- implementation guidance that preserves the current design system shape
- source-backed review path when a generated visual comparison is needed

Avoid:

- creating one-off tokens for every visual exception
- solving a component API problem with raw styling overrides
- assuming the design tool naming is already production-ready
