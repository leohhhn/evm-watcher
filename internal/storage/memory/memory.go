// Package memory provides an in-memory Storage implementation intended for use
// in tests. It records every SaveTransfer call so assertions can inspect what
// the watcher emitted without touching a real database.
package memory

import (
	"context"
	"sync"

	"github.com/leohhhn/tokentail/internal/storage"
)

// Store is a thread-safe in-memory Storage spy.
type Store struct {
	mu        sync.Mutex
	transfers []storage.Transfer
}

// New returns an empty, ready-to-use Store.
func New() *Store {
	return &Store{}
}

// SaveTransfer appends t to the in-memory list for later inspection by tests.
func (s *Store) SaveTransfer(_ context.Context, t storage.Transfer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transfers = append(s.transfers, t)
	return nil
}

// Close is a no-op; the in-memory store requires no teardown.
func (s *Store) Close() error { return nil }

// Len returns the number of transfers recorded so far.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.transfers)
}

// Get returns the transfer at index i under the lock.
func (s *Store) Get(i int) storage.Transfer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transfers[i]
}

// Reset clears all recorded transfers.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transfers = nil
}
