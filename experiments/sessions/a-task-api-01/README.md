# Task Management REST API

A Task Management REST API built with Node.js stdlib only (`node:http`, `node:crypto`). No Express, no external packages.

## Setup

```bash
npm install
npm test
```

Requires Node.js >= 22.

## Starting the server

```js
import { createServer } from './server.js';
const server = createServer();
server.listen(3000, () => console.log('Listening on port 3000'));
```

## Authentication

All task endpoints require an `Authorization: Bearer <token>` header obtained from `POST /auth/login`.

Built-in test users:

| Username | Password              |
|----------|-----------------------|
| alice    | T3stP@ssw0rd!Alice    |
| bob      | T3stP@ssw0rd!Bob      |

---

## Endpoints

### POST /auth/login

Authenticate and receive a session token.

**Request body**
```json
{ "username": "alice", "password": "T3stP@ssw0rd!Alice" }
```

**Responses**

| Status | Body                    | Condition              |
|--------|-------------------------|------------------------|
| 200    | `{ "token": "<hex>" }` | Valid credentials      |
| 400    | `{ "error": "..." }`   | Missing fields         |
| 401    | `{ "error": "..." }`   | Invalid credentials    |

**Example**
```bash
curl -X POST http://localhost:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"T3stP@ssw0rd!Alice"}'
# {"token":"a1b2c3..."}
```

---

### POST /auth/logout

Invalidate the current session.

**Headers**: `Authorization: Bearer <token>`

**Responses**

| Status | Body | Condition        |
|--------|------|------------------|
| 204    | —    | Session removed  |
| 401    | —    | Missing token    |

**Example**
```bash
curl -X POST http://localhost:3000/auth/logout \
  -H 'Authorization: Bearer a1b2c3...'
```

---

### GET /tasks

List all tasks belonging to the authenticated user.

**Headers**: `Authorization: Bearer <token>`

**Responses**

| Status | Body                           | Condition     |
|--------|--------------------------------|---------------|
| 200    | `{ "tasks": [ ... ] }`        | Success       |
| 401    | `{ "error": "Unauthorized" }` | Bad/no token  |

**Example**
```bash
curl http://localhost:3000/tasks \
  -H 'Authorization: Bearer a1b2c3...'
# {"tasks":[{"id":"...","title":"Buy groceries","status":"todo",...}]}
```

---

### POST /tasks

Create a new task.

**Headers**: `Authorization: Bearer <token>`

**Request body**
```json
{ "title": "Buy groceries", "description": "Milk, eggs, bread" }
```

**Responses**

| Status | Body                           | Condition        |
|--------|--------------------------------|------------------|
| 201    | Task object                    | Created          |
| 400    | `{ "error": "..." }`          | Missing title    |
| 401    | `{ "error": "Unauthorized" }` | Bad/no token     |

**Task object shape**
```json
{
  "id": "4f3a...",
  "title": "Buy groceries",
  "description": "Milk, eggs, bread",
  "status": "todo",
  "createdAt": "2026-04-01T10:00:00.000Z",
  "updatedAt": "2026-04-01T10:00:00.000Z"
}
```

**Example**
```bash
curl -X POST http://localhost:3000/tasks \
  -H 'Authorization: Bearer a1b2c3...' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy groceries","description":"Milk, eggs, bread"}'
```

---

### PUT /tasks/:id

Update an existing task (partial update — only provided fields are changed).

**Headers**: `Authorization: Bearer <token>`

**Request body** (all fields optional)
```json
{ "title": "Updated title", "description": "New desc", "status": "in_progress" }
```

Valid status values: `todo`, `in_progress`, `done`.

**Responses**

| Status | Body                           | Condition            |
|--------|--------------------------------|----------------------|
| 200    | Updated task object            | Success              |
| 400    | `{ "error": "..." }`          | Invalid status       |
| 401    | `{ "error": "Unauthorized" }` | Bad/no token         |
| 404    | `{ "error": "..." }`          | Not found / not owned|

**Example**
```bash
curl -X PUT http://localhost:3000/tasks/4f3a... \
  -H 'Authorization: Bearer a1b2c3...' \
  -H 'Content-Type: application/json' \
  -d '{"status":"done"}'
```

---

### DELETE /tasks/:id

Delete a task.

**Headers**: `Authorization: Bearer <token>`

**Responses**

| Status | Body                           | Condition            |
|--------|--------------------------------|----------------------|
| 204    | —                              | Deleted              |
| 401    | `{ "error": "Unauthorized" }` | Bad/no token         |
| 404    | `{ "error": "..." }`          | Not found / not owned|

**Example**
```bash
curl -X DELETE http://localhost:3000/tasks/4f3a... \
  -H 'Authorization: Bearer a1b2c3...'
```

---

## Architecture

| File             | Spec       | Description                                   |
|------------------|------------|-----------------------------------------------|
| `auth.js`        | AUTH-001   | Login/logout/session validation (SHA-256+salt)|
| `tasks.js`       | TASKS-001  | In-memory task CRUD, user-scoped              |
| `server.js`      | API-001    | `node:http` server factory, JSON routing      |
| `server.test.js` | API-001    | 28 integration tests (`node:test`)            |

## Notes

- Sessions expire after 30 minutes of inactivity (sliding window).
- Tasks are user-scoped: user A cannot see or modify user B's tasks.
- All data is in-memory; restarts clear all sessions and tasks.
