//go:generate protoc --go_out=plugins=grpc:. --graphql_out=. -I=../../ -I . sample/sample.proto sample/test.proto
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/ncrypthic/graphql-grpc-edge/example/sample"
	"github.com/ncrypthic/graphql-grpc-edge/example/server"
	edge "github.com/ncrypthic/graphql-grpc-edge/graphql"
	grpc "google.golang.org/grpc"
)

const (
	GRPCPort string = ":9090"
	HTTPPort        = ":8080"
)

func main() {
	// GRPC Server
	srv := server.HelloServer{}
	lis, err := net.Listen("tcp", GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	sample.RegisterHelloServiceServer(grpcServer, &srv)
	sample.RegisterHelloTestServiceServer(grpcServer, &srv)
	go grpcServer.Serve(lis)

	// GraphQL Edge Server
	grpcClient, err := grpc.Dial(GRPCPort, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to grpc server: %v", err)
	}
	testClient := sample.NewHelloTestServiceClient(grpcClient)
	helloClient := sample.NewHelloServiceClient(grpcClient)
	sample.RegisterTestGraphQLTypes()
	sample.RegisterHelloTestServiceQueries(testClient)
	sample.RegisterHelloTestServiceMutations(testClient)
	sample.RegisterSampleGraphQLTypes()
	sample.RegisterHelloServiceQueries(helloClient)
	sample.RegisterHelloServiceMutations(helloClient)
	schema, err := edge.GetSchema()
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}
	h := handler.New(&handler.Config{
		Schema:   schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	fmt.Printf("GraphQL gRPC edge server running on %s\n", HTTPPort)
	http.ListenAndServe(HTTPPort, nil)
}
