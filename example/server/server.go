package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/ncrypthic/graphql-grpc-edge/example/sample"
)

type HelloServer struct{}

func (h *HelloServer) Greeting(ctx context.Context, req *sample.HelloRequest) (*sample.HelloResponse, error) {
	return &sample.HelloResponse{
		Data: &sample.Hello{
			Name: req.Name,
		},
	}, nil
}

func (h *HelloServer) SetGreeting(ctx context.Context, req *sample.Hello) (*sample.HelloResponse, error) {
	return &sample.HelloResponse{
		Data: &sample.Hello{
			Name:      req.Name,
			Type:      req.Type,
			Messages:  req.Messages,
			CreatedAt: req.CreatedAt,
		},
	}, nil
}

func (h *HelloServer) HelloQuery(ctx context.Context, req *sample.Test) (*sample.Test, error) {
	return req, nil
}

func (h *HelloServer) MetGreeting(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return req, nil
}

func (h *HelloServer) SetDuration(ctx context.Context, req *duration.Duration) (*duration.Duration, error) {
	return req, nil
}

func (h *HelloServer) SetTimestamp(ctx context.Context, req *timestamp.Timestamp) (*timestamp.Timestamp, error) {
	return req, nil
}

func (h *HelloServer) SetInt32Value(ctx context.Context, req *wrappers.Int32Value) (*wrappers.Int32Value, error) {
	return req, nil
}

func (h *HelloServer) SetBoolValue(ctx context.Context, req *wrappers.BoolValue) (*wrappers.BoolValue, error) {
	return req, nil
}

func (h *HelloServer) SetStringValue(ctx context.Context, req *wrappers.StringValue) (*wrappers.StringValue, error) {
	return req, nil
}
