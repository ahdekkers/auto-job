package controller

import (
	"fmt"
	"strings"

	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/storage"
	"github.com/ahdekkers/auto-job.git/users"
)

type AppController struct {
	JobAPI          *jobs.FantasticJobsAPI
	Analyser        analyser.JobAnalyser
	JobStorage      storage.JobStorage
	UserStorage     storage.UserStorage
	AnalysisStorage storage.AnalysisStorage
	SearchStorage   storage.SearchStorage
}

func NewAppController(
	jobAPI *jobs.FantasticJobsAPI,
	analyser analyser.JobAnalyser,
	jobStorage storage.JobStorage,
	userStorage storage.UserStorage,
	analysisStorage storage.AnalysisStorage,
	searchStorage storage.SearchStorage,
) *AppController {
	return &AppController{
		JobAPI:          jobAPI,
		Analyser:        analyser,
		JobStorage:      jobStorage,
		UserStorage:     userStorage,
		AnalysisStorage: analysisStorage,
		SearchStorage:   searchStorage,
	}
}

func (c *AppController) CreateUser(
	name string,
	email string,
	yearsOfExperience int,
	currentTitle string,
	preferredWorkModes []users.WorkMode,
	preferredLocations []string,
	salaryExpectation users.SalaryExpectation,
	cv string,
	additionalNotes []string,
) (users.User, error) {
	user := users.User{
		Name:               strings.TrimSpace(name),
		Email:              strings.TrimSpace(email),
		YearsOfExperience:  yearsOfExperience,
		CurrentTitle:       strings.TrimSpace(currentTitle),
		PreferredWorkModes: append([]users.WorkMode(nil), preferredWorkModes...),
		PreferredLocations: append([]string(nil), preferredLocations...),
		SalaryExpectation:  salaryExpectation,
		CV:                 strings.TrimSpace(cv),
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

func (c *AppController) SaveSearch(searchName string, params jobs.JobSearchParams) error {
	searchName = strings.TrimSpace(searchName)
	if searchName == "" {
		return fmt.Errorf("search name is required")
	}
	if c.SearchStorage == nil {
		return fmt.Errorf("search storage is nil")
	}

	if err := c.SearchStorage.StoreSearch(searchName, params); err != nil {
		return fmt.Errorf("store search: %w", err)
	}

	return nil
}

func (c *AppController) ExecuteJobSearch(
	userID string,
	searchName string,
) ([]jobs.Job, map[string]analyser.SuitabilityRating, error) {
	userID = strings.TrimSpace(userID)
	searchName = strings.TrimSpace(searchName)

	if userID == "" {
		return nil, nil, fmt.Errorf("userID is required")
	}
	if searchName == "" {
		return nil, nil, fmt.Errorf("search name is required")
	}
	if c.JobAPI == nil {
		return nil, nil, fmt.Errorf("job api client is nil")
	}
	if c.Analyser == nil {
		return nil, nil, fmt.Errorf("analyser is nil")
	}
	if c.JobStorage == nil || c.UserStorage == nil || c.AnalysisStorage == nil || c.SearchStorage == nil {
		return nil, nil, fmt.Errorf("storage dependency is nil")
	}

	user, err := c.UserStorage.RetrieveUsers(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve user: %w", err)
	}

	params, err := c.SearchStorage.RetrieveSearch(searchName)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve search: %w", err)
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
