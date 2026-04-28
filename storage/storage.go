package storage

import (
	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/users"
)

// JobStorage defines the contract for persisting and retrieving job listings.
type JobStorage interface {
	StoreJobs(jobs []jobs.Job) error
	RetrieveJobs(jobs []jobs.Job) ([]jobs.Job, error)
	DeleteJobs(jobIDs []string) error
}

type UserStorage interface {
	StoreUsers(user users.User) error
	RetrieveUsers(userID string) (users.User, error)
	DeleteUsers(userID string) error
}

type AnalysisStorage interface {
	StoreAnalysis(analysis analyser.Analysis) error
	RetrieveAnalysis(userID string) ([]analyser.Analysis, error)
	DeleteAnalysis(userID string) error
}

type SearchStorage interface {
	StoreSearch(searchName string, params jobs.JobSearchParams) error
	RetrieveSearch(searchName string) (jobs.JobSearchParams, error)
	DeleteSearch(searchName string) error
}
