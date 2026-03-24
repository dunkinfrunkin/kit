---
sidebar_position: 2
---

# Installing Items

## Install everything from a team

```bash
kit install backend
```

This installs all skills, hooks, and configs from the `backend` team namespace.

## Install a specific item

```bash
kit install backend/deploy-k8s
```

## Filter by type

Install only the item types you need:

```bash
kit install backend --skills
kit install backend --hooks
kit install backend --configs
```

## Target a specific tool

By default, kit installs to all detected tools. To target one:

```bash
kit install backend --target claude
```

See [Multi-Tool Targeting](./multi-tool.md) for details.

## Install to the current project

```bash
kit install backend --project
```

This writes items to the project directory instead of the global config paths.

## Install from your personal namespace

```bash
kit install my-snippet
```

## Delete an item from the server

```bash
kit delete backend/deploy-k8s
```

## Check what's installed

```bash
kit status
```

Shows installed items, their sources, and whether any are outdated.

## Update outdated items

```bash
kit update
```

Pulls the latest versions for all installed items.
