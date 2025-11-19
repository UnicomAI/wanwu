package util

import (
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/xuri/excelize/v2"
)

func ReadExcelColumn(filePath string, columnNo int) ([]string, error) {
	// 1. Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Errorf("打开Excel文件失败: %v", err)
	}
	defer func() {
		// close file
		if err := f.Close(); err != nil {
			log.Errorf("关闭Excel文件时出错: %v", err)
		}
	}()

	// 2. Get the worksheet list
	sheets := f.GetSheetList()
	sheet := sheets[0]

	// 3. Get all rows in the worksheet
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	var result []string
	// 4. Traverse rows and cells
	for _, row := range rows {
		result = append(result, row[columnNo])
	}

	return result, nil
}
