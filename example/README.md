# gRPC grahql example

1. Run `go install github.com/ncrypthic/graphql-grpc-edge/protoc-gen-graphql` to install `protoc-gen-graphql` plugin
2. Run `MODULE=github.com/ncrypthic/graphql-grpc-edge/example/grpc go generate`
3. Run `go run main.go`
4. Open `localhost:8080` (Port can be changes on [main.go](main.go#L19))
