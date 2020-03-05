package generator

import (
	"fmt"
	"strings"

	protoparser "github.com/yoheimuta/go-protoparser"
	"github.com/yoheimuta/go-protoparser/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/parser"
)

const (
	OptionGraphQL string = "(graphql.type)"
)

var (
	wellKnownTypes_imports map[string]string = map[string]string{
		"google/protobuf/empty.proto":             `pbEmpty "github.com/golang/protobuf/ptypes/empty"`,
		"google/protobuf/timestamp.proto":         `pbTimestamp "github.com/golang/protobuf/ptypes/timestamp"`,
		"google/protobuf/duration.proto":          `pbDuration "github.com/golang/protobuf/ptypes/duration"`,
		"google/protobuf/wrappers.proto":          `pbWrappers "github.com/golang/protobuf/ptypes/wrappers"`,
		"graphql-grpc-edge/graphql/graphql.proto": `pbGraphql "github.com/ncrypthic/graphql-grpc-edge/graphql"`,
	}

	wellKnownTypes_input_map map[string]string = map[string]string{
		"google.protobuf.Empty":       "pbGraphql.GraphQL_Empty",
		"google.protobuf.Timestamp":   "pbGraphql.GraphQL_Timestamp",
		"google.protobuf.Duration":    "pbGraphql.GraphQL_Duration",
		"google.protobuf.StringValue": "pbGraphql.GraphQL_StringValueInput",
		"google.protobuf.FloatValue":  "pbGraphql.GraphQL_FloatValueInput",
		"google.protobuf.DoubleValue": "pbGraphql.GraphQL_DoubleValueInput",
		"google.protobuf.BoolValue":   "pbGraphql.GraphQL_BoolValueInput",
		"google.protobuf.Int32Value":  "pbGraphql.GraphQL_Int32ValueInput",
		"google.protobuf.UInt32Value": "pbGraphql.GraphQL_UInt32ValueInput",
		"google.protobuf.Int64Value":  "pbGraphql.GraphQL_Int64ValueInput",
		"google.protobuf.UInt64Value": "pbGraphql.GraphQL_UInt64ValueInput",
	}

	wellKnownTypes_map map[string]string = map[string]string{
		"google.protobuf.Empty":       "pbGraphql.GraphQL_Empty",
		"google.protobuf.Timestamp":   "pbGraphql.GraphQL_Timestamp",
		"google.protobuf.Duration":    "pbGraphql.GraphQL_Duration",
		"google.protobuf.StringValue": "pbGraphql.GraphQL_StringValue",
		"google.protobuf.FloatValue":  "pbGraphql.GraphQL_FloatValue",
		"google.protobuf.DoubleValue": "pbGraphql.GraphQL_DoubleValue",
		"google.protobuf.BoolValue":   "pbGraphql.GraphQL_BoolValue",
		"google.protobuf.Int32Value":  "pbGraphql.GraphQL_Int32Value",
		"google.protobuf.UInt32Value": "pbGraphql.GraphQL_UInt32Value",
		"google.protobuf.Int64Value":  "pbGraphql.GraphQL_Int64Value",
		"google.protobuf.UInt64Value": "pbGraphql.GraphQL_UInt64Value",
	}

	wellKnownTypes_base map[string]string = map[string]string{
		"google.protobuf.Empty":       "pbEmpty.Empty",
		"google.protobuf.Timestamp":   "pbTimestamp.Timestamp",
		"google.protobuf.Duration":    "pbDuration.Duration",
		"google.protobuf.StringValue": "pbWrappers.StringValue",
		"google.protobuf.FloatValue":  "pbWrappers.FloatValue",
		"google.protobuf.DoubleValue": "pbWrappers.DoubleValue",
		"google.protobuf.BoolValue":   "pbWrappers.BoolValue",
		"google.protobuf.Int32Value":  "pbWrappers.Int32Value",
		"google.protobuf.UInt32Value": "pbWrappers.UInt32Value",
		"google.protobuf.Int64Value":  "pbWrappers.Int64Value",
		"google.protobuf.UInt64Value": "pbWrappers.UInt64Value",
	}
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
	Prefix     string
	Suffix     string
	IsScalar   bool
	IsRepeated bool
	IsEnum     bool
	IsNonNull  bool
}

