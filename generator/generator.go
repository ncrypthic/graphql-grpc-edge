package generator

import (
	"strings"

	"github.com/ncrypthic/graphql-grpc-edge/generator/funcs"
	protoparser "github.com/yoheimuta/go-protoparser"
	"github.com/yoheimuta/go-protoparser/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/parser"
)

const (
	ContextPackage    string = "package"
	ContextMessage           = "message"
	ContextService           = "service"
	ContextRPC               = "rpc"
	ContextRPCInput          = "rpcInput"
	ContenxtRPCOutput        = "rpcOutput"
	ContextOption            = "option"
	ContextEnum              = "enum"

	OptionGraphQL string = "(graphql.type)"
)

//TypeNameGenerator is function type to generate GraphQL type
//from protobuf object type
type TypeNameGenerator func(packageName, typeName string) string

//DefaultNameGenerator is a default type name generator
func DefaultNameGenerator(packageName, typeName string) string {
	return strings.Title(packageName + typeName)
}

type TypeInfo struct {
	Name       string
	IsScalar   bool
	IsRepeated bool
	IsEnum     bool
}

//Generator is an interface of graphql code generator
type Generator interface {
	FromProto(*parser.Proto) error
	GetTypeInfo(*unordered.Message, *parser.Field) *TypeInfo
	GetFieldName(string) string
}

type generator struct {
	TypeNameGenerator TypeNameGenerator
	PackageName       string
	Enums             map[string]*unordered.Enum
	Objects           map[string]*unordered.Message
	Unions            map[string]*parser.Oneof
	Inputs            map[string]*unordered.Message
	Queries           map[string]*parser.RPC
	Mutations         map[string]*parser.RPC
	Services          []*unordered.Service
}

func NewGenerator(typeNameGenerator TypeNameGenerator) Generator {
	return &generator{
		TypeNameGenerator: typeNameGenerator,
		Enums:             make(map[string]*unordered.Enum),
		Inputs:            make(map[string]*unordered.Message),
		Objects:           make(map[string]*unordered.Message),
		Unions:            make(map[string]*parser.Oneof),
		Queries:           make(map[string]*parser.RPC),
		Mutations:         make(map[string]*parser.RPC),
	}
}

func (g *generator) FromProto(p *parser.Proto) error {
	proto, err := protoparser.UnorderedInterpret(p)
	if err != nil {
		return err
	}
	g.PackageName = proto.ProtoBody.Packages[0].Name
	g.Services = make([]*unordered.Service, len(proto.ProtoBody.Services))
	for i, svc := range proto.ProtoBody.Services {
		g.Services[i] = svc
		for _, rpc := range svc.ServiceBody.RPCs {
			for _, opt := range rpc.Options {
				if opt.OptionName != "(graphql.type)" {
					continue
				}
				val := opt.Constant[1 : len(opt.Constant)-1]
				segments := strings.Split(val, ":")
				op := segments[0]
				opName := strings.ReplaceAll(segments[1], "\\\"", "\"")[1 : len(segments[1])-1]
				if op == "query" {
					g.Queries[opName] = rpc
				} else if op == "mutation" {
					g.Mutations[opName] = rpc
				}
				if _, ok := g.Inputs[rpc.RPCRequest.MessageType]; !ok {
					g.Inputs[rpc.RPCRequest.MessageType] = funcs.LookUpMessage(rpc.RPCRequest.MessageType, proto.ProtoBody)
				}
				if _, ok := g.Objects[rpc.RPCResponse.MessageType]; !ok {
					g.Objects[rpc.RPCResponse.MessageType] = funcs.LookUpMessage(rpc.RPCResponse.MessageType, proto.ProtoBody)
				}
			}
		}
	}
	for _, msg := range proto.ProtoBody.Messages {
		if _, ok := g.Inputs[msg.MessageName]; !ok {
			g.Inputs[msg.MessageName] = msg
		}
		if _, ok := g.Objects[msg.MessageName]; !ok {
			g.Objects[msg.MessageName] = msg
		}
	}
	for _, msg := range proto.ProtoBody.Messages {
		for _, union := range msg.MessageBody.Oneofs {
			g.Unions[msg.MessageName+"_"+union.OneofName] = union
		}
		for _, enum := range msg.MessageBody.Enums {
			g.Enums[msg.MessageName+"_"+enum.EnumName] = enum
		}
	}
	for _, enum := range proto.ProtoBody.Enums {
		g.Enums[enum.EnumName] = enum
	}

	return nil
}

