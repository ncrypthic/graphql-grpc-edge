package generator

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ncrypthic/graphql-grpc-edge/graphql"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	graphqlImport = "github.com/graphql-go/graphql"
	edgeImport    = "github.com/ncrypthic/graphql-grpc-edge/graphql"

	GQLTypeObject   GQLType = GQLType("Object")
	GQLTypeInput            = GQLType("Input")
	GQLTypeScalar           = GQLType("Scalar")
	GQLTypeEnum             = GQLType("Enum")
	GQLTypeQuery            = GQLType("Query")
	GQLTypeMutation         = GQLType("Mutation")
)

type GQLType string

type GQLIdent struct {
	protogen.GoIdent
	Type GQLType
	g    *protogen.GeneratedFile
}

func (g *GQLIdent) String() string {
	return string(g.Type) + "_" + strings.ReplaceAll(g.g.QualifiedGoIdent(g.GoIdent), ".", "_")
}

type Symbol struct {
	Ident  GQLIdent
	Parent *Symbol
}

func NewSymbol(parent *Symbol, ident GQLIdent) *Symbol {
	return &Symbol{
		Ident:  ident,
		Parent: parent,
	}
}

type SymbolTable struct {
	symbols    []*Symbol
	mapSymbols map[string]*Symbol
}

func (t *SymbolTable) Append(s *Symbol) {
	t.symbols = append(t.symbols, s)
	t.mapSymbols[s.Ident.String()] = s
}

func (t *SymbolTable) LookUp(ident GQLIdent) (*Symbol, bool) {
	s, ok := t.mapSymbols[ident.String()]
	return s, ok
}

func (t *SymbolTable) Exist(ident GQLIdent) bool {
	_, ok := t.mapSymbols[ident.String()]
	return ok
}

var (
	tbl  *SymbolTable = &SymbolTable{make([]*Symbol, 0), make(map[string]*Symbol)}
	root *Symbol      = &Symbol{}
)

type Printer interface {
	Enter()
	Exit()
	P(...interface{}) Printer
}

type Visitor interface {
	Content() ([]byte, error)
	Visit(parent *Symbol, f *protogen.File)
	VisitEnum(parent *Symbol, enumDescriptor *protogen.Enum)
	VisitMessage(parent *Symbol, p *protogen.Message, typ GQLType)
	VisitField(symbol *Symbol, p *protogen.Field, typ GQLType)
	VisitOneOf(symbol *Symbol, p *protogen.Oneof, typ GQLType)
	VisitService(symbol *Symbol, p *protogen.Service)
}

type visitor struct {
	*protogen.GeneratedFile
	*protogen.File
	indent []string
	root   *Symbol
}

func NewVisitor(f *protogen.File, g *protogen.GeneratedFile, importPath string) Visitor {
	v := &visitor{g, f, make([]string, 0), &Symbol{}}
	pkgName := strings.ReplaceAll(filepath.Base(importPath), "\"", "")
	v.P("package ", pkgName)

	return v
}

func (v *visitor) Enter() {
	v.indent = append(v.indent, "\t")
}

func (v *visitor) Exit() {
	v.indent = v.indent[1:]
}

func (v *visitor) P(data ...interface{}) Printer {
	output := []interface{}{strings.Join(v.indent, "")}
	output = append(output, data...)
	v.GeneratedFile.P(output...)
	return v
}

func (v *visitor) Visit(root *Symbol, p *protogen.File) {
	registerType := goIdent(edgeImport, "RegisterType")
	for _, enum := range p.Enums {
		v.VisitEnum(root, enum)
	}
	for _, msg := range p.Messages {
		v.VisitMessage(root, msg, GQLTypeObject)
	}
	for _, svc := range p.Services {
		v.VisitService(root, svc)
	}
	v.P("func init() {")
	v.Enter()
	for _, sym := range tbl.mapSymbols {
		if sym.Ident.GoImportPath != v.GoImportPath {
			continue
		}
		switch sym.Ident.Type {
		case GQLTypeEnum:
			fallthrough
		case GQLTypeInput:
			fallthrough
		case GQLTypeObject:
			v.P(registerType, "(", sym.Ident.String(), ")")
		}
	}
	v.Exit()
	v.P("}")
}

