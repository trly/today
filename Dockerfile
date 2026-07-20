FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /today ./cmd/today

FROM scratch
COPY --from=builder /today /today
LABEL org.opencontainers.image.licenses="MIT"
EXPOSE 8080
ENTRYPOINT ["/today"]
