'use strict';

/**
 * auth.js — User authentication module
 *
 * Provides login, logout, and session management backed entirely by
 * in-memory storage. No external dependencies are required; everything
 * relies on the Node.js standard library.
 *
 * Password hashing: SHA-256 with a per-user salt (hex-encoded).
 * Session tokens: 32 random bytes, hex-encoded (64 chars).
 * Session expiry: 30 minutes of inactivity (sliding window).
 */

import { createHash, randomBytes } from 'node:crypto';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const SESSION_TTL_MS = 30 * 60 * 1000; // 30 minutes
const RESET_TOKEN_TTL_MS = 15 * 60 * 1000; // 15 minutes

// ---------------------------------------------------------------------------
// Password utilities
// ---------------------------------------------------------------------------

/**
 * Hash a plaintext password with the provided salt using SHA-256.
 *
 * @param {string} password - Plaintext password (never stored).
 * @param {string} salt     - Hex-encoded salt string.
 * @returns {string}        - Hex-encoded digest of salt+password.
 */
function hashPassword(password, salt) {
  return createHash('sha256')
    .update(salt + password)
    .digest('hex');
}

/**
 * Generate a random hex salt (16 bytes → 32 hex chars).
 *
 * @returns {string}
 */
function generateSalt() {
  return randomBytes(16).toString('hex');
}

// ---------------------------------------------------------------------------
// In-memory user store
//
// Each entry: { salt: string, hash: string }
// Passwords are NEVER stored; only the salt and the hash of salt+password.
// ---------------------------------------------------------------------------

/**
 * Build the in-memory user store from a plain { username: password } map.
 * Intended for module initialisation only.
 *
 * @param {{ [username: string]: string }} plaintextMap
 * @returns {{ [username: string]: { salt: string, hash: string } }}
 */
function buildUserStore(plaintextMap) {
  const store = {};
  for (const [username, password] of Object.entries(plaintextMap)) {
    const salt = generateSalt();
    store[username] = { salt, hash: hashPassword(password, salt) };
  }
  return store;
}

// Default users. In real usage you would derive these offline and hard-code
// the resulting { salt, hash } objects — never the plaintext passwords.
const DEFAULT_USERS = {
  alice: 'correct-horse-battery-staple',
  bob:   'hunter2',
};

const USER_STORE = buildUserStore(DEFAULT_USERS);

// ---------------------------------------------------------------------------
// Session store
//
// Map<token, { username: string, lastActive: number }>
// ---------------------------------------------------------------------------

const sessions = new Map();

// ---------------------------------------------------------------------------
// Reset token store
//
// Map<token, { username: string, createdAt: number }>
// Each token is single-use and expires after RESET_TOKEN_TTL_MS.
// ---------------------------------------------------------------------------

const resetTokens = new Map();

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Attempt to log in with the supplied credentials.
 *
 * @param {string} username
 * @param {string} password - Plaintext password (compared by hashing, never stored).
 * @returns {{ token: string, username: string }}
 * @throws {Error} If credentials are invalid.
 */
function login(username, password) {
  const user = USER_STORE[username];
  if (!user) {
    throw new Error('Invalid credentials.');
  }

  const candidate = hashPassword(password, user.salt);
  if (candidate !== user.hash) {
    throw new Error('Invalid credentials.');
  }

  const token = randomBytes(32).toString('hex');
  sessions.set(token, { username, lastActive: Date.now() });

  return { token, username };
}

/**
 * Invalidate an active session token.
 *
 * If the token does not exist (already expired or was never valid) this
 * function returns silently — callers do not need to handle that case.
 *
 * @param {string} token
 */
function logout(token) {
  sessions.delete(token);
}

/**
 * Check whether a session token is currently valid.
 *
 * A token is valid if it exists in the session store AND its last-active
 * timestamp is within the 30-minute TTL window. A valid check refreshes the
 * last-active timestamp (sliding expiry).
 *
 * @param {string} token
 * @returns {boolean}
 */
