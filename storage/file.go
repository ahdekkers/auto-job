package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ahdekkers/auto-job.git/jobs"
)

// fileStore is the in-file data structure — a map of job ID to Job.
type fileStore map[string]jobs.Job

// FileJobStorage is an implementation of interfaces.JobStorage that persists
// all jobs as a single JSON file on disk.
type FileJobStorage struct {
	mu       sync.RWMutex
	filePath string
}

// NewFileJobStorage creates a new FileJobStorage backed by the file at filePath.
// If the file does not exist it will be created on the first write.
func NewFileJobStorage(filePath string) *FileJobStorage {
	return &FileJobStorage{filePath: filePath}
}

// load reads and deserialises the JSON file from disk.
// Must be called with at least a read lock held.
func (s *FileJobStorage) load() (fileStore, error) {
	data, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return make(fileStore), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading storage file: %w", err)
	}

	var store fileStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parsing storage file: %w", err)
	}
	return store, nil
}

// save serialises the store and writes it atomically to disk.
// Must be called with the write lock held.
func (s *FileJobStorage) save(store fileStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising store: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing storage file: %w", err)
	}
	return nil
}

// Store persists a job. The job must already have its ID set.
func (s *FileJobStorage) Store(job jobs.Job) error {
	if job.ID == "" {
		return fmt.Errorf("job ID must be set before storing")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	store, err := s.load()
	if err != nil {
		return err
	}

	store[job.ID] = job
	return s.save(store)
}

// Retrieve returns the job with the given ID, or an error if not found.
func (s *FileJobStorage) Retrieve(jobID string) (jobs.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	store, err := s.load()
	if err != nil {
		return jobs.Job{}, err
	}

	job, ok := store[jobID]
	if !ok {
		return jobs.Job{}, fmt.Errorf("job %q not found", jobID)
	}
	return job, nil
}

// RetrieveAll returns the jobs corresponding to the given IDs.
// IDs that are not found are silently skipped.
func (s *FileJobStorage) RetrieveAll(jobIDs []string) ([]jobs.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	store, err := s.load()
	if err != nil {
		return nil, err
	}

	jobs := make([]jobs.Job, 0, len(jobIDs))
	for _, id := range jobIDs {
		if job, ok := store[id]; ok {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

// Delete removes the job with the given ID.
func (s *FileJobStorage) Delete(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	store, err := s.load()
	if err != nil {
		return err
	}

	delete(store, jobID)
	return s.save(store)
}

// DeleteAll removes all jobs with the given IDs.
func (s *FileJobStorage) DeleteAll(jobIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	store, err := s.load()
	if err != nil {
		return err
	}

	for _, id := range jobIDs {
		delete(store, id)
	}
	return s.save(store)
}
