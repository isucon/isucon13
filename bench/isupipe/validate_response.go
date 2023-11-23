package isupipe

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"go.uber.org/zap"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func ValidateResponse(req *http.Request, response interface{}) error {
	lgr := zap.S()
	if err := validate.Struct(response); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			lgr.Warnf("ValidateResponse: invalid validation error発生: %s\n", err.Error())
			return bencherror.NewInternalError(err)
		}

		var errorFields []string
		for _, err := range err.(validator.ValidationErrors) {
			errorFields = append(errorFields, err.Namespace())
		}

		return bencherror.NewEmptyHttpResponseError(errorFields, req)
	}

	return nil
}

func ValidateSlice(req *http.Request, slice interface{}) error {
	lgr := zap.S()
	if err := validate.Var(slice, "dive,required"); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			lgr.Warnf("ValidateSlice: invalid validation error発生: %s\n", err.Error())
			return bencherror.NewInternalError(err)
		}

		var errorFields []string
		for _, err := range err.(validator.ValidationErrors) {
			errorFields = append(errorFields, err.Namespace())
		}

		return bencherror.NewEmptyHttpResponseError(errorFields, req)
	}

	return nil
}
