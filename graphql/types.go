package graphql

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	. "github.com/graphql-go/graphql"
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

func parseBytesValue(val interface{}) interface{} {
	b, ok := val.([]byte)
	if !ok {
		return errors.New("invalid bytes input value")
	}
	return b
}

func parseBytesLiteral(valueAST ast.Value) interface{} {
	s, ok := valueAST.GetValue().(string)
	if !ok {
		return errors.New("invalid bytes input value")
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return b
}

func serializeBytesValue(value interface{}) interface{} {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("invalid bytes value")
	}
	return base64.StdEncoding.EncodeToString(b)
}

func parseJSONValue(val interface{}) interface{} {
	return val
}

func parseJSONLiteral(valueAST ast.Value) interface{} {
	m := make(map[string]interface{})
	switch t := valueAST.GetValue().(type) {
	case string:
		err := json.Unmarshal([]byte(t), &m)
		if err != nil {
			return nil
		}
	case []byte:
		err := json.Unmarshal(t, &m)
		if err != nil {
			return nil
		}
	}
	return m
}

func serializeJSONValue(value interface{}) interface{} {
	_, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return value
}

var Scalar_emptypb_Empty *Scalar = NewScalar(ScalarConfig{
	Name:         "Empty",
	Description:  "Empty accepts only `null` value",
	ParseValue:   parseEmptyValue,
	Serialize:    parseEmptyValue,
	ParseLiteral: parseEmptyLiteral,
})

var Scalar_timestamppb_Timestamp *Scalar = NewScalar(ScalarConfig{
	Name:         "Timestamp",
	ParseValue:   parseTimestampValue,
	Serialize:    serializeTimestampValue,
	ParseLiteral: parseTimestampLiteral,
})

var Scalar_durationpb_Duration *Scalar = NewScalar(ScalarConfig{
	Name:         "Duration",
	Description:  "Duration represent time duration",
	ParseValue:   parseDurationValue,
	Serialize:    serializeDurationValue,
	ParseLiteral: parseDurationLiteral,
})

var Scalar_bytes *Scalar = NewScalar(ScalarConfig{
	Name:         "bytes",
	Description:  "base64 encoded bytes value",
	ParseValue:   parseBytesValue,
	Serialize:    serializeBytesValue,
	ParseLiteral: parseBytesLiteral,
})

var Scalar_JSON *Scalar = NewScalar(ScalarConfig{
	Name:         "JSON",
	Description:  "JSON map object",
	ParseValue:   parseJSONValue,
	Serialize:    serializeJSONValue,
	ParseLiteral: parseJSONLiteral,
})

var Input_wrapperspb_FloatValue *InputObject = NewInputObject(InputObjectConfig{
	Name:        "FloatValueInput",
	Description: "FloatValueInput accepts `null` or an object with `value` field typed float",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Float,
		},
	},
})

var Object_wrapperspb_FloatValue *Object = NewObject(ObjectConfig{
	Name:        "FloatValue",
	Description: "FloatValue returns `null` or an object with `value` field typed float",
	Fields: Fields{
		"value": &Field{
			Type: Float,
		},
	},
})

var Input_wrapperspb_DoubleValue *InputObject = NewInputObject(InputObjectConfig{
	Name:        "DoubleValueInput",
	Description: "DoubleValueInput accepts `null` or an object with `value` field typed double",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Float,
		},
	},
})

var Object_wrapperspb_DoubleValue *Object = NewObject(ObjectConfig{
	Name:        "DoubleValue",
	Description: "DoubleValue accepts `null` or an object with `value` field typed double",
	Fields: Fields{
		"value": &Field{
			Type: Float,
		},
	},
})

