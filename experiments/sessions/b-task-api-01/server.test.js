/**
 * server.test.js — Integration tests for Task Management REST API
 *
 * Uses node:test + node:assert/strict.
 * Starts the server on a random port, runs all tests, then closes.
 */

import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from './server.js';

let server;
let baseUrl;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Make an HTTP request and return { status, body }.
 * @param {string} method
 * @param {string} path
 * @param {object|null} body
 * @param {string|null} token
 * @returns {Promise<{ status: number, body: any }>}
 */
async function request(method, path, body = null, token = null) {
  const { default: http } = await import('node:http');

  return new Promise((resolve, reject) => {
    const url = new URL(path, baseUrl);
    const payload = body !== null ? JSON.stringify(body) : null;

    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;
    if (payload) headers['Content-Length'] = Buffer.byteLength(payload);

    const req = http.request(
      { hostname: url.hostname, port: url.port, path: url.pathname, method, headers },
      (res) => {
        let raw = '';
        res.on('data', (chunk) => { raw += chunk; });
        res.on('end', () => {
          let parsed;
          try { parsed = raw ? JSON.parse(raw) : null; } catch { parsed = raw; }
          resolve({ status: res.statusCode, body: parsed });
        });
      }
    );

    req.on('error', reject);
    if (payload) req.write(payload);
    req.end();
  });
}

/** Login as alice and return her token. */
async function loginAlice() {
  const { body } = await request('POST', '/auth/login', { username: 'alice', password: 'alicepassword' });
  return body.token;
}

/** Login as bob and return his token. */
async function loginBob() {
  const { body } = await request('POST', '/auth/login', { username: 'bob', password: 'bobpassword' });
  return body.token;
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

before(async () => {
  server = createServer();
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  const { port } = server.address();
  baseUrl = `http://127.0.0.1:${port}`;
});

after(async () => {
  await new Promise((resolve, reject) =>
    server.close((err) => (err ? reject(err) : resolve()))
  );
});

// ---------------------------------------------------------------------------
// POST /auth/login
// ---------------------------------------------------------------------------

test('POST /auth/login — valid credentials returns token', async () => {
  const { status, body } = await request('POST', '/auth/login', { username: 'alice', password: 'alicepassword' });
  assert.equal(status, 200);
  assert.ok(typeof body.token === 'string', 'token should be a string');
  assert.ok(body.token.length > 0, 'token should not be empty');
});

test('POST /auth/login — bob can log in too', async () => {
  const { status, body } = await request('POST', '/auth/login', { username: 'bob', password: 'bobpassword' });
  assert.equal(status, 200);
  assert.ok(typeof body.token === 'string');
});

test('POST /auth/login — wrong password returns 401', async () => {
  const { status } = await request('POST', '/auth/login', { username: 'alice', password: 'wrongpassword' });
  assert.equal(status, 401);
});

test('POST /auth/login — unknown user returns 401', async () => {
  const { status } = await request('POST', '/auth/login', { username: 'eve', password: 'anything' });
  assert.equal(status, 401);
});

test('POST /auth/login — missing credentials returns 400', async () => {
  const { status } = await request('POST', '/auth/login', { username: 'alice' });
  assert.equal(status, 400);
});

// ---------------------------------------------------------------------------
// POST /auth/logout
// ---------------------------------------------------------------------------

test('POST /auth/logout — valid token returns 204', async () => {
  const token = await loginAlice();
  const { status } = await request('POST', '/auth/logout', null, token);
  assert.equal(status, 204);
});

test('POST /auth/logout — after logout, token is invalid', async () => {
  const token = await loginAlice();
  await request('POST', '/auth/logout', null, token);
  // Now try to use the token — should get 401
  const { status } = await request('GET', '/tasks', null, token);
  assert.equal(status, 401);
});

test('POST /auth/logout — missing Authorization returns 401', async () => {
  const { status } = await request('POST', '/auth/logout');
  assert.equal(status, 401);
});

// ---------------------------------------------------------------------------
// GET /tasks
// ---------------------------------------------------------------------------

test('GET /tasks — returns empty array for new user', async () => {
  const token = await loginAlice();
  const { status, body } = await request('GET', '/tasks', null, token);
  assert.equal(status, 200);
  assert.ok(Array.isArray(body.tasks), 'tasks should be an array');
});

test('GET /tasks — no auth returns 401', async () => {
  const { status } = await request('GET', '/tasks');
  assert.equal(status, 401);
});

test('GET /tasks — invalid token returns 401', async () => {
  const { status } = await request('GET', '/tasks', null, 'invalid-token-xyz');
  assert.equal(status, 401);
});

// ---------------------------------------------------------------------------
// POST /tasks
// ---------------------------------------------------------------------------

test('POST /tasks — creates a task and returns 201', async () => {
  const token = await loginAlice();
  const { status, body } = await request(
    'POST', '/tasks',
    { title: 'Buy groceries', description: 'Milk, eggs, bread' },
    token
  );
  assert.equal(status, 201);
  assert.ok(typeof body.id === 'string');
  assert.equal(body.title, 'Buy groceries');
  assert.equal(body.description, 'Milk, eggs, bread');
  assert.equal(body.status, 'todo');
  assert.ok(typeof body.createdAt === 'string');
  assert.ok(typeof body.updatedAt === 'string');
});

test('POST /tasks — task appears in GET /tasks', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Unique task for list test' }, token);
  const { body: list } = await request('GET', '/tasks', null, token);
  const found = list.tasks.find((t) => t.id === created.id);
  assert.ok(found, 'created task should appear in task list');
});

