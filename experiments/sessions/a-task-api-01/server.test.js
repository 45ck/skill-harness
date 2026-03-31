/**
 * @file server.test.js
 * @spec API-001
 * Integration tests for all 6 HTTP routes. Starts the server on a random
 * ephemeral port, runs all assertions, and closes on completion.
 */

import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from './server.js';

/** @type {import('node:http').Server} */
let server;
/** @type {number} */
let port;

/** Shared token for tests that require auth. */
let aliceToken = '';
let bobToken = '';
/** Shared task id created in tests. */
let taskId = '';

// Known test credentials — hashed at rest in auth.js; plaintext only in tests.
const ALICE = { username: 'alice', password: 'T3stP@ssw0rd!Alice' };
const BOB   = { username: 'bob',   password: 'T3stP@ssw0rd!Bob' };

import http from 'node:http';

/**
 * Make an HTTP request to the local test server.
 *
 * @param {string} method
 * @param {string} path
 * @param {{ headers?: object, body?: object }} [opts]
 * @returns {Promise<{ status: number, body: any, headers: object }>}
 */
function _request(method, path, { headers = {}, body } = {}) {
  return new Promise((resolve, reject) => {
    const payload = body !== undefined ? JSON.stringify(body) : undefined;
    const reqHeaders = {
      'Content-Type': 'application/json',
      ...headers,
    };
    if (payload !== undefined) {
      reqHeaders['Content-Length'] = Buffer.byteLength(payload);
    }
    const req = http.request(
      { hostname: '127.0.0.1', port, path, method, headers: reqHeaders },
      (res) => {
        const chunks = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => {
          let parsed;
          const raw = Buffer.concat(chunks).toString('utf8');
          try { parsed = JSON.parse(raw); } catch { parsed = raw; }
          resolve({ status: res.statusCode, body: parsed, headers: res.headers });
        });
      }
    );
    req.on('error', reject);
    if (payload !== undefined) req.write(payload);
    req.end();
  });
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────

before(async () => {
  server = createServer();
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  port = server.address().port;
});

after(async () => {
  await new Promise((resolve) => server.close(resolve));
});

// ── POST /auth/login ──────────────────────────────────────────────────────────

test('POST /auth/login — valid credentials returns 200 and token', async () => {
  const res = await _request('POST', '/auth/login', { body: ALICE });
  assert.equal(res.status, 200);
  assert.ok(typeof res.body.token === 'string', 'token should be a string');
  assert.equal(res.body.token.length, 64, 'token should be 64 hex chars');
  aliceToken = res.body.token;
});

test('POST /auth/login — valid credentials for bob returns 200 and token', async () => {
  const res = await _request('POST', '/auth/login', { body: BOB });
  assert.equal(res.status, 200);
  assert.ok(typeof res.body.token === 'string');
  bobToken = res.body.token;
});

test('POST /auth/login — wrong password returns 401', async () => {
  const res = await _request('POST', '/auth/login', { body: { username: 'alice', password: 'wrong' } });
  assert.equal(res.status, 401);
  assert.ok(res.body.error, 'should have error field');
});

test('POST /auth/login — unknown user returns 401', async () => {
  const res = await _request('POST', '/auth/login', { body: { username: 'nobody', password: 'x' } });
  assert.equal(res.status, 401);
});

test('POST /auth/login — missing password returns 400', async () => {
  const res = await _request('POST', '/auth/login', { body: { username: 'alice' } });
  assert.equal(res.status, 400);
});

test('POST /auth/login — missing username returns 400', async () => {
  const res = await _request('POST', '/auth/login', { body: { password: 'x' } });
  assert.equal(res.status, 400);
});

// ── GET /tasks ────────────────────────────────────────────────────────────────

test('GET /tasks — no token returns 401', async () => {
  const res = await _request('GET', '/tasks');
  assert.equal(res.status, 401);
});

test('GET /tasks — invalid token returns 401', async () => {
  const res = await _request('GET', '/tasks', { headers: { Authorization: 'Bearer invalidtoken' } });
  assert.equal(res.status, 401);
});

test('GET /tasks — valid token returns 200 with tasks array', async () => {
  const res = await _request('GET', '/tasks', { headers: { Authorization: `Bearer ${aliceToken}` } });
  assert.equal(res.status, 200);
  assert.ok(Array.isArray(res.body.tasks), 'tasks should be an array');
});

// ── POST /tasks ───────────────────────────────────────────────────────────────

test('POST /tasks — no token returns 401', async () => {
  const res = await _request('POST', '/tasks', { body: { title: 'Test' } });
  assert.equal(res.status, 401);
});

test('POST /tasks — missing title returns 400', async () => {
  const res = await _request('POST', '/tasks', {
    headers: { Authorization: `Bearer ${aliceToken}` },
    body: { description: 'no title here' },
  });
  assert.equal(res.status, 400);
});

