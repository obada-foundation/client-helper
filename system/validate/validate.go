package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator wraps validator package
type Validator struct {
	validate   *validator.Validate
	translator ut.Translator
}

// NewValidator returns a new Validator instance
func NewValidator() (*Validator, error) {
	validate := validator.New()

	translator, ok := ut.New(en.New(), en.New()).GetTranslator("en")
	if !ok {
		return nil, fmt.Errorf("cannot find translator for english")
	}

	if err := en_translations.RegisterDefaultTranslations(validate, translator); err != nil {
		return nil, err
	}

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate:   validate,
		translator: translator,
	}, nil
}

// Check validates the provided model against it's declared tags.
func (v *Validator) Check(val interface{}) error {
	if err := v.validate.Struct(val); err != nil {

		// Use a type assertion to get the real error value.
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		var fields FieldErrors
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(v.translator),
			}
			fields = append(fields, field)
		}

		return fields
	}

	return nil
}
