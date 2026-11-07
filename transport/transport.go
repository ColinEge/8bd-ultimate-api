package transport

import "context"

type Transport interface {
	Write(ctx context.Context, data []byte) error
	Read(ctx context.Context) ([]byte, error)
	Close() error
}
