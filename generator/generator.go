package generator

import (
	"fmt"
	"path"
	"strings"

	"github.com/ncrypthic/graphql-grpc-edge/graphql"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	OptionGraphQL string = "(graphql.type)"

	ImportPrefixOption = "import_prefix"
	ImportPathOption   = "import_path"
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

type Option interface {
	Name() string
	Value() string
}

type option struct {
	name  string
	value string
}

func (opt *option) Name() string {
	return opt.name
}

func (opt *option) Value() string {
	return opt.value
}

func NewOption(key, value string) Option {
	return &option{name: key, value: value}
}

type Options []Option

func (opts Options) Get(key string) (string, bool) {
	for _, opt := range opts {
		if opt.Name() == key {
			return opt.Value(), true
		}
	}
	return "", false
}

type Extern struct {
	ProtoPackage string
	ImportAlias  string
	SourcePath   string
}

type ObjectField struct {
	Fields []*descriptorpb.FieldDescriptorProto
	Unions map[string][]*descriptorpb.FieldDescriptorProto
}

//TypeNameGenerator is function type to generate GraphQL type
//from protobuf object type
type TypeNameGenerator func(packageName, typeName string) string

//DefaultNameGenerator is a default type name generator
func DefaultNameGenerator(packageName, typeName string) string {
	return strings.Title(packageName + typeName)
}

type TypeInfo struct {
	Name        string
	Prefix      string
	PackageName string
	Suffix      string
	IsScalar    bool
	IsRepeated  bool
	IsEnum      bool
	IsNonNull   bool
}

func (t *TypeInfo) GetName() string {
	return t.formatRepeated()
}

func (t *TypeInfo) String() string {
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
		return t.Prefix + "." + t.PackageName + t.Name
	}
	if t.Prefix != "" {
		return t.Prefix + ".GraphQL_" + t.PackageName + t.Name + t.Suffix
	}
	return "GraphQL_" + t.PackageName + t.Name + t.Suffix
}

//Generator is an interface of graphql code generator
type Generator interface {
	FromProto(*descriptorpb.FileDescriptorProto, string, string) (bool, error)
	GetFieldType(*descriptorpb.DescriptorProto, *descriptorpb.FieldDescriptorProto, string) *TypeInfo
	GetObjectFields(*descriptorpb.DescriptorProto) *ObjectField
	GetProtobufFieldName(string) string
	GetInputType(string) string
	GetOutputType(string) string
	GetEnumType(string) string
	GetLanguageType(string) string
	GetGraphQLTypeName(string) string
	GetExternal(string) (*Extern, bool)
	HasOperation() bool
}

type UnionField struct {
	Message *descriptorpb.DescriptorProto
	Field   *descriptorpb.FieldDescriptorProto
}

type generator struct {
	TypeNameGenerator TypeNameGenerator
	GoPackageName     string
	PackageName       string
	BaseFileName      string
	Enums             map[string]*descriptorpb.EnumDescriptorProto
	Objects           map[string]*descriptorpb.DescriptorProto
	Unions            map[string]*descriptorpb.OneofDescriptorProto
	UnionFields       map[string][]*UnionField
	Inputs            map[string]*descriptorpb.DescriptorProto
	Queries           map[string]*descriptorpb.MethodDescriptorProto
	Mutations         map[string]*descriptorpb.MethodDescriptorProto
	Services          []*descriptorpb.ServiceDescriptorProto
	LanguageType      map[string]string
	Imports           []string
	Options           Options
	Externals         []*Extern
}

