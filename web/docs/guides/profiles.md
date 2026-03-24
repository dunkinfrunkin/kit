---
sidebar_position: 4
---

# Profiles

Profiles bundle multiple items for one-command install.

## Create a profile

Personal profile:

```bash
kit profile create my-setup
```

Team profile:

```bash
kit profile create default --team backend
```

## Add items to a profile

```bash
kit profile add my-setup deploy-k8s
kit profile add my-setup go-conventions
```

## List profile contents

```bash
kit profile list my-setup
```

## Install a profile

Personal profile:

```bash
kit install -p my-setup
```

Team profile:

```bash
kit install -p backend/default
```

## The "new laptop" flow

Three commands from zero to fully configured:

```bash
brew install kitctl/tap/kit
kit login
kit install -p my-setup
```

That's it. Every skill, hook, and config — installed across all your AI tools.
