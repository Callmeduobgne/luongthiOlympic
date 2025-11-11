package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator instance
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate validates a struct
func (v *Validator) Validate(data interface{}) error {
	if err := v.validate.Struct(data); err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors formats validation errors into a readable message
func (v *Validator) formatValidationErrors(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			messages = append(messages, fmt.Sprintf(
				"field '%s' validation failed on '%s' tag",
				fieldError.Field(),
				fieldError.Tag(),
			))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}
	return err
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

