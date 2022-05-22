package graphql

import (
	"errors"

	. "github.com/graphql-go/graphql"
)

var (
	ErrDuplicateMutation error = errors.New("Duplicate mutation")
	ErrDuplicateQuery          = errors.New("Duplicate query")
)

var (
	typeMap   map[string]Type = make(map[string]Type)
	types     []Type          = make([]Type, 0)
	queries   Fields          = Fields{}
	mutations                 = Fields{}
)

func init() {
	RegisterType(Scalar_bytes)
	RegisterType(Scalar_durationpb_Duration)
	RegisterType(Scalar_emptypb_Empty)
	RegisterType(Scalar_timestamppb_Timestamp)
	RegisterType(Object_wrapperspb_Fixed64Value)
	RegisterType(Object_wrapperspb_SFixed64Value)
	RegisterType(Object_wrapperspb_SInt64Value)
	RegisterType(Object_wrapperspb_UInt64Value)
	RegisterType(Object_wrapperspb_BoolValue)
	RegisterType(Object_wrapperspb_DoubleValue)
	RegisterType(Object_wrapperspb_Fixed32Value)
	RegisterType(Object_wrapperspb_FloatValue)
	RegisterType(Object_wrapperspb_Int32Value)
	RegisterType(Object_wrapperspb_Int64Value)
	RegisterType(Object_wrapperspb_SFixed32Value)
	RegisterType(Object_wrapperspb_SInt32Value)
	RegisterType(Object_wrapperspb_StringValue)
	RegisterType(Object_wrapperspb_UInt32Value)
	RegisterType(Input_wrapperspb_BoolValue)
	RegisterType(Input_wrapperspb_DoubleValue)
	RegisterType(Input_wrapperspb_Fixed64Value)
	RegisterType(Input_wrapperspb_FloatValue)
	RegisterType(Input_wrapperspb_Int32Value)
	RegisterType(Input_wrapperspb_SFixed64Value)
	RegisterType(Input_wrapperspb_SInt64Value)
	RegisterType(Input_wrapperspb_StringValue)
	RegisterType(Input_wrapperspb_UInt64Value)
	RegisterType(Input_wrapperspb_Fixed32Value)
	RegisterType(Input_wrapperspb_Int64Value)
	RegisterType(Input_wrapperspb_SFixed32Value)
	RegisterType(Input_wrapperspb_SInt32Value)
	RegisterType(Input_wrapperspb_UInt32Value)
}

func RegisterType(newType Type) {
	name := newType.Name()
	_, exists := typeMap[name]
	if !exists {
		types = append(types, newType)
		typeMap[name] = newType
		return
	}
	for idx, t := range types {
		if t.Name() == name {
			types[idx] = newType
			return
		}
	}
}

func LookupType(name string) (Type, bool) {
	for _, t := range types {
		if t.Name() == name {
			return t, true
		}
	}
	return nil, false
}

func RegisterQuery(name string, field *Field) error {
	if _, exist := queries[name]; exist {
		return ErrDuplicateQuery
	}

	queries[name] = field
	return nil
}

func RegisterMutation(name string, field *Field) error {
	if _, exist := mutations[name]; exist {
		return ErrDuplicateMutation
	}

	mutations[name] = field
	return nil
}

func GetSchema() (*Schema, error) {
	rootQuery := ObjectConfig{Name: "RootQuery", Fields: queries}
	rootMutation := ObjectConfig{Name: "RootMutation", Fields: mutations}
	schemaConfig := SchemaConfig{
		Query:    NewObject(rootQuery),
		Mutation: NewObject(rootMutation),
		Types:    types,
	}
	schema, err := NewSchema(schemaConfig)
	return &schema, err
}
