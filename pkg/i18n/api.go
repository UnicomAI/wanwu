package i18n

import (
	"errors"
	"fmt"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
)

type Lang string

type Config struct {
	Type        string         `json:"type" mapstructure:"type"`
	XlsxPath    string         `json:"xlsxPath" mapstructure:"xlsxPath"`
	XlsxSheets  []string       `json:"xlsxSheets" mapstructure:"xlsxSheets"`
	JsonlPath   string         `json:"jsonlPath" mapstructure:"jsonlPath"`
	Langs       []LangCodeName `json:"langs" mapstructure:"langs"`
	DefaultLang string         `json:"defaultLang" mapstructure:"defaultLang"`
}

type LangCodeName struct {
	// https://i18ns.com/languagecode.html
	Code string `json:"code" mapstructure:"code"`
	Name string `json:"name" mapstructure:"name"`
}

func Init(cfg Config) error {
	if _i18n != nil {
		return errors.New("i18n already init")
	}

	var textCfgs []*textConfig
	var err error
	switch cfg.Type {
	case "xlsx":
		var langs []string
		for _, lang := range cfg.Langs {
			langs = append(langs, lang.Code)
		}
		textCfgs, err = loadXlsxTextConfigs(cfg.XlsxPath, cfg.XlsxSheets, langs)
	case "jsonl":
		textCfgs, err = loadJsonlTextConfigs(cfg.JsonlPath)
	default:
		return errors.New("invalid i18n type")
	}
	if err != nil {
		return err
	}

	_i18n, err = initI18n(Lang(cfg.DefaultLang), textCfgs)
	return err
}

func DefaultLang() Lang {
	if _i18n != nil {
		return _i18n.DefaultLang
	}
	return ""
}

func ByCode(lang Lang, code err_code.Code, args []string) string {
	return ByCodeOrKey(lang, code, "", args)
}

func ByKey(lang Lang, key string, args []string) string {
	return ByCodeOrKey(lang, 0, key, args)
}

func ByCodeOrKey(lang Lang, code err_code.Code, key string, args []string) string {
	// Defensive nil check to prevent panic
	if _i18n == nil {
		return fmt.Sprintf("[i18n not initialized] lang(%v) err_code(%v) text_key(%v) args: %v", lang, code, key, args)
	}
	
	var textCfg *textConfig
	if key != "" {
		textCfg = _i18n.Keys[key]
	} else if code != 0 {
		if codeKeys, ok := _i18n.CodeKeys[code]; ok {
			if key != "" {
				textCfg = codeKeys[key]
			} else if len(codeKeys) == 1 {
				for key := range codeKeys {
					textCfg = codeKeys[key]
				}
			} else {
				textCfg = codeKeys[""]
			}
		}
	}
	if textCfg != nil {
		return textCfg.langMsg(lang, _i18n.DefaultLang, args)
	}
	return fmt.Sprintf("[i18n] lang(%v) err_code(%v) text_key(%v) args: %v", lang, code, key, args)
}
