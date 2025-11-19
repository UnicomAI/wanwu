package util

import (
	"errors"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	_v     *validator.Validate
	_trans ut.Translator
)

// InitValidator validator中文本地化 [EN] InitValidator validator Chinese localization
func InitValidator() error {
	_v = validator.New()
	_trans, _ = ut.New(zh.New(), zh.New()).GetTranslator("zh")
	return translations.RegisterDefaultTranslations(_v, _trans)
}

func Validate(s interface{}) error {
	if _v == nil || _trans == nil {
		return errors.New("validator未初始化")
	}
	if err := _v.Struct(s); err != nil {
		if errs := err.(validator.ValidationErrors); len(errs) > 0 {
			return errors.New(errs[0].Translate(_trans))
		}
		return err
	}
	return nil
}
