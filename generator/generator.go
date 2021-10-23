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
	Fields    []*descriptorpb.FieldDescriptorProto
	Unions    map[string][]*descriptorpb.FieldDescriptorProto
	Maps      []*descriptorpb.FieldDescriptorProto
	HasFields bool
}

//TypeNameGenerator is function type to generate GraphQL type
//from protobuf object type
type TypeNameGenerator func(packageName, typeName string) string

//DefaultNameGenerator is a default type name generator
func DefaultNameGenerator(packageName, typeName string) string {
	return strings.Title(packageName + typeName)
}

type TypeInfo struct {
	Name         string
	Prefix       string
	PackageName  string
	Suffix       string
	IsScalar     bool
	IsRepeated   bool
	IsNonNull    bool
	LanguageType string
}

func (t *TypeInfo) GetName() string {
	return t.formatRepeated("GraphQL_")
}

func (t *TypeInfo) String() string {
	return t.formatRepeated("GraphQL_")
}

func (t *TypeInfo) formatRepeated(annotation string) string {
	if t.IsRepeated {
		return fmt.Sprintf(`graphql.NewList(%s)`, t.formatNonNull(annotation))
	} else {
		return t.formatNonNull(annotation)
	}
}

func (t *TypeInfo) formatNonNull(annotation string) string {
	if t.IsNonNull {
		return fmt.Sprintf(`graphql.NewNonNull(%s)`, t.formatName(annotation))
	} else {
		return t.formatName(annotation)
	}
}

func (t *TypeInfo) formatName(annotation string) string {
	if t.IsScalar {
		return t.Prefix + "." + t.PackageName + t.Name
	}
	if t.Prefix != "" {
		return t.Prefix + ".GraphQL_" + t.PackageName + t.Name + t.Suffix
	}
	return annotation + t.PackageName + t.Name + t.Suffix
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
	File              *descriptorpb.FileDescriptorProto
	Maps              map[string]*descriptorpb.DescriptorProto
	Enums             map[string]*descriptorpb.EnumDescriptorProto
	Objects           map[string]*descriptorpb.DescriptorProto
	Unions            map[string]*descriptorpb.OneofDescriptorProto
	UnionFields       map[string][]*UnionField
	Inputs            map[string]*descriptorpb.DescriptorProto
	Queries           map[string]map[string]*descriptorpb.MethodDescriptorProto
	Mutations         map[string]map[string]*descriptorpb.MethodDescriptorProto
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
		Maps:              make(map[string]*descriptorpb.DescriptorProto),
		Inputs:            make(map[string]*descriptorpb.DescriptorProto),
		Objects:           make(map[string]*descriptorpb.DescriptorProto),
		Unions:            make(map[string]*descriptorpb.OneofDescriptorProto),
		UnionFields:       make(map[string][]*UnionField),
		Queries:           make(map[string]map[string]*descriptorpb.MethodDescriptorProto),
		Mutations:         make(map[string]map[string]*descriptorpb.MethodDescriptorProto),
		Imports:           make([]string, 0),
		LanguageType:      make(map[string]string),
		Options:           options,
		Externals:         make([]*Extern, 0),
	}
}

func (g *generator) FromProto(p *descriptorpb.FileDescriptorProto, importPrefix, packageName string) (bool, error) {
	g.File = p
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
		svcQueries := make(map[string]*descriptorpb.MethodDescriptorProto)
		svcMutations := make(map[string]*descriptorpb.MethodDescriptorProto)
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
			fqOutputType := g.FullyQualifiedName(rpc.GetOutputType())
			if inputType, ok := g.FindMessage(fqInputType); ok {
				g.visitDescriptor(inputType, "", g.Inputs)
			}
			if outputType, ok := g.FindMessage(fqOutputType); ok {
				g.visitDescriptor(outputType, "", g.Objects)
			}
			if opName := opt.GetQuery(); opName != "" {
				if _, existing := svcQueries[opName]; !existing {
					svcQueries[opName] = rpc
				} else {
					return true, fmt.Errorf("Duplicate query `%s`", opName)
				}
			}
			if opName := opt.GetMutation(); opName != "" {
				if _, existing := svcMutations[opName]; !existing {
					svcMutations[opName] = rpc
				} else {
					return true, fmt.Errorf("Duplicate mutation `%s`", opName)
				}
			}
		}
		if svcHasGraphQL {
			g.Services = append(g.Services, svc)
			g.Queries[svc.GetName()] = svcQueries
			g.Mutations[svc.GetName()] = svcMutations
		}
	}

	return true, nil
}

