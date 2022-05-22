package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	fds   *descriptorpb.FileDescriptorSet = &descriptorpb.FileDescriptorSet{}
	files *protoregistry.Files            = &protoregistry.Files{}
	p     *protogen.Plugin
)

func TestMain(m *testing.M) {
	os.RemoveAll("test.pb.descriptor")
	cmd := exec.Command("protoc", "-o", "test.pb.descriptor", "--include_imports", "-I", "../../", "-I", ".", "test.proto")
	err := cmd.Start()
	if err != nil {
		panic(fmt.Errorf("failed to generate protobuf descriptor: %s, %q", err.Error(), strings.Join(cmd.Args, " ")))
	}
	err = cmd.Wait()
	if err != nil {
		panic(fmt.Errorf("failed to generate protobuf descriptor: %s, %q", err.Error(), strings.Join(cmd.Args, " ")))
	}
	gengocmd := exec.Command("protoc", "--go_out=:.", "--go_opt=paths=source_relative", "-I", "../../", "-I", ".", "test.proto")
	err = gengocmd.Start()
	if err != nil {
		panic(fmt.Errorf("failed to generate protobuf descriptor: %s, %q", err.Error(), strings.Join(gengocmd.Args, " ")))
	}
	err = gengocmd.Wait()
	if err != nil {
		panic(fmt.Errorf("failed to generate protobuf descriptor: %s, %q", err.Error(), strings.Join(gengocmd.Args, " ")))
	}
	b, err := ioutil.ReadFile("./test.pb.descriptor")
	if err != nil {
		panic(fmt.Errorf("failed to parse proto descriptor: %w", err))
	}
	err = proto.Unmarshal(b, fds)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshall proto descriptor: %w", err))
	}
	files, err = protodesc.NewFiles(fds)
	if err != nil {
		panic(fmt.Errorf("failed to read FileDescriptorSet: %w", err))
	}
	p, err = protogen.Options{}.New(&pluginpb.CodeGeneratorRequest{
		ProtoFile: fds.File,
	})
	if err != nil {
		panic(fmt.Errorf("failed to generate plugin: %w", err))
	}

	os.Exit(m.Run())
}

func TestVisitEnum(t *testing.T) {
	f, ok := p.FilesByPath["test.proto"]
	if !ok {
		panic("failed to read FileDescriptor")
	}
	cases := []struct {
		name       string
		enum       *protogen.Enum
		wantString string
	}{
		{
			name: "positive case",
			enum: f.Enums[0],
			wantString: `package descriptorpb

import (
	generator "github.com/ncrypthic/graphql-grpc-edge/generator"
)

var Enum_generator_UserStatus *graphql.Enum = graphql.NewEnum(
	graphql.EnumConfig{
		Name: "Enum_generator_UserStatus",
		Values: graphql.EnumValueConfigMap{
			"UNKNOWN_USER_STATE": &graphql.EnumValueConfig{
				Value: UserStatus_value["UNKNOWN_USER_STATE"],
			},
			"REGISTERED": &graphql.EnumValueConfig{
				Value: UserStatus_value["REGISTERED"],
			},
			"UNREGISTERED": &graphql.EnumValueConfig{
				Value: UserStatus_value["UNREGISTERED"],
			},
		},
	},
)
`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			importPath := p.FilesByPath["test.proto"].GoImportPath
			g := p.NewGeneratedFile("test.pb.graphql.go", importPath)
			v := NewVisitor(f, g, importPath.String())
			v.VisitEnum(root, c.enum)
			res, err := v.Content()
			if err != nil {
				t.Fatalf("failed to parse golang enum: %s", err.Error())
			}
			if res := cmp.Diff(c.wantString, string(res)); res != "" {
				t.Error(res)
			}
		})
	}
}

