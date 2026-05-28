package storage

import (
	"context"
	"time"
)

// Transfer is the canonical record written to any storage backend.
type Transfer struct {
	Block     uint64
	TxHash    string
	LogIndex  uint
	Token     string
	From      string
	To        string
	Amount    float64
	CreatedAt time.Time
}

// Storage is the interface any persistence backend must satisfy.
// Implementations must be safe for concurrent use.
type Storage interface {
	SaveTransfer(ctx context.Context, t Transfer) error
	Close() error
}
