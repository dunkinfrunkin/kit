---
title: Deployment
sidebar_position: 1
---

# Deployment

## Docker Compose (recommended)

Create a `docker-compose.yml`:

```yaml
services:
  db:
    image: postgres:17
    environment:
      POSTGRES_DB: kit
      POSTGRES_USER: kit
      POSTGRES_PASSWORD: changeme
    volumes:
      - kit-db:/var/lib/postgresql/data

  server:
    image: dunkinfrunkin/kit
    ports:
      - "8430:8430"
    environment:
      KIT_SECRET: change-me-to-a-real-secret
      DATABASE_URL: postgres://kit:changeme@db:5432/kit?sslmode=disable
    depends_on:
      - db

volumes:
  kit-db:
```

Start it:

```bash
docker compose up -d
```

The server is available at `http://localhost:8430`. The web dashboard is at `http://localhost:8430/ui`.

## Docker run with existing Postgres

If you already have a Postgres instance:

```bash
docker run -d \
  -p 8430:8430 \
  -e KIT_SECRET=change-me-to-a-real-secret \
  -e DATABASE_URL=postgres://user:pass@your-db-host:5432/kit?sslmode=disable \
  dunkinfrunkin/kit
```

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `KIT_SECRET` | Yes | -- | Used for JWT signing and encryption key derivation. |
| `DATABASE_URL` | Yes | -- | Postgres connection string. |
| `PORT` | No | `8430` | HTTP port the server listens on. |

## Migrations

The server auto-runs database migrations on startup. No manual migration step needed. It creates two tables (`items` and `profiles`) if they don't exist.

## Web dashboard

A built-in web dashboard is served at `/ui`. Point your browser at `http://your-server:8430/ui` to browse namespaces, items, and profiles.

## Backup

Kit uses standard Postgres. Back up with `pg_dump`:

```bash
pg_dump kit > backup.sql
```

Restore:

```bash
psql kit < backup.sql
```

All standard Postgres backup tooling works -- `pg_dump`, `pg_restore`, managed provider snapshots, etc.
