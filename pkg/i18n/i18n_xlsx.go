package i18n

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/xuri/excelize/v2"
)

func loadXlsxTextConfigs(i18nXlsxPath string, sheets, langs []string) ([]*textConfig, error) {
	// open xlsx
	f, err := excelize.OpenFile(i18nXlsxPath)
	if err != nil {
		return nil, fmt.Errorf("i18n open %v err: %v", i18nXlsxPath, err)
	}
	defer f.Close()

	// load sheet
	var ret []*textConfig
	for _, sheet := range sheets {
		textCfgs, err := loadXlsxSheetTextConfigs(f, sheet, langs)
		if err != nil {
			return nil, fmt.Errorf("i18n load %v steet %v err: %v", i18nXlsxPath, sheet, err)
		}
		ret = append(ret, textCfgs...)
	}
	return ret, nil
}

func loadXlsxSheetTextConfigs(f *excelize.File, sheet string, langs []string) ([]*textConfig, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("sheet(%v) rows err: %v", sheet, err)
	}
	// check title
	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet(%v) check title empty", sheet)
	}
	// get title col idx
	colIdxes := make(map[string]int)
	colIdxes[_errCodeCol] = -1
	colIdxes[_textKeyCol] = -1
	for _, lang := range langs {
		colIdxes[lang] = -1
	}
	for idx, col := range rows[0] {
		colIdx, ok := colIdxes[col]
		if !ok {
			continue
		}
		if colIdx >= 0 {
			return nil, fmt.Errorf("sheet(%v) check title(%v) duplicate", sheet, col)
		}
		colIdxes[col] = idx
	}
	// get text configs
	var ret []*textConfig
	for idx := 1; idx < len(rows); idx++ {
		if len(rows[idx]) == 0 {
			// Allow skipping absolutely empty lines
			continue
		}
		textCfg, err := row2textConfig(rows[idx], colIdxes)
		if err != nil {
			return nil, fmt.Errorf("sheet(%v) row(%v) to textConfig err: %v", sheet, idx+1, err)
		}
		ret = append(ret, textCfg)
	}
	return ret, nil
}

func row2textConfig(row []string, colIdxes map[string]int) (*textConfig, error) {
	ret := &textConfig{
		Langs: make(map[Lang]string),
	}
	for col, colIdx := range colIdxes {
		if colIdx < 0 || colIdx >= len(row) {
			switch col {
			case _errCodeCol, _textKeyCol:
				return nil, fmt.Errorf("invalid col(%v) index(%v)", col, colIdx)
			default:
				// There is no corresponding language column, or the corresponding language text is empty
			}
			continue
		}
		colValue := trimInvisibleSpace(row[colIdx])
		switch col {
		case _errCodeCol:
			code, err := str2code(colValue)
			if err != nil {
				return nil, err
			}
			ret.Code = code
		case _textKeyCol:
			ret.Key = colValue
		default:
			if colValue != "" {
				// The corresponding language text is empty and no key is added.
				ret.Langs[Lang(col)] = colValue
			}
		}
	}
	return ret, nil
}

func str2code(s string) (err_code.Code, error) {
	if s == "" {
		return 0, nil
	}
	code, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("convert (%v) to err_code err: %v", s, err)
	}
	return err_code.Code(code), nil
}

func trimInvisibleSpace(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		// Remove Unicode format characters (such as zero-width spaces U+200B, U+FEFF, etc.)
		return unicode.Is(unicode.Cf, r) || // Format character (Format)
			unicode.Is(unicode.Cc, r) || // Control character (Control)
			unicode.IsSpace(r) // Standard whitespace characters
	})
}
