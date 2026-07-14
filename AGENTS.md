# Repository Guidance

## Boundaries

- `cmd/today` wires the CalDAV client, event service, and HTTP server. `internal/caldav` handles provider access, `internal/events` projects and sorts events, and `internal/httpapi` exposes the Connect RPC API plus `/api-docs`.
- HTTP API tests should exercise the public surface through generated Connect clients, not handler internals.
- The web app is a Svelte 5/Vite workspace under `web/`; root scripts delegate to it. There is no repository lint or typecheck script.

## Generated API

- `protobuf/today/v1/today.proto` is the source of truth for Go protobuf/Connect stubs, web TypeScript protobuf descriptors, and `docs/api/today-v1.md`.
- After changing the proto, run `buf generate`; this requires the `buf` CLI, Go tool plugins, root Node dependencies for `protoc-gen-es`, and network access for the remote documentation plugin.
- Do not edit `gen/today/v1/**`, `web/src/gen/**`, or `docs/api/today-v1.md` directly. `/api-docs` serves the generated Markdown embedded from `docs/api`.

## Svelte Diagnostics

- For Svelte changes, prefer deterministic local diagnostics such as `npm run check` in `web/` or direct `svelte-check`/TypeScript commands when available.
- Do not rely on the Svelte MCP server as the source of truth for diagnostics; use it only for documentation or examples when needed.
