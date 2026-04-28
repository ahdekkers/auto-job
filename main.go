package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ahdekkers/auto-job.git/analyser"
	"github.com/ahdekkers/auto-job.git/controller"
	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/storage"
	"github.com/ahdekkers/auto-job.git/users"
)

func main() {
	action := flag.String("action", "", "Controller action to run: create-user, save-search, or search")

	apiKey := flag.String("api-key", "", "RapidAPI key for the jobs API")
	ollamaBaseURL := flag.String("ollama-base-url", "http://localhost:11434", "Ollama base URL")
	model := flag.String("model", "gemma3", "Ollama model name")
	storageFile := flag.String("storage-file", "data.json", "Path to JSON storage file")

	name := flag.String("name", "", "User name")
	email := flag.String("email", "", "User email")
	yearsOfExperience := flag.String("years-of-experience", "", "Years of experience")
	currentTitle := flag.String("current-title", "", "Current job title")
	preferredWorkModes := flag.String("preferred-work-modes", "", "Comma-separated work modes: remote,onsite,hybrid,any")
	preferredLocations := flag.String("preferred-locations", "", "Comma-separated preferred locations")
	salaryMin := flag.String("salary-min", "", "Minimum salary expectation")
	salaryMax := flag.String("salary-max", "", "Maximum salary expectation")
	cvFile := flag.String("cv-file", "", "Path to CV file")
	additionalNotes := flag.String("additional-notes", "", "Comma-separated additional notes")

	searchName := flag.String("search-name", "", "Unique saved search name")
	userID := flag.String("user-id", "", "User ID for job search, usually the user's email")

	limit := flag.String("limit", "", "API limit")
	offset := flag.String("offset", "", "API offset")
	titleFilter := flag.String("title-filter", "", "Title filter")
	locationFilter := flag.String("location-filter", "", "Location filter")
	descriptionFilter := flag.String("description-filter", "", "Description filter")
	organizationDescriptionFilter := flag.String("organization-description-filter", "", "Organization description filter")
	organizationSpecialtiesFilter := flag.String("organization-specialties-filter", "", "Organization specialties filter")
	organizationSlugFilter := flag.String("organization-slug-filter", "", "Organization slug filter")
	descriptionType := flag.String("description-type", "", "Description type: text or html")
	typeFilter := flag.String("type-filter", "", "Job type filter")
	remote := flag.String("remote", "", "Remote only filter: true or false")
	agency := flag.String("agency", "", "Agency filter: true or false")
	industryFilter := flag.String("industry-filter", "", "Industry filter")
	seniorityFilter := flag.String("seniority-filter", "", "Seniority filter")
	dateFilter := flag.String("date-filter", "", "Date filter")
	excludeAtsDuplicate := flag.String("exclude-ats-duplicate", "", "Exclude ATS duplicates: true or false")
	externalApplyURL := flag.String("external-apply-url", "", "Only jobs with external apply URL: true or false")
	directApply := flag.String("directapply", "", "Direct apply filter: true or false")
	employeesLTE := flag.String("employees-lte", "", "Maximum company size")
	employeesGTE := flag.String("employees-gte", "", "Minimum company size")
	order := flag.String("order", "", "Order: asc or default descending")
	advancedTitleFilter := flag.String("advanced-title-filter", "", "Advanced title filter")

	flag.Parse()

	if strings.TrimSpace(*action) == "" {
		exitErr(fmt.Errorf("missing required --action"))
	}

	jobAPI := jobs.NewFantasticJobsAPI(*apiKey)

	aiAnalyser, err := analyser.NewAIAnalyserFromFile(*model, *ollamaBaseURL)
	if err != nil {
		exitErr(fmt.Errorf("initialise analyser: %w", err))
	}

	fileStorage := storage.NewFileStorage(*storageFile)

	ctrl := controller.NewAppController(
		jobAPI,
		aiAnalyser,
		fileStorage,
		fileStorage,
		fileStorage,
		fileStorage,
	)

	switch strings.ToLower(strings.TrimSpace(*action)) {
	case "create-user":
		runCreateUser(ctrl, *name, *email, *yearsOfExperience, *currentTitle, *preferredWorkModes, *preferredLocations, *salaryMin, *salaryMax, *cvFile, *additionalNotes)

	case "save-search":
		runSaveSearch(ctrl, *searchName, buildSearchParams(
			limit, offset, titleFilter, locationFilter, descriptionFilter,
			organizationDescriptionFilter, organizationSpecialtiesFilter, organizationSlugFilter,
			descriptionType, typeFilter, remote, agency, industryFilter, seniorityFilter,
			dateFilter, excludeAtsDuplicate, externalApplyURL, directApply, employeesLTE,
			employeesGTE, order, advancedTitleFilter,
		))

	case "search":
		if strings.TrimSpace(*searchName) == "" {
			exitErr(fmt.Errorf("missing required --search-name for search"))
		}
		if strings.TrimSpace(*userID) == "" {
			exitErr(fmt.Errorf("missing required --user-id for search"))
		}

		runSearch(ctrl, *userID, *searchName)

	default:
		exitErr(fmt.Errorf("unknown --action %q (use create-user, save-search, or search)", *action))
	}
}

