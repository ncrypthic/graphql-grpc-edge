package generator

import (
	"io"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/ncrypthic/graphql-edge/generator/funcs"
	protoparser "github.com/yoheimuta/go-protoparser"
)

func Generate(src io.Reader, dst io.Writer) error {
	g := NewGenerator(DefaultNameGenerator)
	proto, err := protoparser.Parse(src)
	if err != nil {
		log.Fatalf("Failed to parse input...")
		return err
	}
	if err = g.FromProto(proto); err != nil {
		log.Fatalf("Failed to parse input...")
		return err
	}

	tmpl := template.New("./template/graphql.go.tmpl")
	tmpl.Funcs(template.FuncMap{
		"lcfirst":       funcs.LCFirst,
		"ucfirst":       funcs.UCFirst,
		"concat":        funcs.Concat,
		"title":         funcs.Title,
		"lookUpMessage": funcs.LookUpMessage,
		"GetTypeInfo":   g.GetTypeInfo,
	})
	rawTmpl, err := ioutil.ReadFile("./template/graphql.go.tmpl")
	if err != nil {
		return err
	}

	tmpl, err = tmpl.Parse(string(rawTmpl))
	if err != nil {
		return err
	}
	return tmpl.Execute(dst, g)
}
