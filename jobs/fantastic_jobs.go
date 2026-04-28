package jobs

// FantasticJobsAPI is an implementation of interfaces.JobsAPI that retrieves
// job listings from the Fantastic Jobs external API.
type FantasticJobsAPI struct {
	APIKey  string
	BaseURL string
}

// NewFantasticJobsAPI creates a new FantasticJobsAPI client.
func NewFantasticJobsAPI(apiKey, baseURL string) *FantasticJobsAPI {
	return &FantasticJobsAPI{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

// GetJobs fetches job listings from the Fantastic Jobs API.
// TODO: implement HTTP request to the Fantastic Jobs API endpoint.
func (f *FantasticJobsAPI) GetJobs() ([]Job, error) {
	panic("FantasticJobsAPI.GetJobs not yet implemented")
}