func NewGenerator(typeNameGenerator TypeNameGenerator, baseFileName string, options Options) Generator {
	return &generator{
		TypeNameGenerator: typeNameGenerator,
		BaseFileName:      baseFileName,
		GoPackageName:     "",
		PackageName:       "",
		Enums:             make(map[string]*descriptorpb.EnumDescriptorProto),
		Inputs:            make(map[string]*descriptorpb.DescriptorProto),
		Objects:           make(map[string]*descriptorpb.DescriptorProto),
		Unions:            make(map[string]*descriptorpb.OneofDescriptorProto),
		UnionFields:       make(map[string][]*UnionField),
		Queries:           make(map[string]*descriptorpb.MethodDescriptorProto),
		Mutations:         make(map[string]*descriptorpb.MethodDescriptorProto),
		Imports:           make([]string, 0),
		LanguageType:      make(map[string]string),
		Options:           options,
		Externals:         make([]*Extern, 0),
	}
}

func (g *generator) FromProto(p *descriptorpb.FileDescriptorProto, importPrefix, packageName string) (bool, error) {
	for _, imp := range p.Dependency {
		importLine, ok := wellKnownTypes_imports[imp]
		if ok {
			g.Imports = append(g.Imports, importLine)
			continue
		}
		// External imports
		g.Externals = append(g.Externals, g.GenerateExternal(importPrefix, imp))
	}
	for _, ext := range g.Externals {
		g.Imports = append(g.Imports, fmt.Sprintf(`%s "%s"`, ext.ImportAlias, ext.SourcePath))
	}
	g.GoPackageName = packageName
	g.PackageName = p.GetPackage()
	for _, msg := range p.GetMessageType() {
		g.visitDescriptor(msg, "", g.Objects)
	}
	for _, enum := range p.GetEnumType() {
		fqn := g.FullyQualifiedName(enum.GetName())
		g.Enums[fqn] = enum
	}
	g.Services = make([]*descriptorpb.ServiceDescriptorProto, 0)
	for _, svc := range p.GetService() {
		svcHasGraphQL := false
		for _, rpc := range svc.GetMethod() {
			graphqlOpt := proto.GetExtension(rpc.GetOptions(), graphql.E_Type)
			if graphqlOpt == nil {
				continue
			}
			opt, ok := graphqlOpt.(*graphql.GraphQLOption)
			if !ok {
				panic("unexpected graphql option: " + rpc.GetOptions().String())
			}
			if opt == nil {
				continue
			}
			svcHasGraphQL = true
			fqInputType := g.FullyQualifiedName(rpc.GetInputType())
			if inputType, ok := g.FindMessage(fqInputType, p); ok {
				g.visitDescriptor(inputType, "", g.Inputs)
				for _, f := range inputType.GetField() {
					if msg, ok := g.Objects[f.GetTypeName()]; ok {
						g.Inputs[f.GetTypeName()] = msg
					}
				}
			}
			if opName := opt.GetQuery(); opName != "" {
				if _, existing := g.Queries[opName]; !existing {
					g.Queries[opName] = rpc
				} else {
					return true, fmt.Errorf("Duplicate query `%s`", opName)
				}
			}
			if opName := opt.GetMutation(); opName != "" {
				if _, existing := g.Mutations[opName]; !existing {
					g.Mutations[opName] = rpc
				} else {
					return true, fmt.Errorf("Duplicate mutation `%s`", opName)
				}
			}
		}
		if svcHasGraphQL {
			g.Services = append(g.Services, svc)
		}
	}

	return true, nil
}

func (g *generator) visitDescriptor(d *descriptorpb.DescriptorProto, parentType string, targetMap map[string]*descriptorpb.DescriptorProto) {
	fqmn := g.FullyQualifiedName(d.GetName())
	if parentType != "" {
		fqmn = g.FullyQualifiedName(parentType + "_" + d.GetName())
	}
	if _, ok := targetMap[fqmn]; !ok {
		targetMap[fqmn] = d
	}
	for idx, union := range d.GetOneofDecl() {
		g.Unions[fqmn+"_"+union.GetName()] = union
		fields := make([]*UnionField, 0)
		for _, field := range d.GetField() {
			if field.OneofIndex != nil && field.GetOneofIndex() == int32(idx) {
				fields = append(fields, &UnionField{
					Message: d,
					Field:   field,
				})
			}
		}
		g.UnionFields[fqmn+"_"+union.GetName()] = fields
	}
	for _, enum := range d.GetEnumType() {
		g.Enums[fqmn+"_"+enum.GetName()] = enum
	}
	for _, nestedMsg := range d.GetNestedType() {
		g.visitDescriptor(nestedMsg, d.GetName(), targetMap)
	}
}

