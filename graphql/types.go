package graphql

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	graphql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

const (
	jsISOString string = "2006-01-02T15:04:05.999Z07:00"
)

func parseEmptyValue(interface{}) interface{} {
	return nil
}

func parseEmptyLiteral(valueAST ast.Value) interface{} {
	return nil
}

func parseTimestampValue(value interface{}) interface{} {
	switch t := value.(type) {
	case timestamp.Timestamp:
		return time.Unix(t.GetSeconds(), int64(t.GetNanos()))
	case *timestamp.Timestamp:
		return time.Unix(t.GetSeconds(), int64(t.GetNanos()))
	default:
		return fmt.Errorf("Invalid DateTime value: %#v", value)
	}
}

func parseTimestampLiteral(valueAST ast.Value) interface{} {
	date, err := time.Parse(jsISOString, valueAST.GetValue().(string))
	if err != nil {
		return err
	}
	seconds := time.Duration(date.UnixNano()) / time.Second
	nanos := int32(time.Duration(date.UnixNano()) - (time.Duration(seconds) * time.Second))
	return timestamp.Timestamp{
		Seconds: int64(seconds),
		Nanos:   nanos,
	}
}

func parseDurationValue(value interface{}) interface{} {
	switch t := value.(type) {
	case duration.Duration:
		d := (time.Duration(t.GetSeconds()) * time.Second) + time.Duration(t.GetNanos())
		return d.String()
	case *duration.Duration:
		d := (time.Duration(t.GetSeconds()) * time.Second) + time.Duration(t.GetNanos())
		return d.String()
	default:
		return fmt.Errorf("Invalid Duration value: %#v", value)
	}
}

func parseDurationLiteral(valueAST ast.Value) interface{} {
	raw, err := time.ParseDuration(valueAST.GetValue().(string))
	if err != nil {
		return err
	}
	return duration.Duration{
		Seconds: int64(raw.Seconds()),
		Nanos:   int32(raw.Nanoseconds()),
	}
}

var GraphQL_Empty *graphql.Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "Empty",
	Description:  "empty value `null`",
	ParseValue:   parseEmptyValue,
	Serialize:    parseEmptyValue,
	ParseLiteral: parseEmptyLiteral,
})

var GraphQL_Timestamp *graphql.Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "Timestamp",
	ParseValue:   parseTimestampValue,
	Serialize:    parseTimestampValue,
	ParseLiteral: parseTimestampLiteral,
})

var GraphQL_Duration *graphql.Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "Duration",
	ParseValue:   parseDurationValue,
	Serialize:    parseDurationValue,
	ParseLiteral: parseDurationLiteral,
})

var GraphQL_BoolValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "BoolValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
	},
})

var GraphQL_BoolValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "BoolValue",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var GraphQL_StringValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "StringValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var GraphQL_StringValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "StringValue",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var GraphQL_Int32ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "Int32ValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_Int32Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Int32Value",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt32ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UInt32ValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt32Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "UInt32Value",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_Int64ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "Int64ValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_Int64Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Int64Value",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt64ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UInt64ValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt64Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "UInt64Value",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_FloatValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "FloatValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
	},
})

var GraphQL_FloatValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "FloatValue",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Float,
		},
	},
})

var GraphQL_DoubleValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DoubleValueInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
	},
})

var GraphQL_DoubleValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "DoubleValue",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Float,
		},
	},
})
