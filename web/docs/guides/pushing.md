---
sidebar_position: 1
---

# Pushing Items

Kit handles three item types:

| Type | Structure | Example |
|------|-----------|---------|
| **Skill** | Directory containing a `SKILL.md` | `./deploy-k8s/SKILL.md` |
| **Hook** | JavaScript file (`.js`) | `./pre-commit.js` |
| **Config** | Markdown file (`.md`) | `./go-conventions.md` |

Kit auto-detects the type from the file structure and extension. Use `--type` to override if needed:

```bash
kit push ./ambiguous-item --type skill
```

## Push to your personal namespace

By default, items push to your personal namespace (`@you`):

```bash
kit push ./deploy-k8s
```

## Push to a team

```bash
kit push ./deploy-k8s --team backend
```

## Naming rules

The item name is derived automatically:

- **Pushing a directory** — name comes from the directory name:
  ```bash
  kit push ./deploy-k8s
  # → name: deploy-k8s
  ```

- **Pushing a `SKILL.md` directly** — name comes from the parent directory:
  ```bash
  kit push ./deploy-k8s/SKILL.md
  # → name: deploy-k8s
  ```

## Examples by type

### Skill

```bash
# Directory with SKILL.md inside
kit push ./deploy-k8s --team backend
```

### Hook

```bash
# Single .js file
kit push ./pre-commit.js --team backend
```

### Config

```bash
# Single .md file
kit push ./go-conventions.md --team backend
```
