package controller

import (
	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/users"
)

type Controller interface {
	CreateUser(
		name string,
		email string,
		yearsOfExperience int,
		currentTitle string,
		preferredWorkModes []users.WorkMode,
		preferredLocations []string,
		salaryExpectation users.SalaryExpectation,
		cv string,
		additionalNotes []string,
	) (users.User, error)

	SaveSearch(searchName string, params jobs.JobSearchParams) error

	ExecuteJobSearch(
		userID string,
		searchName string,
	) ([]jobs.Job, map[string]analyser.SuitabilityRating, error)
}
