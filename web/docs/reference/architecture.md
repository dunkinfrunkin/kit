---
title: Architecture
sidebar_position: 2
---

# Architecture

## Overview

Kit has two components:

1. **CLI** (`kit`) -- a Go binary that runs on developer machines. Talks to the server over HTTP, writes files to local AI tool config directories.
2. **Server** (`kitd`) -- a Go binary that exposes a REST API and stores data in Postgres.

```
┌──────────────────────────────────────────────────┐
│                  Developer Machine                │
│                                                   │
│  ┌─────────┐     ┌───────────────────────────┐   │
│  │ kit CLI  │────>│  ~/.claude/               │   │
│  │          │────>│  ~/.codex/                │   │
│  │          │────>│  .cursor/rules/           │   │
│  └────┬─────┘     └───────────────────────────┘   │
│       │                                           │
└───────┼───────────────────────────────────────────┘
        │ HTTP (REST API)
        │
┌───────┼───────────────────────────────────────────┐
│       v              Kit Server                   │
│  ┌─────────┐     ┌───────────────────────────┐   │
│  │  kitd   │────>│  Postgres                 │   │
│  │  :8430  │     │  - items table            │   │
│  │         │     │  - profiles table         │   │
│  └─────────┘     └───────────────────────────┘   │
│                                                   │
└───────────────────────────────────────────────────┘
```

## Data flow

**Push:** CLI reads a local file, base64-encodes it, POSTs to the server. The server stores it in the `items` table. If the item targets a personal namespace, the server encrypts the content before storing.

**Install:** CLI GETs items from the server, base64-decodes the content, and writes files to the correct directories for Claude Code, Codex CLI, and Cursor.

## Database schema

Two tables:

**items** -- stores skills, hooks, and configs.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| namespace | TEXT | Team name or `@email` |
| type | TEXT | `skill`, `hook`, or `config` |
| name | TEXT | Item name |
| content | BYTEA | File content (encrypted for personal namespaces) |
| author | TEXT | Email of the user who pushed it |
| version | INTEGER | Auto-incremented on each push |
| created_at | TIMESTAMPTZ | First push timestamp |
| updated_at | TIMESTAMPTZ | Last push timestamp |

Unique constraint: `(namespace, type, name)`.

**profiles** -- bundles of item references.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| namespace | TEXT | Team name or `@email` |
| name | TEXT | Profile name |
| items | JSONB | Array of `{"name": "...", "type": "..."}` references |
| author | TEXT | Email of the creator |
| created_at | TIMESTAMPTZ | Creation timestamp |
| updated_at | TIMESTAMPTZ | Last update timestamp |

Unique constraint: `(namespace, name)`.

## Authentication

Stateless JWT tokens signed with `KIT_SECRET` using HMAC-SHA256. No user table exists. The email address from the JWT payload is the user's identity.

The flow:

1. `POST /login` with `{"email": "..."}` returns a signed JWT.
2. All subsequent requests include `Authorization: Bearer {token}`.
3. The server verifies the signature and extracts the email on each request.

## Encryption

Personal namespaces (prefixed with `@`) are encrypted at rest using AES-256-GCM.

Key derivation:

```
encryption_key = HMAC-SHA256(KIT_SECRET, user_email)
```

The key is never stored. It is derived per-request from the server's secret and the authenticated user's email. This means:

- Only the server with the correct `KIT_SECRET` can decrypt personal data.
- Each user gets a unique encryption key.
- If `KIT_SECRET` changes, all encrypted data becomes unreadable.

## Tech stack

| Component | Technology |
|-----------|------------|
| CLI | Go, Cobra |
| Server | Go, net/http |
| Database | Postgres 17 |
| Auth | HMAC-SHA256 JWT |
| Encryption | AES-256-GCM |
| Key derivation | HMAC-SHA256 |
