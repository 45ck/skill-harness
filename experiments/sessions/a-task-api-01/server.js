/**
 * @file server.js
 * @spec API-001
 * HTTP API layer. Exposes auth and task operations over a JSON REST interface
 * using Node.js built-in `node:http` only. Call `createServer()` to obtain a
 * configured server; bind to a port with `.listen()` externally.
 */

import { createServer as httpCreateServer } from 'node:http';
import { login, logout, isValidSession } from './auth.js';
import { createTask, getTasks, getTask, updateTask, deleteTask } from './tasks.js';

/**
 * Read the full request body as a UTF-8 string.
 *
 * @spec API-001
 * @implements API-001-R2
 * @evidence E0
 * @param {import('node:http').IncomingMessage} req
 * @returns {Promise<string>}
 */
function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (chunk) => chunks.push(chunk));
    req.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
    req.on('error', reject);
  });
}

/**
 * Parse a JSON request body, returning null on parse failure.
 *
 * @spec API-001
 * @implements API-001-R2
 * @evidence E0
 * @param {string} raw
 * @returns {object|null}
 */
function parseJSON(raw) {
  try {
    return JSON.parse(raw || '{}');
  } catch {
    return null;
  }
}

/**
 * Send a JSON response.
 *
 * @spec API-001
 * @implements API-001-R11
 * @evidence E0
 * @param {import('node:http').ServerResponse} res
 * @param {number} status
 * @param {object} body
 */
function sendJSON(res, status, body) {
  const payload = JSON.stringify(body);
  res.writeHead(status, {
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(payload),
  });
  res.end(payload);
}

/**
 * Send an empty 204 No Content response.
 *
 * @spec API-001
 * @implements API-001-R11
 * @evidence E0
 * @param {import('node:http').ServerResponse} res
 */
function send204(res) {
  res.writeHead(204);
  res.end();
}

/**
 * Extract the Bearer token from the Authorization header.
 * Returns null if the header is absent or malformed.
 *
 * @spec API-001
 * @implements API-001-R3
 * @evidence E0
 * @param {import('node:http').IncomingMessage} req
 * @returns {string|null}
 */
function extractToken(req) {
  const header = req.headers['authorization'] ?? '';
  const match = header.match(/^Bearer\s+(\S+)$/i);
  return match ? match[1] : null;
}

/**
 * Require a valid auth token, sending 401 if missing or invalid.
 * Returns the token string on success or null (after sending the response).
 *
 * @spec API-001
 * @implements API-001-R3
 * @evidence E0
 * @param {import('node:http').IncomingMessage} req
 * @param {import('node:http').ServerResponse} res
 * @returns {string|null}
 */
function requireAuth(req, res) {
  const token = extractToken(req);
  if (!token || !isValidSession(token)) {
    sendJSON(res, 401, { error: 'Unauthorized' });
    return null;
  }
  return token;
}

/**
 * Create and return a configured HTTP server without calling .listen().
 *
 * @spec API-001
 * @implements API-001-R1
 * @evidence E0
 * @returns {import('node:http').Server}
 */
export function createServer() {
  const server = httpCreateServer(async (req, res) => {
    const { method, url } = req;

    try {
      // POST /auth/login
      if (method === 'POST' && url === '/auth/login') {
        const raw = await readBody(req);
        const body = parseJSON(raw);
        if (!body) return sendJSON(res, 400, { error: 'Invalid JSON body' });
        const { username, password } = body;
        if (!username || !password) {
          return sendJSON(res, 400, { error: 'username and password are required' });
        }
        try {
          const result = login(username, password);
          return sendJSON(res, 200, result);
        } catch {
          return sendJSON(res, 401, { error: 'Invalid credentials' });
        }
      }

      // POST /auth/logout
      if (method === 'POST' && url === '/auth/logout') {
        const token = extractToken(req);
        if (!token) return sendJSON(res, 401, { error: 'Unauthorized' });
        logout(token);
        return send204(res);
      }

      // GET /tasks
      if (method === 'GET' && url === '/tasks') {
        const token = requireAuth(req, res);
        if (!token) return;
        const tasks = getTasks(token);
        return sendJSON(res, 200, { tasks });
      }

      // POST /tasks
      if (method === 'POST' && url === '/tasks') {
        const token = requireAuth(req, res);
        if (!token) return;
        const raw = await readBody(req);
        const body = parseJSON(raw);
        if (!body) return sendJSON(res, 400, { error: 'Invalid JSON body' });
        if (!body.title) return sendJSON(res, 400, { error: 'title is required' });
        try {
          const task = createTask(token, { title: body.title, description: body.description });
          return sendJSON(res, 201, task);
        } catch (err) {
          if (err.code === 'BAD_REQUEST') return sendJSON(res, 400, { error: err.message });
          return sendJSON(res, 400, { error: err.message });
        }
      }

      // PUT /tasks/:id
      const putMatch = method === 'PUT' && url.match(/^\/tasks\/([^/]+)$/);
      if (putMatch) {
        const token = requireAuth(req, res);
        if (!token) return;
        const id = putMatch[1];
        const raw = await readBody(req);
        const body = parseJSON(raw);
        if (!body) return sendJSON(res, 400, { error: 'Invalid JSON body' });
        try {
          const task = updateTask(token, id, body);
          return sendJSON(res, 200, task);
        } catch (err) {
          if (err.code === 'NOT_FOUND') return sendJSON(res, 404, { error: err.message });
          if (err.code === 'BAD_REQUEST') return sendJSON(res, 400, { error: err.message });
          return sendJSON(res, 400, { error: err.message });
        }
      }

      // DELETE /tasks/:id
      const deleteMatch = method === 'DELETE' && url.match(/^\/tasks\/([^/]+)$/);
      if (deleteMatch) {
        const token = requireAuth(req, res);
        if (!token) return;
        const id = deleteMatch[1];
        try {
          deleteTask(token, id);
          return send204(res);
        } catch (err) {
          if (err.code === 'NOT_FOUND') return sendJSON(res, 404, { error: err.message });
          return sendJSON(res, 404, { error: err.message });
        }
      }

      // 404 fallthrough
      sendJSON(res, 404, { error: 'Not found' });
    } catch (err) {
      sendJSON(res, 500, { error: 'Internal server error' });
    }
  });

  return server;
}
