package main

import (
	"log"
	"net/http"

	graphql "github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/ncrypthic/graphql-edge/example/sample"
)

func main() {
	queries := graphql.Fields{}
	mutations := graphql.Fields{}
	types := make([]graphql.Type, 0)
	sample.RegisterGraphQLTypes(types)
	sample.RegisterHelloServiceQueries(queries, sample.NewHelloServiceClient())
	sample.RegisterHelloServiceMutations(mutations, sample.NewHelloServiceClient())
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