function isValidSession(token) {
  const session = sessions.get(token);
  if (!session) return false;

  const now = Date.now();
  if (now - session.lastActive >= SESSION_TTL_MS) {
    // Expired — clean up eagerly.
    sessions.delete(token);
    return false;
  }

  // Refresh sliding window.
  session.lastActive = now;
  return true;
}

/**
 * Return the username associated with a valid session token, or null if the
 * token is invalid or expired.
 *
 * @param {string} token
 * @returns {string|null}
 */
function getSessionUser(token) {
  if (!isValidSession(token)) return null;
  return sessions.get(token)?.username ?? null;
}

/**
 * Request a password reset for the given username.
 *
 * Generates a one-time, time-limited reset token and stores it. The caller
 * is responsible for delivering this token to the user (e.g. via email).
 *
 * @param {string} username
 * @returns {string} A hex-encoded reset token (64 chars).
 * @throws {Error} If the username is not found in the user store.
 */
function requestPasswordReset(username) {
  if (!USER_STORE[username]) {
    throw new Error('Unknown username.');
  }

  const token = randomBytes(32).toString('hex');
  resetTokens.set(token, { username, createdAt: Date.now() });
  return token;
}

/**
 * Complete a password reset using the provided reset token.
 *
 * Validates the token (existence, expiry, single-use), updates the user's
 * stored password hash, invalidates all existing sessions for that user, and
 * consumes the reset token so it cannot be reused.
 *
 * @param {string} resetToken  - The token returned by requestPasswordReset.
 * @param {string} newPassword - The new plaintext password to set.
 * @throws {Error} If the token is invalid or has expired.
 */
function resetPassword(resetToken, newPassword) {
  const entry = resetTokens.get(resetToken);
  if (!entry) {
    throw new Error('Invalid or expired reset token.');
  }

  if (Date.now() - entry.createdAt >= RESET_TOKEN_TTL_MS) {
    resetTokens.delete(resetToken);
    throw new Error('Invalid or expired reset token.');
  }

  // Consume the token immediately (single-use).
  resetTokens.delete(resetToken);

  // Update the user's password.
  const { username } = entry;
  const salt = generateSalt();
  USER_STORE[username] = { salt, hash: hashPassword(newPassword, salt) };

  // Invalidate all existing sessions for this user.
  for (const [token, session] of sessions) {
    if (session.username === username) {
      sessions.delete(token);
    }
  }
}

/**
 * Backdoor for testing: directly manipulate the lastActive timestamp of a
 * session so expiry scenarios can be simulated without real clock delays.
 *
 * @param {string} token
 * @param {number} lastActive - Epoch ms to set.
 */
function _setLastActive(token, lastActive) {
  const session = sessions.get(token);
  if (session) session.lastActive = lastActive;
}

/**
 * Backdoor for testing: set the createdAt timestamp of a reset token so
 * expiry scenarios can be simulated without real clock delays.
 *
 * @param {string} token
 * @param {number} createdAt - Epoch ms to set.
 */
function _setResetTokenCreatedAt(token, createdAt) {
  const entry = resetTokens.get(token);
  if (entry) entry.createdAt = createdAt;
}

/**
 * Clear all sessions. Useful for test isolation.
 */
function _clearSessions() {
  sessions.clear();
}

/**
 * Clear all reset tokens. Useful for test isolation.
 */
function _clearResetTokens() {
  resetTokens.clear();
}

/**
 * Reset the user store to the default set of users with freshly generated
 * salts. Useful for test isolation when a test mutates user passwords.
 */
function _resetUserStore() {
  const fresh = buildUserStore(DEFAULT_USERS);
  for (const key of Object.keys(USER_STORE)) {
    delete USER_STORE[key];
  }
  Object.assign(USER_STORE, fresh);
}

export {
  login,
  logout,
  isValidSession,
  getSessionUser,
  requestPasswordReset,
  resetPassword,
  // Test helpers (prefixed with _ to signal internal use)
  _setLastActive,
  _setResetTokenCreatedAt,
  _clearSessions,
  _clearResetTokens,
  _resetUserStore,
};
