package users

type WorkMode string

const (
	WorkModeRemote WorkMode = "remote"
	WorkModeOnsite WorkMode = "onsite"
	WorkModeHybrid WorkMode = "hybrid"
	WorkModeAny    WorkMode = "any"
)

// SalaryExpectation represents the user's expected salary range and currency.
type SalaryExpectation struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// User represents the person searching for a job and their preferences.
type User struct {
	Name               string            `json:"name"`
	Email              string            `json:"email"`
	YearsOfExperience  int               `json:"years_of_experience"`
	CurrentTitle       string            `json:"current_title"`
	PreferredWorkModes []WorkMode        `json:"preferred_work_modes"`
	PreferredLocations []string          `json:"preferred_locations"`
	SalaryExpectation  SalaryExpectation `json:"salary_expectation"`
	CV                 string            `json:"cv"`               // plain text CV
	AdditionalNotes    []string          `json:"additional_notes"` // bullet points of extra preferences or dealbreakers
}
