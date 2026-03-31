---
id: AUTH-001
title: "User Authentication Module"
state: in_progress
kind: functional
required_evidence:
  implementation: E0
waivers:
  - kind: missing-verification
    target: AUTH-001
    owner: "agent"
    reason: "No VERIFIED_BY cross-reference claims exist yet. Beads issue tracking is not available in this session environment. The implementation is complete with 19 passing tests; formal verification linkage deferred until Beads is available."
    expires: "2026-07-01"
---

## Overview

This module provides in-memory user authentication for a Node.js application. It handles login, logout, and session management without an external database. Passwords are hashed using SHA-256 with a salt. Sessions expire after 30 minutes of inactivity.

Target audience: backend Node.js developers integrating authentication into a server-side application.

## Requirements

1. **Login** — Accept a username and password string. Validate the credentials against a static in-memory user store (hardcoded map of username → hashed password). Return a non-empty session token string on success. Reject with a descriptive error on invalid credentials.

2. **Logout** — Accept a session token string. Invalidate the token so subsequent validity checks return false.

3. **Session validity** — Provide a `isValidSession(token)` function that returns true if the token exists and has not expired, false otherwise.

4. **Session expiry** — Sessions expire after 30 minutes of inactivity. Each successful `isValidSession` call resets the inactivity timer. An expired token must report as invalid.

5. **Password hashing** — Passwords must be hashed using SHA-256 with a per-user salt using Node's built-in `node:crypto` module. No plaintext passwords may appear in source code, test code, or stored state.

6. **Tests** — Unit tests covering login (success + failure), logout, session validity, and session expiry. Tests must run with `npm test` using Node's built-in `node:test` and `node:assert/strict` modules.

7. **Documentation** — A README.md at the project root explaining the module purpose, how to install dependencies, and how to call `login`, `logout`, and `isValidSession` with a short code example each.

## Acceptance Criteria

- `npm test` exits with code 0 and all tests pass.
- A valid `login(username, password)` call returns a non-empty string token.
- Calling `isValidSession(token)` immediately after login returns `true`.
- Calling `logout(token)` followed by `isValidSession(token)` returns `false`.
- A token whose last-activity timestamp has been backdated beyond 30 minutes reports as invalid.
- No plaintext passwords appear anywhere in source or test code.
- README.md is present at the project root and documents all three operations.

## Out of Scope

- HTTP server or REST API endpoints.
- Persistent storage or external databases.
- Refresh tokens, JWT, or OAuth flows.
- Role-based access control.
- Frontend or browser-side code.
- Rate limiting or brute-force protection.

## Open Questions

- None at this time. All requirements are fully specified.

## Assumptions

- Node.js >= 18 is available (for `node:crypto` and `node:test` built-ins).
- A single module file (`auth.js`) plus a single test file (`auth.test.js`) is the intended scope.
- The static user store may contain a small number of hardcoded test users for demonstration purposes.
