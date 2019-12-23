package generator_test

import (
	"testing"

	generator "github.com/ncrypthic/graphql-grpc-edge/generator"
)

func TestTypeInfo(t *testing.T) {
	types := []generator.TypeInfo{
		generator.TypeInfo{
			Name:     "String",
			Prefix:   "graphql",
			IsScalar: true,
		},
		generator.TypeInfo{
			Name:       "String",
			Prefix:     "graphql",
			IsScalar:   true,
			IsRepeated: true,
		},
		generator.TypeInfo{
			Name:      "String",
			Prefix:    "graphql",
			IsScalar:  true,
			IsNonNull: true,
		},
		generator.TypeInfo{
			Name:       "String",
			Prefix:     "graphql",
			IsScalar:   true,
			IsRepeated: true,
			IsNonNull:  true,
		},
		generator.TypeInfo{
			Name:   "Empty",
			Suffix: "Input",
		},
		generator.TypeInfo{
			Name:       "Empty",
			Suffix:     "Input",
			IsRepeated: true,
		},
		generator.TypeInfo{
			Name:      "Empty",
			Suffix:    "Input",
			IsNonNull: true,
		},
		generator.TypeInfo{
			Name:       "Empty",
			Suffix:     "Input",
			IsRepeated: true,
			IsNonNull:  true,
		},
	}
	results := []string{
		"graphql.String",
		"graphql.NewList(graphql.String)",
		"graphql.NewNonNull(graphql.String)",
		"graphql.NewList(graphql.NewNonNull(graphql.String))",
		"GraphQL_EmptyInput",
		"graphql.NewList(GraphQL_EmptyInput)",
		"graphql.NewNonNull(GraphQL_EmptyInput)",
		"graphql.NewList(graphql.NewNonNull(GraphQL_EmptyInput))",
	}
	for idx, typ := range types {
		if results[idx] != typ.GetName() {
			t.Fatalf("TypeInfo.GetName %s != %s", results[idx], typ.GetName())
		}
	}
}

func TestGetFieldName(t *testing.T) {
	g := generator.NewGenerator(generator.DefaultNameGenerator, "")
	pbFieldNames := []string{
		"first_name",
		"firstName",
		"very_long_name",
		"veryLongName",
	}
	graphqlFieldNames := []string{
		"firstName",
		"firstName",
		"veryLongName",
		"veryLongName",
	}
	for idx, fieldName := range pbFieldNames {
		if graphqlFieldNames[idx] != g.GetFieldName(fieldName) {
			t.Fatalf("Generator.GetFieldName %s != %s", graphqlFieldNames[idx], g.GetFieldName(fieldName))
		}
	}
}
