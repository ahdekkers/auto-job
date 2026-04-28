package jobs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FantasticJobsAPI is an implementation of interfaces.JobsAPI that retrieves
// job listings from the Fantastic Jobs external API.
type FantasticJobsAPI struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NewFantasticJobsAPI creates a new FantasticJobsAPI client.
func NewFantasticJobsAPI(apiKey string) *FantasticJobsAPI {
	return &FantasticJobsAPI{
		APIKey:  apiKey,
		BaseURL: "https://linkedin-job-search-api.p.rapidapi.com/active-jb-7d",
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetJobs fetches job listings from the Fantastic Jobs API.
func (f *FantasticJobsAPI) GetJobs(params JobSearchParams) ([]Job, error) {
	endpoint, err := url.Parse(f.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	q := endpoint.Query()
	addQueryParams(q, params)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Add("x-rapidapi-key", f.APIKey)
	req.Header.Add("x-rapidapi-host", "linkedin-job-search-api.p.rapidapi.com")
	req.Header.Add("Content-Type", "application/json")

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var rawJobs []upstreamJob
	if err := json.NewDecoder(resp.Body).Decode(&rawJobs); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	jobs := make([]Job, 0, len(rawJobs))
	for _, r := range rawJobs {
		jobs = append(jobs, mapUpstreamJob(r))
	}

	return jobs, nil
}

func addQueryParams(q url.Values, p JobSearchParams) {
	if p.Limit != nil {
		q.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		q.Set("offset", strconv.Itoa(*p.Offset))
	}
	if p.TitleFilter != "" {
		q.Set("title_filter", p.TitleFilter)
	}
	if p.LocationFilter != "" {
		q.Set("location_filter", p.LocationFilter)
	}
	if p.DescriptionFilter != "" {
		q.Set("description_filter", p.DescriptionFilter)
	}
	if p.OrganizationDescription != "" {
		q.Set("organization_description_filter", p.OrganizationDescription)
	}
	if p.OrganizationSpecialties != "" {
		q.Set("organization_specialties_filter", p.OrganizationSpecialties)
	}
	if p.OrganizationSlugFilter != "" {
		q.Set("organization_slug_filter", p.OrganizationSlugFilter)
	}
	if p.DescriptionType != "" {
		q.Set("description_type", p.DescriptionType)
	}
	if p.TypeFilter != "" {
		q.Set("type_filter", p.TypeFilter)
	}
	if p.Remote != nil {
		q.Set("remote", strconv.FormatBool(*p.Remote))
	}
	if p.Agency != nil {
		q.Set("agency", strconv.FormatBool(*p.Agency))
	}
	if p.IndustryFilter != "" {
		q.Set("industry_filter", p.IndustryFilter)
	}
	if p.SeniorityFilter != "" {
		q.Set("seniority_filter", p.SeniorityFilter)
	}
	if p.DateFilter != "" {
		q.Set("date_filter", p.DateFilter)
	}
	if p.ExcludeATSDuplicate != nil {
		q.Set("exclude_ats_duplicate", strconv.FormatBool(*p.ExcludeATSDuplicate))
	}
	if p.ExternalApplyURL != nil {
		q.Set("external_apply_url", strconv.FormatBool(*p.ExternalApplyURL))
	}
	if p.DirectApply != nil {
		q.Set("directapply", strconv.FormatBool(*p.DirectApply))
	}
	if p.EmployeesLTE != nil {
		q.Set("employees_lte", strconv.Itoa(*p.EmployeesLTE))
	}
	if p.EmployeesGTE != nil {
		q.Set("employees_gte", strconv.Itoa(*p.EmployeesGTE))
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}
	if p.AdvancedTitleFilter != "" {
		q.Set("advanced_title_filter", p.AdvancedTitleFilter)
	}
}

type upstreamJob struct {
	ID                                  string     `json:"id"`
	Title                               string     `json:"title"`
	Organization                        string     `json:"organization"`
	LocationsDerived                    []string   `json:"locations_derived"`
	RemoteDerived                       bool       `json:"remote_derived"`
	SalaryRaw                           *salaryRaw `json:"salary_raw"`
	EmploymentType                      []string   `json:"employment_type"`
	URL                                 string     `json:"url"`
	ExternalApplyURL                    string     `json:"external_apply_url"`
	DescriptionText                     string     `json:"description_text"`
	LinkedinOrgRecruitmentAgencyDerived bool       `json:"linkedin_org_recruitment_agency_derived"`
	LocationType                        *string    `json:"location_type"`
}

type salaryRaw struct {
	Currency string      `json:"currency"`
	Value    salaryValue `json:"value"`
}

type salaryValue struct {
	MinValue *float64 `json:"minValue"`
	MaxValue *float64 `json:"maxValue"`
	UnitText string   `json:"unitText"`
}

func mapUpstreamJob(r upstreamJob) Job {
	location := ""
	if len(r.LocationsDerived) > 0 {
		location = r.LocationsDerived[0]
	}

	workMode := ""
	switch {
	case r.RemoteDerived:
		workMode = "remote"
	case r.LocationType != nil && *r.LocationType != "":
		workMode = strings.ToLower(*r.LocationType)
	default:
		workMode = "onsite"
	}

	salary := ""
	if r.SalaryRaw != nil {
		salary = formatSalary(r.SalaryRaw)
	}

	jobType := strings.Join(r.EmploymentType, ", ")

	applyURL := r.ExternalApplyURL
	if applyURL == "" {
		applyURL = r.URL
	}

	return Job{
		ID:           r.ID,
		Title:        r.Title,
		Company:      r.Organization,
		Location:     location,
		WorkMode:     workMode,
		Salary:       salary,
		JobType:      jobType,
		Description:  r.DescriptionText,
		Requirements: []string{},
		Benefits:     []string{},
		ApplyURL:     applyURL,
		SourceID:     r.ID,
	}
}

func formatSalary(s *salaryRaw) string {
	if s == nil {
		return ""
	}

	unit := strings.ToLower(s.Value.UnitText)
	if unit == "" {
		unit = "period"
	}

	parts := make([]string, 0, 3)

	if s.Value.MinValue != nil && s.Value.MaxValue != nil {
		parts = append(parts, fmt.Sprintf("%g-%g", *s.Value.MinValue, *s.Value.MaxValue))
	} else if s.Value.MinValue != nil {
		parts = append(parts, fmt.Sprintf("from %g", *s.Value.MinValue))
	} else if s.Value.MaxValue != nil {
		parts = append(parts, fmt.Sprintf("up to %g", *s.Value.MaxValue))
	}

	if s.Currency != "" {
		parts = append(parts, s.Currency)
	}
	if unit != "" {
		parts = append(parts, unit)
	}

	return strings.Join(parts, " ")
}