func (g *generator) GetInputType(typeName string) string {
	wellKnownType, ok := wellKnownTypes_input_map[typeName[1:]]
	if ok {
		return wellKnownType
	}
	ext, ok := g.GetExternal(typeName)
	if ok {
		return ext.ImportAlias + strings.Replace(typeName, ext.ProtoPackage, "", 1) + "Input"
	}
	return "GraphQL_" + g.GetGraphQLTypeName(typeName) + "Input"
}

func (g *generator) GetOutputType(typeName string) string {
	wellKnownType, ok := wellKnownTypes_map[typeName[1:]]
	if ok {
		return wellKnownType
	}
	ext, ok := g.GetExternal(typeName)
	if ok {
		return ext.ImportAlias + strings.Replace(typeName, ext.ProtoPackage, "", 1)
	}
	return "GraphQL_" + g.GetGraphQLTypeName(typeName)
}

func (g *generator) GetEnumType(typeName string) string {
	wellKnownType, ok := wellKnownTypes_map[typeName[1:]]
	if ok {
		return wellKnownType
	}
	ext, ok := g.GetExternal(typeName)
	if ok {
		return ext.ImportAlias + strings.Replace(typeName, ext.ProtoPackage, "", 1)
	}
	_, ok = g.Enums[typeName]
	if ok {
		segments := strings.Split(typeName, ".")
		return segments[len(segments)-1]
	}
	return typeName
}

func (g *generator) GetEnumName(typeName string) string {
	wellKnownType, ok := wellKnownTypes_map[typeName[1:]]
	if ok {
		return wellKnownType
	}
	ext, ok := g.GetExternal(typeName)
	if ok {
		return ext.ImportAlias + strings.Replace(typeName, ext.ProtoPackage, "", 1)
	}
	_, ok = g.Enums[typeName]
	if ok {
		segments := strings.Split(typeName, ".")
		return "GraphQL_" + segments[len(segments)-1]
	}
	return typeName
}

func (g *generator) GetExternal(typeName string) (*Extern, bool) {
	for _, ext := range g.Externals {
		if strings.HasPrefix(typeName, ext.ProtoPackage) {
			return ext, true
		}
	}
	return nil, false
}

func (g *generator) GenerateExternal(importPrefix, location string) *Extern {
	importPath, hasImportPath := g.Options.Get(ImportPathOption)
	location = strings.Trim(location, `"`)
	segments := strings.Split(strings.TrimSuffix(location, ".proto"), "/")
	protoPackage := "."
	importAlias := "pb"
	sourcePath := ""
	for _, segment := range segments {
		protoPackage = protoPackage + segment + "."
		importAlias = importAlias + strings.Title(segment)
		sourcePath = sourcePath + segment + "/"
	}
	if hasImportPath && strings.HasPrefix(sourcePath, importPath) {
		sourcePath = importPrefix + sourcePath
	} else {
		sourcePath = importPrefix + sourcePath
	}
	return &Extern{
		ProtoPackage: protoPackage[0 : len(protoPackage)-1],
		ImportAlias:  importAlias[0 : len(importAlias)-1],
		SourcePath:   path.Dir(sourcePath[0 : len(sourcePath)-1]),
	}
}

