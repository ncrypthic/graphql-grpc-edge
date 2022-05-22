package generator

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func Generate() {
	var flags flag.FlagSet
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			filename := f.GeneratedFilenamePrefix + "_graphql.pb.go"
			gen := gen.NewGeneratedFile(filename, f.GoImportPath)
			v := NewVisitor(f, gen, f.GoImportPath.String())

			v.Visit(root, f)
		}
		return nil
	})
}