func (v *visitor) VisitOneOf(root *Symbol, p *protogen.Oneof, typ GQLType) {
	gqlUnion := goIdent(graphqlImport, "Union")
	gqlNewUnion := goIdent(graphqlImport, "NewUnion")
	gqlUnionConfig := goIdent(graphqlImport, "UnionConfig")
	gqlObject := goIdent(graphqlImport, "Object")
	v.P("var ", typ, "_", p.GoIdent, " *", gqlUnion, " = ", gqlNewUnion, "(", gqlUnionConfig, "{")
	v.Enter()
	v.P("Name: ", `"`, typ, "_", p.GoIdent, `",`)
	v.P("Types: []*", gqlObject, "{")
	v.Enter()
	for _, f := range p.Fields {
		v.P(v.getEdgeType(f.Desc.Kind(), f.GoIdent, f, typ), ",")
	}
	v.Exit()
	v.P("},")
	gqlResolveTypeParams := protogen.GoIdent{
		GoName:       "ResolveTypeParams",
		GoImportPath: graphqlImport,
	}
	v.P("ResolveType: func(p ", gqlResolveTypeParams, ") *", gqlObject, "{")
	v.Enter()
	v.P("switch p.Value.(type) {")
	for _, f := range p.Fields {
		v.P("case *", f.Message.GoIdent, ":")
		v.Enter()
		v.P("return ", v.getEdgeType(f.Desc.Kind(), f.GoIdent, f, typ), "")
		v.Exit()
	}
	v.P("default:")
	v.Enter()
	v.P("return nil")
	v.Exit()
	v.P("}")
	v.Exit()
	v.P("},")
	v.Exit()
	v.P("})")
}

func (v *visitor) VisitOneOfField(root *Symbol, o *protogen.Oneof, typ GQLType) {
	gqlResolveParams := goIdent(graphqlImport, "ResolveParams")
	v.P(quot(string(o.Desc.Name())), ": &", goIdent(graphqlImport, "Field{"))
	v.Enter()
	v.P("Type: ", typ, "_", o.GoIdent, ",")
	v.P("Resolve: func(p ", gqlResolveParams, ") (interface{}, error) {")
	v.Enter()
	{
		v.P("if pdata, ok := p.Source.(*", o.Parent.GoIdent, "); ok {")
		v.Enter()
		{
			v.P("data := pdata.", o.GoName)
			for _, f := range o.Fields {
				v.P("if d, ok := data.(*", f.GoIdent, "); ok {")
				v.Enter()
				{
					v.P("return d.", f.GoName, ", nil")
				}
				v.Exit()
				v.P("}")
			}
		}
		v.Exit()
		v.P("}")
		v.P("return nil, nil")
	}
	v.Exit()
	v.P("},")
	v.Exit()
	v.P("},")
}

func (v *visitor) VisitField(symbol *Symbol, p *protogen.Field, typ GQLType) {
	isList := p.Desc.IsList()
	isEnum := p.Enum != nil
	isMap := p.Desc.IsMap()
	switch {
	case isMap:
		v.visitMapField(symbol, p, typ)
	case isEnum:
		v.visitEnumField(symbol, p, typ, isList)
	default:
		v.visitGeneralField(symbol, p, typ, isList)
	}
}

func (v *visitor) visitEnumField(symbol *Symbol, p *protogen.Field, typ GQLType, isList bool) {
	fieldType := v.getEdgeType(p.Desc.Kind(), p.GoIdent, p, typ)
	gqlField := goIdent(graphqlImport, "Field")
	gqlList := goIdent(graphqlImport, "NewList")
	gqlResolveParams := goIdent(graphqlImport, "ResolveParams")
	if typ == GQLTypeInput {
		gqlField = goIdent(graphqlImport, "InputObjectFieldConfig")
	}
	v.P(quot(p.Desc.JSONName()), ": &", gqlField, "{")
	v.Enter()
	if isList {
		v.P("Type: ", gqlList, "(", fieldType, "),")
	} else {
		v.P("Type: ", fieldType, ",")
	}
	if typ == GQLTypeObject {
		v.P("Resolve: func(p ", gqlResolveParams, ") (interface{}, error) {")
		v.Enter()
		{
			v.P("var res interface{}")
			v.P("if pdata, ok := p.Source.(*", p.Parent.GoIdent, "); ok {")
			v.Enter()
			{
				v.P("res = pdata.", p.GoName, ".String()")
			}
			v.Exit()
			v.P("}")
			v.P("return res, nil")
		}
		v.Exit()
		v.P("},")
	}
	v.Exit()
	v.P("},")
}

