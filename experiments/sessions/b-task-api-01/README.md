# Task Management REST API

A simple Task Management REST API built with Node.js using only the standard library (`node:http`, `node:crypto`). No Express or external packages.

## Purpose

Demonstrates full CRUD for user-scoped tasks with token-based authentication. Tasks belong to the user who created them — each user sees and manages only their own tasks.

## Setup

Node.js >= 22 is required.

```bash
# No dependencies to install (stdlib only)
npm test       # run all tests
node server.js # start the server manually (listens on PORT env var or 3000)
```

### Default users

| Username | Password        |
|----------|-----------------|
| alice    | alicepassword   |
| bob      | bobpassword     |

Passwords are stored as SHA-256 hashes with per-user salts — never in plaintext.

---

## Endpoints

### Authentication

#### `POST /auth/login`

Authenticate with username and password. Returns a session token valid for 30 minutes of inactivity.

**Request body**
```json
{ "username": "alice", "password": "alicepassword" }
```

**Response — 200 OK**
```json
{ "token": "<session-token>" }
```

**Error responses**
- `400` — missing username or password
- `401` — invalid credentials

---

#### `POST /auth/logout`

Invalidate the current session token.

**Headers**
```
Authorization: Bearer <token>
```

**Response — 204 No Content**

**Error responses**
- `401` — missing or no Authorization header

---

### Tasks

All task endpoints require `Authorization: Bearer <token>`.

#### `GET /tasks`

Return all tasks owned by the authenticated user.

**Response — 200 OK**
```json
{
  "tasks": [
    {
      "id": "a1b2c3d4e5f6a1b2",
      "title": "Buy groceries",
      "description": "Milk, eggs, bread",
      "status": "todo",
      "createdAt": "2026-04-01T12:00:00.000Z",
      "updatedAt": "2026-04-01T12:00:00.000Z"
    }
  ]
}
```

**Error responses**
- `401` — missing/invalid/expired token

---

#### `POST /tasks`

Create a new task for the authenticated user.

**Request body**
```json
{ "title": "Buy groceries", "description": "Milk, eggs, bread" }
```

`description` is optional (defaults to `""`).

**Response — 201 Created**
```json
{
  "id": "a1b2c3d4e5f6a1b2",
  "title": "Buy groceries",
  "description": "Milk, eggs, bread",
  "status": "todo",
  "createdAt": "2026-04-01T12:00:00.000Z",
  "updatedAt": "2026-04-01T12:00:00.000Z"
}
```

**Error responses**
- `400` — missing or empty title, invalid JSON
- `401` — missing/invalid/expired token

---

#### `PUT /tasks/:id`

Update a task owned by the authenticated user. All body fields are optional.

**Request body** (all optional)
```json
{ "title": "New title", "description": "New description", "status": "in_progress" }
```

Valid status values: `todo`, `in_progress`, `done`

**Response — 200 OK** — returns the updated task object (same shape as POST response).

**Error responses**
- `400` — invalid status value, invalid JSON
- `401` — missing/invalid/expired token
- `404` — task not found or not owned by caller

---

#### `DELETE /tasks/:id`

Delete a task owned by the authenticated user.

**Response — 204 No Content**

**Error responses**
- `401` — missing/invalid/expired token
- `404` — task not found or not owned by caller

---

## Implementation notes

- **No external packages** — `node:http`, `node:crypto`, `node:test`, `node:assert/strict` only
- **ESM** — `"type": "module"` in package.json
- **Passwords** — SHA-256 + per-user salt; no plaintext storage
- **Sessions** — cryptographically random 32-byte hex tokens; expire after 30 min inactivity
- **Task isolation** — tasks are owner-scoped; cross-user access returns 404 (not 403) to avoid leaking existence
- **In-memory storage** — all data is lost on process restart (by design for this demo)
