package generator

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	protoFile, _ := os.Open("../example/proto/sample.proto")
	dst, _ := os.Create("../example/sample/sample.edge.go")

	err := Generate(protoFile, dst)
	if err != nil {
		t.Fatalf("Err: %s", err)
	}
}
