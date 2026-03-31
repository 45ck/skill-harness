/**
 * tasks.js — Task CRUD store
 *
 * All tasks are stored in-memory, scoped to the user who created them.
 * Valid statuses: 'todo', 'in_progress', 'done'
 */

import { randomBytes } from 'node:crypto';
import { getUsernameFromToken } from './auth.js';

const VALID_STATUSES = new Set(['todo', 'in_progress', 'done']);

/**
 * In-memory task store: taskId → task object
 * @type {Map<string, { id: string, owner: string, title: string, description: string, status: string, createdAt: string, updatedAt: string }>}
 */
const taskStore = new Map();

/**
 * Generate a short unique task ID.
 * @returns {string}
 */
function generateId() {
  return randomBytes(8).toString('hex');
}

/**
 * Resolve and validate the authenticated user from a token.
 * Throws with a descriptive message on failure.
 * @param {string} token
 * @returns {string} username
 */
function resolveUser(token) {
  return getUsernameFromToken(token); // throws 'Invalid or expired session' if bad
}

/**
 * Create a new task for the authenticated user.
 * @param {string} token
 * @param {{ title: string, description?: string }} fields
 * @returns {{ id: string, title: string, description: string, status: string, createdAt: string, updatedAt: string }}
 */
export function createTask(token, { title, description = '' }) {
  const owner = resolveUser(token);

  if (!title || typeof title !== 'string' || title.trim() === '') {
    throw new Error('title is required and must be a non-empty string');
  }

  const now = new Date().toISOString();
  const task = {
    id: generateId(),
    owner,
    title: title.trim(),
    description: typeof description === 'string' ? description.trim() : '',
    status: 'todo',
    createdAt: now,
    updatedAt: now,
  };

  taskStore.set(task.id, task);

  // Return a copy without the internal owner field
  return taskWithoutOwner(task);
}

/**
 * Get all tasks for the authenticated user.
 * @param {string} token
 * @returns {Array}
 */
export function getTasks(token) {
  const owner = resolveUser(token);
  const results = [];
  for (const task of taskStore.values()) {
    if (task.owner === owner) {
      results.push(taskWithoutOwner(task));
    }
  }
  return results;
}

/**
 * Get a single task by ID for the authenticated user.
 * @param {string} token
 * @param {string} id
 * @returns {object}
 * @throws {Error} if not found or not owned by caller
 */
export function getTask(token, id) {
  const owner = resolveUser(token);
  const task = taskStore.get(id);

  if (!task) {
    throw new Error(`Task not found: ${id}`);
  }
  if (task.owner !== owner) {
    throw new Error(`Task not found: ${id}`);
  }

  return taskWithoutOwner(task);
}

/**
 * Update a task owned by the authenticated user.
 * @param {string} token
 * @param {string} id
 * @param {{ title?: string, description?: string, status?: string }} updates
 * @returns {object} updated task
 */
export function updateTask(token, id, updates) {
  const owner = resolveUser(token);
  const task = taskStore.get(id);

  if (!task) {
    throw new Error(`Task not found: ${id}`);
  }
  if (task.owner !== owner) {
    throw new Error(`Task not found: ${id}`);
  }

  if (updates.status !== undefined) {
    if (!VALID_STATUSES.has(updates.status)) {
      throw new Error(`Invalid status '${updates.status}'. Must be one of: todo, in_progress, done`);
    }
    task.status = updates.status;
  }

  if (updates.title !== undefined) {
    if (typeof updates.title !== 'string' || updates.title.trim() === '') {
      throw new Error('title must be a non-empty string');
    }
    task.title = updates.title.trim();
  }

  if (updates.description !== undefined) {
    task.description = typeof updates.description === 'string' ? updates.description.trim() : '';
  }

  task.updatedAt = new Date().toISOString();
  return taskWithoutOwner(task);
}

/**
 * Delete a task owned by the authenticated user.
 * @param {string} token
 * @param {string} id
 * @returns {void}
 * @throws {Error} if not found or not owned by caller
 */
export function deleteTask(token, id) {
  const owner = resolveUser(token);
  const task = taskStore.get(id);

  if (!task) {
    throw new Error(`Task not found: ${id}`);
  }
  if (task.owner !== owner) {
    throw new Error(`Task not found: ${id}`);
  }

  taskStore.delete(id);
}

/**
 * Return a copy of the task without the internal owner field.
 * @param {object} task
 * @returns {object}
 */
function taskWithoutOwner({ id, title, description, status, createdAt, updatedAt }) {
  return { id, title, description, status, createdAt, updatedAt };
}
