//go:generate protoc --go_out=plugins=grpc:. --graphql_out=:. -I ../../ -I . common/shared.proto
//go:generate protoc --go_out=plugins=grpc,Mcommon/shared.proto=github.com/ncrypthic/graphql-grpc-edge/example/common:. --graphql_out=import_path=common,import_prefix=github.com/ncrypthic/graphql-grpc-edge/example/:. -I ../../ -I . sample/sample.proto
//go:generate protoc --go_out=plugins=grpc:. --graphql_out=:. -I ../../ -I . sample/test.proto
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/ncrypthic/graphql-grpc-edge/example/sample"
	"github.com/ncrypthic/graphql-grpc-edge/example/server"
	edge "github.com/ncrypthic/graphql-grpc-edge/graphql"
	"github.com/opentracing/opentracing-go"
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
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
		grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer())),
	)
	sample.RegisterHelloServiceServer(grpcServer, &srv)
	sample.RegisterHelloTestServiceServer(grpcServer, &srv)
	go grpcServer.Serve(lis)

	// GraphQL Edge Server
	grpcClient, err := grpc.Dial(GRPCPort,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		log.Fatalf("failed to connect to grpc server: %v", err)
	}
	testClient := sample.NewHelloTestServiceClient(grpcClient)
	sample.RegisterTestGraphQLTypes()
	sample.RegisterHelloTestServiceQueries(testClient)
	sample.RegisterHelloTestServiceMutations(testClient)

	helloClient := sample.NewHelloServiceClient(grpcClient)
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

	http.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "entrypoint")
		defer span.Finish()
		h.ContextHandler(ctx, w, req)
	})
	fmt.Printf("GraphQL gRPC edge server running on %s\n", HTTPPort)
	http.ListenAndServe(HTTPPort, nil)
}
