package generator

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	protoparser "github.com/yoheimuta/go-protoparser"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func Generate() {
	var flags flag.FlagSet
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		options := GetOptions(gen.Request.GetParameter())
		for filePath, f := range gen.FilesByPath {
			if !f.Generate {
				continue
			}
			filename := f.GeneratedFilenamePrefix + "_graphql.pb.go"
			baseFileName := filepath.Base(f.GeneratedFilenamePrefix)
			g := NewGenerator(DefaultNameGenerator, baseFileName, options)
			src, err := os.Open(filePath)
			if err != nil {
				return err
			}
			proto, err := protoparser.Parse(src)
			if err != nil {
				return err
			}
			strImportPrefix := string(f.GoImportPath)
			importPrefix := ""
			if len(strImportPrefix) > 0 {
				importPrefix = strings.Trim(strImportPrefix, string(f.GoPackageName))
			}
			if ok, err := g.FromProto(proto, importPrefix, string(f.GoPackageName)); err != nil {
				return err
			} else if !ok {
				continue
			}

			tmpl := template.New("graphql_grpc_template")
			tmpl.Funcs(template.FuncMap{
				"GetFieldType":         g.GetFieldType,
				"GetProtobufFieldName": g.GetProtobufFieldName,
				"GetBaseType":          g.GetBaseType,
				"GetInputType":         g.GetInputType,
				"GetOutputType":        g.GetOutputType,
				"NormalizedFileName":   NormalizedFileName,
				"HasOperation":         g.HasOperation,
			})

			tmpl, err = tmpl.Parse(codeTemplate)
			if err != nil {
				return err
			}
			res := gen.NewGeneratedFile(filename, f.GoImportPath)
			err = tmpl.Execute(res, g)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func GetOptions(str string) []Option {
	options := make([]Option, 0)
	optionSet := strings.Split(str, ",")
	for _, opt := range optionSet {
		packageMap := strings.Split(opt, "=")
		if len(packageMap) != 2 {
			continue
		}
		options = append(options, NewOption(packageMap[0], packageMap[1]))
	}
	return options
}
