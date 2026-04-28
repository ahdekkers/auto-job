package controller

import (
	"fmt"
	"strings"

	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/storage"
	"github.com/ahdekkers/auto-job.git/users"
)

// AppController is the concrete implementation of Controller.
type AppController struct {
	JobAPI          *jobs.FantasticJobsAPI
	Analyser        analyser.JobAnalyser
	JobStorage      storage.JobStorage
	UserStorage     storage.UserStorage
	AnalysisStorage storage.AnalysisStorage
}

// NewAppController constructs a new controller.
func NewAppController(
	jobAPI *jobs.FantasticJobsAPI,
	analyser analyser.JobAnalyser,
	jobStorage storage.JobStorage,
	userStorage storage.UserStorage,
	analysisStorage storage.AnalysisStorage,
) *AppController {
	return &AppController{
		JobAPI:          jobAPI,
		Analyser:        analyser,
		JobStorage:      jobStorage,
		UserStorage:     userStorage,
		AnalysisStorage: analysisStorage,
	}
}

// CreateUser constructs a user from input fields and persists it.
func (c *AppController) CreateUser(
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
) (users.User, error) {
	user := users.User{
		Name:               strings.TrimSpace(name),
		Email:              strings.TrimSpace(email),
		YearsOfExperience:  yearsOfExperience,
		CurrentTitle:       strings.TrimSpace(currentTitle),
		Skills:             append([]string(nil), skills...),
		PreferredWorkModes: append([]users.WorkMode(nil), preferredWorkModes...),
		PreferredLocations: append([]string(nil), preferredLocations...),
		SalaryExpectation:  salaryExpectation,
		CVSummary:          strings.TrimSpace(cvSummary),
		AdditionalNotes:    append([]string(nil), additionalNotes...),
	}

	if user.Email == "" {
		return users.User{}, fmt.Errorf("email is required")
	}
	if c.UserStorage == nil {
		return users.User{}, fmt.Errorf("user storage is nil")
	}

	if err := c.UserStorage.StoreUsers(user); err != nil {
		return users.User{}, fmt.Errorf("store user: %w", err)
	}

	return user, nil
}

// ExecuteJobSearch fetches jobs, loads the user, scores the jobs, and stores the results.
func (c *AppController) ExecuteJobSearch(
	userID string,
	params jobs.JobSearchParams,
) ([]jobs.Job, map[string]analyser.SuitabilityRating, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil, fmt.Errorf("userID is required")
	}
	if c.JobAPI == nil {
		return nil, nil, fmt.Errorf("job api client is nil")
	}
	if c.Analyser == nil {
		return nil, nil, fmt.Errorf("analyser is nil")
	}
	if c.JobStorage == nil || c.UserStorage == nil || c.AnalysisStorage == nil {
		return nil, nil, fmt.Errorf("storage dependency is nil")
	}

	user, err := c.UserStorage.RetrieveUsers(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve user: %w", err)
	}

	foundJobs, err := c.JobAPI.GetJobs(params)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch jobs: %w", err)
	}

	if err := c.JobStorage.StoreJobs(foundJobs); err != nil {
		return nil, nil, fmt.Errorf("store jobs: %w", err)
	}

	ratings, err := c.Analyser.GetAllSuitabilityRatings(foundJobs, user)
	if err != nil {
		return nil, nil, fmt.Errorf("analyse jobs: %w", err)
	}

	for _, job := range foundJobs {
		rating, ok := ratings[job.ID]
		if !ok {
			continue
		}

		analysis := analyser.Analysis{
			JobID:  job.ID,
			UserID: userID,
			Rating: rating,
		}

		if err := c.AnalysisStorage.StoreAnalysis(analysis); err != nil {
			return nil, nil, fmt.Errorf("store analysis for job %s: %w", job.ID, err)
		}
	}

	return foundJobs, ratings, nil
}