func runCreateUser(
	ctrl *controller.AppController,
	name, email, yearsStr, currentTitle, modesStr, locationsStr, salaryMinStr, salaryMaxStr, cvFile, notesStr string,
) {
	years, err := parseRequiredInt(yearsStr, "--years-of-experience")
	if err != nil {
		exitErr(err)
	}

	salaryMin, err := parseRequiredInt(salaryMinStr, "--salary-min")
	if err != nil {
		exitErr(err)
	}

	salaryMax, err := parseRequiredInt(salaryMaxStr, "--salary-max")
	if err != nil {
		exitErr(err)
	}

	data, err := os.ReadFile(cvFile)
	if err != nil {
		exitErr(fmt.Errorf("read cv file: %w", err))
	}
	cvText := string(data)

	user, err := ctrl.CreateUser(
		name,
		email,
		years,
		currentTitle,
		parseWorkModes(modesStr),
		splitCSV(locationsStr),
		users.SalaryExpectation{Min: salaryMin, Max: salaryMax},
		cvText,
		splitCSV(notesStr),
	)
	if err != nil {
		exitErr(err)
	}

	printJSON(user)
}

func runSaveSearch(ctrl *controller.AppController, searchName string, params jobs.JobSearchParams) {
	if err := ctrl.SaveSearch(searchName, params); err != nil {
		exitErr(err)
	}

	printJSON(map[string]string{
		"status":      "saved",
		"search_name": searchName,
	})
}

func runSearch(ctrl *controller.AppController, userID, searchName string) {
	jobsFound, ratings, err := ctrl.ExecuteJobSearch(userID, searchName)
	if err != nil {
		exitErr(err)
	}

	type result struct {
		Jobs    []jobs.Job                            `json:"jobs"`
		Ratings map[string]analyser.SuitabilityRating `json:"ratings"`
	}

	printJSON(result{
		Jobs:    jobsFound,
		Ratings: ratings,
	})
}

func buildSearchParams(
	limit, offset, titleFilter, locationFilter, descriptionFilter,
	organizationDescriptionFilter, organizationSpecialtiesFilter, organizationSlugFilter,
	descriptionType, typeFilter, remote, agency, industryFilter, seniorityFilter,
	dateFilter, excludeAtsDuplicate, externalApplyURL, directApply, employeesLTE,
	employeesGTE, order, advancedTitleFilter *string,
) jobs.JobSearchParams {
	return jobs.JobSearchParams{
		Limit:                   mustIntPtr(limit),
		Offset:                  mustIntPtr(offset),
		TitleFilter:             valueOrEmpty(titleFilter),
		LocationFilter:          valueOrEmpty(locationFilter),
		DescriptionFilter:       valueOrEmpty(descriptionFilter),
		OrganizationDescription: valueOrEmpty(organizationDescriptionFilter),
		OrganizationSpecialties: valueOrEmpty(organizationSpecialtiesFilter),
		OrganizationSlugFilter:  valueOrEmpty(organizationSlugFilter),
		DescriptionType:         valueOrEmpty(descriptionType),
		TypeFilter:              valueOrEmpty(typeFilter),
		Remote:                  mustBoolPtr(remote),
		Agency:                  mustBoolPtr(agency),
		IndustryFilter:          valueOrEmpty(industryFilter),
		SeniorityFilter:         valueOrEmpty(seniorityFilter),
		DateFilter:              valueOrEmpty(dateFilter),
		ExcludeATSDuplicate:     mustBoolPtr(excludeAtsDuplicate),
		ExternalApplyURL:        mustBoolPtr(externalApplyURL),
		DirectApply:             mustBoolPtr(directApply),
		EmployeesLTE:            mustIntPtr(employeesLTE),
		EmployeesGTE:            mustIntPtr(employeesGTE),
		Order:                   valueOrEmpty(order),
		AdvancedTitleFilter:     valueOrEmpty(advancedTitleFilter),
	}
}

func parseRequiredInt(value, flagName string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("missing required %s", flagName)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", flagName, err)
	}
	return n, nil
}

func mustIntPtr(value *string) *int {
	if value == nil {
		return nil
	}
	s := strings.TrimSpace(*value)
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		exitErr(fmt.Errorf("invalid integer value %q: %w", s, err))
	}
	return &n
}

func mustBoolPtr(value *string) *bool {
	if value == nil {
		return nil
	}
	s := strings.TrimSpace(*value)
	if s == "" {
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		exitErr(fmt.Errorf("invalid boolean value %q: %w", s, err))
	}
	return &b
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseWorkModes(s string) []users.WorkMode {
	raw := splitCSV(s)
	if len(raw) == 0 {
		return nil
	}

	out := make([]users.WorkMode, 0, len(raw))
	for _, item := range raw {
		switch strings.ToLower(item) {
		case "remote":
			out = append(out, users.WorkModeRemote)
		case "onsite":
			out = append(out, users.WorkModeOnsite)
		case "hybrid":
			out = append(out, users.WorkModeHybrid)
		case "any":
			out = append(out, users.WorkModeAny)
		default:
			exitErr(fmt.Errorf("invalid work mode %q: use remote, onsite, hybrid, or any", item))
		}
	}
	return out
}

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		exitErr(fmt.Errorf("encode json: %w", err))
	}
	fmt.Println(string(b))
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
