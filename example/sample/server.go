package sample

import (
	"context"
)

type HelloServer struct{}

func (h *HelloServer) Greeting(ctx context.Context, req *Hello) (*HelloResponse, error) {
	return &HelloResponse{
		Data: &Hello{
			Name:     req.Name,
			Type:     req.Type,
			Messages: req.Messages,
		},
	}, nil
}

func (h *HelloServer) SetGreeting(ctx context.Context, req *Hello) (*HelloResponse, error) {
	return &HelloResponse{
		Data: &Hello{
			Name:     req.Name,
			Type:     req.Type,
			Messages: req.Messages,
		},
	}, nil
}
