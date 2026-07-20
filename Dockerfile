FROM node:24-alpine AS frontend
RUN corepack enable
WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm run build

FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /today ./cmd/today

FROM scratch
COPY --from=builder /today /today
LABEL org.opencontainers.image.licenses="MIT"
EXPOSE 8080
ENTRYPOINT ["/today"]
