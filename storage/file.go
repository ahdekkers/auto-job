package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/users"
)

// FileStorage implements JobStorage, UserStorage, and AnalysisStorage
// backed by a single JSON file.
type FileStorage struct {
	FilePath string
	mu       sync.Mutex
}

// fileData is the full persisted structure.
type fileData struct {
	Jobs     map[string]jobs.Job            `json:"jobs"`
	Users    map[string]users.User          `json:"users"`
	Analysis map[string][]analyser.Analysis `json:"analysis"` // keyed by userID
}

// NewFileStorage creates a new file storage instance.
func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		FilePath: path,
	}
}

// --------------------
// internal helpers
// --------------------

func (f *FileStorage) load() (*fileData, error) {
	// If file doesn't exist, return empty structure
	if _, err := os.Stat(f.FilePath); os.IsNotExist(err) {
		return &fileData{
			Jobs:     make(map[string]jobs.Job),
			Users:    make(map[string]users.User),
			Analysis: make(map[string][]analyser.Analysis),
		}, nil
	}

	bytes, err := os.ReadFile(f.FilePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var data fileData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("unmarshal file: %w", err)
	}

	// Ensure maps are initialized
	if data.Jobs == nil {
		data.Jobs = make(map[string]jobs.Job)
	}
	if data.Users == nil {
		data.Users = make(map[string]users.User)
	}
	if data.Analysis == nil {
		data.Analysis = make(map[string][]analyser.Analysis)
	}

	return &data, nil
}

func (f *FileStorage) save(data *fileData) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal file: %w", err)
	}

	if err := os.WriteFile(f.FilePath, bytes, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// --------------------
// JobStorage
// --------------------

func (f *FileStorage) StoreJobs(jobList []jobs.Job) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	for _, job := range jobList {
		data.Jobs[job.ID] = job
	}

	return f.save(data)
}

// NOTE: interface is a bit odd (takes []jobs.Job as input)
// We'll interpret this as a filter by IDs if provided.
func (f *FileStorage) RetrieveJobs(filter []jobs.Job) ([]jobs.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return nil, err
	}

	// If no filter, return all
	if len(filter) == 0 {
		result := make([]jobs.Job, 0, len(data.Jobs))
		for _, job := range data.Jobs {
			result = append(result, job)
		}
		return result, nil
	}

	// Otherwise filter by IDs
	result := []jobs.Job{}
	for _, j := range filter {
		if stored, ok := data.Jobs[j.ID]; ok {
			result = append(result, stored)
		}
	}

	return result, nil
}

func (f *FileStorage) DeleteJobs(jobIDs []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	for _, id := range jobIDs {
		delete(data.Jobs, id)
	}

	return f.save(data)
}

// --------------------
// UserStorage
// --------------------

func (f *FileStorage) StoreUsers(user users.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	// Using Email as unique ID (you can swap this if needed)
	userID := user.Email
	if userID == "" {
		return fmt.Errorf("user email cannot be empty")
	}

	data.Users[userID] = user

	return f.save(data)
}

func (f *FileStorage) RetrieveUsers(userID string) (users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return users.User{}, err
	}

	user, ok := data.Users[userID]
	if !ok {
		return users.User{}, fmt.Errorf("user not found: %s", userID)
	}

	return user, nil
}

func (f *FileStorage) DeleteUsers(userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	delete(data.Users, userID)
	delete(data.Analysis, userID) // also clean up related analysis

	return f.save(data)
}

// --------------------
// AnalysisStorage
// --------------------

func (f *FileStorage) StoreAnalysis(a analyser.Analysis) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	if a.UserID == "" {
		return fmt.Errorf("analysis user_id cannot be empty")
	}

	data.Analysis[a.UserID] = append(data.Analysis[a.UserID], a)

	return f.save(data)
}

func (f *FileStorage) RetrieveAnalysis(userID string) ([]analyser.Analysis, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return nil, err
	}

	return data.Analysis[userID], nil
}

func (f *FileStorage) DeleteAnalysis(userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	delete(data.Analysis, userID)

	return f.save(data)
}
