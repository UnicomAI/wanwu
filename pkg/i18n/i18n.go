package i18n

import (
	"fmt"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
)

const (
	_errCodeCol string = "err_code"
	_textKeyCol string = "text_key"
)

var _i18n *i18nConfig

type i18nConfig struct {
	DefaultLang Lang `json:"default_lang"`

	CodeKeys map[err_code.Code]map[string]*textConfig `json:"codeKeys"` // err_code -> string -> *textConfig
	Keys     map[string]*textConfig                   `json:"keys"`     // text_key -> *textConfig

}

type textConfig struct {
	Code  err_code.Code   `json:"code"`
	Key   string          `json:"key"`
	Langs map[Lang]string `json:"langs"`
}

func (cfg *textConfig) langMsg(lang, defaultLang Lang, args []string) string {
	// Defensive nil check to prevent panic
	if cfg == nil || cfg.Langs == nil {
		return fmt.Sprintf("[i18n] nil config: lang(%v) args: %v", lang, args)
	}
	
	var format string
	var ok bool
	if format, ok = cfg.Langs[lang]; !ok {
		// 没有对应语言，则使用默认语言 [EN] If there is no corresponding language, the default language will be used.
		format = cfg.Langs[defaultLang]
	}
	if format != "" {
		// []string -> []interface{}
		iargs := make([]interface{}, len(args))
		for i, arg := range args {
			iargs[i] = arg
		}
		return fmt.Sprintf(format, iargs...)
	}
	return fmt.Sprintf("lang(%v) err_code(%v) text_key(%v) args: %v", lang, cfg.Code, cfg.Key, args)
}

func initI18n(defaultLang Lang, textCfgs []*textConfig) (*i18nConfig, error) {
	ret := &i18nConfig{
		DefaultLang: defaultLang,
		CodeKeys:    make(map[err_code.Code]map[string]*textConfig),
		Keys:        make(map[string]*textConfig),
	}
	for _, textCfg := range textCfgs {
		if textCfg.Code != 0 {
			var codeKeys map[string]*textConfig
			var ok bool
			if codeKeys, ok = ret.CodeKeys[textCfg.Code]; !ok {
				codeKeys = make(map[string]*textConfig)
				ret.CodeKeys[textCfg.Code] = codeKeys
			}
			if _, ok := codeKeys[textCfg.Key]; ok {
				return nil, fmt.Errorf("i18n init err: err_code(%v) text_key(%v) already exist", textCfg.Code, textCfg.Key)
			}
			codeKeys[textCfg.Key] = textCfg
		}
		if textCfg.Key != "" {
			if _, ok := ret.Keys[textCfg.Key]; ok {
				return nil, fmt.Errorf("i18n init err: text_key(%v) already exist", textCfg.Key)
			}
			ret.Keys[textCfg.Key] = textCfg
		}
	}
	return ret, nil
}