func (t *TypeInfo) GetName() string {
	return t.formatRepeated()
}

func (t *TypeInfo) formatRepeated() string {
	if t.IsRepeated {
		return fmt.Sprintf(`graphql.NewList(%s)`, t.formatNonNull())
	} else {
		return t.formatNonNull()
	}
}

func (t *TypeInfo) formatNonNull() string {
	if t.IsNonNull {
		return fmt.Sprintf(`graphql.NewNonNull(%s)`, t.formatName())
	} else {
		return t.formatName()
	}
}

func (t *TypeInfo) formatName() string {
	if t.IsScalar {
		return t.Prefix + "." + t.Name
	}
	if t.Prefix != "" {
		return "GraphQL_" + t.Prefix + "." + t.Name + t.Suffix
	}
	return "GraphQL_" + t.Name + t.Suffix
}

//Generator is an interface of graphql code generator
type Generator interface {
	FromProto(*parser.Proto) (bool, error)
	GetFieldType(*unordered.Message, *parser.Field, string) *TypeInfo
	GetProtobufFieldName(string) string
	GetBaseType(string) string
	GetInputType(string) string
	GetOutputType(string) string
}

type generator struct {
	TypeNameGenerator TypeNameGenerator
	PackageName       string
	BaseFileName      string
	Enums             map[string]*unordered.Enum
	Objects           map[string]*unordered.Message
	Unions            map[string]*parser.Oneof
	Inputs            map[string]*unordered.Message
	Queries           map[string]*parser.RPC
	Mutations         map[string]*parser.RPC
	Services          []*unordered.Service
	Imports           []string
}

func NewGenerator(typeNameGenerator TypeNameGenerator, baseFileName string) Generator {
	return &generator{
		TypeNameGenerator: typeNameGenerator,
		BaseFileName:      baseFileName,
		Enums:             make(map[string]*unordered.Enum),
		Inputs:            make(map[string]*unordered.Message),
		Objects:           make(map[string]*unordered.Message),
		Unions:            make(map[string]*parser.Oneof),
		Queries:           make(map[string]*parser.RPC),
		Mutations:         make(map[string]*parser.RPC),
		Imports:           make([]string, 0),
	}
}

