# Today

A lightweight CalDAV viewer.

Today is intended to be an easy way to display 1 or more CalDav calendars as a semi-interactive dashboard.

## Quick Start

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `CALDAV_URL` | **yes** | — | CalDAV server endpoint |
| `CALDAV_USER` | no | — | HTTP basic-auth username |
| `CALDAV_PASSWORD` | no | — | HTTP basic-auth password (requires user) |
| `ADDR` | no | `:8080` | Listen address |

### Docker

Published to `ghcr.io/trly/today` on every `v*` tag:

```sh
docker run --rm -p 8080:8080 \
  -e CALDAV_URL=https://caldav.icloud.com \
  -e CALDAV_USER=you@example.com \
  -e CALDAV_PASSWORD=secret \
  ghcr.io/trly/today:latest
```

### Docker Compose

```yaml
services:
  today:
    image: ghcr.io/trly/today:latest
    ports:
      - "8080:8080"
    environment:
      CALDAV_URL: https://caldav.icloud.com
      CALDAV_USER: you@example.com
      CALDAV_PASSWORD: secret
```

```sh
docker compose up
```

## Local Development

Requires Go 1.26+, Node 24+, and pnpm.

```sh
# run tests
go test ./...

# start the API server
go run ./cmd/today --url https://caldav.icloud.com --user you@example.com --password secret

# start the web dev server
pnpm -C web dev
```

The web app (`web/`) runs on Vite (`:5173`); the Go server exposes the Connect RPC API on `:8080`. API requests are proxied via through vite during development.

### API Code Generation

After editing `protobuf/today/v1/today.proto`:

```sh
buf generate
```

Requires the `buf` CLI, Go protobuf plugins, and root Node dependencies.
