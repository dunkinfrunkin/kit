---
title: Configuration
sidebar_position: 2
---

# Configuration

Kit is configured entirely through environment variables. There are no config files on the server side.

## Server environment variables

### KIT_SECRET

The most important configuration value. Used for two things:

1. **JWT signing** -- all authentication tokens are signed with this secret.
2. **Encryption key derivation** -- per-user encryption keys are derived from `HMAC-SHA256(KIT_SECRET, email)`.

Keep it secure. If you change it:
- All existing JWT tokens are invalidated (every user must re-login).
- All encrypted personal namespace data becomes unreadable.

Generate a strong secret:

```bash
openssl rand -hex 32
```

### DATABASE_URL

Standard Postgres connection string. Supports all Postgres providers:

```
postgres://user:password@host:5432/kit?sslmode=disable   # local
postgres://user:password@rds-host:5432/kit?sslmode=require  # AWS RDS
postgres://user:password@cloud-sql-host:5432/kit             # Cloud SQL
postgres://user:password@db.supabase.co:5432/kit             # Supabase
```

### PORT

HTTP port the server listens on. Default: `8430`.

```bash
PORT=9000  # listen on port 9000 instead
```

## CLI configuration

The CLI stores credentials at `~/.kit/credentials`. This file is created by `kit login` with `0600` permissions (owner read/write only).

Contents:

```json
{
  "server": "http://localhost:8430",
  "email": "alice@example.com",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

To change servers, run `kit login --server <new-url>`. To clear credentials, run `kit logout`.