func (g *generator) FromProto(p *parser.Proto) (bool, error) {
	proto, err := protoparser.UnorderedInterpret(p)
	if err != nil {
		return false, err
	}
	for _, imp := range proto.ProtoBody.Imports {
		importLine, ok := wellKnownTypes_imports[strings.Trim(imp.Location, `"`)]
		if ok {
			g.Imports = append(g.Imports, importLine)
		}
	}
	g.PackageName = proto.ProtoBody.Packages[0].Name
	g.Services = make([]*unordered.Service, 0)
	for _, svc := range proto.ProtoBody.Services {
		svcHasGraphQL := false
		for _, rpc := range svc.ServiceBody.RPCs {
			for _, opt := range rpc.Options {
				if opt.OptionName != OptionGraphQL {
					continue
				}
				if !svcHasGraphQL {
					svcHasGraphQL = true
				}
				val := opt.Constant[1 : len(opt.Constant)-1]
				segments := strings.Split(val, ":")
				op := segments[0]
				opName := strings.ReplaceAll(segments[1], "\\\"", "\"")[1 : len(segments[1])-1]
				if op == "query" {
					if _, existing := g.Queries[opName]; !existing {
						g.Queries[opName] = rpc
					} else {
						return true, fmt.Errorf("Duplicate query `%s`", opName)
					}
				} else if op == "mutation" {
					if _, existing := g.Mutations[opName]; !existing {
						g.Mutations[opName] = rpc
					} else {
						return true, fmt.Errorf("Duplicate mutation `%s`", opName)
					}
				}
				if _, ok := g.Inputs[rpc.RPCRequest.MessageType]; !ok {
					g.Inputs[rpc.RPCRequest.MessageType] = FindMessage(rpc.RPCRequest.MessageType, proto.ProtoBody)
				}
				if _, ok := g.Objects[rpc.RPCResponse.MessageType]; !ok {
					g.Objects[rpc.RPCResponse.MessageType] = FindMessage(rpc.RPCResponse.MessageType, proto.ProtoBody)
				}
			}
		}
		if svcHasGraphQL {
			g.Services = append(g.Services, svc)
		}
	}
	if len(g.Queries) == 0 && len(g.Mutations) == 0 {
		return false, nil
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

	return true, nil
}

func (g *generator) GetInputType(typeName string) string {
	wellKnownType, ok := wellKnownTypes_input_map[typeName]
	if ok {
		return wellKnownType
	}
	_, ok = g.Inputs[typeName]
	if ok {
		return "GraphQL_" + typeName + "Input"
	}
	return typeName
}

func (g *generator) GetOutputType(typeName string) string {
	wellKnownType, ok := wellKnownTypes_map[typeName]
	if ok {
		return wellKnownType
	}
	_, ok = g.Objects[typeName]
	if ok {
		return "GraphQL_" + typeName
	}
	return typeName
}

func (g *generator) GetFieldType(msg *unordered.Message, field *parser.Field, suffix string) *TypeInfo {
	switch field.Type {
	case "float":
		fallthrough
	case "double":
		return &TypeInfo{Name: "Float", Prefix: "graphql", IsScalar: true, IsRepeated: field.IsRepeated, IsNonNull: true}
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
		return &TypeInfo{Name: "Int", Prefix: "graphql", IsScalar: true, IsRepeated: field.IsRepeated, IsNonNull: true}
	case "bool":
		return &TypeInfo{Name: "Boolean", Prefix: "graphql", IsScalar: true, IsRepeated: field.IsRepeated, IsNonNull: true}
	case "string":
		return &TypeInfo{Name: "String", Prefix: "graphql", IsScalar: true, IsRepeated: field.IsRepeated, IsNonNull: true}
	case "google.protobuf.Empty":
		return &TypeInfo{Name: "Empty", Prefix: "pbEmpty", IsRepeated: field.IsRepeated}
	case "google.protobuf.Timestamp":
		return &TypeInfo{Name: "GraphQL_Timestamp", Prefix: "pbGraphql", IsScalar: true, IsRepeated: field.IsRepeated}
	case "google.protobuf.Duration":
		return &TypeInfo{Name: "GraphQL_Duration", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.DoubleValue":
		return &TypeInfo{Name: "GraphQL_DoubleValue", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.FloatValue":
		return &TypeInfo{Name: "GraphQL_FloatValue", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.Int64Value":
		return &TypeInfo{Name: "GraphQL_Int64Value", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.UInt64Value":
		return &TypeInfo{Name: "GraphQL_UInt64Value", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.Int32Value":
		return &TypeInfo{Name: "GraphQL_Int32Value", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.UInt32Value":
		return &TypeInfo{Name: "GraphQL_UInt32Value", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.BoolValue":
		return &TypeInfo{Name: "GraphQL_BoolValue", Prefix: "pbGraphql", IsRepeated: field.IsRepeated}
	case "google.protobuf.Any":
		fallthrough
	case "google.protobuf.BytesValue":
		fallthrough
	case "google.protobuf.StringValue":
		return &TypeInfo{Name: "GraphQL_StringValue", Prefix: "pbGraphql", IsRepeated: field.IsRepeated, IsNonNull: false}
	default:
		info := &TypeInfo{Name: field.Type, IsScalar: false, IsRepeated: field.IsRepeated, Suffix: suffix}
		_, isMessageEnum := g.Enums[msg.MessageName+"_"+field.Type]
		if isMessageEnum {
			info.Suffix = "Enum"
			info.IsEnum = true
			return info
		}
		_, isEnum := g.Enums[field.Type]
		if isEnum {
			info.Suffix = "Enum"
			info.IsEnum = true
			return info
		}
		return info
	}
}

func (g *generator) GetProtobufFieldName(s string) string {
	segments := strings.Split(s, "_")
	res := ""
	for _, segment := range segments {
		res = res + strings.Title(segment)
	}
	return res
}

func (g *generator) GetBaseType(t string) string {
	knownType, ok := wellKnownTypes_base[t]
	if ok {
		return knownType
	}
	return t
}

const (
	codeTemplate string = `// DO NOT EDIT! This file is autogenerated by 'github.com/ncrypthic/graphql-grpc-edge/protoc-gen-graphql/generator'
package {{.PackageName}}

import (
	"encoding/json"

	"github.com/graphql-go/graphql"
	opentracing "github.com/opentracing/opentracing-go"
	{{- range $import := .Imports }}
	{{ $import }}
	{{- end }}
)

{{ range $input := .Inputs }}
{{- if $input -}}
var GraphQL_{{$input.MessageName}}Input *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "{{$input.MessageName}}Input",
	Fields: graphql.InputObjectConfigFieldMap{
		{{- range $field := $input.MessageBody.Fields -}}
		{{ $type := GetFieldType $input $field "Input"}}
		"{{$field.FieldName}}": &graphql.InputObjectFieldConfig{
			Type: {{ $type.GetName }},
		},
		{{- end }}
	},
})
{{- end }}

{{ end }}
{{- range $output := .Objects }}
{{- if $output -}}
var GraphQL_{{$output.MessageName}} *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "{{$output.MessageName}}",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		{{- range $field := $output.MessageBody.Fields -}}
		{{ $type := GetFieldType $output $field "" }}
		"{{$field.FieldName}}": &graphql.Field{
			Name: "{{ $field.FieldName }}",
			Type: {{ $type.GetName }},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if pdata, ok := p.Source.(*{{$output.MessageName}}); ok {
					return pdata.{{GetProtobufFieldName $field.FieldName}}, nil
				} else if data, ok := p.Source.({{$output.MessageName}}); ok {
					return data.{{GetProtobufFieldName $field.FieldName}}, nil
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
{{- end }}

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
func Register{{$svc.ServiceName}}Queries(sc {{$svc.ServiceName}}Client) error {
	{{- range $name,$query := $queries }}
	pbGraphql.RegisterQuery("{{$name}}", &graphql.Field{
		Name: "{{$name}}",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: {{GetInputType $query.RPCRequest.MessageType}},
			},
		},
		Type: {{GetOutputType $query.RPCResponse.MessageType}},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			span, ctx := opentracing.StartSpanFromContext(p.Context, "{{$name}}")
			defer span.Finish()
			var req {{ GetBaseType $query.RPCRequest.MessageType }}
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(rawJson, &req)
			if err != nil {
				return nil, err
			}
			return sc.{{$query.RPCName}}(ctx, &req)
		},
	})
	{{- end }}

	return nil
}

func Register{{$svc.ServiceName}}Mutations(sc {{$svc.ServiceName}}Client) error {
	{{- range $name,$mutation := $mutations }}
	pbGraphql.RegisterMutation("{{$name}}", &graphql.Field{
		Name: "{{$name}}",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: {{GetInputType $mutation.RPCRequest.MessageType}},
			},
		},
		Type: {{GetOutputType $mutation.RPCResponse.MessageType}},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			span, ctx := opentracing.StartSpanFromContext(p.Context, "{{$name}}")
			defer span.Finish()
			var req {{ GetBaseType $mutation.RPCRequest.MessageType }}
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(rawJson, &req)
			if err != nil {
				return nil, err
			}
			return sc.{{$mutation.RPCName}}(ctx, &req)
		},
	})
	{{- end }}
	return nil
}
{{ end }}

func Register{{NormalizedFileName .BaseFileName}}GraphQLTypes() {
	{{- range $input := .Inputs }}
	{{- if $input }}
	pbGraphql.RegisterType(GraphQL_{{$input.MessageName}}Input)
	{{- end -}}
	{{- end }}
	{{- range $output := .Objects }}
	{{- if $output }}
	pbGraphql.RegisterType(GraphQL_{{$output.MessageName}})
	{{- end -}}
	{{- end }}
	{{- range $name,$enum := .Enums }}
	pbGraphql.RegisterType(GraphQL_{{$enum.EnumName}}Enum)
	{{- end }}
	{{- range $name,$union := .Unions }}
	pbGraphql.RegisterType(GraphQL_{{$name}}Union)
	{{- end }}
}`
)

func FindMessage(name string, b *unordered.ProtoBody) *unordered.Message {
	for _, m := range b.Messages {
		if m.MessageName == name {
			return m
		}
	}
	return nil
}

func NormalizedFileName(s string) string {
	replaceStr := []string{".", "/", "_"}
	for _, c := range replaceStr {
		s = strings.ReplaceAll(s, c, "")
	}
	return strings.Title(s)
}
