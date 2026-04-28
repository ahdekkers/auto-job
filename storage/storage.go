package storage

import "github.com/ahdekkers/auto-job.git/jobs"

// JobStorage defines the contract for persisting and retrieving job listings.
type JobStorage interface {
	// Store persists a job. The job must already have its ID set.
	Store(job jobs.Job) error

	// Retrieve returns the job with the given ID, or an error if not found.
	Retrieve(jobID string) (jobs.Job, error)

	// RetrieveAll returns the jobs corresponding to the given IDs.
	// If a given ID is not found it is silently skipped.
	RetrieveAll(jobIDs []string) ([]jobs.Job, error)

	// Delete removes the job with the given ID.
	Delete(jobID string) error

	// DeleteAll removes all jobs with the given IDs.
	DeleteAll(jobIDs []string) error
}