func TestVisitMessage(t *testing.T) {
	f, ok := p.FilesByPath["test.proto"]
	if !ok {
		panic("failed to read FileDescriptor")
	}
	messageCode := `var Object_generator_Test *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Object_generator_Test",
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		return true
	},
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Name: "name",
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var res interface{}
				if pdata, ok := p.Source.(*Test); ok {
					res = pdata.Name
				} else if data, ok := p.Source.(Test); ok {
					res = data.Name
				}
				return res, nil
			},
		},
	},
})
`
	// TODO(lim): More field types
	// {{- range $union, $unionFields := $obj.Unions }}
	// "{{ $union }}": &graphql.Field{
	// 	Name: "{{ $union }}",
	// 	Type: GraphQL_{{ GetGraphQLTypeName $fqmn }}_{{ $union }}Union,
	// },
	// {{- end }}
	// {{- range $mapField := $obj.Maps }}
	// "{{ $mapField.GetName }}": &graphql.Field{
	// 	Name: "{{ $mapField.GetName }}",
	// 	Type: graphql.NewList(GraphQL_{{ GetGraphQLTypeName $mapField.GetTypeName }}Map),
	// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
	// 		var tmp interface{}
	// 		if pdata, ok := p.Source.(*{{ GetLanguageType $fqmn }}); ok {
	// 			tmp = pdata.{{ GetProtobufFieldName $mapField.GetName }}
	// 		} else if data, ok := p.Source.({{ GetLanguageType $fqmn }}); ok {
	// 			tmp = data.{{ GetProtobufFieldName $mapField.GetName }}
	// 		}
	// 		kv, ok := tmp.(map[{{ (index $obj.MapTypes $mapField.GetTypeName).KeyType }}]{{ (index $obj.MapTypes $mapField.GetTypeName).ValueType }})
	// 		if !ok {
	// 			return nil, graphqlpb.ErrBadValue
	// 		}
	// 		res := make([]interface{}, 0)
	// 		for key, value := range kv {
	// 		    rec := map[string]interface{}{
	// 			"key": key,
	// 			"value": value,
	// 		    }
	// 		    res = append(res, rec)
	// 		}
	// 		return res, nil
	// 	},
	// },
	cases := []struct {
		name       string
		message    *protogen.Message
		wantString string
	}{
		{
			name:       "Test",
			message:    f.Messages[0],
			wantString: messageCode,
		},
		{
			name:       "TestRepeated",
			message:    f.Messages[1],
			wantString: messageCode,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := p.NewGeneratedFile("test.graphql.pb.go", p.FilesByPath["test.proto"].GoImportPath)
			v := NewVisitor(f, g, p.Files[0].GoImportPath.String())
			v.VisitMessage(root, c.message, GQLTypeObject)
			res, err := v.Content()
			if err != nil {
				t.Fatalf("failed to generate file: %s", err.Error())
			}
			fmt.Println(string(res))
		})
	}
}

