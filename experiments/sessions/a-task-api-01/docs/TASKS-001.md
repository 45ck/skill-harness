---
id: TASKS-001
title: "Task CRUD Store"
state: in_progress
kind: functional
required_evidence:
  implementation: E0
  test_coverage: E0
waivers:
  - kind: missing-verification
    target: TASKS-001
    owner: "agent"
    reason: "Specs declare E0 evidence (declarative). All task CRUD operations are exercised by integration tests in server.test.js (28/28 pass). No runtime verification provider configured; advisory warn is expected at E0."
    expires: "2026-07-01"
---

## Overview

The task store module manages task objects scoped to authenticated users. Every operation requires a valid session token. Tasks are held in an in-memory map; no database is used. The module enforces ownership: a user may only read, modify, or delete tasks they created.

## Requirements

### TASKS-001-R1: Task shape
- A task object has the following fields:
  - `id` — unique string identifier (UUID v4 or random hex)
  - `title` — non-empty string
  - `description` — string (may be empty)
  - `status` — one of `"todo"`, `"in_progress"`, `"done"` (default `"todo"`)
  - `createdAt` — ISO-8601 timestamp string set at creation time
  - `updatedAt` — ISO-8601 timestamp string updated on every mutation

### TASKS-001-R2: Create task
- `createTask(token, { title, description })` validates the session via the auth module.
- It throws on invalid session (auth failure).
- It throws if `title` is missing or empty.
- It generates a new task with status `"todo"` and returns the full task object.

### TASKS-001-R3: List tasks
- `getTasks(token)` validates the session.
- Returns an array of all tasks owned by the authenticated user (may be empty).
- Throws on invalid session.

### TASKS-001-R4: Get single task
- `getTask(token, id)` validates the session.
- Returns the task if it exists and is owned by the authenticated user.
- Throws a "not found" error if the id does not exist or belongs to another user.
- Throws on invalid session.

### TASKS-001-R5: Update task
- `updateTask(token, id, { title?, description?, status? })` validates the session and ownership.
- Applies only the provided fields (partial update).
- Throws if `status` is provided but is not one of the valid values.
- Updates `updatedAt` to the current time.
- Returns the updated task object.
- Throws on invalid session or ownership failure.

### TASKS-001-R6: Delete task
- `deleteTask(token, id)` validates the session and ownership.
- Removes the task from the store.
- Returns void on success.
- Throws on invalid session, not-found, or ownership failure.

### TASKS-001-R7: User scoping
- Tasks created by user A are never returned or mutated by operations authenticated as user B.

## Acceptance Criteria

- AC1: `createTask` with a valid token and title returns a task with `status === "todo"` and both timestamps set.
- AC2: `getTasks` returns only the tasks belonging to the authenticated user.
- AC3: `getTask` with a valid token and existing own task returns that task.
- AC4: `getTask` with a valid token but another user's task ID throws.
- AC5: `updateTask` with `status: "done"` returns a task with `status === "done"` and an updated `updatedAt`.
- AC6: `updateTask` with an invalid status throws.
- AC7: `deleteTask` removes the task; subsequent `getTask` for the same id throws.
- AC8: All operations throw descriptively on invalid session token.

## Out of Scope

- Persistent storage.
- Pagination, sorting, or filtering.
- Task assignment to other users.
- Task attachments or comments.
