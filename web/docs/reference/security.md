---
title: Security
sidebar_position: 3
---

# Security

## Authentication model

Kit uses email + JWT. No passwords, no OAuth, no user table.

1. User sends their email to `POST /login`.
2. Server returns a JWT signed with `KIT_SECRET`.
3. All API requests include the JWT in the `Authorization: Bearer` header.
4. Server verifies the signature and extracts the email on each request.

This is a trust-based model suited for internal teams and self-hosted deployments. Anyone who can reach the server can get a token for any email address.

## Token storage

Credentials are stored at `~/.kit/credentials` with file permissions `0600` (owner read/write only). The file contains the server URL, email, and JWT token in JSON format.

```bash
# Verify permissions
ls -la ~/.kit/credentials
# -rw-------  1 user  staff  ... credentials
```

## Personal namespace encryption

Items in personal namespaces (`@email`) are encrypted at rest using AES-256-GCM.

**Key derivation:**

```
encryption_key = HMAC-SHA256(KIT_SECRET, user_email)
```

- Each user gets a unique 256-bit encryption key.
- The key is never stored -- it is derived per-request.
- Encryption and decryption happen server-side.

## What is encrypted

| Data | Encrypted | Reason |
|------|-----------|--------|
| Personal namespace content | Yes | Private to the user |
| Team namespace content | No | Shared by definition -- all team members need to read it |
| Item metadata (names, authors, versions) | No | Needed for listing and search |
| Profile data | No | References only, no content |

## Database exposure scenario

If the database is dumped or leaked:

- **Team items** are readable in plaintext. They are shared content by design.
- **Personal items** are encrypted blobs. Unreadable without `KIT_SECRET` and the user's email to derive the decryption key.
- **Metadata** (item names, authors, versions, timestamps) is visible for all items.

## KIT_SECRET

`KIT_SECRET` is the master key for the entire system. It protects:

1. **JWT signing** -- anyone with the secret can forge tokens for any user.
2. **Encryption key derivation** -- anyone with the secret and a user's email can decrypt their personal items.

Treat it like a private key:
- Generate it with `openssl rand -hex 32`.
- Store it in a secrets manager or environment variable, not in source control.
- Rotate it only if compromised (rotation invalidates all tokens and makes all encrypted personal data unreadable).

## Network security

Kit does not terminate TLS. In production, run a reverse proxy in front of the server:

```
Client --TLS--> nginx/caddy --HTTP--> kitd:8430
```

Example Caddy configuration:

```
kit.mycompany.com {
    reverse_proxy localhost:8430
}
```

This gives you automatic HTTPS with Let's Encrypt certificates. Without TLS, JWT tokens and item content are transmitted in plaintext.
