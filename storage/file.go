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

// FileStorage implements JobStorage, UserStorage, AnalysisStorage and SearchStorage
// backed by a single JSON file.
type FileStorage struct {
	FilePath string
	mu       sync.Mutex
}

type fileData struct {
	Jobs     map[string]jobs.Job             `json:"jobs"`
	Users    map[string]users.User           `json:"users"`
	Analysis map[string][]analyser.Analysis  `json:"analysis"`
	Searches map[string]jobs.JobSearchParams `json:"searches"`
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{FilePath: path}
}

func (f *FileStorage) load() (*fileData, error) {
	if _, err := os.Stat(f.FilePath); os.IsNotExist(err) {
		return &fileData{
			Jobs:     make(map[string]jobs.Job),
			Users:    make(map[string]users.User),
			Analysis: make(map[string][]analyser.Analysis),
			Searches: make(map[string]jobs.JobSearchParams),
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

	if data.Jobs == nil {
		data.Jobs = make(map[string]jobs.Job)
	}
	if data.Users == nil {
		data.Users = make(map[string]users.User)
	}
	if data.Analysis == nil {
		data.Analysis = make(map[string][]analyser.Analysis)
	}
	if data.Searches == nil {
		data.Searches = make(map[string]jobs.JobSearchParams)
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

// JobStorage

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

func (f *FileStorage) RetrieveJobs(filter []jobs.Job) ([]jobs.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return nil, err
	}

	if len(filter) == 0 {
		result := make([]jobs.Job, 0, len(data.Jobs))
		for _, job := range data.Jobs {
			result = append(result, job)
		}
		return result, nil
	}

	result := make([]jobs.Job, 0, len(filter))
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

// UserStorage

func (f *FileStorage) StoreUsers(user users.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

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
	delete(data.Analysis, userID)

	return f.save(data)
}

// AnalysisStorage

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

// SearchStorage

func (f *FileStorage) StoreSearch(searchName string, params jobs.JobSearchParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if searchName == "" {
		return fmt.Errorf("search name cannot be empty")
	}

	data, err := f.load()
	if err != nil {
		return err
	}

	if _, exists := data.Searches[searchName]; exists {
		return fmt.Errorf("search already exists: %s", searchName)
	}

	data.Searches[searchName] = params
	return f.save(data)
}

func (f *FileStorage) RetrieveSearch(searchName string) (jobs.JobSearchParams, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return jobs.JobSearchParams{}, err
	}

	params, ok := data.Searches[searchName]
	if !ok {
		return jobs.JobSearchParams{}, fmt.Errorf("search not found: %s", searchName)
	}

	return params, nil
}

func (f *FileStorage) DeleteSearch(searchName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := f.load()
	if err != nil {
		return err
	}

	delete(data.Searches, searchName)
	return f.save(data)
}
