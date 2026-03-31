/**
 * @file auth.js
 * @spec AUTH-001
 * User authentication module. Provides login, logout, and session validation
 * using SHA-256 + per-user salt. Sessions expire after 30 minutes of inactivity.
 */

import { createHash, randomBytes } from 'node:crypto';

/** Session expiry in milliseconds (30 minutes). */
const SESSION_TTL_MS = 30 * 60 * 1000;

/**
 * Hash a password with a given salt using SHA-256.
 * @param {string} password - Plaintext password.
 * @param {string} salt - Hex-encoded salt string.
 * @returns {string} Hex-encoded SHA-256 digest.
 */
function hashPassword(password, salt) {
  return createHash('sha256').update(salt + password).digest('hex');
}

/**
 * Static in-memory user store.
 * Passwords are pre-hashed; plaintext never appears here.
 * Users: alice / T3stP@ssw0rd!Alice  and  bob / T3stP@ssw0rd!Bob
 * @implements AUTH-001-R4
 * @evidence E0
 * @type {Map<string, { hash: string, salt: string }>}
 */
const USERS = (() => {
  const aliceSalt = 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4';
  const bobSalt   = 'f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3';
  return new Map([
    ['alice', { hash: hashPassword('T3stP@ssw0rd!Alice', aliceSalt), salt: aliceSalt }],
    ['bob',   { hash: hashPassword('T3stP@ssw0rd!Bob',   bobSalt),   salt: bobSalt   }],
  ]);
})();

/**
 * In-memory session store: token → { username, lastUsed }.
 * @type {Map<string, { username: string, lastUsed: number }>}
 */
const SESSIONS = new Map();

/**
 * Resolve the username for a token without refreshing the timestamp.
 * Used internally.
 * @param {string} token
 * @returns {string|null}
 */
function usernameForToken(token) {
  const session = SESSIONS.get(token);
  if (!session) return null;
  if (Date.now() - session.lastUsed > SESSION_TTL_MS) {
    SESSIONS.delete(token);
    return null;
  }
  return session.username;
}

/**
 * Authenticate a user and create a new session.
 *
 * @spec AUTH-001
 * @implements AUTH-001-R1
 * @evidence E0
 * @param {string} username
 * @param {string} password - Plaintext password (never stored).
 * @returns {{ token: string }} Session token.
 * @throws {Error} On invalid credentials.
 */
export function login(username, password) {
  const user = USERS.get(username);
  if (!user) {
    throw new Error('Invalid credentials');
  }
  const attempt = hashPassword(password, user.salt);
  if (attempt !== user.hash) {
    throw new Error('Invalid credentials');
  }
  const token = randomBytes(32).toString('hex');
  SESSIONS.set(token, { username, lastUsed: Date.now() });
  return { token };
}

/**
 * Invalidate an existing session. No-op if token is unknown.
 *
 * @spec AUTH-001
 * @implements AUTH-001-R2
 * @evidence E0
 * @param {string} token
 * @returns {void}
 */
export function logout(token) {
  SESSIONS.delete(token);
}

/**
 * Check whether a session token is currently valid.
 * Refreshes the last-used timestamp on success (sliding expiry).
 *
 * @spec AUTH-001
 * @implements AUTH-001-R3
 * @evidence E0
 * @param {string} token
 * @returns {boolean}
 */
export function isValidSession(token) {
  const session = SESSIONS.get(token);
  if (!session) return false;
  if (Date.now() - session.lastUsed > SESSION_TTL_MS) {
    SESSIONS.delete(token);
    return false;
  }
  session.lastUsed = Date.now();
  return true;
}

/**
 * Retrieve the username associated with a valid token.
 * Also refreshes the session timestamp.
 *
 * @spec AUTH-001
 * @implements AUTH-001-R3
 * @evidence E0
 * @param {string} token
 * @returns {string} Username.
 * @throws {Error} If the token is invalid or expired.
 */
export function requireUsername(token) {
  const session = SESSIONS.get(token);
  if (!session) throw new Error('Invalid or expired session');
  if (Date.now() - session.lastUsed > SESSION_TTL_MS) {
    SESSIONS.delete(token);
    throw new Error('Session expired');
  }
  session.lastUsed = Date.now();
  return session.username;
}
