---
id: AUTH-001
title: "User Authentication Module"
state: in_progress
kind: functional
required_evidence:
  implementation: E0
  test_coverage: E0
waivers:
  - kind: missing-verification
    target: AUTH-001
    owner: "agent"
    reason: "Specs declare E0 evidence (declarative). Integration tests in server.test.js exercise all auth paths and pass (28/28). No runtime verification provider is configured; advisory warn is expected and acceptable at this evidence level."
    expires: "2026-07-01"
---

## Overview

The authentication module provides user identity and session management for the Task Management REST API. It uses SHA-256 with a per-user salt to hash passwords, and issues opaque session tokens that expire after 30 minutes of inactivity. All credentials are stored in-memory; no external persistence is used.

## Requirements

### AUTH-001-R1: Login
- `login(username, password)` accepts a username and plaintext password.
- The function salts and hashes the password with SHA-256, then compares against the stored hash.
- On success it generates a cryptographically random session token, stores it with the current timestamp, and returns `{ token }`.
- On failure it throws a descriptive error (do not reveal whether the username or password was wrong).

### AUTH-001-R2: Logout
- `logout(token)` removes the session from the in-memory store.
- If the token does not exist the call is a no-op (idempotent).

### AUTH-001-R3: Session validation
- `isValidSession(token)` returns `true` if the token exists and was last used within the past 30 minutes.
- It also refreshes the last-used timestamp on every successful check (sliding expiry).
- Returns `false` for unknown or expired tokens.

### AUTH-001-R4: User store
- The module contains a static in-memory map of `{ username → { hash, salt } }`.
- At least two hardcoded test users must be present so integration tests can run without external setup.
- Passwords must never appear in plaintext anywhere in source code or test files.

### AUTH-001-R5: Token format
- Tokens are generated with `node:crypto` `randomBytes(32)` encoded as hex (64-character string).

## Acceptance Criteria

- AC1: `login` with valid credentials returns an object containing a `token` string of length 64.
- AC2: `login` with invalid credentials throws an error.
- AC3: `isValidSession` returns `true` immediately after a successful login.
- AC4: `isValidSession` returns `false` after `logout`.
- AC5: `isValidSession` returns `false` for an unknown token.
- AC6: No plaintext password appears in source or test code.

## Out of Scope

- Persistent storage of users or sessions.
- Password reset, registration, or multi-factor authentication.
- JWT or signed tokens.
- Role-based access control.
