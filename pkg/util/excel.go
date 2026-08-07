package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/xuri/excelize/v2"
)

// ReadColumnsOptions 读取多列数据时的配置项
//   - Sheet: 工作表名称（为空时默认第一个工作表）
//   - ColIndexes: 需要读取的列索引（从0开始）
//   - SkipRows: 需要跳过的行数（通常用于跳过表头）
type ReadColumnsOptions struct {
	Sheet      string
	ColIndexes []int
	SkipRows   int
}

// ReadWithHeaderMappingOptions 带表头映射读取时的配置项
//   - Sheet: 工作表名称（为空时默认第一个工作表）
//   - HeaderRow: 表头所在行号（从0开始）
//   - HeaderMapping: 原始表头 -> 目标字段名 的映射
type ReadWithHeaderMappingOptions struct {
	Sheet         string
	HeaderRow     int
	HeaderMapping map[string]string
}

// Workbook Excel 工作簿封装，仿照 openapi3_util.Client 思路
// 集中管理 *excelize.File，并提供高层 API。
type Workbook struct {
	f *excelize.File
}

// NewWorkbook 创建空工作簿（用于导出/写入），与 OpenWorkbook* 对称。
func NewWorkbook() *Workbook {
	return &Workbook{f: excelize.NewFile()}
}

// OpenWorkbookFromBytes 从字节数组打开 Excel 工作簿
func OpenWorkbookFromBytes(data []byte) (*Workbook, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		log.Errorf("打开Excel文件失败: %v", err)
		return nil, err
	}
	return &Workbook{f: f}, nil
}

// OpenWorkbookFromFile 从文件路径打开 Excel 工作簿
func OpenWorkbookFromFile(path string) (*Workbook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return OpenWorkbookFromBytes(data)
}

// Close 关闭工作簿
func (wb *Workbook) Close() error {
	if wb == nil || wb.f == nil {
		return nil
	}
	return wb.f.Close()
}

// WriteTo 将工作簿写入 w（例如 HTTP ResponseWriter），写入后应 Close。
func (wb *Workbook) WriteTo(w io.Writer) (int64, error) {
	if wb == nil || wb.f == nil {
		return 0, fmt.Errorf("workbook is nil")
	}
	return wb.f.WriteTo(w)
}

// GetSheets 返回所有工作表名称列表
func (wb *Workbook) GetSheets() ([]string, error) {
	if wb == nil || wb.f == nil {
		return nil, fmt.Errorf("workbook is nil")
	}
	return wb.f.GetSheetList(), nil
}

// GetRows 读取指定工作表全部行（sheet 为空时使用第一个工作表）。
func (wb *Workbook) GetRows(sheet string) ([][]string, error) {
	if wb == nil || wb.f == nil {
		return nil, fmt.Errorf("workbook is nil")
	}
	name, err := resolveSheetName(wb.f, sheet)
	if err != nil {
		return nil, err
	}
	return wb.f.GetRows(name)
}

// CreateSheet 新建工作表并设为活动工作表。
func (wb *Workbook) CreateSheet(sheet string) (int, error) {
	if wb == nil || wb.f == nil {
		return -1, fmt.Errorf("workbook is nil")
	}
	index, err := wb.f.NewSheet(sheet)
	if err != nil {
		return -1, err
	}
	wb.f.SetActiveSheet(index)
	return index, nil
}

// WriteRow 写入一行（row 从 1 开始，与 excelize 一致）。
func (wb *Workbook) WriteRow(sheet string, row int, values []any) error {
	if wb == nil || wb.f == nil {
		return fmt.Errorf("workbook is nil")
	}
	for i, val := range values {
		cell, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			return err
		}
		if err := wb.f.SetCellValue(sheet, cell, val); err != nil {
			return err
		}
	}
	return nil
}

