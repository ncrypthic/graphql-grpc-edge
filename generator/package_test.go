package generator

import (
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func TestGenerate(t *testing.T) {
	b, _ := ioutil.ReadFile("../example/sample/sample.proto")
	fd := new(descriptor.FileDescriptorProto)
	err := fd.XXX_Unmarshal(b)
	if err != nil {
		t.Fatalf("Err: %s", err)
	}

	req := &plugin.CodeGeneratorRequest{
		FileToGenerate: []string{},
		ProtoFile:      []*descriptor.FileDescriptorProto{fd},
	}
	_, err = Generate(req)
	if err != nil {
		t.Fatalf("Err: %s", err)
	}
}
