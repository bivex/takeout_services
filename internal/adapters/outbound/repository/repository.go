package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"takeout_services/internal/domain/model"
)

// InMemoryRepository stores all parsed emails in RAM.
type InMemoryRepository struct {
	mu     sync.RWMutex
	emails []*model.Email
}

// NewInMemoryRepository creates an instance of InMemoryRepository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		emails: make([]*model.Email, 0),
	}
}

// Save implements outbound.EmailRepository.
func (r *InMemoryRepository) Save(ctx context.Context, email *model.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emails = append(r.emails, email)
	return nil
}

// Emails returns the emails stored in memory.
func (r *InMemoryRepository) Emails() []*model.Email {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.emails
}

// JSONLinesRepository streams parsed emails directly to a JSON Lines (.jsonl) file.
type JSONLinesRepository struct {
	file *os.File
	mu   sync.Mutex
}

// NewJSONLinesRepository creates a new JSONLinesRepository writing to targetPath.
func NewJSONLinesRepository(targetPath string) (*JSONLinesRepository, error) {
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONLinesRepository{file: file}, nil
}

// Save implements outbound.EmailRepository by writing email as a JSON line.
func (r *JSONLinesRepository) Save(ctx context.Context, email *model.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bytesVal, err := json.Marshal(email)
	if err != nil {
		return err
	}

	if _, err := r.file.Write(append(bytesVal, '\n')); err != nil {
		return err
	}
	return nil
}

// Close closes the underlying JSON lines file.
func (r *JSONLinesRepository) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}