// WriteSheet 创建工作簿并写入表头 + 数据行（调用方负责 Close）。
func WriteSheet(sheet string, title []any, rows [][]any) (*Workbook, error) {
	wb := NewWorkbook()
	if _, err := wb.CreateSheet(sheet); err != nil {
		_ = wb.Close()
		return nil, err
	}
	if err := wb.WriteRow(sheet, 1, title); err != nil {
		_ = wb.Close()
		return nil, err
	}
	for i, row := range rows {
		if err := wb.WriteRow(sheet, i+2, row); err != nil {
			_ = wb.Close()
			return nil, err
		}
	}
	return wb, nil
}

// ReadColumn 读取单列数据（带跳过行、表名配置）
func (wb *Workbook) ReadColumn(opts ReadColumnsOptions) ([]string, error) {
	if wb == nil || wb.f == nil {
		return nil, fmt.Errorf("workbook is nil")
	}
	sheet, err := resolveSheetName(wb.f, opts.Sheet)
	if err != nil {
		return nil, err
	}
	if opts.ColIndexes == nil || len(opts.ColIndexes) != 1 {
		return nil, fmt.Errorf("ReadColumn requires exactly one ColIndex")
	}
	colIndex := opts.ColIndexes[0]
	if colIndex < 0 {
		return nil, fmt.Errorf("column index must be >= 0, got %d", colIndex)
	}
	if opts.SkipRows < 0 {
		return nil, fmt.Errorf("skipRows must be >= 0, got %d", opts.SkipRows)
	}

	rows, err := wb.f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	var result []string
	for i, row := range rows {
		if i < opts.SkipRows {
			continue
		}
		if colIndex < len(row) {
			result = append(result, row[colIndex])
		} else {
			result = append(result, "")
		}
	}
	return result, nil
}

// ReadWithHeaderMapping 使用表头映射读取为 key-value 形式
func (wb *Workbook) ReadWithHeaderMapping(opts ReadWithHeaderMappingOptions) ([]map[string]string, error) {
	if wb == nil || wb.f == nil {
		return nil, fmt.Errorf("workbook is nil")
	}
	if opts.HeaderRow < 0 {
		return nil, fmt.Errorf("headerRow must be >= 0, got %d", opts.HeaderRow)
	}
	if len(opts.HeaderMapping) == 0 {
		return nil, fmt.Errorf("headerMapping must not be empty")
	}

	targetSheet, err := resolveSheetName(wb.f, opts.Sheet)
	if err != nil {
		return nil, err
	}

	rows, err := wb.f.GetRows(targetSheet)
	if err != nil {
		return nil, err
	}
	if len(rows) <= opts.HeaderRow {
		return nil, fmt.Errorf("invalid excel: not enough rows (rows=%d, headerRow=%d)", len(rows), opts.HeaderRow)
	}

	// 构建表头映射：原始表头名称 -> 目标名称 + 列索引
	targetHeaderMap := make(map[int]string) // 列索引 -> 目标名称
	for colIdx, colName := range rows[opts.HeaderRow] {
		colName = trimInvisibleSpace(colName)
		if targetName, ok := opts.HeaderMapping[colName]; ok {
			targetHeaderMap[colIdx] = targetName
		}
	}

	var result []map[string]string
	for i := opts.HeaderRow + 1; i < len(rows); i++ {
		row := rows[i]
		rowMap := make(map[string]string)
		for colIdx, targetName := range targetHeaderMap {
			if colIdx < len(row) {
				rowMap[targetName] = trimInvisibleSpace(row[colIdx])
			}
		}
		result = append(result, rowMap)
	}
	return result, nil
}

// --- internal ---

// trimInvisibleSpace 去除字符串首尾的可见空白，并去除 Excel 表头常见的不可见字符
// 例如零宽空格（U+200B）、字节序标记（U+FEFF）等。
func trimInvisibleSpace(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.Is(unicode.Cf, r) || // 格式字符（Format）
			unicode.Is(unicode.Cc, r) || // 控制字符（Control）
			unicode.IsSpace(r) // 标准空白字符
	})
}

func resolveSheetName(f *excelize.File, sheet string) (string, error) {
	targetSheet := sheet
	if targetSheet == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return "", fmt.Errorf("invalid excel: no sheets")
		}
		targetSheet = sheets[0]
	}
	return targetSheet, nil
}
