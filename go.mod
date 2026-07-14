module github.com/trly/today

go 1.26.4

require (
	connectrpc.com/connect v1.20.0
	github.com/emersion/go-ical v0.0.0-20240127095438-fc1c9d8fb2b6
	github.com/emersion/go-webdav v0.7.0
	github.com/urfave/cli/v3 v3.10.1
	google.golang.org/protobuf v1.36.11
)

require github.com/teambition/rrule-go v1.8.2 // indirect

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)
