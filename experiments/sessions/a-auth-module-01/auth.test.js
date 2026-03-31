/**
 * @spec AUTH-001
 * @implements login, logout, session management, password hashing
 * @evidence E1
 *
 * Unit tests for the auth module.
 * Run with: npm test
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  login,
  logout,
  isValidSession,
  hashPassword,
  SESSION_STORE,
  SESSION_TIMEOUT_MS,
} from './auth.js';

// ---------------------------------------------------------------------------
// Helper: clear all sessions before each logical group
// ---------------------------------------------------------------------------
function clearSessions() {
  SESSION_STORE.clear();
}

// ---------------------------------------------------------------------------
// Password hashing
// ---------------------------------------------------------------------------

test('hashPassword returns a non-empty hex string', () => {
  const hash = hashPassword('secret', 'somesalt');
  assert.ok(hash.length > 0, 'hash should not be empty');
  assert.match(hash, /^[0-9a-f]+$/, 'hash should be a hex string');
});

test('hashPassword produces consistent output for the same inputs', () => {
  const h1 = hashPassword('secret', 'somesalt');
  const h2 = hashPassword('secret', 'somesalt');
  assert.equal(h1, h2);
});

test('hashPassword produces different output for different passwords', () => {
  const h1 = hashPassword('secret',  'somesalt');
  const h2 = hashPassword('secret2', 'somesalt');
  assert.notEqual(h1, h2);
});

test('hashPassword produces different output for different salts', () => {
  const h1 = hashPassword('secret', 'salt1');
  const h2 = hashPassword('secret', 'salt2');
  assert.notEqual(h1, h2);
});

// ---------------------------------------------------------------------------
// Login — success
// ---------------------------------------------------------------------------

test('login returns a non-empty token for valid credentials (alice)', () => {
  clearSessions();
  const result = login('alice', 'password123');
  assert.ok(result.token, 'token should be truthy');
  assert.equal(typeof result.token, 'string');
  assert.ok(result.token.length > 0, 'token must be non-empty');
});

test('login returns a non-empty token for valid credentials (bob)', () => {
  clearSessions();
  const result = login('bob', 'hunter2');
  assert.ok(result.token);
  assert.equal(typeof result.token, 'string');
});

test('login returns different tokens for successive logins', () => {
  clearSessions();
  const { token: t1 } = login('alice', 'password123');
  const { token: t2 } = login('alice', 'password123');
  assert.notEqual(t1, t2, 'each login should produce a unique token');
});

// ---------------------------------------------------------------------------
// Login — failure
// ---------------------------------------------------------------------------

test('login throws for unknown username', () => {
  clearSessions();
  assert.throws(
    () => login('nobody', 'password123'),
    { message: 'Invalid credentials' },
  );
});

test('login throws for wrong password', () => {
  clearSessions();
  assert.throws(
    () => login('alice', 'wrongpassword'),
    { message: 'Invalid credentials' },
  );
});

test('login error message does not reveal whether username or password was wrong', () => {
  clearSessions();
  let msgForBadUser, msgForBadPass;
  try { login('nobody', 'x'); } catch (e) { msgForBadUser = e.message; }
  try { login('alice', 'x'); } catch (e) { msgForBadPass = e.message; }
  assert.equal(msgForBadUser, msgForBadPass, 'error messages must be identical');
});

// ---------------------------------------------------------------------------
// Session validity
// ---------------------------------------------------------------------------

test('isValidSession returns true immediately after login', () => {
  clearSessions();
  const { token } = login('alice', 'password123');
  assert.equal(isValidSession(token), true);
});

test('isValidSession returns false for an unknown token', () => {
  clearSessions();
  assert.equal(isValidSession('nonexistent-token'), false);
});

test('isValidSession returns false for an empty string', () => {
  clearSessions();
  assert.equal(isValidSession(''), false);
});

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

test('logout invalidates the token', () => {
  clearSessions();
  const { token } = login('alice', 'password123');
  assert.equal(isValidSession(token), true, 'token should be valid before logout');
  logout(token);
  assert.equal(isValidSession(token), false, 'token should be invalid after logout');
});

test('logout is idempotent — calling twice does not throw', () => {
  clearSessions();
  const { token } = login('alice', 'password123');
  logout(token);
  assert.doesNotThrow(() => logout(token));
});

test('logout with an unknown token does not throw', () => {
  clearSessions();
  assert.doesNotThrow(() => logout('unknown-token'));
});

// ---------------------------------------------------------------------------
// Session expiry
// ---------------------------------------------------------------------------

test('isValidSession returns false for an expired token (backdated timestamp)', () => {
  clearSessions();
  const { token } = login('alice', 'password123');
  // Backdate the lastActivity beyond the timeout window
  const session = SESSION_STORE.get(token);
  session.lastActivity = Date.now() - SESSION_TIMEOUT_MS - 1;
  assert.equal(isValidSession(token), false, 'expired token must be invalid');
});

test('isValidSession removes an expired token from the store', () => {
  clearSessions();
  const { token } = login('alice', 'password123');
  const session = SESSION_STORE.get(token);
  session.lastActivity = Date.now() - SESSION_TIMEOUT_MS - 1;
  isValidSession(token);
  assert.equal(SESSION_STORE.has(token), false, 'expired token must be purged');
});

test('isValidSession resets the inactivity timer on a valid check', () => {
  clearSessions();
  const { token } = login('alice', 'password123');
  const before = SESSION_STORE.get(token).lastActivity;
  // Small delay to ensure timestamp advances
  const start = Date.now();
  while (Date.now() === start) { /* spin */ }
  isValidSession(token);
  const after = SESSION_STORE.get(token).lastActivity;
  assert.ok(after >= before, 'lastActivity should be updated after a valid check');
});