func (v *visitor) visitOneOfField(symbol *Symbol, p *protogen.Field, typ GQLType, isList bool) {
	fieldType := v.getEdgeType(p.Desc.Kind(), p.GoIdent, p, typ)
	gqlField := goIdent(graphqlImport, "Field")
	gqlList := goIdent(graphqlImport, "NewList")
	gqlResolveParams := goIdent(graphqlImport, "ResolveParams")
	if typ == GQLTypeInput {
		gqlField = goIdent(graphqlImport, "InputObjectFieldConfig")
	}
	v.P(quot(p.Desc.JSONName()), ": &", gqlField, "{")
	v.Enter()
	if isList {
		v.P("Type: ", gqlList, "(", fieldType, "),")
	} else {
		v.P("Type: ", fieldType, ",")
	}
	if typ == GQLTypeObject {
		v.P("Resolve: func(p ", gqlResolveParams, ") (interface{}, error) {")
		v.Enter()
		{
		}
		v.Exit()
		v.P("},")
	}
	v.Exit()
	v.P("},")
}

func (v *visitor) visitMapField(symbol *Symbol, p *protogen.Field, typ GQLType) {
	gqlField := goIdent(graphqlImport, "Field")
	gqlJson := goIdent(edgeImport, "Scalar_JSON")
	if typ == GQLTypeInput {
		gqlField = goIdent(graphqlImport, "InputObjectFieldConfig")
	}
	v.P(quot(p.Desc.JSONName()), ": &", gqlField, "{")
	v.Enter()
	v.P("Type: ", gqlJson, ",")
	v.Exit()
	v.P("},")
}

func (v *visitor) visitGeneralField(symbol *Symbol, p *protogen.Field, typ GQLType, isList bool) {
	fieldType := v.getEdgeType(p.Desc.Kind(), p.GoIdent, p, typ)
	gqlField := goIdent(graphqlImport, "Field")
	gqlList := goIdent(graphqlImport, "NewList")
	gqlResolveParams := goIdent(graphqlImport, "ResolveParams")
	if typ == GQLTypeInput {
		gqlField = goIdent(graphqlImport, "InputObjectFieldConfig")
	}
	v.P(quot(p.Desc.JSONName()), ": &", gqlField, "{")
	v.Enter()
	if isList {
		v.P("Type: ", gqlList, "(", fieldType, "),")
	} else {
		v.P("Type: ", fieldType, ",")
	}
	if typ == GQLTypeObject {
		v.P("Resolve: func(p ", gqlResolveParams, ") (interface{}, error) {")
		v.Enter()
		{
			v.P("var res interface{}")
			v.P("if pdata, ok := p.Source.(*", p.Parent.GoIdent, "); ok {")
			v.Enter()
			{
				v.P("res = pdata.", p.GoName)
			}
			v.Exit()
			v.P("}")
			v.P("return res, nil")
		}
		v.Exit()
		v.P("},")
	}
	v.Exit()
	v.P("},")
}

func (v *visitor) VisitEnum(parent *Symbol, p *protogen.Enum) {
	ident := GQLIdent{
		p.GoIdent,
		GQLTypeEnum,
		v.GeneratedFile,
	}
	if tbl.Exist(ident) {
		return
	}
	gqlEnum := goIdent(graphqlImport, "Enum")
	gqlNewEnum := goIdent(graphqlImport, "NewEnum")
	gqlEnumConfig := goIdent(graphqlImport, "EnumConfig")
	gqlEnumValueConfigMap := goIdent(graphqlImport, "EnumValueConfigMap")
	gqlEnumValueConfig := goIdent(graphqlImport, "EnumValueConfig")
	sym := NewSymbol(parent, ident)
	v.P("var ", ident.String(), " *", gqlEnum, "= ", gqlNewEnum, "(")
	v.Enter()
	v.P(gqlEnumConfig, "{")
	v.Enter()
	v.P("Name: ", quot(ident.String()), ",")
	v.P("Values: ", gqlEnumValueConfigMap, "{")
	v.Enter()
	for _, val := range p.Values {
		v.P(quot(string(val.Desc.Name())), ": &", gqlEnumValueConfig, "{")
		v.Enter()
		v.P("Value: ", val.Parent.GoIdent.GoName, "_name[", val.Desc.Index(), "],")
		v.Exit()
		v.P("},")
	}
	v.Exit()
	v.P("},")
	v.Exit()
	v.P("},")
	v.Exit()
	v.P(")")
	tbl.Append(sym)
}

