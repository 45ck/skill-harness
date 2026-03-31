/**
 * auth.js — User authentication module
 *
 * Provides login, logout, and session validation.
 * Passwords are stored as SHA-256 hashes with a per-user salt.
 * Sessions expire after 30 minutes of inactivity.
 */

import { createHash, randomBytes } from 'node:crypto';

// Session lifetime: 30 minutes in milliseconds
const SESSION_TTL_MS = 30 * 60 * 1000;

/**
 * Hash a password with a given salt using SHA-256.
 * @param {string} password
 * @param {string} salt - hex string
 * @returns {string} hex digest
 */
function hashPassword(password, salt) {
  return createHash('sha256').update(salt + password).digest('hex');
}

// Static user store: username → { salt, passwordHash }
// Passwords are never stored in plaintext.
const SALT_ALICE = 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4';
const SALT_BOB   = 'b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5';

const USER_STORE = {
  alice: {
    salt: SALT_ALICE,
    passwordHash: hashPassword('alicepassword', SALT_ALICE),
  },
  bob: {
    salt: SALT_BOB,
    passwordHash: hashPassword('bobpassword', SALT_BOB),
  },
};

/**
 * Active sessions: token → { username, lastActivity }
 * @type {Map<string, { username: string, lastActivity: number }>}
 */
const sessions = new Map();

/**
 * Remove sessions that have exceeded the inactivity TTL.
 */
function pruneExpiredSessions() {
  const now = Date.now();
  for (const [token, session] of sessions) {
    if (now - session.lastActivity > SESSION_TTL_MS) {
      sessions.delete(token);
    }
  }
}

/**
 * Authenticate a user with username and password.
 * @param {string} username
 * @param {string} password
 * @returns {{ token: string }}
 * @throws {Error} if credentials are invalid
 */
export function login(username, password) {
  pruneExpiredSessions();

  const user = USER_STORE[username];
  if (!user) {
    throw new Error('Invalid credentials');
  }

  const hash = hashPassword(password, user.salt);
  if (hash !== user.passwordHash) {
    throw new Error('Invalid credentials');
  }

  // Generate a cryptographically random token
  const token = randomBytes(32).toString('hex');
  sessions.set(token, { username, lastActivity: Date.now() });

  return { token };
}

/**
 * Invalidate a session token.
 * @param {string} token
 */
export function logout(token) {
  sessions.delete(token);
}

/**
 * Check whether a token represents a valid, non-expired session.
 * If valid, refreshes the lastActivity timestamp.
 * @param {string} token
 * @returns {boolean}
 */
export function isValidSession(token) {
  pruneExpiredSessions();

  const session = sessions.get(token);
  if (!session) return false;

  // Refresh activity timestamp
  session.lastActivity = Date.now();
  return true;
}

/**
 * Get the username associated with a valid session token.
 * @param {string} token
 * @returns {string}
 * @throws {Error} if session is invalid or expired
 */
export function getUsernameFromToken(token) {
  if (!isValidSession(token)) {
    throw new Error('Invalid or expired session');
  }
  return sessions.get(token).username;
}