func (g *generator) GetTypeInfo(msg *unordered.Message, field *parser.Field) *TypeInfo {
	switch field.Type {
	case "float":
		fallthrough
	case "double":
		return &TypeInfo{Name: "graphql.Float", IsScalar: true, IsRepeated: field.IsRepeated}
	case "int32":
		fallthrough
	case "int64":
		fallthrough
	case "uint32":
		fallthrough
	case "uint64":
		fallthrough
	case "sint32":
		fallthrough
	case "sint64":
		fallthrough
	case "fixed32":
		fallthrough
	case "fixed64":
		fallthrough
	case "sfixed32":
		fallthrough
	case "sfixed64":
		return &TypeInfo{Name: "graphql.Int", IsScalar: true, IsRepeated: field.IsRepeated}
	case "bool":
		return &TypeInfo{Name: "graphql.Boolean", IsScalar: true, IsRepeated: field.IsRepeated}
	case "string":
		return &TypeInfo{Name: "graphql.String", IsScalar: true, IsRepeated: field.IsRepeated}
	default:
		info := &TypeInfo{Name: field.Type, IsScalar: false, IsRepeated: field.IsRepeated}
		_, isEnum := g.Enums[msg.MessageName+"_"+field.Type]
		if !isEnum {
			_, isEnum = g.Enums[field.Type]
		}
		info.IsEnum = isEnum
		return info
	}
}

func (g *generator) GetFieldName(s string) string {
	segments := strings.Split(s, "_")
	res := ""
	for _, segment := range segments {
		res = res + strings.Title(segment)
	}
	return res
}

