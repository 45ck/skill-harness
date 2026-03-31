/**
 * @file tasks.js
 * @spec TASKS-001
 * In-memory task CRUD store. All operations require a valid session token.
 * Tasks are scoped to the authenticated user.
 */

import { randomBytes } from 'node:crypto';
import { requireUsername, isValidSession } from './auth.js';

/** @type {Set<string>} */
const VALID_STATUSES = new Set(['todo', 'in_progress', 'done']);

/**
 * In-memory task store: taskId → task object.
 * @type {Map<string, { id: string, owner: string, title: string, description: string, status: string, createdAt: string, updatedAt: string }>}
 */
const STORE = new Map();

/**
 * Generate a random 16-byte hex ID.
 * @returns {string}
 */
function generateId() {
  return randomBytes(16).toString('hex');
}

/**
 * Assert the token is valid and return the associated username.
 * @param {string} token
 * @returns {string} username
 * @throws {Error} On invalid session.
 */
function assertAuth(token) {
  return requireUsername(token);
}

/**
 * Create a new task for the authenticated user.
 *
 * @spec TASKS-001
 * @implements TASKS-001-R2
 * @evidence E0
 * @param {string} token - Valid session token.
 * @param {{ title: string, description?: string }} fields
 * @returns {{ id: string, title: string, description: string, status: string, createdAt: string, updatedAt: string }}
 * @throws {Error} On invalid session or missing title.
 */
export function createTask(token, { title, description = '' }) {
  const owner = assertAuth(token);
  if (!title || typeof title !== 'string' || title.trim() === '') {
    throw new Error('title is required and must be a non-empty string');
  }
  const now = new Date().toISOString();
  const task = {
    id: generateId(),
    owner,
    title: title.trim(),
    description: description ?? '',
    status: 'todo',
    createdAt: now,
    updatedAt: now,
  };
  STORE.set(task.id, task);
  return publicView(task);
}

/**
 * List all tasks owned by the authenticated user.
 *
 * @spec TASKS-001
 * @implements TASKS-001-R3
 * @evidence E0
 * @param {string} token
 * @returns {Array<object>}
 * @throws {Error} On invalid session.
 */
export function getTasks(token) {
  const owner = assertAuth(token);
  const result = [];
  for (const task of STORE.values()) {
    if (task.owner === owner) result.push(publicView(task));
  }
  return result;
}

/**
 * Get a single task by ID, enforcing ownership.
 *
 * @spec TASKS-001
 * @implements TASKS-001-R4
 * @evidence E0
 * @param {string} token
 * @param {string} id
 * @returns {object}
 * @throws {Error} On invalid session, not found, or ownership mismatch.
 */
export function getTask(token, id) {
  const owner = assertAuth(token);
  const task = STORE.get(id);
  if (!task || task.owner !== owner) {
    throw Object.assign(new Error(`Task not found: ${id}`), { code: 'NOT_FOUND' });
  }
  return publicView(task);
}

/**
 * Update fields of an existing task, enforcing ownership.
 *
 * @spec TASKS-001
 * @implements TASKS-001-R5
 * @evidence E0
 * @param {string} token
 * @param {string} id
 * @param {{ title?: string, description?: string, status?: string }} patch
 * @returns {object} Updated task.
 * @throws {Error} On invalid session, not found, ownership mismatch, or invalid status.
 */
export function updateTask(token, id, patch = {}) {
  const owner = assertAuth(token);
  const task = STORE.get(id);
  if (!task || task.owner !== owner) {
    throw Object.assign(new Error(`Task not found: ${id}`), { code: 'NOT_FOUND' });
  }
  if (patch.status !== undefined && !VALID_STATUSES.has(patch.status)) {
    throw Object.assign(
      new Error(`Invalid status "${patch.status}". Valid values: ${[...VALID_STATUSES].join(', ')}`),
      { code: 'BAD_REQUEST' }
    );
  }
  if (patch.title !== undefined) {
    if (typeof patch.title !== 'string' || patch.title.trim() === '') {
      throw Object.assign(new Error('title must be a non-empty string'), { code: 'BAD_REQUEST' });
    }
    task.title = patch.title.trim();
  }
  if (patch.description !== undefined) task.description = patch.description;
  if (patch.status !== undefined) task.status = patch.status;
  task.updatedAt = new Date().toISOString();
  return publicView(task);
}

/**
 * Delete a task by ID, enforcing ownership.
 *
 * @spec TASKS-001
 * @implements TASKS-001-R6
 * @evidence E0
 * @param {string} token
 * @param {string} id
 * @returns {void}
 * @throws {Error} On invalid session, not found, or ownership mismatch.
 */
export function deleteTask(token, id) {
  const owner = assertAuth(token);
  const task = STORE.get(id);
  if (!task || task.owner !== owner) {
    throw Object.assign(new Error(`Task not found: ${id}`), { code: 'NOT_FOUND' });
  }
  STORE.delete(id);
}

/**
 * Return a task object without the internal `owner` field.
 * @param {object} task
 * @returns {object}
 */
function publicView({ id, title, description, status, createdAt, updatedAt }) {
  return { id, title, description, status, createdAt, updatedAt };
}
