package repository

import (
	"bufio"
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
	file    *os.File
	writer  *bufio.Writer
	encoder *json.Encoder
	mu      sync.Mutex
}

// NewJSONLinesRepository creates a new JSONLinesRepository writing to targetPath.
func NewJSONLinesRepository(targetPath string) (*JSONLinesRepository, error) {
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	writer := bufio.NewWriterSize(file, 64*1024) // 64KB buffer
	return &JSONLinesRepository{
		file:    file,
		writer:  writer,
		encoder: json.NewEncoder(writer),
	}, nil
}

// Save implements outbound.EmailRepository by writing email as a JSON line.
func (r *JSONLinesRepository) Save(ctx context.Context, email *model.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.encoder.Encode(email)
}

// Close closes the underlying JSON lines file.
func (r *JSONLinesRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var flushErr error
	if r.writer != nil {
		flushErr = r.writer.Flush()
	}

	var closeErr error
	if r.file != nil {
		closeErr = r.file.Close()
	}

	if flushErr != nil {
		return flushErr
	}
	return closeErr
}
