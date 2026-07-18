package repository

import (
	"encoding/json"
	"os"
	"sync"
)

// FileStateRepository manages persistence of deleted services checkmarks to a local JSON file.
type FileStateRepository struct {
	filePath string
	mu       sync.RWMutex
	deleted  map[string]bool
}

// NewFileStateRepository creates a new FileStateRepository.
func NewFileStateRepository(filePath string) *FileStateRepository {
	repo := &FileStateRepository{
		filePath: filePath,
		deleted:  make(map[string]bool),
	}
	repo.load()
	return repo
}

// load loads the saved deleted domains from disk.
func (r *FileStateRepository) load() {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, err := os.Open(r.filePath)
	if err != nil {
		return // File might not exist yet, which is fine
	}
	defer f.Close()

	var list []string
	if err := json.NewDecoder(f).Decode(&list); err == nil {
		for _, domain := range list {
			r.deleted[domain] = true
		}
	}
}

// Save persists the deleted status of a domain.
func (r *FileStateRepository) Save(domain string, deleted bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if deleted {
		r.deleted[domain] = true
	} else {
		delete(r.deleted, domain)
	}

	var list []string
	for dom := range r.deleted {
		list = append(list, dom)
	}

	f, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(list)
}

// IsDeleted checks if a domain is marked as deleted.
func (r *FileStateRepository) IsDeleted(domain string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.deleted[domain]
}

// IsDeletedList returns the list of all domains marked as deleted.
func (r *FileStateRepository) IsDeletedList() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []string
	for dom, val := range r.deleted {
		if val {
			list = append(list, dom)
		}
	}
	return list
}
