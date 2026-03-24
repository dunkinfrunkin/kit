---
title: Installation
sidebar_position: 1
---

# Installation

## CLI

### Homebrew

```bash
brew install dunkinfrunkin/tap/kit
```

### Go install

```bash
go install github.com/dunkinfrunkin/kit/cmd/kit@latest
```

Requires Go 1.23 or later.

### Build from source

```bash
git clone https://github.com/dunkinfrunkin/kit.git
cd kit
make build
```

The binary is written to `./bin/kit`.

## Server

### Docker Compose (recommended)

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
    build: .
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

Then start it:

```bash
docker compose up -d
```

The server is available at `http://localhost:8430`.

### Docker run with existing Postgres

If you already have a Postgres instance running:

```bash
docker run -d \
  -p 8430:8430 \
  -e KIT_SECRET=change-me-to-a-real-secret \
  -e DATABASE_URL=postgres://user:pass@your-db-host:5432/kit?sslmode=disable \
  dunkinfrunkin/kit
```

## Requirements

| Component | Requirement |
|-----------|-------------|
| CLI (build from source) | Go 1.23+ |
| Server (Docker) | Docker, Docker Compose |
| Server (manual) | Postgres 17 |
