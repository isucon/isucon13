package isupipe

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/isucon/isucon13/bench/internal/bencherror"
)

var validate *validator.Validate

func ValidateResponse(req *http.Request, response interface{}) error {
	if err := validate.Struct(response); err != nil {
		_, ok := err.(*validator.InvalidValidationError)
		if !ok {
			return bencherror.NewInternalError(err)
		}

		var errorFields []string
		for _, err := range err.(validator.ValidationErrors) {
			errorFields = append(errorFields, err.Field())
		}

		return bencherror.NewEmptyHttpResponseError(errorFields, req)
	}

	return nil
}
