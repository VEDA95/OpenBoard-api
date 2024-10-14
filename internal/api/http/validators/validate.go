package validators

import "github.com/go-playground/validator/v10"

type ErrorResponse struct {
	FailedField string      `json:"failed_field"`
	Tag         string      `json:"tag"`
	Value       interface{} `json:"value"`
	ErrValue    string      `json:"err_value"`
}

type Validator struct {
	validator *validator.Validate
}

var Instance *Validator

func NewValidator() *Validator {
	validate := validator.New()

	return &Validator{validator: validate}
}

func InitializeValidatorInstance() {
	Instance = NewValidator()
}

func (validate *Validator) Validate(data interface{}) []*ErrorResponse {
	errs := validate.validator.Struct(data)

	if errs == nil {
		return nil
	}

	validationErrors := make([]*ErrorResponse, 0)

	for _, err := range errs.(validator.ValidationErrors) {
		elem := &ErrorResponse{
			FailedField: err.Field(),
			Tag:         err.Tag(),
			Value:       err.Value(),
			ErrValue:    err.Error(),
		}
		validationErrors = append(validationErrors, elem)
	}

	return validationErrors
}
