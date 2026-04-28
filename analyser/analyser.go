package analyser

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ahdekkers/auto-job.git/jobs"
	"github.com/ahdekkers/auto-job.git/ollama"
	"github.com/ahdekkers/auto-job.git/users"
)

type Analysis struct {
	JobID  string            `json:"job_id"`
	UserID string            `json:"user_id"`
	Rating SuitabilityRating `json:"rating"`
}

// SuitabilityRating is a value between 0.0 and 1.0 representing how well a job
// matches a user's profile, where 1.0 is a perfect match.
type SuitabilityRating float64

// JobAnalyser defines the contract for analysing how suitable a job is for a user.
type JobAnalyser interface {
	GetSuitabilityRating(job jobs.Job, user users.User) (SuitabilityRating, error)
	GetAllSuitabilityRatings(jobs []jobs.Job, user users.User) (map[string]SuitabilityRating, error)
}

// AIAnalyser is a JobAnalyser implementation backed by Ollama.
type AIAnalyser struct {
	Model     string
	Ollama    *ollama.OllamaClient
	UseReason bool
}

// NewAIAnalyserFromFile loads the prompt from a file and creates a new analyser.
func NewAIAnalyserFromFile(model, ollamaBaseURL string) (*AIAnalyser, error) {

	return &AIAnalyser{
		Model:  model,
		Ollama: ollama.NewOllamaClient(ollamaBaseURL),
	}, nil
}

// NewAIAnalyser creates a new analyser from an already-loaded prompt string.
func NewAIAnalyser(model string, ollamaClient *ollama.OllamaClient) *AIAnalyser {
	if ollamaClient == nil {
		ollamaClient = ollama.NewOllamaClient("")
	}

	return &AIAnalyser{
		Model:  model,
		Ollama: ollamaClient,
	}
}

type analysisRequest struct {
	Job  jobs.Job   `json:"job"`
	User users.User `json:"user"`
}

type analysisResponse struct {
	Rating float64 `json:"rating"`
	Reason string  `json:"reason,omitempty"`
}

func (a *AIAnalyser) GetSuitabilityRating(job jobs.Job, user users.User) (SuitabilityRating, error) {
	if a.Ollama == nil {
		return 0, fmt.Errorf("ollama client is nil")
	}
	if strings.TrimSpace(a.Model) == "" {
		return 0, fmt.Errorf("model is required")
	}

	payload := analysisRequest{
		Job:  job,
		User: user,
	}

	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("marshal analysis payload: %w", err)
	}

	fullPrompt := prompt + string(payloadJSON)
	ctx := context.Background()
	raw, err := a.Ollama.Generate(ctx, a.Model, fullPrompt)
	if err != nil {
		return 0, err
	}

	rating, err := parseRatingResponse(raw)
	if err != nil {
		return 0, fmt.Errorf("parse model response: %w\nraw response: %s", err, raw)
	}
	return SuitabilityRating(rating), nil
}

func (a *AIAnalyser) GetAllSuitabilityRatings(jobList []jobs.Job, user users.User) (map[string]SuitabilityRating, error) {
	results := make(map[string]SuitabilityRating, len(jobList))

	for _, job := range jobList {
		rating, err := a.GetSuitabilityRating(job, user)
		if err != nil {
			return nil, fmt.Errorf("analyse job %s: %w", job.ID, err)
		}
		results[job.ID] = rating
	}

	return results, nil
}

func parseRatingResponse(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)

	// First try strict JSON.
	var resp analysisResponse
	if err := json.Unmarshal([]byte(raw), &resp); err == nil {
		if resp.Rating < 0 || resp.Rating > 1 {
			return 0, fmt.Errorf("rating out of range: %v", resp.Rating)
		}
		return resp.Rating, nil
	}

	// Fallback: try plain numeric response.
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		if f < 0 || f > 1 {
			return 0, fmt.Errorf("rating out of range: %v", f)
		}
		return f, nil
	}

	return 0, fmt.Errorf("response was not valid JSON or a float")
}

const prompt = `
You are a job suitability analyser.

Your task is to estimate how well a job matches a user's profile and preferences.

Score the job from 0.0 to 1.0 using these factors:
- title relevance
- skill match
- years of experience fit
- location fit
- work mode fit
- salary fit
- seniority fit
- alignment with extra notes and dealbreakers
- quality of the CV summary match

Scoring guidance:
- 1.0 = excellent match, very likely a strong fit
- 0.8 = strong match with only minor gaps
- 0.6 = reasonable match but with some notable gaps
- 0.4 = weak match
- 0.2 = very poor match
- 0.0 = completely unsuitable

Be strict when there are dealbreakers such as:
- wrong work mode
- wrong location
- salary far below expectation
- insufficient experience
- missing key required skills

Be generous when the job is adjacent to the user's current profile and the user has transferable skills.

Important:
- Return only valid JSON.
- Return exactly one field named "rating".
- The rating must be a number between 0.0 and 1.0.
- Do not include markdown, prose, or code fences.

Example output:
{"rating":0.74}

INPUT JSON:
`