func TestVisitService(t *testing.T) {
	f, ok := p.FilesByPath["test.proto"]
	if !ok {
		panic("failed to read FileDescriptor")
	}
	wantSvcString := `package descriptorpb

import (
	json "encoding/json"
	graphql "github.com/ncrypthic/graphql-grpc-edge/graphql"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	wrappers "google.golang.org/protobuf/types/known/wrappers"
)

var Input_TestScalar *graphql.InputObject = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "Input_TestScalar",
		IsTypeOf: func(g graphql.IsTypeOfParams) bool {
			return true
		},
		Fields: graphql.InputObjectConfigFieldMap{
			"field1": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field1
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field1
					}
					return res, nil
				},
			},
			"field2": &graphql.Field{
				Type: graphql.Bool,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field2
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field2
					}
					return res, nil
				},
			},
			"field3": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field3
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field3
					}
					return res, nil
				},
			},
			"field4": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field4
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field4
					}
					return res, nil
				},
			},
			"field5": &graphql.Field{
				Type: graphql.Float,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field5
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field5
					}
					return res, nil
				},
			},
			"field6": &graphql.Field{
				Type: graphql.Float,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field6
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field6
					}
					return res, nil
				},
			},
			"field7": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field7
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field7
					}
					return res, nil
				},
			},
			"field8": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field8
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field8
					}
					return res, nil
				},
			},
			"field9": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field9
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field9
					}
					return res, nil
				},
			},
			"field10": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field10
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field10
					}
					return res, nil
				},
			},
			"field11": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field11
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field11
					}
					return res, nil
				},
			},
			"field12": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field12
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field12
					}
					return res, nil
				},
			},
			"field13": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field13
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field13
					}
					return res, nil
				},
			},
			"field14": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestScalar); ok {
						res = pdata.Field14
					} else if data, ok := p.Source.(TestScalar); ok {
						res = data.Field14
					}
					return res, nil
				},
			},
		},
	},
)
var Input_Test_TestDetail *graphql.InputObject = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "Input_Test_TestDetail",
		IsTypeOf: func(g graphql.IsTypeOfParams) bool {
			return true
		},
		Fields: graphql.InputObjectConfigFieldMap{
			"name": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test_TestDetail); ok {
						res = pdata.Name
					} else if data, ok := p.Source.(Test_TestDetail); ok {
						res = data.Name
					}
					return res, nil
				},
			},
			"photo": &graphql.Field{
				Type: graphql.graphql.Scalar_bytes,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test_TestDetail); ok {
						res = pdata.Photo
					} else if data, ok := p.Source.(Test_TestDetail); ok {
						res = data.Photo
					}
					return res, nil
				},
			},
		},
	},
)
var Input_Test_AttributesEntry *graphql.InputObject = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "Input_Test_AttributesEntry",
		IsTypeOf: func(g graphql.IsTypeOfParams) bool {
			return true
		},
		Fields: graphql.InputObjectConfigFieldMap{
			"key": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test_AttributesEntry); ok {
						res = pdata.Key
					} else if data, ok := p.Source.(Test_AttributesEntry); ok {
						res = data.Key
					}
					return res, nil
				},
			},
			"value": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test_AttributesEntry); ok {
						res = pdata.Value
					} else if data, ok := p.Source.(Test_AttributesEntry); ok {
						res = data.Value
					}
					return res, nil
				},
			},
		},
	},
)
var Input_Test *graphql.InputObject = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "Input_Test",
		IsTypeOf: func(g graphql.IsTypeOfParams) bool {
			return true
		},
		Fields: graphql.InputObjectConfigFieldMap{
			"name": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.Name
					} else if data, ok := p.Source.(Test); ok {
						res = data.Name
					}
					return res, nil
				},
			},
			"maybeString": &graphql.Field{
				Type: graphql.Input_wrappers_StringValue,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.MaybeString
					} else if data, ok := p.Source.(Test); ok {
						res = data.MaybeString
					}
					return res, nil
				},
			},
			"detail": &graphql.Field{
				Type: Input_Test_TestDetail,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.Detail
					} else if data, ok := p.Source.(Test); ok {
						res = data.Detail
					}
					return res, nil
				},
			},
			"country": &graphql.Field{
				Type: Enum_Test_Country,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.Country
					} else if data, ok := p.Source.(Test); ok {
						res = data.Country
					}
					return res, nil
				},
			},
			"state": &graphql.Field{
				Type: Enum_Test_State,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.State
					} else if data, ok := p.Source.(Test); ok {
						res = data.State
					}
					return res, nil
				},
			},
			"createdAt": &graphql.Field{
				Type: graphql.Scalar_timestamppb_Timestamp,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.CreatedAt
					} else if data, ok := p.Source.(Test); ok {
						res = data.CreatedAt
					}
					return res, nil
				},
			},
			"lastSession": &graphql.Field{
				Type: graphql.Scalar_durationpb_Duration,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.LastSession
					} else if data, ok := p.Source.(Test); ok {
						res = data.LastSession
					}
					return res, nil
				},
			},
			"attributes": &graphql.Field{
				Type: graphql.NewList(Input_Test_AttributesEntry),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*Test); ok {
						res = pdata.Attributes
					} else if data, ok := p.Source.(Test); ok {
						res = data.Attributes
					}
					return res, nil
				},
			},
		},
	},
)
var Object_TestRepeated *graphql.Object = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Object_TestRepeated",
		IsTypeOf: func(g graphql.IsTypeOfParams) bool {
			return true
		},
		Fields: graphql.Fields{
			"tags": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestRepeated); ok {
						res = pdata.Tags
					} else if data, ok := p.Source.(TestRepeated); ok {
						res = data.Tags
					}
					return res, nil
				},
			},
			"failedAttempts": &graphql.Field{
				Type: graphql.NewList(graphql.Scalar_timestamppb_Timestamp),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestRepeated); ok {
						res = pdata.FailedAttempts
					} else if data, ok := p.Source.(TestRepeated); ok {
						res = data.FailedAttempts
					}
					return res, nil
				},
			},
			"history": &graphql.Field{
				Type: graphql.NewList(Enum_TestRepeated_History),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestRepeated); ok {
						res = pdata.History
					} else if data, ok := p.Source.(TestRepeated); ok {
						res = data.History
					}
					return res, nil
				},
			},
		},
	},
)
var Input_TestRepeated *graphql.InputObject = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "Input_TestRepeated",
		IsTypeOf: func(g graphql.IsTypeOfParams) bool {
			return true
		},
		Fields: graphql.InputObjectConfigFieldMap{
			"tags": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestRepeated); ok {
						res = pdata.Tags
					} else if data, ok := p.Source.(TestRepeated); ok {
						res = data.Tags
					}
					return res, nil
				},
			},
			"failedAttempts": &graphql.Field{
				Type: graphql.NewList(graphql.Scalar_timestamppb_Timestamp),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestRepeated); ok {
						res = pdata.FailedAttempts
					} else if data, ok := p.Source.(TestRepeated); ok {
						res = data.FailedAttempts
					}
					return res, nil
				},
			},
			"history": &graphql.Field{
				Type: graphql.NewList(Enum_TestRepeated_History),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res interface{}
					if pdata, ok := p.Source.(*TestRepeated); ok {
						res = pdata.History
					} else if data, ok := p.Source.(TestRepeated); ok {
						res = data.History
					}
					return res, nil
				},
			},
		},
	},
)

func RegisterHelloTestServiceQueries(sc HelloTestServiceClient) error {
	graphql.RegisterQuery("hello", &graphql.Field{
		Name: "hello",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: Input_Test,
			},
		},
		Type: Object_Test,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var req Test
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(rawJson, &req)
			if err != nil {
				return nil, err
			}
			return sc.HelloQuery(ctx, &req)
		},
	})
}

func RegisterHelloTestServiceMutations(sc HelloTestServiceClient) error {
	graphql.RegisterMutation("mutateHello", &graphql.Field{
		Name: "mutateHello",
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: Input_Test,
			},
		},
		Type: graphql.Scalar_emptypb_Empty,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var req Test
			rawJson, err := json.Marshal(p.Args["input"])
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(rawJson, &req)
			if err != nil {
				return nil, err
			}
			return sc.HelloMutation(ctx, &req)
		},
	})
}
`
	cases := []struct {
		name       string
		message    *protogen.Service
		wantString string
	}{
		// {
		// 	name:       "no grapqhl methods",
		// 	message:    f.Services[0],
		// 	wantString: wantEmptyString,
		// },
		{
			name:       "positive",
			message:    f.Services[1],
			wantString: wantSvcString,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := p.NewGeneratedFile("test.graphql.pb.go", p.FilesByPath["test.proto"].GoImportPath)
			v := NewVisitor(f, g, p.Files[0].GoImportPath.String())
			for _, m := range f.Messages {
				v.VisitMessage(root, m, GQLTypeObject)
				v.VisitMessage(root, m, GQLTypeInput)
			}
			v.VisitService(root, c.message)
			res, err := v.Content()
			if err != nil {
				t.Fatalf("failed to generate file: %s", err.Error())
			}
			fmt.Println(string(res))
			// if res := cmp.Diff(c.wantString, string(res)); res != "" {
			// 	t.Error(res)
			// }
		})
	}
}
