package limit

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

func TestClientValidator_ValidateClaudeCode(t *testing.T) {
	logger := zap.NewNop()
	validator := NewClientValidator(logger)

	t.Run("valid claude-code user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "claude-code/1.0.0")

		valid, err := validator.ValidateClaudeCode(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("valid claude_code user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "claude_code/2.0")

		valid, err := validator.ValidateClaudeCode(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "CLAUDE-CODE/1.0")

		valid, err := validator.ValidateClaudeCode(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("invalid user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "curl/7.68.0")

		valid, err := validator.ValidateClaudeCode(headers, nil)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("empty user agent", func(t *testing.T) {
		headers := http.Header{}

		valid, err := validator.ValidateClaudeCode(headers, nil)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestClientValidator_ValidateCodex(t *testing.T) {
	logger := zap.NewNop()
	validator := NewClientValidator(logger)

	t.Run("valid cursor user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "cursor/1.0.0")

		valid, err := validator.ValidateCodex(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("valid vscode user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "vscode/1.85.0")

		valid, err := validator.ValidateCodex(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("valid codex user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "codex-client/1.0")

		valid, err := validator.ValidateCodex(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "CURSOR/1.0")

		valid, err := validator.ValidateCodex(headers, nil)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("invalid user agent", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("User-Agent", "curl/7.68.0")

		valid, err := validator.ValidateCodex(headers, nil)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestClientValidator_IsClientAllowed(t *testing.T) {
	logger := zap.NewNop()
	validator := NewClientValidator(logger)

	t.Run("allowed when restriction disabled", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: false,
			AllowedClients:          model.StringArray{"cursor"},
		}

		allowed := validator.IsClientAllowed(apiKey, "curl/7.68.0")
		assert.True(t, allowed)
	})

	t.Run("allowed when allowed list is empty", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{},
		}

		allowed := validator.IsClientAllowed(apiKey, "curl/7.68.0")
		assert.True(t, allowed)
	})

	t.Run("allowed when client in allowed list", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor", "vscode"},
		}

		allowed := validator.IsClientAllowed(apiKey, "cursor/1.0.0")
		assert.True(t, allowed)
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor"},
		}

		allowed := validator.IsClientAllowed(apiKey, "CURSOR/1.0.0")
		assert.True(t, allowed)
	})

	t.Run("partial match allowed", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor"},
		}

		allowed := validator.IsClientAllowed(apiKey, "cursor-editor/1.0.0")
		assert.True(t, allowed)
	})

	t.Run("not allowed when client not in list", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor", "vscode"},
		}

		allowed := validator.IsClientAllowed(apiKey, "curl/7.68.0")
		assert.False(t, allowed)
	})

	t.Run("multiple allowed clients", func(t *testing.T) {
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor", "vscode", "codex"},
		}

		assert.True(t, validator.IsClientAllowed(apiKey, "cursor/1.0"))
		assert.True(t, validator.IsClientAllowed(apiKey, "vscode/1.85"))
		assert.True(t, validator.IsClientAllowed(apiKey, "codex-client/1.0"))
		assert.False(t, validator.IsClientAllowed(apiKey, "wget/1.0"))
	})
}
