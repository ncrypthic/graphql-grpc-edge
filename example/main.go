package main

import (
	"log"
	"net"
	"net/http"

	graphql "github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/ncrypthic/graphql-edge/example/sample"
	grpc "google.golang.org/grpc"
)

func main() {
	// GRPC Server
	srv := sample.HelloServer{}
	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	sample.RegisterHelloServiceServer(grpcServer, &srv)
	go grpcServer.Serve(lis)

	// GraphQL Edge Server
	queries := graphql.Fields{}
	mutations := graphql.Fields{}
	types := make([]graphql.Type, 0)
	grpcClient, err := grpc.Dial(":9999", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to grpc server: %v", err)
	}
	sample.RegisterGraphQLTypes(types)
	sample.RegisterHelloServiceQueries(queries, sample.NewHelloServiceClient(grpcClient))
	sample.RegisterHelloServiceMutations(mutations, sample.NewHelloServiceClient(grpcClient))
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: queries}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mutations}
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
		Types:    types,
	}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
