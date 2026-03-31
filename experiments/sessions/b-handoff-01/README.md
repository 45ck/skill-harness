# b-auth-module-01 — User Authentication Module

A lightweight, self-contained user authentication module for Node.js. It
handles login, logout, and in-memory session management with no external
dependencies and no database requirement.

## Features

- Password hashing with SHA-256 + per-user random salt (no plaintext passwords stored).
- Session tokens: 32 cryptographically random bytes, hex-encoded.
- Sliding 30-minute session expiry (inactivity timeout).
- Unit tests for all four areas using the Node.js built-in test runner.

## Requirements

- Node.js >= 18 (uses `node:crypto`, `node:test`, and ES modules).
- No npm packages to install.

## Installation

```bash
# No dependencies — nothing to install.
# Clone or copy the files into your project directory, then:
npm test   # verify everything passes
```

## Usage

All functions are named exports from `auth.js`.

### login(username, password)

Validates credentials against the in-memory user store. Returns an object
with `{ token, username }` on success, or throws an `Error` with the message
`"Invalid credentials."` on failure.

```js
import { login } from './auth.js';

try {
  const { token, username } = login('alice', 'correct-horse-battery-staple');
  console.log(`Logged in as ${username}. Token: ${token}`);
} catch (err) {
  console.error(err.message); // "Invalid credentials."
}
```

### logout(token)

Immediately invalidates the given session token. Subsequent calls to
`isValidSession` with that token will return `false`. Calling `logout` with an
unknown or already-expired token is safe and returns silently.

```js
import { logout } from './auth.js';

logout(token); // session is now gone
```

### isValidSession(token)

Returns `true` if the token exists and was last used within the past 30
minutes; `false` otherwise. A successful check refreshes the sliding expiry
window.

```js
import { isValidSession } from './auth.js';

if (isValidSession(token)) {
  // proceed with the authenticated request
} else {
  // redirect to login
}
```

### getSessionUser(token)

Returns the username associated with a valid session token, or `null` if the
token is invalid or expired.

```js
import { getSessionUser } from './auth.js';

const username = getSessionUser(token);
if (username) {
  console.log(`Request made by: ${username}`);
}
```

## Running tests

```bash
npm test
```

The test suite covers:

- Login: valid credentials, wrong password, unknown user, empty inputs, token uniqueness.
- Session validity: fresh token valid, unknown token invalid, per-user independence.
- Logout: immediate invalidation, idempotency, no cross-session side effects.
- Expiry: backdated timestamp triggers invalid result, boundary conditions, store cleanup.

## Design notes

- The user store is initialised once at module load from a hardcoded
  `{ username: plaintext }` map that is immediately hashed. In a real
  application you would store pre-computed `{ salt, hash }` pairs and never
  include plaintext passwords in source code.
- Sessions are stored in a `Map` inside the module. They survive for the
  lifetime of the Node.js process and are not persisted to disk.
- The `_setLastActive` and `_clearSessions` exports are test helpers only
  (prefixed with `_` to signal internal use). Do not call them in production
  code.
