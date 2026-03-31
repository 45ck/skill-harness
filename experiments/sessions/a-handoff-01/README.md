# auth-module

An in-memory user authentication module for Node.js. Handles login, logout, and session management. No external database required.

## What it does

- Validates credentials against a static in-memory user store.
- Issues a cryptographically random session token on successful login.
- Tracks active sessions in memory with a 30-minute inactivity timeout.
- Hashes passwords using SHA-256 with a per-user salt (via Node's built-in `node:crypto`). No plaintext passwords are stored or compared.

## Requirements

- Node.js >= 18 (uses built-in `node:crypto` and `node:test`)
- No additional npm packages needed

## Install dependencies

```bash
npm install
```

## Usage

```js
import { login, logout, isValidSession } from './auth.js';
```

### login(username, password)

Validates credentials and returns a session token on success. Throws an `Error` with message `"Invalid credentials"` on failure.

```js
try {
  const { token } = login('alice', 'password123');
  console.log('Logged in, token:', token);
} catch (err) {
  console.error('Login failed:', err.message);
}
```

### logout(token)

Invalidates the session token. Silently succeeds even if the token is already invalid or unknown.

```js
logout(token);
console.log('Session ended.');
```

### isValidSession(token)

Returns `true` if the token exists and has not expired. Resets the 30-minute inactivity timer on a successful check. Returns `false` for unknown or expired tokens.

```js
if (isValidSession(token)) {
  console.log('Session is active.');
} else {
  console.log('Session has expired or is invalid.');
}
```

## Running tests

```bash
npm test
```

All 19 tests cover password hashing, login success/failure, logout, session validity, and session expiry.

## Pre-seeded demo users

| Username | Password      |
|----------|---------------|
| alice    | password123   |
| bob      | hunter2       |

These are for demonstration only. In a real application, replace the `USER_STORE` map with your own user records.

## Security notes

- Passwords are never stored or transmitted in plaintext. Only `SHA-256(salt + password)` is stored.
- Session tokens are 32 random bytes (64 hex characters) generated with `node:crypto`'s `randomBytes`.
- This module is intentionally minimal. It does not provide HTTP endpoints, rate limiting, or persistent storage.
