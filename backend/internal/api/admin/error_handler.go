package admin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// ValidationErrorResponse represents a structured validation error response.
type ValidationErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Fields  []FieldValidationError `json:"fields,omitempty"`
}

// FieldValidationError represents a single field validation error.
type FieldValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// handleValidationError processes gin validation errors and returns a structured response.
// It extracts field-level validation errors and provides helpful error messages.
func handleValidationError(c *gin.Context, err error, logger *zap.Logger) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fields := make([]FieldValidationError, 0, len(validationErrors))

		// Track whether this is the special "missing account_type" case
		missingAccountType := false

		for _, fieldErr := range validationErrors {
			fields = append(fields, FieldValidationError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
				Message: getValidationErrorMessage(fieldErr),
			})
			if (fieldErr.Field() == "AccountType" || strings.HasSuffix(fieldErr.Field(), ".AccountType")) &&
				fieldErr.Tag() == "required" {
				missingAccountType = true
			}
		}

		errorText := "Validation failed"
		if missingAccountType && len(fields) == 1 {
			// For backward compatibility with existing tests and clearer UX,
			// promote missing account_type to a top-level error.
			errorText = "Missing account_type field"
		}

		logger.Warn("Validation failed",
			zap.Int("field_count", len(fields)),
			zap.Any("fields", fields),
		)

		c.JSON(400, ValidationErrorResponse{
			Error:   errorText,
			Message: "One or more required fields are missing or invalid",
			Fields:  fields,
		})
		return
	}

	// Fallback for non-validation errors
	logger.Warn("Request binding failed", zap.Error(err))
	c.JSON(400, gin.H{
		"error":   "Invalid request",
		"message": err.Error(),
	})
}

// getValidationErrorMessage returns a human-friendly error message for a validation error.
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	field := fieldErr.Field()
	tag := fieldErr.Tag()
	param := fieldErr.Param()
	value := fmt.Sprintf("%v", fieldErr.Value())

	switch tag {
	case "required":
		return fmt.Sprintf("The '%s' field is required and cannot be empty", field)

	case "oneof":
		// Special handling for account_type field
		if field == "AccountType" || strings.HasSuffix(field, ".AccountType") {
			return fmt.Sprintf("The 'account_type' field must be one of: %s (received: '%s')",
				param, value)
		}
		return fmt.Sprintf("The '%s' field must be one of: %s (received: '%s')",
			field, param, value)

	case "min":
		return fmt.Sprintf("The '%s' field must be at least %s (received: '%s')",
			field, param, value)

	case "max":
		return fmt.Sprintf("The '%s' field must be at most %s (received: '%s')",
			field, param, value)

	case "email":
		return fmt.Sprintf("The '%s' field must be a valid email address (received: '%s')",
			field, value)

	case "url":
		return fmt.Sprintf("The '%s' field must be a valid URL (received: '%s')",
			field, value)

	case "len":
		return fmt.Sprintf("The '%s' field must have exactly %s characters (received: %s characters)",
			field, param, value)

	case "gte":
		return fmt.Sprintf("The '%s' field must be greater than or equal to %s (received: '%s')",
			field, param, value)

	case "lte":
		return fmt.Sprintf("The '%s' field must be less than or equal to %s (received: '%s')",
			field, param, value)

	default:
		return fmt.Sprintf("The '%s' field failed validation (tag: '%s', value: '%s')",
			field, tag, value)
	}
}

// maskSensitiveData masks sensitive fields in a request for safe logging.
// Returns a sanitized copy of the data suitable for logging.
func maskSensitiveData(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	result := make(map[string]any, len(data))
	sensitiveFields := map[string]bool{
		"code":          true,
		"api_key":       true,
		"access_token":  true,
		"refresh_token": true,
		"password":      true,
		"secret":        true,
	}

	for key, value := range data {
		lowerKey := strings.ToLower(key)

		// Check if this is a sensitive field
		if sensitiveFields[lowerKey] || strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "password") {
			// Mask the value
			if str, ok := value.(string); ok && len(str) > 0 {
				if len(str) <= 8 {
					result[key] = "***"
				} else {
					result[key] = str[:4] + "***" + str[len(str)-4:]
				}
			} else {
				result[key] = "***"
			}
		} else {
			// Handle nested objects
			if nestedMap, ok := value.(map[string]any); ok {
				result[key] = maskSensitiveData(nestedMap)
			} else {
				result[key] = value
			}
		}
	}

	return result
}

// validateOAuthCode checks if the OAuth authorization code looks valid.
func validateOAuthCode(code string) error {
	if code == "" {
		return errors.New("authorization code is required")
	}

	// OAuth authorization codes are typically 20-256 characters
	if len(code) < 10 {
		return errors.New("authorization code appears to be too short (minimum 10 characters)")
	}

	if len(code) > 512 {
		return errors.New("authorization code exceeds maximum length (512 characters)")
	}

	// Check for obvious invalid characters (whitespace, control characters)
	for _, r := range code {
		// Reject control characters (0-31) and DEL (127)
		if r < 32 || r == 127 {
			return errors.New("authorization code contains invalid characters")
		}
		// Reject whitespace characters (space, tab, etc.)
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return errors.New("authorization code contains invalid characters")
		}
	}

	return nil
}

// validateOAuthState checks if the OAuth state parameter looks valid.
func validateOAuthState(state string) error {
	if state == "" {
		return errors.New("state parameter is required")
	}

	// State should be a hex-encoded random value (our implementation uses 32 bytes = 64 hex chars)
	if len(state) < 16 {
		return errors.New("state parameter appears to be too short (minimum 16 characters)")
	}

	if len(state) > 128 {
		return errors.New("state parameter exceeds maximum length (128 characters)")
	}

	// Check for hex-only characters (our implementation)
	for _, r := range state {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			// Allow non-hex for compatibility, but it should be alphanumeric at least
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-' || r == '_') {
				return errors.New("state parameter contains invalid characters")
			}
		}
	}

	return nil
}
