package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
)

func ValidateRequestData(dataFromRequest any) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(dataFromRequest)
	if err != nil {

		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return err
		}

		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
				switch e.Tag() {
				case "required":
					return commonErrors.NewAPIError(http.StatusUnprocessableEntity, fmt.Sprintf("the %s field is required", e.Field()))
				case "min":
					return commonErrors.NewAPIError(http.StatusUnprocessableEntity, fmt.Sprintf("the %s field must be at least %s characters long", e.Field(), e.Param()))
				case "max":
					return commonErrors.NewAPIError(http.StatusUnprocessableEntity, fmt.Sprintf("the %s field must be at most %s characters long", e.Field(), e.Param()))
				case "email":
					return commonErrors.NewAPIError(http.StatusUnprocessableEntity, fmt.Sprintf("the %s field must be a valid email address", e.Field()))
				case "eqfield":
					return commonErrors.NewAPIError(http.StatusUnprocessableEntity, fmt.Sprintf("the %s field must be the same as %s field", e.Field(), e.Param()))
				}
			}
			return err
		}

		return err
	}
	return nil
}
