---
title: Quickstart
sidebar_position: 2
---

# Quickstart

This guide walks you through starting the server, pushing your first artifacts, and installing them.

## 1. Start the server

```bash
docker compose up -d
```

This starts Postgres and the kit server on port 8430.

## 2. Log in

```bash
kit login --server http://localhost:8430
```

You'll be prompted for your email address. Kit sends a verification code to complete login.

## 3. Push a skill

```bash
kit push ./my-skill/ --team backend
```

This uploads the skill directory to the `backend` team namespace on the server.

## 4. Push a config

```bash
kit push ./conventions.md --team backend
```

Individual files work too. Markdown files are treated as coding conventions.

## 5. List what's on the server

```bash
kit list backend
```

Shows all skills, hooks, and configs in the `backend` namespace.

## 6. Install everything

```bash
kit install backend
```

This pulls all artifacts from the `backend` namespace and writes them into the correct locations for Claude Code, Codex CLI, and Cursor.

## 7. Check what's installed

```bash
kit status
```

Shows what's installed locally, which tool each artifact targets, and whether anything is out of date.

## 8. Open the dashboard

Visit [http://localhost:8430/ui](http://localhost:8430/ui) in your browser to see your team's artifacts, manage access, and review installation status.
