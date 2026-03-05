package utils

import (
	"fmt"
	"reflect"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelColumn Excel列定义
type ExcelColumn struct {
	Header    string  // 列标题
	FieldName string  // 字段名（对应struct字段）
	Width     float64 // 列宽（可选）
}

// ExportToExcel 通用Excel导出函数
// data: 数据切片（必须是struct切片）
// columns: 列定义
// sheetName: 工作表名称
// 返回: Excel文件的字节数组
func ExportToExcel(data interface{}, columns []ExcelColumn, sheetName string) ([]byte, error) {
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 创建工作表
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("创建工作表失败: %w", err)
	}
	f.SetActiveSheet(index)

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建样式失败: %w", err)
	}

	// 写入表头
	for colIdx, col := range columns {
		cellName, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cellName, col.Header)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)

		// 设置列宽
		if col.Width > 0 {
			colName, _ := excelize.ColumnNumberToName(colIdx + 1)
			f.SetColWidth(sheetName, colName, colName, col.Width)
		}
	}

	// 通过反射获取数据
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("data must be a slice")
	}

	// 写入数据行
	for rowIdx := 0; rowIdx < dataValue.Len(); rowIdx++ {
		item := dataValue.Index(rowIdx)

		// 如果是指针，解引用
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		if item.Kind() != reflect.Struct {
			continue
		}

		// 写入每个字段
		for colIdx, col := range columns {
			field := item.FieldByName(col.FieldName)
			if !field.IsValid() {
				continue
			}

			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			cellValue := formatFieldValue(field)
			f.SetCellValue(sheetName, cellName, cellValue)
		}
	}

	// 自动筛选
	if dataValue.Len() > 0 {
		lastCol, _ := excelize.ColumnNumberToName(len(columns))
		lastRow := dataValue.Len() + 1
		f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)
	}

	// 冻结首行
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// 生成字节数组
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成Excel文件失败: %w", err)
	}

	return buf.Bytes(), nil
}

// formatFieldValue 格式化字段值
func formatFieldValue(field reflect.Value) interface{} {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint()
	case reflect.Float32, reflect.Float64:
		return field.Float()
	case reflect.Bool:
		if field.Bool() {
			return "是"
		}
		return "否"
	case reflect.Struct:
		// 处理time.Time类型
		if t, ok := field.Interface().(time.Time); ok {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02 15:04:05")
		}
		return field.String()
	case reflect.Ptr:
		if field.IsNil() {
			return ""
		}
		return formatFieldValue(field.Elem())
	default:
		return fmt.Sprintf("%v", field.Interface())
	}
}

// SetExcelResponse 设置Excel响应头
func SetExcelResponse(filename string) map[string]string {
	return map[string]string{
		"Content-Type":        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s.xlsx", filename),
	}
}
