package model

import "time"

const (
	AIProtocolOpenAI          = "openai"
	AIProtocolOpenAIResponses = "openai_responses"
	AIProtocolAnthropic       = "anthropic"
	AIProtocolGemini          = "gemini"
)

type AISettings struct {
	Enabled   bool      `json:"enabled"`
	Protocol  string    `json:"protocol"`
	BaseURL   string    `json:"base_url"`
	APIKey    string    `json:"api_key"`
	Model     string    `json:"model"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AISettingsStatus struct {
	Enabled bool `json:"enabled"`
}

type AIInterpretationResponse struct {
	RunID        int64  `json:"run_id"`
	Protocol     string `json:"protocol"`
	Model        string `json:"model"`
	Content      string `json:"content"`
	LogTruncated bool   `json:"log_truncated"`
}

var DefaultAISettings = AISettings{
	Enabled:  false,
	Protocol: AIProtocolOpenAI,
	BaseURL:  "",
	APIKey:   "",
	Model:    "",
	UserAgent: "",
}
