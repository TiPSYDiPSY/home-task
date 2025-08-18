package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) ValidateStruct(s any) error {
	if err := v.validate.Struct(s); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var errorMessages []string
			for _, validationErr := range validationErrors {
				errorMessages = append(errorMessages, getErrorMessage(validationErr))
			}

			return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
		}

		return fmt.Errorf("validation error: %w", err)
	}

	return nil
}

func ValidateStruct(s interface{}) error {
	return NewValidator().ValidateStruct(s)
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fe.Field(), fe.Param())
	default:
		return fe.Field() + " is invalid"
	}
}