func (v *visitor) VisitMessage(parent *Symbol, p *protogen.Message, typ GQLType) {
	ident := GQLIdent{
		p.GoIdent,
		typ,
		v.GeneratedFile,
	}
	if p.Desc.IsMapEntry() || tbl.Exist(ident) {
		return
	}
	sym := NewSymbol(parent, ident)
	for _, e := range p.Enums {
		v.VisitEnum(sym, e)
	}
	for _, o := range p.Oneofs {
		v.VisitOneOf(sym, o, typ)
	}
	for _, m := range p.Messages {
		if m.Desc.IsMapEntry() {
			continue
		}
		v.VisitMessage(sym, m, typ)
	}
	gqlObject := goIdent(graphqlImport, "Object")
	gqlNewObject := goIdent(graphqlImport, "NewObject")
	gqlObjectConfig := goIdent(graphqlImport, "ObjectConfig")
	gqlIsTypeOfParams := goIdent(graphqlImport, "IsTypeOfParams")
	gqlFields := goIdent(graphqlImport, "Fields")
	gqlInputObject := goIdent(graphqlImport, "InputObject")
	gqlNewInputObject := goIdent(graphqlImport, "NewInputObject")
	gqlInputObjectConfig := goIdent(graphqlImport, "InputObjectConfig")
	gqlInputObjectConfigFieldMap := goIdent(graphqlImport, "InputObjectConfigFieldMap")
	switch typ {
	case GQLTypeObject:
		v.P("var ", sym.Ident.String(), " *", gqlObject, " = ", gqlNewObject, "(")
		v.Enter()
		v.P(gqlObjectConfig, "{")
		v.Enter()
		v.P("Name: ", quot(sym.Ident.String()), ",")
		v.P("IsTypeOf: func(g ", gqlIsTypeOfParams, ") bool {")
		v.Enter()
		v.P("return true")
		v.Exit()
		v.P("},")
		v.P("Fields: ", gqlFields, "{")
		v.Enter()
		for _, f := range p.Fields {
			if f.Oneof != nil {
				continue
			}
			v.VisitField(sym, f, typ)
		}
		for _, o := range p.Oneofs {
			v.VisitOneOfField(root, o, typ)
		}
		v.Exit()
		v.P("},")
		v.Exit()
		v.P("},")
		v.Exit()
		v.P(")")
	case GQLTypeInput:
		v.P("var ", sym.Ident.String(), " *", gqlInputObject, " = ", gqlNewInputObject, "(")
		v.Enter()
		v.P(gqlInputObjectConfig, "{")
		v.Enter()
		v.P("Name: ", quot(sym.Ident.String()), ",")
		v.P("Fields: ", gqlInputObjectConfigFieldMap, "{")
		v.Enter()
		for _, f := range p.Fields {
			v.VisitField(sym, f, typ)
		}
		v.Exit()
		v.P("},")
		v.Exit()
		v.P("},")
		v.Exit()
		v.P(")")
	default:
	}
	tbl.Append(sym)
	for _, f := range p.Fields {
		if f.Message != nil {
			v.VisitMessage(root, f.Message, typ)
		}
	}
}

