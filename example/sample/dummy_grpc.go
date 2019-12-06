package sample

import (
	"context"
)

type Hello struct {
	Name string
}

type HelloResponse struct{}

type HelloServiceClient interface {
	Greeting(context.Context, *Hello) (*HelloResponse, error)
	SetGreeting(context.Context, *Hello) (*HelloResponse, error)
}

type hellosvc struct{}

func (svc *hellosvc) Greeting(ctx context.Context, req *Hello) (*HelloResponse, error) {
	return nil, nil
}

func (svc *hellosvc) SetGreeting(ctx context.Context, req *Hello) (*HelloResponse, error) {
	return nil, nil
}

func NewHelloServiceClient() HelloServiceClient {
	return &hellosvc{}
}
