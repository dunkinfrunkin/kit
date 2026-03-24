---
sidebar_position: 5
---

# Multi-Tool Targeting

Kit auto-detects which AI coding tools are installed and writes to all of them.

## Install path mapping

| Type | Claude Code | Codex | Cursor |
|------|-------------|-------|--------|
| **Skills** | `~/.claude/skills/{name}/SKILL.md` | `~/.codex/skills/{name}/SKILL.md` | N/A |
| **Hooks** | Merged into `settings.json` | Merged into `config.toml` | N/A |
| **Configs** | Appended to `CLAUDE.md` | Appended to `AGENTS.md` | `.cursor/rules/{name}.mdc` |

## Config translation

Kit stores configs as plain markdown. On install, it adapts the format to each tool:

**Claude Code and Codex** — content is wrapped in marker comments:

```markdown
<!-- kit:go-conventions -->
Use table-driven tests. Always handle errors explicitly.
<!-- /kit:go-conventions -->
```

These markers let `kit update` and `kit uninstall` find and replace managed sections without touching your other content.

**Cursor** — content is written as a `.mdc` file with frontmatter:

```markdown
---
description: Go conventions
globs:
alwaysApply: true
---

Use table-driven tests. Always handle errors explicitly.
```

## Target a single tool

Skip auto-detection and install to one tool only:

```bash
kit install backend --target claude
kit install backend --target codex
kit install backend --target cursor
```

## Project-level vs global install

By default, kit installs to global config paths (`~/.claude/`, etc.).

Use `--project` to install into the current project directory instead:

```bash
kit install backend --project
```

This writes configs to the project root (`CLAUDE.md`, `AGENTS.md`, `.cursor/rules/`) and skills to a local `.kit/skills/` directory, keeping your global config clean.
