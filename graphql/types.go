package graphql

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	graphql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var (
	ErrBadValue         error = fmt.Errorf("invalid value")
	ErrUpstreamResponse error = fmt.Errorf("invalid upstream response")
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

func serializeTimestampValue(value interface{}) interface{} {
	switch t := value.(type) {
	case timestamp.Timestamp:
		ti := time.Unix(t.GetSeconds(), int64(t.GetNanos()))
		return ti.Format(jsISOString)
	case *timestamp.Timestamp:
		ti := time.Unix(t.GetSeconds(), int64(t.GetNanos()))
		return ti.Format(jsISOString)
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

func serializeDurationValue(value interface{}) interface{} {
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
	Description:  "Empty accepts only `null` value",
	ParseValue:   parseEmptyValue,
	Serialize:    parseEmptyValue,
	ParseLiteral: parseEmptyLiteral,
})

var GraphQL_Timestamp *graphql.Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "Timestamp",
	ParseValue:   parseTimestampValue,
	Serialize:    serializeTimestampValue,
	ParseLiteral: parseTimestampLiteral,
})

var GraphQL_Duration *graphql.Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "Duration",
	Description:  "Duration represent time duration",
	ParseValue:   parseDurationValue,
	Serialize:    serializeDurationValue,
	ParseLiteral: parseDurationLiteral,
})

var GraphQL_BoolValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "BoolValueInput",
	Description: "BoolValueInput accept `null` or an object with field `value` with type boolean",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
	},
})

var GraphQL_BoolValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "BoolValue",
	Description: "BoolValue returns `null` or an object with `value` field type boolean",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var GraphQL_StringValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "StringValueInput",
	Description: "StringValueInput accepts string or null",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

var GraphQL_StringValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "StringValue",
	Description: "StringValue returns `null` or an object with `value` field typed string",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var GraphQL_Int32ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "Int32ValueInput",
	Description: "Int32ValueInput accepts `null` or an object with `value` field typed int32",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_Int32Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Int32Value",
	Description: "Int32Value returns `null` or an object with `value` field typed int32",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt32ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "UInt32ValueInput",
	Description: "UInt32ValueInput accepts `null` or an object with `value` field typed uint32",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt32Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UInt32Value",
	Description: "UInt32Value returns `null` or an object with `value` field typed uint32",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_Int64ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "Int64ValueInput",
	Description: "Int64ValueInput accepts `null` or an object with `value` field typed int64",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_Int64Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Int64Value",
	Description: "Int64Value returns `null` or an object with `value` field typed int64",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt64ValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "UInt64ValueInput",
	Description: "UInt64ValueInput accepts `null` or an object with `value` field typed uint64",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var GraphQL_UInt64Value *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UInt64Value",
	Description: "UInt64ValueInput returns `null` or an object with `value` field typed uint64",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var GraphQL_FloatValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "FloatValueInput",
	Description: "FloatValueInput accepts `null` or an object with `value` field typed float",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
	},
})

var GraphQL_FloatValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "FloatValue",
	Description: "FloatValue returns `null` or an object with `value` field typed float",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Float,
		},
	},
})

var GraphQL_DoubleValueInput *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "DoubleValueInput",
	Description: "DoubleValueInput accepts `null` or an object with `value` field typed double",
	Fields: graphql.InputObjectConfigFieldMap{
		"value": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
	},
})

var GraphQL_DoubleValue *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DoubleValue",
	Description: "DoubleValue accepts `null` or an object with `value` field typed double",
	Fields: graphql.Fields{
		"value": &graphql.Field{
			Type: graphql.Float,
		},
	},
})
