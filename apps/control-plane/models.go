package main

import "time"

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
)

const (
	JobKindBrowser = "browser"
	JobKindAPI     = "api"
)

type CreateProjectRequest struct {
	Name                string   `json:"name"`
	FrontendURL         string   `json:"frontend_url"`
	APIBaseURL          string   `json:"api_base_url"`
	OpenAPIURL          string   `json:"openapi_url"`
	AllowedHosts        []string `json:"allowed_hosts"`
	SecurityMode        string   `json:"security_mode"`
	DestructiveActions  bool     `json:"destructive_actions"`
	AllowPrivateTargets bool     `json:"allow_private_targets,omitempty"`
}

type Project struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	FrontendURL         string    `json:"frontend_url"`
	APIBaseURL          string    `json:"api_base_url"`
	OpenAPIURL          string    `json:"openapi_url"`
	AllowedHosts        []string  `json:"allowed_hosts"`
	SecurityMode        string    `json:"security_mode"`
	DestructiveActions  bool      `json:"destructive_actions"`
	AllowPrivateTargets bool      `json:"allow_private_targets"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type TestRun struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	PageTitle    string     `json:"page_title,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RunJob struct {
	ID           string     `json:"id"`
	RunID        string     `json:"run_id"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Finding struct {
	ID             string    `json:"id"`
	RunID          string    `json:"run_id,omitempty"`
	Title          string    `json:"title"`
	Severity       string    `json:"severity"`
	Category       string    `json:"category"`
	Confidence     string    `json:"confidence"`
	Description    string    `json:"description"`
	Recommendation string    `json:"recommendation"`
	EvidenceIDs    []string  `json:"evidence_ids"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

type Evidence struct {
	ID        string         `json:"id"`
	RunID     string         `json:"run_id,omitempty"`
	Type      string         `json:"type"`
	URI       string         `json:"uri"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

type Report struct {
	RunID     string         `json:"run_id"`
	ProjectID string         `json:"project_id"`
	Status    string         `json:"status"`
	Summary   ReportSummary  `json:"summary"`
	Findings  []Finding      `json:"findings"`
	Evidence  []Evidence     `json:"evidence"`
	Metadata  map[string]any `json:"metadata"`
}

type ReportSummary struct {
	TotalFindings int `json:"total_findings"`
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Info          int `json:"info"`
}
