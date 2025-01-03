package validators

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"reflect"
	"strings"
)

type ErrorResponse struct {
	FailedField string      `json:"failed_field"`
	Tag         string      `json:"tag"`
	Value       interface{} `json:"value"`
	ErrValue    string      `json:"err_value"`
}

type Validator struct {
	validator  *validator.Validate
	translator ut.Translator
}

type ErrorResponseMap = map[string]*ErrorResponse

var Instance *Validator

func NewValidator() *Validator {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	enInstance := en.New()
	uni := ut.New(enInstance, enInstance)
	trans, _ := uni.GetTranslator("en")

	enTranslations.RegisterDefaultTranslations(validate, trans)

	validate.RegisterTranslation(
		"required",
		trans,
		func(translator ut.Translator) error {
			return translator.Add("required", "A value must be provided", true)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			translatedError, _ := translator.T("required")

			return translatedError
		},
	)

	validate.RegisterTranslation(
		"min",
		trans,
		func(translator ut.Translator) error {
			return translator.Add("min", "The value provided needs to be longer", true)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			translatedError, _ := translator.T("min")

			return translatedError
		},
	)

	validate.RegisterTranslation(
		"max",
		trans,
		func(translator ut.Translator) error {
			return translator.Add("max", "The value provided needs to shorter", true)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			translatedError, _ := translator.T("max")

			return translatedError
		},
	)

	validate.RegisterTranslation(
		"email",
		trans,
		func(translator ut.Translator) error {
			return translator.Add("email", "A valid email address must be provided", true)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			translatedError, _ := translator.T("email")

			return translatedError
		},
	)

	validate.RegisterTranslation(
		"oneof",
		trans,
		func(translator ut.Translator) error {
			return translator.Add("oneof", "Only the following values may be selected: {0}", true)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			splitParams := strings.Split(fieldError.Param(), " ")
			translatedError, _ := translator.T("oneof", strings.Join(splitParams, ", "))

			return translatedError
		},
	)

	validate.RegisterTranslation(
		"eqfield",
		trans,
		func(translator ut.Translator) error {
			return translator.Add("eqfield", "The value provided must be the same as the value provided for the field: {0}", true)
		},
		func(translator ut.Translator, fieldError validator.FieldError) string {
			translatedError, _ := translator.T("eqfield", fieldError.Param())

			return translatedError
		},
	)

	return &Validator{validator: validate, translator: trans}
}

func InitializeValidatorInstance() {
	Instance = NewValidator()
}

func (validate *Validator) Validate(data interface{}) ErrorResponseMap {
	errs := validate.validator.Struct(data)

	if errs == nil {
		return nil
	}

	validationErrors := make(ErrorResponseMap)

	for _, err := range errs.(validator.ValidationErrors) {
		errField := err.Field()
		validationErrors[errField] = &ErrorResponse{
			FailedField: errField,
			Tag:         err.Tag(),
			Value:       err.Value(),
			ErrValue:    err.Translate(validate.translator),
		}
	}

	return validationErrors
}
