FROM node:24-alpine AS frontend
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY web/package.json ./web/
RUN pnpm install --frozen-lockfile
COPY web/ ./web/
RUN pnpm --filter web build

FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /today ./cmd/today

FROM gcr.io/distroless/static-debian12
COPY --from=builder /today /today
LABEL org.opencontainers.image.licenses="MIT"
EXPOSE 8080
ENTRYPOINT ["/today"]
