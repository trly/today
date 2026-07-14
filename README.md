# Today

## Development

Run the Go tests:

```sh
go test ./...
```

Run a focused test or package:

```sh
go test ./internal/httpapi -run TestHealth
```

Build the web app from the repository root:

```sh
pnpm build
```

Start the Vite development server:

```sh
pnpm dev
```

`pnpm dev` starts only the web app. Start the Go server separately when API access is needed:

```sh
go run ./cmd/today --url <caldav-url>
```

The server requires `--url` or `CALDAV_URL`. Basic authentication uses `--user`/`CALDAV_USER` and `--password`/`CALDAV_PASSWORD`; a password requires a user. The default address is `:8080`, configurable with `--addr` or `ADDR`.

## API Code Generation

After changing `protobuf/today/v1/today.proto`, regenerate the Go protobuf/Connect stubs, web TypeScript protobuf descriptors, and Markdown API documentation:

```sh
buf generate
```

This requires the `buf` CLI, the Go tool plugins, root Node dependencies for `protoc-gen-es`, and network access for the remote documentation plugin.
