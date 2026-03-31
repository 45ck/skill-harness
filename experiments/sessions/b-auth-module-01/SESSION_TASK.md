# Session Task: User Authentication Module

This is the shared task prompt given verbatim to both Group A and Group B. Do not modify it between sessions. Copy and paste the text in the "Prompt" section below as the first and only message in the Claude Code conversation.

---

## Prompt

Build a user authentication module for a Node.js application. The module must handle login, logout, and session management.

### Requirements

1. **Login** — Accept a username and password. Validate credentials against a static in-memory user store (no real database needed — a hardcoded map of username → hashed password is fine). Return a session token on success. Reject with a clear error on invalid credentials.

2. **Logout** — Accept a session token. Invalidate it so it can no longer be used.

3. **Session management** — Maintain active sessions in memory. Provide a function to check whether a given token is currently valid. Sessions should expire after 30 minutes of inactivity.

4. **Password hashing** — Passwords must be hashed using a standard algorithm (bcrypt or SHA-256 with salt are both acceptable). Do not store or compare plaintext passwords.

5. **Tests** — Write unit tests for all four areas above (login, logout, session validity, expiry). Tests must be runnable with `npm test`.

6. **Documentation** — Write a README that explains what the module does, how to install dependencies, and how to use each function with a short code example.

### Acceptance criteria

- `npm test` runs and all tests pass.
- A valid login call returns a non-empty token string.
- Using that token immediately after login reports it as valid.
- Logging out with that token reports it as invalid on the next validity check.
- An expired token (simulate by backdating its timestamp) reports as invalid.
- No plaintext passwords appear anywhere in source or test code.
- README is present and covers all three operations.

### Constraints

- Node.js only. No frontend required.
- No external database. In-memory storage is correct.
- Keep the implementation small. A single module file plus a test file is sufficient. Do not build an HTTP server unless you want to — the core logic should be callable as plain functions.
- Do not add packages beyond what is strictly necessary for the requirements above.

### Done when

The project directory contains working source code, passing tests, and a README. Declare completion when `npm test` exits with code 0.

---

## Notes for reviewers

This task was chosen because it is:

- **Realistic** — authentication is a common, well-understood feature.
- **Bounded** — clear acceptance criteria with no ambiguity about what "done" means.
- **Fair to both groups** — Group A's toolkit adds process overhead; if the toolkit helps, it should show up in spec compliance and evidence quality without hurting output correctness.
- **Naturally scope-limited** — over-engineering is easy to detect (e.g. building an Express server when none was required).
