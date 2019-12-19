package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
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
	res := &plugin.CodeGeneratorResponse{}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		write(res, err)
		return
	}

	var req plugin.CodeGeneratorRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		write(res, err)
		return
	}

	if len(req.FileToGenerate) == 0 {
		write(res, errors.New("No file to generate"))
		return
	}

	res, err = generator.Generate(&req)
	if err != nil {
		write(res, err)
		return
	}
	write(res, nil)
}

func write(res *plugin.CodeGeneratorResponse, err error) {
	if err != nil {
		errMsg := err.Error()
		res = &plugin.CodeGeneratorResponse{Error: &errMsg}
	}
	resData, _ := proto.Marshal(res)
	os.Stdout.Write(resData)
}