const (
	codeTemplate string = `// DO NOT EDIT! This file is autogenerated by graphql-grpc-edge generator
package {{.PackageName}}

import (
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	graphql "github.com/graphql-go/graphql"
)
{{ range $input := .Inputs }}
var GraphQL_Input{{$input.MessageName}} *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "Input{{$input.MessageName}}",
	Fields: graphql.InputObjectConfigFieldMap{
		{{- range $field := $input.MessageBody.Fields -}}
		{{ $type := GetTypeInfo $input $field }}
		"{{$field.FieldName}}": &graphql.InputObjectFieldConfig{
			Type: {{ if $type.IsRepeated -}} graphql.NewList( {{- end -}}
				{{- if $type.IsScalar -}} {{- $type.Name -}}
				{{- else if $type.IsEnum -}} GraphQL_{{- $type.Name -}}Enum
				{{- else -}} GraphQL_Input{{- $type.Name -}} {{- end -}}
			{{- if $type.IsRepeated -}} ) {{- end -}},
		},
		{{- end }}
	},
})

{{ end }}
{{- range $output := .Objects }}
var GraphQL_{{$output.MessageName}} *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "{{$output.MessageName}}",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		{{- range $field := $output.MessageBody.Fields -}}
		{{ $type := GetTypeInfo $output $field }}
		"{{$field.FieldName}}": &graphql.Field{
			Name: "{{ $field.FieldName }}",
			Type: {{if $type.IsRepeated -}} graphql.NewList( {{- end -}}
				{{- if $type.IsScalar -}} {{- $type.Name -}}
				{{- else if $type.IsEnum -}} GraphQL_{{- $type.Name -}}Enum
				{{- else -}} GraphQL_{{- $type.Name -}} {{- end -}}
				{{- if $type.IsRepeated -}} ) {{- end }},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if pdata, ok := p.Source.(*{{$output.MessageName}}); ok {
					return pdata.{{GetFieldName $field.FieldName}}, nil
				} else if data, ok := p.Source.({{$output.MessageName}}); ok {
					return data.{{GetFieldName $field.FieldName}}, nil
				}
				return nil, nil
			},
		},
		{{- end }}
		{{- range $union := $output.MessageBody.Oneofs -}}
		"{{$union.OneofName}}": &graphql.Field{
			Name: "{{ $union.OneofName }}",
			Type: GraphQL_{{ $output.MessageName }}_{{ $union.OneofName }}Union,
		},
		{{- end }}
	},
})

{{ end }}
{{- range $name,$enum := .Enums }}
var GraphQL_{{$enum.EnumName}}Enum *graphql.Enum = graphql.NewEnum(graphql.EnumConfig{
	Name: "{{$enum.EnumName}}Enum",
	Values: graphql.EnumValueConfigMap{
		{{- range $enumField := $enum.EnumBody.EnumFields }}
		"{{$enumField.Ident}}": &graphql.EnumValueConfig{
			Value: {{$name}}_value["{{$enumField.Ident}}"],
		},
		{{- end }}
	},
})
{{ end }}
{{- range $name,$union := .Unions }}
var GraphQL_{{$name}}Union *graphql.Union = graphql.NewUnion(graphql.UnionConfig{
	Name: "{{$name}}Union",
	Types: []*graphql.Object{
		{{- range $unionField := $union.OneofFields }}
		GraphQL_{{$unionField.Type}},
		{{- end }}
	},
})
{{ end }}

{{- $queries := .Queries -}}
{{- $mutations := .Mutations -}}
{{ range $svc := .Services }}
func Register{{$svc.ServiceName}}Queries(queries graphql.Fields, sc {{$svc.ServiceName}}Client) error {
	{{- range $name,$query := $queries }}
	queries["{{$name}}"] = &graphql.Field{
		Name: "{{$name}}",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: GraphQL_Input{{$query.RPCRequest.MessageType}},
			},
		},
		Type: GraphQL_{{$query.RPCResponse.MessageType}},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var req {{$query.RPCRequest.MessageType}}
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			jsonpb.UnmarshalString(string(rawJson), &req)
			if err != nil {
				return nil, err
			}
			return sc.{{GetFieldName $name}}(p.Context, &req)
		},
	}
	{{- end }}

	return nil
}

func Register{{$svc.ServiceName}}Mutations(mutations graphql.Fields, sc {{$svc.ServiceName}}Client) error {
	{{- range $name,$mutation := $mutations }}
	mutations["{{$name}}"] = &graphql.Field{
		Name: "{{$name}}",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: GraphQL_Input{{$mutation.RPCRequest.MessageType}},
			},
		},
		Type: GraphQL_{{$mutation.RPCResponse.MessageType}},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var req {{$mutation.RPCRequest.MessageType}}
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			jsonpb.UnmarshalString(string(rawJson), &req)
			if err != nil {
				return nil, err
			}
			return sc.{{GetFieldName $name}}(p.Context, &req)
		},
	}
	{{- end }}
	return nil
}
{{ end }}

func RegisterGraphQLTypes(types []graphql.Type) {
	{{- range $input := .Inputs }}
	types = append(types, GraphQL_Input{{$input.MessageName}})
	{{- end }}
	{{- range $output := .Objects }}
	types = append(types, GraphQL_{{$output.MessageName}})
	{{- end }}
	{{- range $name,$enum := .Enums }}
	types = append(types, GraphQL_{{$enum.EnumName}}Enum)
	{{- end }}
	{{- range $name,$union := .Unions }}
	types = append(types, GraphQL_{{$name}}Union)
	{{- end }}
}`
)
