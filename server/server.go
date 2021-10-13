package server

import (
	"context"
	"sync"

	pb "github.com/ksonny4/link-tracking/proto"
)

// Backend implements the protobuf interface
type Backend struct {
	mu *sync.RWMutex
}

// New initializes a new Backend struct.
func New() *Backend {
	init_main() // TODO move and close DB!!!!
	return &Backend{
		mu: &sync.RWMutex{},
	}
}

func (b *Backend) GetUrl(ctx context.Context, input *pb.URLGenerateRequest) (*pb.Url, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return GetUrl(input)
}