var Input_wrapperspb_Int64Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "Int64ValueInput",
	Description: "Int64ValueInput accepts `null` or an object with `value` field typed int64",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_Int64Value *Object = NewObject(ObjectConfig{
	Name:        "Int64Value",
	Description: "Int64Value returns `null` or an object with `value` field typed int64",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_Int32Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "Int32ValueInput",
	Description: "Int32ValueInput accepts `null` or an object with `value` field typed int32",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_Int32Value *Object = NewObject(ObjectConfig{
	Name:        "Int32Value",
	Description: "Int32Value returns `null` or an object with `value` field typed int32",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_UInt64Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "UInt64ValueInput",
	Description: "UInt64ValueInput accepts `null` or an object with `value` field typed uint64",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_UInt64Value *Object = NewObject(ObjectConfig{
	Name:        "UInt64Value",
	Description: "UInt64ValueInput returns `null` or an object with `value` field typed uint64",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_UInt32Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "UInt32ValueInput",
	Description: "UInt32ValueInput accepts `null` or an object with `value` field typed uint32",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_UInt32Value *Object = NewObject(ObjectConfig{
	Name:        "UInt32Value",
	Description: "UInt32Value returns `null` or an object with `value` field typed uint32",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_SInt64Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "SInt64ValueInput",
	Description: "SInt64ValueInput accepts `null` or an object with `value` field typed uint64",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_SInt64Value *Object = NewObject(ObjectConfig{
	Name:        "SInt64Value",
	Description: "SInt64ValueInput returns `null` or an object with `value` field typed uint64",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_SInt32Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "SInt32ValueInput",
	Description: "SInt32ValueInput accepts `null` or an object with `value` field typed uint32",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_SInt32Value *Object = NewObject(ObjectConfig{
	Name:        "SInt32Value",
	Description: "SInt32Value returns `null` or an object with `value` field typed uint32",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_Fixed64Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "Fixed64ValueInput",
	Description: "Fixed64ValueInput accepts `null` or an object with `value` field typed uint64",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_Fixed64Value *Object = NewObject(ObjectConfig{
	Name:        "Fixed64Value",
	Description: "Fixed64ValueInput returns `null` or an object with `value` field typed uint64",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_Fixed32Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "Fixed32ValueInput",
	Description: "Fixed32ValueInput accepts `null` or an object with `value` field typed uint32",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_Fixed32Value *Object = NewObject(ObjectConfig{
	Name:        "Fixed32Value",
	Description: "Fixed32Value returns `null` or an object with `value` field typed uint32",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_SFixed64Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "SFixed64ValueInput",
	Description: "SFixed64ValueInput accepts `null` or an object with `value` field typed uint64",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_SFixed64Value *Object = NewObject(ObjectConfig{
	Name:        "SFixed64Value",
	Description: "SFixed64ValueInput returns `null` or an object with `value` field typed uint64",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_SFixed32Value *InputObject = NewInputObject(InputObjectConfig{
	Name:        "SFixed32ValueInput",
	Description: "SFixed32ValueInput accepts `null` or an object with `value` field typed uint32",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Int,
		},
	},
})

var Object_wrapperspb_SFixed32Value *Object = NewObject(ObjectConfig{
	Name:        "SFixed32Value",
	Description: "SFixed32Value returns `null` or an object with `value` field typed uint32",
	Fields: Fields{
		"value": &Field{
			Type: Int,
		},
	},
})

var Input_wrapperspb_BoolValue *InputObject = NewInputObject(InputObjectConfig{
	Name:        "BoolValueInput",
	Description: "BoolValueInput accept `null` or an object with field `value` with type boolean",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: Boolean,
		},
	},
})

var Object_wrapperspb_BoolValue *Object = NewObject(ObjectConfig{
	Name:        "BoolValue",
	Description: "BoolValue returns `null` or an object with `value` field type boolean",
	Fields: Fields{
		"value": &Field{
			Type: Boolean,
		},
	},
})

var Input_wrapperspb_StringValue *InputObject = NewInputObject(InputObjectConfig{
	Name:        "StringValueInput",
	Description: "StringValueInput accepts string or null",
	Fields: InputObjectConfigFieldMap{
		"value": &InputObjectFieldConfig{
			Type: String,
		},
	},
})

var Object_wrapperspb_StringValue *Object = NewObject(ObjectConfig{
	Name:        "StringValue",
	Description: "StringValue returns `null` or an object with `value` field typed string",
	Fields: Fields{
		"value": &Field{
			Type: String,
		},
	},
})
