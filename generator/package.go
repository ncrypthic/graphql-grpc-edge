package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	protoparser "github.com/yoheimuta/go-protoparser"
)

func Generate(req *plugin.CodeGeneratorRequest) (res *plugin.CodeGeneratorResponse, err error) {
	res = &plugin.CodeGeneratorResponse{}
	for _, f := range req.GetFileToGenerate() {
		baseFileName, fileName := generateFilename(f)
		g := NewGenerator(DefaultNameGenerator, baseFileName)
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
		if ok, err := g.FromProto(proto); err != nil {
			res = withError(res, err)
			break
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

func generateFilename(s string) (base string, generatedFileName string) {
	ext := filepath.Ext(s)
	fileName := s[0 : len(s)-len(ext)]
	generatedFileName = fileName + ".graphql.pb.go"
	base = filepath.Base(s)
	base = base[0 : len(base)-len(ext)]
	return
}

func withError(res *plugin.CodeGeneratorResponse, err error) *plugin.CodeGeneratorResponse {
	msg := err.Error()
	res.Error = &msg
	return res
}