func (v *visitor) VisitService(symbol *Symbol, p *protogen.Service) {
	queries := make(map[string]*protogen.Method)
	mutations := make(map[string]*protogen.Method)
	inputs := make(map[*protogen.Message]struct{})
	for _, rpc := range p.Methods {
		edgeOpt := proto.GetExtension(rpc.Desc.Options(), graphql.E_Type)
		if edgeOpt == nil {
			continue
		}
		opt, ok := edgeOpt.(*graphql.GraphQLOption)
		if !ok {
			continue
		}
		queryName := opt.GetQuery()
		mutationName := opt.GetMutation()
		if queryName != "" && mutationName != "" {
			panic("conflict graphql operator for method: " + rpc.GoName)
		}
		if queryName != "" {
			queries[queryName] = rpc
		}
		if mutationName != "" {
			mutations[mutationName] = rpc
		}
		inputs[rpc.Input] = struct{}{}
	}
	for m := range inputs {
		v.VisitMessage(root, m, GQLTypeInput)
	}
	if len(queries) > 0 {
		v.P("func Register", p.GoName, "Queries(sc ", p.GoName, "Client) error {")
		v.Enter()
		for name, q := range queries {
			v.visitMethod(symbol, q, name, GQLTypeQuery)
		}
		v.P("return nil")
		v.Exit()
		v.P("}")
	}
	v.P("")
	if len(mutations) > 0 {
		v.P("func Register", p.GoName, "Mutations(sc ", p.GoName, "Client) error {")
		v.Enter()
		for name, m := range mutations {
			v.visitMethod(symbol, m, name, GQLTypeMutation)
		}
		v.P("return nil")
		v.Exit()
		v.P("}")
	}
}

func (v *visitor) visitMethod(symbol *Symbol, p *protogen.Method, optionName string, methodType GQLType) {
	jsonMarshal := goIdent("encoding/json", "Marshal")
	jsonUnmarshal := goIdent("google.golang.org/protobuf/encoding/protojson", "Unmarshal")
	gqlField := goIdent(graphqlImport, "Field")
	switch methodType {
	case GQLTypeQuery:
		edgeQuery := goIdent(edgeImport, "RegisterQuery")
		v.P(edgeQuery, "(", quot(optionName), ", &", gqlField, "{")
	case GQLTypeMutation:
		edgeMutation := goIdent(edgeImport, "RegisterMutation")
		v.P(edgeMutation, "(", quot(optionName), ", &", gqlField, "{")
	default:
		panic("graphql method must be `query` or `mutation`, got: " + optionName)
	}
	gqlFieldConfigArgument := goIdent(graphqlImport, "FieldConfigArgument")
	gqlArgumentConfig := goIdent(graphqlImport, "ArgumentConfig")
	gqlResolveParams := goIdent(graphqlImport, "ResolveParams")
	v.Enter()
	v.P("Name: ", quot(optionName), ",")
	v.P("Args: ", gqlFieldConfigArgument, "{")
	v.Enter()
	v.P(quot("input"), ": &", gqlArgumentConfig, "{")
	v.Enter()
	input := v.getType(protoreflect.MessageKind, p.Input.GoIdent, p.Input.Desc, GQLTypeInput)
	v.P("Type: ", input, ",")
	v.Exit()
	v.P("},")
	v.Exit()
	v.P("},")
	output := v.getType(protoreflect.MessageKind, p.Output.GoIdent, p.Output.Desc, GQLTypeObject)
	v.P("Type: ", output, ",")
	v.P("Resolve: func(p ", gqlResolveParams, ") (interface{}, error) {")
	v.Enter()
	v.P("var req ", p.Input.GoIdent)
	v.P("rawJson, err := ", jsonMarshal, "(p.Args[", quot("input"), "])")
	v.P("if err != nil {")
	v.Enter()
	v.P("return nil, err")
	v.Exit()
	v.P("}")
	v.P("err = ", jsonUnmarshal, "(rawJson, &req)")
	v.P("if err != nil {")
	v.Enter()
	v.P("return nil, err")
	v.Exit()
	v.P("}")
	v.P("var res *", p.Output.GoIdent)
	v.P("res, err = sc.", p.GoName, "(p.Context, &req)")
	v.P("return res, err")
	v.Exit()
	v.P("},")
	v.Exit()
	v.P("})")
}

