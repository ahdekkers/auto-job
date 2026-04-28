package controller

import (
	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/users"
)

// Controller defines the orchestration layer of the application.
type Controller interface {
	// CreateUser constructs a user from input fields and persists it.
	CreateUser(
		name string,
		email string,
		yearsOfExperience int,
		currentTitle string,
		skills []string,
		preferredWorkModes []users.WorkMode,
		preferredLocations []string,
		salaryExpectation users.SalaryExpectation,
		cvSummary string,
		additionalNotes []string,
	) (users.User, error)

	// ExecuteJobSearch:
	// - fetches jobs from the jobs API
	// - retrieves the user from storage
	// - runs suitability analysis
	// - stores jobs and analysis results
	// - returns the jobs and rating map
	ExecuteJobSearch(
		userID string,
		params jobs.JobSearchParams,
	) ([]jobs.Job, map[string]analyser.SuitabilityRating, error)
}
