package main

import (
	"github.com/ncrypthic/graphql-grpc-edge/generator"
)

type option struct {
	SourcePath      string
	DestinationPath string
}

func main() {
	generator.Generate()
}
