---
title: Introduction
sidebar_position: 1
---

# Introduction

Kit is an open-source CLI and server for sharing AI coding skills, hooks, and configuration across developers and teams. It works with Claude Code, Codex CLI, and Cursor.

## The problem

Every AI coding tool stores configuration differently. Claude Code uses `CLAUDE.md` files and `.claude/` directories. Codex CLI uses `codex.md` and its own settings. Cursor uses `.cursor/rules/`. None of them have a distribution story.

When you write a useful skill, coding convention, or hook, sharing it means copying files between repos, Slack messages, or wikis. Onboarding a new developer means manually setting up each tool. Switching laptops means doing it all over again.

## What kit does

Kit gives you a central server to **push** skills, hooks, and configs to, and a CLI to **install** them into any supported tool.

- **Push once, install everywhere.** Push a skill to the kit server. Install it into Claude Code, Codex CLI, or Cursor with one command.
- **Team namespaces.** Organize artifacts by team (`backend`, `frontend`, `infra`). Everyone on the team gets the same setup.
- **Personal encryption.** Private artifacts are encrypted at rest. Only you can decrypt them.
- **Profiles.** Bundle everything you need into a profile. New laptop? `kit install` and you're done.

## Quick example

```bash
# Push a coding skill to your team's namespace
kit push ./sql-conventions.md --team backend

# On another machine (or another developer's machine), install it
kit install backend

# Check what's installed and where
kit status
```
