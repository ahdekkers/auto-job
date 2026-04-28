package jobs

// JobSearchParams holds optional query parameters for the API request.
type JobSearchParams struct {
	Limit                   *int
	Offset                  *int
	TitleFilter             string
	LocationFilter          string
	DescriptionFilter       string
	OrganizationDescription string
	OrganizationSpecialties string
	OrganizationSlugFilter  string
	DescriptionType         string
	TypeFilter              string
	Remote                  *bool
	Agency                  *bool
	IndustryFilter          string
	SeniorityFilter         string
	DateFilter              string
	ExcludeATSDuplicate     *bool
	ExternalApplyURL        *bool
	DirectApply             *bool
	EmployeesLTE            *int
	EmployeesGTE            *int
	Order                   string
	AdvancedTitleFilter     string
}

// Job represents a job listing retrieved from a jobs API.
type Job struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Company      string   `json:"company"`
	Location     string   `json:"location"`
	WorkMode     string   `json:"work_mode"`
	Salary       string   `json:"salary"`
	JobType      string   `json:"job_type"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	Benefits     []string `json:"benefits"`
	ApplyURL     string   `json:"apply_url"`
	SourceID     string   `json:"source_id"` // original ID from the upstream API
}
