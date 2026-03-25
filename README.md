# kit

An open-source CLI + server for sharing AI coding skills, hooks, and config across developers and teams. Supports Claude Code, Codex CLI, and Cursor.

## Quick Start

### Server (Docker)

```bash
docker run -d \
  -p 80:80 \
  -e KIT_SECRET=your-secret-here \
  -e DATABASE_URL=postgres://user:pass@host:5432/kit \
  ghcr.io/dunkinfrunkin/kit:latest
```

Or with Docker Compose:

```yaml
services:
  db:
    image: postgres:17
    environment:
      POSTGRES_DB: kit
      POSTGRES_USER: kit
      POSTGRES_PASSWORD: changeme
    volumes:
      - kit-db:/var/lib/postgresql/data

  server:
    image: ghcr.io/dunkinfrunkin/kit:latest
    ports:
      - "80:80"
    environment:
      KIT_SECRET: ${KIT_SECRET:-change-me-to-a-real-secret}
      DATABASE_URL: postgres://kit:changeme@db:5432/kit?sslmode=disable
      KIT_OIDC_ISSUER: ${KIT_OIDC_ISSUER:-}
      KIT_OIDC_CLIENT_ID: ${KIT_OIDC_CLIENT_ID:-}
      KIT_OIDC_CLIENT_SECRET: ${KIT_OIDC_CLIENT_SECRET:-}
    depends_on:
      - db

volumes:
  kit-db:
```

### CLI

```bash
brew install dunkinfrunkin/tap/kit
```

### Login

```bash
# Email auth (simple/personal)
kit login --server https://kit.yourcompany.com

# SSO (enterprise)
kit login --server https://kit.yourcompany.com --sso
```

## SSO Setup

### Okta

1. In Okta Admin, create a new App Integration:
   - Sign-in method: OIDC
   - Application type: Native Application
   - Grant type: Authorization Code
   - Sign-in redirect URI: `http://localhost:9876/callback`
   - Assignments: assign to users/groups

2. Set environment variables on the server:
```bash
KIT_OIDC_ISSUER=https://yourcompany.okta.com/oauth2/default
KIT_OIDC_CLIENT_ID=0oa1b2c3d4e5f6g7h8
```

3. Users login with:
```bash
kit login --server https://kit.yourcompany.com --sso
```

### Google Workspace

1. In Google Cloud Console, create OAuth 2.0 credentials:
   - Application type: Desktop app (or Web application)
   - Authorized redirect URI: `http://localhost:9876/callback`

2. Set environment variables on the server:
```bash
KIT_OIDC_ISSUER=https://accounts.google.com
KIT_OIDC_CLIENT_ID=your-client-id.apps.googleusercontent.com
KIT_OIDC_CLIENT_SECRET=your-client-secret
```

3. Users login with:
```bash
kit login --server https://kit.yourcompany.com --sso
```

### Azure AD / Entra ID

1. Register an application in Azure Portal:
   - Redirect URI: `http://localhost:9876/callback` (Mobile and desktop applications)
   - API permissions: openid, email, profile

2. Set environment variables:
```bash
KIT_OIDC_ISSUER=https://login.microsoftonline.com/{tenant-id}/v2.0
KIT_OIDC_CLIENT_ID=your-application-client-id
KIT_OIDC_CLIENT_SECRET=your-client-secret
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `KIT_SECRET` | Yes | Secret key for JWT signing and encryption. Keep secure. |
| `DATABASE_URL` | Yes | Postgres connection string |
| `PORT` | No | HTTP port (default: `80`) |
| `KIT_OIDC_ISSUER` | No | OIDC issuer URL for SSO |
| `KIT_OIDC_CLIENT_ID` | No | OIDC client ID for SSO |
| `KIT_OIDC_CLIENT_SECRET` | No | OIDC client secret (required for Google) |

## Usage

```bash
# Push a skill
kit push ./my-skill/ --team backend

# Push a config
kit push ./conventions.md --team backend

# Install everything from a team
kit install backend

# List items
kit list backend

# Delete an item
kit delete backend/my-skill

# Check status
kit status

# Create API token (for CI/CD)
kit token create ci-deploy
```

## Dashboard

The server includes a web dashboard at the root URL. When SSO is configured, users can sign in directly from the browser.

## License

MIT
