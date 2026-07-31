package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelParams contains parameters for Excel generation
type ExcelParams struct {
	Symbol      string
	CompanyName string
	Period      string
	TTMEPS      float64
	IncludePE   bool
	Data        []StockData
	PeriodData  []PeriodData
}

// GenerateExcel creates an Excel file from stock data
func GenerateExcel(params ExcelParams) (*excelize.File, error) {
	f := excelize.NewFile()

	sheetName := "Stock Data"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Style for header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	// Style for right-aligned text
	rightAlignStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	// Style for numbers (right-aligned)
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt:    4, // #,##0.00
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	// Style for negative values (light red background, right-aligned)
	negStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFC7CE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	// Add metadata
	setCell(f, sheetName, 1, 1, "Symbol:", rightAlignStyle)
	setCell(f, sheetName, 2, 1, params.Symbol, rightAlignStyle)
	setCell(f, sheetName, 1, 2, "Company:", rightAlignStyle)
	setCell(f, sheetName, 2, 2, params.CompanyName, rightAlignStyle)
	setCell(f, sheetName, 1, 3, "Period:", rightAlignStyle)
	setCell(f, sheetName, 2, 3, params.Period, rightAlignStyle)
	if params.IncludePE {
		setCell(f, sheetName, 1, 4, "TTM EPS:", rightAlignStyle)
		setCell(f, sheetName, 2, 4, params.TTMEPS, rightAlignStyle)
	}

	row := 6

	if params.PeriodData != nil {
		row = writePeriodData(f, sheetName, row, params.PeriodData, params.IncludePE, headerStyle, negStyle, rightAlignStyle)
	} else {
		row = writeDailyData(f, sheetName, row, params.Data, params.IncludePE, headerStyle, numberStyle, negStyle, rightAlignStyle)
	}

	// Auto-fit columns
	_ = row // silence unused warning
	for col := 1; col <= 16; col++ {
		colName, _ := excelize.ColumnNumberToName(col)
		_ = f.SetColWidth(sheetName, colName, colName, 12)
	}

	return f, nil
}

// writeDailyData writes daily stock data to Excel
func writeDailyData(f *excelize.File, sheet string, startRow int, data []StockData, includePE bool, headerStyle, numberStyle, negStyle, rightAlignStyle int) int {
	headers := []string{"Date", "Open", "High", "Low", "Close", "Volume", "Change", "HChange"}
	if includePE {
		headers = append(headers, "PE")
	}

	// Write headers
	for col, h := range headers {
		setCell(f, sheet, col+1, startRow, h, headerStyle)
	}
	startRow++

	// Write data rows
	for _, d := range data {
		setCell(f, sheet, 1, startRow, d.Date, rightAlignStyle)
		setCellNum(f, sheet, 2, startRow, d.Open, numberStyle)
		setCellNum(f, sheet, 3, startRow, d.High, numberStyle)
		setCellNum(f, sheet, 4, startRow, d.Low, numberStyle)
		setCellNum(f, sheet, 5, startRow, d.Close, numberStyle)
		setCell(f, sheet, 6, startRow, d.Volume, rightAlignStyle)
		setChangeCell(f, sheet, 7, startRow, d.Change, negStyle, rightAlignStyle)
		setChangeCell(f, sheet, 8, startRow, d.HChange, negStyle, rightAlignStyle)
		if includePE {
			setCell(f, sheet, 9, startRow, d.PE, rightAlignStyle)
		}
		startRow++
	}
	return startRow
}

// writePeriodData writes period aggregated data to Excel
func writePeriodData(f *excelize.File, sheet string, startRow int, data []PeriodData, includePE bool, headerStyle, negStyle, rightAlignStyle int) int {
	headers := []string{"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "HChange"}
	if includePE {
		headers = append(headers, "PE")
	}
	headers = append(headers, "Days", "C/L-2%", "C/L-3%", "C/L-4%", "C/L-5%")

	// Write headers
	for col, h := range headers {
		setCell(f, sheet, col+1, startRow, h, headerStyle)
	}
	startRow++

	// Write data rows
	for _, p := range data {
		col := 1
		setCell(f, sheet, col, startRow, p.Period, rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, p.StartDate, rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, p.EndDate, rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.Open), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.High), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.Low), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.Close), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, p.Volume, rightAlignStyle)
		col++
		setChangeCell(f, sheet, col, startRow, p.Change, negStyle, rightAlignStyle)
		col++
		setChangeCell(f, sheet, col, startRow, p.HChange, negStyle, rightAlignStyle)
		col++
		if includePE {
			setCell(f, sheet, col, startRow, p.PE, rightAlignStyle)
			col++
		}
		setCell(f, sheet, col, startRow, p.Days, rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop2Pct.Close, p.Drop2Pct.Low), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop3Pct.Close, p.Drop3Pct.Low), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop4Pct.Close, p.Drop4Pct.Low), rightAlignStyle)
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop5Pct.Close, p.Drop5Pct.Low), rightAlignStyle)
		startRow++
	}
	return startRow
}

// Helper functions
func setCell(f *excelize.File, sheet string, col, row int, value interface{}, style int) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, value)
	_ = f.SetCellStyle(sheet, cell, cell, style)
}

func setCellNum(f *excelize.File, sheet string, col, row int, value string, style int) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	v, _ := strconv.ParseFloat(value, 64)
	_ = f.SetCellValue(sheet, cell, v)
	_ = f.SetCellStyle(sheet, cell, cell, style)
}

func parseFloatStr(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func setChangeCell(f *excelize.File, sheet string, col, row int, value string, negStyle, rightAlignStyle int) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, value)
	if isNegPct(value) {
		_ = f.SetCellStyle(sheet, cell, cell, negStyle)
	} else {
		_ = f.SetCellStyle(sheet, cell, cell, rightAlignStyle)
	}
}

func isNegPct(s string) bool {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return v < 0
}