/**
 * @module auth
 * @spec AUTH-001
 * @implements login, logout, session management, password hashing
 * @evidence E0
 *
 * In-memory user authentication module.
 * Provides login, logout, and session validity checks.
 * Sessions expire after 30 minutes of inactivity.
 * Passwords are stored and compared as SHA-256 hashes with a per-user salt.
 */

import { createHash, randomBytes } from 'node:crypto';

// ---------------------------------------------------------------------------
// Password hashing
// ---------------------------------------------------------------------------

/**
 * Hash a plaintext password using SHA-256 with the provided salt.
 *
 * @spec AUTH-001
 * @implements password hashing
 * @evidence E0
 *
 * @param {string} password - Plaintext password (never stored).
 * @param {string} salt     - Hex-encoded salt string.
 * @returns {string} Hex-encoded SHA-256 digest of salt+password.
 */
export function hashPassword(password, salt) {
  return createHash('sha256').update(salt + password).digest('hex');
}

// ---------------------------------------------------------------------------
// Static user store
// ---------------------------------------------------------------------------

/**
 * In-memory user store. Each entry contains a hex salt and the SHA-256 hash
 * of (salt + plaintext_password). No plaintext passwords are stored.
 *
 * @spec AUTH-001
 * @implements login
 * @evidence E0
 *
 * Pre-seeded with two demo accounts:
 *   alice / password123
 *   bob   / hunter2
 */
const SALT_ALICE = 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4';
const SALT_BOB   = 'b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5';

const USER_STORE = new Map([
  ['alice', { salt: SALT_ALICE, hash: hashPassword('password123', SALT_ALICE) }],
  ['bob',   { salt: SALT_BOB,   hash: hashPassword('hunter2',     SALT_BOB)   }],
]);

// ---------------------------------------------------------------------------
// Session store
// ---------------------------------------------------------------------------

/**
 * Active sessions map: token → { username, lastActivity (ms timestamp) }.
 *
 * @spec AUTH-001
 * @implements session management
 * @evidence E0
 */
const SESSION_STORE = new Map();

/** Session inactivity timeout in milliseconds (30 minutes). */
export const SESSION_TIMEOUT_MS = 30 * 60 * 1000;

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Authenticate a user and create a new session.
 *
 * @spec AUTH-001
 * @implements login
 * @evidence E0
 *
 * @param {string} username
 * @param {string} password - Plaintext password supplied by the caller.
 * @returns {{ token: string }} Object containing the session token on success.
 * @throws {Error} If the username does not exist or the password is incorrect.
 */
export function login(username, password) {
  const user = USER_STORE.get(username);
  if (!user) {
    throw new Error('Invalid credentials');
  }
  const candidate = hashPassword(password, user.salt);
  if (candidate !== user.hash) {
    throw new Error('Invalid credentials');
  }
  const token = randomBytes(32).toString('hex');
  SESSION_STORE.set(token, { username, lastActivity: Date.now() });
  return { token };
}

/**
 * Invalidate an active session token.
 *
 * @spec AUTH-001
 * @implements logout
 * @evidence E0
 *
 * @param {string} token - Session token to invalidate.
 * @returns {void} Silently succeeds even if the token was already invalid.
 */
export function logout(token) {
  SESSION_STORE.delete(token);
}

/**
 * Check whether a session token is currently valid.
 * Resets the inactivity timer on a successful check.
 *
 * @spec AUTH-001
 * @implements session management, session expiry
 * @evidence E0
 *
 * @param {string} token - Session token to validate.
 * @returns {boolean} True if the token exists and has not expired.
 */
export function isValidSession(token) {
  const session = SESSION_STORE.get(token);
  if (!session) return false;

  const now = Date.now();
  if (now - session.lastActivity > SESSION_TIMEOUT_MS) {
    SESSION_STORE.delete(token);
    return false;
  }

  session.lastActivity = now;
  return true;
}

/**
 * Expose the session store for testing purposes only (backdating timestamps).
 * Not part of the public API.
 *
 * @internal
 */
export { SESSION_STORE };
