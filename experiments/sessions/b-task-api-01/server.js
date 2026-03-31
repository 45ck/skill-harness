/**
 * server.js — HTTP API server
 *
 * Creates an HTTP server handling all Task Management API routes.
 * Uses only node:http — no Express or external dependencies.
 *
 * Routes:
 *   POST /auth/login        — { username, password } → { token }
 *   POST /auth/logout       — Authorization: Bearer <token> → 204
 *   GET  /tasks             — auth → { tasks: [...] }
 *   POST /tasks             — auth + { title, description } → 201 + task
 *   PUT  /tasks/:id         — auth + body → updated task
 *   DELETE /tasks/:id       — auth → 204
 */

import { createServer as httpCreateServer } from 'node:http';
import { login, logout } from './auth.js';
import { createTask, getTasks, getTask, updateTask, deleteTask } from './tasks.js';

/**
 * Parse the request body as JSON.
 * @param {import('node:http').IncomingMessage} req
 * @returns {Promise<any>}
 */
function parseBody(req) {
  return new Promise((resolve, reject) => {
    let raw = '';
    req.on('data', (chunk) => { raw += chunk; });
    req.on('end', () => {
      if (!raw) {
        resolve({});
        return;
      }
      try {
        resolve(JSON.parse(raw));
      } catch {
        reject(new Error('Invalid JSON body'));
      }
    });
    req.on('error', reject);
  });
}

/**
 * Send a JSON response.
 * @param {import('node:http').ServerResponse} res
 * @param {number} statusCode
 * @param {any} body
 */
function sendJSON(res, statusCode, body) {
  const payload = JSON.stringify(body);
  res.writeHead(statusCode, {
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(payload),
  });
  res.end(payload);
}

/**
 * Send a 204 No Content response.
 * @param {import('node:http').ServerResponse} res
 */
function send204(res) {
  res.writeHead(204);
  res.end();
}

/**
 * Extract the Bearer token from the Authorization header.
 * @param {import('node:http').IncomingMessage} req
 * @returns {string|null}
 */
function extractToken(req) {
  const auth = req.headers['authorization'];
  if (!auth || !auth.startsWith('Bearer ')) return null;
  return auth.slice(7).trim() || null;
}

/**
 * Route handler: POST /auth/login
 */
async function handleLogin(req, res) {
  let body;
  try {
    body = await parseBody(req);
  } catch {
    return sendJSON(res, 400, { error: 'Invalid JSON body' });
  }

  const { username, password } = body;
  if (!username || !password) {
    return sendJSON(res, 400, { error: 'username and password are required' });
  }

  try {
    const result = login(username, password);
    sendJSON(res, 200, result);
  } catch (err) {
    sendJSON(res, 401, { error: err.message });
  }
}

/**
 * Route handler: POST /auth/logout
 */
async function handleLogout(req, res) {
  const token = extractToken(req);
  if (!token) {
    return sendJSON(res, 401, { error: 'Authorization header required' });
  }
  logout(token);
  send204(res);
}

/**
 * Route handler: GET /tasks
 */
async function handleGetTasks(req, res) {
  const token = extractToken(req);
  if (!token) {
    return sendJSON(res, 401, { error: 'Authorization header required' });
  }

  try {
    const tasks = getTasks(token);
    sendJSON(res, 200, { tasks });
  } catch (err) {
    sendJSON(res, 401, { error: err.message });
  }
}

/**
 * Route handler: POST /tasks
 */
async function handleCreateTask(req, res) {
  const token = extractToken(req);
  if (!token) {
    return sendJSON(res, 401, { error: 'Authorization header required' });
  }

  let body;
  try {
    body = await parseBody(req);
  } catch {
    return sendJSON(res, 400, { error: 'Invalid JSON body' });
  }

  try {
    const task = createTask(token, body);
    sendJSON(res, 201, task);
  } catch (err) {
    const status = err.message.includes('session') ? 401 : 400;
    sendJSON(res, status, { error: err.message });
  }
}

/**
 * Route handler: PUT /tasks/:id
 */
async function handleUpdateTask(req, res, id) {
  const token = extractToken(req);
  if (!token) {
    return sendJSON(res, 401, { error: 'Authorization header required' });
  }

  let body;
  try {
    body = await parseBody(req);
  } catch {
    return sendJSON(res, 400, { error: 'Invalid JSON body' });
  }

  try {
    const task = updateTask(token, id, body);
    sendJSON(res, 200, task);
  } catch (err) {
    if (err.message.includes('session')) return sendJSON(res, 401, { error: err.message });
    if (err.message.includes('not found')) return sendJSON(res, 404, { error: err.message });
    sendJSON(res, 400, { error: err.message });
  }
}

/**
 * Route handler: DELETE /tasks/:id
 */
async function handleDeleteTask(req, res, id) {
  const token = extractToken(req);
  if (!token) {
    return sendJSON(res, 401, { error: 'Authorization header required' });
  }

  try {
    deleteTask(token, id);
    send204(res);
  } catch (err) {
    if (err.message.includes('session')) return sendJSON(res, 401, { error: err.message });
    if (err.message.includes('not found')) return sendJSON(res, 404, { error: err.message });
    sendJSON(res, 400, { error: err.message });
  }
}

/**
 * Create and return an HTTP server (caller is responsible for calling .listen()).
 * @returns {import('node:http').Server}
 */
export function createServer() {
  const server = httpCreateServer(async (req, res) => {
    const url = req.url || '/';
    const method = req.method || 'GET';

    // Route: POST /auth/login
    if (method === 'POST' && url === '/auth/login') {
      return handleLogin(req, res);
    }

    // Route: POST /auth/logout
    if (method === 'POST' && url === '/auth/logout') {
      return handleLogout(req, res);
    }

    // Route: GET /tasks
    if (method === 'GET' && url === '/tasks') {
      return handleGetTasks(req, res);
    }

    // Route: POST /tasks
    if (method === 'POST' && url === '/tasks') {
      return handleCreateTask(req, res);
    }

    // Route: PUT /tasks/:id
    const putMatch = url.match(/^\/tasks\/([^/]+)$/);
    if (method === 'PUT' && putMatch) {
      return handleUpdateTask(req, res, putMatch[1]);
    }

    // Route: DELETE /tasks/:id
    const deleteMatch = url.match(/^\/tasks\/([^/]+)$/);
    if (method === 'DELETE' && deleteMatch) {
      return handleDeleteTask(req, res, deleteMatch[1]);
    }

    // 404 for everything else
    sendJSON(res, 404, { error: 'Not found' });
  });

  return server;
}
