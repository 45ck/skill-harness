# Agent instructions

This project uses specgraph for spec tracking and noslop for quality gates.

## Before writing any code

1. Run `/spec-writer` to create a spec document in `docs/` for the feature you are implementing.
2. Confirm the spec is committed before proceeding.

## While writing code

3. Annotate every source file with `@spec` and `@implements` JSDoc tags using `/annotation-writer`.
4. Run `npx specgraph verify` after each significant change and review output with `/verify-interpreter`.

## Before committing

5. Run `/noslop-commit-gate` and resolve all failures before committing.
6. Run `npx specgraph verify` one final time. All requirements must have E1 or higher evidence.

## If a requirement cannot be satisfied

Use `/waiver-writer` to write a justified waiver with an expiry date.
