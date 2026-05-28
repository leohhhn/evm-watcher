// Package memory provides an in-memory Storage implementation intended for use
// in tests. It records every SaveTransfer call so assertions can inspect what
// the watcher emitted without touching a real database.
package memory

import (
	"context"
	"sync"

	"github.com/leohhhn/evm-watcher/internal/storage"
)

// Store is a thread-safe in-memory Storage spy.
type Store struct {
	mu        sync.Mutex
	Transfers []storage.Transfer
}

func New() *Store {
	return &Store{}
}

func (s *Store) SaveTransfer(_ context.Context, t storage.Transfer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Transfers = append(s.Transfers, t)
	return nil
}

func (s *Store) Close() error { return nil }

// Len returns the number of transfers recorded so far.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Transfers)
}

// Reset clears all recorded transfers.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Transfers = nil
}
