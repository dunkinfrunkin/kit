---
title: CLI Reference
sidebar_position: 1
---

# CLI Reference

## login

Authenticate with a kit server.

```
kit login --server <url>
```

Prompts for your email address, then stores credentials to `~/.kit/credentials`.

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--server` | Yes | Server URL (e.g., `http://localhost:8430`) |

**Example:**

```bash
kit login --server https://kit.mycompany.com
```

## logout

Remove stored credentials.

```
kit logout
```

Deletes `~/.kit/credentials`.

## whoami

Show the current authenticated user and server.

```
kit whoami
```

**Output:**

```
alice@example.com @ https://kit.mycompany.com
```

## push

Push a skill, hook, or config to the server.

```
kit push <path> [flags]
```

Kit auto-detects the item type from the file structure:
- Directory with `SKILL.md` inside = skill
- `.js` file = hook
- `.md` file = config

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | No | Push to a team namespace instead of your personal namespace |
| `--type` | No | Override auto-detected type (`skill`, `hook`, `config`) |

**Examples:**

```bash
# Push a skill to your personal namespace
kit push ./deploy-k8s

# Push a config to a team
kit push ./go-conventions.md --team backend

# Push a hook to a team
kit push ./pre-commit.js --team frontend
```

## install

Install skills, hooks, or configs from the server into local AI tool directories.

```
kit install <ref> [flags]
```

The `ref` can be a namespace (install everything), a specific item (`namespace/name`), or a profile name (with `-p`).

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--target` | No | Target tool (`claude`, `codex`, `cursor`) |
| `--project` | No | Install to project directory instead of user-level |
| `--skills` | No | Only install skills |
| `--hooks` | No | Only install hooks |
| `--configs` | No | Only install configs |
| `-p`, `--profile` | No | Install a profile |

**Examples:**

```bash
# Install everything from a team namespace
kit install backend

# Install a single item
kit install backend/deploy-k8s

# Install only skills from a namespace
kit install backend --skills

# Install a personal profile
kit install -p my-setup

# Install a team profile
kit install -p backend/default

# Install to a specific tool
kit install backend --target claude
```

## delete

Delete an item from the server and remove it locally.

```
kit delete <namespace/name>
```

Only the original author can delete an item.

**Example:**

```bash
kit delete backend/deploy-k8s
```

## list

List items on the server.

```
kit list [namespace] [flags]
```

Without a namespace, lists items across all accessible namespaces.

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--mine` | No | List only your personal items |

**Examples:**

```bash
# List everything you have access to
kit list

# List items in a team namespace
kit list backend

# List your personal items
kit list --mine
```

## search

Search for items by name across team namespaces.

```
kit search <query>
```

**Example:**

```bash
kit search deploy
```

## info

Show details about a specific item.

```
kit info <namespace/name>
```

**Example:**

```bash
kit info backend/deploy-k8s
```

**Output:**

```
Namespace: backend
Type:      skill
Name:      deploy-k8s
Author:    alice@example.com
Version:   3
```

## status

Show installed items and check for updates.

```
kit status
```

Compares locally installed versions against the server and shows which items are up-to-date, outdated, or unknown.

**Output:**

```
NAMESPACE  TYPE    NAME          INSTALLED  LATEST  STATUS
backend    skill   deploy-k8s    v2         v3      outdated
backend    config  go-conventions v1        v1      up-to-date
```

## update

Update installed items to their latest versions.

```
kit update [namespace]
```

Without a namespace, updates all installed items. With a namespace, updates only items from that namespace.

**Examples:**

```bash
# Update everything
kit update

# Update only items from a specific namespace
kit update backend
```

## sync

Update all installed items to latest versions. Equivalent to `kit update` with no arguments.

```
kit sync
```

## doctor

Check kit configuration and connectivity.

```
kit doctor
```

Verifies credentials, server connectivity, and detects installed AI tools.

**Output:**

```
Checking kit health...

[ OK ] Credentials: alice@example.com @ https://kit.mycompany.com
[ OK ] Server connectivity: https://kit.mycompany.com
[ OK ] Detected tool: claude
[ OK ] Detected tool: cursor
```

## profile create

Create a new profile.

```
kit profile create <name> [flags]
```

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | No | Create in a team namespace |

**Examples:**

```bash
# Personal profile
kit profile create my-setup

# Team profile
kit profile create default --team backend
```

## profile add

Add an item to a profile.

```
kit profile add <profile> <item-ref> [flags]
```

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | No | Use a team namespace |

**Examples:**

```bash
# Add an item to a personal profile
kit profile add my-setup deploy-k8s

# Add an item to a team profile
kit profile add default go-conventions --team backend
```

## profile list

List items in a profile.

```
kit profile list <name> [flags]
```

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--team` | No | Use a team namespace |

**Example:**

```bash
kit profile list my-setup
```

**Output:**

```
Profile: @alice@example.com/my-setup
  skill  deploy-k8s
  config go-conventions
```
