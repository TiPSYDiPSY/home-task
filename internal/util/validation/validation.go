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

const (
	maxDecimalPlaces = 2
)

func NewValidator() *Validator {
	v := &Validator{
		validate: validator.New(),
	}

	if err := v.validate.RegisterValidation("decimal2", validateDecimal2); err != nil {
		panic(fmt.Sprintf("failed to register decimal2 validator: %v", err))
	}

	return v
}

func validateDecimal2(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if idx := strings.IndexByte(value, '.'); idx != -1 {
		return len(value)-idx-1 <= maxDecimalPlaces
	}

	return true
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

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fe.Field(), fe.Param())
	case "decimal2":
		return fe.Field() + " must have at most 2 decimal places"
	default:
		return fe.Field() + " is invalid"
	}
}
