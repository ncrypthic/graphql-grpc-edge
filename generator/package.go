package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/ncrypthic/graphql-grpc-edge/generator/funcs"
	protoparser "github.com/yoheimuta/go-protoparser"
)

func Generate(req *plugin.CodeGeneratorRequest) (res *plugin.CodeGeneratorResponse, err error) {
	res = &plugin.CodeGeneratorResponse{}
	for _, f := range req.GetFileToGenerate() {
		fileName := generateFilename(f)
		g := NewGenerator(DefaultNameGenerator)
		src, err := os.Open(f)
		if err != nil {
			res = withError(res, err)
			break
		}
		proto, err := protoparser.Parse(src)
		if err != nil {
			res = withError(res, err)
			break
		}
		if err = g.FromProto(proto); err != nil {
			res = withError(res, err)
			break
		}

		tmpl := template.New("graphql_grpc_template")
		tmpl.Funcs(template.FuncMap{
			"lcfirst":       funcs.LCFirst,
			"ucfirst":       funcs.UCFirst,
			"concat":        funcs.Concat,
			"lookUpMessage": funcs.LookUpMessage,
			"GetTypeInfo":   g.GetTypeInfo,
			"GetFieldName":  g.GetFieldName,
		})

		tmpl, err = tmpl.Parse(codeTemplate)
		if err != nil {
			res = withError(res, err)
			break
		}
		bb := bytes.NewBuffer(nil)
		err = tmpl.Execute(bb, g)
		if err != nil {
			res = withError(res, err)
			break
		}
		output, err := ioutil.ReadAll(bb)
		if err != nil {
			res = withError(res, err)
			break
		}
		content := string(output)
		res.File = append(res.File, &plugin.CodeGeneratorResponse_File{
			Name:           &fileName,
			InsertionPoint: nil,
			Content:        &content,
		})
	}
	return res, err
}

func generateFilename(s string) string {
	ext := filepath.Ext(s)
	base := s[0 : len(s)-len(ext)]
	return base + ".graphql.pb.go"
}

func withError(res *plugin.CodeGeneratorResponse, err error) *plugin.CodeGeneratorResponse {
	msg := err.Error()
	res.Error = &msg
	return res
}
