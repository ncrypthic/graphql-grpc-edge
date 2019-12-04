package generator

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	protoFile, _ := os.Open("../../proto/sample.proto")

	err := Generate(protoFile)
	if err != nil {
		t.Fatalf("Err: %s", err)
	}
}
