package util

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	_v     *validator.Validate
	_trans ut.Translator
)

// InitValidator configures the validator with English localization
func InitValidator() error {
	_v = validator.New()
	_trans, _ = ut.New(en.New(), en.New()).GetTranslator("en")
	return translations.RegisterDefaultTranslations(_v, _trans)
}

func Validate(s interface{}) error {
	if _v == nil || _trans == nil {
		return errors.New("validator not initialized")
	}
	if err := _v.Struct(s); err != nil {
		if errs := err.(validator.ValidationErrors); len(errs) > 0 {
			return errors.New(errs[0].Translate(_trans))
		}
		return err
	}
	return nil
}
