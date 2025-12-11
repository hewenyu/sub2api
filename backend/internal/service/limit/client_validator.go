package limit

import (
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// ClientValidator manages client validation for API keys.
type ClientValidator interface {
	ValidateClaudeCode(headers http.Header, body []byte) (bool, error)
	ValidateCodex(headers http.Header, body []byte) (bool, error)
	IsClientAllowed(apiKey *model.APIKey, userAgent string) bool
}

type clientValidator struct {
	logger *zap.Logger
}

// NewClientValidator creates a new client validator.
func NewClientValidator(logger *zap.Logger) ClientValidator {
	return &clientValidator{logger: logger}
}

// ValidateClaudeCode validates Claude Code client requests.
func (v *clientValidator) ValidateClaudeCode(headers http.Header, body []byte) (bool, error) {
	userAgent := headers.Get("User-Agent")

	// Claude Code User-Agent pattern
	matched, _ := regexp.MatchString(`(?i)claude[-_]code`, userAgent)
	return matched, nil
}

// ValidateCodex validates Codex client requests.
func (v *clientValidator) ValidateCodex(headers http.Header, body []byte) (bool, error) {
	userAgent := headers.Get("User-Agent")

	// Codex client patterns (Cursor, VSCode, etc.)
	codexPatterns := []string{
		`(?i)cursor`,
		`(?i)vscode`,
		`(?i)codex`,
	}

	for _, pattern := range codexPatterns {
		matched, _ := regexp.MatchString(pattern, userAgent)
		if matched {
			return true, nil
		}
	}

	return false, nil
}

// IsClientAllowed checks if a client is allowed for the API key.
func (v *clientValidator) IsClientAllowed(apiKey *model.APIKey, userAgent string) bool {
	if !apiKey.EnableClientRestriction {
		return true
	}

	if len(apiKey.AllowedClients) == 0 {
		return true
	}

	// Check allowed list
	userAgentLower := strings.ToLower(userAgent)
	for _, allowed := range apiKey.AllowedClients {
		allowedLower := strings.ToLower(allowed)
		if strings.Contains(userAgentLower, allowedLower) {
			return true
		}
	}

	v.logger.Warn("Client not allowed",
		zap.Int64("api_key_id", apiKey.ID),
		zap.String("user_agent", userAgent),
	)

	return false
}
