package storage

import (
	"context"
	"errors"
)

type Storage interface {
	GetClient(ctx context.Context) (*Client, error)
	CreateClient(ctx context.Context, login, password string) (*Client, error)
}

type Client struct {
	Login    string
	password string
}

var ErrClientNotFound = errors.New("client not found")
