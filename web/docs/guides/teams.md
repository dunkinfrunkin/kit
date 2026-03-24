---
sidebar_position: 3
---

# Teams

Teams are shared namespaces. No setup needed — they're auto-created on first push.

## Access model

- Anyone with a valid token can **read from** and **push to** team namespaces.
- **Personal namespaces** (`@email`) are encrypted and private.
- **Team namespaces** are plaintext — they're shared by definition.

## Push to a team

```bash
kit push ./deploy-k8s --team backend
kit push ./lint-rules.js --team frontend
kit push ./go-conventions.md --team platform
```

## List items

List your own items:

```bash
kit list
```

List a team's items:

```bash
kit list backend
```

## Search across namespaces

```bash
kit search deploy
```

## Install from a team

```bash
# Everything
kit install backend

# One item
kit install backend/deploy-k8s
```