func (v *visitor) getType(kind protoreflect.Kind, ident protogen.GoIdent, desc protoreflect.Descriptor, typ GQLType) protogen.GoIdent {
	wellKnownImports := map[string]GQLIdent{
		"google.protobuf.Empty": {
			GoIdent: protogen.GoIdent{
				GoName:       "Empty",
				GoImportPath: "google.golang.org/protobuf/types/known/emptypb",
			},
			Type: GQLTypeScalar,
			g:    v.GeneratedFile,
		},
		"google.protobuf.Timestamp": {
			GoIdent: protogen.GoIdent{
				GoName:       "Timestamp",
				GoImportPath: "google.golang.org/protobuf/types/known/timestamppb",
			},
			Type: GQLTypeScalar,
			g:    v.GeneratedFile,
		},
		"google.protobuf.Duration": {
			GoIdent: protogen.GoIdent{
				GoName:       "Duration",
				GoImportPath: "google.golang.org/protobuf/types/known/durationpb",
			},
			Type: GQLTypeScalar,
			g:    v.GeneratedFile,
		},
	}
	wrappers := []string{
		"BoolValue", "StringValue", "FloatValue", "Int64Value",
		"Int32Value", "UInt64Value", "UInt32Value", "SInt64Value",
		"SInt32Value", "Fixed64Value", "Fixed32Value", "SFixed64Value",
		"SFixed32Value",
	}
	for _, t := range wrappers {
		wellKnownImports["google.protobuf."+t] = GQLIdent{
			GoIdent: protogen.GoIdent{
				GoName:       t,
				GoImportPath: "google.golang.org/protobuf/types/known/wrapperspb",
			},
			Type: typ,
			g:    v.GeneratedFile,
		}
	}
	switch kind {
	case protoreflect.FloatKind:
		fallthrough
	case protoreflect.DoubleKind:
		return protogen.GoIdent{
			GoName:       "Float",
			GoImportPath: graphqlImport,
		}
	case protoreflect.Int64Kind:
		fallthrough
	case protoreflect.Int32Kind:
		fallthrough
	case protoreflect.Uint64Kind:
		fallthrough
	case protoreflect.Uint32Kind:
		fallthrough
	case protoreflect.Sint64Kind:
		fallthrough
	case protoreflect.Sint32Kind:
		fallthrough
	case protoreflect.Fixed64Kind:
		fallthrough
	case protoreflect.Fixed32Kind:
		fallthrough
	case protoreflect.Sfixed64Kind:
		fallthrough
	case protoreflect.Sfixed32Kind:
		return protogen.GoIdent{
			GoName:       "Int",
			GoImportPath: graphqlImport,
		}
	case protoreflect.BoolKind:
		return protogen.GoIdent{
			GoName:       "Boolean",
			GoImportPath: graphqlImport,
		}
	case protoreflect.BytesKind:
		return protogen.GoIdent{
			GoName:       "Scalar_bytes",
			GoImportPath: edgeImport,
		}
	case protoreflect.StringKind:
		return protogen.GoIdent{
			GoName:       "String",
			GoImportPath: graphqlImport,
		}
	case protoreflect.EnumKind:
		ident := GQLIdent{
			ident,
			GQLTypeEnum,
			v.GeneratedFile,
		}
		return protogen.GoIdent{
			GoName:       ident.String(),
			GoImportPath: ident.GoImportPath,
		}
	case protoreflect.MessageKind:
		if ident, ok := wellKnownImports[string(desc.FullName())]; ok {
			return protogen.GoIdent{
				GoName:       ident.String(),
				GoImportPath: edgeImport,
			}
		}
		return protogen.GoIdent{
			GoName:       string(typ) + "_" + string(ident.GoName),
			GoImportPath: ident.GoImportPath,
		}
	}
	panic("failed to get type for: " + ident.String())
}

func (v *visitor) getEdgeType(kind protoreflect.Kind, ident protogen.GoIdent, p *protogen.Field, typ GQLType) protogen.GoIdent {
	switch kind {
	case protoreflect.EnumKind:
		return v.getType(kind, p.Enum.GoIdent, p.Enum.Desc, typ)
	case protoreflect.MessageKind:
		return v.getType(kind, p.Message.GoIdent, p.Message.Desc, typ)
	default:
		return v.getType(kind, ident, p.Desc, typ)
	}
}

func quot(str string) string {
	return strconv.Quote(str)
}

func goIdent(imp protogen.GoImportPath, name string) protogen.GoIdent {
	return protogen.GoIdent{GoName: name, GoImportPath: imp}
}
