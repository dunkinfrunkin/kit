---
title: API Reference
sidebar_position: 3
---

# API Reference

All endpoints return JSON. All endpoints except `/login` and `/ui` require an `Authorization: Bearer {token}` header.

## Authentication

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `POST` | `/login` | Get a JWT token | No |

**Request body:**

```json
{"email": "alice@example.com"}
```

**Response:**

```json
{"token": "eyJhbGci...", "email": "alice@example.com"}
```

## Items

### List all items of a type

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/skills` | List all skills | Yes |
| `GET` | `/hooks` | List all hooks | Yes |
| `GET` | `/configs` | List all configs | Yes |

Returns items from all team namespaces plus the authenticated user's personal namespace. Other users' personal items are excluded.

### List items in a namespace

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/{namespace}/skills` | List skills in a namespace | Yes |
| `GET` | `/{namespace}/hooks` | List hooks in a namespace | Yes |
| `GET` | `/{namespace}/configs` | List configs in a namespace | Yes |

### Get a specific item

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/{namespace}/skills/{name}` | Get a skill | Yes |
| `GET` | `/{namespace}/hooks/{name}` | Get a hook | Yes |
| `GET` | `/{namespace}/configs/{name}` | Get a config | Yes |

Response includes `content` as a base64-encoded string. Personal namespace items are decrypted server-side before encoding.

### Push an item

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `POST` | `/{namespace}/skills` | Push a skill | Yes |
| `POST` | `/{namespace}/hooks` | Push a hook | Yes |
| `POST` | `/{namespace}/configs` | Push a config | Yes |

**Request body:**

```json
{"name": "deploy-k8s", "content": "base64-encoded-content"}
```

If the item already exists, it is updated and the version is incremented.

### Delete an item

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `DELETE` | `/{namespace}/skills/{name}` | Delete a skill | Yes |
| `DELETE` | `/{namespace}/hooks/{name}` | Delete a hook | Yes |
| `DELETE` | `/{namespace}/configs/{name}` | Delete a config | Yes |

Only the original author can delete an item.

## Search

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/search?q={query}` | Search items by name | Yes |

Searches across team namespaces only (personal namespaces are excluded from search results).

## Profiles

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/{namespace}/profiles` | List profiles in a namespace | Yes |
| `POST` | `/{namespace}/profiles` | Create a profile | Yes |
| `GET` | `/{namespace}/profiles/{name}` | Get a profile | Yes |
| `DELETE` | `/{namespace}/profiles/{name}` | Delete a profile (author only) | Yes |
| `POST` | `/{namespace}/profiles/{name}/items` | Add an item to a profile | Yes |

**Create profile body:**

```json
{"name": "my-setup"}
```

**Add profile item body:**

```json
{"name": "deploy-k8s", "type": "skill"}
```

## Web dashboard

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/ui` | Web dashboard | No |

## Personal namespaces

Namespaces starting with `@` are personal namespaces (e.g., `@alice@example.com`). They are:

- **Access-controlled** -- only the matching authenticated user can read or write.
- **Encrypted** -- content is encrypted at rest with AES-256-GCM using a key derived from the user's email and `KIT_SECRET`.