func (g *generator) isRepeatedField(field *descriptorpb.FieldDescriptorProto) bool {
	return field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

func (g *generator) GetFieldType(msg *descriptorpb.DescriptorProto, f *descriptorpb.FieldDescriptorProto, suffix string) *TypeInfo {
	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return &TypeInfo{Name: "Float", Prefix: "graphql", IsScalar: true, IsRepeated: g.isRepeatedField(f), IsNonNull: true}
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return &TypeInfo{Name: "Int", Prefix: "graphql", IsScalar: true, IsRepeated: g.isRepeatedField(f), IsNonNull: true}
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return &TypeInfo{Name: "Boolean", Prefix: "graphql", IsScalar: true, IsRepeated: g.isRepeatedField(f), IsNonNull: true}
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return &TypeInfo{Name: "String", Prefix: "graphql", IsScalar: true, IsRepeated: g.isRepeatedField(f), IsNonNull: true}
	}
	switch f.GetTypeName() {
	case ".google.protobuf.Empty":
		return &TypeInfo{Name: "Empty", Prefix: "pbEmpty", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.Timestamp":
		return &TypeInfo{Name: "GraphQL_Timestamp", Prefix: "pbGraphql", IsScalar: true, IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.Duration":
		return &TypeInfo{Name: "GraphQL_Duration", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.DoubleValue":
		return &TypeInfo{Name: "GraphQL_DoubleValue", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.FloatValue":
		return &TypeInfo{Name: "GraphQL_FloatValue", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.Int64Value":
		return &TypeInfo{Name: "GraphQL_Int64Value", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.UInt64Value":
		return &TypeInfo{Name: "GraphQL_UInt64Value", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.Int32Value":
		return &TypeInfo{Name: "GraphQL_Int32Value", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.UInt32Value":
		return &TypeInfo{Name: "GraphQL_UInt32Value", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.BoolValue":
		return &TypeInfo{Name: "GraphQL_BoolValue", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f)}
	case ".google.protobuf.Any":
		fallthrough
	case ".google.protobuf.BytesValue":
		fallthrough
	case ".google.protobuf.StringValue":
		return &TypeInfo{Name: "GraphQL_StringValue", Prefix: "pbGraphql", IsRepeated: g.isRepeatedField(f), IsNonNull: false}
	default:
		nameSegments := strings.Split(f.GetTypeName(), ".")
		name := nameSegments[len(nameSegments)-1:][0]
		info := &TypeInfo{Name: g.GetGraphQLTypeName(f.GetTypeName()), IsScalar: false, IsRepeated: g.isRepeatedField(f), Suffix: suffix}
		fqmn := g.FullyQualifiedName(msg.GetName())
		_, isMessageEnum := g.Enums[fqmn+"_"+name]
		if isMessageEnum {
			info.Suffix = "Enum"
			info.IsEnum = true
			return info
		}
		_, isEnum := g.Enums[f.GetTypeName()]
		if isEnum {
			info.Suffix = "Enum"
			info.IsEnum = true
			return info
		}
		if ext, ok := g.GetExternal(f.GetTypeName()); ok {
			info.Prefix = ext.ImportAlias
		}
		return info
	}
}

func (g *generator) GetObjectFields(msg *descriptorpb.DescriptorProto) *ObjectField {
	fields := make([]*descriptorpb.FieldDescriptorProto, 0)
	unions := make(map[string][]*descriptorpb.FieldDescriptorProto, 0)
	for _, f := range msg.GetField() {
		if f.OneofIndex == nil {
			fields = append(fields, f)
			continue
		}
		unionFieldName := msg.OneofDecl[f.GetOneofIndex()].GetName()
		_, exists := unions[unionFieldName]
		if !exists {
			unions[unionFieldName] = make([]*descriptorpb.FieldDescriptorProto, 0)
		}
		unions[unionFieldName] = append(unions[unionFieldName], f)
	}
	return &ObjectField{
		Fields: fields,
		Unions: unions,
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

func (g *generator) HasOperation() bool {
	return len(g.Services) > 0
}

func (g *generator) GetLanguageType(fqn string) string {
	t, ok := wellKnownTypes_base[fqn[1:]]
	if ok {
		return t
	}
	ext, ok := g.GetExternal(fqn)
	if ok {
		return ext.ImportAlias + strings.Replace(fqn, ext.ProtoPackage, "", 1)
	}
	return strings.Replace(fqn, "."+g.PackageName, "", 1)[1:]
}

func (g *generator) GetGraphQLTypeName(fqn string) string {
	t, ok := wellKnownTypes_base[fqn[1:]]
	if ok {
		return t
	}
	var (
		packageName string
		typeName    string
	)
	if ext, ok := g.GetExternal(fqn); ok {
		typeName = strings.TrimPrefix(fqn, ext.ProtoPackage)
		packageName = ext.ProtoPackage
	} else {
		typeName = strings.TrimPrefix(fqn, "."+g.PackageName)
		packageName = g.PackageName
	}
	typeName = strings.ReplaceAll(typeName, ".", "_")
	segments := strings.Split(packageName, ".")
	res := ""
	for _, seg := range segments {
		res += strings.Title(seg)
	}

	return res + typeName
}

const (
	codeTemplate string = `// DO NOT EDIT! This file is autogenerated by 'github.com/ncrypthic/graphql-grpc-edge/protoc-gen-graphql/generator'
package {{.GoPackageName}}

import (
{{ if HasOperation -}}
	"encoding/json"

	opentracing "github.com/opentracing/opentracing-go"
{{ end -}}
	{{- range $import := .Imports }}
	{{ $import }}
	{{- end }}
	"github.com/graphql-go/graphql"
)

{{- range $fqmn,$output := .Objects }}
var GraphQL_{{ GetGraphQLTypeName $fqmn  }} *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "{{ GetGraphQLTypeName $fqmn }}",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	{{- $obj := GetObjectFields $output }}
	Fields: graphql.Fields{
		{{ range $field := $obj.Fields }}
		{{- $type := GetFieldType $output $field "" -}}
		"{{ $field.GetName }}": &graphql.Field{
			Name: "{{ $field.GetName }}",
			Type: {{ $type }},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if pdata, ok := p.Source.(*{{ GetLanguageType $fqmn }}); ok {
					return pdata.{{ GetProtobufFieldName $field.GetName }}, nil
				} else if data, ok := p.Source.({{ GetLanguageType $fqmn }}); ok {
					return data.{{ GetProtobufFieldName $field.GetName }}, nil
				}
				return nil, nil
			},
		},
		{{ end }}
		{{- range $union, $unionFields := $obj.Unions -}}
		"{{ $union }}": &graphql.Field{
			Name: "{{ $union }}",
			Type: GraphQL_{{ GetGraphQLTypeName $fqmn }}_{{ $union }}Union,
		},
		{{ end -}}
	},
})
{{ end }}
{{ range $inputType, $input := .Inputs }}

var GraphQL_{{ GetGraphQLTypeName $inputType }}Input *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "{{ GetGraphQLTypeName $inputType }}Input",
	{{- $obj := GetObjectFields $input }}
	Fields: graphql.InputObjectConfigFieldMap{
		{{- range $field := $obj.Fields }}
		{{- $type := GetFieldType $input $field "Input" }}
		"{{ $field.GetName }}": &graphql.InputObjectFieldConfig{
			Type: {{ $type }},
		},
		{{- end }}
		{{- range $union, $unionFields := $obj.Unions }}
		"{{ $union }}": &graphql.InputObjectFieldConfig{
			Type: GraphQL_{{ GetGraphQLTypeName $inputType }}_{{ $union }}Union,
		},
		{{ end }}
	},
})
{{ end }}
{{- range $name,$enum := .Enums }}

var GraphQL_{{ GetGraphQLTypeName $name }}Enum *graphql.Enum = graphql.NewEnum(graphql.EnumConfig{
	Name: "{{ GetGraphQLTypeName $name }}Enum",
	Values: graphql.EnumValueConfigMap{
		{{- range $enumField := $enum.Value }}
		"{{ $enumField.GetName }}": &graphql.EnumValueConfig{
			Value: {{ GetEnumType $name }}_value["{{ $enumField.GetName }}"],
		},
		{{- end }}
	},
})
{{ end }}
{{- range $name,$unionFields := .UnionFields }}
var GraphQL_{{ GetGraphQLTypeName $name }}Union *graphql.Union = graphql.NewUnion(graphql.UnionConfig{
	Name: "{{ GetGraphQLTypeName $name }}Union",
	Types: []*graphql.Object{
		{{- range $unionField := $unionFields }}
		{{ GetFieldType $unionField.Message $unionField.Field "" }},
		{{- end }}
	},
})
{{ end }}

{{- $queries := .Queries -}}
{{- $mutations := .Mutations -}}
{{ range $svc := .Services }}
func Register{{ $svc.GetName }}Queries(sc {{ $svc.GetName }}Client) error {
	{{- range $name,$query := $queries }}
	pbGraphql.RegisterQuery("{{ $name }}", &graphql.Field{
		Name: "{{ $name }}",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: {{ GetInputType $query.GetInputType }},
			},
		},
		Type: {{ GetOutputType $query.GetOutputType }},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			span, ctx := opentracing.StartSpanFromContext(p.Context, "{{$name}}")
			defer span.Finish()
			var req {{ GetLanguageType $query.GetInputType }}
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(rawJson, &req)
			if err != nil {
				return nil, err
			}
			return sc.{{ $query.GetName }}(ctx, &req)
		},
	})
	{{- end }}

	return nil
}

func Register{{ $svc.GetName }}Mutations(sc {{ $svc.GetName }}Client) error {
	{{- range $name,$mutation := $mutations }}
	pbGraphql.RegisterMutation("{{ $name }}", &graphql.Field{
		Name: "{{ $name }}",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: {{ GetInputType $mutation.GetInputType }},
			},
		},
		Type: {{ GetOutputType $mutation.GetOutputType }},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			span, ctx := opentracing.StartSpanFromContext(p.Context, "{{ $name }}")
			defer span.Finish()
			var req {{ GetLanguageType $mutation.GetInputType }}
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(rawJson, &req)
			if err != nil {
				return nil, err
			}
			return sc.{{ $mutation.GetName }}(ctx, &req)
		},
	})
	{{- end }}
	return nil
}
{{ end }}

func Register{{NormalizedFileName .BaseFileName}}GraphQLTypes() {
	{{- range $fqn, $input := .Inputs }}
	pbGraphql.RegisterType(GraphQL_{{ GetGraphQLTypeName $fqn }}Input)
	{{- end }}
	{{- range $fqn, $output := .Objects }}
	{{- if $output }}
	pbGraphql.RegisterType(GraphQL_{{ GetGraphQLTypeName $fqn }})
	{{- end -}}
	{{- end }}
	{{- range $name,$enum := .Enums }}
	pbGraphql.RegisterType(GraphQL_{{ GetGraphQLTypeName $name }}Enum)
	{{- end }}
	{{- range $name,$unionFields := .UnionFields }}
	pbGraphql.RegisterType(GraphQL_{{ GetGraphQLTypeName $name }}Union)
	{{- end }}
}`
)

func (g *generator) FindMessage(name string, b *descriptorpb.FileDescriptorProto) (*descriptorpb.DescriptorProto, bool) {
	for _, m := range b.GetMessageType() {
		if g.FullyQualifiedName(m.GetName()) == name {
			return m, true
		}
	}
	return nil, false
}

func (g *generator) FullyQualifiedName(name string) string {
	if strings.HasPrefix(name, ".") {
		return name
	}
	return "." + g.PackageName + "." + name
}

func NormalizedFileName(s string) string {
	replaceStr := []string{".", "/", "_"}
	for _, c := range replaceStr {
		s = strings.ReplaceAll(s, c, "")
	}
	return strings.Title(s)
}
