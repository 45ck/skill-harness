'use strict';

/**
 * auth.test.js — Unit tests for the auth module.
 *
 * Run with: npm test
 * Uses Node.js built-in test runner (node:test) and assert (node:assert/strict).
 */

import { describe, it, beforeEach } from 'node:test';
import assert from 'node:assert/strict';

import {
  login,
  logout,
  isValidSession,
  getSessionUser,
  _setLastActive,
  _clearSessions,
} from './auth.js';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const SESSION_TTL_MS = 30 * 60 * 1000;

// Reset session state before every test so tests are fully independent.
beforeEach(() => {
  _clearSessions();
});

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

describe('login', () => {
  it('returns a non-empty token string for valid credentials', () => {
    const result = login('alice', 'correct-horse-battery-staple');
    assert.equal(typeof result.token, 'string');
    assert.ok(result.token.length > 0, 'token must be non-empty');
  });

  it('returns the username in the result object', () => {
    const result = login('alice', 'correct-horse-battery-staple');
    assert.equal(result.username, 'alice');
  });

  it('returns a 64-character hex token (32 random bytes)', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    assert.match(token, /^[0-9a-f]{64}$/);
  });

  it('generates a different token on each login call', () => {
    const { token: t1 } = login('alice', 'correct-horse-battery-staple');
    const { token: t2 } = login('alice', 'correct-horse-battery-staple');
    assert.notEqual(t1, t2);
  });

  it('works for a second user (bob)', () => {
    const { token } = login('bob', 'hunter2');
    assert.match(token, /^[0-9a-f]{64}$/);
  });

  it('throws on an unknown username', () => {
    assert.throws(
      () => login('nobody', 'somepassword'),
      { message: 'Invalid credentials.' },
    );
  });

  it('throws on a wrong password for a known user', () => {
    assert.throws(
      () => login('alice', 'wrong-password'),
      { message: 'Invalid credentials.' },
    );
  });

  it('throws on an empty password', () => {
    assert.throws(
      () => login('alice', ''),
      { message: 'Invalid credentials.' },
    );
  });

  it('throws on an empty username', () => {
    assert.throws(
      () => login('', 'correct-horse-battery-staple'),
      { message: 'Invalid credentials.' },
    );
  });
});

// ---------------------------------------------------------------------------
// Session validity
// ---------------------------------------------------------------------------

describe('isValidSession', () => {
  it('reports a freshly issued token as valid', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    assert.equal(isValidSession(token), true);
  });

  it('reports an unknown token as invalid', () => {
    assert.equal(isValidSession('0'.repeat(64)), false);
  });

  it('reports an empty string as invalid', () => {
    assert.equal(isValidSession(''), false);
  });

  it('reports tokens from different users as independently valid', () => {
    const { token: ta } = login('alice', 'correct-horse-battery-staple');
    const { token: tb } = login('bob', 'hunter2');
    assert.equal(isValidSession(ta), true);
    assert.equal(isValidSession(tb), true);
  });

  it('getSessionUser returns the username for a valid token', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    assert.equal(getSessionUser(token), 'alice');
  });

  it('getSessionUser returns null for an unknown token', () => {
    assert.equal(getSessionUser('deadbeef'.repeat(8)), null);
  });
});

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

describe('logout', () => {
  it('invalidates the token immediately after logout', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    assert.equal(isValidSession(token), true, 'precondition: token is valid before logout');
    logout(token);
    assert.equal(isValidSession(token), false, 'token must be invalid after logout');
  });

  it('getSessionUser returns null after logout', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    logout(token);
    assert.equal(getSessionUser(token), null);
  });

  it('does not throw when called with an unknown token', () => {
    assert.doesNotThrow(() => logout('0'.repeat(64)));
  });

  it('does not affect other active sessions', () => {
    const { token: ta } = login('alice', 'correct-horse-battery-staple');
    const { token: tb } = login('bob', 'hunter2');
    logout(ta);
    assert.equal(isValidSession(ta), false);
    assert.equal(isValidSession(tb), true, "bob's session must still be valid");
  });

  it('is idempotent — calling logout twice does not throw', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    logout(token);
    assert.doesNotThrow(() => logout(token));
  });
});

// ---------------------------------------------------------------------------
// Session expiry
// ---------------------------------------------------------------------------

describe('session expiry', () => {
  it('reports an expired token as invalid (backdated timestamp)', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    // Simulate the session being 31 minutes old.
    _setLastActive(token, Date.now() - SESSION_TTL_MS - 60_000);
    assert.equal(isValidSession(token), false);
  });

  it('getSessionUser returns null for an expired token', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    _setLastActive(token, Date.now() - SESSION_TTL_MS - 1);
    assert.equal(getSessionUser(token), null);
  });

  it('a token exactly at the TTL boundary is still invalid', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    // lastActive exactly SESSION_TTL_MS ago means (now - lastActive) === TTL, which is > TTL: false.
    _setLastActive(token, Date.now() - SESSION_TTL_MS);
    assert.equal(isValidSession(token), false);
  });

  it('a token one millisecond inside the TTL window is still valid', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    _setLastActive(token, Date.now() - SESSION_TTL_MS + 100);
    assert.equal(isValidSession(token), true);
  });

  it('an expired token is removed from the store (subsequent checks also return false)', () => {
    const { token } = login('alice', 'correct-horse-battery-staple');
    _setLastActive(token, Date.now() - SESSION_TTL_MS - 1);
    isValidSession(token); // trigger eviction
    assert.equal(isValidSession(token), false, 'must remain invalid after eviction');
  });
});
