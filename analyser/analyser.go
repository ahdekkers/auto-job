package analyser

import (
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/users"
)

// SuitabilityRating is a value between 0.0 and 1.0 representing how well a job
// matches a user's profile, where 1.0 is a perfect match and 0.0 is no match.
type SuitabilityRating float64

// JobAnalyser defines the contract for analysing how suitable a job is for a user.
type JobAnalyser interface {
	// GetSuitabilityRating returns a score between 0.0 and 1.0 indicating how
	// well the given job matches the given user's profile and preferences.
	GetSuitabilityRating(job jobs.Job, user users.User) (SuitabilityRating, error)

	// GetAllSuitabilityRatings evaluates a list of jobs against a single user and
	// returns a map of job ID to suitability rating for each job.
	GetAllSuitabilityRatings(jobs []jobs.Job, user users.User) (map[string]SuitabilityRating, error)
}
