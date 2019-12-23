package graphql

import (
	"errors"

	graphql "github.com/graphql-go/graphql"
)

var (
	ErrDuplicateMutation error = errors.New("Duplicate mutation")
	ErrDuplicateQuery          = errors.New("Duplicate query")
)

var (
	types     []graphql.Type = make([]graphql.Type, 0)
	queries   graphql.Fields = graphql.Fields{}
	mutations                = graphql.Fields{}
)

func init() {
	RegisterType(GraphQL_Empty)
	RegisterType(GraphQL_Duration)
	RegisterType(GraphQL_Timestamp)
	RegisterType(GraphQL_BoolValue)
	RegisterType(GraphQL_BoolValueInput)
	RegisterType(GraphQL_StringValue)
	RegisterType(GraphQL_StringValueInput)
	RegisterType(GraphQL_FloatValue)
	RegisterType(GraphQL_FloatValueInput)
	RegisterType(GraphQL_DoubleValue)
	RegisterType(GraphQL_DoubleValueInput)
	RegisterType(GraphQL_Int32Value)
	RegisterType(GraphQL_Int32ValueInput)
	RegisterType(GraphQL_UInt32Value)
	RegisterType(GraphQL_UInt32ValueInput)
	RegisterType(GraphQL_Int64Value)
	RegisterType(GraphQL_Int64ValueInput)
	RegisterType(GraphQL_UInt64Value)
	RegisterType(GraphQL_UInt64ValueInput)
}

func RegisterType(newType graphql.Type) {
	types = append(types, newType)
}

func RegisterQuery(name string, field *graphql.Field) error {
	if _, exist := queries[name]; exist {
		return ErrDuplicateQuery
	}

	queries[name] = field
	return nil
}

func RegisterMutation(name string, field *graphql.Field) error {
	if _, exist := mutations[name]; exist {
		return ErrDuplicateMutation
	}

	mutations[name] = field
	return nil
}

func GetSchema() (*graphql.Schema, error) {
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: queries}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mutations}
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
		Types:    types,
	}
	schema, err := graphql.NewSchema(schemaConfig)
	return &schema, err
}
