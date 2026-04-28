package jobs

// JobsAPI defines the contract for retrieving job listings from an external source.
type JobsAPI interface {
	// GetJobs fetches a list of job listings. The returned jobs will not yet
	// have an ID assigned — that is handled by the storage layer.
	GetJobs() ([]Job, error)
}
