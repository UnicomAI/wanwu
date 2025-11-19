package pkg

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/xuri/excelize/v2"
)

var sensitiveTypeMapping = map[string]string{
	"涉政":   "Political",
	"辱骂":   "Revile",
	"涉黄":   "Pornography",
	"暴恐":   "ViolentTerror",
	"违禁":   "Illegal",
	"信息安全": "InformationSecurity",
	"其他":   "Other",
}

type SensitiveRawData struct {
	Content       string
	SensitiveType string
}

// ParseSensitiveExcel 解析Excel文件返回敏感词列表 [EN] ParseSensitiveExcel parses Excel files and returns a list of sensitive words
func ParseSensitiveExcel(fileData []byte) ([]SensitiveRawData, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetList()[0]
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("invalid excel data")
	}

	headerMap, err := parseSensitiveHeader(rows[0])
	if err != nil {
		return nil, err
	}

	dataMap := make(map[string]bool)
	rawDataList := make([]SensitiveRawData, 0)

	for lineCount := 1; lineCount < len(rows); lineCount++ {
		row := rows[lineCount]
		if len(row) == 0 {
			continue
		}

		rowItems, err := processSensitiveRow(row, headerMap)
		if err != nil {
			log.Errorf("Process row %d failed: %v", lineCount+1, err)
			continue
		}

		for _, item := range rowItems {
			if item.Content == "" || dataMap[item.Content] {
				continue
			}
			dataMap[item.Content] = true
			rawDataList = append(rawDataList, item)
		}
	}

	if len(rawDataList) == 0 {
		return nil, fmt.Errorf("no valid data")
	}

	return rawDataList, nil
}

// 解析表头映射关系 [EN] Parse header mapping relationship
func parseSensitiveHeader(headerRow []string) (map[int]string, error) {

	// 创建表头映射：列索引 -> 敏感类型 [EN] Create header mapping: column index -> sensitive type
	headerMap := make(map[int]string)
	validColumns := 0

	for colIndex, colName := range headerRow {
		colName = strings.TrimSpace(colName)
		if mappedType, ok := sensitiveTypeMapping[colName]; ok {
			headerMap[colIndex] = mappedType
			validColumns++
		}
	}

	if validColumns == 0 {
		return nil, fmt.Errorf("sensitive excel header invalid")
	}

	return headerMap, nil
}

// 处理单行数据 [EN] Process single row of data
func processSensitiveRow(row []string, headerMap map[int]string) ([]SensitiveRawData, error) {
	results := make([]SensitiveRawData, 0, len(headerMap))

	for colIndex, text := range row {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		sensitiveType, ok := headerMap[colIndex]
		if !ok {
			continue
		}

		results = append(results, SensitiveRawData{
			Content:       text,
			SensitiveType: sensitiveType,
		})
	}

	return results, nil
}
