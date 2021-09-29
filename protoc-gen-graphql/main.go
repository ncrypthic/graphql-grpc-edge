package main

import (
	"flag"

	"github.com/ncrypthic/graphql-grpc-edge/generator"
)

type option struct {
	SourcePath      string
	DestinationPath string
}

func initFlags() *option {
	opt := option{}
	rawOpt := ""
	flag.StringVar(&rawOpt, "graphql_out", "", "GraphQL gRPC edge generator option")
	flag.Parse()
	return &opt
}

func main() {
	generator.Generate()
}
