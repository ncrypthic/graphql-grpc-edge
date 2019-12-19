package server

import (
	"context"

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
			Name:     req.Name,
			Type:     req.Type,
			Messages: req.Messages,
		},
	}, nil
}
func (h *HelloServer) HelloQuery(ctx context.Context, req *sample.Test) (*sample.Test, error) {
	return req, nil
}
