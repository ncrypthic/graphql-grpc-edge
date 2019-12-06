package sample

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	graphql "github.com/graphql-go/graphql"
)

var GraphQL_InputError *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputError",
	Fields: graphql.InputObjectConfigFieldMap{
		"error_message": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})


var GraphQL_InputFieldValidationError *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputFieldValidationError",
	Fields: graphql.InputObjectConfigFieldMap{
		"field": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"error": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})


var GraphQL_InputHello *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputHello",
	Fields: graphql.InputObjectConfigFieldMap{
		"name": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: GraphQL_HelloTypeEnum,
		},
		"messages": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
	},
})


var GraphQL_InputHelloRequest *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputHelloRequest",
	Fields: graphql.InputObjectConfigFieldMap{
		"name": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})


var GraphQL_InputHelloResponse *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputHelloResponse",
	Fields: graphql.InputObjectConfigFieldMap{
		"data": &graphql.InputObjectFieldConfig{
			Type: GraphQL_InputHello,
		},
	},
})


var GraphQL_InputServerError *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputServerError",
	Fields: graphql.InputObjectConfigFieldMap{
		"code": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"decription": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})


var GraphQL_InputValidationError *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "InputValidationError",
	Fields: graphql.InputObjectConfigFieldMap{
		"fields": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(GraphQL_InputFieldValidationError),
		},
	},
})


var GraphQL_Error *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Error",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"error_message": &graphql.Field{
			Name: "error_message",
			Type: graphql.String,
		},
	},
})


var GraphQL_FieldValidationError *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "FieldValidationError",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"field": &graphql.Field{
			Name: "field",
			Type: graphql.String,
		},
		"error": &graphql.Field{
			Name: "error",
			Type: graphql.String,
		},
	},
})


var GraphQL_Hello *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Hello",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Name: "name",
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Name: "type",
			Type: GraphQL_HelloTypeEnum,
		},
		"messages": &graphql.Field{
			Name: "messages",
			Type: graphql.NewList(graphql.String),
		},
	},
})


var GraphQL_HelloRequest *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "HelloRequest",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Name: "name",
			Type: graphql.String,
		},
	},
})


var GraphQL_HelloResponse *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "HelloResponse",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"data": &graphql.Field{
			Name: "data",
			Type: GraphQL_Hello,
		},"error": &graphql.Field{
			Name: "error",
			Type: GraphQL_HelloResponse_errorUnion,
		},
	},
})


var GraphQL_ServerError *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "ServerError",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"code": &graphql.Field{
			Name: "code",
			Type: graphql.Int,
		},
		"decription": &graphql.Field{
			Name: "decription",
			Type: graphql.String,
		},
	},
})


var GraphQL_ValidationError *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "ValidationError",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"fields": &graphql.Field{
			Name: "fields",
			Type: graphql.NewList(GraphQL_FieldValidationError),
		},
	},
})


var GraphQL_HelloTypeEnum *graphql.Enum = graphql.NewEnum(graphql.EnumConfig{
	Name: "HelloTypeEnum",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: "NONE",
		},
		"ANY": &graphql.EnumValueConfig{
			Value: "ANY",
		},
	},
})

var GraphQL_HelloResponse_errorUnion *graphql.Union = graphql.NewUnion(graphql.UnionConfig{
	Name: "HelloResponse_errorUnion",
	Types: []*graphql.Object{
		GraphQL_ServerError,
		GraphQL_ValidationError,
	},
})

func RegisterHelloServiceQueries(queries graphql.Fields, sc HelloServiceClient) error {
	queries["greeting"] = &graphql.Field{
		Name: "greeting",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: GraphQL_InputHello,
			},
		},
		Type: GraphQL_HelloResponse,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			inputArgs, ok := p.Args["input"].(map[string]interface{})
			if !ok || inputArgs == nil {
				return nil, errors.New("Missing required parameter `input`")
			}
			var req Hello
			mapErr := mapstructure.Decode(inputArgs, &req)
			if mapErr != nil {
				return nil, mapErr
			}
			return sc.Greeting(p.Context, &req)
		},
	}

	return nil
}

func RegisterHelloServiceMutations(mutations graphql.Fields, sc HelloServiceClient) error {
	mutations["setGreeting"] = &graphql.Field{
		Name: "setGreeting",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: GraphQL_InputHello,
			},
		},
		Type: GraphQL_HelloResponse,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			inputArgs, ok := p.Args["input"].(map[string]interface{})
			if !ok || inputArgs == nil {
				return nil, errors.New("Missing required parameter `input`")
			}
			var req Hello
			mapErr := mapstructure.Decode(inputArgs, &req)
			if mapErr != nil {
				return nil, mapErr
			}
			return sc.SetGreeting(p.Context, &req)
		},
	}
	return nil
}


func RegisterGraphQLTypes(types []graphql.Type) {
	types = append(types, GraphQL_InputError)
	types = append(types, GraphQL_InputFieldValidationError)
	types = append(types, GraphQL_InputHello)
	types = append(types, GraphQL_InputHelloRequest)
	types = append(types, GraphQL_InputHelloResponse)
	types = append(types, GraphQL_InputServerError)
	types = append(types, GraphQL_InputValidationError)
	types = append(types, GraphQL_Error)
	types = append(types, GraphQL_FieldValidationError)
	types = append(types, GraphQL_Hello)
	types = append(types, GraphQL_HelloRequest)
	types = append(types, GraphQL_HelloResponse)
	types = append(types, GraphQL_ServerError)
	types = append(types, GraphQL_ValidationError)
	types = append(types, GraphQL_HelloTypeEnum)
	types = append(types, GraphQL_HelloResponse_errorUnion)
}