test('POST /tasks — valid request returns 201 with task object', async () => {
  const res = await _request('POST', '/tasks', {
    headers: { Authorization: `Bearer ${aliceToken}` },
    body: { title: 'Buy groceries', description: 'Milk, eggs, bread' },
  });
  assert.equal(res.status, 201);
  assert.ok(res.body.id, 'task should have id');
  assert.equal(res.body.title, 'Buy groceries');
  assert.equal(res.body.status, 'todo');
  assert.ok(res.body.createdAt, 'task should have createdAt');
  assert.ok(res.body.updatedAt, 'task should have updatedAt');
  taskId = res.body.id;
});

test('POST /tasks — task is visible in GET /tasks for same user', async () => {
  const res = await _request('GET', '/tasks', { headers: { Authorization: `Bearer ${aliceToken}` } });
  assert.equal(res.status, 200);
  const found = res.body.tasks.find((t) => t.id === taskId);
  assert.ok(found, 'created task should appear in list');
});

// ── User scoping ──────────────────────────────────────────────────────────────

test('GET /tasks — bob cannot see alice tasks', async () => {
  const res = await _request('GET', '/tasks', { headers: { Authorization: `Bearer ${bobToken}` } });
  assert.equal(res.status, 200);
  const aliceTask = res.body.tasks.find((t) => t.id === taskId);
  assert.equal(aliceTask, undefined, "bob should not see alice's task");
});

// ── PUT /tasks/:id ────────────────────────────────────────────────────────────

test('PUT /tasks/:id — no token returns 401', async () => {
  const res = await _request('PUT', `/tasks/${taskId}`, { body: { status: 'done' } });
  assert.equal(res.status, 401);
});

test('PUT /tasks/:id — valid update returns 200 with updated task', async () => {
  const res = await _request('PUT', `/tasks/${taskId}`, {
    headers: { Authorization: `Bearer ${aliceToken}` },
    body: { status: 'in_progress', title: 'Buy groceries updated' },
  });
  assert.equal(res.status, 200);
  assert.equal(res.body.status, 'in_progress');
  assert.equal(res.body.title, 'Buy groceries updated');
});

test('PUT /tasks/:id — invalid status returns 400', async () => {
  const res = await _request('PUT', `/tasks/${taskId}`, {
    headers: { Authorization: `Bearer ${aliceToken}` },
    body: { status: 'flying' },
  });
  assert.equal(res.status, 400);
});

test('PUT /tasks/:id — non-existent id returns 404', async () => {
  const res = await _request('PUT', '/tasks/nonexistent999', {
    headers: { Authorization: `Bearer ${aliceToken}` },
    body: { status: 'done' },
  });
  assert.equal(res.status, 404);
});

test("PUT /tasks/:id — bob cannot update alice's task (returns 404)", async () => {
  const res = await _request('PUT', `/tasks/${taskId}`, {
    headers: { Authorization: `Bearer ${bobToken}` },
    body: { status: 'done' },
  });
  assert.equal(res.status, 404);
});

// ── DELETE /tasks/:id ─────────────────────────────────────────────────────────

test('DELETE /tasks/:id — no token returns 401', async () => {
  const res = await _request('DELETE', `/tasks/${taskId}`);
  assert.equal(res.status, 401);
});

test('DELETE /tasks/:id — non-existent id returns 404', async () => {
  const res = await _request('DELETE', '/tasks/doesnotexist', {
    headers: { Authorization: `Bearer ${aliceToken}` },
  });
  assert.equal(res.status, 404);
});

test("DELETE /tasks/:id — bob cannot delete alice's task (returns 404)", async () => {
  const res = await _request('DELETE', `/tasks/${taskId}`, {
    headers: { Authorization: `Bearer ${bobToken}` },
  });
  assert.equal(res.status, 404);
});

test('DELETE /tasks/:id — alice can delete her own task, returns 204', async () => {
  const res = await _request('DELETE', `/tasks/${taskId}`, {
    headers: { Authorization: `Bearer ${aliceToken}` },
  });
  assert.equal(res.status, 204);
});

test('GET /tasks — deleted task no longer appears', async () => {
  const res = await _request('GET', '/tasks', { headers: { Authorization: `Bearer ${aliceToken}` } });
  assert.equal(res.status, 200);
  const found = res.body.tasks.find((t) => t.id === taskId);
  assert.equal(found, undefined, 'deleted task should not appear');
});

// ── POST /auth/logout ─────────────────────────────────────────────────────────

test('POST /auth/logout — no token returns 401', async () => {
  const res = await _request('POST', '/auth/logout');
  assert.equal(res.status, 401);
});

test('POST /auth/logout — valid token returns 204', async () => {
  const res = await _request('POST', '/auth/logout', {
    headers: { Authorization: `Bearer ${aliceToken}` },
  });
  assert.equal(res.status, 204);
});

test('GET /tasks — after logout, old token returns 401', async () => {
  const res = await _request('GET', '/tasks', { headers: { Authorization: `Bearer ${aliceToken}` } });
  assert.equal(res.status, 401);
});

// ── Unknown routes ────────────────────────────────────────────────────────────

test('GET /unknown — returns 404', async () => {
  const res = await _request('GET', '/unknown');
  assert.equal(res.status, 404);
});