func (g *generator) visitDescriptor(d *descriptorpb.DescriptorProto, parentTypeName string, targetMap map[string]*descriptorpb.DescriptorProto) {
	fqmn := g.FullyQualifiedName(d.GetName())
	if parentTypeName != "" {
		fqmn = g.FullyQualifiedName(parentTypeName + "." + d.GetName())
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
		g.Enums[fqmn+"."+enum.GetName()] = enum
	}
	for _, nestedMsg := range d.GetNestedType() {
		if nestedMsg.GetOptions().GetMapEntry() {
			g.Maps[fqmn+"."+nestedMsg.GetName()] = nestedMsg
			continue
		}
		g.visitDescriptor(nestedMsg, d.GetName(), targetMap)
	}
	for _, f := range d.GetField() {
		if f.GetTypeName() == "" {
			continue
		}
		if _, ok := targetMap[f.GetTypeName()]; ok {
			continue
		}
		msg, ok := g.FindMessage(f.GetTypeName())
		if ok {
			targetMap[f.GetTypeName()] = msg
		}
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

func (g *generator) GetScalarFieldType(f *descriptorpb.FieldDescriptorProto) (t *TypeInfo, ok bool) {
	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		t = &TypeInfo{
			Name:         "Float",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "float64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int32",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "uint64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "uint32",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int32",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int32",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		t = &TypeInfo{
			Name:         "Int",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "int64",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		t = &TypeInfo{
			Name:         "Boolean",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "bool",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		t = &TypeInfo{
			Name:         "Boolean",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "[]byte",
		}
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		t = &TypeInfo{
			Name:         "String",
			Prefix:       "graphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			IsNonNull:    true,
			LanguageType: "string",
		}
	}
	return t, t != nil
}

func (g *generator) GetWellknownFieldType(msg *descriptorpb.DescriptorProto, f *descriptorpb.FieldDescriptorProto, suffix string) (t *TypeInfo, ok bool) {
	switch f.GetTypeName() {
	case ".google.protobuf.Empty":
		t = &TypeInfo{
			Name:         "Empty",
			Prefix:       "pbEmpty",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbEmpty.Empty",
		}
	case ".google.protobuf.Timestamp":
		t = &TypeInfo{
			Name:         "GraphQL_Timestamp",
			Prefix:       "pbGraphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbTimestamp.Timestamp",
		}
	case ".google.protobuf.Duration":
		t = &TypeInfo{
			Name:         "GraphQL_Duration",
			Prefix:       "pbGraphql",
			IsScalar:     true,
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbTimestamp.Duration",
		}
	case ".google.protobuf.DoubleValue":
		t = &TypeInfo{
			Name:         "GraphQL_DoubleValue",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.DoubleValue",
		}
	case ".google.protobuf.FloatValue":
		t = &TypeInfo{
			Name:         "GraphQL_FloatValue",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.FloatValue",
		}
	case ".google.protobuf.Int64Value":
		t = &TypeInfo{
			Name:         "GraphQL_Int64Value",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.Int64Value",
		}
	case ".google.protobuf.UInt64Value":
		t = &TypeInfo{
			Name:         "GraphQL_UInt64Value",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.UInt64Value",
		}
	case ".google.protobuf.Int32Value":
		t = &TypeInfo{
			Name:         "GraphQL_Int32Value",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.Int32Value",
		}
	case ".google.protobuf.UInt32Value":
		t = &TypeInfo{
			Name:         "GraphQL_UInt32Value",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.UInt32Value",
		}
	case ".google.protobuf.BoolValue":
		t = &TypeInfo{
			Name:         "GraphQL_BoolValue",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.BoolValue",
		}
	case ".google.protobuf.Any":
		t = &TypeInfo{
			Name:         "GraphQL_StringValue",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.Any",
		}
	case ".google.protobuf.BytesValue":
		t = &TypeInfo{
			Name:         "GraphQL_StringValue",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.BytesValue",
		}
	case ".google.protobuf.StringValue":
		t = &TypeInfo{
			Name:         "GraphQL_StringValue",
			Prefix:       "pbGraphql",
			IsRepeated:   g.isRepeatedField(f),
			LanguageType: "pbWrappers.StringValue",
		}
	}
	return t, t != nil
}

func (g *generator) isMapField(msg *descriptorpb.DescriptorProto, f *descriptorpb.FieldDescriptorProto) bool {
	_, ok := g.Maps[f.GetTypeName()]
	return ok
}

func (g *generator) GetFieldType(msg *descriptorpb.DescriptorProto, f *descriptorpb.FieldDescriptorProto, suffix string) *TypeInfo {
	t, ok := g.GetScalarFieldType(f)
	if ok {
		return t
	}
	t, ok = g.GetWellknownFieldType(msg, f, suffix)
	if ok {
		return t
	}
	info := &TypeInfo{
		Name:         g.GetGraphQLTypeName(f.GetTypeName()),
		LanguageType: g.GetLanguageType(f.GetTypeName()),
		IsScalar:     false,
		IsRepeated:   g.isRepeatedField(f),
		Suffix:       suffix,
	}
	_, isEnum := g.Enums[f.GetTypeName()]
	if isEnum {
		info.Suffix = "Enum"
		return info
	}
	_, isMap := g.Maps[f.GetTypeName()]
	if isMap {
		info.Suffix = "Map"
		return info
	}
	if ext, ok := g.GetExternal(f.GetTypeName()); ok {
		info.Prefix = ext.ImportAlias
	}
	return info
}

func (g *generator) GetObjectFields(msg *descriptorpb.DescriptorProto) *ObjectField {
	fields := make([]*descriptorpb.FieldDescriptorProto, 0)
	unions := make(map[string][]*descriptorpb.FieldDescriptorProto, 0)
	maps := make([]*descriptorpb.FieldDescriptorProto, 0)
	for _, f := range msg.GetField() {
		if _, ok := g.Maps[f.GetTypeName()]; ok {
			maps = append(maps, f)
			continue
		}
		if f.OneofIndex != nil {
			unionFieldName := msg.OneofDecl[f.GetOneofIndex()].GetName()
			_, exists := unions[unionFieldName]
			if !exists {
				unions[unionFieldName] = make([]*descriptorpb.FieldDescriptorProto, 0)
			}
			unions[unionFieldName] = append(unions[unionFieldName], f)
			continue
		}
		fields = append(fields, f)
	}
	return &ObjectField{
		Fields:    fields,
		Unions:    unions,
		Maps:      maps,
		HasFields: len(fields) > 0 || len(unions) > 0 || len(maps) > 0,
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
	if len(fqn) == 0 {
		return ""
	}
	t, ok := wellKnownTypes_base[fqn[1:]]
	if ok {
		return t
	}
	ext, ok := g.GetExternal(fqn)
	if ok {
		return ext.ImportAlias + strings.Replace(fqn, ext.ProtoPackage, "", 1)
	}
	return strings.ReplaceAll(strings.Replace(fqn, "."+g.PackageName, "", 1)[1:], ".", "_")
}

func (g *generator) GetGraphQLTypeName(fqn string) string {
	if len(fqn) == 0 {
		return ""
	}
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

func (g *generator) FindMessage(name string) (*descriptorpb.DescriptorProto, bool) {
	for _, m := range g.File.GetMessageType() {
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

const (
	codeTemplate string = `// DO NOT EDIT! This file is autogenerated by 'github.com/ncrypthic/graphql-grpc-edge/protoc-gen-graphql/generator'
package {{.GoPackageName}}

import (
{{- if HasOperation -}}
	"encoding/json"

	opentracing "github.com/opentracing/opentracing-go"
{{- end -}}
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
		{{- range $field := $obj.Fields }}
		"{{ $field.GetName }}": &graphql.Field{
			Name: "{{ $field.GetName }}",
			Type: {{ (GetFieldType $output $field "") }},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var res interface{}
				if pdata, ok := p.Source.(*{{ GetLanguageType $fqmn }}); ok {
					res = pdata.{{ GetProtobufFieldName $field.GetName }}
				} else if data, ok := p.Source.({{ GetLanguageType $fqmn }}); ok {
					res = data.{{ GetProtobufFieldName $field.GetName }}
				}
				if res == nil {
					return nil, nil
				}
				switch t := res.(type) {
				case {{ (GetFieldType $output $field "").LanguageType }}:
					return t, nil
				default:
					return nil, pbGraphql.ErrBadValue
				}
			},
		},
		{{- end }}
		{{- range $union, $unionFields := $obj.Unions }}
		"{{ $union }}": &graphql.Field{
			Name: "{{ $union }}",
			Type: GraphQL_{{ GetGraphQLTypeName $fqmn }}_{{ $union }}Union,
		},
		{{- end }}
		{{- range $mapField := $obj.Maps }}
		"{{ $mapField.GetName }}": &graphql.Field{
			Name: "{{ $mapField.GetName }}",
			Type: graphql.NewList(GraphQL_{{ GetGraphQLTypeName $mapField.GetTypeName }}Map),
		},
		{{- end }}
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
		{{- end }}
		{{- range $mapField:= $obj.Maps }}
		"{{ $mapField.GetName }}": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(GraphQL_{{ GetGraphQLTypeName $mapField.GetTypeName }}MapInput),
		},
		{{- end }}
	},
})
{{ end }}
{{- range $name,$enum := .Enums }}

var GraphQL_{{ GetGraphQLTypeName $name }}Enum *graphql.Enum = graphql.NewEnum(graphql.EnumConfig{
	Name: "{{ GetGraphQLTypeName $name }}Enum",
	Values: graphql.EnumValueConfigMap{
		{{- range $enumField := $enum.Value }}
		"{{ $enumField.GetName }}": &graphql.EnumValueConfig{
			Value: {{ GetLanguageType $name }}_value["{{ $enumField.GetName }}"],
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
{{- range $name,$map := .Maps }}

var GraphQL_{{ GetGraphQLTypeName $name }}Map *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "{{ GetGraphQLTypeName $name }}Map",
	Fields: graphql.Fields{
		{{ $key := index $map.Field 0 -}}
		{{- $value := index $map.Field 1 -}}
		"{{ $key.GetName }}": &graphql.Field{
			Type: {{ GetFieldType $map $key "" }},
		},
		"{{ $value.GetName }}": &graphql.Field{
			Type: {{ GetFieldType $map $value "" }},
		},
	},
})

var GraphQL_{{ GetGraphQLTypeName $name }}MapInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "{{ GetGraphQLTypeName $name }}MapInput",
	Fields: graphql.InputObjectConfigFieldMap{
		{{ $key := index $map.Field 0 -}}
		{{- $value := index $map.Field 1 -}}
		"{{ $key.GetName }}": &graphql.InputObjectFieldConfig{
			Type: {{ GetFieldType $map $key "" }},
		},
		"{{ $value.GetName }}": &graphql.InputObjectFieldConfig{
			Type: {{ GetFieldType $map $value "" }},
		},
	},
})
{{ end }}
{{- $queries := .Queries -}}
{{- $mutations := .Mutations -}}
{{ range $svc := .Services }}
{{- $svcName := $svc.GetName -}}

func Register{{ $svc.GetName }}Queries(sc {{ $svc.GetName }}Client) error {
	{{- range $name,$query :=  (index $queries $svcName) }}
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
			var res *{{ GetLanguageType $query.GetOutputType }}
			res, err = sc.{{ $query.GetName }}(ctx, &req)
			return res, err
		},
	})
	{{- end }}

	return nil
}

func Register{{ $svc.GetName }}Mutations(sc {{ $svc.GetName }}Client) error {
	{{- range $name,$mutation := (index $mutations $svcName) }}
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
			var res *{{ GetLanguageType $mutation.GetOutputType }}
			res, err = sc.{{ $mutation.GetName }}(ctx, &req)
			return res, err
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
}
`
)
