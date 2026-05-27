// pkg/validator/validator.go
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// single instance — validator is safe for concurrent use, create once
var validate = validator.New()

// ValidationError holds all field errors in a clean structure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate runs validation and returns human-readable errors
// returns nil if everything is valid
func Validate(input interface{}) []ValidationError {
	err := validate.Struct(input)
	if err == nil {
		return nil // no errors — all good
	}

	// Cast to ValidationErrors to get individual field errors
	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   toSnakeCase(err.Field()),
			Message: toMessage(err),
		})
	}
	return errors
}

// toMessage converts validator error tags to human readable messages
func toMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", toSnakeCase(err.Field()))
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", err.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", err.Param())
	case "uuid4":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", err.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", err.Param())
	default:
		return fmt.Sprintf("failed validation: %s", err.Tag())
	}
}

// toSnakeCase converts "FullName" → "full_name" for consistent error field names
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