test('POST /tasks — no auth returns 401', async () => {
  const { status } = await request('POST', '/tasks', { title: 'Unauthorized task' });
  assert.equal(status, 401);
});

test('POST /tasks — missing title returns 400', async () => {
  const token = await loginAlice();
  const { status } = await request('POST', '/tasks', { description: 'No title here' }, token);
  assert.equal(status, 400);
});

// ---------------------------------------------------------------------------
// PUT /tasks/:id
// ---------------------------------------------------------------------------

test('PUT /tasks/:id — updates a task', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Original title' }, token);
  const { status, body } = await request(
    'PUT', `/tasks/${created.id}`,
    { title: 'Updated title', status: 'in_progress' },
    token
  );
  assert.equal(status, 200);
  assert.equal(body.title, 'Updated title');
  assert.equal(body.status, 'in_progress');
});

test('PUT /tasks/:id — invalid status returns 400', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Status test' }, token);
  const { status } = await request(
    'PUT', `/tasks/${created.id}`,
    { status: 'flying' },
    token
  );
  assert.equal(status, 400);
});

test('PUT /tasks/:id — non-existent id returns 404', async () => {
  const token = await loginAlice();
  const { status } = await request('PUT', '/tasks/nonexistentid', { title: 'Ghost' }, token);
  assert.equal(status, 404);
});

test('PUT /tasks/:id — no auth returns 401', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Auth test task' }, token);
  const { status } = await request('PUT', `/tasks/${created.id}`, { title: 'No token' });
  assert.equal(status, 401);
});

// ---------------------------------------------------------------------------
// DELETE /tasks/:id
// ---------------------------------------------------------------------------

test('DELETE /tasks/:id — deletes a task and returns 204', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Task to delete' }, token);
  const { status } = await request('DELETE', `/tasks/${created.id}`, null, token);
  assert.equal(status, 204);
});

test('DELETE /tasks/:id — deleted task no longer in GET /tasks', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Delete and verify' }, token);
  await request('DELETE', `/tasks/${created.id}`, null, token);
  const { body: list } = await request('GET', '/tasks', null, token);
  const found = list.tasks.find((t) => t.id === created.id);
  assert.equal(found, undefined, 'deleted task should not appear in list');
});

test('DELETE /tasks/:id — non-existent id returns 404', async () => {
  const token = await loginAlice();
  const { status } = await request('DELETE', '/tasks/doesnotexist', null, token);
  assert.equal(status, 404);
});

test('DELETE /tasks/:id — no auth returns 401', async () => {
  const token = await loginAlice();
  const { body: created } = await request('POST', '/tasks', { title: 'Delete no auth' }, token);
  const { status } = await request('DELETE', `/tasks/${created.id}`);
  assert.equal(status, 401);
});

// ---------------------------------------------------------------------------
// User isolation (tasks scoped to owner)
// ---------------------------------------------------------------------------

test('Tasks are scoped: bob cannot see alice\'s tasks', async () => {
  const aliceToken = await loginAlice();
  const bobToken = await loginBob();

  const { body: created } = await request('POST', '/tasks', { title: 'Alice private task' }, aliceToken);

  const { body: bobList } = await request('GET', '/tasks', null, bobToken);
  const found = bobList.tasks.find((t) => t.id === created.id);
  assert.equal(found, undefined, 'bob should not see alice\'s tasks');
});

test('Tasks are scoped: bob cannot delete alice\'s task', async () => {
  const aliceToken = await loginAlice();
  const bobToken = await loginBob();

  const { body: created } = await request('POST', '/tasks', { title: 'Alice task bob tries to delete' }, aliceToken);
  const { status } = await request('DELETE', `/tasks/${created.id}`, null, bobToken);

  assert.equal(status, 404, 'bob deleting alice\'s task should return 404');
});

test('Tasks are scoped: bob cannot update alice\'s task', async () => {
  const aliceToken = await loginAlice();
  const bobToken = await loginBob();

  const { body: created } = await request('POST', '/tasks', { title: 'Alice task bob tries to edit' }, aliceToken);
  const { status } = await request('PUT', `/tasks/${created.id}`, { title: 'Bob hijack' }, bobToken);

  assert.equal(status, 404, 'bob updating alice\'s task should return 404');
});

// ---------------------------------------------------------------------------
// Unknown routes
// ---------------------------------------------------------------------------

test('Unknown route returns 404', async () => {
  const { status } = await request('GET', '/unknown-route');
  assert.equal(status, 404);
});
